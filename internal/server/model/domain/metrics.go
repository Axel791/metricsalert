package domain

import (
	"errors"

	"gopkg.in/guregu/null.v4"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Metrics struct {
	ID    int64      `db:"id"`
	Name  string     `db:"name"`
	MType string     `db:"metric_type"`
	Delta null.Int   `db:"delta"`
	Value null.Float `db:"value"`
}

func (m *Metrics) ValidateMetricsType() error {
	if m.MType != Counter && m.MType != Gauge {
		return errors.New("invalid metrics type")
	}
	return nil
}

func (m *Metrics) ValidateMetricID() error {
	if m.Name == "" {
		return errors.New("metric id is required")
	}
	return nil
}

func (m *Metrics) SetMetricValue(value interface{}) error {
	switch m.MType {
	case Counter:
		v, ok := value.(int64)
		if !ok {
			return errors.New("invalid value type for counter metric, expected int64")
		}
		m.Delta = null.IntFrom(v)
		m.Value = null.Float{}

	case Gauge:
		v, ok := value.(float64)
		if !ok {
			return errors.New("invalid value type for gauge metric, expected float64")
		}
		m.Value = null.FloatFrom(v)
		m.Delta = null.Int{}

	default:
		return errors.New("unsupported metric type")
	}
	return nil
}
