package config

import (
	"flag"
	"strings"
	"time"
)

func ParseFlags(cfg *Config) (string, time.Duration, time.Duration, string) {
	address := flag.String("a", cfg.Address, "HTTP server address")
	reportInterval := flag.Int(
		"r",
		int(cfg.ReportInterval),
		"Frequency of sending metrics to the server (in seconds)",
	)
	pollInterval := flag.Int(
		"p",
		int(cfg.PollInterval),
		"Frequency of collecting metrics from runtime (in seconds)",
	)
	key := flag.String("k", cfg.Key, "secret key")

	flag.Parse()

	addr := *address
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	return addr, time.Duration(*reportInterval) * time.Second, time.Duration(*pollInterval) * time.Second, *key
}
