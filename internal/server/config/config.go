package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

// Config структура для хранения конфигурации
type Config struct {
	Address         string `mapstructure:"address"`
	StoreInterval   int64  `mapstructure:"store_interval"`
	FileStoragePath string `mapstructure:"file_storage_path"`
	Restore         bool   `mapstructure:"restore"`
	UseFileStorage  bool   `mapstructure:"use_file_storage"`
	DatabaseDSN     string `mapstructure:"database_dsn"`
	MigrationsPath  string `mapstructure:"migrations_path"`
}

// ServerLoadConfig загружает конфигурацию из .env, переменных окружения и задает значения по умолчанию
func ServerLoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.SetDefault("ADDRESS", "localhost:8080")
	viper.SetDefault("STORE_INTERVAL", 300)
	viper.SetDefault("FILE_STORAGE_PATH", "./data.txt")
	viper.SetDefault("RESTORE", true)
	viper.SetDefault("USE_FILE_STORAGE", true)
	viper.SetDefault("MIGRATIONS_PATH", "./migrations")

	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Infof("filed find file config set defoult value: %v", err)
	}

	rawDSN := viper.GetString("DATABASE_DSN")
	log.Infof("DEBUG: viper.GetString(\"DATABASE_DSN\") => %q", rawDSN)

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	log.Infof("DEBUG: cfg.DatabaseDSN after Unmarshal => %q", cfg.DatabaseDSN)

	return &cfg, nil
}
