package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

type libraryBrowseResponse struct {
	Items             []plugins.MediaItem `json:"items"`
	Total             int                 `json:"total"`
	Page              int                 `json:"page"`
	PageSize          int                 `json:"page_size"`
	AvailableGenres   []string            `json:"available_genres"`
	RequiresLink      bool                `json:"requires_link"`
	PluginDisplayName string              `json:"plugin_display_name,omitempty"`
}

type libraryPlayRequest struct {
	ItemID string `json:"item_id"`
}

type libraryPlayResponse struct {
	URL string `json:"url"`
}

func userHasExternalLink(user *plugins.MortarUser, pluginID string) bool {
	if user == nil {
		return false
	}
	for _, ext := range user.ExternalAccounts {
		if ext.PluginID == pluginID {
			return true
		}
	}
	return false
}

func sortedPlugins(reg *plugins.Registry) []plugins.Plugin {
	all := reg.All()
	ids := make([]string, 0, len(all))
	for id := range all {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]plugins.Plugin, 0, len(ids))
	for _, id := range ids {
		out = append(out, all[id])
	}
	return out
}

func firstLibraryBrowser(reg *plugins.Registry) (plugins.PluginManifest, plugins.LibraryBrowser, bool) {
	for _, plugin := range sortedPlugins(reg) {
		library, ok := plugin.(plugins.LibraryBrowser)
		if !ok {
			continue
		}
		return plugin.Manifest(), library, true
	}
	return plugins.PluginManifest{}, nil, false
}

func firstContinueWatchingPlugin(reg *plugins.Registry) (plugins.PluginManifest, plugins.LibraryResumeReadable, bool) {
	for _, plugin := range sortedPlugins(reg) {
		reader, ok := plugin.(plugins.LibraryResumeReadable)
		if !ok {
			continue
		}
		return plugin.Manifest(), reader, true
	}
	return plugins.PluginManifest{}, nil, false
}

func distinctGenres(items []plugins.MediaItem) []string {
	set := map[string]struct{}{}
	for _, item := range items {
		for _, genre := range item.Genres {
			if genre != "" {
				set[genre] = struct{}{}
			}
		}
	}
	genres := make([]string, 0, len(set))
	for genre := range set {
		genres = append(genres, genre)
	}
	sort.Strings(genres)
	return genres
}

func (h *handler) handleLibraryBrowse(w http.ResponseWriter, r *http.Request) {
	manifest, library, ok := firstLibraryBrowser(h.registry)
	resp := libraryBrowseResponse{
		Items:           []plugins.MediaItem{},
		Page:            1,
		PageSize:        24,
		AvailableGenres: []string{},
	}
	if !ok {
		writeJSON(w, http.StatusOK, resp)
		return
	}

	if h.enforceExternalLinks() && !userHasExternalLink(currentUser(r), manifest.ID) {
		resp.RequiresLink = true
		resp.PluginDisplayName = manifest.DisplayName
		writeJSON(w, http.StatusOK, resp)
		return
	}

	options := plugins.BrowseOptions{}
	if mediaType := r.URL.Query().Get("type"); mediaType != "" {
		t := plugins.MediaType(mediaType)
		options.Type = &t
	}
	if genre := strings.TrimSpace(r.URL.Query().Get("genre")); genre != "" {
		options.Genre = &genre
	}
	if sortValue := strings.TrimSpace(r.URL.Query().Get("sort")); sortValue != "" {
		options.Sort = &sortValue
	}
	if pageValue := r.URL.Query().Get("page"); pageValue != "" {
		if page, err := strconv.Atoi(pageValue); err == nil && page > 0 {
			options.Page = &page
		}
	}
	if pageSizeValue := r.URL.Query().Get("page_size"); pageSizeValue != "" {
		if pageSize, err := strconv.Atoi(pageSizeValue); err == nil && pageSize > 0 {
			options.PageSize = &pageSize
		}
	}
	if options.Page == nil {
		page := 1
		options.Page = &page
	}
	if options.PageSize == nil {
		pageSize := 24
		options.PageSize = &pageSize
	}

	result, err := library.Browse(options)
	if err != nil {
		jsonError(w, "library browse failed", http.StatusBadGateway)
		return
	}

	resp.Items = result.Items
	resp.Total = result.Total
	resp.Page = result.Page
	resp.PageSize = result.PageSize
	resp.AvailableGenres = distinctGenres(result.Items)
	writeJSON(w, http.StatusOK, resp)
}

func (h *handler) handleLibraryPlay(w http.ResponseWriter, r *http.Request) {
	var body libraryPlayRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.ItemID == "" {
		jsonError(w, "item_id is required", http.StatusBadRequest)
		return
	}

	pluginID, _, ok := strings.Cut(body.ItemID, ":")
	if !ok || pluginID == "" {
		jsonError(w, "invalid item_id", http.StatusBadRequest)
		return
	}

	plugin := h.registry.Get(pluginID)
	if plugin == nil {
		jsonError(w, "library plugin not found", http.StatusNotFound)
		return
	}
	library, ok := plugin.(plugins.LibraryBrowser)
	if !ok {
		jsonError(w, "plugin does not support playback handoff", http.StatusBadRequest)
		return
	}
	if h.enforceExternalLinks() && !userHasExternalLink(currentUser(r), pluginID) {
		jsonError(w, "linked account required for playback handoff", http.StatusForbidden)
		return
	}

	item, err := library.GetItem(body.ItemID)
	if err != nil {
		jsonError(w, "failed to load library item", http.StatusBadGateway)
		return
	}
	if item == nil {
		jsonError(w, "library item not found", http.StatusNotFound)
		return
	}

	url, err := library.GetPlayURL(*item, *currentUser(r))
	if err != nil {
		jsonError(w, "failed to build playback handoff", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, libraryPlayResponse{URL: url})
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
