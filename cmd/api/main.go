package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agkmw/reddit-clone/internal/app/sdk/errs"
	"github.com/agkmw/reddit-clone/internal/platform/logger"
	"github.com/agkmw/reddit-clone/internal/platform/web"
)

const version = "1.0.0"

var build = "dev"

type config struct {
	port        int
	environment string
}

type app struct {
	logger *logger.Logger
	config config
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Application server port")
	flag.StringVar(&cfg.environment, "environment", "development", "Environment (development|staging|production)")

	flag.Parse()

	var log *logger.Logger

	traceIDFn := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			// TODO: Implement sending email to the dev
			log.Info(ctx, "******* SEND ALERT! *******")
		},
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "reddit-clone", traceIDFn, events)

	ctx := context.Background()

	app := &app{
		config: cfg,
		logger: log,
	}

	logFn := func(ctx context.Context, msg string, args ...any) {
		log.Info(ctx, msg, args...)
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
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	log.Info(ctx, "server starting", "addr", server.Addr, "env", app.config.environment)

	errChan := make(chan error)
	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-shutdown

		log.Info(ctx, "server shutting down", "sig", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		errChan <- server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Error(ctx, "error shutting down the server", "error", err)
	}

	if err := <-errChan; err != nil {
		log.Error(ctx, "error shutting down the server", "ERROR", err)
	}

	log.Info(ctx, "server shut down")
}

func (app *app) healthcheckHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	app.logger.Info(ctx, "request started")

	defer func() {
		status := web.GetStatusCode(ctx)
		app.logger.Info(ctx, "request completed", "status", status)
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
	app.logger.Info(ctx, "request started")

	defer func() {
		status := web.GetStatusCode(ctx)
		app.logger.Info(ctx, "request completed", "status", status)
	}()

	se := errs.NewServerError(http.StatusInternalServerError, errors.New("test error"), errors.New("internal server error"))

	app.logger.Error(ctx, "handled error during request", "ERROR", se.LogErr)

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
	app.logger.Info(ctx, "request started")

	defer func() {
		status := web.GetStatusCode(ctx)
		app.logger.Info(ctx, "request completed", "status", status)
	}()

	ce := errs.NewClientError(http.StatusBadRequest, "there's nothing here, but chickens!", errors.New("you've messed up"))

	app.logger.Info(ctx, ce.ResErr.Error(), "method", r.Method, "status", ce.Code, "uri", r.RequestURI)

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
			err,
			errors.New("internal server error"),
		)
	}

	return nil
}
