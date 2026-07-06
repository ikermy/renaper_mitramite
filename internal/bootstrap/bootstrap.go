package bootstrap

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"renaper_mitramite/internal/application"
	"renaper_mitramite/internal/domain"
	"renaper_mitramite/internal/infrastructure/config"
	"renaper_mitramite/internal/infrastructure/scraper"
	telegramadapter "renaper_mitramite/internal/infrastructure/telegram"

	tb "gopkg.in/telebot.v4"
)

func Run(ctx context.Context) {
	cfg := config.Load()
	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN is not set")
	}
	if cfg.UseWebhook && cfg.PublicURL == "" {
		log.Fatal("PUBLIC_URL is required when webhook mode is enabled")
	}

	checker := scraper.NewChecker(10 * time.Second)
	usecase := application.NewBot(checker, cfg)

	settings := tb.Settings{Token: cfg.TelegramToken}
	if cfg.UseWebhook {
		settings.Poller = &tb.Webhook{
			Listen:   ":" + cfg.Port,
			Endpoint: &tb.WebhookEndpoint{PublicURL: cfg.PublicURL},
		}
		log.Printf("Starting Telegram bot with webhook on port %s using public URL %s", cfg.Port, cfg.PublicURL)
	} else {
		settings.Poller = &tb.LongPoller{Timeout: 10 * time.Second}
		log.Printf("Starting Telegram bot with long polling")
	}

	bot, err := tb.NewBot(settings)
	if err != nil {
		log.Fatalf("create bot: %v", err)
	}

	adapter := telegramadapter.NewAdapter(bot, usecase)
	adapter.Register()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", HealthHandler)
	mux.HandleFunc("/readyz", HealthHandler)

	healthServer := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("health server stopped: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		log.Println("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := healthServer.Shutdown(ctx); err != nil {
			log.Printf("health server shutdown error: %v", err)
		}
		bot.Stop()
		domain.CloseExit()
	}()

	log.Printf("Telegram bot started")
	bot.Start()
}
