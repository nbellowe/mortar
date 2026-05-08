// Package api wires up the Mortar HTTP router.
// Feature routes are added in later phases; only the /health ping is
// provided here so the server can be verified to be running.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "X-Request-Id"},
		MaxAge:         300,
	}))

	r.Get("/health", h.handleHealth)

	// Feature routes.
	// TODO: restrict to admin role once auth middleware is wired up (health spec AC).
	r.Get("/api/v1/health", h.handlePluginHealth)
	r.Get("/api/v1/search", h.handleSearch)
	r.Get("/api/v1/requests", h.handleListRequests)
	r.Post("/api/v1/requests", h.handleSubmitRequest)
	r.Get("/api/v1/requests/{id}", h.handleGetRequest)

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
