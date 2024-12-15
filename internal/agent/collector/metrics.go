package collector

import (
	"runtime"
)

func Collector() runtime.MemStats {
	var metric runtime.MemStats
	runtime.ReadMemStats(&metric)
	return metric
}
