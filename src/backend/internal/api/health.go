// Package api — plugin health endpoint.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// pluginHealthResponse is the per-plugin entry returned by GET /api/v1/health.
// It embeds the HealthStatus fields and adds plugin identity fields.
type pluginHealthResponse struct {
	PluginID    string  `json:"plugin_id"`
	PluginType  string  `json:"plugin_type"`
	DisplayName string  `json:"display_name"`
	Status      string  `json:"status"`
	Reachable   bool    `json:"reachable"`
	LatencyMs   int64   `json:"latency_ms"`
	CheckedAt   string  `json:"checked_at"`
	Detail      *string `json:"detail"`
}

// handlePluginHealth is the handler for GET /api/v1/health.
// It performs a live health check against every registered plugin and returns
// a JSON array of per-plugin results. The endpoint always returns 200 OK;
// individual plugin failures are surfaced as "unreachable" entries in the
// payload rather than as HTTP errors.
func (h *handler) handlePluginHealth(w http.ResponseWriter, _ *http.Request) {
	results := h.health.list()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(results)
}

// buildHealthEntry calls Health() on a plugin and constructs the response
// entry. If Health() returns an error, the entry is synthesized as
// "unreachable" with the error message in Detail.
func buildHealthEntry(manifest plugins.PluginManifest, p plugins.Plugin) pluginHealthResponse {
	checkedAt := time.Now().UTC().Format(time.RFC3339)
	entry := pluginHealthResponse{
		PluginID:    manifest.ID,
		PluginType:  manifest.Type,
		DisplayName: manifest.DisplayName,
	}

	status, err := p.Health()
	if err != nil {
		detail := err.Error()
		entry.Status = "unreachable"
		entry.Reachable = false
		entry.LatencyMs = 0
		entry.Detail = &detail
		entry.CheckedAt = checkedAt
		return entry
	}

	entry.Status = status.Status
	entry.Reachable = status.Reachable
	entry.LatencyMs = status.LatencyMs
	entry.CheckedAt = status.CheckedAt
	entry.Detail = status.Detail
	return entry
}
