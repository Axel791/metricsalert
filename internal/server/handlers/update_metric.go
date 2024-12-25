package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/services"
	log "github.com/sirupsen/logrus"
)

type UpdateMetricHandler struct {
	metricService services.Metric
}

func NewUpdateMetricHandler(metricService services.Metric) *UpdateMetricHandler {
	return &UpdateMetricHandler{metricService: metricService}
}

func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input api.Metrics
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Infof("UpdateMetricHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Infof("input value: %v", input)

	var value interface{}

	switch input.MType {
	case domain.Counter:
		if input.Delta == nil {
			http.Error(w, "missing delta value for counter metric", http.StatusBadRequest)
			return
		}
		value = *input.Delta

	case domain.Gauge:
		if input.Value == nil {
			http.Error(w, "missing value for gauge metric", http.StatusBadRequest)
			return
		}
		value = *input.Value

	default:
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	log.Infof("value: %v", value)

	metricDTO, err := h.metricService.CreateOrUpdateMetric(input.MType, input.ID, value)
	if err != nil {
		log.Infof("UpdateMetricHandler: failed to update metric: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ
	response := api.Metrics{
		ID:    metricDTO.ID,
		MType: metricDTO.MType,
	}

	// Добавляем поле в зависимости от типа метрики
	switch metricDTO.MType {
	case domain.Counter:
		response.Delta = &metricDTO.Delta.Int64
	case domain.Gauge:
		response.Value = &metricDTO.Value.Float64
	}

	log.Infof("API response: %v", response)

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("UpdateMetricHandler: failed to encode response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
