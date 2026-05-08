package jellyseerr_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/jellyseerr"
)

// newTestPlugin creates a Plugin pointed at the given test server URL.
func newTestPlugin(t *testing.T, serverURL string) plugins.Plugin {
	t.Helper()
	p, err := jellyseerr.New(config.PluginConfig{
		ID:     "jellyseerr",
		Type:   "jellyseerr",
		URL:    serverURL,
		APIKey: "test-api-key",
	})
	if err != nil {
		t.Fatalf("jellyseerr.New: %v", err)
	}
	return p
}

// ---------------------------------------------------------------------------
// Health tests
// ---------------------------------------------------------------------------

func TestHealth_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/status" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Api-Key") != "test-api-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version":"1.0.0"}`))
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health() returned unexpected error: %v", err)
	}
	if !h.Reachable {
		t.Errorf("expected Reachable=true, got false; detail: %v", h.Detail)
	}
	if h.Status != "healthy" && h.Status != "degraded" {
		t.Errorf("expected status healthy or degraded, got %q", h.Status)
	}
}

func TestHealth_Unhealthy(t *testing.T) {
	// Point at a server that immediately closes connections.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health() returned unexpected error: %v", err)
	}
	if h.Reachable {
		t.Error("expected Reachable=false for 500 response")
	}
	if h.Status != "unreachable" {
		t.Errorf("expected status unreachable, got %q", h.Status)
	}
}

func TestHealth_Unreachable(t *testing.T) {
	// Use a non-listening address.
	p := newTestPlugin(t, "http://127.0.0.1:19999")
	h, err := p.Health()
	if err != nil {
		t.Fatalf("Health() returned unexpected error: %v", err)
	}
	if h.Reachable {
		t.Error("expected Reachable=false for unreachable server")
	}
	if h.Status != "unreachable" {
		t.Errorf("expected status unreachable, got %q", h.Status)
	}
	if h.Detail == nil {
		t.Error("expected Detail to be set for unreachable server")
	}
}

// ---------------------------------------------------------------------------
// Search tests
// ---------------------------------------------------------------------------

func TestSearch_ReturnsMappedMediaItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"id":          123,
					"tmdbId":      123,
					"mediaType":   "movie",
					"title":       "Inception",
					"releaseDate": "2010-07-16",
					"overview":    "A thief who steals corporate secrets.",
					"posterPath":  "/poster.jpg",
					"imdbId":      "tt1375666",
				},
				{
					"id":           456,
					"tmdbId":       456,
					"mediaType":    "tv",
					"name":         "Breaking Bad",
					"firstAirDate": "2008-01-20",
					"overview":     "A chemistry teacher turns to crime.",
					"posterPath":   "/bb.jpg",
				},
			},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, ok := p.(plugins.Requester)
	if !ok {
		t.Fatal("plugin does not implement plugins.Requester")
	}

	items, err := req.Search("inception")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Movie assertions.
	movie := items[0]
	if movie.Title != "Inception" {
		t.Errorf("movie title: got %q, want %q", movie.Title, "Inception")
	}
	if movie.Type != plugins.MediaTypeMovie {
		t.Errorf("movie type: got %q, want %q", movie.Type, plugins.MediaTypeMovie)
	}
	if movie.Year == nil || *movie.Year != 2010 {
		t.Errorf("movie year: got %v, want 2010", movie.Year)
	}
	if movie.PosterURL == nil || *movie.PosterURL != "https://image.tmdb.org/t/p/w500/poster.jpg" {
		t.Errorf("movie poster_url: got %v", movie.PosterURL)
	}
	if movie.ImdbID == nil || *movie.ImdbID != "tt1375666" {
		t.Errorf("movie imdb_id: got %v, want tt1375666", movie.ImdbID)
	}
	if movie.ID != "jellyseerr:123" {
		t.Errorf("movie id: got %q, want %q", movie.ID, "jellyseerr:123")
	}
	if movie.TmdbID == nil || *movie.TmdbID != "123" {
		t.Errorf("movie tmdb_id: got %v, want 123", movie.TmdbID)
	}

	// TV show assertions.
	show := items[1]
	if show.Title != "Breaking Bad" {
		t.Errorf("show title: got %q, want %q", show.Title, "Breaking Bad")
	}
	if show.Type != plugins.MediaTypeShow {
		t.Errorf("show type: got %q, want %q", show.Type, plugins.MediaTypeShow)
	}
	if show.Year == nil || *show.Year != 2008 {
		t.Errorf("show year: got %v, want 2008", show.Year)
	}
}

// ---------------------------------------------------------------------------
// ListRequests tests
// ---------------------------------------------------------------------------

func makeTestRequest(id, status int, requesterID int, mediaType, title string) map[string]interface{} {
	return map[string]interface{}{
		"id":        id,
		"status":    status,
		"createdAt": "2024-01-01T00:00:00Z",
		"updatedAt": "2024-01-02T00:00:00Z",
		"media": map[string]interface{}{
			"id":        id * 100,
			"mediaType": mediaType,
			"tmdbId":    id * 100,
			"title":     title,
		},
		"requestedBy": map[string]interface{}{
			"id": requesterID,
		},
	}
}

func TestListRequests_NoFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				makeTestRequest(1, 1, 10, "movie", "Inception"),
				makeTestRequest(2, 2, 11, "tv", "Breaking Bad"),
			},
			"pageInfo": map[string]interface{}{"pages": 1, "pageSize": 20, "results": 2, "skip": 0},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	reqs, err := req.ListRequests(plugins.ListRequestsOptions{})
	if err != nil {
		t.Fatalf("ListRequests() error: %v", err)
	}
	if len(reqs) != 2 {
		t.Errorf("expected 2 requests, got %d", len(reqs))
	}
}

func TestListRequests_FilterByStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				makeTestRequest(1, 1, 10, "movie", "Inception"),       // pending
				makeTestRequest(2, 2, 11, "tv", "Breaking Bad"),       // approved
				makeTestRequest(3, 3, 12, "movie", "The Dark Knight"), // declined
			},
			"pageInfo": map[string]interface{}{"pages": 1, "pageSize": 20, "results": 3, "skip": 0},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	status := plugins.RequestStatusApproved
	reqs, err := req.ListRequests(plugins.ListRequestsOptions{Status: &status})
	if err != nil {
		t.Fatalf("ListRequests() error: %v", err)
	}
	if len(reqs) != 1 {
		t.Errorf("expected 1 approved request, got %d", len(reqs))
	}
	if reqs[0].Status != plugins.RequestStatusApproved {
		t.Errorf("expected status approved, got %q", reqs[0].Status)
	}
}

func TestListRequests_FilterByRequesterID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				makeTestRequest(1, 1, 10, "movie", "Inception"),
				makeTestRequest(2, 1, 11, "tv", "Breaking Bad"),
			},
			"pageInfo": map[string]interface{}{"pages": 1, "pageSize": 20, "results": 2, "skip": 0},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	requesterID := "10"
	reqs, err := req.ListRequests(plugins.ListRequestsOptions{RequesterID: &requesterID})
	if err != nil {
		t.Fatalf("ListRequests() error: %v", err)
	}
	if len(reqs) != 1 {
		t.Errorf("expected 1 request for requester 10, got %d", len(reqs))
	}
	if reqs[0].RequesterID != "10" {
		t.Errorf("expected requester_id 10, got %q", reqs[0].RequesterID)
	}
}

// ---------------------------------------------------------------------------
// SubmitRequest tests
// ---------------------------------------------------------------------------

func TestSubmitRequest_Movie(t *testing.T) {
	var capturedBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "expected POST", http.StatusMethodNotAllowed)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			http.Error(w, "bad request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        99,
			"status":    1,
			"createdAt": "2024-01-01T00:00:00Z",
			"updatedAt": "2024-01-01T00:00:00Z",
			"media": map[string]interface{}{
				"id":        550,
				"mediaType": "movie",
				"tmdbId":    550,
				"title":     "Fight Club",
			},
			"requestedBy": map[string]interface{}{"id": 1},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	tmdbID := "550"
	item := plugins.MediaItem{
		ID:         "jellyseerr:550",
		ExternalID: "550",
		PluginID:   "jellyseerr",
		Type:       plugins.MediaTypeMovie,
		Title:      "Fight Club",
		TmdbID:     &tmdbID,
	}
	user := plugins.MortarUser{ID: "mortar-user-1", Username: "alice", Role: "user"}

	result, err := req.SubmitRequest(item, user)
	if err != nil {
		t.Fatalf("SubmitRequest() error: %v", err)
	}

	if result.ID != "jellyseerr:99" {
		t.Errorf("request id: got %q, want %q", result.ID, "jellyseerr:99")
	}
	if result.Status != plugins.RequestStatusPending {
		t.Errorf("request status: got %q, want pending", result.Status)
	}

	// Verify the request body sent to Jellyseerr.
	if capturedBody["mediaType"] != "movie" {
		t.Errorf("request body mediaType: got %v, want movie", capturedBody["mediaType"])
	}
	if capturedBody["mediaId"] != float64(550) {
		t.Errorf("request body mediaId: got %v, want 550", capturedBody["mediaId"])
	}
}

func TestSubmitRequest_Show(t *testing.T) {
	var capturedBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        100,
			"status":    1,
			"createdAt": "2024-01-01T00:00:00Z",
			"updatedAt": "2024-01-01T00:00:00Z",
			"media": map[string]interface{}{
				"id":        1396,
				"mediaType": "tv",
				"tmdbId":    1396,
				"title":     "Breaking Bad",
			},
			"requestedBy": map[string]interface{}{"id": 1},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	tmdbID := "1396"
	item := plugins.MediaItem{
		Type:   plugins.MediaTypeShow,
		Title:  "Breaking Bad",
		TmdbID: &tmdbID,
	}

	_, err := req.SubmitRequest(item, plugins.MortarUser{})
	if err != nil {
		t.Fatalf("SubmitRequest() error: %v", err)
	}

	if capturedBody["mediaType"] != "tv" {
		t.Errorf("expected mediaType tv, got %v", capturedBody["mediaType"])
	}
	if capturedBody["seasons"] != "all" {
		t.Errorf("expected seasons all, got %v", capturedBody["seasons"])
	}
}

// ---------------------------------------------------------------------------
// GetActivity tests
// ---------------------------------------------------------------------------

func TestGetActivity_NoFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				makeTestRequest(1, 1, 10, "movie", "Inception"),
				makeTestRequest(2, 2, 11, "tv", "Breaking Bad"),
				makeTestRequest(3, 4, 12, "movie", "The Matrix"),
			},
			"pageInfo": map[string]interface{}{"pages": 1, "pageSize": 50, "results": 3, "skip": 0},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	ar, ok := p.(plugins.ActivityReadable)
	if !ok {
		t.Fatal("plugin does not implement plugins.ActivityReadable")
	}

	events, err := ar.GetActivity(nil)
	if err != nil {
		t.Fatalf("GetActivity() error: %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}

	// Check event types.
	if events[0].Type != plugins.ActivityEventRequested {
		t.Errorf("event[0].Type: got %q, want requested", events[0].Type)
	}
	if events[1].Type != plugins.ActivityEventApproved {
		t.Errorf("event[1].Type: got %q, want approved", events[1].Type)
	}
	if events[2].Type != plugins.ActivityEventAddedToLibrary {
		t.Errorf("event[2].Type: got %q, want added_to_library", events[2].Type)
	}

	// "added_to_library" should be all_users.
	if events[2].Visibility != plugins.ActivityVisibilityAllUsers {
		t.Errorf("event[2].Visibility: got %q, want all_users", events[2].Visibility)
	}

	// "requested" and "approved" should be requester_and_admin.
	if events[0].Visibility != plugins.ActivityVisibilityRequesterAndAdmin {
		t.Errorf("event[0].Visibility: got %q, want requester_and_admin", events[0].Visibility)
	}
}

func TestGetActivity_WithSinceFilter(t *testing.T) {
	// All test requests have updatedAt = "2024-01-02T00:00:00Z".
	// Use a since value that is before that timestamp to include all, and
	// after it to exclude all.

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				makeTestRequest(1, 1, 10, "movie", "Inception"),
			},
			"pageInfo": map[string]interface{}{"pages": 1, "pageSize": 50, "results": 1, "skip": 0},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	ar, _ := p.(plugins.ActivityReadable)

	// Since is after updatedAt — event should be excluded.
	after := time.Now().UTC().Format(time.RFC3339)
	events, err := ar.GetActivity(&after)
	if err != nil {
		t.Fatalf("GetActivity() error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events after future timestamp, got %d", len(events))
	}

	// Since is before updatedAt — event should be included.
	before := "2024-01-01T00:00:00Z"
	events, err = ar.GetActivity(&before)
	if err != nil {
		t.Fatalf("GetActivity() error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event after past timestamp, got %d", len(events))
	}
}

// ---------------------------------------------------------------------------
// Manifest tests
// ---------------------------------------------------------------------------

func TestManifest(t *testing.T) {
	p := newTestPlugin(t, "http://localhost")
	m := p.Manifest()

	if m.ID != "jellyseerr" {
		t.Errorf("manifest.ID: got %q, want jellyseerr", m.ID)
	}
	if m.Type != "jellyseerr" {
		t.Errorf("manifest.Type: got %q, want jellyseerr", m.Type)
	}

	hasVideo := false
	hasActivity := false
	for _, c := range m.Capabilities {
		if c == plugins.CapabilityRequestsVideo {
			hasVideo = true
		}
		if c == plugins.CapabilityActivityRead {
			hasActivity = true
		}
	}
	if !hasVideo {
		t.Error("manifest missing requests.video capability")
	}
	if !hasActivity {
		t.Error("manifest missing activity.read capability")
	}
}

// ---------------------------------------------------------------------------
// Constructor validation tests
// ---------------------------------------------------------------------------

func TestNew_MissingURL(t *testing.T) {
	_, err := jellyseerr.New(config.PluginConfig{
		ID:     "jellyseerr",
		Type:   "jellyseerr",
		APIKey: "key",
	})
	if err == nil {
		t.Error("expected error for missing URL, got nil")
	}
}

func TestNew_MissingAPIKey(t *testing.T) {
	_, err := jellyseerr.New(config.PluginConfig{
		ID:   "jellyseerr",
		Type: "jellyseerr",
		URL:  "http://localhost",
	})
	if err == nil {
		t.Error("expected error for missing API key, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetRequest tests
// ---------------------------------------------------------------------------

func TestGetRequest_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request/42" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        42,
			"status":    2,
			"createdAt": "2024-03-01T10:00:00Z",
			"updatedAt": "2024-03-02T12:00:00Z",
			"media": map[string]interface{}{
				"id":        4200,
				"mediaType": "movie",
				"tmdbId":    4200,
				"title":     "Dune",
			},
			"requestedBy": map[string]interface{}{"id": 7},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	result, err := req.GetRequest("jellyseerr:42")
	if err != nil {
		t.Fatalf("GetRequest() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("GetRequest() returned nil, expected a result")
	}
	if result.ID != "jellyseerr:42" {
		t.Errorf("ID: got %q, want %q", result.ID, "jellyseerr:42")
	}
	if result.Status != plugins.RequestStatusApproved {
		t.Errorf("Status: got %q, want approved", result.Status)
	}
	if result.Item.Title != "Dune" {
		t.Errorf("Item.Title: got %q, want %q", result.Item.Title, "Dune")
	}
	if result.RequesterID != "7" {
		t.Errorf("RequesterID: got %q, want %q", result.RequesterID, "7")
	}
}

func TestGetRequest_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	result, err := req.GetRequest("jellyseerr:99")
	if err != nil {
		t.Fatalf("GetRequest() unexpected error on 404: %v", err)
	}
	if result != nil {
		t.Errorf("GetRequest() expected nil result for 404, got %+v", result)
	}
}

// ---------------------------------------------------------------------------
// ReviewRequest tests
// ---------------------------------------------------------------------------

func TestReviewRequest_Approve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/request/42/approve" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        42,
			"status":    2,
			"createdAt": "2024-03-01T10:00:00Z",
			"updatedAt": "2024-03-02T13:00:00Z",
			"media": map[string]interface{}{
				"id":        4200,
				"mediaType": "movie",
				"tmdbId":    4200,
				"title":     "Dune",
			},
			"requestedBy": map[string]interface{}{"id": 7},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	result, err := req.ReviewRequest("jellyseerr:42", plugins.RequestReview{Decision: "approve"})
	if err != nil {
		t.Fatalf("ReviewRequest(approve) error: %v", err)
	}
	if result.ID != "jellyseerr:42" {
		t.Errorf("ID: got %q, want %q", result.ID, "jellyseerr:42")
	}
	if result.Status != plugins.RequestStatusApproved {
		t.Errorf("Status: got %q, want approved", result.Status)
	}
}

func TestReviewRequest_Decline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/request/42/decline" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        42,
			"status":    3,
			"createdAt": "2024-03-01T10:00:00Z",
			"updatedAt": "2024-03-02T14:00:00Z",
			"media": map[string]interface{}{
				"id":        4200,
				"mediaType": "movie",
				"tmdbId":    4200,
				"title":     "Dune",
			},
			"requestedBy": map[string]interface{}{"id": 7},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	reason := "already available"
	result, err := req.ReviewRequest("jellyseerr:42", plugins.RequestReview{Decision: "decline", Reason: &reason})
	if err != nil {
		t.Fatalf("ReviewRequest(decline) error: %v", err)
	}
	if result.ID != "jellyseerr:42" {
		t.Errorf("ID: got %q, want %q", result.ID, "jellyseerr:42")
	}
	if result.Status != plugins.RequestStatusDeclined {
		t.Errorf("Status: got %q, want declined", result.Status)
	}
}

// ---------------------------------------------------------------------------
// ListRequests combined-filter test
// ---------------------------------------------------------------------------

func TestListRequests_BothFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				makeTestRequest(1, 2, 10, "movie", "Inception"),    // requester=10, approved
				makeTestRequest(2, 1, 10, "tv", "Breaking Bad"),    // requester=10, pending
				makeTestRequest(3, 2, 11, "movie", "The Matrix"),   // requester=11, approved
				makeTestRequest(4, 3, 10, "movie", "Interstellar"), // requester=10, declined
			},
			"pageInfo": map[string]interface{}{"pages": 1, "pageSize": 20, "results": 4, "skip": 0},
		})
	}))
	defer srv.Close()

	p := newTestPlugin(t, srv.URL)
	req, _ := p.(plugins.Requester)

	requesterID := "10"
	status := plugins.RequestStatusApproved
	reqs, err := req.ListRequests(plugins.ListRequestsOptions{
		RequesterID: &requesterID,
		Status:      &status,
	})
	if err != nil {
		t.Fatalf("ListRequests() error: %v", err)
	}
	// Only request #1 matches both requester=10 and status=approved.
	if len(reqs) != 1 {
		t.Errorf("expected 1 request matching both filters, got %d", len(reqs))
	}
	if len(reqs) > 0 {
		if reqs[0].RequesterID != "10" {
			t.Errorf("RequesterID: got %q, want %q", reqs[0].RequesterID, "10")
		}
		if reqs[0].Status != plugins.RequestStatusApproved {
			t.Errorf("Status: got %q, want approved", reqs[0].Status)
		}
	}
}
