package config

import (
	"flag"
)

func ParseFlags(cfg *Config) string {
	addr := flag.String("a", cfg.Address, "HTTP server address")
	flag.Parse()
	return *addr
}
