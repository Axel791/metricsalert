package services

import "github.com/Axel791/metricsalert/internal/server/model/dto"

type Metric interface {
	GetMetric(metricType, name string) (dto.Metrics, error)
	CreateOrUpdateMetric(metricType, name string, value interface{}) (dto.Metrics, error)
	GetAllMetric() []dto.Metrics
}
