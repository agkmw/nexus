package healthcheckapi

import (
	"net/http"

	"github.com/agkmw/reddit-clone/internal/platform/web"
)

type Config struct {
	Environment string
	Build       string
	Version     string
}

func Routes(app *web.App, cfg Config) {
	api := newAPI(cfg)

	app.HandlerFunc(http.MethodGet, "/v1", "/healthcheck", api.healthcheckHandler)
}
