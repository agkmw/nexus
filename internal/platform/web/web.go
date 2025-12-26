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

func NewApp(logFn LogFn, mw ...Middleware) *App {
	mux := chi.NewMux()

	app := App{
		log: logFn,
		mux: mux,
		mw:  mw,
	}

	app.NotFound(NotFound)
	app.MethodNotAllowed(MethodNotAllowed)

	return &app
}

func (app *App) HandlerFunc(
	method,
	group,
	path string,
	handler Handler,
) {
	handler = wrapMiddleware(app.mw, handler)

	if group != "" {
		path = group + path
	}

	app.mux.MethodFunc(method, path, app.handle(handler))
}

func (app *App) HandlerFuncWithMid(
	method,
	group,
	path string,
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
			app.log(ctx, "unexpected error occurred", "error", err)
		}
	}

	return h
}
