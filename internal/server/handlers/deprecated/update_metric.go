package deprecated

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Axel791/metricsalert/internal/server/repositories"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type UpdateMetricHandler struct {
	storage repositories.Store
}

func NewUpdateMetricHandler(storage repositories.Store) *UpdateMetricHandler {
	return &UpdateMetricHandler{storage}
}

func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	if metricType == "" || name == "" || value == "" {
		http.Error(w, "Required parameters are missing", http.StatusNotFound)
		return
	}

	switch metricType {
	case Gauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "invalid gauge value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(name, v)
	case Counter:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(name, v)
	default:
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length:", value)
	w.WriteHeader(http.StatusOK)
}
