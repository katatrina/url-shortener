package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/katatrina/url-shortener/internal/config"
)

// Setup .
func Setup(cfg *config.Config) {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     resolveLogLevel(cfg),
	}

	var handler slog.Handler
	if cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func resolveLogLevel(cfg *config.Config) slog.Level {
	switch cfg.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

type ctxKey struct{}

var loggerKey = ctxKey{}

// WithContext .
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromRequestContext .
func FromRequestContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
