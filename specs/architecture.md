# Architecture

## Overview

Mortar has three layers:

```
Browser
  └── Mortar UI (web app)
        └── Mortar Server (API + plugin runtime)
              ├── Plugin: Jellyseerr
              ├── Plugin: Jellyfin
              ├── Plugin: SABnzbd
              ├── Plugin: qBittorrent
              ├── Plugin: Sonarr
              ├── Plugin: Radarr
              └── Plugin: AudioBookRequest
```

The server holds API keys, proxies requests to backend services, normalizes responses, and serves the UI. The browser never talks directly to backend services.

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
| `library.exists` | Can answer "does this item exist in the library?" |
| `downloads.read` | Can report current download queue and progress |
| `activity.read` | Can report a stream of recent events |
| `health.ping` | Can report connectivity status |

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
    capabilities: [activity.read]  # override: use sonarr only for activity

  - id: audiobookrequest
    type: audiobookrequest
    url: http://audiobookrequest:8080
    api_key: ${ABR_API_KEY}
```

Multiple plugins of the same type are allowed (e.g. two Sonarr instances).

## Auth model

Two separate auth concerns:

**Mortar user auth** — users log in to Mortar with a username/password (or SSO in a future version). Mortar maintains its own user table with roles: `admin` and `user`.

**Service API keys** — stored in the Mortar server config, never exposed to the browser. The server proxies all service API calls. Users interact only with Mortar's API.

## Request routing

When a user submits a request, Mortar routes it to the appropriate plugin based on content type:

```
video request   → first plugin with requests.video capability
audio request   → first plugin with requests.audio capability
ebook request   → first plugin with requests.ebook capability
```

In configs where multiple plugins share a capability (unlikely but possible), the first declared plugin wins. A future version may allow explicit routing rules.

## Data model

Mortar normalizes all data from backend services into shared types:

- `MediaItem` — title, year, type (movie/show/audiobook/ebook), poster, external IDs
- `Request` — item, requester, status, submitted_at, fulfilled_at
- `ActivityEvent` — source plugin, event type, item, timestamp, message
- `DownloadItem` — name, progress, size, speed, eta, source plugin
- `HealthStatus` — plugin id, reachable, latency_ms, checked_at

## Deployment

Mortar is distributed as a single Docker image. Configuration is provided via a mounted config file and environment variables for secrets.

```
docker run \
  -v ./mortar.yaml:/config/mortar.yaml \
  -e JELLYSEERR_API_KEY=... \
  -p 3000:3000 \
  mortar:latest
```

A Kubernetes manifest and Helm chart are planned.

## Open questions

- **Tech stack:** TypeScript (Next.js full-stack) vs. Go backend + React frontend. Decision pending.
- **Database:** Does Mortar need persistent storage? Possibly for user accounts and request history cache. SQLite is sufficient.
- **Real-time updates:** WebSocket or SSE for live download progress and activity feed?
- **Caching:** How aggressively should plugin responses be cached to avoid hammering upstream services?
