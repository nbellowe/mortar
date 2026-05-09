package api_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// ---------------------------------------------------------------------------
// mockDownloadsPlugin — plugins.Plugin + plugins.DownloadsReadable
// ---------------------------------------------------------------------------

type mockDownloadsPlugin struct {
	manifest  plugins.PluginManifest
	healthErr error
	items     []plugins.DownloadItem
	queueErr  error
}

func (m *mockDownloadsPlugin) Manifest() plugins.PluginManifest { return m.manifest }
func (m *mockDownloadsPlugin) Health() (plugins.HealthStatus, error) {
	return plugins.HealthStatus{Status: "healthy", Reachable: true}, m.healthErr
}
func (m *mockDownloadsPlugin) GetQueue() ([]plugins.DownloadItem, error) {
	return m.items, m.queueErr
}

// downloadsResp mirrors the JSON shape of downloadsResponse.
type downloadsResp struct {
	Items         []plugins.DownloadItem `json:"items"`
	FailedPlugins []string               `json:"failed_plugins"`
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestDownloads_HappyPath(t *testing.T) {
	p := &mockDownloadsPlugin{
		manifest: plugins.PluginManifest{
			ID:           "sabnzbd",
			Type:         "sabnzbd",
			DisplayName:  "SABnzbd",
			Capabilities: []plugins.Capability{plugins.CapabilityDownloadsRead},
		},
		items: []plugins.DownloadItem{
			{ID: "1", Name: "Show.S01E01", Status: "downloading", Progress: 0.5},
			{ID: "2", Name: "Show.S01E02", Status: "queued", Progress: 0.0},
			{ID: "3", Name: "Movie", Status: "failed", Progress: 0.0},
		},
	}

	reg := buildRegistry(map[string]plugins.Plugin{"sabnzbd": p}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/downloads")
	if err != nil {
		t.Fatalf("GET /api/v1/downloads: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	var got downloadsResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(got.Items))
	}
	// Should be sorted by status priority: downloading → queued → failed.
	if got.Items[0].Status != "downloading" {
		t.Errorf("items[0].status: got %q, want downloading", got.Items[0].Status)
	}
	if got.Items[1].Status != "queued" {
		t.Errorf("items[1].status: got %q, want queued", got.Items[1].Status)
	}
	if got.Items[2].Status != "failed" {
		t.Errorf("items[2].status: got %q, want failed", got.Items[2].Status)
	}
	if len(got.FailedPlugins) != 0 {
		t.Errorf("expected no failed plugins, got %v", got.FailedPlugins)
	}
}

func TestDownloads_EmptyRegistry(t *testing.T) {
	reg := buildRegistry(nil, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/downloads")
	if err != nil {
		t.Fatalf("GET /api/v1/downloads: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got downloadsResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(got.Items))
	}
	if len(got.FailedPlugins) != 0 {
		t.Errorf("expected 0 failed plugins, got %v", got.FailedPlugins)
	}
}

func TestDownloads_PluginError_GracefulDegradation(t *testing.T) {
	good := &mockDownloadsPlugin{
		manifest: plugins.PluginManifest{
			ID:           "sabnzbd",
			Type:         "sabnzbd",
			DisplayName:  "SABnzbd",
			Capabilities: []plugins.Capability{plugins.CapabilityDownloadsRead},
		},
		items: []plugins.DownloadItem{
			{ID: "1", Name: "Movie.mkv", Status: "downloading", Progress: 0.75},
		},
	}
	// A second "downloads" plugin that errors — simulated via a fake plugin ID.
	bad := &mockDownloadsPlugin{
		manifest: plugins.PluginManifest{
			ID:           "nzbget",
			Type:         "nzbget",
			DisplayName:  "NZBGet",
			Capabilities: []plugins.Capability{plugins.CapabilityDownloadsRead},
		},
		queueErr: errors.New("connection refused"),
	}

	reg := buildRegistry(map[string]plugins.Plugin{
		"sabnzbd": good,
		"nzbget":  bad,
	}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/downloads")
	if err != nil {
		t.Fatalf("GET /api/v1/downloads: %v", err)
	}
	defer resp.Body.Close()

	// Must always return 200.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got downloadsResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// One item from the working plugin.
	if len(got.Items) != 1 {
		t.Errorf("expected 1 item from working plugin, got %d", len(got.Items))
	}
	// Failed plugin name in failed_plugins.
	if len(got.FailedPlugins) != 1 {
		t.Errorf("expected 1 failed plugin, got %v", got.FailedPlugins)
	} else if got.FailedPlugins[0] != "NZBGet" {
		t.Errorf("failed_plugins[0]: got %q, want NZBGet", got.FailedPlugins[0])
	}
}

func TestDownloads_StatusPriorityOrdering(t *testing.T) {
	// All five known statuses in reverse priority order to verify sorting.
	p := &mockDownloadsPlugin{
		manifest: plugins.PluginManifest{
			ID:           "sabnzbd",
			Type:         "sabnzbd",
			DisplayName:  "SABnzbd",
			Capabilities: []plugins.Capability{plugins.CapabilityDownloadsRead},
		},
		items: []plugins.DownloadItem{
			{ID: "5", Name: "E", Status: "failed"},
			{ID: "4", Name: "D", Status: "paused"},
			{ID: "3", Name: "C", Status: "processing"},
			{ID: "2", Name: "B", Status: "queued"},
			{ID: "1", Name: "A", Status: "downloading"},
		},
	}

	reg := buildRegistry(map[string]plugins.Plugin{"sabnzbd": p}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/downloads")
	if err != nil {
		t.Fatalf("GET /api/v1/downloads: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var got downloadsResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	wantOrder := []string{"downloading", "queued", "processing", "paused", "failed"}
	if len(got.Items) != len(wantOrder) {
		t.Fatalf("expected %d items, got %d", len(wantOrder), len(got.Items))
	}
	for i, want := range wantOrder {
		if got.Items[i].Status != want {
			t.Errorf("items[%d].status: got %q, want %q", i, got.Items[i].Status, want)
		}
	}
}
