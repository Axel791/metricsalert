package sender

import (
	"testing"

	"github.com/Axel791/metricsalert/internal/agent/model/api"
	"github.com/Axel791/metricsalert/internal/agent/sender/mocks"

	"github.com/stretchr/testify/require"
)

func TestSendMetrics(t *testing.T) {
	mockClient := new(mocks.MockMetricClient)
	metrics := api.Metrics{
		Alloc:         1024,
		BuckHashSys:   512,
		Frees:         300,
		GCCPUFraction: 0.01,
	}

	mockClient.On("SendMetrics", metrics).Return(nil).Once()

	err := mockClient.SendMetrics(metrics)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}
