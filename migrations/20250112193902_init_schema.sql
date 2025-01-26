-- +goose Up
-- +goose StatementBegin
CREATE TABLE metrics (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    metric_type VARCHAR(10) NOT NULL CHECK (metric_type IN ('gauge', 'counter')),
    value DOUBLE PRECISION NOT NULL,
    delta BIGINT NOT NULL,
    UNIQUE (name, metric_type)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS metrics;
-- +goose StatementEnd
