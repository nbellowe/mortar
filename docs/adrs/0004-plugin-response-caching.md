# ADR 0004: Plugin Response Caching

## Status

Accepted

## Date

2026-05-07

## Context

Mortar fans out to multiple upstream services for searches, activity, downloads, health, and library views. Without a caching policy, two failure modes appear quickly:

- Upstream services get hammered by repeated identical requests
- Feature specs cannot make honest latency claims

The current repo mentions caching as an open question but does not define scope, invalidation, or where cached data lives.

## Decision

Use an in-memory read-through cache for volatile upstream reads. Use SQLite only for durable snapshots and Mortar-owned state, not as the primary cache for short-lived plugin responses.

## Policy

- Cache only idempotent read operations.
- Mutations must invalidate affected cached entries immediately.
- Do not use request coalescing in `v1`; simultaneous cache misses may fan out independently.
- Do not use stale-while-revalidate behavior in `v1`; expired entries must be refreshed before responding.
- Cache keys must include plugin identity, read operation, and normalized query parameters.
- If a read result depends on effective user or role context, that context must be part of the cache key.
- Cache behavior uses internal defaults in `v1`; no operator-facing cache tuning is required yet.

## TTL posture

`v1` intentionally does not standardize exact TTL values yet. Freshness expectations still matter, but the repo is not locking concrete cache timings until implementation and feature rewrites make them easier to validate.

## Rationale

- Memory caching matches the single-server deployment model and keeps implementation simple.
- SQLite is a better fit for durable state than for a fast-changing hot cache.
- A simple cache model is easier to reason about than mixing coalescing, stale reads, and operator tuning before the first implementation exists.

## Consequences

- Specs should call out freshness expectations where they matter.
- The cache key design must include plugin id, read operation, query parameters, and user/role context where relevant.
- Any admin mutation flow must invalidate related request and activity cache entries.
- Activity, browse, download, health, and request views may rely on short-lived in-memory read caching, but not on durable cached responses in SQLite.

## Related

- [Plugin Response Caching Decision Session](../sessions/2026-05-07-plugin-response-caching.md)
- [Architecture Spec](../../specs/architecture.md)
