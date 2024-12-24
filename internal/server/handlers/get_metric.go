package handlers

import (
	"encoding/json"
	"time"

	"net/http"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/services"

	log "github.com/sirupsen/logrus"
)

type GetMetricHandler struct {
	metricService services.Metric
}

func NewGetMetricHandler(metricService services.Metric) *GetMetricHandler {
	return &GetMetricHandler{metricService: metricService}
}

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second)
	var input api.GetMetric

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Infof("UpdateMetricHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Infof("UpdateMetricHandler: received request for metric: %v", input)

	metricDTO, err := h.metricService.GetMetric(input.MType, input.ID)
	if err != nil {
		log.Printf("GetMetricHandler: error getting metric: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	apiResponse := api.Metrics{
		ID:    metricDTO.ID,
		MType: metricDTO.MType,
	}

	if metricDTO.MType == domain.Counter {
		apiResponse.Delta = metricDTO.Delta.Int64
	} else if metricDTO.MType == domain.Gauge {
		apiResponse.Value = metricDTO.Value.Float64
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiResponse); err != nil {
		log.Printf("GetMetricHandler: error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
