package config

import (
	"flag"
	log "github.com/sirupsen/logrus"
)

func ParseFlags(cfg *Config) (string, string, int64, string, bool) {
	// Переменные для хранения флагов
	var flagAddress string
	var flagDatabaseDSN string
	var flagStoreInterval int64
	var flagFileStoragePath string
	var flagRestore bool
	var flagUseFileStorage bool
	var flagMigrationsPath string

	// Привязываем флаги к переменным
	flag.StringVar(&flagAddress, "a", cfg.Address, "HTTP server address")
	flag.StringVar(&flagDatabaseDSN, "d", cfg.DatabaseDSN, "database DSN")
	flag.Int64Var(&flagStoreInterval, "i", cfg.StoreInterval, "interval in seconds for storing metrics (0 means sync)")
	flag.StringVar(&flagFileStoragePath, "f", cfg.FileStoragePath, "path to file for storing metrics")
	flag.BoolVar(&flagRestore, "r", cfg.Restore, "restore metrics from file on start (true/false)")
	flag.BoolVar(&flagUseFileStorage, "use-file", cfg.UseFileStorage, "use file storage (true/false)")
	flag.StringVar(&flagMigrationsPath, "m", cfg.MigrationsPath, "path to database migrations")

	// Парсим флаги
	flag.Parse()

	// Логирование для отладки
	log.Printf("Parsed flags:")
	log.Printf("  Address: %s", flagAddress)
	log.Printf("  DatabaseDSN: %s", flagDatabaseDSN)
	log.Printf("  StoreInterval: %d", flagStoreInterval)
	log.Printf("  FileStoragePath: %s", flagFileStoragePath)
	log.Printf("  Restore: %v", flagRestore)
	log.Printf("  UseFileStorage: %v", flagUseFileStorage)
	log.Printf("  MigrationsPath: %s", flagMigrationsPath)

	if flagAddress != "" {
		cfg.Address = flagAddress
	}
	if flagDatabaseDSN != "" {
		cfg.DatabaseDSN = flagDatabaseDSN
	}
	if flagStoreInterval != 0 {
		cfg.StoreInterval = flagStoreInterval
	}
	if flagFileStoragePath != "" {
		cfg.FileStoragePath = flagFileStoragePath
	}
	cfg.Restore = flagRestore
	cfg.UseFileStorage = flagUseFileStorage
	if flagMigrationsPath != "" {
		cfg.MigrationsPath = flagMigrationsPath
	}
	log.Infof("config: %+v", cfg)
	return cfg.Address, cfg.DatabaseDSN, cfg.StoreInterval, cfg.FileStoragePath, cfg.Restore
}
