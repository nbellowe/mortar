# ADR 0003: Real-Time Update Delivery

## Status

Proposed

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

## Rationale

- Polling is the simplest transport to implement, observe, and debug across web and native clients.
- It avoids introducing connection lifecycle complexity before the API surface is stable.
- A shared client-side subscription abstraction keeps a later move to SSE or WebSocket possible without rewriting feature screens.

## Consequences

- Specs should describe freshness budgets rather than claiming fully live updates.
- The API should support incremental fetches where possible, especially for activity.
- The frontend should centralize refresh policy instead of per-screen ad hoc timers.

## Follow-up

- Revisit this ADR after the first implementation if polling causes unacceptable load, latency, or battery impact.
