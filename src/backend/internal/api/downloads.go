// Package api — download queue endpoint.
package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// downloadsResponse is the JSON body returned by GET /api/v1/downloads.
type downloadsResponse struct {
	Items         []plugins.DownloadItem `json:"items"`
	FailedPlugins []string               `json:"failed_plugins"`
}

// downloadStatusPriority maps status strings to sort priority (lower = first).
var downloadStatusPriority = map[string]int{
	"downloading": 0,
	"queued":      1,
	"processing":  2,
	"paused":      3,
	"failed":      4,
}

func downloadPriority(status string) int {
	if p, ok := downloadStatusPriority[status]; ok {
		return p
	}
	// Unknown statuses sort after all known ones.
	return len(downloadStatusPriority)
}

// handleDownloads handles GET /api/v1/downloads.
// It fans out to all plugins implementing DownloadsReadable, merges the
// results, sorts them by status priority, and always returns HTTP 200.
// Plugin-level errors are reported in the failed_plugins field.
func (h *handler) handleDownloads(w http.ResponseWriter, _ *http.Request) {
	all := h.registry.All()

	var (
		mu            sync.Mutex
		items         []plugins.DownloadItem
		failedPlugins []string
		wg            sync.WaitGroup
	)

	for _, p := range all {
		dr, ok := p.(plugins.DownloadsReadable)
		if !ok {
			continue
		}

		manifest := p.Manifest()
		wg.Add(1)
		go func(dr plugins.DownloadsReadable, displayName string) {
			defer wg.Done()
			queue, err := dr.GetQueue()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failedPlugins = append(failedPlugins, displayName)
				return
			}
			items = append(items, queue...)
		}(dr, manifest.DisplayName)
	}

	wg.Wait()

	sort.SliceStable(items, func(i, j int) bool {
		return downloadPriority(items[i].Status) < downloadPriority(items[j].Status)
	})

	if items == nil {
		items = []plugins.DownloadItem{}
	}
	if failedPlugins == nil {
		failedPlugins = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(downloadsResponse{
		Items:         items,
		FailedPlugins: failedPlugins,
	})
}
