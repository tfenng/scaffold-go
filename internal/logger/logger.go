package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"scaffold-api/internal/config"
)

func New(cfg *config.Config) (*slog.Logger, error) {
	level, err := parseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	options := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if cfg.AppEnv == "dev" {
		handler = slog.NewTextHandler(os.Stdout, options)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, options)
	}

	return slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
		slog.String("env", cfg.AppEnv),
	), nil
}

func parseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid LOG_LEVEL: %s", raw)
	}
}
