# Roadmap

## Current phase

Mortar is in **`v0.1.0-alpha` execution**. The goal right now is not polish; it is proving that the supported stack works end to end with the simplest architecture we can keep.

## Supported path

- **Primary supported client:** web
- **Primary supported backend stack:** Jellyfin, Jellyseerr, Sonarr, Radarr, SABnzbd
- **Committed product flows:** Search & Request, Activity Feed, Download Queue, Browse & Play, Service Health

## Status summary

- The server and web client both run locally.
- The supported plugin stack is wired into the app.
- The alpha-critical product slice is implemented in code on the `codex/p0-alpha-critical` branch.
- Automated backend and frontend tests cover the main new API and client behavior.
- Manual smoke testing has been run against the local Docker stack.
- Distribution, broader end-to-end testing depth, and early marketing are still open.

## P0: Finish Alpha-Critical Product Work

This branch executes the current P0.

- [x] Replace the placeholder library screen with real Browse & Play behavior.
- [x] Add simple session auth with seeded users and role-aware navigation.
- [x] Restrict health details to admins while keeping a public liveness check.
- [x] Cache plugin health snapshots instead of probing every plugin on each request.
- [x] Expand Search & Request so it searches across request-capable plugins and shows useful request state.
- [x] Persist enough local request state to prevent duplicate pending requests.
- [x] Gate Jellyfin-linked experiences when the Mortar user lacks the required external account link.
- [x] Add automated tests for the new auth, library, and request behavior.
- [x] Run manual smoke tests against the local Docker stack.
- [x] Capture the current product state in screenshots.

## Next Up

These are the next items to keep current after P0 merges.

### P1: Activity feed quality

- [ ] Sonarr and Radarr activity events display raw release filenames (e.g. `FireFly.S01E13.1080p.BluRay.x264`) instead of show/movie titles. The `item.title` field is now correctly populated — the fix is to build a human-readable `Message` in the Sonarr/Radarr plugins from `series.Title` + episode info / `movie.Title` rather than passing through `sourceTitle`.
- [ ] Radarr `mapEventType` has a catch-all fallback that maps unknown event types (e.g. `movieAdded`, `health`) to `ActivityEventDownloaded` instead of skipping them. Spurious "Downloaded" events appear for non-download activity. Fix: return `("", false)` for unrecognised types, matching Sonarr's behavior.

### P1: Prove the real end-to-end experience

- [ ] Write and keep a release smoke-test checklist for all five committed flows.
- [ ] Do a focused Browse & Play pass on desktop and mobile browsers, especially the Jellyfin handoff.
- [ ] Decide whether the current "open Jellyfin to play" behavior is good enough for alpha or needs a different plan.
- [ ] Add at least one deeper integration test path that exercises auth plus a real plugin-backed flow.
- [ ] Add decent telemetry to make it easy to track how well its working.

### P1: Distribution

- [ ] Decide the first distribution targets: web-only, desktop wrappers, or app-store builds.
- [ ] Document the minimal release path for outside testers.
- [ ] Prepare app store/distribution prerequisites if native packaging stays in scope.

### P1: Early adoption

- [ ] Write a short operator install guide for the supported stack.
- [ ] Create a simple "what Mortar is" landing/demo description for sharing.
- [ ] Do a mild first wave of outreach to a few friendly testers.

## Exit criteria

### `v0.1.0-alpha`

Goal: prove the architecture with a real, rough, end-to-end Mortar system.

- The server runs locally.
- The web client runs locally.
- A Docker-based server install path exists.
- The supported reference stack can be configured without code changes.
- All five committed flows are demonstrable end to end on the supported path.
- The rough edges are understood well enough to guide the next product decisions.

### `v0.5.0-beta`

- Setup and upgrade steps are documented for operators.
- Smoke-test steps are documented and repeatable.
- CI is required and trusted.
- The supported web experience is dependable enough for outside testers.

### `v1.0.0`

Goal: ship a stable, supported, web-first baseline for the reference stack.

- The web experience is the clearly supported primary surface.
- The reference stack is stable and compatibility-defined.
- All five committed flows are stable on the supported reference stack.
- Release, upgrade, and security policies are documented.
- No release-critical roadmap or spec blockers remain.

## Notes

- This roadmap is scope-based, not date-based.
- Native clients are still part of the long-term architecture direction, but they are not `v1.0` release-blocking.
- Additional plugins may land before `v1.0`, but only the reference stack defines `v1.0` readiness.
