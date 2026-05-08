// Package plugins provides the plugin registry that holds all registered
// plugin instances and validates routing configuration at startup.
package plugins

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nbellowe/mortar/src/backend/internal/config"
)

// Factory is a function that constructs a Plugin from its config entry.
type Factory func(cfg config.PluginConfig) (Plugin, error)

// Registry holds registered plugin instances and their resolved routing.
type Registry struct {
	plugins   map[string]Plugin  // keyed by plugin id
	factories map[string]Factory // keyed by plugin type

	// Resolved request routing: capability string → plugin id
	requestRoutes map[string]string
}

// NewRegistry creates an empty Registry with no factories registered.
// Call RegisterFactory for each plugin type before calling Init.
func NewRegistry() *Registry {
	return &Registry{
		plugins:       make(map[string]Plugin),
		factories:     make(map[string]Factory),
		requestRoutes: make(map[string]string),
	}
}

// RegisterFactory associates a plugin type name with its constructor.
// This must be called at program startup, before Init.
func (r *Registry) RegisterFactory(pluginType string, factory Factory) {
	r.factories[pluginType] = factory
}

// Init instantiates plugins from cfg, calls health on each, and
// validates routing configuration. Returns an error if any plugin type
// is unknown, any routing config is invalid, or any routing ambiguity
// is unresolved.
func (r *Registry) Init(cfg *config.Config) error {
	// Instantiate plugins.
	for _, pc := range cfg.Plugins {
		factory, ok := r.factories[pc.Type]
		if !ok {
			return fmt.Errorf("registry: unknown plugin type %q (id: %q)", pc.Type, pc.ID)
		}
		plugin, err := factory(pc)
		if err != nil {
			return fmt.Errorf("registry: failed to create plugin %q (%s): %w", pc.ID, pc.Type, err)
		}
		r.plugins[pc.ID] = plugin
	}

	// Validate and resolve routing.
	if err := r.resolveRouting(cfg); err != nil {
		return err
	}

	// Call health on each plugin and log the result. Non-fatal at startup.
	for id, p := range r.plugins {
		h, err := p.Health()
		if err != nil {
			fmt.Printf("registry: health check failed for plugin %q: %v\n", id, err)
			continue
		}
		fmt.Printf("registry: plugin %q health: %s (latency %dms)\n", id, h.Status, h.LatencyMs)
	}

	return nil
}

// resolveRouting builds the requestRoutes map and validates the config.
func (r *Registry) resolveRouting(cfg *config.Config) error {
	capabilities := []struct {
		name       string
		configured string
	}{
		{"requests.video", cfg.Routing.Requests.Video},
		{"requests.audio", cfg.Routing.Requests.Audio},
		{"requests.ebook", cfg.Routing.Requests.Ebook},
	}

	for _, cap := range capabilities {
		candidates := r.pluginsWithCapability(Capability(cap.name))

		switch {
		case len(candidates) == 0:
			// No plugin supports this capability — that request type is unavailable.
			// This is allowed; the UI simply won't surface it.
			continue

		case len(candidates) == 1 && cap.configured == "":
			// Exactly one candidate; auto-route.
			r.requestRoutes[cap.name] = candidates[0]

		case len(candidates) == 1 && cap.configured != "":
			// Explicit route provided for a single-candidate capability.
			// Still validate that the named plugin exists and has the capability.
			if err := r.validateRoute(cap.name, cap.configured); err != nil {
				return err
			}
			r.requestRoutes[cap.name] = cap.configured

		case len(candidates) > 1 && cap.configured == "":
			// Ambiguous — multiple candidates and no explicit routing.
			return fmt.Errorf(
				"registry: multiple plugins support %q (%s) but no routing config specifies which to use; "+
					"add a routing.requests.%s entry to your config",
				cap.name,
				strings.Join(candidates, ", "),
				routingKey(cap.name),
			)

		case len(candidates) > 1 && cap.configured != "":
			if err := r.validateRoute(cap.name, cap.configured); err != nil {
				return err
			}
			r.requestRoutes[cap.name] = cap.configured
		}
	}

	return nil
}

// validateRoute confirms that pluginID exists and declares the given capability.
func (r *Registry) validateRoute(capabilityName, pluginID string) error {
	plugin, ok := r.plugins[pluginID]
	if !ok {
		return fmt.Errorf(
			"registry: routing.requests.%s references unknown plugin id %q",
			routingKey(capabilityName), pluginID,
		)
	}
	manifest := plugin.Manifest()
	for _, c := range manifest.Capabilities {
		if string(c) == capabilityName {
			return nil
		}
	}
	return fmt.Errorf(
		"registry: routing.requests.%s references plugin %q, which does not declare the %q capability",
		routingKey(capabilityName), pluginID, capabilityName,
	)
}

// pluginsWithCapability returns the IDs of all plugins that declare cap,
// sorted alphabetically for deterministic output.
func (r *Registry) pluginsWithCapability(cap Capability) []string {
	var out []string
	for id, p := range r.plugins {
		for _, c := range p.Manifest().Capabilities {
			if c == cap {
				out = append(out, id)
				break
			}
		}
	}
	sort.Strings(out)
	return out
}

// routingKey converts "requests.video" → "video".
func routingKey(capabilityName string) string {
	parts := strings.SplitN(capabilityName, ".", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return capabilityName
}

// Get returns the plugin with the given id, or nil if not found.
func (r *Registry) Get(id string) Plugin {
	return r.plugins[id]
}

// All returns all registered plugins.
func (r *Registry) All() map[string]Plugin {
	out := make(map[string]Plugin, len(r.plugins))
	for k, v := range r.plugins {
		out[k] = v
	}
	return out
}

// RouteRequest returns the plugin ID that should handle a request of the
// given capability (e.g. "requests.video"). Returns an empty string if no
// route is configured for that capability.
func (r *Registry) RouteRequest(capability string) string {
	return r.requestRoutes[capability]
}
