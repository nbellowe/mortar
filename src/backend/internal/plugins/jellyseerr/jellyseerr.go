// Package jellyseerr implements the Mortar plugin for Jellyseerr, a media
// request management service. It declares the requests.video and activity.read
// capabilities.
package jellyseerr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// Plugin is the Jellyseerr plugin implementation. It satisfies plugins.Plugin,
// plugins.Requester, and plugins.ActivityReadable.
type Plugin struct {
	id          string
	baseURL     string
	externalURL string // browser-accessible URL for admin review links; falls back to baseURL
	httpClient  *http.Client
	// apiKey is used only in outbound request headers; never logged or returned.
	apiKey string
}

// New constructs a Jellyseerr Plugin from its config entry. Returns an error if
// the config is missing required fields.
func New(cfg config.PluginConfig) (plugins.Plugin, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("jellyseerr: plugin %q is missing a url", cfg.ID)
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("jellyseerr: plugin %q is missing an api_key", cfg.ID)
	}

	baseURL := strings.TrimRight(cfg.URL, "/")
	externalURL := strings.TrimRight(cfg.ExternalURL, "/")
	if externalURL == "" {
		externalURL = baseURL
	}
	return &Plugin{
		id:          cfg.ID,
		baseURL:     baseURL,
		externalURL: externalURL,
		apiKey:      cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// ---------------------------------------------------------------------------
// Plugin base interface
// ---------------------------------------------------------------------------

// Manifest returns the plugin's identity and declared capabilities.
func (p *Plugin) Manifest() plugins.PluginManifest {
	return plugins.PluginManifest{
		ID:          p.id,
		Type:        "jellyseerr",
		DisplayName: "Jellyseerr",
		Capabilities: []plugins.Capability{
			plugins.CapabilityRequestsVideo,
			plugins.CapabilityActivityRead,
		},
	}
}

// Health performs a live health check against the Jellyseerr /api/v1/status
// endpoint and returns a HealthStatus. It never returns the API key in any
// field of the response.
func (p *Plugin) Health() (plugins.HealthStatus, error) {
	start := time.Now()
	checkedAt := start.UTC().Format(time.RFC3339)

	resp, err := p.get("/api/v1/status")
	if err != nil {
		detail := err.Error()
		return plugins.HealthStatus{
			Status:    "unreachable",
			Reachable: false,
			LatencyMs: 0,
			CheckedAt: checkedAt,
			Detail:    &detail,
		}, nil
	}
	defer resp.Body.Close()

	latencyMs := time.Since(start).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		detail := fmt.Sprintf("unexpected status %d", resp.StatusCode)
		return plugins.HealthStatus{
			Status:    "unreachable",
			Reachable: false,
			LatencyMs: latencyMs,
			CheckedAt: checkedAt,
			Detail:    &detail,
		}, nil
	}

	status := "healthy"
	if latencyMs > 2000 {
		status = "degraded"
	}

	return plugins.HealthStatus{
		Status:    status,
		Reachable: true,
		LatencyMs: latencyMs,
		CheckedAt: checkedAt,
	}, nil
}

// ---------------------------------------------------------------------------
// Requester interface (requests.video)
// ---------------------------------------------------------------------------

// Search fans the query to the Jellyseerr search endpoint and returns
// normalized MediaItems.
func (p *Plugin) Search(query string) ([]plugins.MediaItem, error) {
	// Jellyseerr requires %20 for spaces, not the + that url.QueryEscape produces.
	endpoint := fmt.Sprintf("/api/v1/search?query=%s&page=1", strings.ReplaceAll(url.QueryEscape(query), "+", "%20"))
	resp, err := p.get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("jellyseerr: search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyseerr: search returned status %d", resp.StatusCode)
	}

	var body jsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("jellyseerr: search decode failed: %w", err)
	}

	items := make([]plugins.MediaItem, 0, len(body.Results))
	for _, r := range body.Results {
		item := p.mapSearchResult(r)
		items = append(items, item)
	}
	return items, nil
}

// GetRequest fetches a single request by its Mortar-internal ID
// ("jellyseerr:<numericId>").
func (p *Plugin) GetRequest(id string) (*plugins.Request, error) {
	numericID, err := extractNumericID(p.id, id)
	if err != nil {
		return nil, err
	}

	resp, err := p.get("/api/v1/request/" + numericID)
	if err != nil {
		return nil, fmt.Errorf("jellyseerr: get request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyseerr: get request returned status %d", resp.StatusCode)
	}

	var jsReq jsRequest
	if err := json.NewDecoder(resp.Body).Decode(&jsReq); err != nil {
		return nil, fmt.Errorf("jellyseerr: get request decode failed: %w", err)
	}

	req := p.mapRequest(jsReq)
	return &req, nil
}

// ListRequests returns all requests, optionally filtered by status or
// requesterID. Jellyseerr paginates at 20 items; this implementation fetches
// up to 200 requests (10 pages) before stopping to avoid unbounded loops.
func (p *Plugin) ListRequests(opts plugins.ListRequestsOptions) ([]plugins.Request, error) {
	const pageSize = 20
	const maxPages = 10

	var all []plugins.Request

	for page := 0; page < maxPages; page++ {
		skip := page * pageSize
		endpoint := fmt.Sprintf("/api/v1/request?take=%d&skip=%d&sort=added", pageSize, skip)
		resp, err := p.get(endpoint)
		if err != nil {
			return nil, fmt.Errorf("jellyseerr: list requests failed: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("jellyseerr: list requests returned HTTP %d", resp.StatusCode)
		}

		var body jsRequestListResponse
		decErr := json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if decErr != nil {
			return nil, fmt.Errorf("jellyseerr: list requests decode failed: %w", decErr)
		}

		for _, jsReq := range body.Results {
			req := p.mapRequest(jsReq)

			// Apply optional filters.
			if opts.RequesterID != nil && req.RequesterID != *opts.RequesterID {
				continue
			}
			if opts.Status != nil && req.Status != *opts.Status {
				continue
			}

			all = append(all, req)
		}

		// Stop if this page was the last.
		if len(body.Results) < pageSize {
			break
		}
	}

	return all, nil
}

// SubmitRequest submits a new media request to Jellyseerr.
func (p *Plugin) SubmitRequest(item plugins.MediaItem, requester plugins.MortarUser) (plugins.Request, error) {
	if item.TmdbID == nil {
		return plugins.Request{}, fmt.Errorf("jellyseerr: cannot submit request without a tmdb_id")
	}

	tmdbID, err := strconv.Atoi(*item.TmdbID)
	if err != nil {
		return plugins.Request{}, fmt.Errorf("jellyseerr: invalid tmdb_id %q: %w", *item.TmdbID, err)
	}

	var payload interface{}
	switch item.Type {
	case plugins.MediaTypeMovie:
		payload = map[string]interface{}{
			"mediaType": "movie",
			"mediaId":   tmdbID,
		}
	case plugins.MediaTypeShow:
		payload = map[string]interface{}{
			"mediaType": "tv",
			"mediaId":   tmdbID,
			"seasons":   "all",
		}
	default:
		return plugins.Request{}, fmt.Errorf("jellyseerr: unsupported media type %q for request", item.Type)
	}

	resp, err := p.post("/api/v1/request", payload)
	if err != nil {
		return plugins.Request{}, fmt.Errorf("jellyseerr: submit request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return plugins.Request{}, fmt.Errorf("jellyseerr: submit request returned status %d: %s", resp.StatusCode, string(body))
	}

	var jsReq jsRequest
	if err := json.NewDecoder(resp.Body).Decode(&jsReq); err != nil {
		return plugins.Request{}, fmt.Errorf("jellyseerr: submit request decode failed: %w", err)
	}

	return p.mapRequest(jsReq), nil
}

// ReviewRequest approves or declines an existing request.
func (p *Plugin) ReviewRequest(id string, review plugins.RequestReview) (plugins.Request, error) {
	numericID, err := extractNumericID(p.id, id)
	if err != nil {
		return plugins.Request{}, err
	}

	var endpoint string
	var payload interface{}

	switch review.Decision {
	case "approve":
		endpoint = fmt.Sprintf("/api/v1/request/%s/approve", numericID)
		payload = map[string]interface{}{}
	case "decline":
		endpoint = fmt.Sprintf("/api/v1/request/%s/decline", numericID)
		if review.Reason != nil {
			payload = map[string]interface{}{"reason": *review.Reason}
		} else {
			payload = map[string]interface{}{}
		}
	default:
		return plugins.Request{}, fmt.Errorf("jellyseerr: unknown review decision %q", review.Decision)
	}

	resp, err := p.post(endpoint, payload)
	if err != nil {
		return plugins.Request{}, fmt.Errorf("jellyseerr: review request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return plugins.Request{}, fmt.Errorf("jellyseerr: review request returned status %d: %s", resp.StatusCode, string(body))
	}

	var jsReq jsRequest
	if err := json.NewDecoder(resp.Body).Decode(&jsReq); err != nil {
		return plugins.Request{}, fmt.Errorf("jellyseerr: review request decode failed: %w", err)
	}

	return p.mapRequest(jsReq), nil
}

// ReviewURL returns the upstream request-management surface for admins.
func (p *Plugin) ReviewURL(_ string) string {
	return p.externalURL + "/requests"
}

// ---------------------------------------------------------------------------
// ActivityReadable interface (activity.read)
// ---------------------------------------------------------------------------

// GetActivity returns activity events derived from the last 50 Jellyseerr
// requests, optionally filtered to events after `since` (ISO 8601).
func (p *Plugin) GetActivity(since *string) ([]plugins.ActivityEvent, error) {
	resp, err := p.get("/api/v1/request?take=50&skip=0&sort=added")
	if err != nil {
		return nil, fmt.Errorf("jellyseerr: get activity request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyseerr: get activity returned status %d", resp.StatusCode)
	}

	var body jsRequestListResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("jellyseerr: get activity decode failed: %w", err)
	}

	var sinceTime *time.Time
	if since != nil && *since != "" {
		t, err := time.Parse(time.RFC3339, *since)
		if err != nil {
			return nil, fmt.Errorf("jellyseerr: invalid since value %q: %w", *since, err)
		}
		sinceTime = &t
	}

	events := make([]plugins.ActivityEvent, 0, len(body.Results))
	for _, jsReq := range body.Results {
		ts, err := time.Parse(time.RFC3339, jsReq.UpdatedAt)
		if err != nil {
			// Fall back to createdAt if updatedAt is not parseable.
			ts, _ = time.Parse(time.RFC3339, jsReq.CreatedAt)
		}

		// Filter by since timestamp (>= semantics: include events at exactly since).
		if sinceTime != nil && ts.Before(*sinceTime) {
			continue
		}

		event := p.mapRequestToActivityEvent(jsReq, ts)
		events = append(events, event)
	}

	return events, nil
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func (p *Plugin) get(path string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, p.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", p.apiKey)
	return p.httpClient.Do(req)
}

func (p *Plugin) post(path string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, p.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return p.httpClient.Do(req)
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

// mapSearchResult converts a Jellyseerr search result to a plugins.MediaItem.
func (p *Plugin) mapSearchResult(r jsSearchResult) plugins.MediaItem {
	externalID := strconv.Itoa(r.ID)
	mortarID := p.id + ":" + externalID

	var mediaType plugins.MediaType
	if r.MediaType == "movie" {
		mediaType = plugins.MediaTypeMovie
	} else {
		mediaType = plugins.MediaTypeShow
	}

	title := r.Title
	if title == "" {
		title = r.Name
	}

	item := plugins.MediaItem{
		ID:         mortarID,
		ExternalID: externalID,
		PluginID:   p.id,
		Type:       mediaType,
		Title:      title,
	}

	// Year from release date or first air date.
	dateStr := r.ReleaseDate
	if dateStr == "" {
		dateStr = r.FirstAirDate
	}
	if len(dateStr) >= 4 {
		if y, err := strconv.Atoi(dateStr[:4]); err == nil {
			item.Year = &y
		}
	}

	if r.Overview != "" {
		ov := r.Overview
		item.Overview = &ov
	}

	if r.PosterPath != "" {
		posterURL := "https://image.tmdb.org/t/p/w500" + r.PosterPath
		item.PosterURL = &posterURL
	}

	// In Jellyseerr search results the "id" field is the TMDB ID. There is no
	// separate "tmdbId" field in the search response — set TmdbID from r.ID so
	// that SubmitRequest can forward it to Jellyseerr's request API.
	tmdbIDStr := strconv.Itoa(r.ID)
	item.TmdbID = &tmdbIDStr

	if r.ImdbID != "" {
		imdb := r.ImdbID
		item.ImdbID = &imdb
	}

	return item
}

// fetchMediaTitle looks up the title for a media item from Jellyseerr's
// movie or TV detail endpoint. Returns an empty string if the lookup fails
// so callers can fall back gracefully.
func (p *Plugin) fetchMediaTitle(tmdbID int, mediaType string) string {
	if tmdbID == 0 {
		return ""
	}
	var path string
	if mediaType == "movie" {
		path = fmt.Sprintf("/api/v1/movie/%d", tmdbID)
	} else {
		path = fmt.Sprintf("/api/v1/tv/%d", tmdbID)
	}

	resp, err := p.get(path)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var detail struct {
		Title string `json:"title"` // movies
		Name  string `json:"name"`  // TV shows
	}
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return ""
	}
	if detail.Title != "" {
		return detail.Title
	}
	return detail.Name
}

// mapMediaFromRequest converts a jsMedia embedded in a request to a
// plugins.MediaItem.
func (p *Plugin) mapMediaFromRequest(m jsMedia) plugins.MediaItem {
	externalID := strconv.Itoa(m.ID)
	mortarID := p.id + ":" + externalID

	var mediaType plugins.MediaType
	if m.MediaType == "movie" {
		mediaType = plugins.MediaTypeMovie
	} else {
		mediaType = plugins.MediaTypeShow
	}

	title := m.Title
	if title == "" && m.TmdbID != 0 {
		title = p.fetchMediaTitle(m.TmdbID, m.MediaType)
	}

	item := plugins.MediaItem{
		ID:         mortarID,
		ExternalID: externalID,
		PluginID:   p.id,
		Type:       mediaType,
		Title:      title,
	}

	if m.TmdbID != 0 {
		tmdbStr := strconv.Itoa(m.TmdbID)
		item.TmdbID = &tmdbStr
	}

	if m.ImdbID != "" {
		imdb := m.ImdbID
		item.ImdbID = &imdb
	}

	return item
}

// mapRequest converts a jsRequest to a plugins.Request.
func (p *Plugin) mapRequest(jsReq jsRequest) plugins.Request {
	id := strconv.Itoa(jsReq.ID)
	mortarID := p.id + ":" + id

	return plugins.Request{
		ID:          mortarID,
		PluginID:    p.id,
		Item:        p.mapMediaFromRequest(jsReq.Media),
		RequesterID: strconv.Itoa(jsReq.RequestedBy.ID),
		Status:      mapRequestStatus(jsReq.Status),
		SubmittedAt: jsReq.CreatedAt,
		UpdatedAt:   jsReq.UpdatedAt,
	}
}

// mapRequestToActivityEvent converts a jsRequest to a plugins.ActivityEvent.
func (p *Plugin) mapRequestToActivityEvent(jsReq jsRequest, ts time.Time) plugins.ActivityEvent {
	id := strconv.Itoa(jsReq.ID)
	eventID := p.id + ":" + id

	eventType, visibility := activityEventTypeAndVisibility(jsReq.Status)

	actorID := strconv.Itoa(jsReq.RequestedBy.ID)
	var actorPtr *string
	if actorID != "" && actorID != "0" {
		actorPtr = &actorID
	}

	item := p.mapMediaFromRequest(jsReq.Media)
	message := buildActivityMessage(jsReq, eventType)

	return plugins.ActivityEvent{
		ID:           eventID,
		SourcePlugin: p.id,
		Type:         eventType,
		Item:         &item,
		Message:      message,
		Timestamp:    ts.UTC().Format(time.RFC3339),
		Visibility:   visibility,
		ActorUserID:  actorPtr,
	}
}

// buildActivityMessage generates a human-readable message for an activity event.
func buildActivityMessage(jsReq jsRequest, eventType plugins.ActivityEventType) string {
	title := jsReq.Media.Title
	if title == "" {
		title = "unknown media"
	}

	switch eventType {
	case plugins.ActivityEventApproved:
		return fmt.Sprintf("Request for %q was approved", title)
	case plugins.ActivityEventDeclined:
		return fmt.Sprintf("Request for %q was declined", title)
	case plugins.ActivityEventAddedToLibrary:
		return fmt.Sprintf("%q is now available in the library", title)
	default:
		return fmt.Sprintf("%q was requested", title)
	}
}

// activityEventTypeAndVisibility maps a Jellyseerr request status integer to
// the appropriate ActivityEventType and ActivityVisibility.
//
// Jellyseerr status codes:
//
//	1 = pending, 2 = approved, 3 = declined, 4 = available, 5 = failed
func activityEventTypeAndVisibility(status int) (plugins.ActivityEventType, plugins.ActivityVisibility) {
	switch status {
	case 2:
		return plugins.ActivityEventApproved, plugins.ActivityVisibilityRequesterAndAdmin
	case 3:
		return plugins.ActivityEventDeclined, plugins.ActivityVisibilityRequesterAndAdmin
	case 4:
		return plugins.ActivityEventAddedToLibrary, plugins.ActivityVisibilityAllUsers
	default:
		// 1 = pending, 5 = failed, anything else = treat as requested
		return plugins.ActivityEventRequested, plugins.ActivityVisibilityRequesterAndAdmin
	}
}

// mapRequestStatus converts a Jellyseerr status integer to a RequestStatus.
func mapRequestStatus(status int) plugins.RequestStatus {
	switch status {
	case 1:
		return plugins.RequestStatusPending
	case 2:
		return plugins.RequestStatusApproved
	case 3:
		return plugins.RequestStatusDeclined
	case 4:
		return plugins.RequestStatusAvailable
	case 5:
		return plugins.RequestStatusFailed
	default:
		return plugins.RequestStatusPending
	}
}

// extractNumericID strips the "pluginID:" prefix from a Mortar-internal ID
// and returns the raw numeric ID string.
func extractNumericID(pluginID, mortarID string) (string, error) {
	prefix := pluginID + ":"
	if !strings.HasPrefix(mortarID, prefix) {
		return "", fmt.Errorf("jellyseerr: id %q does not belong to plugin %q", mortarID, pluginID)
	}
	return strings.TrimPrefix(mortarID, prefix), nil
}

// ---------------------------------------------------------------------------
// Jellyseerr API response types (internal)
// ---------------------------------------------------------------------------

type jsSearchResponse struct {
	Results []jsSearchResult `json:"results"`
}

type jsSearchResult struct {
	ID           int    `json:"id"`
	TmdbID       int    `json:"tmdbId"`
	MediaType    string `json:"mediaType"`
	Title        string `json:"title"`        // movies
	Name         string `json:"name"`         // TV shows
	ReleaseDate  string `json:"releaseDate"`  // movies
	FirstAirDate string `json:"firstAirDate"` // TV shows
	Overview     string `json:"overview"`
	PosterPath   string `json:"posterPath"`
	ImdbID       string `json:"imdbId"`
}

type jsRequestListResponse struct {
	Results  []jsRequest `json:"results"`
	PageInfo struct {
		Pages    int `json:"pages"`
		PageSize int `json:"pageSize"`
		Results  int `json:"results"`
		Skip     int `json:"skip"`
	} `json:"pageInfo"`
}

type jsRequest struct {
	ID          int     `json:"id"`
	Status      int     `json:"status"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	Media       jsMedia `json:"media"`
	RequestedBy jsUser  `json:"requestedBy"`
}

type jsMedia struct {
	ID        int    `json:"id"`
	MediaType string `json:"mediaType"`
	TmdbID    int    `json:"tmdbId"`
	ImdbID    string `json:"imdbId"`
	Title     string `json:"title"`
}

type jsUser struct {
	ID int `json:"id"`
}
