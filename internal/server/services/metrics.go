package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/model/dto"
	"github.com/Axel791/metricsalert/internal/server/repositories"
)

// MetricsService - сервис, работающий с метриками
type MetricsService struct {
	store repositories.Store
}

func NewMetricsService(store repositories.Store) *MetricsService {
	return &MetricsService{store: store}
}

// GetMetric - получение метрики по (type, name).
func (ms *MetricsService) GetMetric(ctx context.Context, metricType, name string) (dto.Metrics, error) {
	var metricsDTO dto.Metrics

	metric := domain.Metrics{
		Name:  name,
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
		return metricsDTO, fmt.Errorf("GetMetric: error getting metric domain: %v", err)
	}

	if metricsDomain.Name == "" {
		return metricsDTO, errors.New("metric not found")
	}

	metricsDTO = dto.Metrics{
		ID:    metricsDomain.Name,
		MType: metricsDomain.MType,
		Delta: metricsDomain.Delta,
		Value: metricsDomain.Value,
	}

	return metricsDTO, nil
}

// CreateOrUpdateMetric - создаёт или обновляет метрику.
func (ms *MetricsService) CreateOrUpdateMetric(ctx context.Context, metricApi api.Metrics) (dto.Metrics, error) {
	var metricsDTO dto.Metrics

	if metricApi.ID == "" {
		return metricsDTO, errors.New("metric name (ID) is required")
	}

	if metricApi.MType != domain.Counter && metricApi.MType != domain.Gauge {
		return metricsDTO, fmt.Errorf("invalid metric type: %s", metricApi.MType)
	}

	switch metricApi.MType {
	case domain.Counter:
		if metricApi.Delta == nil {
			return metricsDTO, fmt.Errorf("missing delta for counter '%s'", metricApi.ID)
		}
	case domain.Gauge:
		if metricApi.Value == nil {
			return metricsDTO, fmt.Errorf("missing value for gauge '%s'", metricApi.ID)
		}
	}

	metric := domain.Metrics{
		Name:  metricApi.ID,
		MType: metricApi.MType,
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
		updatedMetric, err = ms.store.UpdateCounter(ctx, metric.Name, metric.Delta.Int64)
		if err != nil {
			return metricsDTO, fmt.Errorf("UpdateMetric (counter): %v", err)
		}
	case domain.Gauge:
		updatedMetric, err = ms.store.UpdateGauge(ctx, metric.Name, metric.Value.Float64)
		if err != nil {
			return metricsDTO, fmt.Errorf("UpdateMetric (gauge): %v", err)
		}
	default:
		return metricsDTO, fmt.Errorf("unsupported metric type: %s", metric.MType)
	}

	metricsDTO = dto.Metrics{
		ID:    updatedMetric.Name, // !!!
		MType: updatedMetric.MType,
		Delta: updatedMetric.Delta,
		Value: updatedMetric.Value,
	}

	return metricsDTO, nil
}

// GetAllMetric - получение всех метрик
func (ms *MetricsService) GetAllMetric(ctx context.Context) ([]dto.Metrics, error) {
	var metricsDTO []dto.Metrics

	metricsMap, err := ms.store.GetAllMetrics(ctx)
	if err != nil {
		return metricsDTO, fmt.Errorf("GetAllMetrics: error getting from store: %v", err)
	}

	for _, domainM := range metricsMap {
		metricsDTO = append(metricsDTO, dto.Metrics{
			ID:    domainM.Name, // DTO.ID = domainM.Name
			MType: domainM.MType,
			Delta: domainM.Delta,
			Value: domainM.Value,
		})
	}
	return metricsDTO, nil
}

// BatchMetricsUpdate - батчевое обновление
func (ms *MetricsService) BatchMetricsUpdate(ctx context.Context, metrics []api.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	domainMetrics := make([]domain.Metrics, 0, len(metrics))

	for _, m := range metrics {
		if m.ID == "" {
			return fmt.Errorf("BatchMetricsUpdate: metric name (ID) is empty")
		}

		d := domain.Metrics{
			Name:  m.ID,
			MType: m.MType,
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
				return fmt.Errorf(
					"BatchMetricsUpdate: metric '%s': delta is required for counter",
					m.ID,
				)
			}
			if err := d.SetMetricValue(*m.Delta); err != nil {
				return fmt.Errorf("BatchMetricsUpdate: metric '%s': %v", m.ID, err)
			}
		case domain.Gauge:
			if m.Value == nil {
				return fmt.Errorf(
					"BatchMetricsUpdate: metric '%s': value is required for gauge",
					m.ID,
				)
			}
			if err := d.SetMetricValue(*m.Value); err != nil {
				return fmt.Errorf("BatchMetricsUpdate: metric '%s': %v", m.ID, err)
			}
		default:
			return fmt.Errorf("BatchMetricsUpdate: metric '%s': unsupported metric type '%s'", m.ID, m.MType)
		}

		domainMetrics = append(domainMetrics, d)
	}

	if err := ms.store.BatchUpdateMetrics(ctx, domainMetrics); err != nil {
		return fmt.Errorf("BatchMetricsUpdate: error batch update failed: %v", err)
	}
	return nil
}
