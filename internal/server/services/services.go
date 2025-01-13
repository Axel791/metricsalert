package services

import (
	"context"

	"github.com/Axel791/metricsalert/internal/server/model/dto"
)

type Metric interface {
	GetMetric(ctx context.Context, metricType, name string) (dto.Metrics, error)
	CreateOrUpdateMetric(ctx context.Context, metricType, name string, value interface{}) (dto.Metrics, error)
	GetAllMetric(ctx context.Context) ([]dto.Metrics, error)
}
