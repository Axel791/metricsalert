package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/model/dto"
	"github.com/Axel791/metricsalert/internal/server/repositories"
)

type MetricsService struct {
	store repositories.Store
}

func NewMetricsService(store repositories.Store) *MetricsService {
	return &MetricsService{
		store: store,
	}
}

func (ms *MetricsService) GetMetric(ctx context.Context, metricType, name string) (dto.Metrics, error) {
	var metricsDTO dto.Metrics

	metric := domain.Metrics{
		ID:    name,
		MType: metricType,
	}

	if err := metric.ValidateMetricID(); err != nil {
		return metricsDTO, err
	}

	if err := metric.ValidateMetricsType(); err != nil {
		return metricsDTO, err
	}

	metricsDomain, err := ms.store.GetMetric(ctx, metric)
	if err != nil {
		return metricsDTO, errors.New(fmt.Sprintf("GetMetric: error getting metric domain: %v", err))
	}

	if metricsDomain.ID == "" {
		return metricsDTO, errors.New("metric not found")
	}

	metricsDTO = dto.Metrics{
		ID:    metricsDomain.ID,
		MType: metricsDomain.MType,
		Delta: metricsDomain.Delta,
		Value: metricsDomain.Value,
	}

	return metricsDTO, nil
}

func (ms *MetricsService) CreateOrUpdateMetric(
	ctx context.Context, metricType, name string, value interface{},
) (dto.Metrics, error) {
	var metricsDTO dto.Metrics

	metric := domain.Metrics{
		ID:    name,
		MType: metricType,
	}

	if err := metric.ValidateMetricID(); err != nil {
		return metricsDTO, err
	}

	if err := metric.ValidateMetricsType(); err != nil {
		return metricsDTO, err
	}

	if err := metric.SetMetricValue(value); err != nil {
		return metricsDTO, err
	}

	var updatedMetric domain.Metrics
	var err error

	switch metric.MType {
	case domain.Gauge:
		updatedMetric, err = ms.store.UpdateGauge(ctx, metric.ID, metric.Value.Float64)
		if err != nil {
			return metricsDTO, errors.New(fmt.Sprintf("UpdateMetric: error updating metric: %v", err))
		}
	case domain.Counter:
		updatedMetric, err = ms.store.UpdateCounter(ctx, metric.ID, metric.Delta.Int64)
		if err != nil {
			return metricsDTO, errors.New(fmt.Sprintf("UpdateMetric: error updating metric: %v", err))
		}
	default:
		return metricsDTO, errors.New("unsupported metric type")
	}

	metricsDTO = dto.Metrics{
		ID:    updatedMetric.ID,
		MType: updatedMetric.MType,
		Delta: updatedMetric.Delta,
		Value: updatedMetric.Value,
	}

	return metricsDTO, nil
}

func (ms *MetricsService) GetAllMetric(ctx context.Context) ([]dto.Metrics, error) {
	var metricsDTO []dto.Metrics
	metrics, err := ms.store.GetAllMetrics(ctx)
	if err != nil {
		return metricsDTO, errors.New(fmt.Sprintf("GetAllMetrics: error getting metrics: %v", err))
	}

	for _, metric := range metrics {
		metricsDTO = append(
			metricsDTO,
			dto.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: metric.Delta,
				Value: metric.Value,
			},
		)
	}
	return metricsDTO, nil
}
