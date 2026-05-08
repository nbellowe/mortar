# ADR 0002: Persistence and State Model

## Status

Proposed

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

Use SQLite as Mortar's local persistent store for Mortar-owned state and durable snapshots. Mortar should not create a second source of truth for mutable media state that already belongs to upstream services.

## What will be stored

- `users`
- `sessions`
- `external_account_links`
- `health_snapshots`
- `request_snapshots`
- `activity_cursors`
- `cache_metadata`

## What will not be stored as authoritative data

- Full library contents
- Full download queues
- Playback state owned by Jellyfin or another media server
- Request approval logic owned by Jellyseerr or another request system

## Rationale

- SQLite is sufficient for the single-server self-hosted deployment model.
- It keeps setup simple for homelab users and avoids introducing a second service.
- Durable local state solves the identity-linking and cached-health requirements without distorting the plugin model.

## Consequences

- Mortar needs a migration system from day one.
- Specs should distinguish between upstream source-of-truth data and Mortar snapshots.
- API responses that include cached or snapshot-backed data should expose freshness where relevant.

## Open questions

- Should request snapshots be retained indefinitely or bounded by age?
- Should cache metadata and health snapshots live in the same SQLite file or be split later if scale changes?
