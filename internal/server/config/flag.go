package config

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
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

func SetServerParams() (string, time.Duration, string, bool, bool, string, string) {
	var (
		flagRestore, restore             bool
		flagStoreFile, storeFile         string
		flagAddress                      string
		flagStoreInterval, storeInterval time.Duration
		flagDebug                        bool
		flagKey                          string
		flagDataBase                     string
	)

	flag.BoolVar(&flagRestore, "r", false, "restore_true/false")
	flag.StringVar(&flagStoreFile, "f", "hh", "store_file")
	flag.StringVar(&flagAddress, "a", "hh", "server_address")
	flag.DurationVar(&flagStoreInterval, "i", 2, "store_interval_in_seconds")
	flag.BoolVar(&flagDebug, "debug", false, "debug_true/false")
	flag.StringVar(&flagKey, "k", "", "hash_key")
	flag.StringVar(&flagDataBase, "d", "", "db_address")
	flag.Parse()
	address, exists := os.LookupEnv("ADDRESS")
	if !exists {
		address = flagAddress
	}
	if storeFile, exists = os.LookupEnv("STORE_FILE"); !exists {
		storeFile = flagStoreFile
	}
	if strStoreInterval, exists := os.LookupEnv("STORE_INTERVAL"); !exists {
		storeInterval = flagStoreInterval
	} else {
		var err error
		if storeInterval, err = time.ParseDuration(strStoreInterval); err != nil {
			storeInterval = flagStoreInterval
		}
	}
	if strRestore, exists := os.LookupEnv("RESTORE"); !exists {
		restore = flagRestore
	} else {
		var err error
		if restore, err = strconv.ParseBool(strRestore); err != nil {
			restore = flagRestore
		}
	}
	key, exists := os.LookupEnv("KEY")
	if !exists {
		key = flagKey
	}
	database, exists := os.LookupEnv("DATABASE_DSN")
	if !exists {
		database = flagDataBase
	}
	return address, storeInterval, storeFile, restore, flagDebug, key, database
}
