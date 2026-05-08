// Package sonarr implements the Mortar plugin for Sonarr (TV show management).
// It declares the activity.read capability, exposing TV show download and
// import history as Mortar ActivityEvents.
package sonarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	plugins "github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// Plugin is the Sonarr plugin instance.
type Plugin struct {
	id         string
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New constructs a Sonarr Plugin from a PluginConfig.
// The config URL is used as the Sonarr base URL; api_key must be provided.
func New(cfg config.PluginConfig) (plugins.Plugin, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("sonarr: plugin %q requires api_key", cfg.ID)
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("sonarr: plugin %q requires url", cfg.ID)
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

// Manifest implements plugins.Plugin.
func (p *Plugin) Manifest() plugins.PluginManifest {
	return plugins.PluginManifest{
		ID:           p.id,
		Type:         "sonarr",
		DisplayName:  "Sonarr",
		Capabilities: []plugins.Capability{plugins.CapabilityActivityRead},
	}
}

// Health implements plugins.Plugin by calling the Sonarr system/status endpoint.
func (p *Plugin) Health() (plugins.HealthStatus, error) {
	start := time.Now()
	checkedAt := start.UTC().Format(time.RFC3339)

	req, err := http.NewRequest(http.MethodGet, p.baseURL+"/api/v3/system/status", nil)
	if err != nil {
		return unreachable(checkedAt, err.Error()), nil
	}
	req.Header.Set("X-Api-Key", p.apiKey)

	resp, err := p.httpClient.Do(req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		return unreachable(checkedAt, err.Error()), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail := fmt.Sprintf("unexpected status %d", resp.StatusCode)
		return unreachable(checkedAt, detail), nil
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

// unreachable builds a HealthStatus for an unreachable Sonarr instance.
func unreachable(checkedAt, detail string) plugins.HealthStatus {
	return plugins.HealthStatus{
		Status:    "unreachable",
		Reachable: false,
		LatencyMs: 0,
		CheckedAt: checkedAt,
		Detail:    &detail,
	}
}

// ---------------------------------------------------------------------------
// ActivityReadable
// ---------------------------------------------------------------------------

// sonarrHistoryResponse is the top-level response from GET /api/v3/history.
type sonarrHistoryResponse struct {
	Records []sonarrHistoryRecord `json:"records"`
}

// sonarrHistoryRecord is one entry from the Sonarr history API.
type sonarrHistoryRecord struct {
	ID          int                  `json:"id"`
	EpisodeID   int                  `json:"episodeId"`
	SourceTitle string               `json:"sourceTitle"`
	Date        string               `json:"date"`
	EventType   string               `json:"eventType"`
	Series      sonarrSeriesSummary  `json:"series"`
	Episode     sonarrEpisodeSummary `json:"episode"`
}

type sonarrSeriesSummary struct {
	Title  string `json:"title"`
	TvdbID int    `json:"tvdbId"`
}

type sonarrEpisodeSummary struct {
	Title string `json:"title"`
}

// GetActivity implements plugins.ActivityReadable.
// It fetches the 50 most recent history events from Sonarr and returns them
// as ActivityEvents. If since is non-nil, events older than that ISO 8601
// timestamp are filtered out.
func (p *Plugin) GetActivity(since *string) ([]plugins.ActivityEvent, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		p.baseURL+"/api/v3/history?pageSize=50&sortKey=date&sortDirection=descending",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("sonarr: build history request: %w", err)
	}
	req.Header.Set("X-Api-Key", p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sonarr: fetch history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sonarr: history returned status %d", resp.StatusCode)
	}

	var body sonarrHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("sonarr: decode history response: %w", err)
	}

	var sinceTime *time.Time
	if since != nil && *since != "" {
		t, err := time.Parse(time.RFC3339Nano, *since)
		if err != nil {
			t, err = time.Parse(time.RFC3339, *since)
			if err != nil {
				// Try without nanoseconds / with varying precision.
				t, err = time.Parse("2006-01-02T15:04:05Z", *since)
				if err != nil {
					return nil, fmt.Errorf("sonarr: invalid since timestamp %q: %w", *since, err)
				}
			}
		}
		sinceTime = &t
	}

	events := make([]plugins.ActivityEvent, 0, len(body.Records))
	for _, r := range body.Records {
		if sinceTime != nil {
			recordTime, err := time.Parse(time.RFC3339, r.Date)
			if err != nil {
				// Try alternate format (Sonarr sometimes omits sub-seconds).
				recordTime, err = time.Parse("2006-01-02T15:04:05Z", r.Date)
				if err != nil {
					// Skip unparseable dates rather than aborting.
					continue
				}
			}
			if !recordTime.After(*sinceTime) {
				continue
			}
		}

		event, ok := p.recordToEvent(r)
		if !ok {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// recordToEvent maps a Sonarr history record to an ActivityEvent.
// Returns (event, false) if the eventType is not a known Mortar ActivityEventType.
func (p *Plugin) recordToEvent(r sonarrHistoryRecord) (plugins.ActivityEvent, bool) {
	eventType, ok := mapEventType(r.EventType)
	if !ok {
		return plugins.ActivityEvent{}, false
	}

	externalID := strconv.Itoa(r.EpisodeID)
	mediaItem := &plugins.MediaItem{
		ID:         p.id + ":" + externalID,
		ExternalID: externalID,
		PluginID:   p.id,
		Type:       plugins.MediaTypeShow,
		Title:      r.Series.Title,
		Genres:     []string{},
	}
	if r.Series.TvdbID != 0 {
		tvdbID := strconv.Itoa(r.Series.TvdbID)
		mediaItem.TvdbID = &tvdbID
	}

	return plugins.ActivityEvent{
		ID:           fmt.Sprintf("%s-%d", p.id, r.ID),
		SourcePlugin: p.id,
		Type:         eventType,
		Item:         mediaItem,
		Message:      r.SourceTitle,
		Timestamp:    r.Date,
		Visibility:   plugins.ActivityVisibilityAllUsers,
	}, true
}

// mapEventType converts a Sonarr eventType string to a Mortar ActivityEventType.
// Returns (type, true) for known types; ("", false) for any unrecognised eventType.
func mapEventType(sonarrType string) (plugins.ActivityEventType, bool) {
	switch sonarrType {
	case "grabbed":
		return plugins.ActivityEventDownloaded, true
	case "downloadFolderImported", "downloadImported":
		return plugins.ActivityEventAddedToLibrary, true
	case "seriesDelete", "episodeFileDelete":
		return plugins.ActivityEventDeleted, true
	case "failed":
		return plugins.ActivityEventFailed, true
	default:
		return "", false
	}
}
