package api

type Metrics struct {
	Alloc         float64 `json:"Alloc"`
	BuckHashSys   float64 `json:"BuckHashSys"`
	Frees         float64 `json:"Frees"`
	GCCPUFraction float64 `json:"GCCPUFraction"`
	GCSys         float64 `json:"GCSys"`
	HeapAlloc     float64 `json:"HeapAlloc"`
	HeapIdle      float64 `json:"HeapIdle"`
	HeapInuse     float64 `json:"HeapInuse"`
	HeapObjects   float64 `json:"HeapObjects"`
	HeapReleased  float64 `json:"HeapReleased"`
	HeapSys       float64 `json:"HeapSys"`
	LastGC        float64 `json:"LastGC"`
	Lookups       float64 `json:"Lookups"`
	MCacheInuse   float64 `json:"MCacheInuse"`
	MSpanInuse    float64 `json:"MSpanInuse"`
	MSpanSys      float64 `json:"MSpanSys"`
	Mallocs       float64 `json:"Mallocs"`
	NextGC        float64 `json:"NextGC"`
	NumGC         float64 `json:"NumGC"`
	NumForcedGC   float64 `json:"NumForcedGC"`
	OtherSys      float64 `json:"OtherSys"`
	PauseTotalNs  float64 `json:"PauseTotalNs"`
	StackInuse    float64 `json:"StackInuse"`
	Sys           float64 `json:"Sys"`
	TotalAlloc    float64 `json:"TotalAlloc"`
}

type MetricPost struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta,omitempty"`
	Value float64 `json:"value,omitempty"`
}
