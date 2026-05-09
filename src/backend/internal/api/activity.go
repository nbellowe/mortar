// Package api — activity feed endpoint.
package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

// activityResponse is the JSON body returned by GET /api/v1/activity.
type activityResponse struct {
	Events        []plugins.ActivityEvent `json:"events"`
	FailedPlugins []string                `json:"failed_plugins"`
}

// handleActivity handles GET /api/v1/activity.
// It fans out to all plugins implementing ActivityReadable, merges the results,
// sorts them by timestamp descending, and always returns HTTP 200.
// Plugin-level errors are reported in the failed_plugins field.
func (h *handler) handleActivity(w http.ResponseWriter, r *http.Request) {
	sinceParam := r.URL.Query().Get("since")
	var since *string
	if sinceParam != "" {
		since = &sinceParam
	}

	all := h.registry.All()

	var (
		mu            sync.Mutex
		events        []plugins.ActivityEvent
		failedPlugins []string
		wg            sync.WaitGroup
	)

	for _, p := range all {
		ar, ok := p.(plugins.ActivityReadable)
		if !ok {
			continue
		}

		manifest := p.Manifest()
		wg.Add(1)
		go func(ar plugins.ActivityReadable, displayName string) {
			defer wg.Done()
			evts, err := ar.GetActivity(since)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failedPlugins = append(failedPlugins, displayName)
				return
			}
			events = append(events, evts...)
		}(ar, manifest.DisplayName)
	}

	wg.Wait()

	user := currentUser(r)
	if user != nil {
		events = h.filterVisibleEvents(events, *user)
	}

	// Sort events by timestamp descending; events with unparseable timestamps
	// sort last.
	sort.SliceStable(events, func(i, j int) bool {
		ti, erri := time.Parse(time.RFC3339, events[i].Timestamp)
		tj, errj := time.Parse(time.RFC3339, events[j].Timestamp)
		if erri != nil && errj != nil {
			return false
		}
		if erri != nil {
			return false
		}
		if errj != nil {
			return true
		}
		return ti.After(tj)
	})

	if events == nil {
		events = []plugins.ActivityEvent{}
	}
	if failedPlugins == nil {
		failedPlugins = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(activityResponse{
		Events:        events,
		FailedPlugins: failedPlugins,
	})
}

func (h *handler) filterVisibleEvents(events []plugins.ActivityEvent, user plugins.MortarUser) []plugins.ActivityEvent {
	filtered := make([]plugins.ActivityEvent, 0, len(events))
	for _, event := range events {
		switch event.Visibility {
		case plugins.ActivityVisibilityAllUsers:
			filtered = append(filtered, event)
		case plugins.ActivityVisibilityAdminOnly:
			if user.Role == "admin" {
				filtered = append(filtered, event)
			}
		case plugins.ActivityVisibilityRequesterAndAdmin:
			if user.Role == "admin" {
				filtered = append(filtered, event)
				continue
			}
			if event.ActorUserID == nil || h.store == nil {
				continue
			}
			mortarUserID, err := h.store.LookupMortarUserIDByExternalAccount(event.SourcePlugin, *event.ActorUserID)
			if err == nil && mortarUserID == user.ID {
				cloned := event
				cloned.ActorUserID = &mortarUserID
				filtered = append(filtered, cloned)
			}
		}
	}
	return filtered
}
