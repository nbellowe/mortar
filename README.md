# Mortar

A unified frontend for homelab media stacks. Mortar is the front door — a single, consistent UI for household members to search, request, browse, and track media across a heterogeneous set of self-hosted services.

Mortar does not replace any existing service. It connects to them via their APIs and presents a unified experience on top.

## Status

Early spec phase. Product direction and release scope are defined, but implementation has not started yet. See [`ROADMAP.md`](ROADMAP.md) for milestone goals and [`specs/`](specs/) for design documents.

## Core idea

Every homelab media stack runs a different set of services — Jellyfin, Jellyseerr, Sonarr, Radarr, SABnzbd, Audiobookshelf, and more. Each has its own UI, its own auth, and its own design language. Mortar aggregates the daily-use workflows across all of them into one place, without replacing any of them.

## Design principles

- **Plugin-first.** Every integration is a plugin with a standard interface. No service is hardcoded.
- **Front door, not replacement.** Complex operations belong in the native app. Mortar owns the 80% daily-use flows.
- **Built for the household, not just the owner.** Non-technical users should be able to use it without knowing what Sonarr is.
- **Homelab-owner configured.** The owner defines which plugins are enabled. Users just use it.
- **Spec-driven.** All features begin as specs. Agents implement against specs.

## Release scope

- **Supported for `v1.0`:** Mortar server + web client
- **Experimental during `v0.x`:** iOS, Android, macOS, and Windows clients
- **Supported `v1.0` reference stack:** Jellyfin, Jellyseerr, Sonarr, Radarr, SABnzbd

## Docs

- [`ROADMAP.md`](ROADMAP.md) — Milestones, supported scope, and release targets
- [DESIGN.md](DESIGN.md) — Frontend design contract and Stitch handoff rules
- [`specs/`](specs/) — Feature and architecture specs
- [`docs/adrs/`](docs/adrs/) — Architecture decision records
- [`docs/sessions/`](docs/sessions/) — Historical brainstorm and decision logs

## License

Apache-2.0. See [`LICENSE`](LICENSE).
