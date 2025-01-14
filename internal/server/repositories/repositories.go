package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
	"time"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
)

type StoreOptions struct {
	FilePath        string
	RestoreFromFile bool
	StoreInterval   time.Duration
	UseFileStore    bool
}

type Store interface {
	UpdateGauge(ctx context.Context, name string, value float64) (domain.Metrics, error)
	UpdateCounter(ctx context.Context, name string, value int64) (domain.Metrics, error)
	GetMetric(ctx context.Context, metric domain.Metrics) (domain.Metrics, error)
	GetAllMetrics(ctx context.Context) (map[string]domain.Metrics, error)
	BatchUpdateMetrics(ctx context.Context, metrics []domain.Metrics) error
}

func StoreFactory(ctx context.Context, db *sqlx.DB, opts StoreOptions) (Store, error) {
	store := NewMetricRepository(db)
	if opts.UseFileStore {
		return NewFileStore(ctx, store, opts.FilePath, opts.RestoreFromFile, opts.StoreInterval)
	}
	return store, nil
}
