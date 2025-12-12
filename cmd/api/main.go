package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "ERROR", err)
		os.Exit(1)
	}
}

func (app *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	traceID := "00000000-0000-0000-0000-000000000000"

	app.logger.Info("request started", "trace_id", traceID)

	defer func() {
		app.logger.Info("request completed", "trace_id", traceID)
	}()

	data := map[string]string{
		"environment": app.config.environment,
		"version":     version,
		"build":       build,
	}

	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		app.logger.Error("unable to marshal", "ERROR", err)
		return
	}

	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
