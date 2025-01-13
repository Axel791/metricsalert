package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/services"
)

type UpdateMetricHandler struct {
	metricService services.Metric
	logger        *log.Logger
}

func NewUpdateMetricHandler(metricService services.Metric, logger *log.Logger) *UpdateMetricHandler {
	return &UpdateMetricHandler{
		metricService: metricService,
		logger:        logger,
	}
}

func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input api.Metrics
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Infof("UpdateMetricHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

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

	metricDTO, err := h.metricService.CreateOrUpdateMetric(r.Context(), input.MType, input.ID, value)
	if err != nil {
		h.logger.Infof("UpdateMetricHandler: failed to update metric: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := api.Metrics{
		ID:    metricDTO.ID,
		MType: metricDTO.MType,
	}

	switch metricDTO.MType {
	case domain.Counter:
		response.Delta = &metricDTO.Delta.Int64
	case domain.Gauge:
		response.Value = &metricDTO.Value.Float64
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("UpdateMetricHandler: failed to encode response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
