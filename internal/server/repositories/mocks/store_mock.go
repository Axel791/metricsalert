package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) UpdateGauge(name string, value float64) float64 {
	args := m.Called(name, value)
	if result := args.Get(0); result != nil {
		return result.(float64)
	}
	return 0
}

func (m *MockStore) UpdateCounter(name string, value int64) int64 {
	args := m.Called(name, value)
	if result := args.Get(0); result != nil {
		return result.(int64)
	}
	return 0
}

func (m *MockStore) GetMetric(name string) interface{} {
	args := m.Called(name)
	if result := args.Get(0); result != nil {
		return result
	}
	return 0
}

func (m *MockStore) GetAllMetrics() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}
