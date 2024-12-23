package repositories

import (
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"gopkg.in/guregu/null.v4"
)

type MetricRepository struct {
	metrics map[string]domain.Metrics
}

func NewMetricRepository() *MetricRepository {
	return &MetricRepository{metrics: make(map[string]domain.Metrics)}
}

func (r *MetricRepository) UpdateGauge(name string, value float64) domain.Metrics {
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
	return metric
}

func (r *MetricRepository) UpdateCounter(name string, value int64) domain.Metrics {
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
	return metric
}

func (r *MetricRepository) GetMetric(metricsDomain domain.Metrics) domain.Metrics {
	if metric, exists := r.metrics[metricsDomain.ID]; exists {
		return metric
	}
	return domain.Metrics{}
}

func (r *MetricRepository) GetAllMetrics() map[string]domain.Metrics {
	return r.metrics
}
