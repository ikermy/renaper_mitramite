package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"renaper_mitramite/internal/domain"

	"github.com/chromedp/chromedp"
	"github.com/ikermy/AiR_Logger/v2/pkg/logger"
)

var (
	ErrUnavailable     = errors.New("scraper unavailable")
	ErrInvalidResponse = errors.New("scraper returned an invalid response")
	ErrCaptchaBlocked  = errors.New("captcha blocked")
)

type apiErrorResponse struct {
	Errors struct {
		Status int    `json:"status"`
		Source string `json:"source"`
		Tipo   string `json:"tipo"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

type Checker struct {
	timeout time.Duration
}

func NewChecker(timeout time.Duration) *Checker {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Checker{timeout: timeout}
}

func (c *Checker) CheckTramite(tramiteID string) (string, string, bool, error) {
	const maxAttempts = 2
	var lastErr error
	restarted := false

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		browserCtx, cleanup := newBrowserContext(ctx)
		var raw string
		jsCode := fmt.Sprintf(`
        (() => {
          window.__rnpResult = null;
          const waitForRecaptcha = () => new Promise((resolve, reject) => {
            const startedAt = Date.now();
            const check = () => {
              if (window.grecaptcha && typeof window.grecaptcha.ready === "function" && typeof window.grecaptcha.execute === "function") {
                resolve();
                return;
              }
              if (Date.now() - startedAt > 15000) {
                reject(new Error("gRecaptcha no está listo"));
                return;
              }
              setTimeout(check, 250);
            };
            check();
          });

          (async () => {
            try {
              await waitForRecaptcha();
              const token = await grecaptcha.execute("6Ld2mMAbAAAAAM9grHC4aJ6pJT1TtvUz04q4Fvjs", { action: "submit_tramite" });
              const formData = new FormData();
              formData.append("tramite", "%s");
              formData.append("token", token);
              formData.append("action", "submit_tramite");
              const response = await fetch("busqueda.php", { method: "POST", body: formData, headers: { "Accept": "application/json" } });
              window.__rnpResult = await response.text();
            } catch (err) {
              window.__rnpResult = JSON.stringify({ error: String(err && err.message ? err.message : err) });
            }
          })();
          return "started";
        })()
    `, tramiteID)

		readCode := `(() => window.__rnpResult || "")()`
		err := chromedp.Run(browserCtx,
			chromedp.Navigate("https://mitramite.renaper.gob.ar/"),
			chromedp.EvaluateAsDevTools(jsCode, &raw),
			// Poll every second until __rnpResult is populated or 25s elapse.
			chromedp.ActionFunc(func(pctx context.Context) error {
				deadline := time.Now().Add(25 * time.Second)
				for time.Now().Before(deadline) {
					if evalErr := chromedp.Evaluate(readCode, &raw).Do(pctx); evalErr != nil {
						return evalErr
					}
					if raw != "" {
						return nil
					}
					select {
					case <-pctx.Done():
						return pctx.Err()
					case <-time.After(time.Second):
					}
				}
				return nil // raw stays empty; handled below
			}),
		)
		logger.Debug("Scraper attempt %d raw result: %q err: %v", attempt, raw, err)
		cleanup()
		cancel()
		if err == nil {
			// JS catch block writes {"error":"..."} on failure.
			var jsErr struct {
				Error string `json:"error"`
			}
			if jsonErr := json.Unmarshal([]byte(raw), &jsErr); jsonErr == nil && jsErr.Error != "" {
				logger.Debug("Scraper attempt %d JS error: %s", attempt, jsErr.Error)
				lastErr = fmt.Errorf("%w: JS error: %s", ErrInvalidResponse, jsErr.Error)
				if attempt < maxAttempts {
					time.Sleep(time.Duration(attempt) * time.Second)
				}
				continue
			}

			normalized, err := normalizeResponse(raw)
			if err != nil {
				lastErr = fmt.Errorf("%w: %w", ErrInvalidResponse, err)
				if attempt < maxAttempts {
					time.Sleep(time.Duration(attempt) * time.Second)
				}
				continue
			}
			if normalized == "" {
				lastErr = fmt.Errorf("%w: empty result after polling", ErrInvalidResponse)
				if attempt < maxAttempts {
					time.Sleep(time.Duration(attempt) * time.Second)
				}
				continue
			}
			if isCaptchaPayload(normalized) {
				if attempt == 1 && !restarted {
					restarted = true
					time.Sleep(2 * time.Second)
					continue
				}
				return "", normalized, restarted, fmt.Errorf("%w: %w", ErrCaptchaBlocked, errors.New(normalized))
			}

			var resp domain.TramiteResponse
			if err := json.Unmarshal([]byte(normalized), &resp); err != nil {
				lastErr = fmt.Errorf("%w: %w", ErrInvalidResponse, err)
				if attempt < maxAttempts {
					time.Sleep(time.Duration(attempt) * time.Second)
				}
				continue
			}

			message := formatTramite(resp)
			return message, normalized, restarted, nil
		}

		if isCaptchaPayload(err.Error()) {
			if attempt == 1 && !restarted {
				restarted = true
				time.Sleep(2 * time.Second)
				continue
			}
			return "", "", restarted, fmt.Errorf("%w: %w", ErrCaptchaBlocked, err)
		}

		lastErr = err
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	if lastErr == nil {
		return "", "", restarted, fmt.Errorf("%w: empty result", ErrUnavailable)
	}
	return "", "", restarted, fmt.Errorf("%w after %d attempts: %w", ErrUnavailable, maxAttempts, lastErr)
}

func isCaptchaPayload(raw string) bool {
	payload := strings.TrimSpace(raw)
	if payload == "" {
		return false
	}
	lower := strings.ToLower(payload)
	if strings.Contains(lower, "recaptcha") || strings.Contains(lower, "captcha") {
		return true
	}
	var data struct {
		Errors map[string]any `json:"errors"`
	}
	if err := json.Unmarshal([]byte(payload), &data); err == nil && data.Errors != nil {
		if status, ok := data.Errors["status"]; ok && fmt.Sprint(status) == "100" {
			return true
		}
		for _, key := range []string{"title", "detail", "tipo"} {
			if value, ok := data.Errors[key]; ok {
				if strings.Contains(strings.ToLower(fmt.Sprint(value)), "captcha") || strings.Contains(strings.ToLower(fmt.Sprint(value)), "recaptcha") {
					return true
				}
			}
		}
	}
	return false
}

func normalizeResponse(raw any) (string, error) {
	switch v := raw.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

func newBrowserContext(ctx context.Context) (context.Context, func()) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("remote-allow-origins", "*"),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("single-process", true),
	)
	if browserPath := strings.TrimSpace(os.Getenv("CHROMIUM_PATH")); browserPath != "" {
		opts = append(opts, chromedp.ExecPath(browserPath))
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	return browserCtx, func() {
		cancelBrowser()
		cancelAlloc()
	}
}

func formatTramite(resp domain.TramiteResponse) string {
	if resp.Data.IDTramite == "" {
		return "No data"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Trámite %s\n", resp.Data.IDTramite)
	fmt.Fprintf(&b, "Descripción: %s\n", resp.Data.DescripcionTramite)
	fmt.Fprintf(&b, "Estado: %s\n", resp.Data.DescripcionUltimoEstado)
	fmt.Fprintf(&b, "Fecha: %s\n", resp.Data.FechaUltimoEstado)
	fmt.Fprintf(&b, "Oficina: %s\n", resp.Data.OficinaRemitente.Descripcion)
	fmt.Fprintf(&b, "Correo: %s\n", resp.Data.Correo)
	if len(resp.Data.TramitesUI) > 0 && len(resp.Data.TramitesUI[0].Historico) > 0 {
		lastEvent := resp.Data.TramitesUI[0].Historico[0]
		fmt.Fprintf(&b, "Último evento: %s (%s)", lastEvent.Evento, lastEvent.Fecha)
	}
	return b.String()
}
