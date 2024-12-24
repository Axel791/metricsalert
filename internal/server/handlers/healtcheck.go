package handlers

import (
	"encoding/json"
	"github.com/Axel791/metricsalert/internal/server/model/api"
	"net/http"
)

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	response := api.HealthCheckResponse{Status: "true"}

	// Устанавливаем заголовок контента
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Кодируем ответ в JSON
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
