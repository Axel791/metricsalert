package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/Axel791/metricsalert/internal/agent/model/api"
)

type MockMetricClient struct {
	mock.Mock
}

func (m *MockMetricClient) SendMetrics(metrics api.Metrics) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func (m *MockMetricClient) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}
