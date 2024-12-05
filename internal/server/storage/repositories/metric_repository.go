package repositories

type MetricRepository struct {
	metrics map[string]interface{}
}

func NewMetricRepository() *MetricRepository {
	return &MetricRepository{metrics: make(map[string]interface{})}
}

func (r *MetricRepository) UpdateGauge(name string, value float64) float64 {
	r.metrics[name] = value
	return value
}

func (r *MetricRepository) UpdateCounter(name string, value int64) int64 {
	if current, ok := r.metrics[name].(int64); ok {
		r.metrics[name] = current + value
	} else {
		r.metrics[name] = value
	}
	return value
}

func (r *MetricRepository) GetMetric(name string) interface{} {
	metric := r.metrics[name]
	return metric
}

func (r *MetricRepository) GetAllMetrics() map[string]interface{} {
	return r.metrics
}
