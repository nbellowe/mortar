package jellyfin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/jellyfin"
)

// newTestPlugin creates a Jellyfin plugin pointed at the given test server URL.
func newTestPlugin(t *testing.T, serverURL string) plugins.Plugin {
	t.Helper()
	cfg := config.PluginConfig{
		ID:       "jellyfin-test",
		Type:     "jellyfin",
		URL:      serverURL,
		APIKey:   "test-api-key",
		Username: "user-uuid-123",
	}
	p, err := jellyfin.New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return p
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

func TestHealth_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if h.Status != "healthy" {
		t.Errorf("Status = %q; want %q", h.Status, "healthy")
	}
	if !h.Reachable {
		t.Error("Reachable = false; want true")
	}
}

func TestHealth_Unhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
	if h.Status != "unreachable" {
		t.Errorf("Status = %q; want %q", h.Status, "unreachable")
	}
	if h.Reachable {
		t.Error("Reachable = true; want false")
	}
	if h.Detail == nil || !strings.Contains(*h.Detail, "500") {
		t.Errorf("Detail = %v; want message containing 500", h.Detail)
	}
}

// ---------------------------------------------------------------------------
// Browse
// ---------------------------------------------------------------------------

func TestBrowse_Happy(t *testing.T) {
	year := 2020
	resp := map[string]interface{}{
		"Items": []map[string]interface{}{
			{
				"Id":             "abc123",
				"Type":           "Movie",
				"Name":           "Test Movie",
				"ProductionYear": year,
				"ImageTags":      map[string]string{"Primary": "tag"},
				"ProviderIds":    map[string]string{"Imdb": "tt9999999"},
			},
		},
		"TotalRecordCount": 1,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	browser, ok := p.(plugins.LibraryBrowser)
	if !ok {
		t.Fatal("plugin does not implement LibraryBrowser")
	}

	result, err := browser.Browse(plugins.BrowseOptions{})
	if err != nil {
		t.Fatalf("Browse() error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d; want 1", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(Items) = %d; want 1", len(result.Items))
	}

	item := result.Items[0]
	if item.ID != "jellyfin-test:abc123" {
		t.Errorf("ID = %q; want %q", item.ID, "jellyfin-test:abc123")
	}
	if item.Type != plugins.MediaTypeMovie {
		t.Errorf("Type = %q; want %q", item.Type, plugins.MediaTypeMovie)
	}
	if item.Title != "Test Movie" {
		t.Errorf("Title = %q; want %q", item.Title, "Test Movie")
	}
	if item.Year == nil || *item.Year != year {
		t.Errorf("Year = %v; want %d", item.Year, year)
	}
	if item.PosterURL == nil || !strings.Contains(*item.PosterURL, "abc123") {
		t.Errorf("PosterURL = %v; want URL containing item ID", item.PosterURL)
	}
	if item.ImdbID == nil || *item.ImdbID != "tt9999999" {
		t.Errorf("ImdbID = %v; want tt9999999", item.ImdbID)
	}
}

// ---------------------------------------------------------------------------
// FindMatch
// ---------------------------------------------------------------------------

func TestFindMatch_ByIMDBId(t *testing.T) {
	resp := map[string]interface{}{
		"Items": []map[string]interface{}{
			{
				"Id":          "match456",
				"Type":        "Movie",
				"Name":        "Found Movie",
				"ProviderIds": map[string]string{"Imdb": "tt1111111"},
			},
		},
		"TotalRecordCount": 1,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	exists, ok := p.(plugins.LibraryExists)
	if !ok {
		t.Fatal("plugin does not implement LibraryExists")
	}

	imdbID := "tt1111111"
	item := plugins.MediaItem{
		ID:     "other:some-id",
		Title:  "Found Movie",
		Type:   plugins.MediaTypeMovie,
		ImdbID: &imdbID,
	}

	match, err := exists.FindMatch(item)
	if err != nil {
		t.Fatalf("FindMatch() error: %v", err)
	}
	if match == nil {
		t.Fatal("FindMatch() returned nil; want a match")
	}
	if match.MatchedBy != "imdb_id" {
		t.Errorf("MatchedBy = %q; want %q", match.MatchedBy, "imdb_id")
	}
	if match.Item.ID != "jellyfin-test:match456" {
		t.Errorf("Item.ID = %q; want %q", match.Item.ID, "jellyfin-test:match456")
	}
}

func TestFindMatch_NoMatch(t *testing.T) {
	emptyResp := map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(emptyResp)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	exists, ok := p.(plugins.LibraryExists)
	if !ok {
		t.Fatal("plugin does not implement LibraryExists")
	}

	item := plugins.MediaItem{
		ID:    "other:some-id",
		Title: "Nonexistent Movie",
		Type:  plugins.MediaTypeMovie,
	}

	match, err := exists.FindMatch(item)
	if err != nil {
		t.Fatalf("FindMatch() error: %v", err)
	}
	if match != nil {
		t.Errorf("FindMatch() returned %+v; want nil", match)
	}
}

// ---------------------------------------------------------------------------
// GetContinueWatching
// ---------------------------------------------------------------------------

func TestGetContinueWatching_WithAccountLink(t *testing.T) {
	positionTicks := int64(3_000_000_000) // 300 seconds
	runtimeTicks := int64(6_000_000_000)  // 600 seconds
	progress := 50.0
	lastPlayed := "2026-05-07T12:00:00Z"

	resp := map[string]interface{}{
		"Items": []map[string]interface{}{
			{
				"Id":           "resume789",
				"Type":         "Movie",
				"Name":         "Half Watched",
				"RunTimeTicks": runtimeTicks,
				"UserData": map[string]interface{}{
					"PlayedPercentage":      progress,
					"PlaybackPositionTicks": positionTicks,
					"LastPlayedDate":        lastPlayed,
				},
			},
		},
		"TotalRecordCount": 1,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	resumable, ok := p.(plugins.LibraryResumeReadable)
	if !ok {
		t.Fatal("plugin does not implement LibraryResumeReadable")
	}

	user := plugins.MortarUser{
		ID:       "mortar-user-1",
		Username: "alice",
		ExternalAccounts: []plugins.ExternalAccountLink{
			{PluginID: "jellyfin-test", ExternalUserID: "jellyfin-user-uuid"},
		},
	}

	items, err := resumable.GetContinueWatching(user, plugins.ContinueWatchingOptions{})
	if err != nil {
		t.Fatalf("GetContinueWatching() error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d; want 1", len(items))
	}

	cwi := items[0]
	if cwi.Item.ID != "jellyfin-test:resume789" {
		t.Errorf("Item.ID = %q; want %q", cwi.Item.ID, "jellyfin-test:resume789")
	}
	if cwi.Item.Title != "Half Watched" {
		t.Errorf("Title = %q; want %q", cwi.Item.Title, "Half Watched")
	}
	// Progress is stored as 0.0–1.0; 50% → 0.5
	if cwi.Progress != 0.5 {
		t.Errorf("Progress = %f; want 0.5", cwi.Progress)
	}
	if cwi.PositionSeconds == nil || *cwi.PositionSeconds != 300 {
		t.Errorf("PositionSeconds = %v; want 300", cwi.PositionSeconds)
	}
	if cwi.DurationSeconds == nil || *cwi.DurationSeconds != 600 {
		t.Errorf("DurationSeconds = %v; want 600", cwi.DurationSeconds)
	}
	if cwi.LastWatchedAt != lastPlayed {
		t.Errorf("LastWatchedAt = %q; want %q", cwi.LastWatchedAt, lastPlayed)
	}
}

func TestGetContinueWatching_NoAccountLink(t *testing.T) {
	// Server should never be called; if it is, fail the test.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("HTTP request made despite no account link")
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	resumable, ok := p.(plugins.LibraryResumeReadable)
	if !ok {
		t.Fatal("plugin does not implement LibraryResumeReadable")
	}

	user := plugins.MortarUser{
		ID:               "mortar-user-2",
		Username:         "bob",
		ExternalAccounts: nil, // no link
	}

	items, err := resumable.GetContinueWatching(user, plugins.ContinueWatchingOptions{})
	if err != nil {
		t.Fatalf("GetContinueWatching() error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("len(items) = %d; want 0", len(items))
	}
}

// ---------------------------------------------------------------------------
// GetPlayURL
// ---------------------------------------------------------------------------

func TestGetPlayURL(t *testing.T) {
	const serverIDVal = "server-id-abc"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/System/Info/Public" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"Id": serverIDVal})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	browser, ok := p.(plugins.LibraryBrowser)
	if !ok {
		t.Fatal("plugin does not implement LibraryBrowser")
	}

	user := plugins.MortarUser{
		ID:       "mortar-user-3",
		Username: "charlie",
		ExternalAccounts: []plugins.ExternalAccountLink{
			{PluginID: "jellyfin-test", ExternalUserID: "jellyfin-user-uuid"},
		},
	}
	item := plugins.MediaItem{
		ID:         "jellyfin-test:item-xyz",
		ExternalID: "item-xyz",
		PluginID:   "jellyfin-test",
		Title:      "Some Movie",
		Type:       plugins.MediaTypeMovie,
	}

	playURL, err := browser.GetPlayURL(item, user)
	if err != nil {
		t.Fatalf("GetPlayURL() error: %v", err)
	}
	if !strings.Contains(playURL, srv.URL) {
		t.Errorf("GetPlayURL() = %q; want URL containing server base", playURL)
	}
	if !strings.Contains(playURL, "item-xyz") {
		t.Errorf("GetPlayURL() = %q; want URL containing item ID", playURL)
	}
	if !strings.Contains(playURL, serverIDVal) {
		t.Errorf("GetPlayURL() = %q; want URL containing server ID", playURL)
	}
	if !strings.Contains(playURL, "/web/index.html") {
		t.Errorf("GetPlayURL() = %q; want deep-link URL format", playURL)
	}
}

func TestGetPlayURL_NoAccountLink(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No requests should be made for the no-account-link case.
		t.Error("unexpected HTTP request when no account link present")
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	browser, ok := p.(plugins.LibraryBrowser)
	if !ok {
		t.Fatal("plugin does not implement LibraryBrowser")
	}

	user := plugins.MortarUser{
		ID:               "mortar-user-4",
		Username:         "dave",
		ExternalAccounts: nil, // no account link
	}
	item := plugins.MediaItem{
		ID:         "jellyfin-test:item-abc",
		ExternalID: "item-abc",
		PluginID:   "jellyfin-test",
		Title:      "Some Movie",
		Type:       plugins.MediaTypeMovie,
	}

	_, err := browser.GetPlayURL(item, user)
	if err == nil {
		t.Error("GetPlayURL() returned nil error; want error for missing account link")
	}
}
