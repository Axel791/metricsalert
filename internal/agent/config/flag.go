package config

import (
	"flag"
	"strings"
	"time"
)

func ParseFlags(cfg *Config) (string, time.Duration, time.Duration, string, int) {
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
	rateLimit := flag.Int("l", cfg.RateLimit, "rate limit")

	flag.Parse()

	addr := *address
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	return addr,
		time.Duration(*reportInterval) * time.Second,
		time.Duration(*pollInterval) * time.Second,
		*key,
		*rateLimit
}
