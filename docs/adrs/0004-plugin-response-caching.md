# ADR 0004: Plugin Response Caching

## Status

Proposed

## Date

2026-05-07

## Context

Mortar fans out to multiple upstream services for searches, activity, downloads, health, and library views. Without a caching policy, two failure modes appear quickly:

- Upstream services get hammered by repeated identical requests
- Feature specs cannot make honest latency claims

The current repo mentions caching as an open question but does not define scope, invalidation, or where cached data lives.

## Decision

Use an in-memory read-through cache with request coalescing for volatile upstream reads. Use SQLite only for durable snapshots and metadata, not as the primary cache for short-lived plugin responses.

## Policy

- Cache only idempotent read operations.
- Mutations must invalidate affected cached entries immediately.
- Use request coalescing for identical in-flight requests.
- Prefer stale-while-revalidate behavior for non-critical list views.

## Recommended TTLs

- Health: 60 seconds
- Download queue: 5-10 seconds
- Activity: 15-30 seconds
- Library browse pages: 30-120 seconds
- Search results: 30-60 seconds
- Request detail / request lists: 15-30 seconds

## Rationale

- Memory caching matches the single-server deployment model and keeps implementation simple.
- SQLite is a better fit for durable state than for a fast-changing hot cache.
- Capability-specific TTLs avoid pretending that all plugin data has the same freshness requirements.

## Consequences

- Specs should call out freshness expectations where they matter.
- The cache key design must include plugin id, capability, and query parameters.
- Any admin mutation flow must invalidate related request and activity cache entries.
