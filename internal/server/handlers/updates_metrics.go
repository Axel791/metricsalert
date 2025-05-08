package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/services"
)

// UpdatesMetricsHandler обрабатывает пакетное (batch) обновление метрик.
//
// Клиент отправляет массив JSON‑объектов в формате `api.Metrics`; каждый элемент
// описывает одну метрику (counter или gauge). Хендлер валидирует входные данные
// и делегирует их обработку `MetricService.BatchMetricsUpdate`, позволяя
// применить несколько изменений атомарно.
//
// # Пример запроса
//
// POST /api/metrics HTTP/1.1
// Content-Type: application/json
//
// [
//
//	{"id":"Alloc","type":"gauge","value":6.27},
//	{"id":"PollCount","type":"counter","delta":5}
//
// ]
//
// # Ответы
// | Код | Когда возвращается                                        |
// |-----|-----------------------------------------------------------|
// | 200 | Метрики успешно сохранены                                 |
// | 400 | Невалидный JSON или бизнес‑ошибка сервиса                 |
// | 500 | Неожиданная внутренняя ошибка при сериализации/логике      |
//
// Логи записываются через переданный `*log.Logger`. Экземпляр
// `UpdatesMetricsHandler` потокобезопасен.
type UpdatesMetricsHandler struct {
	metricService services.Metric
	logger        *log.Logger
}

// NewUpdatesMetricsHandler создаёт и инициализирует UpdatesMetricsHandler.
func NewUpdatesMetricsHandler(metricService services.Metric, logger *log.Logger) *UpdatesMetricsHandler {
	return &UpdatesMetricsHandler{
		metricService: metricService,
		logger:        logger,
	}
}

// ServeHTTP реализует http.Handler. Последовательность действий:
//  1. Декодирует входной JSON‑массив в `[]api.Metrics`.
//  2. Передаёт данные в `MetricService.BatchMetricsUpdate`.
//  3. Возвращает HTTP 200 при успехе либо соответствующий код ошибки.
func (h *UpdatesMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input []api.Metrics
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Printf("UpdatesMetricsHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.metricService.BatchMetricsUpdate(r.Context(), input); err != nil {
		h.logger.Printf("UpdatesMetricsHandler: failed to update metrics: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
