package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

type submitRequestBody struct {
	ItemID   string  `json:"item_id,omitempty"`
	MediaID  string  `json:"media_id,omitempty"`
	PluginID string  `json:"plugin_id,omitempty"`
	Type     string  `json:"type"`
	Title    string  `json:"title,omitempty"`
	TmdbID   *string `json:"tmdb_id,omitempty"`
	ImdbID   *string `json:"imdb_id,omitempty"`
	TvdbID   *string `json:"tvdb_id,omitempty"`
	ISBN     *string `json:"isbn,omitempty"`
	ASIN     *string `json:"asin,omitempty"`
}

type searchResponse struct {
	Items            []plugins.MediaItem    `json:"items"`
	FailedPlugins    []string               `json:"failed_plugins"`
	ExistingRequests []plugins.Request      `json:"existing_requests"`
	AvailableMatches []plugins.LibraryMatch `json:"available_matches"`
}

type listRequestsResponse struct {
	Items      []plugins.Request `json:"items"`
	ReviewURLs map[string]string `json:"review_urls"`
}

func requestCapabilityForType(mediaType plugins.MediaType) string {
	switch mediaType {
	case plugins.MediaTypeMovie, plugins.MediaTypeShow:
		return string(plugins.CapabilityRequestsVideo)
	case plugins.MediaTypeAudiobook:
		return string(plugins.CapabilityRequestsAudio)
	case plugins.MediaTypeEbook:
		return string(plugins.CapabilityRequestsEbook)
	default:
		return ""
	}
}

func (h *handler) requesterPluginForType(w http.ResponseWriter, mediaType plugins.MediaType) (plugins.Requester, bool) {
	capability := requestCapabilityForType(mediaType)
	if capability == "" {
		jsonError(w, "unsupported request media type", http.StatusBadRequest)
		return nil, false
	}

	pluginID := h.registry.RouteRequest(capability)
	if pluginID == "" {
		jsonError(w, "no request plugin configured for media type", http.StatusServiceUnavailable)
		return nil, false
	}

	plugin := h.registry.Get(pluginID)
	if plugin == nil {
		jsonError(w, "configured request plugin not found", http.StatusServiceUnavailable)
		return nil, false
	}

	requester, ok := plugin.(plugins.Requester)
	if !ok {
		jsonError(w, "configured request plugin does not implement requests", http.StatusServiceUnavailable)
		return nil, false
	}
	return requester, true
}

func searchKey(item plugins.MediaItem) string {
	switch {
	case item.TmdbID != nil && *item.TmdbID != "":
		return "tmdb:" + *item.TmdbID
	case item.ImdbID != nil && *item.ImdbID != "":
		return "imdb:" + *item.ImdbID
	case item.TvdbID != nil && *item.TvdbID != "":
		return "tvdb:" + *item.TvdbID
	case item.ISBN != nil && *item.ISBN != "":
		return "isbn:" + *item.ISBN
	case item.ASIN != nil && *item.ASIN != "":
		return "asin:" + *item.ASIN
	default:
		return "id:" + item.ID
	}
}

func mergeMediaItem(base, candidate plugins.MediaItem) plugins.MediaItem {
	if base.Title == "" && candidate.Title != "" {
		base.Title = candidate.Title
	}
	if base.Year == nil && candidate.Year != nil {
		base.Year = candidate.Year
	}
	if base.Overview == nil && candidate.Overview != nil {
		base.Overview = candidate.Overview
	}
	if base.PosterURL == nil && candidate.PosterURL != nil {
		base.PosterURL = candidate.PosterURL
	}
	if len(base.Genres) == 0 && len(candidate.Genres) > 0 {
		base.Genres = candidate.Genres
	}
	if base.TmdbID == nil && candidate.TmdbID != nil {
		base.TmdbID = candidate.TmdbID
	}
	if base.ImdbID == nil && candidate.ImdbID != nil {
		base.ImdbID = candidate.ImdbID
	}
	if base.TvdbID == nil && candidate.TvdbID != nil {
		base.TvdbID = candidate.TvdbID
	}
	if base.ISBN == nil && candidate.ISBN != nil {
		base.ISBN = candidate.ISBN
	}
	if base.ASIN == nil && candidate.ASIN != nil {
		base.ASIN = candidate.ASIN
	}
	return base
}

func requesterPlugins(reg *plugins.Registry) []plugins.Plugin {
	var out []plugins.Plugin
	for _, plugin := range sortedPlugins(reg) {
		if _, ok := plugin.(plugins.Requester); ok {
			out = append(out, plugin)
		}
	}
	return out
}

func libraryExistsPlugins(reg *plugins.Registry) []plugins.Plugin {
	var out []plugins.Plugin
	for _, plugin := range sortedPlugins(reg) {
		if _, ok := plugin.(plugins.LibraryExists); ok {
			out = append(out, plugin)
		}
	}
	return out
}

func (h *handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		jsonError(w, "missing required query parameter: q", http.StatusBadRequest)
		return
	}

	pluginsWithRequests := requesterPlugins(h.registry)
	if len(pluginsWithRequests) == 0 {
		jsonError(w, "no request plugins configured", http.StatusServiceUnavailable)
		return
	}

	var (
		mu            sync.Mutex
		deduped       = map[string]plugins.MediaItem{}
		failedPlugins []string
		wg            sync.WaitGroup
	)

	for _, plugin := range pluginsWithRequests {
		requester := plugin.(plugins.Requester)
		manifest := plugin.Manifest()
		wg.Add(1)
		go func(displayName string, requester plugins.Requester) {
			defer wg.Done()
			items, err := requester.Search(query)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failedPlugins = append(failedPlugins, displayName)
				return
			}
			for _, item := range items {
				key := searchKey(item)
				if existing, ok := deduped[key]; ok {
					deduped[key] = mergeMediaItem(existing, item)
					continue
				}
				deduped[key] = item
			}
		}(manifest.DisplayName, requester)
	}
	wg.Wait()
	if failedPlugins == nil {
		failedPlugins = []string{}
	}

	items := make([]plugins.MediaItem, 0, len(deduped))
	keys := make([]string, 0, len(deduped))
	for key := range deduped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		items = append(items, deduped[key])
	}

	existingRequests := h.searchExistingRequests(items, pluginsWithRequests)
	availableMatches := h.searchAvailableMatches(items)

	writeJSON(w, http.StatusOK, searchResponse{
		Items:            items,
		FailedPlugins:    failedPlugins,
		ExistingRequests: existingRequests,
		AvailableMatches: availableMatches,
	})
}

func (h *handler) searchExistingRequests(items []plugins.MediaItem, requestPlugins []plugins.Plugin) []plugins.Request {
	if len(items) == 0 {
		return []plugins.Request{}
	}

	keys := make(map[string]struct{}, len(items))
	for _, item := range items {
		keys[searchKey(item)] = struct{}{}
	}

	var (
		mu      sync.Mutex
		matched []plugins.Request
		wg      sync.WaitGroup
	)
	for _, plugin := range requestPlugins {
		requester := plugin.(plugins.Requester)
		wg.Add(1)
		go func(requester plugins.Requester) {
			defer wg.Done()
			requests, err := requester.ListRequests(plugins.ListRequestsOptions{})
			if err != nil {
				return
			}
			local := make([]plugins.Request, 0)
			for _, req := range requests {
				if _, ok := keys[searchKey(req.Item)]; ok {
					local = append(local, req)
				}
			}
			if len(local) == 0 {
				return
			}
			mu.Lock()
			matched = append(matched, local...)
			mu.Unlock()
		}(requester)
	}
	wg.Wait()
	if matched == nil {
		return []plugins.Request{}
	}
	return matched
}

func (h *handler) searchAvailableMatches(items []plugins.MediaItem) []plugins.LibraryMatch {
	if len(items) == 0 {
		return []plugins.LibraryMatch{}
	}

	existsPlugins := libraryExistsPlugins(h.registry)
	if len(existsPlugins) == 0 {
		return []plugins.LibraryMatch{}
	}

	var (
		mu      sync.Mutex
		matches []plugins.LibraryMatch
		wg      sync.WaitGroup
	)

	for _, item := range items {
		item := item
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, plugin := range existsPlugins {
				reader := plugin.(plugins.LibraryExists)
				match, err := reader.FindMatch(item)
				if err != nil || match == nil {
					continue
				}
				mu.Lock()
				matches = append(matches, *match)
				mu.Unlock()
				return
			}
		}()
	}
	wg.Wait()

	if matches == nil {
		return []plugins.LibraryMatch{}
	}
	return matches
}

func (h *handler) handleListRequests(w http.ResponseWriter, r *http.Request) {
	requestPlugins := requesterPlugins(h.registry)
	if len(requestPlugins) == 0 {
		jsonError(w, "no request plugins configured", http.StatusServiceUnavailable)
		return
	}

	var opts plugins.ListRequestsOptions
	if s := r.URL.Query().Get("status"); s != "" {
		status := plugins.RequestStatus(s)
		opts.Status = &status
	}
	if rid := r.URL.Query().Get("requester_id"); rid != "" {
		opts.RequesterID = &rid
	}

	var items []plugins.Request
	reviewURLs := map[string]string{}
	for _, plugin := range requestPlugins {
		requester := plugin.(plugins.Requester)
		requests, err := requester.ListRequests(opts)
		if err != nil {
			continue
		}
		items = append(items, requests...)
		if currentUser(r) != nil && currentUser(r).Role == "admin" {
			if linker, ok := plugin.(plugins.RequestReviewURLProvider); ok {
				for _, req := range requests {
					reviewURLs[req.ID] = linker.ReviewURL(req.ID)
				}
			}
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	if items == nil {
		items = []plugins.Request{}
	}
	writeJSON(w, http.StatusOK, listRequestsResponse{
		Items:      items,
		ReviewURLs: reviewURLs,
	})
}

func (h *handler) handleSubmitRequest(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body submitRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Type == "" {
		jsonError(w, "type is required", http.StatusBadRequest)
		return
	}

	mediaType := plugins.MediaType(body.Type)
	requester, ok := h.requesterPluginForType(w, mediaType)
	if !ok {
		return
	}

	itemID := strings.TrimSpace(body.ItemID)
	externalID := strings.TrimSpace(body.MediaID)
	if externalID == "" && itemID != "" {
		_, externalID, _ = strings.Cut(itemID, ":")
	}
	if externalID == "" {
		jsonError(w, "media_id is required", http.StatusBadRequest)
		return
	}
	if itemID == "" && strings.TrimSpace(body.PluginID) != "" {
		itemID = body.PluginID + ":" + externalID
	}

	item := plugins.MediaItem{
		ID:         itemID,
		ExternalID: externalID,
		PluginID:   body.PluginID,
		Type:       mediaType,
		Title:      body.Title,
		TmdbID:     body.TmdbID,
		ImdbID:     body.ImdbID,
		TvdbID:     body.TvdbID,
		ISBN:       body.ISBN,
		ASIN:       body.ASIN,
	}

	if h.store != nil && item.ID != "" {
		exists, err := h.store.PendingRequestExists(item.ID)
		if err == nil && exists {
			jsonError(w, "a pending request already exists for this item", http.StatusConflict)
			return
		}
	}

	if pending, err := requester.ListRequests(plugins.ListRequestsOptions{Status: requestStatusPtr(plugins.RequestStatusPending)}); err == nil {
		for _, req := range pending {
			if searchKey(req.Item) == searchKey(item) {
				jsonError(w, "a pending request already exists for this item", http.StatusConflict)
				return
			}
		}
	}

	user := currentUser(r)
	created, err := requester.SubmitRequest(item, *user)
	if err != nil {
		jsonError(w, "submit request failed", http.StatusInternalServerError)
		return
	}
	if h.store != nil {
		_ = h.store.UpsertRequestSnapshot(created, user.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func requestStatusPtr(status plugins.RequestStatus) *plugins.RequestStatus {
	return &status
}

func (h *handler) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	pluginID, _, ok := strings.Cut(id, ":")
	if !ok {
		jsonError(w, "invalid request id", http.StatusBadRequest)
		return
	}

	plugin := h.registry.Get(pluginID)
	if plugin == nil {
		jsonError(w, "request plugin not found", http.StatusNotFound)
		return
	}
	requester, ok := plugin.(plugins.Requester)
	if !ok {
		jsonError(w, "plugin does not support requests", http.StatusBadRequest)
		return
	}

	req, err := requester.GetRequest(id)
	if err != nil {
		jsonError(w, "get request failed", http.StatusInternalServerError)
		return
	}
	if req == nil {
		jsonError(w, "request not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(req)
}
