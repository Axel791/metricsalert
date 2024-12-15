package collector

import (
	"testing"

	"github.com/Axel791/metricsalert/internal/agent/collector/mocks"

	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	metric := mocks.MockCollector()

	require.Equal(t, uint64(1024*1024), metric.Alloc)
	require.Equal(t, uint64(512*1024), metric.BuckHashSys)
	require.Equal(t, uint64(300), metric.Frees)
	require.Equal(t, 0.01, metric.GCCPUFraction)
}
