package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DSN string

	MaxConns     int
	MinConns     int
	MinIdleConns int

	MaxConnIdleTime time.Duration
	MaxConnLifeTime time.Duration

	HealthCheckPeriod time.Duration
}

func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("unable to parse the connection string: %w", err)
	}

	pgxCfg.MaxConns = int32(cfg.MaxConns)
	pgxCfg.MinConns = int32(cfg.MinConns)
	pgxCfg.MinIdleConns = int32(cfg.MinIdleConns)
	pgxCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	pgxCfg.MaxConnLifetime = cfg.MaxConnLifeTime
	pgxCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping the database: %w", err)
	}

	return pool, nil
}
