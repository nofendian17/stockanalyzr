package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config stores runtime configuration for user-service.
type Config struct {
	GRPCPort                string `mapstructure:"grpc_port"`
	PostgresDSN             string `mapstructure:"postgres_dsn"`
	RedisDSN                string `mapstructure:"redis_dsn"`
	JWTSecret               string `mapstructure:"jwt_secret"`
	JWTIssuer               string `mapstructure:"jwt_issuer"`
	JWTAccessExpiryMinutes  int    `mapstructure:"jwt_access_expiry_minutes"`
	JWTRefreshExpiryMinutes int    `mapstructure:"jwt_refresh_expiry_minutes"`
	MigrationsPath          string `mapstructure:"migrations_path"`
}

// Load reads configuration from file (if exists) and environment variables.
func Load() *Config {
	v := viper.New()

	// Defaults
	v.SetDefault("grpc_port", "50051")
	v.SetDefault("postgres_dsn", "postgres://postgres:postgres@localhost:5432/stockanalyzr?sslmode=disable")
	v.SetDefault("redis_dsn", "redis://localhost:6379/0")
	v.SetDefault("jwt_secret", "change-me-in-production")
	v.SetDefault("jwt_issuer", "user-service")
	v.SetDefault("jwt_access_expiry_minutes", 60)
	v.SetDefault("jwt_refresh_expiry_minutes", 10080) // 7 days
	v.SetDefault("migrations_path", "migrations")

	// Config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	_ = v.ReadInConfig() // ignore if not found

	// Environment variables
	v.SetEnvPrefix("USER_SERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	cfg := &Config{}
	_ = v.Unmarshal(cfg)

	return cfg
}
