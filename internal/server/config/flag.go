package config

import (
	"flag"
	log "github.com/sirupsen/logrus"
)

func ParseFlags(cfg *Config) (string, string, int64, string, bool) {
	addr := flag.String("a", cfg.Address, "HTTP server address")

	databaseDSN := flag.String("d", cfg.DatabaseDSN, "database DSN")

	storeIntervalFlag := flag.Int64(
		"i", cfg.StoreInterval, "interval in seconds for storing metrics (0 means sync)",
	)
	filePathFlag := flag.String("f", cfg.FileStoragePath, "path to file for storing metrics")
	restoreFlag := flag.Bool("r", cfg.Restore, "restore metrics from file on start (true/false)")
	flag.Parse()

	log.Infof("addr %s", *addr)
	log.Infof("databasedsn: %s", *databaseDSN)
	log.Infof("store interval: %d", *storeIntervalFlag)
	log.Infof("file path: %s", *filePathFlag)
	log.Infof("restore interval: %v", *restoreFlag)

	return *addr, *databaseDSN, *storeIntervalFlag, *filePathFlag, *restoreFlag
}
