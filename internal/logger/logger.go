package logger

import (
	"io"
	"os"
	"time"

	"scaffold-api/internal/config"

	"github.com/rs/zerolog"
)

func New(cfg *config.Config) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return zerolog.Logger{}, err
	}

	zerolog.SetGlobalLevel(level)

	var writer io.Writer = os.Stdout
	if cfg.LogPretty {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	logger := zerolog.New(writer).With().
		Timestamp().
		Str("app", cfg.AppName).
		Logger().
		Level(level)

	return logger, nil
}
