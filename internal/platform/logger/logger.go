package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

type TraceIDFn func(ctx context.Context) string

type Logger struct {
	handler   slog.Handler
	traceIDFn TraceIDFn
}

func New(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn) *Logger {
	return new(w, minLevel, serviceName, traceIDFn, Events{})
}

func NewWithEvents(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn, events Events) *Logger {
	return new(w, minLevel, serviceName, traceIDFn, events)
}

func NewStdLogger(log *Logger, level Level) *log.Logger {
	return slog.NewLogLogger(log.handler, slog.Level(level))
}

func (log *Logger) Debug(ctx context.Context, msg string, args ...any) {
	log.write(ctx, 3, LevelDebug, msg, args...)
}

func (log *Logger) Debugc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, caller, LevelDebug, msg, args...)
}

func (log *Logger) Info(ctx context.Context, msg string, args ...any) {
	log.write(ctx, 3, LevelInfo, msg, args...)
}

func (log *Logger) Infoc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, caller, LevelInfo, msg, args...)
}

func (log *Logger) Warn(ctx context.Context, msg string, args ...any) {
	log.write(ctx, 3, LevelWarn, msg, args...)
}

func (log *Logger) Warnc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, caller, LevelWarn, msg, args...)
}

func (log *Logger) Error(ctx context.Context, msg string, args ...any) {
	log.write(ctx, 3, LevelError, msg, args...)
}

func (log *Logger) Errorc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, caller, LevelError, msg, args...)
}

func (log *Logger) write(ctx context.Context, caller int, level Level, msg string, args ...any) {
	slogLevel := slog.Level(level)

	if !log.handler.Enabled(ctx, slogLevel) {
		return
	}

	var pc [1]uintptr
	runtime.Callers(caller, pc[:])

	r := slog.NewRecord(time.Now(), slogLevel, msg, pc[0])

	args = append(args, "trace id", log.traceIDFn(ctx))

	r.Add(args...)

	log.handler.Handle(ctx, r)
}

func new(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn, events Events) *Logger {
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source, ok := a.Value.Any().(*slog.Source)
			if ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.Level(minLevel),
		ReplaceAttr: f,
	}))

	if events.Debug != nil || events.Info != nil || events.Warn != nil || events.Error != nil {
		handler = newLogHandler(handler, events)
	}

	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}

	handler = handler.WithAttrs(attrs)

	return &Logger{
		handler:   handler,
		traceIDFn: traceIDFn,
	}
}

// =============================================================================

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
