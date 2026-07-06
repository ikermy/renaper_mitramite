package config

import (
	"os"
	"strings"

	"github.com/ikermy/AiR_Logger/v2/pkg/logger"
)

type Config struct {
	TelegramToken string
	PublicURL     string
	Port          string
	Lang          string
	UseWebhook    bool
}

func Load() Config {
	telegramToken := firstNonEmpty(strings.TrimSpace(os.Getenv("TELEGRAM_TOKEN")), strings.TrimSpace(os.Getenv("BOT_TOKEN")))
	if telegramToken == "" {
		logger.Warn("TELEGRAM_TOKEN environment variable is not set")
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	publicURL := strings.TrimSpace(os.Getenv("PUBLIC_URL"))
	useWebhook := strings.EqualFold(strings.TrimSpace(os.Getenv("USE_WEBHOOK")), "true") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("USE_WEBHOOK")), "1") ||
		publicURL != ""

	lang := strings.TrimSpace(os.Getenv("LANGUAGE"))
	if lang == "" {
		lang = "ru"
	}

	if !useWebhook {
		logger.Info("Webhook disabled; using long polling")
	}

	return Config{
		TelegramToken: telegramToken,
		PublicURL:     publicURL,
		Port:          port,
		Lang:          lang,
		UseWebhook:    useWebhook,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
