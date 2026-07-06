package bootstrap

import (
	"fmt"
	"net/http"

	"github.com/ikermy/AiR_Logger/v2/pkg/logger"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	if _, err := fmt.Fprint(w, "ok"); err != nil {
		logger.Error("ошибка записи ответа: %v", err)
	}
}
