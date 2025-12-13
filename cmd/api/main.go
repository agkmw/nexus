package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/agkmw/reddit-clone/internal/app/sdk/errs"
	"github.com/agkmw/reddit-clone/internal/platform/web"
)

const version = "1.0.0"

var build = "dev"

type config struct {
	port        int
	environment string
}

type app struct {
	logger *slog.Logger
	config config
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Application server port")
	flag.StringVar(&cfg.environment, "environment", "development", "Environment (development|staging|production)")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, ok := a.Value.Any().(*slog.Source)
				if ok {
					v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
					return slog.Attr{Key: "file", Value: slog.StringValue(v)}
				}
			}
			return a
		},
	}))

	app := &app{
		config: cfg,
		logger: logger,
	}

	logFn := func(ctx context.Context, msg string, args ...any) {
		logger.Info(msg, args...)
	}

	api := web.NewApp(logFn, nil)

	api.HandlerFunc(http.MethodGet, "/v1", "/healthcheck", app.healthcheckHandler)
	api.HandlerFunc(http.MethodGet, "/v1", "/testServerError", app.testServerError)
	api.HandlerFunc(http.MethodGet, "/v1", "/testClientError", app.testClientError)

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      api,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logger.Info("server starting", "addr", server.Addr, "env", app.config.environment)

	errChan := make(chan error)
	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-shutdown

		logger.Info("server shutting down", "sig", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		errChan <- server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("error shutting down the server", "error", err)
		os.Exit(1)
	}

	if err := <-errChan; err != nil {
		logger.Error("error shutting down the server", "ERROR", err)
		os.Exit(1)
	}

	logger.Info("server shut down")
}

func (app *app) healthcheckHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	traceID := web.GetTraceID(ctx)

	app.logger.Info("request started", "trace_id", traceID)

	defer func() {
		status := web.GetStatusCode(ctx)
		app.logger.Info("request completed", "trace_id", traceID, "status", status)
	}()

	data := web.Envelope{
		"environment": app.config.environment,
		"version":     version,
		"build":       build,
	}

	err := web.Respond(ctx, w, http.StatusOK, data)
	// TODO: Just return the error from the web.Respond after centralized logging is implemented
	if err != nil {
		return errs.NewServerError(
			http.StatusInternalServerError,
			errors.New("unable to respond"),
			errors.New("internal server error"),
		)
	}

	return nil
}

func (app *app) testServerError(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	traceID := web.GetTraceID(ctx)

	app.logger.Info("request started", "trace_id", traceID)

	defer func() {
		status := web.GetStatusCode(ctx)
		app.logger.Info("request completed", "trace_id", traceID, "status", status)
	}()

	se := errs.NewServerError(http.StatusInternalServerError, errors.New("test error"), errors.New("internal server error"))

	app.logger.Error("handled error during request", "ERROR", se.LogErr)

	env := web.Envelope{
		"status":  "error",
		"message": se.ResErr.Error(),
	}

	err := web.Respond(ctx, w, se.Code, env)
	// TODO: Just return the error from the web.Respond after centralized logging is implemented
	if err != nil {
		return errs.NewServerError(
			http.StatusInternalServerError,
			errors.New("unable to respond"),
			errors.New("internal server error"),
		)
	}

	return nil
}

func (app *app) testClientError(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	traceID := web.GetTraceID(ctx)

	app.logger.Info("request started", "trace_id", traceID)

	defer func() {
		status := web.GetStatusCode(ctx)
		app.logger.Info("request completed", "trace_id", traceID, "status", status)
	}()

	ce := errs.NewClientError(http.StatusBadRequest, "there's nothing here, but chickens!", errors.New("you've messed up"))

	app.logger.Info(ce.ResErr.Error(), "method", r.Method, "status", ce.Code, "uri", r.RequestURI)

	env := web.Envelope{
		"status":  "fail",
		"message": ce.ResErr.Error(),
		"data":    ce.Data,
	}

	err := web.Respond(ctx, w, ce.Code, env)
	// TODO: Just return the error from the web.Respond after centralized logging is implemented
	if err != nil {
		return errs.NewServerError(
			http.StatusInternalServerError,
			errors.New("unable to respond"),
			errors.New("internal server error"),
		)
	}

	return nil
}
