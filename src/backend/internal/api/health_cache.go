package api

import (
	"sort"
	"sync"
	"time"

	"github.com/nbellowe/mortar/src/backend/internal/appstate"
	"github.com/nbellowe/mortar/src/backend/internal/plugins"
)

const healthRefreshInterval = 60 * time.Second

type healthCache struct {
	store    *appstate.Store
	registry *plugins.Registry

	mu      sync.RWMutex
	entries map[string]pluginHealthResponse
}

func newHealthCache(registry *plugins.Registry, store *appstate.Store) *healthCache {
	hc := &healthCache{
		store:    store,
		registry: registry,
		entries:  make(map[string]pluginHealthResponse),
	}
	hc.bootstrap()
	go hc.loop()
	return hc
}

func (hc *healthCache) bootstrap() {
	if hc.store != nil {
		if snapshots, err := hc.store.LoadHealthSnapshots(); err == nil {
			for _, plugin := range hc.registry.All() {
				manifest := plugin.Manifest()
				snapshot, ok := snapshots[manifest.ID]
				if !ok {
					continue
				}
				hc.entries[manifest.ID] = pluginHealthResponse{
					PluginID:    manifest.ID,
					PluginType:  manifest.Type,
					DisplayName: manifest.DisplayName,
					Status:      snapshot.Status,
					Reachable:   snapshot.Reachable,
					LatencyMs:   snapshot.LatencyMs,
					CheckedAt:   snapshot.CheckedAt,
					Detail:      snapshot.Detail,
				}
			}
		}
	}

	hc.refresh()
}

func (hc *healthCache) loop() {
	ticker := time.NewTicker(healthRefreshInterval)
	defer ticker.Stop()
	for range ticker.C {
		hc.refresh()
	}
}

func (hc *healthCache) refresh() {
	next := make(map[string]pluginHealthResponse)
	for _, plugin := range hc.registry.All() {
		manifest := plugin.Manifest()
		entry := buildHealthEntry(manifest, plugin)
		next[manifest.ID] = entry
		if hc.store != nil {
			_ = hc.store.RecordHealthSnapshot(manifest.ID, plugins.HealthStatus{
				Status:    entry.Status,
				Reachable: entry.Reachable,
				LatencyMs: entry.LatencyMs,
				CheckedAt: entry.CheckedAt,
				Detail:    entry.Detail,
			})
		}
	}

	hc.mu.Lock()
	hc.entries = next
	hc.mu.Unlock()
}

func (hc *healthCache) list() []pluginHealthResponse {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	out := make([]pluginHealthResponse, 0, len(hc.entries))
	for _, entry := range hc.entries {
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].PluginID < out[j].PluginID
	})
	return out
}
