# ADR 0003: Real-Time Update Delivery

## Status

Accepted

## Date

2026-05-07

## Context

Mortar needs near-real-time updates for the download queue, activity feed, and health status. The current specs mention polling, SSE, and WebSocket but do not choose one. The product is early, the server is self-hosted, and the initial workloads are modest.

## Decision

Use polling in v1 behind a shared frontend subscription abstraction. Do not commit to SSE or WebSocket until a concrete product need appears that polling cannot satisfy.

## Default intervals

- Downloads: 10 seconds
- Activity: 30 seconds
- Health: 60 seconds

## Policy

- Frontend refresh behavior should be centralized in a shared subscription or refresh abstraction rather than per-screen ad hoc timers.
- Use incremental fetches where the capability naturally supports them, especially for activity-style feeds.
- Do not force delta semantics onto capabilities where full refreshes remain the clearer model.
- Polling should throttle or suspend when the relevant screen is inactive or the app is backgrounded.
- `v1` defines the principle above, but does not lock exact inactive-state pause or throttle timings yet.

## Rationale

- Polling is the simplest transport to implement, observe, and debug across web and native clients.
- It avoids introducing connection lifecycle complexity before the API surface is stable.
- A shared client-side subscription abstraction keeps a later move to SSE or WebSocket possible without rewriting feature screens.
- Throttling or suspending inactive polling reduces unnecessary server and upstream load in the self-hosted deployment model.

## Consequences

- Specs should describe freshness budgets rather than claiming fully live updates.
- The API should support incremental fetches where possible, especially for activity.
- The frontend should centralize refresh policy instead of per-screen ad hoc timers.
- A later move to SSE or WebSocket should be treated as an implementation swap behind the shared subscription abstraction, not a per-screen rewrite.

## Follow-up

- Revisit this ADR after the first implementation if polling causes unacceptable load, latency, or battery impact.

## Related

- [Real-Time Delivery Decision Session](../sessions/2026-05-07-realtime-updates.md)
- [Architecture Spec](../../specs/architecture.md)
