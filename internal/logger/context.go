package logger

import (
	"context"
	"log/slog"
)

type ctxKey int

const fieldsKey ctxKey = iota

func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	old, _ := ctx.Value(fieldsKey).([]slog.Attr)

	next := make([]slog.Attr, 0, len(old)+len(attrs))
	next = append(next, old...)
	next = append(next, attrs...)

	return context.WithValue(ctx, fieldsKey, next)
}

type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(fieldsKey).([]slog.Attr); ok {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}
