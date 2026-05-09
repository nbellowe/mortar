// Command server is the Mortar API server entry point.
// It loads configuration, initialises the plugin registry, and starts the
// HTTP server. All plugin implementations are registered here via their
// factory functions; see src/backend/internal/plugins/ for how to add a new one.
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/nbellowe/mortar/src/backend/internal/api"
	"github.com/nbellowe/mortar/src/backend/internal/appstate"
	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/jellyfin"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/jellyseerr"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/radarr"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/sabnzbd"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/sonarr"
	"github.com/nbellowe/mortar/src/db"
)

//go:embed web
var webFS embed.FS

func main() {
	configPath := flag.String("config", "config.yaml", "path to the Mortar config file")
	dbPath := flag.String("db", "mortar.db", "path to the SQLite database file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mortar: %v\n", err)
		os.Exit(1)
	}

	database, err := db.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mortar: database init failed: %v\n", err)
		os.Exit(1)
	}
	if err := appstate.New(database).SyncUsersFromConfig(cfg.Users); err != nil {
		fmt.Fprintf(os.Stderr, "mortar: user bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	registry := plugins.NewRegistry()
	// Plugin factories are registered here as plugin packages are added.
	registry.RegisterFactory("jellyfin", jellyfin.New)
	registry.RegisterFactory("jellyseerr", jellyseerr.New)
	registry.RegisterFactory("radarr", radarr.New)
	registry.RegisterFactory("sabnzbd", sabnzbd.New)
	registry.RegisterFactory("sonarr", sonarr.New)

	if err := registry.Init(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "mortar: plugin registry init failed: %v\n", err)
		os.Exit(1)
	}

	webAssets, err := fs.Sub(webFS, "web")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mortar: web assets unavailable: %v\n", err)
		os.Exit(1)
	}

	router := api.NewRouter(cfg, registry, database, webAssets)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("mortar: listening on %s\n", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		fmt.Fprintf(os.Stderr, "mortar: server error: %v\n", err)
		os.Exit(1)
	}
}
