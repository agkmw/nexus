package mux

import (
	"context"

	"github.com/agkmw/reddit-clone/internal/api/domain/healthcheckapi"
	"github.com/agkmw/reddit-clone/internal/api/domain/userapi"
	"github.com/agkmw/reddit-clone/internal/api/sdk/mid"
	"github.com/agkmw/reddit-clone/internal/database/userdb"
	"github.com/agkmw/reddit-clone/internal/platform/logger"
	"github.com/agkmw/reddit-clone/internal/platform/web"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Environment string
	Version     string
	Build       string
	Limiter     mid.LimiterConfig
	Pool        *pgxpool.Pool
	Log         *logger.Logger
}

func WebAPI(cfg Config) *web.App {
	logFn := func(ctx context.Context, msg string, args ...any) {
		cfg.Log.Info(ctx, msg, args...)
	}

	app := web.NewApp(
		logFn,
		mid.HandleLogs(cfg.Log),
		mid.HandleErrors(cfg.Log),
		mid.RecoverPanics(),
		mid.RateLimit(cfg.Limiter),
	)

	RouteAdder(cfg, app)

	return app
}

func RouteAdder(cfg Config, app *web.App) {
	userapi.Routes(
		app,
		userdb.New(cfg.Pool),
	)

	healthcheckapi.Routes(
		app,
		healthcheckapi.Config{
			Environment: cfg.Environment,
			Version:     cfg.Version,
			Build:       cfg.Build,
		},
	)
}
