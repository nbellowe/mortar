package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// requesterReg builds a Registry with a single mockPluginWithRequester
// routed as requests.video, pre-populated with the given state.
func requesterReg(mp *mockPluginWithRequester) *plugins.Registry {
	return buildRegistry(
		map[string]plugins.Plugin{"jellyseerr": mp},
		map[string]string{"requests.video": "jellyseerr"},
	)
}

// defaultRequesterPlugin returns a mockPluginWithRequester with sensible
// defaults for tests that don't need to control the health response.
func defaultRequesterPlugin() *mockPluginWithRequester {
	return &mockPluginWithRequester{
		manifest: plugins.PluginManifest{
			ID:           "jellyseerr",
			Type:         "jellyseerr",
			DisplayName:  "Jellyseerr",
			Capabilities: []plugins.Capability{plugins.CapabilityRequestsVideo},
		},
		healthResp: plugins.HealthStatus{Status: "healthy", Reachable: true},
	}
}

type searchResp struct {
	Items            []plugins.MediaItem    `json:"items"`
	FailedPlugins    []string               `json:"failed_plugins"`
	ExistingRequests []plugins.Request      `json:"existing_requests"`
	AvailableMatches []plugins.LibraryMatch `json:"available_matches"`
}

type requestsResp struct {
	Items      []plugins.Request `json:"items"`
	ReviewURLs map[string]string `json:"review_urls"`
}

// ---------------------------------------------------------------------------
// GET /api/v1/search
// ---------------------------------------------------------------------------

func TestSearch_ReturnsTwoResults(t *testing.T) {
	year := 2012
	mp := defaultRequesterPlugin()
	mp.req.searchResults = []plugins.MediaItem{
		{
			ID:       "jellyseerr:1",
			Title:    "The Avengers",
			Type:     plugins.MediaTypeMovie,
			Year:     &year,
			PluginID: "jellyseerr",
		},
		{
			ID:       "jellyseerr:2",
			Title:    "Avengers: Endgame",
			Type:     plugins.MediaTypeMovie,
			PluginID: "jellyseerr",
		},
	}

	srv := newTestServer(requesterReg(mp))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/search?q=avengers")
	if err != nil {
		t.Fatalf("GET /api/v1/search: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var payload searchResp
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(payload.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(payload.Items))
	}
	if payload.Items[0].Title != "The Avengers" {
		t.Errorf("items[0].Title: got %q, want The Avengers", payload.Items[0].Title)
	}
	if payload.Items[1].Title != "Avengers: Endgame" {
		t.Errorf("items[1].Title: got %q, want Avengers: Endgame", payload.Items[1].Title)
	}
	if payload.FailedPlugins == nil {
		t.Fatal("expected failed_plugins to be a non-nil slice")
	}
}

func TestSearch_MissingQ_Returns400(t *testing.T) {
	mp := defaultRequesterPlugin()
	srv := newTestServer(requesterReg(mp))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/search")
	if err != nil {
		t.Fatalf("GET /api/v1/search: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing q, got %d", resp.StatusCode)
	}
}

func TestSearch_NoPlugin_Returns503(t *testing.T) {
	// Registry with no requests.video route.
	reg := buildRegistry(nil, nil)
	srv := newTestServer(reg)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/search?q=test")
	if err != nil {
		t.Fatalf("GET /api/v1/search: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when no plugin configured, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/requests
// ---------------------------------------------------------------------------

func TestListRequests_ReturnsResults(t *testing.T) {
	mp := defaultRequesterPlugin()
	mp.req.listResults = []plugins.Request{
		{
			ID:          "jellyseerr:10",
			PluginID:    "jellyseerr",
			Status:      plugins.RequestStatusPending,
			RequesterID: "user1",
			SubmittedAt: "2026-05-01T00:00:00Z",
			UpdatedAt:   "2026-05-01T00:00:00Z",
			Item: plugins.MediaItem{
				ID:    "jellyseerr:100",
				Title: "Inception",
				Type:  plugins.MediaTypeMovie,
			},
		},
	}

	srv := newTestServer(requesterReg(mp))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/requests")
	if err != nil {
		t.Fatalf("GET /api/v1/requests: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var payload requestsResp
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 request, got %d", len(payload.Items))
	}
	if payload.Items[0].ID != "jellyseerr:10" {
		t.Errorf("ID: got %q, want jellyseerr:10", payload.Items[0].ID)
	}
	if payload.Items[0].Status != plugins.RequestStatusPending {
		t.Errorf("Status: got %q, want pending", payload.Items[0].Status)
	}
}

func TestListRequests_EmptyList(t *testing.T) {
	mp := defaultRequesterPlugin()
	// listResults is nil by default — should return empty JSON array.
	srv := newTestServer(requesterReg(mp))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/requests")
	if err != nil {
		t.Fatalf("GET /api/v1/requests: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var payload requestsResp
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Items == nil {
		t.Error("expected non-nil (empty) slice, got null JSON")
	}
	if len(payload.Items) != 0 {
		t.Errorf("expected 0 requests, got %d", len(payload.Items))
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/requests
// ---------------------------------------------------------------------------

func TestSubmitRequest_Returns201(t *testing.T) {
	mp := defaultRequesterPlugin()
	mp.req.submitResult = plugins.Request{
		ID:          "jellyseerr:99",
		PluginID:    "jellyseerr",
		Status:      plugins.RequestStatusPending,
		RequesterID: "anonymous",
		SubmittedAt: "2026-05-08T00:00:00Z",
		UpdatedAt:   "2026-05-08T00:00:00Z",
		Item: plugins.MediaItem{
			ID:    "jellyseerr:42",
			Title: "Dune",
			Type:  plugins.MediaTypeMovie,
		},
	}

	srv := newTestServer(requesterReg(mp))
	defer srv.Close()

	payload := map[string]string{
		"media_id": "jellyseerr:42",
		"type":     "movie",
	}
	req := jsonBody(http.MethodPost, srv.URL+"/api/v1/requests", payload)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/requests: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)
	var created plugins.Request
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if created.ID != "jellyseerr:99" {
		t.Errorf("ID: got %q, want jellyseerr:99", created.ID)
	}
	if created.Status != plugins.RequestStatusPending {
		t.Errorf("Status: got %q, want pending", created.Status)
	}
}

func TestSubmitRequest_MissingBody_Returns400(t *testing.T) {
	mp := defaultRequesterPlugin()
	srv := newTestServer(requesterReg(mp))
	defer srv.Close()

	req := jsonBody(http.MethodPost, srv.URL+"/api/v1/requests", map[string]string{
		"type": "movie",
		// media_id intentionally omitted
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/requests: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing media_id, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/requests/{id}
// ---------------------------------------------------------------------------

func TestGetRequest_Found_Returns200(t *testing.T) {
	r := &plugins.Request{
		ID:          "jellyseerr:42",
		PluginID:    "jellyseerr",
		Status:      plugins.RequestStatusApproved,
		RequesterID: "user1",
		SubmittedAt: "2026-05-01T00:00:00Z",
		UpdatedAt:   "2026-05-02T00:00:00Z",
		Item: plugins.MediaItem{
			ID:    "jellyseerr:4200",
			Title: "Dune",
			Type:  plugins.MediaTypeMovie,
		},
	}

	mp := defaultRequesterPlugin()
	mp.req.requests = map[string]*plugins.Request{
		"jellyseerr:42": r,
	}

	srv := newTestServer(requesterReg(mp))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/requests/jellyseerr:42")
	if err != nil {
		t.Fatalf("GET /api/v1/requests/jellyseerr:42: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got plugins.Request
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != "jellyseerr:42" {
		t.Errorf("ID: got %q, want jellyseerr:42", got.ID)
	}
	if got.Status != plugins.RequestStatusApproved {
		t.Errorf("Status: got %q, want approved", got.Status)
	}
	if got.Item.Title != "Dune" {
		t.Errorf("Item.Title: got %q, want Dune", got.Item.Title)
	}
}

func TestGetRequest_NotFound_Returns404(t *testing.T) {
	mp := defaultRequesterPlugin()
	// No requests registered — any id returns nil.
	srv := newTestServer(requesterReg(mp))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/requests/jellyseerr:99")
	if err != nil {
		t.Fatalf("GET /api/v1/requests/jellyseerr:99: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown id, got %d", resp.StatusCode)
	}
}
