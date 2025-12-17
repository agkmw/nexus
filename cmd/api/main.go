package main

import (
	"context"
	"flag"
	"os"
	"sync"

	"github.com/agkmw/reddit-clone/internal/api/sdk/mid"
	"github.com/agkmw/reddit-clone/internal/api/sdk/mux"
	"github.com/agkmw/reddit-clone/internal/platform/db"
	"github.com/agkmw/reddit-clone/internal/platform/logger"
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
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Application server port")
	flag.StringVar(&cfg.environment, "environment", "development", "Environment (development|staging|production)")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	// -------------------------------------------------------------------------

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

	// -------------------------------------------------------------------------

	ctx := context.Background()

	pool, err := db.Open("postgres://rdcadmin:pa55word@localhost/rdc?sslmode=disable")
	if err != nil {
		log.Error(ctx, "failed to open db", "ERROR", err)
		os.Exit(1)
	}
	defer pool.Close()

	log.Info(ctx, "connected to the database")

	// -------------------------------------------------------------------------

	app := &app{
		config: cfg,
		logger: log,
	}

	muxCfg := mux.Config{
		Environment: cfg.environment,
		Version:     version,
		Build:       build,
		Limiter: mid.LimiterConfig{
			Enabled: cfg.limiter.enabled,
			RPS:     cfg.limiter.rps,
			Burst:   cfg.limiter.burst,
		},
		Pool: pool,
		Log:  log,
	}

	if err := app.serve(ctx, mux.WebAPI(muxCfg)); err != nil {
		log.Error(ctx, "server failed", "ERROR", err)
	}
}
