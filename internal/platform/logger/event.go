package logger

import (
	"context"
	"log/slog"
	"time"
)

type Level slog.Level

const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// =============================================================================

type EventHandler func(ctx context.Context, r Record)

type Events struct {
	Debug EventHandler
	Info  EventHandler
	Warn  EventHandler
	Error EventHandler
}

// =============================================================================

type Record struct {
	Time       time.Time
	Level      Level
	Message    string
	Attributes map[string]any
}

func toRecord(r slog.Record) Record {
	attrs := make(map[string]any, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	return Record{
		Time:       r.Time,
		Level:      Level(r.Level),
		Message:    r.Message,
		Attributes: attrs,
	}
}
