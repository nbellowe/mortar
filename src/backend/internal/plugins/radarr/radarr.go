// Package radarr implements a Mortar plugin for Radarr (v3 API).
// It declares the activity.read capability and exposes movie download
// and import history as ActivityEvents.
package radarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// Plugin is the Radarr plugin implementation.
type Plugin struct {
	id         string
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New constructs a Radarr Plugin from the provided PluginConfig.
func New(cfg config.PluginConfig) (plugins.Plugin, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("radarr: url is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("radarr: api_key is required")
	}
	return &Plugin{
		id:      cfg.ID,
		baseURL: cfg.URL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Manifest returns the plugin's identity and declared capabilities.
func (p *Plugin) Manifest() plugins.PluginManifest {
	return plugins.PluginManifest{
		ID:          p.id,
		Type:        "radarr",
		DisplayName: "Radarr",
		Capabilities: []plugins.Capability{
			plugins.CapabilityActivityRead,
		},
	}
}

// Health calls the Radarr system/status endpoint and derives a HealthStatus.
func (p *Plugin) Health() (plugins.HealthStatus, error) {
	start := time.Now()
	checkedAt := start.UTC().Format(time.RFC3339)

	req, err := http.NewRequest(http.MethodGet, p.baseURL+"/api/v3/system/status", nil)
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
	req.Header.Set("X-Api-Key", p.apiKey)

	resp, err := p.httpClient.Do(req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		detail := err.Error()
		return plugins.HealthStatus{
			Status:    "unreachable",
			Reachable: false,
			LatencyMs: latencyMs,
			CheckedAt: checkedAt,
			Detail:    &detail,
		}, nil
	}
	defer resp.Body.Close()

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

	// Confirm the response body contains a "version" field.
	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		detail := fmt.Sprintf("failed to decode response: %v", err)
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

// radarrHistoryRecord is a single record from GET /api/v3/history.
type radarrHistoryRecord struct {
	ID          int    `json:"id"`
	EventType   string `json:"eventType"`
	SourceTitle string `json:"sourceTitle"`
	Date        string `json:"date"`
	MovieID     int    `json:"movieId"`
	Movie       struct {
		Title  string `json:"title"`
		Year   int    `json:"year"`
		ImdbID string `json:"imdbId"`
		TmdbID int    `json:"tmdbId"`
	} `json:"movie"`
}

// radarrHistoryResponse is the top-level response from GET /api/v3/history.
type radarrHistoryResponse struct {
	Records []radarrHistoryRecord `json:"records"`
}

// mapEventType converts a Radarr history eventType string to an ActivityEventType.
func mapEventType(eventType string) plugins.ActivityEventType {
	switch eventType {
	case "grabbed":
		return plugins.ActivityEventDownloaded
	case "downloadFolderImported", "downloadImported":
		return plugins.ActivityEventAddedToLibrary
	case "movieFileDelete", "movieDelete":
		return plugins.ActivityEventDeleted
	case "failed":
		return plugins.ActivityEventFailed
	default:
		return plugins.ActivityEventDownloaded
	}
}

// GetActivity fetches movie history from Radarr and returns it as ActivityEvents.
// If since is non-nil it is parsed as RFC3339 and records with date < since are excluded.
func (p *Plugin) GetActivity(since *string) ([]plugins.ActivityEvent, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		p.baseURL+"/api/v3/history?pageSize=50&sortKey=date&sortDirection=descending&includeMovie=true",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("radarr: build history request: %w", err)
	}
	req.Header.Set("X-Api-Key", p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("radarr: history request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("radarr: history returned status %d", resp.StatusCode)
	}

	var history radarrHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("radarr: decode history response: %w", err)
	}

	// Parse the since threshold once if provided.
	var sinceTime time.Time
	hasSince := false
	if since != nil {
		t, err := time.Parse(time.RFC3339, *since)
		if err != nil {
			return nil, fmt.Errorf("radarr: parse since %q: %w", *since, err)
		}
		sinceTime = t
		hasSince = true
	}

	events := make([]plugins.ActivityEvent, 0, len(history.Records))
	for _, rec := range history.Records {
		// Apply since filter.
		if hasSince {
			recTime, err := time.Parse(time.RFC3339, rec.Date)
			if err != nil {
				// If we can't parse the date, skip the record rather than error.
				continue
			}
			if recTime.Before(sinceTime) {
				continue
			}
		}

		movieID := rec.MovieID
		imdbID := rec.Movie.ImdbID
		year := rec.Movie.Year

		item := &plugins.MediaItem{
			ID:         fmt.Sprintf("%s:%d", p.id, movieID),
			ExternalID: strconv.Itoa(movieID),
			Type:       plugins.MediaTypeMovie,
			Title:      rec.Movie.Title,
			Year:       &year,
			PluginID:   p.id,
		}
		if rec.Movie.TmdbID != 0 {
			tmdbIDStr := strconv.Itoa(rec.Movie.TmdbID)
			item.TmdbID = &tmdbIDStr
		}
		if imdbID != "" {
			item.ImdbID = &imdbID
		}

		events = append(events, plugins.ActivityEvent{
			ID:           fmt.Sprintf("%s:%d", p.id, rec.ID),
			SourcePlugin: p.id,
			Type:         mapEventType(rec.EventType),
			Item:         item,
			Message:      rec.SourceTitle,
			Timestamp:    rec.Date,
			Visibility:   plugins.ActivityVisibilityAllUsers,
		})
	}

	return events, nil
}
