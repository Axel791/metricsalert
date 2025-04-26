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

// UpdateMetricHandler - структура хэндлера обновления метрик [устаревший]
type UpdateMetricHandler struct {
	storage repositories.Store
}

// NewUpdateMetricHandler - конструктор хэндлера обновления метрик [устаревший]
func NewUpdateMetricHandler(storage repositories.Store) *UpdateMetricHandler {
	return &UpdateMetricHandler{storage}
}

// ServeHTTP - обработчик запроса
func (h *UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	if metricType == "" || name == "" || value == "" {
		http.Error(w, "Required parameters are missing", http.StatusNotFound)
		return
	}

	ctx := r.Context()

	switch metricType {
	case Gauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "invalid gauge value", http.StatusBadRequest)
			return
		}
		_, err = h.storage.UpdateGauge(ctx, name, v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case Counter:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		_, err = h.storage.UpdateCounter(ctx, name, v)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
		}
	default:
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length:", value)
	w.WriteHeader(http.StatusOK)
}
