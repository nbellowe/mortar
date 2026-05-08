# ADR 0002: Persistence and State Model

## Status

Accepted

## Date

2026-05-07

## Context

Mortar proxies most user-visible data from upstream services, but it still owns some data outright:

- Mortar user accounts and roles
- Session state
- Per-plugin external account links
- Health snapshots
- Request-history snapshots or cursors used for faster views
- Cache metadata and polling cursors

The current specs say Mortar does not own request state, but they also require user auth, request history views, and health caching. That means Mortar needs durable local state even if upstream systems remain the source of truth for media operations.

## Decision

Use a single SQLite database as Mortar's `v1` durable store for Mortar-owned state and durable snapshots.

Mortar should not create a second source of truth for mutable media state that already belongs to upstream services.

## Durable core state

- `users`
- `sessions`
- `external_account_links`
- `health_snapshots`
- `request_snapshots`

## Snapshot behavior

- `request_snapshots` are a durable history and convenience index retained indefinitely in `v1`
- `request_snapshots` support request history, duplicate-request checks, and faster views
- Upstream request systems remain authoritative for current request state and review actions
- `health_snapshots` store only the last-known health record per plugin
- `health_snapshots` are overwritten on each health check and back the cached health view

## Auth behavior

- Sessions are durable and server-side
- The supported web client uses `HttpOnly` session cookies
- Native-client auth is intentionally deferred until native clients become a supported surface

## Transient or optional state

- In-memory poll state
- Short-lived cache entries
- Cache metadata or polling cursors only when they materially improve correctness or startup continuity
- If cache metadata or cursors need persistence, they should use the same SQLite file rather than a separate store

## What will not be stored as authoritative data

- Full library contents
- Full download queues
- Playback state owned by Jellyfin or another media server
- Request approval logic owned by Jellyseerr or another request system

## Rationale

- SQLite is sufficient for the single-server self-hosted deployment model.
- It keeps setup simple for homelab users and avoids introducing a second service.
- Durable local state solves the auth, identity-linking, request-history, and cached-health requirements without distorting the plugin model.

## Consequences

- Mortar needs a migration system from day one.
- Specs should distinguish between upstream source-of-truth data and Mortar snapshots or convenience indexes.
- API responses that include cached or snapshot-backed data should expose freshness where relevant.
- Web auth design can stay simple and server-centered for `v1`.
- Restart continuity exists for durable auth and snapshot-backed views without persisting every internal optimization detail.

## Related

- [Persistence and State Decision Session](../sessions/2026-05-07-persistence-state.md)
- [Architecture Spec](../../specs/architecture.md)
