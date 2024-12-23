package api

type Metrics struct {
	Alloc         float64 `json:"alloc"`
	BuckHashSys   float64 `json:"buckHashSys"`
	Frees         float64 `json:"frees"`
	GCCPUFraction float64 `json:"gcCPUFraction"`
	GCSys         float64 `json:"gcSys"`
	HeapAlloc     float64 `json:"heapAlloc"`
	HeapIdle      float64 `json:"heapIdle"`
	HeapInuse     float64 `json:"heapInuse"`
	HeapObjects   float64 `json:"heapObjects"`
	HeapReleased  float64 `json:"heapReleased"`
	HeapSys       float64 `json:"heapSys"`
	LastGC        float64 `json:"lastGC"`
	Lookups       float64 `json:"lookups"`
	MCacheInuse   float64 `json:"mCacheInuse"`
	MSpanInuse    float64 `json:"mSpanInuse"`
	MSpanSys      float64 `json:"mSpanSys"`
	Mallocs       float64 `json:"mallocs"`
	NextGC        float64 `json:"nextGC"`
	NumGC         float64 `json:"numGC"`
	NumForcedGC   float64 `json:"numForcedGC"`
	OtherSys      float64 `json:"otherSys"`
	PauseTotalNs  float64 `json:"pauseTotalNs"`
	StackInuse    float64 `json:"stackInuse"`
	Sys           float64 `json:"sys"`
	TotalAlloc    float64 `json:"totalAlloc"`
}

type MetricPost struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta,omitempty"`
	Value float64 `json:"value,omitempty"`
}
