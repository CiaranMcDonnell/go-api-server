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
	RequestTimeoutSecs int    `mapstructure:"REQUEST_TIMEOUT_SECS"`
	MaxBodyBytes       int64  `mapstructure:"MAX_BODY_BYTES"`

	RateLimitStrictRate    float64 `mapstructure:"RATE_LIMIT_STRICT_RATE"`
	RateLimitStrictBurst   int     `mapstructure:"RATE_LIMIT_STRICT_BURST"`
	RateLimitStandardRate  float64 `mapstructure:"RATE_LIMIT_STANDARD_RATE"`
	RateLimitStandardBurst int     `mapstructure:"RATE_LIMIT_STANDARD_BURST"`
	RateLimitRelaxedRate   float64 `mapstructure:"RATE_LIMIT_RELAXED_RATE"`
	RateLimitRelaxedBurst  int     `mapstructure:"RATE_LIMIT_RELAXED_BURST"`
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

	for _, key := range []string{
		"DB_DRIVER", "DB_SOURCE", "DB_MAX_CONNS", "DB_MIN_CONNS",
		"SERVER_ADDRESS", "ENVIRONMENT",
		"JWT_SECRET", "JWT_EXPIRATION_HOURS", "COOKIE_MAX_AGE_SECS",
		"SCHEMA_PATH", "CORS_ORIGINS",
		"REQUEST_TIMEOUT_SECS", "MAX_BODY_BYTES",
		"RATE_LIMIT_STRICT_RATE", "RATE_LIMIT_STRICT_BURST",
		"RATE_LIMIT_STANDARD_RATE", "RATE_LIMIT_STANDARD_BURST",
		"RATE_LIMIT_RELAXED_RATE", "RATE_LIMIT_RELAXED_BURST",
	} {
		viper.BindEnv(key)
	}

	viper.SetDefault("SERVER_ADDRESS", "0.0.0.0:8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("JWT_EXPIRATION_HOURS", 8)
	viper.SetDefault("COOKIE_MAX_AGE_SECS", 8*3600)
	viper.SetDefault("DB_MAX_CONNS", 100)
	viper.SetDefault("DB_MIN_CONNS", 20)
	viper.SetDefault("REQUEST_TIMEOUT_SECS", 10)
	viper.SetDefault("MAX_BODY_BYTES", 1<<20) // 1MB

	viper.SetDefault("RATE_LIMIT_STRICT_RATE", 0.2)
	viper.SetDefault("RATE_LIMIT_STRICT_BURST", 5)
	viper.SetDefault("RATE_LIMIT_STANDARD_RATE", 2.0)
	viper.SetDefault("RATE_LIMIT_STANDARD_BURST", 10)
	viper.SetDefault("RATE_LIMIT_RELAXED_RATE", 5.0)
	viper.SetDefault("RATE_LIMIT_RELAXED_BURST", 15)

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
