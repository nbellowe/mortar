package api_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// healthEntry mirrors the pluginHealthResponse JSON shape.
type healthEntry struct {
	PluginID    string  `json:"plugin_id"`
	PluginType  string  `json:"plugin_type"`
	DisplayName string  `json:"display_name"`
	Status      string  `json:"status"`
	Reachable   bool    `json:"reachable"`
	LatencyMs   int64   `json:"latency_ms"`
	CheckedAt   string  `json:"checked_at"`
	Detail      *string `json:"detail"`
}

func TestPluginHealth_HappyPath(t *testing.T) {
	// Two plugins: one healthy, one degraded.
	healthy := &mockPlugin{
		manifest: plugins.PluginManifest{
			ID:          "jellyfin",
			Type:        "jellyfin",
			DisplayName: "Jellyfin",
		},
		healthResp: plugins.HealthStatus{
			Status:    "healthy",
			Reachable: true,
			LatencyMs: 42,
			CheckedAt: "2026-05-08T10:00:00Z",
		},
	}

	degraded := &mockPlugin{
		manifest: plugins.PluginManifest{
			ID:          "jellyseerr",
			Type:        "jellyseerr",
			DisplayName: "Jellyseerr",
		},
		healthResp: plugins.HealthStatus{
			Status:    "degraded",
			Reachable: true,
			LatencyMs: 3500,
			CheckedAt: "2026-05-08T10:00:00Z",
		},
	}

	reg := buildRegistry(
		map[string]plugins.Plugin{
			"jellyfin":   healthy,
			"jellyseerr": degraded,
		},
		nil,
	)

	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("GET /api/v1/health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	var entries []healthEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	byID := make(map[string]healthEntry)
	for _, e := range entries {
		byID[e.PluginID] = e
	}

	j, ok := byID["jellyfin"]
	if !ok {
		t.Fatal("missing jellyfin entry")
	}
	if j.Status != "healthy" {
		t.Errorf("jellyfin status: got %q, want healthy", j.Status)
	}
	if !j.Reachable {
		t.Error("jellyfin: expected Reachable=true")
	}
	if j.LatencyMs != 42 {
		t.Errorf("jellyfin latency_ms: got %d, want 42", j.LatencyMs)
	}
	if j.DisplayName != "Jellyfin" {
		t.Errorf("jellyfin display_name: got %q, want Jellyfin", j.DisplayName)
	}

	js, ok := byID["jellyseerr"]
	if !ok {
		t.Fatal("missing jellyseerr entry")
	}
	if js.Status != "degraded" {
		t.Errorf("jellyseerr status: got %q, want degraded", js.Status)
	}
	if !js.Reachable {
		t.Error("jellyseerr: expected Reachable=true")
	}
}

func TestPluginHealth_PluginReturnsError(t *testing.T) {
	// Plugin whose Health() returns an error.
	errPlugin := &mockPlugin{
		manifest: plugins.PluginManifest{
			ID:          "broken",
			Type:        "broken",
			DisplayName: "Broken Plugin",
		},
		healthErr: errors.New("connection refused"),
	}

	reg := buildRegistry(
		map[string]plugins.Plugin{"broken": errPlugin},
		nil,
	)

	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("GET /api/v1/health: %v", err)
	}
	defer resp.Body.Close()

	// Endpoint must always return 200.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK even for erroring plugin, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var entries []healthEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.Status != "unreachable" {
		t.Errorf("status: got %q, want unreachable", e.Status)
	}
	if e.Reachable {
		t.Error("expected Reachable=false")
	}
	if e.Detail == nil {
		t.Error("expected Detail to be set")
	} else if *e.Detail != "connection refused" {
		t.Errorf("detail: got %q, want %q", *e.Detail, "connection refused")
	}
	if e.PluginID != "broken" {
		t.Errorf("plugin_id: got %q, want broken", e.PluginID)
	}
	if e.DisplayName != "Broken Plugin" {
		t.Errorf("display_name: got %q, want Broken Plugin", e.DisplayName)
	}
}

func TestPluginHealth_EmptyRegistry(t *testing.T) {
	reg := buildRegistry(nil, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("GET /api/v1/health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var entries []healthEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty array, got %d entries", len(entries))
	}
}
