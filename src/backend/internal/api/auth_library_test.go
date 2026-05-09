package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

func loginCookie(t *testing.T, baseURL, username, password string) *http.Cookie {
	t.Helper()

	payload, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		t.Fatalf("marshal login payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/auth/login", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("new login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 login, got %d: %s", resp.StatusCode, body)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "mortar_session" {
			return cookie
		}
	}
	t.Fatal("expected mortar_session cookie")
	return nil
}

func TestAuthSessionAndAdminHealth(t *testing.T) {
	reg := buildRegistry(map[string]plugins.Plugin{
		"jellyfin": &mockPlugin{
			manifest: plugins.PluginManifest{
				ID:          "jellyfin",
				Type:        "jellyfin",
				DisplayName: "Jellyfin",
			},
			healthResp: plugins.HealthStatus{Status: "healthy", Reachable: true, CheckedAt: "2026-05-08T10:00:00Z"},
		},
	}, nil)

	cfg := &config.Config{
		Users: []config.UserConfig{
			{Username: "admin", Password: "secret", Role: "admin"},
			{Username: "alice", Password: "secret", Role: "user"},
		},
	}
	srv, _ := newTestServerWithConfig(t, cfg, reg)

	userCookie := loginCookie(t, srv.URL, "alice", "secret")

	sessionReq, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/auth/session", nil)
	if err != nil {
		t.Fatalf("session request: %v", err)
	}
	sessionReq.AddCookie(userCookie)
	sessionResp, err := http.DefaultClient.Do(sessionReq)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	defer sessionResp.Body.Close()

	if sessionResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 session, got %d", sessionResp.StatusCode)
	}

	var sessionBody struct {
		User plugins.MortarUser `json:"user"`
	}
	if err := json.NewDecoder(sessionResp.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if sessionBody.User.Username != "alice" {
		t.Fatalf("expected alice session, got %q", sessionBody.User.Username)
	}

	userHealthReq, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/health", nil)
	if err != nil {
		t.Fatalf("health request: %v", err)
	}
	userHealthReq.AddCookie(userCookie)
	userHealthResp, err := http.DefaultClient.Do(userHealthReq)
	if err != nil {
		t.Fatalf("user health: %v", err)
	}
	defer userHealthResp.Body.Close()

	if userHealthResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 health for regular user, got %d", userHealthResp.StatusCode)
	}

	adminCookie := loginCookie(t, srv.URL, "admin", "secret")
	adminHealthReq, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/health", nil)
	if err != nil {
		t.Fatalf("admin health request: %v", err)
	}
	adminHealthReq.AddCookie(adminCookie)
	adminHealthResp, err := http.DefaultClient.Do(adminHealthReq)
	if err != nil {
		t.Fatalf("admin health: %v", err)
	}
	defer adminHealthResp.Body.Close()

	if adminHealthResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 health for admin, got %d", adminHealthResp.StatusCode)
	}
}

func TestLibraryBrowseRequiresLink(t *testing.T) {
	lib := &mockLibraryPlugin{
		manifest: plugins.PluginManifest{
			ID:           "jellyfin",
			Type:         "jellyfin",
			DisplayName:  "Jellyfin",
			Capabilities: []plugins.Capability{plugins.CapabilityLibraryBrowse},
		},
		healthResp: plugins.HealthStatus{Status: "healthy", Reachable: true, CheckedAt: "2026-05-08T10:00:00Z"},
	}

	cfg := &config.Config{
		Users: []config.UserConfig{
			{Username: "alice", Password: "secret", Role: "user"},
		},
	}

	srv, _ := newTestServerWithConfig(t, cfg, buildRegistry(map[string]plugins.Plugin{"jellyfin": lib}, nil))
	cookie := loginCookie(t, srv.URL, "alice", "secret")

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/library", nil)
	if err != nil {
		t.Fatalf("new library request: %v", err)
	}
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("library request: %v", err)
	}
	defer resp.Body.Close()

	var body struct {
		RequiresLink bool `json:"requires_link"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode library response: %v", err)
	}
	if !body.RequiresLink {
		t.Fatal("expected requires_link=true when no external account link exists")
	}
}

func TestLibraryPlayUsesLinkedAccount(t *testing.T) {
	item := &plugins.MediaItem{
		ID:         "jellyfin:1",
		ExternalID: "1",
		PluginID:   "jellyfin",
		Type:       plugins.MediaTypeMovie,
		Title:      "Dune",
	}

	lib := &mockLibraryPlugin{
		manifest: plugins.PluginManifest{
			ID:           "jellyfin",
			Type:         "jellyfin",
			DisplayName:  "Jellyfin",
			Capabilities: []plugins.Capability{plugins.CapabilityLibraryBrowse},
		},
		healthResp: plugins.HealthStatus{Status: "healthy", Reachable: true, CheckedAt: "2026-05-08T10:00:00Z"},
		getItem:    item,
		playURL:    "http://jellyfin.example/web/index.html#!/details?id=1",
	}

	cfg := &config.Config{
		Plugins: []config.PluginConfig{{ID: "jellyfin", Type: "jellyfin", URL: "http://jellyfin.example"}},
		Users: []config.UserConfig{
			{
				Username: "alice",
				Password: "secret",
				Role:     "user",
				ExternalAccounts: []config.ExternalAccountConfig{
					{PluginID: "jellyfin", ExternalUserID: "alice-jellyfin"},
				},
			},
		},
	}

	srv, _ := newTestServerWithConfig(t, cfg, buildRegistry(map[string]plugins.Plugin{"jellyfin": lib}, nil))
	cookie := loginCookie(t, srv.URL, "alice", "secret")

	req := jsonBody(http.MethodPost, srv.URL+"/api/v1/library/play", map[string]string{"item_id": "jellyfin:1"})
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("play request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 play response, got %d: %s", resp.StatusCode, body)
	}

	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode play response: %v", err)
	}
	if body.URL != lib.playURL {
		t.Fatalf("expected play URL %q, got %q", lib.playURL, body.URL)
	}
}
