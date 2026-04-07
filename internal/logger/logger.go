package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/katatrina/url-shortener/internal/config"
	"github.com/rs/zerolog"
)

// Setup configs slog global logger with zerolog backend.
func Setup(env config.Environment) {
	var zl zerolog.Logger

	switch env {
	case config.EnvProduction:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		zl = zerolog.New(os.Stdout).With().Timestamp().Logger()
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		zl = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
			With().Timestamp().Logger()
	}

	handler := NewZerologHandler(zl, true)
	slog.SetDefault(slog.New(handler))
}

type ctxKey string

const loggerKey ctxKey = "slog_logger"

// WithContext .
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext retrieves the logger from context.
// It returns slog.Default() if not found - never panic.
// This is important: code shouldn't crash just because of a missing logger.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
