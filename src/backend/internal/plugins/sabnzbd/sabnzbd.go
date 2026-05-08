// Package sabnzbd implements the Mortar plugin for SABnzbd, a Usenet download
// client. It declares the downloads.read capability and exposes the current
// download queue as normalised DownloadItem values.
package sabnzbd

import (
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

// Plugin implements plugins.Plugin and plugins.DownloadsReadable for SABnzbd.
type Plugin struct {
	id         string
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New constructs a SABnzbd plugin from the provided config entry.
// It returns an error if the base URL is missing.
func New(cfg config.PluginConfig) (plugins.Plugin, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("sabnzbd: plugin %q requires a url", cfg.ID)
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("sabnzbd: plugin %q requires an api_key", cfg.ID)
	}
	return &Plugin{
		id:      cfg.ID,
		baseURL: strings.TrimRight(cfg.URL, "/"),
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Manifest returns the plugin's identity and declared capabilities.
func (p *Plugin) Manifest() plugins.PluginManifest {
	return plugins.PluginManifest{
		ID:           p.id,
		Type:         "sabnzbd",
		DisplayName:  "SABnzbd",
		Capabilities: []plugins.Capability{plugins.CapabilityDownloadsRead},
	}
}

// Health checks reachability by calling the SABnzbd version endpoint.
func (p *Plugin) Health() (plugins.HealthStatus, error) {
	start := time.Now()
	checkedAt := start.UTC().Format(time.RFC3339)

	resp, err := p.httpClient.Get(p.apiURL("version"))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		detail := fmt.Sprintf("failed to read response: %v", err)
		return plugins.HealthStatus{
			Status:    "unreachable",
			Reachable: false,
			LatencyMs: latencyMs,
			CheckedAt: checkedAt,
			Detail:    &detail,
		}, nil
	}

	var versionResp struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &versionResp); err != nil || versionResp.Version == "" {
		detail := "version field missing from response"
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

// GetQueue returns all items currently in the SABnzbd download queue.
func (p *Plugin) GetQueue() ([]plugins.DownloadItem, error) {
	resp, err := p.httpClient.Get(p.apiURL("queue"))
	if err != nil {
		return nil, fmt.Errorf("sabnzbd: queue request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sabnzbd: queue returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sabnzbd: failed to read queue response: %w", err)
	}

	var queueResp sabQueueResponse
	if err := json.Unmarshal(body, &queueResp); err != nil {
		return nil, fmt.Errorf("sabnzbd: failed to parse queue response: %w", err)
	}

	// Parse queue-level speed (kbps) once — shared across all slots.
	speedBytesS := parseKbps(queueResp.Queue.Kbps)

	items := make([]plugins.DownloadItem, 0, len(queueResp.Queue.Slots))
	for _, slot := range queueResp.Queue.Slots {
		item := plugins.DownloadItem{
			ID:           fmt.Sprintf("%s:%s", p.id, slot.NzoID),
			Name:         slot.Filename,
			Progress:     parseProgress(slot.Percentage),
			SizeBytes:    parseMB(slot.MB),
			SpeedBytesS:  speedBytesS,
			EtaSeconds:   parseTimeleft(slot.Timeleft),
			Status:       mapStatus(slot.Status),
			SourcePlugin: nil,
		}
		items = append(items, item)
	}

	return items, nil
}

// apiURL builds the SABnzbd API URL for a given mode.
// The API key is placed in the query string as required by SABnzbd's API.
func (p *Plugin) apiURL(mode string) string {
	u := fmt.Sprintf("%s/api", p.baseURL)
	params := url.Values{}
	params.Set("output", "json")
	params.Set("apikey", p.apiKey)
	params.Set("mode", mode)
	return u + "?" + params.Encode()
}

// ---------------------------------------------------------------------------
// SABnzbd API response types
// ---------------------------------------------------------------------------

// sabQueueResponse is the top-level response from mode=queue.
type sabQueueResponse struct {
	Queue sabQueue `json:"queue"`
}

type sabQueue struct {
	// Kbps is the current aggregate download speed in kilobytes per second.
	// SABnzbd may return this as a string or number across API versions.
	Kbps  json.RawMessage `json:"kbps"`
	Slots []sabSlot       `json:"slots"`
}

type sabSlot struct {
	NzoID      string          `json:"nzo_id"`
	Filename   string          `json:"filename"`
	Status     string          `json:"status"`
	Percentage json.RawMessage `json:"percentage"` // string or number
	MB         json.RawMessage `json:"mb"`         // total MB, string or number
	MBLeft     json.RawMessage `json:"mbleft"`     // remaining MB, string or number
	Timeleft   string          `json:"timeleft"`   // "HH:MM:SS"
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

// parseFloat64Raw parses a json.RawMessage that may be a JSON string or number
// into a float64. Returns 0 on any error.
func parseFloat64Raw(raw json.RawMessage) float64 {
	if len(raw) == 0 {
		return 0
	}

	// Try plain number first.
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return f
	}

	// Try quoted string.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		f, _ = strconv.ParseFloat(strings.TrimSpace(s), 64)
		return f
	}

	return 0
}

// parseProgress converts a percentage raw value (0–100) to a 0.0–1.0 float.
func parseProgress(raw json.RawMessage) float64 {
	pct := parseFloat64Raw(raw)
	return pct / 100.0
}

// parseMB converts a megabyte raw value to bytes as int64.
func parseMB(raw json.RawMessage) int64 {
	mb := parseFloat64Raw(raw)
	return int64(mb * 1024 * 1024)
}

// parseKbps converts a kilobytes-per-second raw value to bytes-per-second as int64.
func parseKbps(raw json.RawMessage) int64 {
	kbps := parseFloat64Raw(raw)
	return int64(kbps * 1024)
}

// parseTimeleft parses a "HH:MM:SS" duration string into total seconds.
// Returns nil if the string is empty or represents zero duration.
func parseTimeleft(s string) *int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0:00:00" {
		return nil
	}

	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return nil
	}

	h, err1 := strconv.ParseInt(parts[0], 10, 64)
	m, err2 := strconv.ParseInt(parts[1], 10, 64)
	sec, err3 := strconv.ParseInt(parts[2], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}

	total := h*3600 + m*60 + sec
	if total == 0 {
		return nil
	}
	return &total
}

// mapStatus converts SABnzbd slot status strings to Mortar's canonical values.
func mapStatus(s string) string {
	switch s {
	case "Downloading":
		return "downloading"
	case "Paused":
		return "paused"
	case "Queued":
		return "queued"
	case "Extracting", "Verifying", "Repairing", "Moving":
		return "processing"
	case "Failed":
		return "failed"
	default:
		return "queued"
	}
}
