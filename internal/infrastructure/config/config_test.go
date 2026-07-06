package config

import (
	"testing"
)

func TestLoadUsesDefaultLanguageWhenUnset(t *testing.T) {
	t.Setenv("LANGUAGE", "")
	t.Setenv("TELEGRAM_TOKEN", "")

	cfg := Load()
	if cfg.Lang != "ru" {
		t.Fatalf("expected default language ru, got %s", cfg.Lang)
	}
}

func TestLoadUsesProvidedLanguage(t *testing.T) {
	t.Setenv("LANGUAGE", "en")
	t.Setenv("TELEGRAM_TOKEN", "")

	cfg := Load()
	if cfg.Lang != "en" {
		t.Fatalf("expected language en, got %s", cfg.Lang)
	}
}

func TestLoadUsesBotTokenWhenTelegramTokenMissing(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "")
	t.Setenv("BOT_TOKEN", "test-token")

	cfg := Load()
	if cfg.TelegramToken != "test-token" {
		t.Fatalf("expected TELEGRAM_TOKEN to be used, got %s", cfg.TelegramToken)
	}
}
