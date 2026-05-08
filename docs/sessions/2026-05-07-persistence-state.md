# Persistence and State Decision Session — 2026-05-07

## Why this session happened

Mortar already had a strong plugin-first direction, but several blocked specs depended on a concrete answer to a simple question: how much state does Mortar own locally, and how much should remain upstream?

The risk was not choosing the wrong database. The risk was accidentally inventing a second source of truth during implementation.

## Decisions made

### Storage model

- `v1` uses a single SQLite database
- SQLite is the durable store for Mortar-owned state and durable snapshots
- Mortar remains a proxy and normalization layer, not the authority for media-domain state owned by upstream services

### Durable core state

- `users`
- `sessions`
- `external_account_links`
- `request_snapshots`
- `health_snapshots`

### Request snapshots

- Request snapshots are durable in `v1`
- They are retained indefinitely in `v1`
- They support request history, duplicate-request checks, and faster views
- Upstream request systems remain authoritative for current request state and review actions

### Health snapshots

- Health snapshots store only the last-known record per plugin
- They are overwritten on each health check
- The health UI should read cached state rather than blocking on a fresh live probe
- Long-term health history is intentionally out of scope for `v1`

### Auth and sessions

- Sessions are server-side and durable
- The supported web client uses `HttpOnly` session cookies
- Native-client auth is intentionally deferred until native clients become a supported surface

### Transient and optional persistence

- Short-lived cache state and poll bookkeeping should stay in memory by default
- Cache metadata and cursors should only be persisted when they materially improve correctness or startup continuity
- If they are persisted, they belong in the same SQLite database rather than a separate store

## Why this direction won

- It matches the web-first, single-server self-hosted deployment story
- It keeps setup simple for operators
- It gives Mortar enough durable state to feel reliable without turning it into a replacement for upstream apps
- It preserves the "front door, not replacement" principle in the data model itself

## Immediate follow-up chosen from this session

- Mark ADR 0002 as accepted
- Update the architecture spec to summarize the accepted persistence model
- Adjust the requests and health feature specs to reflect durable snapshots and cached last-known health state
