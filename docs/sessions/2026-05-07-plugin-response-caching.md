# Plugin Response Caching Decision Session — 2026-05-07

## Why this session happened

Mortar needs enough caching to make latency and upstream load manageable, but not so much complexity that the first implementation becomes hard to reason about.

The main goal was to choose a cache model that fits the single-server, web-first, pre-1.0 shape of the project.

## Decisions made

### Core cache model

- `v1` uses an in-memory read-through cache for plugin read responses
- SQLite is reserved for durable snapshots and Mortar-owned state, not the primary hot cache
- Only idempotent read operations are cacheable

### Mutation behavior

- Upstream mutations triggered by Mortar must invalidate affected cached reads immediately
- `v1` does not attempt proactive write-through refresh after mutations

### Simplicity choices

- No request coalescing in `v1`
- No stale-while-revalidate behavior in `v1`
- No operator-facing cache tuning in `v1`
- Exact TTL values are intentionally deferred for now

### Cache key safety

- Cache keys must include plugin identity
- Cache keys must include the read operation being performed
- Cache keys must include normalized query parameters
- User-scoped or role-scoped reads must include effective user context in the cache key

## Why this direction won

- It matches the single-server self-hosted deployment model
- It keeps durable state and volatile cache behavior clearly separated
- It avoids premature complexity while still giving the project a real cache policy
- It preserves correctness for personalized and visibility-scoped reads

## Immediate follow-up chosen from this session

- Mark ADR 0004 as accepted
- Update the architecture spec to summarize the accepted cache model
- Remove the old SQLite-cache ambiguity from the affected feature specs
