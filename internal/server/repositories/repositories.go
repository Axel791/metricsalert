package repositories

import (
	"context"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"time"
)

type StoreOptions struct {
	FilePath        string
	RestoreFromFile bool
	StoreInterval   time.Duration
	UseFileStore    bool
}

type Store interface {
	UpdateGauge(name string, value float64) domain.Metrics
	UpdateCounter(name string, value int64) domain.Metrics
	GetMetric(metricsDomain domain.Metrics) domain.Metrics
	GetAllMetrics() map[string]domain.Metrics
}

func StoreFactory(ctx context.Context, opts StoreOptions) (Store, error) {
	store := NewMetricRepository()
	if opts.UseFileStore {
		return NewFileStore(ctx, store, opts.FilePath, opts.RestoreFromFile, opts.StoreInterval)
	}
	return store, nil
}
