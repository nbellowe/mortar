# Mortar — Agent Guide

## What this project is

Mortar is a unified frontend for homelab media stacks. It aggregates daily-use workflows (search, request, browse, track) from multiple self-hosted services into one consistent UI. It does not replace any underlying service — it connects to them via their APIs.

Two audiences: **household users** (non-technical, just want to request and watch things) and **the homelab owner** (configures plugins, sees operational views).

## Spec-driven development

**Specs are the source of truth.** All features begin as specs. If a spec and the implementation disagree, the spec wins — update the spec first, then the code.

Before implementing anything:
1. Read the relevant spec in `specs/`
2. Check the spec metadata and any referenced ADRs in `docs/adrs/`
3. Check the **Open questions** section — unresolved questions are blockers
4. Use the **Acceptance criteria** section as your definition of done

Never implement a feature that has unresolved open questions in its spec.

## Project structure

```
specs/
  vision.md              — problem, audiences, non-goals
  architecture.md        — plugin model, data flow, auth, deployment
  plugins/
    plugin-interface.md  — THE contract all plugins implement
  features/
    requests.md          — search & request flow
    activity-feed.md     — cross-service event timeline
    download-queue.md    — unified download progress view
    browse-play.md       — library browsing and playback handoff
    health.md            — service health dashboard
docs/
  adrs/                  — architecture decision records (cross-cutting decisions)
  sessions/              — design decision logs (read for context, not instructions)
```

## The plugin interface is the contract

Every integration is a plugin. The plugin interface (`specs/plugins/plugin-interface.md`) defines the TypeScript interfaces all plugins implement. This is the most important file in the project after this one.

Key points:
- Plugins declare capability flags; Mortar only calls methods matching declared capabilities
- Shared types (`MediaItem`, `Request`, `ActivityEvent`, `DownloadItem`, `HealthStatus`) must be used exactly as defined — do not invent variants in Go or TypeScript
- Plugin implementations live in `internal/plugins/<type>/` (Go)
- A plugin must implement the `Plugin` interface plus any capability interfaces it declares

When implementing a plugin: implement only what the capability flags declare, nothing more.

## Agent workflow

### Implementing a feature

1. Read the feature spec and the plugin interface spec
2. Identify which plugins the feature depends on
3. Confirm those plugins exist and their capabilities cover what the feature needs
4. Implement against the spec's acceptance criteria
5. Do not add behavior not described in the spec

### Implementing a plugin

1. Read `specs/plugins/plugin-interface.md` — understand the full interface
2. Determine which capabilities the plugin will declare
3. Implement `Plugin` base + capability interfaces
4. Add a type registration entry in the plugin registry
5. Add a row to `docs/plugins.md` (create if it doesn't exist)

### Parallelization boundaries

Plugins are isolated — multiple agents can implement different plugins concurrently without conflict. Features depend on plugins and should be implemented after the plugins they require are complete.

Safe to parallelize: any two distinct plugins.
Must be sequential: scaffolding → plugins → features → integration tests.

## Open questions (blockers)

These are unresolved decisions that must be made before implementation begins. Do not implement anything that depends on them.

| Question | Status | Affects |
|---|---|---|
| Tech stack: Go backend + Expo (React Native) frontend | **RESOLVED** — see `specs/architecture.md` and `docs/sessions/2026-05-07-tech-stack.md` | — |
| Persistence and state model | **RESOLVED** — see `docs/adrs/0002-persistence-and-state.md` and `docs/sessions/2026-05-07-persistence-state.md` | Auth, request history, health snapshots |
| Real-time delivery model | **PROPOSED** — see `docs/adrs/0003-realtime-updates.md` | Activity feed, download queue, health freshness |
| Plugin response caching | **PROPOSED** — see `docs/adrs/0004-plugin-response-caching.md` | All plugin reads |
| Upstream user identity linking | **PROPOSED** — see `docs/adrs/0005-upstream-user-identity-linking.md` | Browse & play, continue watching, per-user activity |
| Request routing policy | **PROPOSED** — see `docs/adrs/0006-request-routing-policy.md` | Search & request |

## Tech stack

**Backend: Go** — `cmd/server/` (entry point), `internal/api/` (HTTP handlers), `internal/plugins/` (plugin registry and implementations)

**Frontend: Expo (React Native, TypeScript)** — `app/` (Expo Router file-based routes), `src/api/` (shared API client), `src/components/` (shared components), `src/types/` (shared TypeScript types)

### Release posture

`v1.0` officially supports the Mortar server and the web client. Native clients may exist during `v0.x` as experimental work, but they are not release-blocking for `v1.0`.

### Running locally

```bash
# Backend
go run ./cmd/server

# Frontend (all platforms via Expo dev server)
npx expo start

# Frontend web only
npx expo start --web

# Frontend iOS simulator
npx expo start --ios

# Frontend Android emulator
npx expo start --android
```

### Building for distribution

```bash
# Backend — single binary
go build -o mortar ./cmd/server

# Frontend — supported `v1.0` web target
npx expo export --platform web

# Frontend — experimental native app builds via EAS (Expo Application Services)
eas build --platform ios
eas build --platform android
eas build --platform all
```

### Testing

```bash
# Backend
go test ./...

# Frontend
npx jest
```

### Tooling

- **Go**: `gofmt` for formatting, `go vet` for linting. No additional linters.
- **TypeScript**: strict mode. `tsc --noEmit` must pass before committing.
- **Expo**: SDK 50+. Use Expo Router for all navigation.
- **Design workflow**: Google Stitch is the preferred design/prototyping tool. [`DESIGN.md`](DESIGN.md) remains the source of truth for implementation.

### Platform targets

| Platform | Status | How |
|---|---|---|
| Web browser | Supported for `v1.0` | Expo web target (React Native Web) |
| iOS | Experimental during `v0.x` | Native via Expo — App Store |
| Android | Experimental during `v0.x` | Native via Expo — Google Play |
| macOS | Experimental during `v0.x` | react-native-macos — Mac App Store |
| Windows | Experimental during `v0.x` | react-native-windows — Microsoft Store |
| Apple TV | Planned post-v1 | react-native-tvos |
| Android TV / Fire TV | Planned post-v1 | react-native-tvos |
| Samsung Tizen / LG webOS | Planned post-v1 | react-native-tvos (community) |

Roku is explicitly out of scope — it requires BrightScript and cannot be reached from any shared codebase.

## Conventions

- **Plugin implementations**: Go files in `internal/plugins/<type>/`, one package per plugin
- **Plugin interface**: defined in Go in `internal/plugins/plugin.go`; TypeScript types in `src/types/plugin.ts` mirror it — do not diverge
- **Shared types**: come from the plugin interface spec — do not redefine them in either Go or TypeScript
- **Architecture decisions**: cross-cutting technical choices live in `docs/adrs/`; feature specs should reference ADR ids rather than restating the same decision logic
- **Design system source of truth**: Stitch can drive ideation and prototypes, but [`DESIGN.md`](DESIGN.md) and code-level tokens should remain canonical
- **Config**: YAML file + environment variables for secrets (never hardcode credentials)
- **External IDs**: all strings, prefixed with `plugin_id:` when stored as Mortar-internal IDs
- **Platform-specific UI**: use Expo's `.ios.tsx` / `.android.tsx` / `.web.tsx` file extensions, or `Platform.select()` inline. Business logic must not be platform-specific.
- **API client**: all frontend code calls the Go server via `src/api/`. No frontend code calls upstream services directly.
