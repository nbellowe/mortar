package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/nbellowe/mortar/src/backend/internal/api"
	"github.com/nbellowe/mortar/src/backend/internal/appstate"
	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/db"
)

// ---------------------------------------------------------------------------
// mockPlugin — minimal plugins.Plugin for handler tests.
// ---------------------------------------------------------------------------

// mockPlugin implements plugins.Plugin only (no Requester capability).
// Use mockPluginWithRequester when the Requester interface is also needed.
type mockPlugin struct {
	manifest   plugins.PluginManifest
	healthResp plugins.HealthStatus
	healthErr  error
}

func (m *mockPlugin) Manifest() plugins.PluginManifest { return m.manifest }

func (m *mockPlugin) Health() (plugins.HealthStatus, error) {
	return m.healthResp, m.healthErr
}

// ---------------------------------------------------------------------------
// mockRequester — injectable plugins.Requester behaviour.
// ---------------------------------------------------------------------------

type mockRequester struct {
	searchResults []plugins.MediaItem
	searchErr     error

	requests     map[string]*plugins.Request
	submitResult plugins.Request
	submitErr    error
	listResults  []plugins.Request
	listErr      error
}

func (mr *mockRequester) Search(_ string) ([]plugins.MediaItem, error) {
	return mr.searchResults, mr.searchErr
}

func (mr *mockRequester) GetRequest(id string) (*plugins.Request, error) {
	if mr.requests == nil {
		return nil, nil
	}
	r, ok := mr.requests[id]
	if !ok {
		return nil, nil
	}
	return r, nil
}

func (mr *mockRequester) ListRequests(_ plugins.ListRequestsOptions) ([]plugins.Request, error) {
	return mr.listResults, mr.listErr
}

func (mr *mockRequester) SubmitRequest(_ plugins.MediaItem, _ plugins.MortarUser) (plugins.Request, error) {
	return mr.submitResult, mr.submitErr
}

func (mr *mockRequester) ReviewRequest(_ string, _ plugins.RequestReview) (plugins.Request, error) {
	return plugins.Request{}, errors.New("not implemented in mock")
}

// ---------------------------------------------------------------------------
// mockPluginWithRequester — Plugin that also satisfies plugins.Requester.
// ---------------------------------------------------------------------------

type mockPluginWithRequester struct {
	manifest   plugins.PluginManifest
	healthResp plugins.HealthStatus
	healthErr  error
	req        mockRequester
}

func (m *mockPluginWithRequester) Manifest() plugins.PluginManifest { return m.manifest }
func (m *mockPluginWithRequester) Health() (plugins.HealthStatus, error) {
	return m.healthResp, m.healthErr
}

func (m *mockPluginWithRequester) Search(q string) ([]plugins.MediaItem, error) {
	return m.req.Search(q)
}
func (m *mockPluginWithRequester) GetRequest(id string) (*plugins.Request, error) {
	return m.req.GetRequest(id)
}
func (m *mockPluginWithRequester) ListRequests(opts plugins.ListRequestsOptions) ([]plugins.Request, error) {
	return m.req.ListRequests(opts)
}
func (m *mockPluginWithRequester) SubmitRequest(item plugins.MediaItem, user plugins.MortarUser) (plugins.Request, error) {
	return m.req.SubmitRequest(item, user)
}
func (m *mockPluginWithRequester) ReviewRequest(id string, review plugins.RequestReview) (plugins.Request, error) {
	return m.req.ReviewRequest(id, review)
}

// ---------------------------------------------------------------------------
// Registry builder helpers
// ---------------------------------------------------------------------------

// buildRegistry creates a Registry pre-populated with the given plugins and
// optional request route overrides, bypassing Init.
//
// routes maps capability string (e.g. "requests.video") to plugin id.
func buildRegistry(ps map[string]plugins.Plugin, routes map[string]string) *plugins.Registry {
	reg := plugins.NewRegistry()
	reg.InjectForTest(ps, routes)
	return reg
}

// newTestServer creates an httptest.Server backed by the Mortar router.
// The database is nil because none of the handler tests exercise DB calls.
func newTestServer(reg *plugins.Registry) *httptest.Server {
	router := api.NewRouter(&config.Config{}, reg, nil, fstest.MapFS{})
	return httptest.NewServer(router)
}

func newTestServerWithConfig(t *testing.T, cfg *config.Config, reg *plugins.Registry) (*httptest.Server, *db.DB) {
	t.Helper()

	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := appstate.New(database).SyncUsersFromConfig(cfg.Users); err != nil {
		t.Fatalf("sync users: %v", err)
	}

	router := api.NewRouter(cfg, reg, database, fstest.MapFS{})
	server := httptest.NewServer(router)
	t.Cleanup(func() {
		server.Close()
		_ = database.Close()
	})
	return server, database
}

// jsonBody marshals v to JSON and returns an *http.Request with it as the
// body. Panics on encoding failure (acceptable in tests).
func jsonBody(method, url string, v interface{}) *http.Request {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}
