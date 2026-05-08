# Specs

This directory is the source of truth for what Mortar builds. All features begin as specs. Agents implement against specs. If a spec and the implementation disagree, update the spec first, then the code.

## Index

### Foundation

- [Vision](vision.md) — problem statement, audiences, core flows, non-goals
- [Architecture](architecture.md) — plugin model, data flow, auth, deployment

### Plugin interface

- [Plugin Interface](plugins/plugin-interface.md) — capability flags, TypeScript interfaces, shared types

### Features

- [Search & Request](features/requests.md) — unified search and request flow
- [Activity Feed](features/activity-feed.md) — cross-service event timeline
- [Download Queue](features/download-queue.md) — unified download progress view
- [Home](features/home.md) — dashboard rows and health badge
- [Browse & Play](features/browse-play.md) — library browsing and playback handoff
- [Service Health](features/health.md) — plugin connectivity dashboard

## Conventions

- Each spec should begin with a short metadata block: **Status** and **Depends on** for ADRs or other blocking specs. **Last updated** is optional if it helps you track iteration.
- Each spec has an **Acceptance criteria** section. These are the definition of done.
- Feature and architecture specs should include a short **Documentation impact** section whenever the change affects any of Mortar's three documentation audiences:
  - **App users** — how to use the feature
  - **Operators** — how to configure, deploy, or troubleshoot it
  - **Contributors** — how the change affects contributor workflow, conventions, or project policy
- **Open questions** sections track decisions not yet made. Resolve them before implementing.
- Specs are written in terms of user behavior, not implementation. Keep the how out of specs.
- Cross-cutting technical decisions belong in `docs/adrs/`. Feature and architecture specs should reference ADR ids instead of restating the same decision logic in multiple places.
- When a requirement is intended to be normative, prefer explicit language such as `MUST`, `SHOULD`, and `MAY`.
