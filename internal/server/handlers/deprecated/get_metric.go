package deprecated

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/repositories"
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

	metric := domain.Metrics{
		ID:    name,
		MType: metricType,
	}

	if err := metric.ValidateMetricID(); err != nil {
		http.Error(w, "invalid metric name", http.StatusBadRequest)
		return
	}

	if err := metric.ValidateMetricsType(); err != nil {
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	value := h.storage.GetMetric(metric)

	if value.ID == "" {
		log.Printf("GetMetricHandler: metric not found: %s (type: %s)", name, metricType)
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}

	var valueStr string
	switch value.MType {
	case Gauge:
		if value.Value.Valid {
			valueStr = strconv.FormatFloat(value.Value.Float64, 'g', -1, 64)
		} else {
			valueStr = "null"
		}
	case Counter:
		if value.Delta.Valid {
			valueStr = strconv.FormatInt(value.Delta.Int64, 10)
		} else {
			valueStr = "null"
		}
	default:
		log.Printf("unknown metric type: %s", value.MType)
		valueStr = "unknown"
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(valueStr)))
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte(valueStr))
	if err != nil {
		log.Printf(
			"GetMetricHandler: invalid metric %s (type: %s): %v", name, metricType, err,
		)
		http.Error(w, "invalid metric", http.StatusInternalServerError)
		return
	}
}
