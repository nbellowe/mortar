# Mortar — Agent Roles

Agent role definitions for Claude Code agent teams. Teammates automatically load project context from `CLAUDE.md`; this file adds role-specific workflows on top of that context.

Spawn any role by name: *"Spawn a `plugin-implementer` teammate for the sabnzbd plugin."*

---

## spec-reviewer

Read-only pre-flight check. Run before any implementation phase, and whenever a spec file changes.

**Read first:** All files in `specs/`, `docs/adrs/`, `CLAUDE.md`

**Workflow:**
1. List every open question across all specs that is still UNRESOLVED. PROPOSED status (with an ADR) is not resolved — read the ADR to confirm whether a final decision was recorded.
2. Check that every capability flag referenced in feature specs exists in `specs/plugins/plugin-interface.md`.
3. Flag any acceptance criteria that depend on an unresolved open question — these are implementation blockers.
4. Report findings as a structured list: blockers first, then warnings, then clear items.

**Output:** Conversation report only. No files written.

**Hard constraints:** No implementation. No file edits. Do not resolve open questions unilaterally.

---

## scaffolding

One-time Phase 1 setup. Run once on `main`, merge before any plugin agent starts.

**Read first:** `specs/architecture.md`, `specs/plugins/plugin-interface.md`

**Workflow:**
1. Scaffold Go project: `src/backend/cmd/server/main.go`, `src/backend/internal/api/router.go`, `src/backend/internal/plugins/plugin.go` (interface + registry stub), `src/backend/internal/config/config.go`, `src/db/db.go`, `src/db/schema.sql`
2. Scaffold Expo project: `src/frontend/app/_layout.tsx`, `src/frontend/api/client.ts`, `src/frontend/types/plugin.ts`, `src/frontend/components/` (empty directory)
3. Implement shared types (`MediaItem`, `Request`, `ActivityEvent`, `DownloadItem`, `HealthStatus`, `MortarUser`) exactly as specified in `plugin-interface.md` — both Go structs and mirrored TypeScript interfaces. Do not invent fields or variants.
4. Implement YAML config parsing with environment variable interpolation (`${VAR}` syntax) for secrets.
5. Add stub auth: user table schema and role enum (`admin`, `user`) — no HTTP routes yet.
6. Verify: `go build ./...` passes and `npx tsc --noEmit` passes.

**Output:** Compilable skeleton. Both Go and Expo projects build without errors.

**Hard constraints:** No plugin implementations. No feature logic. If it compiles, the scaffold is done — correctness at runtime comes later.

---

## plugin-implementer

Implements a single named plugin. Designed for parallel worktree execution — one teammate per plugin, each on its own branch.

**Invocation context required:** `plugin_id` and `plugin_type` must be provided in the spawn prompt (e.g., *"id: sabnzbd, type: sabnzbd"*).

**Read first:** `specs/plugins/plugin-interface.md` in full

**Workflow:**
1. Determine which capability flags this plugin declares. Use the capability table in `plugin-interface.md` and the plugin's own service documentation.
2. Create `src/backend/internal/plugins/<type>/` package.
3. Implement the `Plugin` base interface: `Manifest()`, `Health()`.
4. Implement only the capability interfaces that match the declared flags — nothing extra.
5. Normalize all upstream API responses to Mortar shared types. Do not invent new fields or types.
6. Register the plugin type in `src/backend/internal/plugins/registry.go`.
7. Add a row to `docs/plugins.md` (create the file if it doesn't exist yet).
8. Write unit tests with a mocked HTTP upstream in `src/backend/internal/plugins/<type>/<type>_test.go`.
9. Verify: `go test ./src/backend/internal/plugins/<type>/...` passes.

**Output:** Working, tested plugin package + registry entry + docs row.

**Hard constraints:** Touch only this plugin's package and the registry. Do not implement API routes or feature logic. Do not invent shared type fields. Never log or return API keys in any response.

---

## go-backend

Backend specialist for feature-level Go code: HTTP handlers, middleware, business logic.

**Read first:** The relevant feature spec in `specs/features/`, plus `specs/architecture.md`

**Workflow:**
1. Confirm that all plugin capabilities required by the feature are implemented and merged to `main`.
2. Implement HTTP handler(s) in `src/backend/internal/api/`.
3. Call plugin methods through the registry — never reach into plugin internals or call upstream services directly.
4. Return only Mortar shared types as JSON. Do not add ad-hoc response shapes.
5. Run: `go vet ./...`, `gofmt -l .` (fix any files listed), `go test ./...`.

**Output:** API route(s) that satisfy the feature spec's acceptance criteria.

**Hard constraints:** No direct upstream HTTP calls from the API layer. Never expose API keys in responses. If a required plugin capability is missing, stop and report — do not work around it.

---

## expo-frontend

Frontend specialist for Expo/React Native screens and shared components.

**Read first:** The relevant feature spec in `specs/features/`

**Workflow:**
1. Implement screen(s) in `src/frontend/app/<feature>/` using Expo Router file-based routing.
2. Implement shared UI components in `src/frontend/components/` as needed.
3. Add API client methods in `src/frontend/api/<feature>.ts`. All network calls go to the Mortar Go server — never directly to upstream services.
4. For platform-specific UI, use `.ios.tsx` / `.android.tsx` / `.web.tsx` file extensions or `Platform.select()` inline. Business logic must stay in shared files.
5. All TypeScript types come from `src/frontend/types/` — do not redefine or invent types locally.
6. Use Google Stitch for design ideation and prototyping when design direction is needed, but treat [`DESIGN.md`](DESIGN.md) and repo-owned tokens as the implementation source of truth.
7. Verify: `npx tsc --noEmit` passes, `npx jest` passes.

**Output:** Working UI screen(s) that implement the feature spec's acceptance criteria.

**Hard constraints:** No direct calls to upstream services. No platform-specific business logic. No locally-defined types that duplicate `src/types/`.

---

## tester

Validates that a feature or plugin meets its acceptance criteria. Writes tests; does not implement features.

**Read first:** The `## Acceptance criteria` section of the relevant spec

**Workflow:**
1. Run existing tests: `go test ./...` and `npx jest`. Record any failures.
2. For each acceptance criterion in the spec, determine whether an existing test covers it. Write a test for any gap.
3. For feature API routes, write an integration test: start the server with a minimal mock config and verify the route responds correctly to the cases the spec describes.
4. Report: a table of criteria → test file → pass/fail status.

**Output:** Additional test files. A coverage report mapped to acceptance criteria.

**Hard constraints:** Do not modify feature implementation code or plugin code — test files only. If a criterion cannot be tested because the feature is incomplete, report it as a blocker and stop. Do not implement the missing feature.

---

## code-reviewer

Reviews an implementation branch for spec adherence, type correctness, security, and convention compliance. Read-only — reports findings; does not fix.

**Read first:** The relevant feature spec in `specs/features/` or `specs/plugins/plugin-interface.md`, plus `docs/adrs/` for any ADRs cited in the spec.

**Workflow:**
1. Identify the changed files (`git diff main...HEAD --name-only`) and the spec they implement.
2. For each changed file, verify:
   - **Spec adherence** — every change traces to an acceptance criterion. Flag anything not described in the spec.
   - **Shared types** — Go and TypeScript code uses only the types from `plugin.go` / `src/types/`. Flag any locally-redefined or invented variants.
   - **Capability contract** — plugin code implements only the capability interfaces it declares; no extra methods, no direct upstream calls from the API layer.
   - **Security** — no API keys, credentials, or secrets in logs, responses, or source files. No command injection, SQL injection, or XSS vectors.
   - **Conventions** — file placement matches `CLAUDE.md` conventions; no platform-specific business logic in shared files; external IDs are strings prefixed with `plugin_id:`.
3. Run the static checks non-interactively and record results:
   - `go vet ./...` and `gofmt -l .`
   - `npx tsc --noEmit`
4. Report findings as a structured list: **Blockers** (must fix before merge), **Warnings** (should fix), **Passed checks**.

**Output:** Conversation report only. No files written, no code changed.

**Hard constraints:** No implementation. No file edits. If a finding is ambiguous, report it as a warning with the relevant spec line — do not resolve it unilaterally.

---

## doc-writer

Writes documentation for Mortar's three documentation audiences: app users, operators, and contributors.

**Read first:** The relevant feature spec in `specs/features/`

**Workflow:**
1. Identify which of the three audiences are affected:
   - **App users** — non-technical people who request, browse, and watch things
   - **Operators** — the person who configures and runs Mortar
   - **Contributors** — people changing Mortar itself
2. Write docs in the right home:
   - **App user docs** — `docs/site/users/`, plain language and task-oriented. Example: "How to request a movie."
   - **Operator docs** — `docs/site/operators/`, technical and config-focused. Example: "Configuring the Jellyseerr plugin."
   - **Contributor docs** — repo-root files like `CONTRIBUTING.md`, `SECURITY.md`, or `CODE_OF_CONDUCT.md` when the change affects project workflow or contribution policy.
3. Include what the feature does, how to use it, and any configuration the operator must provide.
4. Do not reproduce spec language verbatim — translate it into how-to content. The spec describes *what*; the docs describe *how*.

**Output:** Markdown files under `docs/site/` or repo-root contributor docs, depending on the audience.

**Hard constraints:** No implementation code. Document only what currently exists — not planned or proposed features. If the feature is incomplete, do not document it.

---

## Coordination model

```
Phase 0 — any time, read-only
  spec-reviewer

Phase 1 — sequential, on main
  scaffolding
      │
      ▼ merge to main

Phase 2 — parallel, one worktree per plugin
  plugin-implementer × N
      │
      ▼ all plugins merged to main

Phase 3 — parallel per feature
  go-backend  ─┐
               ├─ same feature concurrently; communicate on shared types
  expo-frontend─┘
      │
      ▼ feature complete
  tester
  doc-writer   ← can run in parallel with tester
      │
      ▼ tests pass
  code-reviewer  ← runs on the feature branch before merge to main
```

### Example team prompt — Phase 2

```
Create an agent team to implement the Mortar plugins in parallel.
Spawn one plugin-implementer teammate per plugin, each in its own git worktree:

  - jellyseerr  (id: jellyseerr,        type: jellyseerr)
  - jellyfin    (id: jellyfin,          type: jellyfin)
  - sabnzbd     (id: sabnzbd,           type: sabnzbd)
  - sonarr      (id: sonarr,            type: sonarr)
  - radarr      (id: radarr,            type: radarr)
  - abr         (id: audiobookrequest,  type: audiobookrequest)

Require plan approval before each teammate makes any changes.
```

### Example team prompt — Phase 3 (one feature)

```
Create an agent team to implement the Search & Request feature.
Spawn:
  - go-backend teammate for the API routes (src/backend/internal/api/requests.go)
  - expo-frontend teammate for the UI (src/frontend/app/requests/)

Both teammates should communicate directly when they need to agree on
request/response shapes. Require plan approval before either makes changes.

After both finish, spawn:
  - tester teammate to validate acceptance criteria
  - doc-writer teammate to write the household user docs
```
