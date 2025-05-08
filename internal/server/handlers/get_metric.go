package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/services"

	log "github.com/sirupsen/logrus"
)

// GetMetricHandler обрабатывает извлечение одного значения метрики через HTTP.
//
// Он ожидает тело запроса JSON, соответствующее схеме api.GetMetric
// и отвечает JSON, соответствующим api.Metrics.
//
// # Request example
//
//	POST /value HTTP/1.1
//	Content-Type: application/json
//
//	{"id": "Alloc", "type": "gauge"}
//
// # Successful response example (gauge)
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json
//
//	{"id": "Alloc", "type": "gauge", "value": 6.27}
//
// # Possible error responses
//
//	400 – malformed JSON request body;
//	404 – metric not found;
//	500 – failed to encode response.
//
// Обработчик делегирует бизнес-логику реализации services.Metric
// и регистрирует информационные сообщения через предоставленный *logrus.Logger.
//
// GetMetricHandler is safe for concurrent use and may be reused across requests.
type GetMetricHandler struct {
	metricService services.Metric
	logger        *log.Logger
}

// NewGetMetricHandler создает и возвращает полностью инициализированный экземпляр GetMetricHandler,
// связанный с предоставленной службой Metric и регистратором.
func NewGetMetricHandler(metricService services.Metric, logger *log.Logger) *GetMetricHandler {
	return &GetMetricHandler{
		metricService: metricService,
		logger:        logger,
	}
}

// ServeHTTP implements the http.Handler interface.
//
// Он декодирует входящий запрос, вызывает MetricService.GetMetric, преобразует
// полученный объект домена в HTTP-дружественные DTO и сериализует ответ JSON
// обратно клиенту. Подробную семантику см. в документации GetMetricHandler на уровне типа.
func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input api.GetMetric

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Infof("failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	metricDTO, err := h.metricService.GetMetric(r.Context(), input.MType, input.ID)
	if err != nil {
		h.logger.Infof("error getting metric: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	apiResponse := api.Metrics{
		ID:    metricDTO.ID,
		MType: metricDTO.MType,
	}

	if metricDTO.MType == domain.Counter && metricDTO.Delta.Int64 != 0 {
		apiResponse.Delta = &metricDTO.Delta.Int64
	}

	if metricDTO.MType == domain.Gauge {
		apiResponse.Value = &metricDTO.Value.Float64
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiResponse); err != nil {
		h.logger.Infof("error encoding response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
