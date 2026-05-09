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
// mockActivityPlugin — plugins.Plugin + plugins.ActivityReadable
// ---------------------------------------------------------------------------

type mockActivityPlugin struct {
	manifest  plugins.PluginManifest
	healthErr error
	events    []plugins.ActivityEvent
	actErr    error
}

func (m *mockActivityPlugin) Manifest() plugins.PluginManifest { return m.manifest }
func (m *mockActivityPlugin) Health() (plugins.HealthStatus, error) {
	return plugins.HealthStatus{Status: "healthy", Reachable: true}, m.healthErr
}
func (m *mockActivityPlugin) GetActivity(_ *string) ([]plugins.ActivityEvent, error) {
	return m.events, m.actErr
}

// activityResp mirrors the JSON shape of activityResponse.
type activityResp struct {
	Events        []plugins.ActivityEvent `json:"events"`
	FailedPlugins []string                `json:"failed_plugins"`
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestActivity_HappyPath(t *testing.T) {
	p := &mockActivityPlugin{
		manifest: plugins.PluginManifest{
			ID:           "sonarr",
			Type:         "sonarr",
			DisplayName:  "Sonarr",
			Capabilities: []plugins.Capability{plugins.CapabilityActivityRead},
		},
		events: []plugins.ActivityEvent{
			{
				ID:           "sonarr:1",
				SourcePlugin: "sonarr",
				Type:         plugins.ActivityEventDownloaded,
				Message:      "Downloaded S01E01",
				Timestamp:    "2026-05-08T12:00:00Z",
				Visibility:   plugins.ActivityVisibilityAllUsers,
			},
			{
				ID:           "sonarr:2",
				SourcePlugin: "sonarr",
				Type:         plugins.ActivityEventAddedToLibrary,
				Message:      "Added to library",
				Timestamp:    "2026-05-08T11:00:00Z",
				Visibility:   plugins.ActivityVisibilityAllUsers,
			},
		},
	}

	reg := buildRegistry(map[string]plugins.Plugin{"sonarr": p}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/activity")
	if err != nil {
		t.Fatalf("GET /api/v1/activity: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	var got activityResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got.Events))
	}
	// Events should be sorted descending by timestamp.
	if got.Events[0].ID != "sonarr:1" {
		t.Errorf("events[0].ID: got %q, want sonarr:1 (newer event first)", got.Events[0].ID)
	}
	if got.Events[1].ID != "sonarr:2" {
		t.Errorf("events[1].ID: got %q, want sonarr:2", got.Events[1].ID)
	}
	if len(got.FailedPlugins) != 0 {
		t.Errorf("expected no failed plugins, got %v", got.FailedPlugins)
	}
}

func TestActivity_EmptyRegistry(t *testing.T) {
	reg := buildRegistry(nil, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/activity")
	if err != nil {
		t.Fatalf("GET /api/v1/activity: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got activityResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(got.Events))
	}
	if len(got.FailedPlugins) != 0 {
		t.Errorf("expected 0 failed plugins, got %v", got.FailedPlugins)
	}
}

func TestActivity_PluginError_GracefulDegradation(t *testing.T) {
	good := &mockActivityPlugin{
		manifest: plugins.PluginManifest{
			ID:           "radarr",
			Type:         "radarr",
			DisplayName:  "Radarr",
			Capabilities: []plugins.Capability{plugins.CapabilityActivityRead},
		},
		events: []plugins.ActivityEvent{
			{
				ID:           "radarr:1",
				SourcePlugin: "radarr",
				Type:         plugins.ActivityEventDownloaded,
				Message:      "Downloaded movie",
				Timestamp:    "2026-05-08T10:00:00Z",
				Visibility:   plugins.ActivityVisibilityAllUsers,
			},
		},
	}
	bad := &mockActivityPlugin{
		manifest: plugins.PluginManifest{
			ID:           "sonarr",
			Type:         "sonarr",
			DisplayName:  "Sonarr",
			Capabilities: []plugins.Capability{plugins.CapabilityActivityRead},
		},
		actErr: errors.New("connection refused"),
	}

	reg := buildRegistry(map[string]plugins.Plugin{
		"radarr": good,
		"sonarr": bad,
	}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/activity")
	if err != nil {
		t.Fatalf("GET /api/v1/activity: %v", err)
	}
	defer resp.Body.Close()

	// Must always return 200 even when a plugin errors.
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got activityResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// One event from the working plugin.
	if len(got.Events) != 1 {
		t.Errorf("expected 1 event from working plugin, got %d", len(got.Events))
	}
	// Failed plugin name in failed_plugins.
	if len(got.FailedPlugins) != 1 {
		t.Errorf("expected 1 failed plugin, got %v", got.FailedPlugins)
	} else if got.FailedPlugins[0] != "Sonarr" {
		t.Errorf("failed_plugins[0]: got %q, want Sonarr", got.FailedPlugins[0])
	}
}

func TestActivity_SinceParam_Forwarded(t *testing.T) {
	// Verify that the since query param is forwarded to the plugin.
	var capturedSince *string
	p := &captureActivityPlugin{
		manifest: plugins.PluginManifest{
			ID:           "sonarr",
			Type:         "sonarr",
			DisplayName:  "Sonarr",
			Capabilities: []plugins.Capability{plugins.CapabilityActivityRead},
		},
		capturedSince: &capturedSince,
	}

	reg := buildRegistry(map[string]plugins.Plugin{"sonarr": p}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/activity?since=2026-05-01T00:00:00Z")
	if err != nil {
		t.Fatalf("GET /api/v1/activity: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	if capturedSince == nil {
		t.Fatal("since was not forwarded to plugin (got nil)")
	}
	if *capturedSince != "2026-05-01T00:00:00Z" {
		t.Errorf("since: got %q, want 2026-05-01T00:00:00Z", *capturedSince)
	}
}

// captureActivityPlugin records the since param passed to GetActivity.
type captureActivityPlugin struct {
	manifest      plugins.PluginManifest
	capturedSince **string
}

func (c *captureActivityPlugin) Manifest() plugins.PluginManifest { return c.manifest }
func (c *captureActivityPlugin) Health() (plugins.HealthStatus, error) {
	return plugins.HealthStatus{Status: "healthy", Reachable: true}, nil
}
func (c *captureActivityPlugin) GetActivity(since *string) ([]plugins.ActivityEvent, error) {
	*c.capturedSince = since
	return []plugins.ActivityEvent{}, nil
}
