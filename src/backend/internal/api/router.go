// Package api wires up the Mortar HTTP router.
package api

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/db"
)

// handler holds shared dependencies for HTTP handlers.
type handler struct {
	registry *plugins.Registry
	database *db.DB
}

// NewRouter constructs and returns the root HTTP router.
func NewRouter(cfg *config.Config, reg *plugins.Registry, database *db.DB, webFS fs.FS) http.Handler {
	h := &handler{registry: reg, database: database}

	allowedOrigins := cfg.Server.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
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
	r.Get("/api/v1/activity", h.handleActivity)
	r.Get("/api/v1/downloads", h.handleDownloads)
	r.Get("/api/v1/home", h.handleHome)

	r.Handle("/*", spaHandler(webFS))

	return r
}

// spaHandler serves static files from fsys and falls back to index.html for
// any path that doesn't match a real file, enabling client-side routing.
func spaHandler(fsys fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(fsys))
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			f, err := fsys.Open(path)
			if err == nil {
				stat, serr := f.Stat()
				_ = f.Close()
				if serr == nil && !stat.IsDir() {
					fileServer.ServeHTTP(w, r)
					return
				}
			}
		}
		http.ServeFileFS(w, r, fsys, "index.html")
	}
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
