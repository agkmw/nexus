package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Envelope map[string]any

func Respond(ctx context.Context, w http.ResponseWriter, status int, data Envelope, headers http.Header) error {
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return nil
	}
	
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("web.respond.marshal: %w", err)
	}

	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(js); err != nil {
		return fmt.Errorf("web.respond.write: %w", err)
	}

	return nil
}
