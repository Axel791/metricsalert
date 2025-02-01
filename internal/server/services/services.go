package services

import (
	"context"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/dto"
)

type Metric interface {
	GetMetric(ctx context.Context, metricType, name string) (dto.Metrics, error)
	CreateOrUpdateMetric(ctx context.Context, metricAPI api.Metrics) (dto.Metrics, error)
	GetAllMetric(ctx context.Context) ([]dto.Metrics, error)
	BatchMetricsUpdate(ctx context.Context, metrics []api.Metrics) error
}
