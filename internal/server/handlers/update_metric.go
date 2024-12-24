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

	var value interface{}
	if input.MType == domain.Counter {
		value = input.Delta
	} else {
		value = input.Value
	}

	metricDTO, err := h.metricService.CreateOrUpdateMetric(input.MType, input.ID, value)
	if err != nil {
		log.Infof("UpdateMetricHandler: failed to update metric: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response := api.Metrics{
		ID:    metricDTO.ID,
		MType: metricDTO.MType,
	}

	if metricDTO.MType == domain.Counter {
		response.Delta = metricDTO.Delta.Int64
	} else if metricDTO.MType == domain.Gauge {
		response.Value = metricDTO.Value.Float64
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("UpdateMetricHandler: failed to encode response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
