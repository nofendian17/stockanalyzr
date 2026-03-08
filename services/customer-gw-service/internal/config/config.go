package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config stores runtime configuration for customer-service.
type Config struct {
	HTTPPort        string `mapstructure:"http_port"`
	UserServiceAddr string `mapstructure:"user_service_grpc_addr"`
	JWTSecret       string `mapstructure:"jwt_secret"`
	RedisDSN        string `mapstructure:"redis_dsn"`
}

// Load reads configuration from file (if exists) and environment variables.
func Load() *Config {
	v := viper.New()

	// Defaults
	v.SetDefault("http_port", "3000")
	v.SetDefault("user_service_grpc_addr", "localhost:50051")
	v.SetDefault("jwt_secret", "change-me-in-production")
	v.SetDefault("redis_dsn", "redis://localhost:6379/0")

	// Config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	_ = v.ReadInConfig() // ignore if not found

	// Environment variables
	v.SetEnvPrefix("CUSTOMER_GW_SERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	cfg := &Config{}
	_ = v.Unmarshal(cfg)

	return cfg
}
