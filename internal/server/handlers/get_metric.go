package handlers

import (
	"fmt"
	"github.com/Axel791/metricsalert/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type GetMetricHandler struct {
	storage storage.Store
}

func NewGetMetricHandler(storage storage.Store) *GetMetricHandler {
	return &GetMetricHandler{storage}
}

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	name := chi.URLParam(r, "name")

	if name == "" {
		http.Error(w, "invalid metric name", http.StatusNotFound)
		return
	}

	var value interface{}

	if metricType != Counter && metricType != Gauge {
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}
	value = h.storage.GetMetric(name)

	if value == nil {
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}

	fmt.Println(value)

	var valueStr string

	switch v := value.(type) {
	case string:
		valueStr = v
	case float64:
		valueStr = strconv.FormatFloat(v, 'g', -1, 64)
	case int64:
		valueStr = fmt.Sprintf("%d", v)
	default:
		valueStr = fmt.Sprintf("%v", v)
	}
	fmt.Println(valueStr)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(valueStr)))
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(valueStr))
	if err != nil {
		http.Error(w, "invalid metric", http.StatusInternalServerError)
	}
}
