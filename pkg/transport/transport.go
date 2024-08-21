// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package transport

import (
	"context"
	"fmt"
	"io"
	"net/http"

	logging "log/slog"

	"github.com/go-chi/chi/v5"
)

type Logger interface {
	// Log error during request processing
	ErrorContext(ctx context.Context, msg string, args ...any)
}

func NewHandler(reloader *EndpointReloader) http.Handler {
	routes := chi.NewRouter()
	routes.Handle("/*", reloader)
	routes.Get("/status", handleStatus)
	return routes
}

func handleStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := io.WriteString(w, `{"status":"OK"}`)
	if err != nil {
		logging.Error(fmt.Sprintf("write http response body error for /status: %s", err))
	}
}
