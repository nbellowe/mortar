# Roadmap

## Current phase

Mortar is currently **pre-alpha**. The project has a clear product direction, but it still needs a few architecture and spec decisions closed before implementation can move quickly without inventing behavior in code.

## Scope snapshot

- **Supported `v1.0` surfaces:** Mortar server + web client
- **Experimental during `v0.x`:** iOS, Android, macOS, Windows
- **Supported `v1.0` reference stack:** Jellyfin, Jellyseerr, Sonarr, Radarr, SABnzbd
- **Committed `v1.0` flows:** Search & Request, Activity Feed, Download Queue, Browse & Play, Service Health

## Plugin support levels

- **Supported:** documented, compatibility-defined, tested, part of the reference stack, and release-blocking if broken
- **Experimental:** basic docs, minimal tests, known compatibility target, and not release-blocking

## Milestones

### Pre-alpha: implementation readiness

This is the current phase.

Exit criteria:

- ADRs 0002-0006 are accepted or explicitly revised
- The plugin interface is stable enough for initial scaffolding
- Feature specs no longer rely on unresolved release-critical questions
- The missing home-screen spec is added or its behavior is folded into existing accepted specs

### `v0.1.0-alpha`

Goal: prove the architecture with a real, rough, end-to-end Mortar system.

Exit criteria:

- The server runs locally
- The web client runs locally
- A Docker-based server install path exists
- The supported reference stack can be configured
- All five committed flows are demonstrable end-to-end on the supported stack
- Rough edges are acceptable, but the system is real and usable enough to evaluate

### `v0.5.0-beta`

Goal: make Mortar installable and dependable for outside testers.

Exit criteria:

- A documented install path exists for the supported stack
- Local development and Docker deployment are reproducible
- All five committed flows are implemented on the supported stack
- CI exists and is required for the project
- Release smoke-test steps exist for the supported path
- A compatibility matrix exists for supported plugins and upstream versions
- Major UX gaps in the supported web experience are closed

### `v1.0.0`

Goal: ship a stable, supported, web-first baseline for the reference stack.

Exit criteria:

- The Mortar server and web client are the officially supported `v1.0` surfaces
- All five committed flows are stable on the supported reference stack
- Release process, versioning, and changelog policy are documented
- Upgrade expectations are documented
- Security reporting instructions are documented
- No release-critical ADRs or blocked specs remain

## Notes

- This roadmap is scope-based, not date-based
- Native clients are still part of the long-term architecture direction, but they are not `v1.0` release-blocking
- Additional plugins may land before `v1.0`, but only the reference stack defines `v1.0` readiness
