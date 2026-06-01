package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/katatrina/url-shortener/internal/config"
	"github.com/lmittmann/tint"
)

// Setup .
func Setup(cfg *config.Config) {
	var logLevel slog.Level
	switch cfg.LogLevel {
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
	if cfg.AppEnv.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
		})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:  true,
			Level:      logLevel,
			TimeFormat: time.Kitchen,
		})
	}

	slog.SetDefault(slog.New(handler))
}
