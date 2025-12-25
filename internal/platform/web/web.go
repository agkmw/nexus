package web

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type LogFn func(ctx context.Context, msg string, args ...any)

type App struct {
	log LogFn
	mux *chi.Mux
	mw  []Middleware
}

func NewApp(
	logFn LogFn,
	methodNotAllowed Handler,
	notFound Handler,
	mw ...Middleware,
) *App {
	mux := chi.NewMux()

	return &App{
		log: logFn,
		mux: mux,
		mw:  mw,
	}
}

func (app *App) HandlerFunc(
	method, group, path string,
	handler Handler,
) {
	handler = wrapMiddleware(app.mw, handler)

	if group != "" {
		path = group + path
	}

	app.mux.MethodFunc(method, path, app.handle(handler))
}

func (app *App) HandlerFuncWithMid(
	method, group, path string,
	handler Handler,
	middleware ...Middleware,
) {
	handler = wrapMiddleware(middleware, handler)
	handler = wrapMiddleware(app.mw, handler)

	if group != "" {
		path = group + path
	}

	app.mux.MethodFunc(method, path, app.handle(handler))
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.mux.ServeHTTP(w, r)
}

func (app *App) MethodNotAllowed(handler Handler) {
	app.mux.MethodNotAllowed(app.handle(handler))
}

func (app *App) NotFound(handler Handler) {
	app.mux.NotFound(app.handle(handler))
}

func (app *App) handle(handler Handler) http.HandlerFunc {
	h := func(w http.ResponseWriter, r *http.Request) {
		tracer := Tracer{
			Now:     time.Now(),
			TraceID: uuid.New().String(),
		}

		ctx := setTracer(r.Context(), &tracer)

		err := handler(ctx, w, r)
		if err != nil {
			app.log(ctx, "caught an error propagated through the chain", "ERROR", err)
		}
	}

	return h
}

// =============================================================================

type Middleware func(handler Handler) Handler

func wrapMiddleware(middlewares []Middleware, handler Handler) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		mid := middlewares[i]
		if mid != nil {
			handler = mid(handler)
		}
	}

	return handler
}
