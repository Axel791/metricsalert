package repositories

import (
	"context"
	"fmt"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"gopkg.in/guregu/null.v4"
)

type MetricMapRepositoryHandler struct {
	metrics map[string]domain.Metrics
}

func NewMetricMapRepository() *MetricMapRepositoryHandler {
	return &MetricMapRepositoryHandler{metrics: make(map[string]domain.Metrics)}
}

func (r *MetricMapRepositoryHandler) UpdateGauge(_ context.Context, name string, value float64) (domain.Metrics, error) {
	metric, exists := r.metrics[name]
	if exists && metric.MType == domain.Gauge {
		metric.Value = null.FloatFrom(value)
	} else {
		metric = domain.Metrics{
			ID:    name,
			MType: domain.Gauge,
			Value: null.FloatFrom(value),
		}
	}
	r.metrics[name] = metric
	return metric, nil
}

func (r *MetricMapRepositoryHandler) UpdateCounter(_ context.Context, name string, value int64) (domain.Metrics, error) {
	metric, exists := r.metrics[name]
	if exists && metric.MType == domain.Counter {
		metric.Delta = null.IntFrom(metric.Delta.Int64 + value)
	} else {
		metric = domain.Metrics{
			ID:    name,
			MType: domain.Counter,
			Delta: null.IntFrom(value),
		}
	}
	r.metrics[name] = metric
	return metric, nil
}

func (r *MetricMapRepositoryHandler) GetMetric(_ context.Context, metricsDomain domain.Metrics) (domain.Metrics, error) {
	if metric, exists := r.metrics[metricsDomain.ID]; exists {
		return metric, nil
	}
	return domain.Metrics{}, nil
}

func (r *MetricMapRepositoryHandler) GetAllMetrics(_ context.Context) (map[string]domain.Metrics, error) {
	return r.metrics, nil
}

func (r *MetricMapRepositoryHandler) BatchUpdateMetrics(_ context.Context, metrics []domain.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case domain.Gauge:
			_, err := r.UpdateGauge(context.Background(), metric.ID, metric.Value.Float64)
			if err != nil {
				return fmt.Errorf("failed to update gauge metric %s: %w", metric.ID, err)
			}
		case domain.Counter:
			_, err := r.UpdateCounter(context.Background(), metric.ID, metric.Delta.Int64)
			if err != nil {
				return fmt.Errorf("failed to update counter metric %s: %w", metric.ID, err)
			}
		default:
			return fmt.Errorf("unknown metric type %s for metric %s", metric.MType, metric.ID)
		}
	}
	return nil
}
