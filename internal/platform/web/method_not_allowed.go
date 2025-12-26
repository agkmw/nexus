package web

import (
	"context"
	"fmt"
	"net/http"
)

func MethodNotAllowed(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	env := Envelope{
		"code":    "method_not_allowed",
		"message": fmt.Sprintf("the %s method is not supported for this resource", r.Method),
	}

	return Encode(ctx, w, http.StatusMethodNotAllowed, env, nil)
}
