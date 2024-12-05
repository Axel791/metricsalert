package dto

type Metrics struct {
	Alloc         float64 `json:"alloc"`
	BuckHashSys   float64 `json:"buckHashSys"`
	Frees         float64 `json:"frees"`
	GCCPUFraction float64 `json:"gcCPUFraction"`
}
