package logger

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"

	"github.com/rs/zerolog"
)

type ZerologHandler struct {
	zl        zerolog.Logger
	attrs     []slog.Attr
	group     string
	addSource bool
}

func NewZerologHandler(zl zerolog.Logger, addSource bool) *ZerologHandler {
	return &ZerologHandler{zl: zl, addSource: addSource}
}

func (h *ZerologHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.mapLevel(level) >= zerolog.GlobalLevel()
}

func (h *ZerologHandler) Handle(_ context.Context, record slog.Record) error {
	level := h.mapLevel(record.Level)
	event := h.zl.WithLevel(level)

	if h.addSource {
		if record.PC != 0 {
			frames := runtime.CallersFrames([]uintptr{record.PC})
			f, _ := frames.Next()
			event = event.Str("source", f.File+":"+strconv.Itoa(f.Line))
		}
	}

	for _, attr := range h.attrs {
		event = h.addAttr(event, attr)
	}

	record.Attrs(func(attr slog.Attr) bool {
		event = h.addAttr(event, attr)
		return true
	})

	event.Msg(record.Message)
	return nil
}

func (h *ZerologHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ZerologHandler{
		zl: h.zl, attrs: newAttrs, group: h.group,
		addSource: h.addSource,
	}
}

func (h *ZerologHandler) WithGroup(name string) slog.Handler {
	newGroup := name
	if h.group != "" {
		newGroup = h.group + "." + name
	}
	return &ZerologHandler{
		zl: h.zl, attrs: h.attrs, group: newGroup,
		addSource: h.addSource,
	}
}

func (h *ZerologHandler) addAttr(event *zerolog.Event, attr slog.Attr) *zerolog.Event {
	key := attr.Key
	if h.group != "" {
		key = h.group + "." + key
	}

	val := attr.Value.Resolve()

	switch val.Kind() {
	case slog.KindString:
		return event.Str(key, val.String())
	case slog.KindInt64:
		return event.Int64(key, val.Int64())
	case slog.KindUint64:
		return event.Uint64(key, val.Uint64())
	case slog.KindFloat64:
		return event.Float64(key, val.Float64())
	case slog.KindBool:
		return event.Bool(key, val.Bool())
	case slog.KindTime:
		return event.Time(key, val.Time())
	case slog.KindDuration:
		return event.Dur(key, val.Duration())
	case slog.KindGroup:
		for _, groupAttr := range val.Group() {
			event = h.addAttr(event, groupAttr)
		}
		return event
	default:
		return event.Interface(key, val.Any())
	}
}

func (h *ZerologHandler) mapLevel(level slog.Level) zerolog.Level {
	switch {
	case level >= slog.LevelError:
		return zerolog.ErrorLevel
	case level >= slog.LevelWarn:
		return zerolog.WarnLevel
	case level >= slog.LevelInfo:
		return zerolog.InfoLevel
	default:
		return zerolog.DebugLevel
	}
}
