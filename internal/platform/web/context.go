package web

import (
	"context"
	"time"
)

type ctxKey string

const key ctxKey = "ctxKey"

const defaultTraceID = "00000000-0000-0000-0000-000000000000"

type Tracer struct {
	Now        time.Time
	StatusCode int
	TraceID    string
}

func GetTracer(ctx context.Context) *Tracer {
	tracer, ok := ctx.Value(key).(*Tracer)
	if !ok {
		return &Tracer{
			Now:     time.Now(),
			TraceID: defaultTraceID,
		}
	}

	return tracer
}

func GetTime(ctx context.Context) time.Time {
	tracer, ok := ctx.Value(key).(*Tracer)
	if !ok {
		return time.Now()
	}

	return tracer.Now
}

func GetStatusCode(ctx context.Context) int {
	tracer, ok := ctx.Value(key).(*Tracer)
	if !ok {
		return 0
	}

	return tracer.StatusCode
}

func GetTraceID(ctx context.Context) string {
	tracer, ok := ctx.Value(key).(*Tracer)
	if !ok {
		return defaultTraceID
	}

	return tracer.TraceID
}

func setStatusCode(ctx context.Context, statusCode int) {
	tracer, ok := ctx.Value(key).(*Tracer)
	if !ok {
		return
	}

	tracer.StatusCode = statusCode
}

func setTracer(ctx context.Context, tracer *Tracer) context.Context {
	return context.WithValue(ctx, key, tracer)
}
