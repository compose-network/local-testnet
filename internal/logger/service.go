package logger

import (
	"log/slog"
	"os"
)

func Initialize(level slog.Level) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}
