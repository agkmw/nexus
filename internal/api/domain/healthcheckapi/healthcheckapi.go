package healthcheckapi

import (
	"context"
	"net/http"

	"github.com/agkmw/reddit-clone/internal/platform/web"
)

type api struct {
	cfg Config
}

func newAPI(cfg Config) *api {
	return &api{cfg: cfg}
}

func (api *api) healthcheckHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	data := web.Envelope{
		"environment": api.cfg.Environment,
		"version":     api.cfg.Version,
		"build":       api.cfg.Build,
	}

	return web.Respond(ctx, w, http.StatusOK, data)
}
