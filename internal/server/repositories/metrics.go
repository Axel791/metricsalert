package repositories

import (
	"context"
	"fmt"
	"github.com/Axel791/metricsalert/internal/server/db"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
)

var (
	cursor = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

// MetricsRepositoryHandler хранит ссылку на БД.
type MetricsRepositoryHandler struct {
	db *sqlx.DB
}

// NewMetricRepository — конструктор репозитория PostgreSQL.
func NewMetricRepository(db *sqlx.DB) *MetricsRepositoryHandler {
	return &MetricsRepositoryHandler{db: db}
}

// UpdateGauge - обновление Gauge.
func (r *MetricsRepositoryHandler) UpdateGauge(ctx context.Context, name string, value float64) (domain.Metrics, error) {
	var result domain.Metrics

	err := db.RetryOperation(func() error {
		metric := domain.Metrics{
			ID:    name,
			MType: domain.Gauge,
			Value: null.FloatFrom(value),
			Delta: null.Int{},
		}

		query, args, err := cursor.
			Insert("metrics").
			Columns("id", "metric_type", "value", "delta").
			Values(metric.ID, metric.MType, metric.Value, nil).
			Suffix(`
				ON CONFLICT (id) 
				DO UPDATE SET
					metric_type = EXCLUDED.metric_type,
					value       = EXCLUDED.value,
					delta       = NULL
			`).
			ToSql()

		if err != nil {
			return fmt.Errorf("building upsert gauge query: %w", err)
		}

		_, execErr := r.db.ExecContext(ctx, query, args...)
		if execErr != nil {
			return fmt.Errorf("exec upsert gauge query: %w", execErr)
		}

		result = metric
		return nil
	})

	return result, err
}

// UpdateCounter - обновление counter
func (r *MetricsRepositoryHandler) UpdateCounter(ctx context.Context, name string, value int64) (domain.Metrics, error) {
	var result domain.Metrics

	err := db.RetryOperation(func() error {
		metric := domain.Metrics{
			ID:    name,
			MType: domain.Counter,
			Delta: null.IntFrom(value),
			Value: null.Float{},
		}

		query, args, err := cursor.
			Insert("metrics").
			Columns("id", "metric_type", "delta", "value").
			Values(metric.ID, metric.MType, metric.Delta, nil).
			Suffix(`
				ON CONFLICT (id) 
				DO UPDATE SET
					metric_type = EXCLUDED.metric_type,
					delta       = metrics.delta + EXCLUDED.delta,
					value       = NULL
			`).
			ToSql()
		if err != nil {
			return fmt.Errorf("building upsert counter query: %w", err)
		}

		_, execErr := r.db.ExecContext(ctx, query, args...)
		if execErr != nil {
			return fmt.Errorf("exec upsert counter query: %w", execErr)
		}

		result = metric
		return nil
	})

	return result, err
}

// GetMetric - получение метрики по ее ID.
func (r *MetricsRepositoryHandler) GetMetric(ctx context.Context, metric domain.Metrics) (domain.Metrics, error) {
	var result domain.Metrics

	err := db.RetryOperation(func() error {
		query, args, err := cursor.
			Select("id", "metric_type", "value", "delta").
			From("metrics").
			Where(sq.Eq{"id": metric.ID}).
			Limit(1).
			ToSql()
		if err != nil {
			return fmt.Errorf("build get metric query: %w", err)
		}

		err = r.db.GetContext(ctx, &result, query, args...)
		if err != nil {
			return fmt.Errorf("get metric from db: %w", err)
		}

		return nil
	})

	return result, err
}

// GetAllMetrics - Получение всех метрик
func (r *MetricsRepositoryHandler) GetAllMetrics(ctx context.Context) (map[string]domain.Metrics, error) {
	metricsMap := make(map[string]domain.Metrics)

	err := db.RetryOperation(func() error {
		query, args, err := cursor.
			Select("id", "metric_type", "value", "delta").
			From("metrics").
			ToSql()
		if err != nil {
			return fmt.Errorf("build get all metrics query: %w", err)
		}

		rows, err := r.db.QueryxContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("query all metrics: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var m domain.Metrics
			if scanErr := rows.StructScan(&m); scanErr != nil {
				return fmt.Errorf("scan metric row: %w", scanErr)
			}
			metricsMap[m.ID] = m
		}

		return nil
	})

	return metricsMap, err
}

// BatchUpdateMetrics - обновление метрик батчами
func (r *MetricsRepositoryHandler) BatchUpdateMetrics(ctx context.Context, metrics []domain.Metrics) error {
	return db.RetryOperation(func() error {
		insertBuilder := cursor.Insert("metrics").
			Columns("id", "metric_type", "value", "delta")

		for _, metric := range metrics {
			insertBuilder = insertBuilder.Values(metric.ID, metric.MType, metric.Value, metric.Delta)
		}

		sql, args, err := insertBuilder.ToSql()
		if err != nil {
			return fmt.Errorf("cannot build insert query: %w", err)
		}

		sql += `
			ON CONFLICT (id)
			DO UPDATE SET
				metric_type = EXCLUDED.metric_type,
				value = EXCLUDED.value,
				delta = CASE 
					WHEN EXCLUDED.metric_type = 'Counter' THEN metrics.delta + EXCLUDED.delta
					ELSE EXCLUDED.delta
				END
		`

		if _, err := r.db.ExecContext(ctx, sql, args...); err != nil {
			return fmt.Errorf("cannot exec batch upsert: %w", err)
		}

		return nil
	})
}
