# Writing a plugin

Plugins live in `internal/plugins/<type>/`. Each plugin is its own Go package.

## Steps

### 1. Read the plugin interface spec

`specs/plugins/plugin-interface.md` is the contract. Read it fully before writing any code.

### 2. Create the package

```
internal/plugins/myservice/
  plugin.go
```

### 3. Implement the base interface

Every plugin must implement `Plugin`:

```go
type Plugin interface {
    Manifest() PluginManifest
    Health(ctx context.Context) (HealthStatus, error)
}
```

`Manifest()` returns the plugin's `id`, `type`, `displayName`, and declared `capabilities`.

### 4. Declare and implement capabilities

Only declare capabilities you actually implement. For each declared capability, implement the corresponding interface.

Example — a plugin that can read downloads:

```go
func (p *MyPlugin) Manifest() plugin.PluginManifest {
    return plugin.PluginManifest{
        ID:           p.config.ID,
        Type:         "myservice",
        DisplayName:  "My Service",
        Capabilities: []string{"downloads.read"},
    }
}

// Implements DownloadsReadable
func (p *MyPlugin) ListDownloads(ctx context.Context) ([]plugin.DownloadItem, error) {
    // ...
}
```

### 5. Register the plugin type

Add a case to the plugin registry in `internal/plugins/registry.go` (or equivalent) so Mortar knows how to instantiate your plugin type from a config entry.

### 6. Document it

Add a row to `docs/plugins.md`:

```markdown
| myservice | Experimental | downloads.read | Brief description |
```

## Rules

- Use the shared types (`MediaItem`, `DownloadItem`, etc.) exactly as defined — do not add fields or create variants
- Return an error on auth failures, unreachability, and timeouts; do not swallow errors silently
- Do not call other plugins from within a plugin
- Keep all upstream HTTP calls inside the plugin package; no upstream calls from handlers or other packages
