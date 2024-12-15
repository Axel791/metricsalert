package mocks

import (
	"runtime"
)

func MockCollector() runtime.MemStats {
	return runtime.MemStats{
		Alloc:         1024 * 1024,
		BuckHashSys:   512 * 1024,
		Frees:         300,
		GCCPUFraction: 0.01,
	}
}
