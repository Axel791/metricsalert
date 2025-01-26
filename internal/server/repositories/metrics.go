package repositories

import (
	"context"
	"fmt"
	"github.com/Axel791/metricsalert/internal/db"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
)

var cursor = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// MetricsRepositoryHandler хранит ссылку на БД.
type MetricsRepositoryHandler struct {
	db *sqlx.DB
}

// NewMetricRepository — конструктор репозитория PostgreSQL.
func NewMetricRepository(db *sqlx.DB) *MetricsRepositoryHandler {
	return &MetricsRepositoryHandler{db: db}
}

// UpdateGauge - обновление Gauge.
func (r *MetricsRepositoryHandler) UpdateGauge(ctx context.Context, name string, gaugeVal float64) (domain.Metrics, error) {
	var result domain.Metrics

	err := db.RetryOperation(func() error {
		cteSQL := `
			WITH updated AS (
				UPDATE metrics
				SET value = $2,
					delta = NULL
				WHERE name = $1
				  AND metric_type = 'gauge'
				RETURNING id, name, metric_type, value, delta
			),
			inserted AS (
				INSERT INTO metrics (name, metric_type, value, delta)
				SELECT $1, 'gauge', $2, NULL
				WHERE NOT EXISTS (SELECT 1 FROM updated)
				RETURNING id, name, metric_type, value, delta
			)
			SELECT id, name, metric_type, value, delta FROM updated
			UNION ALL
			SELECT id, name, metric_type, value, delta FROM inserted
		`

		if err := r.db.QueryRowxContext(ctx, cteSQL, name, gaugeVal).StructScan(&result); err != nil {
			return fmt.Errorf("UpdateGauge CTE error: %w", err)
		}
		return nil
	})

	return result, err
}

func (r *MetricsRepositoryHandler) UpdateCounter(ctx context.Context, name string, value int64) (domain.Metrics, error) {
	var result domain.Metrics

	err := db.RetryOperation(func() error {
		cteSQL := `
			WITH updated AS (
				UPDATE metrics
				SET delta = metrics.delta + $2,
					value = NULL
				WHERE name = $1
				  AND metric_type = 'counter'
				RETURNING id, name, metric_type, value, delta
			),
			inserted AS (
				INSERT INTO metrics (name, metric_type, delta, value)
				SELECT $1, 'counter', $2, NULL
				WHERE NOT EXISTS (SELECT 1 FROM updated)
				RETURNING id, name, metric_type, value, delta
			)
			SELECT id, name, metric_type, value, delta FROM updated
			UNION ALL
			SELECT id, name, metric_type, value, delta FROM inserted
			`

		if err := r.db.QueryRowxContext(ctx, cteSQL, name, value).StructScan(&result); err != nil {
			return fmt.Errorf("UpdateCounter CTE error: %w", err)
		}

		return nil
	})

	return result, err
}

// GetMetric - получение метрики по ее ID.
func (r *MetricsRepositoryHandler) GetMetric(ctx context.Context, metric domain.Metrics) (domain.Metrics, error) {
	var result domain.Metrics

	err := db.RetryOperation(func() error {
		query, args, err := cursor.
			Select("id", "name", "metric_type", "value", "delta").
			From("metrics").
			Where(sq.Eq{"name": metric.Name, "metric_type": metric.MType}).
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
			Select("id", "name", "metric_type", "value", "delta").
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

			metricsMap[m.Name] = m
		}
		if err = rows.Err(); err != nil {
			return fmt.Errorf("error during rows iteration: %w", err)
		}

		return nil
	})

	return metricsMap, err
}

func (r *MetricsRepositoryHandler) BatchUpdateMetrics(ctx context.Context, metrics []domain.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	return db.RetryOperation(func() error {
		var sb strings.Builder

		sb.WriteString(`
WITH input (name, metric_type, val, delt) AS (
    VALUES
`)

		args := make([]interface{}, 0, len(metrics)*4)

		for i, m := range metrics {
			if i > 0 {
				sb.WriteString(",")
			}
			placeholderStart := len(args) + 1
			sb.WriteString(fmt.Sprintf(
				"($%d::text, $%d::text, $%d::double precision, $%d::bigint)",
				placeholderStart, placeholderStart+1, placeholderStart+2, placeholderStart+3,
			))

			args = append(args,
				m.Name,
				m.MType,
				m.Value.Float64,
				m.Delta.Int64,
			)
		}

		sb.WriteString(`
		),
		updated AS (
			UPDATE metrics mt
			SET
				value = CASE
					WHEN i.metric_type = 'gauge' THEN i.val
					ELSE NULL
				END,
				delta = CASE
					WHEN i.metric_type = 'counter' THEN mt.delta + i.delt
					ELSE i.delt
				END
			FROM input i
			WHERE mt.name = i.name
			  AND mt.metric_type = i.metric_type
			RETURNING mt.*
		),
		inserted AS (
			INSERT INTO metrics (name, metric_type, value, delta)
			SELECT
				i.name,
				i.metric_type,
				CASE WHEN i.metric_type = 'gauge' THEN i.val ELSE NULL END,
				CASE WHEN i.metric_type = 'counter' THEN i.delt ELSE NULL END
			FROM input i
			WHERE NOT EXISTS (
				SELECT 1 FROM updated u
				WHERE u.name = i.name
				  AND u.metric_type = i.metric_type
			)
			RETURNING *
		)
		SELECT 1 FROM updated
		UNION ALL
		SELECT 1 FROM inserted
		`)

		cteSQL := sb.String()

		if _, err := r.db.ExecContext(ctx, cteSQL, args...); err != nil {
			return fmt.Errorf("BatchUpdateMetrics CTE error: %w", err)
		}
		return nil
	})
}
