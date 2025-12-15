package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agkmw/reddit-clone/internal/api/sdk/mid"
	"github.com/agkmw/reddit-clone/internal/app/sdk/errs"
	"github.com/agkmw/reddit-clone/internal/platform/logger"
	"github.com/agkmw/reddit-clone/internal/platform/validator"
	"github.com/agkmw/reddit-clone/internal/platform/web"
)

const version = "1.0.0"

var build = "dev"

type config struct {
	port        int
	environment string
	limiter     struct {
		enabled bool
		rps     float64
		burst   int
	}
}

type app struct {
	logger *logger.Logger
	config config
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Application server port")
	flag.StringVar(&cfg.environment, "environment", "development", "Environment (development|staging|production)")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

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

	api := web.NewApp(
		logFn,
		mid.HandleLogs(log),
		mid.HandleErrors(log),
		mid.RecoverPanics(),
		mid.RateLimit(app.config.limiter.enabled, app.config.limiter.rps, app.config.limiter.burst),
	)

	api.HandlerFunc(http.MethodGet, "/v1", "/healthcheck", app.healthcheckHandler)
	api.HandlerFunc(http.MethodGet, "/v1", "/testServerError", app.testServerError)
	api.HandlerFunc(http.MethodGet, "/v1", "/testClientError", app.testClientError)
	api.HandlerFunc(http.MethodPost, "/v1", "/posts", app.createPostHandler)

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      api,
		IdleTimeout:  2 * time.Minute,
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

		log.Info(ctx, "server shutting down", "sig", sig.String())

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

// =============================================================================

func (app *app) healthcheckHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	data := web.Envelope{
		"environment": app.config.environment,
		"version":     version,
		"build":       build,
	}

	return web.Respond(ctx, w, http.StatusOK, data)
}

func (app *app) testServerError(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return errs.NewServerError(errs.Internal, errors.New("server error"), errs.InternalMsg)
}

func (app *app) testClientError(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return errs.NewClientError(errs.BadRequest, errors.New("client error"), errs.BadRequestMsg)
}

func (app *app) createPostHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return errs.NewClientError(errs.BadRequest, err, errs.BadRequestMsg)

	}

	v := validator.New()

	v.Check(input.Title != "", "title", "must not be empty")
	v.Check(len(input.Title) <= 10, "title", "must not be longer than 10 bytes")

	v.Check(input.Body != "", "body", "must not be empty")
	v.Check(len(input.Body) <= 50, "body", "must not be longer than 50 bytes")

	if !v.Valid() {
		return errs.NewClientError(errs.FailedValidation, v.Errors, errs.FailedValidationMsg)
	}

	env := web.Envelope{
		"status": "success",
		"data":   input,
	}

	return web.Respond(ctx, w, http.StatusOK, env)
}
