package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ReadParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
