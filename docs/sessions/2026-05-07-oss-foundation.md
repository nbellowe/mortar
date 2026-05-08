# OSS Foundation Planning Session — 2026-05-07

## Why this session happened

The repo already had strong product and architecture direction, but it was still missing some of the repo-level decisions that make an open source project feel deliberate: release scope, roadmap shape, plugin support posture, versioning, and baseline project docs.

## Decisions made

### `v1.0` scope

- `v1.0` is **web-first**
- The officially supported `v1.0` surfaces are the Mortar server and web client
- Native clients for iOS, Android, macOS, and Windows may exist during `v0.x`, but they are **experimental** and not release-blocking for `v1.0`

### `v1.0` feature and plugin commitment

- `v1.0` still includes all five core flows already named in `specs/vision.md`
- The supported `v1.0` reference stack is:
  - Jellyfin
  - Jellyseerr
  - Sonarr
  - Radarr
  - SABnzbd

### Plugin support model

Two support levels are enough for now:

- **Supported** — documented, compatibility-defined, tested, part of the reference stack, and release-blocking if broken
- **Experimental** — basic docs, minimal tests, known compatibility target, and not release-blocking

We explicitly chose **not** to add a "community" tier yet. That can be introduced later if the ecosystem needs it.

### Release maturity ladder

- `alpha` — architecture proven with a real but rough end-to-end system
- `beta` — installable and dependable for outside testers
- `v1.0` — stable supported baseline for the web-first reference stack

### Versioning and release posture

- Use SemVer with prereleases: `v0.1.0-alpha.1`, `v0.5.0-beta.1`, `v1.0.0-rc.1`, `v1.0.0`
- `0.x` releases may introduce breaking changes, but the breakage must be called out clearly
- Use RCs only for especially important stable cuts
- Use a `main`-first branching model with short-lived release branches only when needed
- Official releases should include a Git tag, GitHub release, release notes, and a published Docker image

### Repo-level OSS foundation

- License: **Apache-2.0**
- Public roadmap: **`ROADMAP.md` at repo root**
- Minimum contributor-facing baseline later: `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`

### Documentation structure

Mortar documentation should be thought about early, even before implementation exists. We identified three audiences:

- **App users** — people using Mortar
- **Operators** — people running Mortar
- **Contributors** — people changing Mortar itself

From that, the documentation structure should separate:

- `docs/site/users/`
- `docs/site/operators/`
- contributor-facing repo-root docs such as `CONTRIBUTING.md`, `SECURITY.md`, and `CODE_OF_CONDUCT.md`

## What we intentionally deferred

- Detailed hotfix and rollback procedures
- Full support-window and deprecation policy
- A third plugin support tier for community-maintained integrations
- Heavy release-process mechanics that are better decided once the project is runnable

## Immediate follow-up chosen from this session

- Add the Apache-2.0 license
- Add a public roadmap
- Remove the existing contradiction around whether native clients are part of `v1.0`
- Scaffold the documentation structure around app users, operators, and contributors
