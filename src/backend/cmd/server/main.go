// Command server is the Mortar API server entry point.
// It loads configuration, initialises the plugin registry, and starts the
// HTTP server. All plugin implementations are registered here via their
// factory functions; see src/backend/internal/plugins/ for how to add a new one.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/nbellowe/mortar/src/backend/internal/api"
	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/jellyfin"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/radarr"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/sabnzbd"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/sonarr"
	"github.com/nbellowe/mortar/src/db"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the Mortar config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mortar: %v\n", err)
		os.Exit(1)
	}

	database, err := db.Open("mortar.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mortar: database init failed: %v\n", err)
		os.Exit(1)
	}

	registry := plugins.NewRegistry()
	// Plugin factories are registered here as plugin packages are added.
	registry.RegisterFactory("jellyfin", jellyfin.New)
	// Example (uncomment when implementing the jellyseerr plugin):
	//   registry.RegisterFactory("jellyseerr", jellyseerr.NewPlugin)
	registry.RegisterFactory("radarr", radarr.New)
	registry.RegisterFactory("sabnzbd", sabnzbd.New)
	registry.RegisterFactory("sonarr", sonarr.New)

	if err := registry.Init(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "mortar: plugin registry init failed: %v\n", err)
		os.Exit(1)
	}

	router := api.NewRouter(registry, database)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("mortar: listening on %s\n", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		fmt.Fprintf(os.Stderr, "mortar: server error: %v\n", err)
		os.Exit(1)
	}
}
