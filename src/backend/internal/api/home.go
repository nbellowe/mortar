package api

import (
	"net/http"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

type homeHealthSummary struct {
	AnyUnreachable   bool `json:"any_unreachable"`
	Total            int  `json:"total"`
	UnreachableCount int  `json:"unreachable_count"`
}

type homeResponse struct {
	RecentlyAdded                []plugins.MediaItem            `json:"recently_added"`
	RecentlyAddedRequiresLink    bool                           `json:"recently_added_requires_link"`
	ContinueWatching             []plugins.ContinueWatchingItem `json:"continue_watching"`
	ContinueWatchingEnabled      bool                           `json:"continue_watching_enabled"`
	ContinueWatchingRequiresLink bool                           `json:"continue_watching_requires_link"`
	HealthSummary                homeHealthSummary              `json:"health_summary"`
}

func (h *handler) handleHome(w http.ResponseWriter, r *http.Request) {
	resp := homeResponse{
		RecentlyAdded:    []plugins.MediaItem{},
		ContinueWatching: []plugins.ContinueWatchingItem{},
		HealthSummary:    h.homeHealthSummary(),
	}

	user := currentUser(r)

	if manifest, library, ok := firstLibraryBrowser(h.registry); ok {
		if h.enforceExternalLinks() && !userHasExternalLink(user, manifest.ID) {
			resp.RecentlyAddedRequiresLink = true
		} else {
			pageSize := 10
			sortBy := "added"
			if result, err := library.Browse(plugins.BrowseOptions{
				Sort:     &sortBy,
				PageSize: &pageSize,
			}); err == nil {
				resp.RecentlyAdded = result.Items
			}
		}
	}

	if manifest, reader, ok := firstContinueWatchingPlugin(h.registry); ok {
		resp.ContinueWatchingEnabled = true
		if h.enforceExternalLinks() && !userHasExternalLink(user, manifest.ID) {
			resp.ContinueWatchingRequiresLink = true
		} else if user != nil {
			limit := 10
			if items, err := reader.GetContinueWatching(*user, plugins.ContinueWatchingOptions{Limit: &limit}); err == nil {
				resp.ContinueWatching = items
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *handler) homeHealthSummary() homeHealthSummary {
	entries := h.health.list()
	summary := homeHealthSummary{Total: len(entries)}
	for _, entry := range entries {
		if entry.Status == "unreachable" {
			summary.UnreachableCount++
		}
	}
	summary.AnyUnreachable = summary.UnreachableCount > 0
	return summary
}
