package handlers

import (
	"net/http"
)

// NewHealthCheckHandler - обработчик запросов на проверку http сервера доступности
func NewHealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
