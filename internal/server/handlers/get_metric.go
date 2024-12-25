package handlers

import (
	"encoding/json"
	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/services"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type GetMetricHandler struct {
	metricService services.Metric
}

func NewGetMetricHandler(metricService services.Metric) *GetMetricHandler {
	return &GetMetricHandler{metricService: metricService}
}

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(1 * time.Second)
	var input api.GetMetric

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Infof("UpdateMetricHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Infof("input value: %v", input)

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

	if metricDTO.MType == domain.Counter && metricDTO.Delta.Int64 != 0 {
		log.Infof("GetMetricHandler: delta metric: %v", metricDTO.Delta)
		apiResponse.Delta = metricDTO.Delta.Int64
	}

	if metricDTO.MType == domain.Gauge && metricDTO.Value.Float64 != 0 {
		log.Infof("GetMetricHandler: value metric: %v", metricDTO.Value)
		apiResponse.Value = metricDTO.Value.Float64
	}

	log.Infof("API response: %v", apiResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiResponse); err != nil {
		log.Printf("GetMetricHandler: error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
