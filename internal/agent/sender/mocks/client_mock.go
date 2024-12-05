package mocks

import (
	"github.com/Axel791/metricsalert/internal/agent/model/dto"
	"github.com/stretchr/testify/mock"
)

type MockMetricClient struct {
	mock.Mock
}

func (m *MockMetricClient) SendMetrics(metrics dto.Metrics) error {
	args := m.Called(metrics)
	return args.Error(0)
}
