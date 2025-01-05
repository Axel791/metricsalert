package repositories

import "github.com/Axel791/metricsalert/internal/server/model/domain"

type Store interface {
	UpdateGauge(name string, value float64) domain.Metrics
	UpdateCounter(name string, value int64) domain.Metrics
	GetMetric(metricsDomain domain.Metrics) domain.Metrics
	GetAllMetrics() map[string]domain.Metrics
}

type FileStore interface {
	UpdateGauge(name string, value float64) domain.Metrics
	UpdateCounter(name string, value int64) domain.Metrics
	GetMetric(metricsDomain domain.Metrics) domain.Metrics
	GetAllMetrics() map[string]domain.Metrics
	Load() error
	SaveToFile() error
}
