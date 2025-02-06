package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	h.logger.Infof("token update: %s", token)

	validBody, err := h.validateBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	normalizedBody, err := h.normalizeBody(validBody)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = h.authService.Validate(token, normalizedBody); err != nil {
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
	computedHash := h.authService.ComputedHash(validBody)

	w.Header().Set("HashSHA256", computedHash)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metricDTO); err != nil {
		h.logger.Infof("UpdateMetricHandler: failed to encode response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *UpdateMetricHandler) validateBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	if len(body) == 0 {
		return nil, nil
	}
	return body, nil
}

func (h *UpdateMetricHandler) normalizeBody(body []byte) ([]byte, error) {
	var metricsList []api.Metrics
	if err := json.Unmarshal(body, &metricsList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}
	normalizedBody, err := json.Marshal(metricsList)

	h.logger.Infof("normalized body update metric: %s", string(normalizedBody))

	if err != nil {
		return nil, fmt.Errorf("failed to marshal normalized body: %w", err)
	}
	return normalizedBody, nil
}
