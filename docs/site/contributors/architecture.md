# Architecture

## Overview

Mortar is a three-layer system:

```
Clients (web browser, experimental native)
        ↕ HTTP / JSON
Mortar Server (Go)
        ↕ HTTP APIs
Upstream services (Jellyfin, Jellyseerr, Sonarr, …)
```

The server is the only component that talks to upstream services. The frontend talks exclusively to the Mortar server.

## Plugin model

All upstream service integrations are plugins. A plugin:

1. Declares a set of **capability flags** at registration time
2. Implements only the interfaces matching those flags
3. Is called by the server only for operations its capabilities cover

This means adding a new service never requires touching existing code — you write a new plugin package.

Capability flags:

| Flag | Interface |
|---|---|
| `requests.video` / `.audio` / `.ebook` | Submit and look up media requests |
| `library.browse` | List recently added items |
| `library.exists` | Check whether a title is in the library |
| `library.resume` | Fetch in-progress playback positions |
| `downloads.read` | Read the active download queue |
| `activity.read` | Read recent service activity |

Every plugin also implements `health()`, which Mortar polls to populate the health dashboard.

## Data model

Shared types flow through all layers unchanged. Do not invent variants:

| Type | Used for |
|---|---|
| `MediaItem` | Any piece of media: movie, show, episode, audiobook |
| `Request` | A pending or fulfilled media request |
| `ActivityEvent` | A timestamped event from a service (download completed, etc.) |
| `DownloadItem` | An active download with progress |
| `HealthStatus` | A plugin's current reachability and status |

## Persistence

SQLite stores durable state:

- User accounts and sessions
- External account links (Mortar user ↔ Jellyfin user, etc.)
- Last-known health snapshot per plugin
- Request history (upstream systems remain authoritative)

Volatile upstream data (library contents, download queues, activity) is cached in memory only.

## Polling intervals

Mortar polls upstream services on fixed intervals:

| Data | Default interval |
|---|---|
| Download queue | 10 seconds |
| Activity feed | 30 seconds |
| Service health | 60 seconds |

Polling suspends when no clients are active.

## Auth

Mortar owns authentication. Users log in with a username and password; the server issues a session cookie. API keys for upstream services are stored in the config file (via environment variable references) and never exposed to clients.

## Further reading

- `specs/architecture.md` — full architecture spec
- `docs/adrs/` — cross-cutting decisions with rationale
