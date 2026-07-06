package application

import (
	"testing"

	"renaper_mitramite/internal/domain"
	"renaper_mitramite/internal/infrastructure/config"
)

type fakeChecker struct{}

func (f fakeChecker) CheckTramite(_ string) (string, string, bool, error) {
	return "ok", "{}", false, nil
}

func TestHandleCommandStartShowsMainMenu(t *testing.T) {
	bot := NewBot(fakeChecker{}, config.Config{Lang: "ru"})

	reply, err := bot.HandleCommand(42, "start")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.Text == "" {
		t.Fatal("expected non-empty welcome text")
	}
	if reply.Markup == nil || len(reply.Markup.InlineKeyboard) == 0 {
		t.Fatal("expected main menu markup")
	}
}

func TestHandleTextInvalidNumberReturnsValidationMessage(t *testing.T) {
	bot := NewBot(fakeChecker{}, config.Config{Lang: "en"})

	reply, err := bot.HandleText(7, "invalid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.Text == "" {
		t.Fatal("expected validation message")
	}
	if reply.Markup == nil || len(reply.Markup.InlineKeyboard) == 0 {
		t.Fatal("expected main menu markup")
	}
}

func TestHandleCallbackLanguageChangesLocale(t *testing.T) {
	bot := NewBot(fakeChecker{}, config.Config{Lang: "ru"})

	_, err := bot.HandleCallback(11, "lang_en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := bot.locale(11); got != "en" {
		t.Fatalf("expected locale en, got %s", got)
	}
}

func TestStatePersistsForUser(t *testing.T) {
	bot := NewBot(fakeChecker{}, config.Config{Lang: "ru"})
	bot.setWaiting(99, true)
	state := bot.states[99]
	if !state.WaitingForTramite {
		t.Fatal("expected waiting state to be true")
	}
	if state.Locale != "" {
		t.Fatalf("expected empty locale, got %s", state.Locale)
	}
}

func TestDomainModels(t *testing.T) {
	state := domain.UserState{Locale: "es"}
	if state.Locale != "es" {
		t.Fatalf("expected locale es, got %s", state.Locale)
	}
}
