package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Axel791/metricsalert/internal/server/repositories"
	"github.com/go-chi/chi/v5"
)

type GetMetricHandler struct {
	storage repositories.Store
}

func NewGetMetricHandler(storage repositories.Store) *GetMetricHandler {
	return &GetMetricHandler{storage}
}

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	name := chi.URLParam(r, "name")

	if name == "" {
		log.Printf("GetMetricHandler: missing metric name")
		http.Error(w, "invalid metric name", http.StatusNotFound)
		return
	}

	if metricType != Counter && metricType != Gauge {
		log.Printf("GetMetricHandler: invalid metric type: %s", metricType)
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	value := h.storage.GetMetric(name)
	if value == nil {
		log.Printf("GetMetricHandler: metric not found: %s (type: %s)", name, metricType)
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}

	var valueStr string
	switch v := value.(type) {
	case string:
		valueStr = v
	case float64:
		valueStr = strconv.FormatFloat(v, 'g', -1, 64)
	case int64:
		valueStr = strconv.FormatInt(v, 10)
	default:
		valueStr = strconv.FormatFloat(float64(v.(int)), 'g', -1, 64)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(valueStr)))
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte(valueStr))
	if err != nil {
		log.Printf("GetMetricHandler: failed to write response for metric %s (type: %s): %v", name, metricType, err)
		http.Error(w, "invalid metric", http.StatusInternalServerError)
		return
	}
}
