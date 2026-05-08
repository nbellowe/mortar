# Real-Time Delivery Decision Session — 2026-05-07

## Why this session happened

Mortar needed a concrete answer for freshness-sensitive views like downloads, activity, and health, but the repo was still leaving the transport open between polling, SSE, and WebSockets.

The goal was to choose a `v1` delivery model that fits a web-first, self-hosted, single-server app without creating premature connection-management complexity.

## Decisions made

### Core delivery model

- `v1` uses polling
- Polling sits behind a shared frontend subscription or refresh abstraction
- SSE and WebSockets are explicitly deferred until polling proves insufficient

### Default intervals

- Downloads: `10s`
- Activity: `30s`
- Health: `60s`

### API and frontend behavior

- Incremental fetches should be used where they fit naturally, especially for activity-style feeds
- Delta semantics should not be forced onto every capability
- Polling should throttle or suspend when the relevant screen is inactive or the app is backgrounded
- `v1` records that principle, but does not lock exact inactive-state timing details yet

## Why this direction won

- It is the simplest option to build, observe, and debug
- It matches the modest early workload and self-hosted deployment story
- It avoids connection-lifecycle complexity before the API surface is stable
- The shared subscription abstraction preserves a clean migration path if a later transport upgrade becomes necessary

## Immediate follow-up chosen from this session

- Mark ADR 0003 as accepted
- Update the architecture spec to summarize the polling model
- Remove the old transport-choice ambiguity from the activity and download feature specs
