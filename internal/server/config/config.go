package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config структура для хранения конфигурации
type Config struct {
	Address string `mapstructure:"ADDRESS"`
}

// ServerLoadConfig загружает конфигурацию из .env, переменных окружения и задает значения по умолчанию
func ServerLoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.SetDefault("ADDRESS", "localhost:8080")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("filed find file config set defoult value: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
