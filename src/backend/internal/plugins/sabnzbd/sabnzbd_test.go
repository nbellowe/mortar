package sabnzbd_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nbellowe/mortar/src/backend/internal/config"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
	"github.com/nbellowe/mortar/src/backend/internal/plugins/sabnzbd"
)

// newTestPlugin creates a Plugin pointed at the given test server URL.
func newTestPlugin(t *testing.T, serverURL string) plugins.Plugin {
	t.Helper()
	cfg := config.PluginConfig{
		ID:     "test-sabnzbd",
		Type:   "sabnzbd",
		URL:    serverURL,
		APIKey: "testapikey",
	}
	p, err := sabnzbd.New(cfg)
	if err != nil {
		t.Fatalf("sabnzbd.New: %v", err)
	}
	return p
}

// modeHandler returns an http.HandlerFunc that dispatches on the "mode" query
// param, calling the provided handlers map.
func modeHandler(handlers map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("mode")
		h, ok := handlers[mode]
		if !ok {
			http.Error(w, "unknown mode: "+mode, http.StatusBadRequest)
			return
		}
		h(w, r)
	}
}

// ---------------------------------------------------------------------------
// Health tests
// ---------------------------------------------------------------------------

func TestHealth_Happy(t *testing.T) {
	ts := httptest.NewServer(modeHandler(map[string]http.HandlerFunc{
		"version": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"version":"3.7.2"}`))
		},
	}))
	defer ts.Close()

	p := newTestPlugin(t, ts.URL)
	status, err := p.Health()
	if err != nil {
		t.Fatalf("Health() returned unexpected error: %v", err)
	}
	if status.Status != "healthy" {
		t.Errorf("Status = %q, want %q", status.Status, "healthy")
	}
	if !status.Reachable {
		t.Error("Reachable = false, want true")
	}
	if status.CheckedAt == "" {
		t.Error("CheckedAt is empty")
	}
}

func TestHealth_ServerDown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	p := newTestPlugin(t, ts.URL)
	status, err := p.Health()
	if err != nil {
		t.Fatalf("Health() returned unexpected error: %v", err)
	}
	if status.Status != "unreachable" {
		t.Errorf("Status = %q, want %q", status.Status, "unreachable")
	}
	if status.Reachable {
		t.Error("Reachable = true, want false")
	}
	if status.Detail == nil || *status.Detail == "" {
		t.Error("Detail should be non-nil and non-empty on unreachable status")
	}
}

// ---------------------------------------------------------------------------
// GetQueue tests
// ---------------------------------------------------------------------------

func TestGetQueue_FullMapping(t *testing.T) {
	// Queue JSON with 2 slots covering all the mapping cases listed in the spec.
	queueJSON := `{
		"queue": {
			"kbps": "2048",
			"slots": [
				{
					"nzo_id":     "NZO1",
					"filename":   "Movie.2024.mkv",
					"status":     "Downloading",
					"percentage": "42",
					"mb":         "1024",
					"mbleft":     "594",
					"timeleft":   "01:30:00"
				},
				{
					"nzo_id":     "NZO2",
					"filename":   "Album.zip",
					"status":     "Paused",
					"percentage": "0",
					"mb":         "512",
					"mbleft":     "512",
					"timeleft":   "0:00:00"
				}
			]
		}
	}`

	ts := httptest.NewServer(modeHandler(map[string]http.HandlerFunc{
		"queue": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(queueJSON))
		},
	}))
	defer ts.Close()

	p := newTestPlugin(t, ts.URL)
	dp, ok := p.(plugins.DownloadsReadable)
	if !ok {
		t.Fatal("plugin does not implement DownloadsReadable")
	}

	items, err := dp.GetQueue()
	if err != nil {
		t.Fatalf("GetQueue() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	// --- Slot 1: Downloading ---
	item0 := items[0]

	if item0.ID != "test-sabnzbd:NZO1" {
		t.Errorf("items[0].ID = %q, want %q", item0.ID, "test-sabnzbd:NZO1")
	}
	if item0.Name != "Movie.2024.mkv" {
		t.Errorf("items[0].Name = %q, want %q", item0.Name, "Movie.2024.mkv")
	}

	// percentage "42" → Progress 0.42
	const wantProgress = 0.42
	if diff := item0.Progress - wantProgress; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("items[0].Progress = %v, want %v", item0.Progress, wantProgress)
	}

	// mb "1024" → SizeBytes = 1024 * 1024 * 1024
	const wantSizeBytes = int64(1024 * 1024 * 1024)
	if item0.SizeBytes != wantSizeBytes {
		t.Errorf("items[0].SizeBytes = %d, want %d", item0.SizeBytes, wantSizeBytes)
	}

	// kbps "2048" → SpeedBytesS = 2048 * 1024
	const wantSpeed = int64(2048 * 1024)
	if item0.SpeedBytesS != wantSpeed {
		t.Errorf("items[0].SpeedBytesS = %d, want %d", item0.SpeedBytesS, wantSpeed)
	}

	// timeleft "01:30:00" → EtaSeconds = 5400
	if item0.EtaSeconds == nil {
		t.Error("items[0].EtaSeconds is nil, want 5400")
	} else if *item0.EtaSeconds != 5400 {
		t.Errorf("items[0].EtaSeconds = %d, want 5400", *item0.EtaSeconds)
	}

	if item0.Status != "downloading" {
		t.Errorf("items[0].Status = %q, want %q", item0.Status, "downloading")
	}

	// --- Slot 2: Paused, zero timeleft ---
	item1 := items[1]

	if item1.ID != "test-sabnzbd:NZO2" {
		t.Errorf("items[1].ID = %q, want %q", item1.ID, "test-sabnzbd:NZO2")
	}
	if item1.Status != "paused" {
		t.Errorf("items[1].Status = %q, want %q", item1.Status, "paused")
	}
	// timeleft "0:00:00" → EtaSeconds should be nil
	if item1.EtaSeconds != nil {
		t.Errorf("items[1].EtaSeconds = %d, want nil", *item1.EtaSeconds)
	}
}

func TestGetQueue_StatusMappings(t *testing.T) {
	cases := []struct {
		sabStatus  string
		wantStatus string
	}{
		{"Downloading", "downloading"},
		{"Paused", "paused"},
		{"Extracting", "processing"},
		{"Verifying", "processing"},
		{"Repairing", "processing"},
		{"Moving", "processing"},
		{"Failed", "failed"},
		{"Queued", "queued"},
		{"UnknownStatus", "queued"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.sabStatus, func(t *testing.T) {
			slot := map[string]interface{}{
				"nzo_id":     "NZO1",
				"filename":   "test.nzb",
				"status":     tc.sabStatus,
				"percentage": "0",
				"mb":         "100",
				"mbleft":     "100",
				"timeleft":   "0:00:00",
			}
			queuePayload := map[string]interface{}{
				"queue": map[string]interface{}{
					"kbps":  "0",
					"slots": []interface{}{slot},
				},
			}
			body, _ := json.Marshal(queuePayload)

			ts := httptest.NewServer(modeHandler(map[string]http.HandlerFunc{
				"queue": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write(body)
				},
			}))
			defer ts.Close()

			p := newTestPlugin(t, ts.URL)
			dp := p.(plugins.DownloadsReadable)
			items, err := dp.GetQueue()
			if err != nil {
				t.Fatalf("GetQueue(): %v", err)
			}
			if len(items) != 1 {
				t.Fatalf("len(items) = %d, want 1", len(items))
			}
			if items[0].Status != tc.wantStatus {
				t.Errorf("Status = %q, want %q", items[0].Status, tc.wantStatus)
			}
		})
	}
}

func TestGetQueue_Empty(t *testing.T) {
	ts := httptest.NewServer(modeHandler(map[string]http.HandlerFunc{
		"queue": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"queue":{"kbps":"0","slots":[]}}`))
		},
	}))
	defer ts.Close()

	p := newTestPlugin(t, ts.URL)
	dp := p.(plugins.DownloadsReadable)

	items, err := dp.GetQueue()
	if err != nil {
		t.Fatalf("GetQueue() unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("len(items) = %d, want 0", len(items))
	}
}
