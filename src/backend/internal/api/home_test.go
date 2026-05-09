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
// mockLibraryPlugin — plugins.Plugin + plugins.LibraryBrowser
// ---------------------------------------------------------------------------

type mockLibraryPlugin struct {
	manifest    plugins.PluginManifest
	healthResp  plugins.HealthStatus
	healthErr   error
	browseItems []plugins.MediaItem
	browseErr   error
}

func (m *mockLibraryPlugin) Manifest() plugins.PluginManifest { return m.manifest }
func (m *mockLibraryPlugin) Health() (plugins.HealthStatus, error) {
	return m.healthResp, m.healthErr
}
func (m *mockLibraryPlugin) Browse(_ plugins.BrowseOptions) (plugins.PagedResult[plugins.MediaItem], error) {
	if m.browseErr != nil {
		return plugins.PagedResult[plugins.MediaItem]{}, m.browseErr
	}
	return plugins.PagedResult[plugins.MediaItem]{
		Items:    m.browseItems,
		Total:    len(m.browseItems),
		Page:     1,
		PageSize: 10,
	}, nil
}
func (m *mockLibraryPlugin) GetItem(_ string) (*plugins.MediaItem, error) {
	return nil, nil
}
func (m *mockLibraryPlugin) GetPlayURL(_ plugins.MediaItem, _ plugins.MortarUser) (string, error) {
	return "", nil
}

// homeResp mirrors the JSON shape of homeResponse.
type homeResp struct {
	RecentlyAdded []plugins.MediaItem `json:"recently_added"`
	HealthSummary struct {
		AnyUnreachable   bool `json:"any_unreachable"`
		Total            int  `json:"total"`
		UnreachableCount int  `json:"unreachable_count"`
	} `json:"health_summary"`
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestHome_HappyPath(t *testing.T) {
	year := 2026
	lib := &mockLibraryPlugin{
		manifest: plugins.PluginManifest{
			ID:           "jellyfin",
			Type:         "jellyfin",
			DisplayName:  "Jellyfin",
			Capabilities: []plugins.Capability{plugins.CapabilityLibraryBrowse},
		},
		healthResp: plugins.HealthStatus{Status: "healthy", Reachable: true, CheckedAt: "2026-05-08T10:00:00Z"},
		browseItems: []plugins.MediaItem{
			{ID: "jellyfin:1", Title: "Dune", Type: plugins.MediaTypeMovie, Year: &year},
			{ID: "jellyfin:2", Title: "Severance", Type: plugins.MediaTypeShow},
		},
	}

	reg := buildRegistry(map[string]plugins.Plugin{"jellyfin": lib}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/home")
	if err != nil {
		t.Fatalf("GET /api/v1/home: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	var got homeResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.RecentlyAdded) != 2 {
		t.Fatalf("recently_added: expected 2 items, got %d", len(got.RecentlyAdded))
	}
	if got.RecentlyAdded[0].Title != "Dune" {
		t.Errorf("recently_added[0].title: got %q, want Dune", got.RecentlyAdded[0].Title)
	}
	if got.HealthSummary.Total != 1 {
		t.Errorf("health_summary.total: got %d, want 1", got.HealthSummary.Total)
	}
	if got.HealthSummary.UnreachableCount != 0 {
		t.Errorf("health_summary.unreachable_count: got %d, want 0", got.HealthSummary.UnreachableCount)
	}
	if got.HealthSummary.AnyUnreachable {
		t.Error("health_summary.any_unreachable: expected false")
	}
}

func TestHome_EmptyRegistry(t *testing.T) {
	reg := buildRegistry(nil, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/home")
	if err != nil {
		t.Fatalf("GET /api/v1/home: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got homeResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.RecentlyAdded == nil || len(got.RecentlyAdded) != 0 {
		t.Errorf("recently_added: expected empty slice, got %v", got.RecentlyAdded)
	}
	if got.HealthSummary.Total != 0 {
		t.Errorf("health_summary.total: got %d, want 0", got.HealthSummary.Total)
	}
	if got.HealthSummary.UnreachableCount != 0 {
		t.Errorf("health_summary.unreachable_count: got %d, want 0", got.HealthSummary.UnreachableCount)
	}
}

func TestHome_NoLibraryBrowser_RecentlyAddedEmpty(t *testing.T) {
	// Plugin with no library.browse capability.
	p := &mockPlugin{
		manifest: plugins.PluginManifest{
			ID:          "sabnzbd",
			Type:        "sabnzbd",
			DisplayName: "SABnzbd",
		},
		healthResp: plugins.HealthStatus{Status: "healthy", Reachable: true, CheckedAt: "2026-05-08T10:00:00Z"},
	}

	reg := buildRegistry(map[string]plugins.Plugin{"sabnzbd": p}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/home")
	if err != nil {
		t.Fatalf("GET /api/v1/home: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got homeResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.RecentlyAdded) != 0 {
		t.Errorf("recently_added: expected empty slice, got %d items", len(got.RecentlyAdded))
	}
	// SABnzbd is healthy.
	if got.HealthSummary.Total != 1 {
		t.Errorf("health_summary.total: got %d, want 1", got.HealthSummary.Total)
	}
	if got.HealthSummary.AnyUnreachable {
		t.Error("health_summary.any_unreachable: expected false")
	}
}

func TestHome_UnreachablePlugin_ReflectedInHealthSummary(t *testing.T) {
	healthy := &mockPlugin{
		manifest: plugins.PluginManifest{
			ID:          "sonarr",
			Type:        "sonarr",
			DisplayName: "Sonarr",
		},
		healthResp: plugins.HealthStatus{Status: "healthy", Reachable: true, CheckedAt: "2026-05-08T10:00:00Z"},
	}
	unreachable := &mockPlugin{
		manifest: plugins.PluginManifest{
			ID:          "radarr",
			Type:        "radarr",
			DisplayName: "Radarr",
		},
		healthErr: nil,
		healthResp: plugins.HealthStatus{
			Status:    "unreachable",
			Reachable: false,
			CheckedAt: "2026-05-08T10:00:00Z",
		},
	}

	reg := buildRegistry(map[string]plugins.Plugin{
		"sonarr": healthy,
		"radarr": unreachable,
	}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/home")
	if err != nil {
		t.Fatalf("GET /api/v1/home: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got homeResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.HealthSummary.Total != 2 {
		t.Errorf("health_summary.total: got %d, want 2", got.HealthSummary.Total)
	}
	if got.HealthSummary.UnreachableCount != 1 {
		t.Errorf("health_summary.unreachable_count: got %d, want 1", got.HealthSummary.UnreachableCount)
	}
	if !got.HealthSummary.AnyUnreachable {
		t.Error("health_summary.any_unreachable: expected true")
	}
}

func TestHome_BrowseError_RecentlyAddedEmpty(t *testing.T) {
	// Library plugin that errors on Browse — recently_added should be empty.
	lib := &mockLibraryPlugin{
		manifest: plugins.PluginManifest{
			ID:           "jellyfin",
			Type:         "jellyfin",
			DisplayName:  "Jellyfin",
			Capabilities: []plugins.Capability{plugins.CapabilityLibraryBrowse},
		},
		healthResp: plugins.HealthStatus{Status: "healthy", Reachable: true, CheckedAt: "2026-05-08T10:00:00Z"},
		browseErr:  errBrowse,
	}

	reg := buildRegistry(map[string]plugins.Plugin{"jellyfin": lib}, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/home")
	if err != nil {
		t.Fatalf("GET /api/v1/home: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got homeResp
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.RecentlyAdded) != 0 {
		t.Errorf("recently_added: expected empty after browse error, got %d items", len(got.RecentlyAdded))
	}
}

// errBrowse is a sentinel error used in home tests.
var errBrowse = errors.New("browse failed")
