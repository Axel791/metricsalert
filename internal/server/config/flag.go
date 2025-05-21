package config

import (
	"flag"
)

func ParseFlags(cfg *Config) (string, string, int64, string, bool, string, string) {
	addr := flag.String("a", cfg.Address, "HTTP server address")
	databaseDSN := flag.String("d", cfg.DatabaseDSN, "database DSN")

	storeIntervalFlag := flag.Int64(
		"i", cfg.StoreInterval, "interval in seconds for storing metrics (0 means sync)",
	)
	filePathFlag := flag.String("f", cfg.FileStoragePath, "path to file for storing metrics")
	restoreFlag := flag.Bool("r", cfg.Restore, "restore metrics from file on start (true/false)")
	key := flag.String("k", cfg.Key, "secret key")
	cryptoKey := flag.String(
		"crypto-key",
		cfg.CryptoKey,
		"path to PEM public key for RSA encryption (agent)",
	)

	flag.Parse()

	return *addr, *databaseDSN, *storeIntervalFlag, *filePathFlag, *restoreFlag, *key, *cryptoKey
}
