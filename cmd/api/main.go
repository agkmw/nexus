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

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: mux,
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

func (app *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	traceID := "00000000-0000-0000-0000-000000000000"

	app.logger.Info("request started", "trace_id", traceID)

	defer func() {
		app.logger.Info("request completed", "trace_id", traceID)
	}()

	data := web.Envelope{
		"environment": app.config.environment,
		"version":     version,
		"build":       build,
	}

	err := web.Respond(context.Background(), w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Error("unable to respond", "trace_id", traceID, "ERROR", err)
	}
}
