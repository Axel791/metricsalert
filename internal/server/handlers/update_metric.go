package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/services"
)

// UpdateMetricHandler принимает HTTP‑запрос с JSON‑описанием метрики
// и создаёт новую либо обновляет существующую запись.
//
// # Формат запроса
//
// POST /api/metric HTTP/1.1
// Content‑Type: application/json
//
//	{
//	  "id":    "Alloc",
//	  "type":  "gauge",
//	  "value": 6.27
//	}
//
// Для счётчиков используйте поле `delta`, для измеряемых величин — `value`.
// Оба поля необязательны, но одно из них **обязательно** должно быть непустым.
//
// # Формат успешного ответа
//
// HTTP/1.1 200 OK
// Content‑Type: application/json
//
//	{
//	  "id":    "Alloc",
//	  "type":  "gauge",
//	  "value": 6.27
//	}
//
// # Коды ошибок
//
//	400 – некорректный JSON или нарушены бизнес‑правила (например, пустые value/delta);
//	500 – ошибка кодирования ответа.
//
// Все диагностические сообщения пишет в переданный *log.Logger.
// Экземпляр хэндлера потокобезопасен.
type UpdateMetricHandler struct {
	metricService services.Metric
	logger        *log.Logger
}

// NewUpdateMetricHandler конструирует UpdateMetricHandler с внедрённым
// сервисом метрик и логгером.
func NewUpdateMetricHandler(metricService services.Metric, logger *log.Logger) *UpdateMetricHandler {
	return &UpdateMetricHandler{
		metricService: metricService,
		logger:        logger,
	}
}

// ServeHTTP реализует http.Handler.
//
// Шаги выполнения:
//  1. Декодирует тело запроса в api.Metrics.
//  2. Передаёт DTO в MetricService.CreateOrUpdateMetric.
//  3. Возвращает обновлённый объект в JSON.
func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input api.Metrics
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Printf("UpdateMetricHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	metricDTO, err := h.metricService.CreateOrUpdateMetric(r.Context(), input)
	if err != nil {
		h.logger.Printf("UpdateMetricHandler: failed to update metric: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metricDTO); err != nil {
		h.logger.Printf("UpdateMetricHandler: failed to encode response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
