// Package api wires up the Mortar HTTP router.
// Feature routes are added in later phases; only the /health ping is
// provided here so the server can be verified to be running.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/db"
)

// handler holds shared dependencies for HTTP handlers.
type handler struct {
	registry *plugins.Registry
	database *db.DB
}

// NewRouter constructs and returns the root HTTP router.
// Feature sub-routers are mounted here as they are implemented.
func NewRouter(reg *plugins.Registry, database *db.DB) http.Handler {
	h := &handler{registry: reg, database: database}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", h.handleHealth)

	return r
}

// healthResponse is the JSON body returned by the /health endpoint.
type healthResponse struct {
	Status string `json:"status"`
}

// handleHealth responds with a simple JSON ping indicating the server is up.
func (h *handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}
