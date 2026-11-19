
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName         string
	AppEnv          string
	AppPort         string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	JWTAccessTTL    time.Duration
	JWTRefreshTTL   time.Duration
	MaxUploadSizeMB int64
}

func Load() (Config, error) {
	cfg := Config{
		AppName:         env("APP_NAME", "NovaMessenger"),
		AppEnv:          env("APP_ENV", "development"),
		AppPort:         env("APP_PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisURL:        os.Getenv("REDIS_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		MaxUploadSizeMB: 50,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}

	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}

	var err error

	cfg.JWTAccessTTL, err = durationEnv(
		"JWT_ACCESS_TTL",
		15*time.Minute,
	)
	if err != nil {
		return Config{}, err
	}

	cfg.JWTRefreshTTL, err = durationEnv(
		"JWT_REFRESH_TTL",
		168*time.Hour,
	)
	if err != nil {
		return Config{}, err
	}

	if raw := os.Getenv("MAX_UPLOAD_SIZE_MB"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)

		if err != nil || value < 1 || value > 1024 {
			return Config{}, fmt.Errorf(
				"MAX_UPLOAD_SIZE_MB must be between 1 and 1024",
			)
		}

		cfg.MaxUploadSizeMB = value
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func durationEnv(
	key string,
	fallback time.Duration,
) (time.Duration, error) {
	raw := os.Getenv(key)

	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)

	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}

	return value, nil
}
