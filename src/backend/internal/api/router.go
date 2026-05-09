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
	"github.com/nbellowe/mortar/src/backend/internal/appstate"
	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/db"
)

// handler holds shared dependencies for HTTP handlers.
type handler struct {
	registry *plugins.Registry
	store    *appstate.Store
	health   *healthCache
}

// NewRouter constructs and returns the root HTTP router.
func NewRouter(cfg *config.Config, reg *plugins.Registry, database *db.DB, webFS fs.FS) http.Handler {
	store := appstate.New(database)
	h := &handler{
		registry: reg,
		store:    store,
		health:   newHealthCache(reg, store),
	}

	allowedOrigins := cfg.Server.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(h.sessionMiddleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Request-Id"},
		AllowCredentials: len(allowedOrigins) > 0 && allowedOrigins[0] != "*",
		MaxAge:           300,
	}))

	r.Get("/health", h.handleHealth)
	r.Get("/api/v1/auth/session", h.handleAuthSession)
	r.Post("/api/v1/auth/login", h.handleLogin)
	r.Post("/api/v1/auth/logout", h.handleLogout)

	r.Group(func(auth chi.Router) {
		auth.Use(h.requireAuth)
		auth.Get("/api/v1/search", h.handleSearch)
		auth.Get("/api/v1/requests", h.handleListRequests)
		auth.Post("/api/v1/requests", h.handleSubmitRequest)
		auth.Get("/api/v1/requests/{id}", h.handleGetRequest)
		auth.Get("/api/v1/activity", h.handleActivity)
		auth.Get("/api/v1/downloads", h.handleDownloads)
		auth.Get("/api/v1/home", h.handleHome)
		auth.Get("/api/v1/library", h.handleLibraryBrowse)
		auth.Post("/api/v1/library/play", h.handleLibraryPlay)
		auth.Group(func(admin chi.Router) {
			admin.Use(h.requireAdmin)
			admin.Get("/api/v1/health", h.handlePluginHealth)
		})
	})

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
