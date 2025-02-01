package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/services"
)

type UpdateMetricHandler struct {
	metricService services.Metric
	authService   services.AuthService
	logger        *log.Logger
}

func NewUpdateMetricHandler(
	metricService services.Metric,
	authService services.AuthService,
	logger *log.Logger,
) *UpdateMetricHandler {
	return &UpdateMetricHandler{
		metricService: metricService,
		authService:   authService,
		logger:        logger,
	}
}

func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("HashSHA256")
	if err := h.authService.Validate(token); err != nil {
		h.logger.Infof("error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var input api.Metrics
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Infof("UpdateMetricHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	metricDTO, err := h.metricService.CreateOrUpdateMetric(r.Context(), input)
	if err != nil {
		h.logger.Infof("UpdateMetricHandler: failed to update metric: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metricDTO); err != nil {
		h.logger.Infof("UpdateMetricHandler: failed to encode response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
