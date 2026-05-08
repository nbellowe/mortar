# ADR 0001: Tech Stack

## Status

Accepted

## Date

2026-05-07

## Context

Mortar needs one client codebase that can reach web, iOS, Android, macOS, and Windows in v1, with a credible path to TV platforms post-v1. The server is a long-running API proxy and plugin runtime that fans out to multiple upstream services concurrently.

## Decision

- Backend: Go
- Frontend: Expo (React Native, TypeScript)

## Consequences

- Plugin implementations live in Go packages under `internal/plugins/`.
- The frontend consumes a Mortar API instead of talking to upstream services directly.
- Platform-specific UI is allowed, but business logic and types remain shared.

## Related

- [Tech Stack Decision Session](../sessions/2026-05-07-tech-stack.md)
- [Architecture Spec](../../specs/architecture.md)
