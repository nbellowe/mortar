package radarr_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/radarr"
)

// newTestPlugin creates a Plugin pointed at the given test server URL.
func newTestPlugin(t *testing.T, serverURL string) plugins.Plugin {
	t.Helper()
	p, err := radarr.New(config.PluginConfig{
		ID:     "radarr-test",
		Type:   "radarr",
		URL:    serverURL,
		APIKey: "test-key",
	})
	if err != nil {
		t.Fatalf("radarr.New: %v", err)
	}
	return p
}

// -------------------------------------------------------------------------
// Health tests
// -------------------------------------------------------------------------

func TestHealth_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"version": "5.2.0"})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health() returned error: %v", err)
	}
	if !h.Reachable {
		t.Errorf("expected Reachable=true, got false (detail: %v)", h.Detail)
	}
	if h.Status != "healthy" && h.Status != "degraded" {
		t.Errorf("expected healthy or degraded, got %q", h.Status)
	}
	if h.CheckedAt == "" {
		t.Error("expected CheckedAt to be set")
	}
}

func TestHealth_Down(t *testing.T) {
	// Point at a server that immediately closes the connection.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health() returned error: %v", err)
	}
	if h.Reachable {
		t.Error("expected Reachable=false for a 503 response")
	}
	if h.Status != "unreachable" {
		t.Errorf("expected status=unreachable, got %q", h.Status)
	}
	if h.Detail == nil || *h.Detail == "" {
		t.Error("expected Detail to be set for an unreachable health check")
	}
}

// -------------------------------------------------------------------------
// GetActivity tests
// -------------------------------------------------------------------------

// sampleHistory returns a minimal Radarr history payload with one record of
// each interesting event type and rich movie metadata.
func sampleHistoryResponse() map[string]interface{} {
	return map[string]interface{}{
		"records": []map[string]interface{}{
			{
				"id":          101,
				"eventType":   "grabbed",
				"sourceTitle": "Inception.2010.BluRay.mkv",
				"date":        "2026-05-01T10:00:00Z",
				"movieId":     42,
				"movie": map[string]interface{}{
					"title":  "Inception",
					"year":   2010,
					"imdbId": "tt1375666",
					"tmdbId": 27205,
				},
			},
			{
				"id":          102,
				"eventType":   "downloadFolderImported",
				"sourceTitle": "The.Matrix.1999.mkv",
				"date":        "2026-05-02T12:00:00Z",
				"movieId":     43,
				"movie": map[string]interface{}{
					"title":  "The Matrix",
					"year":   1999,
					"imdbId": "tt0133093",
					"tmdbId": 603,
				},
			},
			{
				"id":          103,
				"eventType":   "movieFileDelete",
				"sourceTitle": "OldFile.mkv",
				"date":        "2026-05-03T08:00:00Z",
				"movieId":     44,
				"movie": map[string]interface{}{
					"title":  "Old Movie",
					"year":   1990,
					"imdbId": "",
					"tmdbId": 0,
				},
			},
			{
				"id":          104,
				"eventType":   "failed",
				"sourceTitle": "BadDownload.mkv",
				"date":        "2026-05-04T06:00:00Z",
				"movieId":     45,
				"movie": map[string]interface{}{
					"title":  "Failed Film",
					"year":   2000,
					"imdbId": "",
					"tmdbId": 0,
				},
			},
			{
				"id":          105,
				"eventType":   "downloadImported",
				"sourceTitle": "Dune.2021.mkv",
				"date":        "2026-05-05T14:00:00Z",
				"movieId":     46,
				"movie": map[string]interface{}{
					"title":  "Dune",
					"year":   2021,
					"imdbId": "tt1160419",
					"tmdbId": 438631,
				},
			},
			{
				"id":          106,
				"eventType":   "movieDelete",
				"sourceTitle": "DeletedMovie.mkv",
				"date":        "2026-05-06T09:00:00Z",
				"movieId":     47,
				"movie": map[string]interface{}{
					"title":  "Deleted Movie",
					"year":   2005,
					"imdbId": "",
					"tmdbId": 0,
				},
			},
		},
	}
}

func historyServer(t *testing.T, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/history" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func TestGetActivity_FullMapping(t *testing.T) {
	srv := historyServer(t, sampleHistoryResponse())
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	ar, ok := p.(plugins.ActivityReadable)
	if !ok {
		t.Fatal("plugin does not implement ActivityReadable")
	}

	events, err := ar.GetActivity(nil)
	if err != nil {
		t.Fatalf("GetActivity: %v", err)
	}
	if len(events) != 6 {
		t.Fatalf("expected 6 events, got %d", len(events))
	}

	// Check event type mapping.
	wantTypes := []plugins.ActivityEventType{
		plugins.ActivityEventDownloaded,     // grabbed
		plugins.ActivityEventAddedToLibrary, // downloadFolderImported
		plugins.ActivityEventDeleted,        // movieFileDelete
		plugins.ActivityEventFailed,         // failed
		plugins.ActivityEventAddedToLibrary, // downloadImported
		plugins.ActivityEventDeleted,        // movieDelete
	}
	for i, ev := range events {
		if ev.Type != wantTypes[i] {
			t.Errorf("event[%d]: expected type %q, got %q", i, wantTypes[i], ev.Type)
		}
	}

	// Check first event (Inception / grabbed) in detail.
	ev := events[0]
	if ev.ID != "radarr-test:101" {
		t.Errorf("expected ID=radarr-test:101, got %q", ev.ID)
	}
	if ev.SourcePlugin != "radarr-test" {
		t.Errorf("expected SourcePlugin=radarr-test, got %q", ev.SourcePlugin)
	}
	if ev.Message != "Inception.2010.BluRay.mkv" {
		t.Errorf("expected Message=Inception.2010.BluRay.mkv, got %q", ev.Message)
	}
	if ev.Timestamp != "2026-05-01T10:00:00Z" {
		t.Errorf("expected Timestamp=2026-05-01T10:00:00Z, got %q", ev.Timestamp)
	}
	if ev.Visibility != plugins.ActivityVisibilityAllUsers {
		t.Errorf("expected Visibility=all_users, got %q", ev.Visibility)
	}

	// Check MediaItem fields.
	item := ev.Item
	if item == nil {
		t.Fatal("expected Item to be non-nil")
	}
	if item.ID != "radarr-test:42" {
		t.Errorf("expected Item.ID=radarr-test:42, got %q", item.ID)
	}
	if item.ExternalID != "42" {
		t.Errorf("expected Item.ExternalID=42, got %q", item.ExternalID)
	}
	if item.Type != plugins.MediaTypeMovie {
		t.Errorf("expected Item.Type=movie, got %q", item.Type)
	}
	if item.Title != "Inception" {
		t.Errorf("expected Item.Title=Inception, got %q", item.Title)
	}
	if item.Year == nil || *item.Year != 2010 {
		t.Errorf("expected Item.Year=2010, got %v", item.Year)
	}
	if item.ImdbID == nil || *item.ImdbID != "tt1375666" {
		t.Errorf("expected Item.ImdbID=tt1375666, got %v", item.ImdbID)
	}
	if item.TmdbID == nil || *item.TmdbID != "27205" {
		t.Errorf("expected Item.TmdbID=27205, got %v", item.TmdbID)
	}

	// Third event has no imdbId and tmdbId=0 — verify both are nil.
	ev3 := events[2]
	if ev3.Item == nil {
		t.Fatal("expected event[2].Item to be non-nil")
	}
	if ev3.Item.ImdbID != nil {
		t.Errorf("expected ImdbID=nil for empty imdbId, got %v", ev3.Item.ImdbID)
	}
	if ev3.Item.TmdbID != nil {
		t.Errorf("expected TmdbID=nil for tmdbId=0, got %v", ev3.Item.TmdbID)
	}

	// Fifth event (downloadImported / Dune) — verify AddedToLibrary.
	ev5 := events[4]
	if ev5.ID != "radarr-test:105" {
		t.Errorf("expected event[4].ID=radarr-test:105, got %q", ev5.ID)
	}
	if ev5.Item == nil {
		t.Fatal("expected event[4].Item to be non-nil")
	}
	if ev5.Item.TmdbID == nil || *ev5.Item.TmdbID != "438631" {
		t.Errorf("expected event[4].Item.TmdbID=438631, got %v", ev5.Item.TmdbID)
	}

	// Sixth event (movieDelete) — verify Deleted type and nil TmdbID.
	ev6 := events[5]
	if ev6.ID != "radarr-test:106" {
		t.Errorf("expected event[5].ID=radarr-test:106, got %q", ev6.ID)
	}
	if ev6.Item == nil {
		t.Fatal("expected event[5].Item to be non-nil")
	}
	if ev6.Item.TmdbID != nil {
		t.Errorf("expected event[5].Item.TmdbID=nil for tmdbId=0, got %v", ev6.Item.TmdbID)
	}
}

func TestGetActivity_SinceFilter(t *testing.T) {
	srv := historyServer(t, sampleHistoryResponse())
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	ar := p.(plugins.ActivityReadable)

	// since = 2026-05-02T00:00:00Z; should exclude the first record (2026-05-01).
	since := "2026-05-02T00:00:00Z"
	events, err := ar.GetActivity(&since)
	if err != nil {
		t.Fatalf("GetActivity with since: %v", err)
	}
	if len(events) != 5 {
		t.Fatalf("expected 5 events after since filter, got %d", len(events))
	}
	// Verify the filtered-out record is gone.
	for _, ev := range events {
		if ev.ID == "radarr-test:101" {
			t.Error("event radarr-test:101 should have been filtered out by since=2026-05-02T00:00:00Z")
		}
	}
}

func TestGetActivity_SinceFilterExcludesAll(t *testing.T) {
	srv := historyServer(t, sampleHistoryResponse())
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	ar := p.(plugins.ActivityReadable)

	// since is after all records — expect empty result.
	since := "2026-06-01T00:00:00Z"
	events, err := ar.GetActivity(&since)
	if err != nil {
		t.Fatalf("GetActivity with future since: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

// -------------------------------------------------------------------------
// Manifest test
// -------------------------------------------------------------------------

func TestManifest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	m := p.Manifest()

	if m.ID != "radarr-test" {
		t.Errorf("expected ID=radarr-test, got %q", m.ID)
	}
	if m.Type != "radarr" {
		t.Errorf("expected Type=radarr, got %q", m.Type)
	}
	if len(m.Capabilities) != 1 || m.Capabilities[0] != plugins.CapabilityActivityRead {
		t.Errorf("expected [activity.read], got %v", m.Capabilities)
	}
}

// -------------------------------------------------------------------------
// Constructor validation tests
// -------------------------------------------------------------------------

func TestNew_MissingURL(t *testing.T) {
	_, err := radarr.New(config.PluginConfig{ID: "x", Type: "radarr", APIKey: "k"})
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestNew_MissingAPIKey(t *testing.T) {
	_, err := radarr.New(config.PluginConfig{ID: "x", Type: "radarr", URL: "http://radarr"})
	if err == nil {
		t.Error("expected error for missing API key")
	}
}
