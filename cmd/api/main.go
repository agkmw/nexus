package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	db struct {
		dsn string

		maxConns     int
		minConns     int
		minIdleConns int

		maxConnIdleTime time.Duration
		maxConnLifeTime time.Duration

		healthCheckPeriod time.Duration
	}
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args[1:], os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	// -------------------------------------------------------------------------

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// -------------------------------------------------------------------------

	var cfg config

	fs := flag.NewFlagSet("reddit-clone", flag.PanicOnError)

	fs.IntVar(
		&cfg.port,
		"port",
		4000,
		"Application server port",
	)
	fs.StringVar(
		&cfg.environment,
		"environment",
		"development",
		"Environment (development|staging|production)",
	)

	fs.Float64Var(
		&cfg.limiter.rps,
		"limiter-rps",
		2,
		"Rate limiter maximum requests per second",
	)
	fs.IntVar(
		&cfg.limiter.burst,
		"limiter-burst",
		4,
		"Rate limiter maximum burst",
	)
	fs.BoolVar(
		&cfg.limiter.enabled,
		"limiter-enabled",
		true,
		"Enable rate limiter",
	)

	fs.StringVar(
		&cfg.db.dsn,
		"db-dsn",
		"",
		"PostgreSQL DSN",
	)
	fs.IntVar(
		&cfg.db.maxConns,
		"db-max-conns",
		25,
		"PostgreSQL max open connections",
	)
	fs.IntVar(
		&cfg.db.minConns,
		"db-min-conns",
		5,
		"PostgreSQL min open connections",
	)
	fs.IntVar(
		&cfg.db.minIdleConns,
		"db-min-idle-conns",
		25,
		"PostgreSQL min idle connections",
	)

	fs.DurationVar(
		&cfg.db.maxConnIdleTime,
		"db-max-idle-time",
		15*time.Minute,
		"PostgeSQL max connection idle time",
	)
	fs.DurationVar(
		&cfg.db.maxConnLifeTime,
		"db-max-life-time",
		2*time.Hour,
		"PostgeSQL max connection life time",
	)
	fs.DurationVar(
		&cfg.db.healthCheckPeriod,
		"db-heathz-period",
		time.Minute,
		"PostgeSQL health check period",
	)

	fs.Parse(args)

	// -------------------------------------------------------------------------

	var log *logger.Logger

	traceIDFn := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT! *******")
		},
	}

	log = logger.NewWithEvents(
		stdout,
		logger.LevelInfo,
		"reddit-clone",
		traceIDFn,
		events)

	// -------------------------------------------------------------------------

	dbCfg := db.Config{
		DSN: cfg.db.dsn,

		MaxConns:     cfg.db.maxConns,
		MinConns:     cfg.db.minConns,
		MinIdleConns: cfg.db.minIdleConns,

		MaxConnIdleTime: cfg.db.maxConnIdleTime,
		MaxConnLifeTime: cfg.db.maxConnLifeTime,

		HealthCheckPeriod: cfg.db.healthCheckPeriod,
	}

	pool, err := db.Open(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	defer pool.Close()

	log.Info(ctx, "successfully connected to the database")

	// -------------------------------------------------------------------------

	// TODO:
	var mux http.Handler

	if err := serve(ctx, cfg, mux, log); err != nil {
		return fmt.Errorf("server failed %w", err)
	}

	return nil
}
