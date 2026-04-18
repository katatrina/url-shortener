package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/katatrina/url-shortener/internal/config"
)

// Setup cấu hình default slog logger dựa trên config.
//
// Prod → JSON handler (máy parse), stdout theo 12-factor.
// Local → Text handler (người đọc), stdout cho nhất quán.
//
// Cả hai đều bật AddSource để log kèm file:line — vô giá khi debug.
func Setup(cfg *config.Config) {
	opts := &slog.HandlerOptions{
		Level:     resolveLevel(cfg),
		AddSource: true,
	}

	var handler slog.Handler
	if cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// resolveLevel ưu tiên LOG_LEVEL từ config;
// rỗng thì fallback theo môi trường.
func resolveLevel(cfg *config.Config) slog.Level {
	switch strings.ToLower(cfg.Logger.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	if cfg.IsProduction() {
		return slog.LevelInfo
	}
	return slog.LevelDebug
}

// ---- Per-request logger in context ----

// ctxKey là struct rỗng unexported — không thể collide với key từ
// package khác, kể cả khi họ cũng định nghĩa `type ctxKey string`.
// Đây là idiom chuẩn cho context key trong Go.
type ctxKey struct{}

var loggerKey = ctxKey{}

func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext lấy logger từ ctx; không có thì trả slog.Default().
// Không bao giờ panic — caller không cần nil-check.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
