package web

import (
	"context"
	"net/http"
)

func NotFound (ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return NotFoundResponse(ctx, w)
}
