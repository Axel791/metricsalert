-- +goose Up
-- +goose StatementBegin
CREATE TABLE metrics (
    id TEXT PRIMARY KEY,
    metric_type VARCHAR(10) NOT NULL CHECK (metric_type IN ('gauge', 'counter')),
    value DOUBLE PRECISION,
    delta BIGINT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metrics;
-- +goose StatementEnd
