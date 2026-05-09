// Package api — home dashboard endpoint.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// homeHealthSummary is the health sub-object in the home response.
type homeHealthSummary struct {
	AnyUnreachable  bool `json:"any_unreachable"`
	Total           int  `json:"total"`
	UnreachableCount int  `json:"unreachable_count"`
}

// homeResponse is the JSON body returned by GET /api/v1/home.
type homeResponse struct {
	RecentlyAdded []plugins.MediaItem `json:"recently_added"`
	HealthSummary homeHealthSummary   `json:"health_summary"`
}

// handleHome handles GET /api/v1/home.
// It assembles the home dashboard: recently-added items from the first
// LibraryBrowser plugin and a health summary across all plugins.
// Continue watching is omitted until auth is implemented.
// Always returns HTTP 200.
func (h *handler) handleHome(w http.ResponseWriter, _ *http.Request) {
	all := h.registry.All()

	// Recently added: use the first LibraryBrowser plugin found.
	var recentlyAdded []plugins.MediaItem
	for _, p := range all {
		lb, ok := p.(plugins.LibraryBrowser)
		if !ok {
			continue
		}
		pageSize := 10
		sort := "added"
		result, err := lb.Browse(plugins.BrowseOptions{
			Sort:     &sort,
			PageSize: &pageSize,
		})
		if err == nil {
			recentlyAdded = result.Items
		}
		// Use only the first browser plugin regardless of success.
		break
	}
	if recentlyAdded == nil {
		recentlyAdded = []plugins.MediaItem{}
	}

	// Health summary: check all plugins.
	total := len(all)
	unreachableCount := 0
	for _, p := range all {
		manifest := p.Manifest()
		entry := buildHealthEntry(manifest, p)
		if entry.Status == "unreachable" {
			unreachableCount++
		}
	}

	resp := homeResponse{
		RecentlyAdded: recentlyAdded,
		HealthSummary: homeHealthSummary{
			AnyUnreachable:  unreachableCount > 0,
			Total:           total,
			UnreachableCount: unreachableCount,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
