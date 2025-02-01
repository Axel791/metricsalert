package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/services"
)

type UpdatesMetricsHandler struct {
	metricService services.Metric
	authService   services.AuthService
	logger        *log.Logger
}

func NewUpdatesMetricsHandler(
	metricService services.Metric,
	authService services.AuthService,
	logger *log.Logger,
) *UpdatesMetricsHandler {
	return &UpdatesMetricsHandler{
		metricService: metricService,
		authService:   authService,
		logger:        logger,
	}
}

func (h *UpdatesMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("HashSHA256")
	if err := h.authService.Validate(token); err != nil {
		h.logger.Infof("error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var input []api.Metrics
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Infof("UpdateMetricHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	if err := h.metricService.BatchMetricsUpdate(ctx, input); err != nil {
		h.logger.Errorf("UpdateMetricHandler: failed to update metrics: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
