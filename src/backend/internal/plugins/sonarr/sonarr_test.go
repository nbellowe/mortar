package sonarr_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	plugins "github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/sonarr"
)

// newTestPlugin builds a Sonarr plugin wired to the given test server URL.
func newTestPlugin(t *testing.T, serverURL string) plugins.Plugin {
	t.Helper()
	p, err := sonarr.New(config.PluginConfig{
		ID:     "sonarr-test",
		Type:   "sonarr",
		URL:    serverURL,
		APIKey: "test-api-key",
	})
	if err != nil {
		t.Fatalf("sonarr.New: %v", err)
	}
	return p
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

func TestHealth_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"appName":"Sonarr","version":"4.0.0"}`))
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}
	if h.Status != "healthy" {
		t.Errorf("Status = %q, want %q", h.Status, "healthy")
	}
	if !h.Reachable {
		t.Error("Reachable = false, want true")
	}
}

func TestHealth_Down(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}
	if h.Status != "unreachable" {
		t.Errorf("Status = %q, want %q", h.Status, "unreachable")
	}
	if h.Reachable {
		t.Error("Reachable = true, want false")
	}
}

// ---------------------------------------------------------------------------
// GetActivity
// ---------------------------------------------------------------------------

// historyResponse mirrors the Sonarr API shape used by the plugin.
type historyResponse struct {
	Records []historyRecord `json:"records"`
}

type historyRecord struct {
	ID          int            `json:"id"`
	EpisodeID   int            `json:"episodeId"`
	SourceTitle string         `json:"sourceTitle"`
	Date        string         `json:"date"`
	EventType   string         `json:"eventType"`
	Series      seriesSummary  `json:"series"`
	Episode     episodeSummary `json:"episode"`
}

type seriesSummary struct {
	Title  string `json:"title"`
	TvdbID int    `json:"tvdbId"`
}

type episodeSummary struct {
	Title string `json:"title"`
}

func makeHistoryServer(t *testing.T, records []historyRecord) *httptest.Server {
	t.Helper()
	body, err := json.Marshal(historyResponse{Records: records})
	if err != nil {
		t.Fatalf("marshal history: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/history":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		case "/api/v3/system/status":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	return srv
}

func TestGetActivity_HappyPath(t *testing.T) {
	records := []historyRecord{
		{
			ID:          1,
			EpisodeID:   101,
			SourceTitle: "Show.S01E01.720p",
			Date:        "2026-05-07T10:00:00Z",
			EventType:   "grabbed",
			Series:      seriesSummary{Title: "Great Show", TvdbID: 12345},
			Episode:     episodeSummary{Title: "Pilot"},
		},
		{
			ID:          2,
			EpisodeID:   102,
			SourceTitle: "Show.S01E02.1080p",
			Date:        "2026-05-07T11:00:00Z",
			EventType:   "downloadFolderImported",
			Series:      seriesSummary{Title: "Great Show", TvdbID: 12345},
			Episode:     episodeSummary{Title: "Episode 2"},
		},
	}

	srv := makeHistoryServer(t, records)
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	ar, ok := p.(plugins.ActivityReadable)
	if !ok {
		t.Fatal("plugin does not implement ActivityReadable")
	}

	events, err := ar.GetActivity(nil)
	if err != nil {
		t.Fatalf("GetActivity returned error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}

	// --- event 0: grabbed → ActivityEventDownloaded ---
	e0 := events[0]
	if e0.Type != plugins.ActivityEventDownloaded {
		t.Errorf("events[0].Type = %q, want %q", e0.Type, plugins.ActivityEventDownloaded)
	}
	if e0.Timestamp != "2026-05-07T10:00:00Z" {
		t.Errorf("events[0].Timestamp = %q, want %q", e0.Timestamp, "2026-05-07T10:00:00Z")
	}
	if e0.Visibility != plugins.ActivityVisibilityAllUsers {
		t.Errorf("events[0].Visibility = %q, want %q", e0.Visibility, plugins.ActivityVisibilityAllUsers)
	}
	if e0.SourcePlugin != "sonarr-test" {
		t.Errorf("events[0].SourcePlugin = %q, want %q", e0.SourcePlugin, "sonarr-test")
	}
	if e0.Item == nil {
		t.Fatal("events[0].Item is nil")
	}
	if e0.Item.Type != plugins.MediaTypeShow {
		t.Errorf("events[0].Item.Type = %q, want %q", e0.Item.Type, plugins.MediaTypeShow)
	}

	// --- event 1: downloadFolderImported → ActivityEventAddedToLibrary ---
	e1 := events[1]
	if e1.Type != plugins.ActivityEventAddedToLibrary {
		t.Errorf("events[1].Type = %q, want %q", e1.Type, plugins.ActivityEventAddedToLibrary)
	}
	if e1.Timestamp != "2026-05-07T11:00:00Z" {
		t.Errorf("events[1].Timestamp = %q, want %q", e1.Timestamp, "2026-05-07T11:00:00Z")
	}
	if e1.Visibility != plugins.ActivityVisibilityAllUsers {
		t.Errorf("events[1].Visibility = %q, want %q", e1.Visibility, plugins.ActivityVisibilityAllUsers)
	}
	if e1.SourcePlugin != "sonarr-test" {
		t.Errorf("events[1].SourcePlugin = %q, want %q", e1.SourcePlugin, "sonarr-test")
	}
	if e1.Item == nil {
		t.Fatal("events[1].Item is nil")
	}
	if e1.Item.Type != plugins.MediaTypeShow {
		t.Errorf("events[1].Item.Type = %q, want %q", e1.Item.Type, plugins.MediaTypeShow)
	}
}

func TestGetActivity_SinceFilter(t *testing.T) {
	// Three records: before, at, and after the since timestamp.
	since := "2026-05-07T10:30:00Z"
	sinceTime, _ := time.Parse(time.RFC3339, since)
	beforeTime := sinceTime.Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	atTime := sinceTime.UTC().Format(time.RFC3339) // exactly at since — should be excluded (not After)
	afterTime := sinceTime.Add(1 * time.Hour).UTC().Format(time.RFC3339)

	records := []historyRecord{
		{ID: 1, EpisodeID: 101, SourceTitle: "old", Date: beforeTime, EventType: "grabbed",
			Series: seriesSummary{Title: "Show A"}, Episode: episodeSummary{Title: "Ep1"}},
		{ID: 2, EpisodeID: 102, SourceTitle: "at", Date: atTime, EventType: "grabbed",
			Series: seriesSummary{Title: "Show A"}, Episode: episodeSummary{Title: "Ep2"}},
		{ID: 3, EpisodeID: 103, SourceTitle: "new", Date: afterTime, EventType: "grabbed",
			Series: seriesSummary{Title: "Show A"}, Episode: episodeSummary{Title: "Ep3"}},
	}

	srv := makeHistoryServer(t, records)
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	ar := p.(plugins.ActivityReadable)

	events, err := ar.GetActivity(&since)
	if err != nil {
		t.Fatalf("GetActivity returned error: %v", err)
	}
	// Only the record strictly after since should pass the filter.
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1 (only records strictly after since)", len(events))
	}
	if events[0].Timestamp != afterTime {
		t.Errorf("events[0].Timestamp = %q, want %q", events[0].Timestamp, afterTime)
	}
}

func TestGetActivity_EventTypeMapping(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		wantType  plugins.ActivityEventType
		wantSkip  bool // true if the event should be absent from results
	}{
		{
			name:      "downloadImported maps to AddedToLibrary",
			eventType: "downloadImported",
			wantType:  plugins.ActivityEventAddedToLibrary,
		},
		{
			name:      "seriesDelete maps to Deleted",
			eventType: "seriesDelete",
			wantType:  plugins.ActivityEventDeleted,
		},
		{
			name:      "episodeFileDelete maps to Deleted",
			eventType: "episodeFileDelete",
			wantType:  plugins.ActivityEventDeleted,
		},
		{
			name:      "failed maps to Failed",
			eventType: "failed",
			wantType:  plugins.ActivityEventFailed,
		},
		{
			name:      "unknown eventType is skipped",
			eventType: "seriesAdd",
			wantSkip:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			records := []historyRecord{
				{
					ID:          10,
					EpisodeID:   200,
					SourceTitle: "Some.Show.S01E01",
					Date:        "2026-05-07T12:00:00Z",
					EventType:   tc.eventType,
					Series:      seriesSummary{Title: "Some Show", TvdbID: 99999},
					Episode:     episodeSummary{Title: "Pilot"},
				},
			}

			srv := makeHistoryServer(t, records)
			defer srv.Close()

			p := newTestPlugin(t, srv.URL)
			ar := p.(plugins.ActivityReadable)

			events, err := ar.GetActivity(nil)
			if err != nil {
				t.Fatalf("GetActivity returned error: %v", err)
			}

			if tc.wantSkip {
				if len(events) != 0 {
					t.Errorf("expected 0 events for unknown eventType %q, got %d", tc.eventType, len(events))
				}
				return
			}

			if len(events) != 1 {
				t.Fatalf("len(events) = %d, want 1", len(events))
			}
			if events[0].Type != tc.wantType {
				t.Errorf("events[0].Type = %q, want %q", events[0].Type, tc.wantType)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Constructor validation
// ---------------------------------------------------------------------------

func TestNew_MissingAPIKey(t *testing.T) {
	_, err := sonarr.New(config.PluginConfig{
		ID:  "sonarr-test",
		URL: "http://localhost:8989",
	})
	if err == nil {
		t.Error("expected error when api_key is missing, got nil")
	}
}

func TestNew_MissingURL(t *testing.T) {
	_, err := sonarr.New(config.PluginConfig{
		ID:     "sonarr-test",
		APIKey: "some-key",
	})
	if err == nil {
		t.Error("expected error when url is missing, got nil")
	}
}
