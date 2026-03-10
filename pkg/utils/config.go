package utils

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver           string `mapstructure:"DB_DRIVER"`
	DBSource           string `mapstructure:"DB_SOURCE"`
	DBMaxConns         int    `mapstructure:"DB_MAX_CONNS"`
	DBMinConns         int    `mapstructure:"DB_MIN_CONNS"`
	ServerAddress      string `mapstructure:"SERVER_ADDRESS"`
	Environment        string `mapstructure:"ENVIRONMENT"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	JWTExpirationHours int    `mapstructure:"JWT_EXPIRATION_HOURS"`
	CookieMaxAgeSecs   int    `mapstructure:"COOKIE_MAX_AGE_SECS"`
	SchemaPath         string `mapstructure:"SCHEMA_PATH"`
	CORSOrigins        string `mapstructure:"CORS_ORIGINS"`
}

var (
	config *Config
	once   sync.Once
	cfgErr error
)

func LoadConfig() (*Config, error) {
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// Bind all config keys so env vars work without a config file
	for _, key := range []string{
		"DB_DRIVER", "DB_SOURCE", "DB_MAX_CONNS", "DB_MIN_CONNS",
		"SERVER_ADDRESS", "ENVIRONMENT",
		"JWT_SECRET", "JWT_EXPIRATION_HOURS", "COOKIE_MAX_AGE_SECS",
		"SCHEMA_PATH", "CORS_ORIGINS",
	} {
		viper.BindEnv(key)
	}

	// Set sensible defaults
	viper.SetDefault("SERVER_ADDRESS", "0.0.0.0:8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("JWT_EXPIRATION_HOURS", 8)
	viper.SetDefault("COOKIE_MAX_AGE_SECS", 8*3600)
	viper.SetDefault("DB_MAX_CONNS", 100)
	viper.SetDefault("DB_MIN_CONNS", 20)

	var cfg Config
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional — env vars can supply everything
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		slog.Warn("JWT_SECRET is shorter than 32 characters — consider using a stronger secret")
	}

	return &cfg, nil
}

func GetConfig() (*Config, error) {
	once.Do(func() {
		config, cfgErr = LoadConfig()
	})
	return config, cfgErr
}
