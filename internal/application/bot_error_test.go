package application

import (
	"errors"
	"testing"

	"renaper_mitramite/internal/infrastructure/config"
	"renaper_mitramite/internal/infrastructure/scraper"
)

type failingChecker struct{}

func (f failingChecker) CheckTramite(_ string) (string, string, bool, error) {
	return "", "", false, errors.New("boom")
}

type noDataChecker struct{}

func (f noDataChecker) CheckTramite(_ string) (string, string, bool, error) {
	return "No data", "", false, nil
}

type recaptchaChecker struct{}

func (f recaptchaChecker) CheckTramite(_ string) (string, string, bool, error) {
	return "", `{"errors":{"status":100,"source":"/busqueda.php","tipo":"ERROR","title":"Error reCAPTCHA. Por favor recargá la página y volvé a intentar","detail":"reCAPTCHA para: tramite:00123456789"}}`, false, nil
}

type captchaBlockedChecker struct{}

func (f captchaBlockedChecker) CheckTramite(_ string) (string, string, bool, error) {
	return "", "", true, scraper.ErrCaptchaBlocked
}

func TestHandleTextShowsErrorForScraperFailure(t *testing.T) {
	bot := NewBot(failingChecker{}, config.Config{Lang: "ru"})

	reply, err := bot.HandleText(1, "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.Text == "" {
		t.Fatal("expected error message")
	}
}

func TestHandleCallbackRawWithoutStoredDataShowsNoRawMessage(t *testing.T) {
	bot := NewBot(fakeChecker{}, config.Config{Lang: "ru"})

	reply, err := bot.HandleCallback(2, "raw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.Text == "" {
		t.Fatal("expected no raw message")
	}
}

func TestHandleTextShowsNoDataMessageForEmptyScraperResponse(t *testing.T) {
	bot := NewBot(noDataChecker{}, config.Config{Lang: "ru"})

	reply, err := bot.HandleText(3, "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.Text == "" {
		t.Fatal("expected no-data message")
	}
	if reply.Markup == nil || len(reply.Markup.InlineKeyboard) == 0 {
		t.Fatal("expected main menu markup after no-data response")
	}
}

func TestHandleTextShowsRecaptchaMessageForCaptchaPayload(t *testing.T) {
	bot := NewBot(recaptchaChecker{}, config.Config{Lang: "ru"})

	reply, err := bot.HandleText(4, "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.Text == "" {
		t.Fatal("expected reCAPTCHA message")
	}
	if reply.Markup == nil || len(reply.Markup.InlineKeyboard) == 0 {
		t.Fatal("expected main menu markup after reCAPTCHA response")
	}
}
