package logger

import (
	"log/slog"
	"os"
)

func Initialize(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}

func Named(name string) *slog.Logger {
	logger := slog.Default()
	if logger == nil {
		return nil
	}

	return logger.With("name", name)
}
