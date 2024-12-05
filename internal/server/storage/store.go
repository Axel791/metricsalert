package storage

type Store interface {
	UpdateGauge(name string, value float64) float64
	UpdateCounter(name string, value int64) int64
	GetMetric(name string) interface{}
	GetAllMetrics() map[string]interface{}
}
