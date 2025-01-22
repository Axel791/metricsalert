-- +goose Up
-- +goose StatementBegin
CREATE TABLE metrics (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    metric_type VARCHAR(10) NOT NULL CHECK (metric_type IN ('gauge', 'counter')),
    value DOUBLE PRECISION,
    delta BIGINT,
    UNIQUE (name, metric_type)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metrics;
-- +goose StatementEnd
