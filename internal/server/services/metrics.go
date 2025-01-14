package services

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/guregu/null.v4"

	"github.com/Axel791/metricsalert/internal/server/model/api"
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

// GetMetric - получение метрики
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

// CreateOrUpdateMetric - Обновление метрик
func (ms *MetricsService) CreateOrUpdateMetric(
	ctx context.Context,
	metricApi api.Metrics,
) (dto.Metrics, error) {

	var metricsDTO dto.Metrics

	if metricApi.ID == "" {
		return metricsDTO, errors.New("metric ID is required")
	}

	if metricApi.MType != domain.Counter && metricApi.MType != domain.Gauge {
		return metricsDTO, fmt.Errorf("invalid metric type: %s", metricApi.MType)
	}

	switch metricApi.MType {
	case domain.Counter:
		if metricApi.Delta == nil {
			return metricsDTO, fmt.Errorf("missing delta value for counter metric '%s'", metricApi.ID)
		}
	case domain.Gauge:
		if metricApi.Value == nil {
			return metricsDTO, fmt.Errorf("missing value for gauge metric '%s'", metricApi.ID)
		}
	}

	metric := domain.Metrics{
		ID:    metricApi.ID,
		MType: metricApi.MType,
		Delta: null.Int{},
		Value: null.Float{},
	}

	if metricApi.MType == domain.Counter {
		if err := metric.SetMetricValue(*metricApi.Delta); err != nil {
			return metricsDTO, err
		}
	} else {
		if err := metric.SetMetricValue(*metricApi.Value); err != nil {
			return metricsDTO, err
		}
	}

	var updatedMetric domain.Metrics
	var err error

	switch metric.MType {
	case domain.Counter:
		updatedMetric, err = ms.store.UpdateCounter(ctx, metric.ID, metric.Delta.Int64)
		if err != nil {
			return metricsDTO, fmt.Errorf("UpdateMetric: error updating counter: %v", err)
		}
	case domain.Gauge:
		updatedMetric, err = ms.store.UpdateGauge(ctx, metric.ID, metric.Value.Float64)
		if err != nil {
			return metricsDTO, fmt.Errorf("UpdateMetric: error updating gauge: %v", err)
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

// GetAllMetric - Получение всех метрик
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

// BatchMetricsUpdate - валидирует входные метрики, конвертирует в доменные
func (ms *MetricsService) BatchMetricsUpdate(ctx context.Context, metrics []api.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	domainMetrics := make([]domain.Metrics, 0, len(metrics))

	for _, m := range metrics {
		d := domain.Metrics{
			ID:    m.ID,
			MType: m.MType,
			Value: null.Float{},
			Delta: null.Int{},
		}

		if err := d.ValidateMetricsType(); err != nil {
			return fmt.Errorf("metric '%s': %w", m.ID, err)
		}

		if err := d.ValidateMetricID(); err != nil {
			return fmt.Errorf("metric '%s': %w", m.ID, err)
		}

		switch d.MType {
		case domain.Counter:
			if m.Delta == nil {
				return errors.New(
					fmt.Sprintf("BatchMetricsUpdate: error metric '%s': delta is required for counter", m.ID),
				)
			}
			if err := d.SetMetricValue(*m.Delta); err != nil {
				return errors.New(fmt.Sprintf("BatchMetricsUpdate: error metric '%s': %v", m.ID, err))
			}
		case domain.Gauge:
			if m.Value == nil {
				return errors.New(
					fmt.Sprintf("BatchMetricsUpdate: error metric '%s': value is required for gauge", m.ID),
				)
			}
			if err := d.SetMetricValue(*m.Value); err != nil {
				return errors.New(fmt.Sprintf("BatchMetricsUpdate: error metric '%s': %v", m.ID, err))
			}
		default:
			return errors.New(
				fmt.Sprintf(
					"BatchMetricsUpdate: error metric '%s': unsupported metric type '%s'", m.ID, m.MType,
				),
			)
		}

		domainMetrics = append(domainMetrics, d)
	}

	if err := ms.store.BatchUpdateMetrics(ctx, domainMetrics); err != nil {
		return errors.New(fmt.Sprintf("BatchMetricsUpdate: error batch update failed: %v", err))
	}

	return nil
}
