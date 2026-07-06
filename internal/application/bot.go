package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"renaper_mitramite/internal/domain"
	"renaper_mitramite/internal/infrastructure/config"
	"renaper_mitramite/internal/infrastructure/scraper"
)

var translations = map[string]map[string]string{
	"ru": {
		"welcome":            "👋 Привет! Я могу проверить статус заявления по номеру trámite.\n\nВыберите действие:",
		"btnCheck":           "🔎 Проверить статус",
		"btnLang":            "🌐 Изменить язык",
		"btnRaw":             "📄 Raw JSON",
		"btnRu":              "🇷🇺 Русский",
		"btnEn":              "🇺🇸 English",
		"btnEs":              "🇪🇸 Español",
		"prompt":             "📝 Отправьте номер заявления (например, 00123456789)",
		"checking":           "⏳ Проверяю статус заявления %s…",
		"timeout":            "⏰ Не удалось получить ответ от сервера в течение %s. Попробуйте позже.",
		"success":            "✅ Статус заявления %s\n\n%s",
		"rawBtn":             "📄 Показать raw JSON",
		"noRaw":              "📭 Raw JSON пока недоступен.",
		"langChanged":        "🌐 Язык изменён на русский.",
		"help":               "🧭 Команды:\n/start — меню\n/help — помощь\n/check — проверить статус",
		"invalid":            "⚠️ Пожалуйста, отправьте корректный номер заявления.",
		"cancel":             "❌ Ожидание номера отменено.",
		"recaptcha":          "🛡️ Сервис временно не может обработать запрос из-за reCAPTCHA.",
		"noData":             "📭 По данному номеру данных пока нет.",
		"serviceUnavailable": "⚠️ Сервис временно недоступен. Попробуйте позже.",
		"captchaRetrying":    "⏳ Пытаюсь обойти временную блокировку.",
		"captchaFallback":    "⚠️ Не удалось обойти защиту. Откройте ссылку, пройдите проверку и затем повторите запрос:\n%s",
	},
	"en": {
		"welcome":            "👋 Hello! I can check the status of your application by trámite number.\n\nChoose an action:",
		"btnCheck":           "🔎 Check status",
		"btnLang":            "🌐 Change language",
		"btnRaw":             "📄 Raw JSON",
		"btnRu":              "🇷🇺 Русский",
		"btnEn":              "🇺🇸 English",
		"btnEs":              "🇪🇸 Español",
		"prompt":             "📝 Send the application number (for example, 00123456789)",
		"checking":           "⏳ Checking status for %s…",
		"timeout":            "⏰ The server did not respond within %s. Please try again later.",
		"success":            "✅ Status for application %s\n\n%s",
		"rawBtn":             "📄 Show raw JSON",
		"noRaw":              "📭 Raw JSON is not available yet.",
		"langChanged":        "🌐 Language changed to English.",
		"help":               "🧭 Commands:\n/start — menu\n/help — help\n/check — check status",
		"invalid":            "⚠️ Please send a valid application number.",
		"cancel":             "❌ Waiting for the number was cancelled.",
		"recaptcha":          "🛡️ The service is temporarily blocked by reCAPTCHA.",
		"noData":             "📭 There is no data for this number yet.",
		"serviceUnavailable": "⚠️ The service is temporarily unavailable. Please try again later.",
		"captchaRetrying":    "⏳ I am trying to bypass the temporary block.",
		"captchaFallback":    "⚠️ I could not bypass the protection. Open the link, complete the check, and then try again:\n%s",
	},
	"es": {
		"welcome":            "👋 ¡Hola! Puedo verificar el estado del trámite por número.\n\nElige una acción:",
		"btnCheck":           "🔎 Verificar estado",
		"btnLang":            "🌐 Cambiar idioma",
		"btnRaw":             "📄 Raw JSON",
		"btnRu":              "🇷🇺 Русский",
		"btnEn":              "🇺🇸 English",
		"btnEs":              "🇪🇸 Español",
		"prompt":             "📝 Envía el número del trámite (por ejemplo, 00123456789)",
		"checking":           "⏳ Comprobando el estado de %s…",
		"timeout":            "⏰ El servidor no respondió en %s. Inténtalo más tarde.",
		"success":            "✅ Estado del trámite %s\n\n%s",
		"rawBtn":             "📄 Mostrar raw JSON",
		"noRaw":              "📭 El raw JSON aún no está disponible.",
		"langChanged":        "🌐 Idioma cambiado a español.",
		"help":               "🧭 Comandos:\n/start — menú\n/help — ayuda\n/check — verificar estado",
		"invalid":            "⚠️ Por favor, envía un número de trámite válido.",
		"cancel":             "❌ Se canceló la espera del número.",
		"recaptcha":          "🛡️ El servicio está temporalmente bloqueado por reCAPTCHA.",
		"noData":             "📭 Todavía no hay datos para este número.",
		"serviceUnavailable": "⚠️ El servicio está temporalmente no disponible. Inténtalo más tarde.",
		"captchaRetrying":    "⏳ Estoy intentando sortear la restricción.",
		"captchaFallback":    "⚠️ No pude sortear la protección. Abre el enlace, completa la verificación y luego vuelve a intentarlo:\n%s",
	},
}

type TramiteChecker interface {
	CheckTramite(tramiteID string) (string, string, bool, error)
}

type Reply struct {
	Text   string
	Markup *Markup
}

type Markup struct {
	InlineKeyboard [][]InlineButton
}

type InlineButton struct {
	Text string
	Data string
}

type Bot struct {
	checker   TramiteChecker
	states    map[int]domain.UserState
	rawByUser map[int]string
	config    config.Config
	mu        sync.RWMutex
}

func NewBot(checker TramiteChecker, cfg config.Config) *Bot {
	return &Bot{
		checker:   checker,
		states:    make(map[int]domain.UserState),
		rawByUser: make(map[int]string),
		config:    cfg,
	}
}

func (u *Bot) HandleCommand(userID int, command string) (Reply, error) {
	switch command {
	case "start":
		return u.showMainMenu(userID), nil
	case "help":
		return u.withMainMenu(userID, Reply{Text: u.t(userID, "help")}), nil
	case "check":
		u.setWaiting(userID, true)
		return u.withMainMenu(userID, Reply{Text: u.t(userID, "prompt")}), nil
	default:
		return Reply{}, nil
	}
}

func (u *Bot) HandleCallback(userID int, data string) (Reply, error) {
	switch data {
	case "check":
		u.setWaiting(userID, true)
		return u.withMainMenu(userID, Reply{Text: u.t(userID, "prompt")}), nil
	case "raw":
		if raw, ok := u.rawByUser[userID]; ok && raw != "" {
			return u.withMainMenu(userID, Reply{Text: raw}), nil
		}
		return u.withMainMenu(userID, Reply{Text: u.t(userID, "noRaw")}), nil
	case "lang":
		return u.withMainMenu(userID, u.showLanguageMenu(userID)), nil
	case "lang_ru":
		u.setLocale(userID, "ru")
		return u.withMainMenu(userID, Reply{Text: u.t(userID, "langChanged")}), nil
	case "lang_en":
		u.setLocale(userID, "en")
		return u.withMainMenu(userID, Reply{Text: u.t(userID, "langChanged")}), nil
	case "lang_es":
		u.setLocale(userID, "es")
		return u.withMainMenu(userID, Reply{Text: u.t(userID, "langChanged")}), nil
	default:
		return Reply{}, nil
	}
}

func (u *Bot) HandleText(userID int, text string) (Reply, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Reply{}, nil
	}

	u.mu.RLock()
	state := u.states[userID]
	u.mu.RUnlock()
	if state.WaitingForTramite {
		u.setWaiting(userID, false)
		if strings.EqualFold(text, "/cancel") {
			return u.withMainMenu(userID, Reply{Text: u.t(userID, "cancel")}), nil
		}
		if !isTramiteID(text) {
			return u.withMainMenu(userID, Reply{Text: u.t(userID, "invalid")}), nil
		}
		reply, err := u.checkAndReply(userID, text)
		if err != nil {
			return Reply{}, err
		}
		return u.withMainMenu(userID, reply), nil
	}

	if strings.HasPrefix(text, "/") {
		return Reply{}, nil
	}

	if isTramiteID(text) {
		reply, err := u.checkAndReply(userID, text)
		if err != nil {
			return Reply{}, err
		}
		return u.withMainMenu(userID, reply), nil
	}

	return u.withMainMenu(userID, Reply{Text: u.t(userID, "invalid")}), nil
}

func (u *Bot) showMainMenu(chatID int) Reply {
	return Reply{Text: u.t(chatID, "welcome"), Markup: u.mainMarkup(chatID)}
}

func (u *Bot) showLanguageMenu(chatID int) Reply {
	markup := &Markup{InlineKeyboard: [][]InlineButton{{
		{Text: u.t(chatID, "btnRu"), Data: "lang_ru"},
		{Text: u.t(chatID, "btnEn"), Data: "lang_en"},
		{Text: u.t(chatID, "btnEs"), Data: "lang_es"},
	}}}
	return Reply{Text: u.t(chatID, "btnLang"), Markup: markup}
}

func (u *Bot) mainMarkup(chatID int) *Markup {
	return &Markup{InlineKeyboard: [][]InlineButton{{
		{Text: u.t(chatID, "btnCheck"), Data: "check"},
		{Text: u.t(chatID, "btnLang"), Data: "lang"},
	}}}
}

func (u *Bot) checkAndReply(chatID int, tramiteID string) (Reply, error) {
	message, rawJSON, restarted, err := u.checker.CheckTramite(tramiteID)
	if err != nil {
		if strings.Contains(err.Error(), context.DeadlineExceeded.Error()) || strings.Contains(strings.ToLower(err.Error()), "deadline exceeded") {
			return Reply{Text: fmt.Sprintf(u.t(chatID, "timeout"), 10*time.Second)}, nil
		}
		if errors.Is(err, scraper.ErrCaptchaBlocked) {
			return Reply{Text: fmt.Sprintf("%s\n\n%s", u.t(chatID, "captchaRetrying"), fmt.Sprintf(u.t(chatID, "captchaFallback"), captchaLink))}, nil
		}
		if errors.Is(err, scraper.ErrUnavailable) {
			if isCaptchaError(err.Error()) {
				return Reply{Text: u.t(chatID, "recaptcha")}, nil
			}
			return Reply{Text: u.t(chatID, "serviceUnavailable")}, nil
		}
		return Reply{Text: fmt.Sprintf("Error: %v", err)}, nil
	}

	if statusText, ok := u.classifyResponse(chatID, message, rawJSON); ok {
		if restarted {
			return Reply{Text: fmt.Sprintf("%s\n\n%s", u.t(chatID, "captchaRetrying"), statusText)}, nil
		}
		return Reply{Text: statusText}, nil
	}

	if rawJSON != "" {
		u.mu.Lock()
		u.rawByUser[chatID] = rawJSON
		u.mu.Unlock()
	}

	markup := &Markup{InlineKeyboard: [][]InlineButton{{
		{Text: u.t(chatID, "rawBtn"), Data: "raw"},
	}}}
	text := fmt.Sprintf(u.t(chatID, "success"), tramiteID, message)
	if restarted {
		text = fmt.Sprintf("%s\n\n%s", u.t(chatID, "captchaRetrying"), text)
	}
	return Reply{Text: text, Markup: markup}, nil
}

func (u *Bot) classifyResponse(chatID int, message, rawJSON string) (string, bool) {
	trimmed := strings.TrimSpace(message)
	if strings.EqualFold(trimmed, "No data") || strings.EqualFold(trimmed, "no data") || strings.Contains(strings.ToLower(trimmed), "no data") {
		return u.t(chatID, "noData"), true
	}
	if isCaptchaPayload(rawJSON) || isCaptchaError(trimmed) {
		return u.t(chatID, "recaptcha"), true
	}
	return "", false
}

func isCaptchaPayload(rawJSON string) bool {
	payload := strings.TrimSpace(rawJSON)
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

func isCaptchaError(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	return strings.Contains(lower, "recaptcha") || strings.Contains(lower, "captcha")
}

func (u *Bot) t(chatID int, key string) string {
	u.mu.RLock()
	locale := u.locale(chatID)
	u.mu.RUnlock()
	if texts, ok := translations[locale]; ok {
		if s, ok := texts[key]; ok {
			return s
		}
	}
	return translations[u.config.Lang][key]
}

func (u *Bot) locale(chatID int) string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if st, ok := u.states[chatID]; ok && st.Locale != "" {
		return st.Locale
	}
	return u.config.Lang
}

func (u *Bot) setWaiting(chatID int, waiting bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	st := u.states[chatID]
	st.WaitingForTramite = waiting
	u.states[chatID] = st
}

func (u *Bot) setLocale(chatID int, locale string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	st := u.states[chatID]
	st.Locale = locale
	u.states[chatID] = st
}

func isTramiteID(text string) bool {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return false
	}
	for _, r := range clean {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (u *Bot) withMainMenu(chatID int, reply Reply) Reply {
	if reply.Text == "" {
		return reply
	}
	menu := u.mainMarkup(chatID)
	if reply.Markup == nil {
		reply.Markup = menu
		return reply
	}
	rows := make([][]InlineButton, 0, len(reply.Markup.InlineKeyboard)+len(menu.InlineKeyboard))
	rows = append(rows, reply.Markup.InlineKeyboard...)
	rows = append(rows, menu.InlineKeyboard...)
	reply.Markup = &Markup{InlineKeyboard: rows}
	return reply
}

var captchaLink = "https://mitramite.renaper.gob.ar/"
