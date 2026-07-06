package scraper

import (
	"testing"
	"time"
)

func TestNewCheckerUsesDefaultTimeout(t *testing.T) {
	checker := NewChecker(0)
	if checker.timeout <= 0 {
		t.Fatal("expected positive timeout")
	}
}

func TestNewCheckerUsesProvidedTimeout(t *testing.T) {
	checker := NewChecker(3 * time.Second)
	if checker.timeout != 3*time.Second {
		t.Fatalf("expected timeout 3s, got %s", checker.timeout)
	}
}

func TestNormalizeResponseString(t *testing.T) {
	out, err := normalizeResponse("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "abc" {
		t.Fatalf("expected abc, got %s", out)
	}
}
