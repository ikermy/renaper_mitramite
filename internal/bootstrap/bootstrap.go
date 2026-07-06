package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"renaper_mitramite/internal/application"
	"renaper_mitramite/internal/domain"
	"renaper_mitramite/internal/infrastructure/config"
	"renaper_mitramite/internal/infrastructure/scraper"
	telegramadapter "renaper_mitramite/internal/infrastructure/telegram"

	"github.com/ikermy/AiR_Logger/v2/pkg/logger"
	tb "gopkg.in/telebot.v4"
)

func Run(ctx context.Context) {
	cfg := config.Load()
	if cfg.TelegramToken == "" {
		logger.Fatal("TELEGRAM_TOKEN is not set")
	}
	if cfg.UseWebhook && cfg.PublicURL == "" {
		logger.Fatal("PUBLIC_URL is required when webhook mode is enabled")
	}

	checker := scraper.NewChecker(30 * time.Second)
	usecase := application.NewBot(checker, cfg)

	settings := tb.Settings{Token: cfg.TelegramToken}
	var webhook *tb.Webhook
	if cfg.UseWebhook {
		webhookURL := buildWebhookURL(cfg.PublicURL)
		webhook = &tb.Webhook{
			// IgnoreSetWebhook: false — telebot сам зарегистрирует URL в Telegram.
			// Listen оставляем пустым — HTTP-сервер управляем сами.
			Endpoint: &tb.WebhookEndpoint{PublicURL: webhookURL},
		}
		settings.Poller = webhook
		logger.Info("Starting Telegram bot with webhook on port %s using public URL %s", cfg.Port, webhookURL)
	} else {
		settings.Poller = &tb.LongPoller{Timeout: 10 * time.Second}
		logger.Info("Starting Telegram bot with long polling")
	}

	bot, err := tb.NewBot(settings)
	if err != nil {
		logger.Fatal("create bot: %v", err)
	}

	adapter := telegramadapter.NewAdapter(bot, usecase)
	adapter.Register()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", HealthHandler)
	mux.HandleFunc("/readyz", HealthHandler)
	if cfg.UseWebhook && webhook != nil {
		// Синхронный обработчик webhook:
		// HTTP-ответ Telegram отправляется только ПОСЛЕ завершения обработки
		// (включая chromedp). Пока запрос открыт, Cloud Run не троттлит CPU.
		mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
			var upd tb.Update
			if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			bot.ProcessUpdate(upd) // синхронно — не возвращается до конца обработки
			w.WriteHeader(http.StatusOK)
		})
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bot is running")
	})

	healthServer := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server stopped: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received")
		// На завершение активных chromedp-запросов.
		shutCtx, cancel := context.WithTimeout(context.Background(), 39*time.Second)
		defer cancel()
		if err := healthServer.Shutdown(shutCtx); err != nil {
			logger.Error("health server shutdown error: %v", err)
		}
		bot.Stop()
		domain.CloseExit()
	}()

	logger.Info("Telegram bot started")
	bot.Start()
}

func buildWebhookURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(strings.ToLower(baseURL), "/webhook") {
		return baseURL
	}
	return baseURL + "/webhook"
}
