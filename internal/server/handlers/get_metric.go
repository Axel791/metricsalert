package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/services"

	log "github.com/sirupsen/logrus"
)

type GetMetricHandler struct {
	metricService services.Metric
	logger        *log.Logger
}

func NewGetMetricHandler(metricService services.Metric, logger *log.Logger) *GetMetricHandler {
	return &GetMetricHandler{
		metricService: metricService,
		logger:        logger,
	}
}

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input api.GetMetric

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Infof("failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	metricDTO, err := h.metricService.GetMetric(input.MType, input.ID)
	if err != nil {
		h.logger.Infof("Error getting metric: %v", err)
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
		h.logger.Infof("Error encoding response: %v", err)
		errorText := http.StatusText(http.StatusInternalServerError)
		http.Error(w, errorText, http.StatusInternalServerError)
		return
	}
}
