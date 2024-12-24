package services

import (
	"errors"

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

func (ms *MetricsService) GetMetric(metricType, name string) (dto.Metrics, error) {
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

	metricsDomain := ms.store.GetMetric(metric)
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

func (ms *MetricsService) CreateOrUpdateMetric(metricType, name string, value interface{}) (dto.Metrics, error) {
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

	switch metric.MType {
	case domain.Gauge:
		updatedMetric = ms.store.UpdateGauge(metric.ID, metric.Value.Float64)
	case domain.Counter:
		updatedMetric = ms.store.UpdateCounter(metric.ID, metric.Delta.Int64)
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

func (ms *MetricsService) GetAllMetric() []dto.Metrics {
	var metricsDTO []dto.Metrics
	metrics := ms.store.GetAllMetrics()

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
	return metricsDTO
}
