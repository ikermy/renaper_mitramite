package bootstrap

import (
	"fmt"
	"log"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	if _, err := fmt.Fprint(w, "ok"); err != nil {
		log.Println("ошибка записи ответа:", err)
	}
}
