package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config структура для хранения конфигурации
type Config struct {
	Address   string `mapstructure:"ADDRESS"`
	Key       string `mapstructure:"KEY"`
	CryptoKey string `mapstructure:"CRYPTO_KEY"`

	UseGRPC        bool   `mapstructure:"USE_GRPC"`
	GRPCAddress    string `mapstructure:"GRPC_ADDRESS"`
	ReportInterval int64  `mapstructure:"REPORT_INTERVAL"`
	PollInterval   int64  `mapstructure:"POLL_INTERVAL"`
	RateLimit      int    `mapstructure:"RATE_LIMIT"`
}

// AgentLoadConfig загружает конфигурацию из .env, переменных окружения и задает значения по умолчанию
func AgentLoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.SetDefault("ADDRESS", "localhost:8080")
	viper.SetDefault("REPORT_INTERVAL", 10)
	viper.SetDefault("POLL_INTERVAL", 2)
	viper.SetDefault("USE_GRPC", false)

	_ = viper.BindEnv("KEY", "KEY")
	_ = viper.BindEnv("CRYPTO_KEY", "CRYPTO_KEY")
	_ = viper.BindEnv("RATE_LIMIT", "RATE_LIMIT")

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
