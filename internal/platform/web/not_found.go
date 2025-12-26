package web

import (
	"context"
	"net/http"
)

func NotFound(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	env := Envelope{
		"code":    "not_found",
		"message": "the requested resource could not be found",
	}

	return Encode(ctx, w, http.StatusNotFound, env, nil)
}
