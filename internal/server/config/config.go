package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config структура для хранения конфигурации
type Config struct {
	Address         string `mapstructure:"ADDRESS"`
	StoreInterval   int64  `mapstructure:"STORE_INTERVAL"`
	FileStoragePath string `mapstructure:"FILE_STORAGE_PATH"`
	Restore         bool   `mapstructure:"RESTORE"`
	UseFileStorage  bool   `mapstructure:"USE_FILE_STORAGE"`
	DatabaseDSN     string `mapstructure:"DATABASE_DSN"`
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

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Infof("filed find file config set defoult value: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
