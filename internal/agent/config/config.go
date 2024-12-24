package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config структура для хранения конфигурации
type Config struct {
	Address        string `mapstructure:"address"`
	ReportInterval int64  `mapstructure:"report_interval"`
	PollInterval   int64  `mapstructure:"poll_interval"`
}

// AgentLoadConfig загружает конфигурацию из .env, переменных окружения и задает значения по умолчанию
func AgentLoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.SetDefault("ADDRESS", "localhost:8080")
	viper.SetDefault("REPORT_INTERVAL", 10)
	viper.SetDefault("POLL_INTERVAL", 2)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Infof("filed to find conf file, set default value: %v.", err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
