package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) UpdateGauge(ctx context.Context, name string, value float64) (domain.Metrics, error) {
	args := m.Called(ctx, name, value)

	if res := args.Get(0); res != nil {
		return res.(domain.Metrics), args.Error(1)
	}
	return domain.Metrics{}, args.Error(1)
}

func (m *MockStore) UpdateCounter(ctx context.Context, name string, value int64) (domain.Metrics, error) {
	args := m.Called(ctx, name, value)
	if res := args.Get(0); res != nil {
		return res.(domain.Metrics), args.Error(1)
	}
	return domain.Metrics{}, args.Error(1)
}

func (m *MockStore) GetMetric(ctx context.Context, metric domain.Metrics) (domain.Metrics, error) {
	args := m.Called(ctx, metric)
	if res := args.Get(0); res != nil {
		return res.(domain.Metrics), args.Error(1)
	}
	return domain.Metrics{}, args.Error(1)
}

func (m *MockStore) GetAllMetrics(ctx context.Context) (map[string]domain.Metrics, error) {
	args := m.Called(ctx)
	if res := args.Get(0); res != nil {
		return res.(map[string]domain.Metrics), args.Error(1)
	}
	return make(map[string]domain.Metrics), args.Error(1)
}

func (m *MockStore) BatchMetricsUpdate(ctx context.Context, metrics []api.Metrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}
