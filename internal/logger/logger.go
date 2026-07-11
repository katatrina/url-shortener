package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// Setup .
func Setup(level string, production bool) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	var handler slog.Handler
	if production {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      logLevel,
			TimeFormat: time.Kitchen,
		})
	}

	slog.SetDefault(slog.New(contextHandler{handler}))
}
