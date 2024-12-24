package api

type Metrics struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta,omitempty"`
	Value float64 `json:"value,omitempty"`
}

type GetMetric struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}

type HealthCheckResponse struct {
	Status string `json:"status"`
}
