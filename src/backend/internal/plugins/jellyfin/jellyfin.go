// Package jellyfin implements a Mortar plugin for the Jellyfin media server.
//
// # Configuration
//
// The plugin reads from a config.PluginConfig:
//   - cfg.URL    — Jellyfin base URL (e.g. "http://jellyfin:8096")
//   - cfg.APIKey — Jellyfin API key (use ${VAR} in YAML to avoid hardcoding)
//   - cfg.Username — Jellyfin user ID to browse as (not the username string;
//     obtain the UUID from Jellyfin's user management page or the /Users endpoint).
//     This field is re-used for the default browsing user ID.
//
// Declared capabilities: library.browse, library.exists, library.resume
package jellyfin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// Plugin implements plugins.Plugin, plugins.LibraryBrowser, plugins.LibraryExists,
// and plugins.LibraryResumeReadable for a Jellyfin server.
type Plugin struct {
	id            string
	baseURL       string
	apiKey        string
	defaultUserID string
	httpClient    *http.Client

	// serverID is lazily fetched from /System/Info/Public and cached.
	serverIDMu sync.Mutex
	serverID   string
}

// New creates a new Jellyfin plugin from the given config.
// cfg.URL must be set; cfg.APIKey must be set; cfg.Username is used as the
// default Jellyfin user ID for library browsing.
func New(cfg config.PluginConfig) (plugins.Plugin, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("jellyfin: url is required")
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("jellyfin: api_key is required")
	}
	baseURL := strings.TrimRight(cfg.URL, "/")
	return &Plugin{
		id:            cfg.ID,
		baseURL:       baseURL,
		apiKey:        cfg.APIKey,
		defaultUserID: cfg.Username,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ---------------------------------------------------------------------------
// Plugin base interface
// ---------------------------------------------------------------------------

// Manifest returns the plugin's identity and declared capabilities.
func (p *Plugin) Manifest() plugins.PluginManifest {
	return plugins.PluginManifest{
		ID:          p.id,
		Type:        "jellyfin",
		DisplayName: "Jellyfin",
		Capabilities: []plugins.Capability{
			plugins.CapabilityLibraryBrowse,
			plugins.CapabilityLibraryExists,
			plugins.CapabilityLibraryResume,
		},
	}
}

// Health checks whether the Jellyfin server is reachable.
func (p *Plugin) Health() (plugins.HealthStatus, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	start := time.Now()

	resp, err := p.doGet("/health", nil)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		detail := err.Error()
		return plugins.HealthStatus{
			Status:    "unreachable",
			Reachable: false,
			LatencyMs: latency,
			CheckedAt: now,
			Detail:    &detail,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail := fmt.Sprintf("upstream returned HTTP %d", resp.StatusCode)
		return plugins.HealthStatus{
			Status:    "unreachable",
			Reachable: false,
			LatencyMs: latency,
			CheckedAt: now,
			Detail:    &detail,
		}, nil
	}

	status := "healthy"
	if latency > 2000 {
		status = "degraded"
	}
	return plugins.HealthStatus{
		Status:    status,
		Reachable: true,
		LatencyMs: latency,
		CheckedAt: now,
	}, nil
}

// ---------------------------------------------------------------------------
// LibraryBrowser
// ---------------------------------------------------------------------------

// jellyfinItem is the subset of Jellyfin's item JSON we care about.
type jellyfinItem struct {
	ID             string            `json:"Id"`
	Type           string            `json:"Type"`
	Name           string            `json:"Name"`
	ProductionYear *int              `json:"ProductionYear"`
	Overview       *string           `json:"Overview"`
	ImageTags      map[string]string `json:"ImageTags"`
	Genres         []string          `json:"Genres"`
	ProviderIds    map[string]string `json:"ProviderIds"`
	UserData       *jellyfinUserData `json:"UserData"`
	RunTimeTicks   *int64            `json:"RunTimeTicks"` // 100-nanosecond ticks
}

// jellyfinUserData holds per-user playback state from Jellyfin.
type jellyfinUserData struct {
	PlayedPercentage *float64 `json:"PlayedPercentage"`
	PlaybackPosition int64    `json:"PlaybackPositionTicks"` // 100-nanosecond ticks
	LastPlayedDate   *string  `json:"LastPlayedDate"`
}

// jellyfinItemsResponse is the envelope returned by Jellyfin browse calls.
type jellyfinItemsResponse struct {
	Items            []jellyfinItem `json:"Items"`
	TotalRecordCount int            `json:"TotalRecordCount"`
}

// Browse returns a paginated list of library items matching the given options.
func (p *Plugin) Browse(options plugins.BrowseOptions) (plugins.PagedResult[plugins.MediaItem], error) {
	userID := p.defaultUserID
	if userID == "" {
		return plugins.PagedResult[plugins.MediaItem]{}, fmt.Errorf("jellyfin: no default user id configured; set username in plugin config")
	}

	page := 1
	if options.Page != nil && *options.Page > 0 {
		page = *options.Page
	}
	pageSize := 50
	if options.PageSize != nil && *options.PageSize > 0 {
		pageSize = *options.PageSize
	}
	startIndex := (page - 1) * pageSize

	params := url.Values{}
	params.Set("Recursive", "true")
	params.Set("Fields", "Overview,Genres,ExternalUrls,ProviderIds,ImageTags")
	params.Set("Limit", strconv.Itoa(pageSize))
	params.Set("StartIndex", strconv.Itoa(startIndex))

	// Sort
	sortBy := "SortName"
	sortOrder := "Ascending"
	if options.Sort != nil {
		switch *options.Sort {
		case "added":
			sortBy = "DateCreated"
			sortOrder = "Descending"
		case "year":
			sortBy = "ProductionYear"
			sortOrder = "Descending"
		case "title":
			sortBy = "SortName"
			sortOrder = "Ascending"
		}
	}
	params.Set("SortBy", sortBy)
	params.Set("SortOrder", sortOrder)

	// Type filter
	itemTypes := "Movie,Series"
	if options.Type != nil {
		switch *options.Type {
		case plugins.MediaTypeMovie:
			itemTypes = "Movie"
		case plugins.MediaTypeShow:
			itemTypes = "Series"
		}
	}
	params.Set("IncludeItemTypes", itemTypes)

	// Genre filter
	if options.Genre != nil && *options.Genre != "" {
		params.Set("Genres", *options.Genre)
	}

	path := fmt.Sprintf("/Users/%s/Items", userID)
	resp, err := p.doGet(path, params)
	if err != nil {
		return plugins.PagedResult[plugins.MediaItem]{}, fmt.Errorf("jellyfin: browse request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return plugins.PagedResult[plugins.MediaItem]{}, fmt.Errorf("jellyfin: browse returned HTTP %d", resp.StatusCode)
	}

	var jellyResp jellyfinItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jellyResp); err != nil {
		return plugins.PagedResult[plugins.MediaItem]{}, fmt.Errorf("jellyfin: browse decode error: %w", err)
	}

	items := make([]plugins.MediaItem, 0, len(jellyResp.Items))
	for _, ji := range jellyResp.Items {
		items = append(items, p.toMediaItem(ji))
	}

	return plugins.PagedResult[plugins.MediaItem]{
		Items:    items,
		Total:    jellyResp.TotalRecordCount,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetItem returns the MediaItem for the given Mortar item ID.
// The id should be in the form "pluginID:jellyfinItemID".
func (p *Plugin) GetItem(id string) (*plugins.MediaItem, error) {
	userID := p.defaultUserID
	if userID == "" {
		return nil, fmt.Errorf("jellyfin: no default user id configured; set username in plugin config")
	}

	jellyfinID := p.externalIDFromMortarID(id)

	params := url.Values{}
	params.Set("Fields", "Overview,Genres,ExternalUrls,ProviderIds,ImageTags")

	path := fmt.Sprintf("/Users/%s/Items/%s", userID, jellyfinID)
	resp, err := p.doGet(path, params)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: get item request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyfin: get item returned HTTP %d", resp.StatusCode)
	}

	var ji jellyfinItem
	if err := json.NewDecoder(resp.Body).Decode(&ji); err != nil {
		return nil, fmt.Errorf("jellyfin: get item decode error: %w", err)
	}

	item := p.toMediaItem(ji)
	return &item, nil
}

// GetPlayURL returns the Jellyfin web deep-link URL for the given item.
// The URL format is: <baseURL>/web/index.html#!/details?id=<itemId>&serverId=<serverId>
// This opens Jellyfin's own web UI; Mortar does not proxy or play video inline.
//
// The user param is used to verify that an ExternalAccountLink exists for this plugin.
// Mortar gates this call on link presence, but the plugin returns an error if missing
// to be safe.
func (p *Plugin) GetPlayURL(item plugins.MediaItem, user plugins.MortarUser) (string, error) {
	if item.PluginID != p.id {
		return "", fmt.Errorf("jellyfin: GetPlayURL called with item from plugin %q, expected %q", item.PluginID, p.id)
	}

	// Verify the user has an account link for this plugin.
	linked := false
	for _, ext := range user.ExternalAccounts {
		if ext.PluginID == p.id {
			linked = true
			break
		}
	}
	if !linked {
		return "", fmt.Errorf("jellyfin: user %q has no external account linked for plugin %q", user.ID, p.id)
	}

	serverID, err := p.getServerID()
	if err != nil {
		return "", fmt.Errorf("jellyfin: could not resolve server id for play URL: %w", err)
	}

	jellyfinID := item.ExternalID
	playURL := fmt.Sprintf("%s/web/index.html#!/details?id=%s&serverId=%s",
		p.baseURL, url.QueryEscape(jellyfinID), url.QueryEscape(serverID))
	return playURL, nil
}

// ---------------------------------------------------------------------------
// LibraryExists
// ---------------------------------------------------------------------------

// FindMatch searches the Jellyfin library for an item matching the given MediaItem.
// It tries IMDB, TMDB, TVDB, then title+year in order, returning the first match.
func (p *Plugin) FindMatch(item plugins.MediaItem) (*plugins.LibraryMatch, error) {
	userID := p.defaultUserID
	if userID == "" {
		return nil, fmt.Errorf("jellyfin: no default user id configured; set username in plugin config")
	}

	type strategy struct {
		key       string
		value     *string
		matchedBy string
	}

	strategies := []strategy{
		{"AnyProviderIdEquals", imdbIDParam(item.ImdbID), "imdb_id"},
		{"AnyProviderIdEquals", tmdbIDParam(item.TmdbID), "tmdb_id"},
		{"AnyProviderIdEquals", tvdbIDParam(item.TvdbID), "tvdb_id"},
	}

	for _, s := range strategies {
		if s.value == nil {
			continue
		}
		ji, err := p.searchByProviderID(userID, *s.value)
		if err != nil {
			return nil, err
		}
		if ji != nil {
			matched := p.toMediaItem(*ji)
			return &plugins.LibraryMatch{
				PluginID:  p.id,
				Item:      matched,
				MatchedBy: s.matchedBy,
			}, nil
		}
	}

	// Fallback: title + year
	ji, err := p.searchByTitleYear(userID, item.Title, item.Year)
	if err != nil {
		return nil, err
	}
	if ji != nil {
		matched := p.toMediaItem(*ji)
		return &plugins.LibraryMatch{
			PluginID:  p.id,
			Item:      matched,
			MatchedBy: "title_year",
		}, nil
	}

	return nil, nil
}

// ---------------------------------------------------------------------------
// LibraryResumeReadable
// ---------------------------------------------------------------------------

// GetContinueWatching returns items the user has partially watched.
// It looks for the Jellyfin user ID in user.ExternalAccounts.
// If no account link exists for this plugin, it returns an empty slice (not an error).
func (p *Plugin) GetContinueWatching(user plugins.MortarUser, opts plugins.ContinueWatchingOptions) ([]plugins.ContinueWatchingItem, error) {
	// Find the Jellyfin user ID from the external account link.
	jellyfinUserID := ""
	for _, ext := range user.ExternalAccounts {
		if ext.PluginID == p.id {
			jellyfinUserID = ext.ExternalUserID
			break
		}
	}
	if jellyfinUserID == "" {
		// No account link — return empty, not an error.
		return []plugins.ContinueWatchingItem{}, nil
	}

	params := url.Values{}
	params.Set("Fields", "Overview,Genres,ProviderIds")
	params.Set("Recursive", "true")
	if opts.Limit != nil && *opts.Limit > 0 {
		params.Set("Limit", strconv.Itoa(*opts.Limit))
	}

	path := fmt.Sprintf("/Users/%s/Items/Resume", jellyfinUserID)
	resp, err := p.doGet(path, params)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: continue watching request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyfin: continue watching returned HTTP %d", resp.StatusCode)
	}

	var jellyResp jellyfinItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jellyResp); err != nil {
		return nil, fmt.Errorf("jellyfin: continue watching decode error: %w", err)
	}

	result := make([]plugins.ContinueWatchingItem, 0, len(jellyResp.Items))
	for _, ji := range jellyResp.Items {
		cwi := p.toContinueWatchingItem(ji)
		result = append(result, cwi)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// doGet performs an authenticated GET request to the Jellyfin API.
// The API key is passed via the X-MediaBrowser-Token header (never in logs/URLs).
func (p *Plugin) doGet(path string, params url.Values) (*http.Response, error) {
	rawURL := p.baseURL + path
	if len(params) > 0 {
		rawURL += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: build request: %w", err)
	}
	req.Header.Set("X-MediaBrowser-Token", p.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: http error: %w", err)
	}
	return resp, nil
}

// jellyfinPublicInfo holds the fields we need from /System/Info/Public.
type jellyfinPublicInfo struct {
	ID string `json:"Id"`
}

// getServerID returns the cached Jellyfin server ID, fetching it if needed.
func (p *Plugin) getServerID() (string, error) {
	p.serverIDMu.Lock()
	defer p.serverIDMu.Unlock()

	if p.serverID != "" {
		return p.serverID, nil
	}

	resp, err := p.doGet("/System/Info/Public", nil)
	if err != nil {
		return "", fmt.Errorf("jellyfin: fetch server info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("jellyfin: /System/Info/Public returned HTTP %d", resp.StatusCode)
	}

	var info jellyfinPublicInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("jellyfin: decode server info: %w", err)
	}
	if info.ID == "" {
		return "", fmt.Errorf("jellyfin: server info returned empty id")
	}

	p.serverID = info.ID
	return p.serverID, nil
}

// toMediaItem converts a Jellyfin API item into a Mortar MediaItem.
func (p *Plugin) toMediaItem(ji jellyfinItem) plugins.MediaItem {
	mortarType := plugins.MediaTypeMovie
	if ji.Type == "Series" {
		mortarType = plugins.MediaTypeShow
	}

	externalID := ji.ID
	mortarID := p.id + ":" + externalID

	item := plugins.MediaItem{
		ID:         mortarID,
		ExternalID: externalID,
		PluginID:   p.id,
		Type:       mortarType,
		Title:      ji.Name,
		Year:       ji.ProductionYear,
		Overview:   ji.Overview,
		Genres:     ji.Genres,
	}

	// Poster URL
	if _, ok := ji.ImageTags["Primary"]; ok {
		posterURL := fmt.Sprintf("%s/Items/%s/Images/Primary", p.baseURL, ji.ID)
		item.PosterURL = &posterURL
	}

	// External IDs from ProviderIds
	if v, ok := ji.ProviderIds["Imdb"]; ok && v != "" {
		item.ImdbID = &v
	}
	if v, ok := ji.ProviderIds["Tmdb"]; ok && v != "" {
		item.TmdbID = &v
	}
	if v, ok := ji.ProviderIds["Tvdb"]; ok && v != "" {
		item.TvdbID = &v
	}

	return item
}

// toContinueWatchingItem converts a Jellyfin item with user data to a ContinueWatchingItem.
func (p *Plugin) toContinueWatchingItem(ji jellyfinItem) plugins.ContinueWatchingItem {
	item := p.toMediaItem(ji)

	cwi := plugins.ContinueWatchingItem{
		Item: item,
	}

	if ji.UserData != nil {
		if ji.UserData.PlayedPercentage != nil {
			cwi.Progress = *ji.UserData.PlayedPercentage / 100.0
		}
		if ji.UserData.LastPlayedDate != nil {
			cwi.LastWatchedAt = *ji.UserData.LastPlayedDate
		}
		// if LastPlayedDate is empty, leave LastWatchedAt as zero value ("") — caller treats empty as unknown
		// PlaybackPositionTicks are in 100-nanosecond units; convert to seconds.
		if ji.UserData.PlaybackPosition > 0 {
			posSeconds := ji.UserData.PlaybackPosition / 10_000_000
			cwi.PositionSeconds = &posSeconds
		}
	}

	// Duration: RunTimeTicks are 100-nanosecond ticks.
	if ji.RunTimeTicks != nil && *ji.RunTimeTicks > 0 {
		durSeconds := *ji.RunTimeTicks / 10_000_000
		cwi.DurationSeconds = &durSeconds
	}

	return cwi
}

// externalIDFromMortarID extracts the Jellyfin item ID from a Mortar ID
// (format: "pluginID:jellyfinID"). Falls back to the raw string if no colon is found.
func (p *Plugin) externalIDFromMortarID(mortarID string) string {
	if idx := strings.Index(mortarID, ":"); idx >= 0 {
		return mortarID[idx+1:]
	}
	return mortarID
}

// searchByProviderID queries Jellyfin for items with a given provider ID value
// (e.g. "Imdb.tt1234567"). Returns the first match, or nil if none.
func (p *Plugin) searchByProviderID(userID, providerIDParam string) (*jellyfinItem, error) {
	params := url.Values{}
	params.Set("AnyProviderIdEquals", providerIDParam)
	params.Set("Recursive", "true")
	params.Set("Fields", "Overview,Genres,ProviderIds")
	params.Set("Limit", "1")

	path := fmt.Sprintf("/Users/%s/Items", userID)
	resp, err := p.doGet(path, params)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: find match request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyfin: find match returned HTTP %d", resp.StatusCode)
	}

	var jellyResp jellyfinItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jellyResp); err != nil {
		return nil, fmt.Errorf("jellyfin: find match decode error: %w", err)
	}

	if len(jellyResp.Items) == 0 {
		return nil, nil
	}
	return &jellyResp.Items[0], nil
}

// searchByTitleYear queries Jellyfin for items with matching title and optional year.
func (p *Plugin) searchByTitleYear(userID, title string, year *int) (*jellyfinItem, error) {
	params := url.Values{}
	params.Set("SearchTerm", title)
	params.Set("Recursive", "true")
	params.Set("Fields", "Overview,Genres,ProviderIds")
	params.Set("Limit", "10")
	params.Set("IncludeItemTypes", "Movie,Series")

	path := fmt.Sprintf("/Users/%s/Items", userID)
	resp, err := p.doGet(path, params)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: title search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyfin: title search returned HTTP %d", resp.StatusCode)
	}

	var jellyResp jellyfinItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jellyResp); err != nil {
		return nil, fmt.Errorf("jellyfin: title search decode error: %w", err)
	}

	normalTitle := strings.ToLower(strings.TrimSpace(title))
	for i, ji := range jellyResp.Items {
		if strings.ToLower(strings.TrimSpace(ji.Name)) != normalTitle {
			continue
		}
		if year != nil && ji.ProductionYear != nil && *ji.ProductionYear != *year {
			continue
		}
		return &jellyResp.Items[i], nil
	}
	return nil, nil
}

// Provider ID param helpers — format expected by Jellyfin's AnyProviderIdEquals filter.
func imdbIDParam(id *string) *string {
	if id == nil || *id == "" {
		return nil
	}
	v := "Imdb." + *id
	return &v
}

func tmdbIDParam(id *string) *string {
	if id == nil || *id == "" {
		return nil
	}
	v := "Tmdb." + *id
	return &v
}

func tvdbIDParam(id *string) *string {
	if id == nil || *id == "" {
		return nil
	}
	v := "Tvdb." + *id
	return &v
}
