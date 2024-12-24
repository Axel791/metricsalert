package mocks

import (
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) UpdateGauge(name string, value float64) domain.Metrics {
	args := m.Called(name, value)
	if result := args.Get(0); result != nil {
		return result.(domain.Metrics)
	}
	return domain.Metrics{}
}

func (m *MockStore) UpdateCounter(name string, value int64) domain.Metrics {
	args := m.Called(name, value)
	if result := args.Get(0); result != nil {
		return result.(domain.Metrics)
	}
	return domain.Metrics{}
}

func (m *MockStore) GetMetric(metricsDomain domain.Metrics) domain.Metrics {
	args := m.Called(metricsDomain.ID)
	if result := args.Get(0); result != nil {
		return result.(domain.Metrics)
	}
	return domain.Metrics{}
}

func (m *MockStore) GetAllMetrics() map[string]domain.Metrics {
	args := m.Called()
	if result := args.Get(0); result != nil {
		return result.(map[string]domain.Metrics)
	}
	return make(map[string]domain.Metrics)
}
