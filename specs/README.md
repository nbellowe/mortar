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
- [Browse & Play](features/browse-play.md) — library browsing and playback handoff
- [Service Health](features/health.md) — plugin connectivity dashboard

## Conventions

- Each spec has an **Acceptance criteria** section. These are the definition of done.
- **Open questions** sections track decisions not yet made. Resolve them before implementing.
- Specs are written in terms of user behavior, not implementation. Keep the how out of specs.
