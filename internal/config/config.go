package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAppEnv          = "dev"
	defaultHTTPHost        = "0.0.0.0"
	defaultHTTPPort        = 8080
	defaultLogLevel        = "info"
	defaultServiceName     = "scaffold-api"
	defaultReadTimeout     = 15 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	ServiceName      string
	AppEnv           string
	HTTPHost         string
	HTTPPort         int
	DBDSN            string
	LogLevel         string
	CORSAllowOrigins []string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	ShutdownTimeout  time.Duration
}

func Load() (*Config, error) {
	httpPort, err := intFromEnv("HTTP_PORT", defaultHTTPPort)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		ServiceName:      defaultServiceName,
		AppEnv:           stringFromEnv("APP_ENV", defaultAppEnv),
		HTTPHost:         stringFromEnv("HTTP_HOST", defaultHTTPHost),
		HTTPPort:         httpPort,
		DBDSN:            strings.TrimSpace(os.Getenv("DB_DSN")),
		LogLevel:         stringFromEnv("LOG_LEVEL", defaultLogLevel),
		CORSAllowOrigins: csvFromEnv("CORS_ALLOW_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),
		ReadTimeout:      defaultReadTimeout,
		WriteTimeout:     defaultWriteTimeout,
		ShutdownTimeout:  defaultShutdownTimeout,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func stringFromEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intFromEnv(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return value, nil
}

func csvFromEnv(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return fallback
	}

	return values
}

func (c Config) Validate() error {
	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP_PORT: %d", c.HTTPPort)
	}
	if c.DBDSN == "" {
		return fmt.Errorf("DB_DSN is required")
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}
	return nil
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}
