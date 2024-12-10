package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/Axel791/metricsalert/internal/agent/model/dto"
)

type MockMetricClient struct {
	mock.Mock
}

func (m *MockMetricClient) SendMetrics(metrics dto.Metrics) error {
	args := m.Called(metrics)
	return args.Error(0)
}
