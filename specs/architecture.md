# Architecture

## Overview

Mortar has three layers:

```
Supported `v1.0` client: Web (Expo web)
Experimental `v0.x` clients: iOS / Android / macOS / Windows (Expo)
  └── Mortar Server (HTTP API + plugin runtime)
        ├── Plugin: Jellyseerr
        ├── Plugin: Jellyfin
        ├── Plugin: SABnzbd
        ├── Plugin: qBittorrent
        ├── Plugin: Sonarr
        ├── Plugin: Radarr
        └── Plugin: AudioBookRequest
```

The server holds API keys, proxies requests to backend services, normalizes responses, and exposes the Mortar API consumed by every client. Web and native clients never talk directly to backend services.

Web is the only supported client target for `v1.0`. Native and desktop clients remain experimental until a later milestone.
The plugin list above illustrates the broader plugin model; the narrower `v1.0` reference stack is defined in [`ROADMAP.md`](../ROADMAP.md).

## Plugin model

Every integration is a plugin. Plugins declare their capabilities via a set of flags, and Mortar activates only the UI components those flags support.

See [`plugins/plugin-interface.md`](plugins/plugin-interface.md) for the full interface spec.

### Capability flags

| Flag | Description |
|---|---|
| `requests.video` | Can accept and track video requests |
| `requests.audio` | Can accept and track audiobook requests |
| `requests.ebook` | Can accept and track ebook requests |
| `library.browse` | Can enumerate library content |
| `library.exists` | Can resolve whether a media item already exists in the library |
| `library.resume` | Can return user-specific in-progress playback items |
| `downloads.read` | Can report current download queue and progress |
| `activity.read` | Can report a stream of recent events |

Health is not a capability flag. Every plugin implements the base `health()` contract.

### Plugin configuration

Each plugin is declared in the owner config file with a `type` and its connection details:

```yaml
plugins:
  - id: jellyseerr
    type: jellyseerr
    url: http://jellyseerr:5055
    api_key: ${JELLYSEERR_API_KEY}

  - id: jellyfin
    type: jellyfin
    url: http://jellyfin:8096
    api_key: ${JELLYFIN_API_KEY}

  - id: sabnzbd
    type: sabnzbd
    url: http://sabnzbd:8080
    api_key: ${SABNZBD_API_KEY}

  - id: sonarr
    type: sonarr
    url: http://sonarr:8989
    api_key: ${SONARR_API_KEY}

  - id: audiobookrequest
    type: audiobookrequest
    url: http://audiobookrequest:8080
    api_key: ${ABR_API_KEY}
```

Multiple plugins of the same type are allowed (e.g. two Sonarr instances).

## Auth model

Two separate auth concerns:

**Mortar user auth** — users log in to Mortar with a username/password (or SSO in a future version). Mortar maintains its own user table with roles: `admin` and `user`, plus durable server-side sessions in local SQLite.

**Service API keys** — stored in the Mortar server config, never exposed to the browser. The server proxies all service API calls. Users interact only with Mortar's API.

For `v1`, the supported web client authenticates using an `HttpOnly` session cookie. Native-client auth is intentionally deferred until native clients become a supported surface.

## Request routing

When a user submits a request, Mortar routes it to the appropriate plugin based on content type and the configured routing policy for that capability:

```
video request   → configured `requests.video` route
audio request   → configured `requests.audio` route
ebook request   → configured `requests.ebook` route
```

Routing lives in a top-level `routing` section rather than inside plugin definitions:

```yaml
routing:
  requests:
    video: jellyseerr
    audio: audiobookrequest
```

`v1` routing is only by request capability. If exactly one plugin exposes a given request capability, Mortar may route to it automatically. If more than one plugin exposes the same request capability, explicit routing config is required and startup validation must fail if the ambiguity is unresolved.

If no plugin exposes a request capability, that request type is simply unavailable in the product.

## Data model

Mortar normalizes all data from backend services into shared types:

- `MediaItem` — title, year, type (movie/show/audiobook/ebook), poster, external IDs
- `Request` — item, requester, status, submitted_at, fulfilled_at
- `ActivityEvent` — source plugin, event type, item, timestamp, message
- `DownloadItem` — name, progress, size, speed, eta, source plugin
- `HealthStatus` — status, reachable, latency_ms, checked_at, detail? (Mortar adds plugin id when aggregating health across plugins; status is derived by Mortar from reachable + latency_ms thresholds — see plugin interface)

## Local persistence

Mortar uses a single SQLite database for Mortar-owned state and durable snapshots.

Durable local state includes:

- `users`
- `sessions`
- `external_account_links`
- `request_snapshots`
- `health_snapshots`

`request_snapshots` are a durable convenience index for request history and duplicate prevention, but upstream request systems remain authoritative.

`health_snapshots` store last-known state only per plugin in `v1`; long-term health history is out of scope.

Short-lived cache state, polling cursors, and similar optimization data should stay in memory unless persistence proves operationally necessary for correctness or startup continuity.

## Upstream identity linking

Mortar owns authentication and roles, but some plugin features require user-specific upstream context.

- External account links are explicit per-user, per-plugin records stored in Mortar's local database
- They are optional overall and are only required for features that need personalized upstream behavior
- In `v1`, link management starts with config-file or bootstrap seeding rather than a dedicated admin UI
- Mortar does not auto-provision guest, shadow, or fallback upstream accounts in `v1`

For `v1`, Browse & Play is link-gated for the relevant plugin. If a required external account link is missing, Mortar should fail explicitly with clear UX rather than silently browsing or playing through a shared fallback.

## Read caching

Mortar uses an in-memory read-through cache for volatile upstream reads in `v1`.

- Only idempotent read operations are cacheable
- SQLite is not the hot cache for short-lived plugin responses
- Upstream mutations triggered by Mortar must invalidate affected cached reads immediately
- Cache keys must include plugin identity, read operation, normalized parameters, and user or role context when the result is scoped
- Expired cache entries are refreshed before responding; `v1` does not use stale-while-revalidate behavior

Exact TTL values are intentionally left open for now, but feature specs should still be honest about freshness expectations and latency goals.

## Refresh model

Mortar uses polling for `v1` freshness-sensitive UI updates behind a shared frontend subscription abstraction.

- Downloads poll every 10 seconds by default
- Activity polls every 30 seconds by default
- Health uses a 60-second default freshness interval
- Incremental fetches should be used where they fit naturally, especially for activity-style feeds
- Inactive or backgrounded views should throttle or suspend polling rather than continue at full rate

SSE and WebSockets remain possible later, but they are intentionally deferred until polling proves insufficient.

## Deployment

The Mortar server is distributed as a single Docker image. Configuration is provided via a mounted config file and environment variables for secrets.

```
docker run \
  -v ./config.yaml:/data/config.yaml:ro \
  -v mortar-data:/data \
  -e JELLYSEERR_API_KEY=... \
  -p 3000:3000 \
  ghcr.io/nbellowe/mortar:latest
```

The container mounts all persistent data (config and SQLite database) under `/data`. Config is bind-mounted read-only; the database lives in a named volume at `/data/mortar.db`.

Frontend delivery is split by platform:

- Web: supported `v1.0` client path via the Expo web target
- iOS / Android / desktop: experimental during `v0.x`; not required for `v1.0` release readiness

A Kubernetes manifest and Helm chart are planned for the server deployment.

## Tech stack

**Backend: Go.** The Mortar server is a long-running API proxy with a plugin system making concurrent upstream calls. Go's goroutines, interface model, and single-binary compilation make it the right fit. The plugin interface is a Go interface; each plugin is a Go package under `internal/plugins/`.

**Frontend: Expo (React Native, TypeScript).** Expo is the long-term client strategy because it reaches all target platforms from one codebase and can expand into native app stores later. Mortar's `v1.0` support promise is still web-first. Platform-specific UI code is allowed via `.ios.tsx` / `.android.tsx` / `.web.tsx` file conventions; business logic is always shared.

## Design workflow

**Design tool: Google Stitch.** Stitch is the preferred tool for frontend ideation, screen exploration, flow prototyping, and visual iteration.

**Source of truth: repo-owned design contract.** Stitch is not the implementation authority by itself. The canonical design-system source of truth lives in the repository as versioned design rules, implementation tokens, and component conventions. See [`DESIGN.md`](../DESIGN.md).

See `docs/adrs/0001-tech-stack.md` and `docs/sessions/2026-05-07-tech-stack.md` for the full rationale including alternatives considered.

