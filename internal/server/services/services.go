package services

import (
	"context"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/dto"
)

// Metric - интерфейс сервиса по работе с метриками
type Metric interface {
	GetMetric(ctx context.Context, metricType, name string) (dto.Metrics, error)
	CreateOrUpdateMetric(ctx context.Context, metricAPI api.Metrics) (dto.Metrics, error)
	GetAllMetric(ctx context.Context) ([]dto.Metrics, error)
	BatchMetricsUpdate(ctx context.Context, metrics []api.Metrics) error
}

// SignService - интерфейс подписи
type SignService interface {
	Validate(token string, body []byte) error
	ComputedHash(body []byte) string
}
