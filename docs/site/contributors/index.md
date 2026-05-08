# Development setup

## Prerequisites

- Go 1.21+
- Node 18+ with npm
- Git

## Running locally

**Backend:**

```bash
go run ./cmd/server
```

The server starts on `http://localhost:8080`. It looks for `config.yaml` in the working directory by default.

**Frontend:**

```bash
npm install
npx expo start --web
```

The web client starts on `http://localhost:8081` and proxies API calls to the backend.

## Tests

```bash
# Backend
go test ./...

# Frontend type check
npx tsc --noEmit

# Frontend unit tests
npx jest
```

## Before you change anything

Mortar is spec-driven. The specs in `specs/` are the source of truth — if a spec and the code disagree, the spec wins.

1. Read the relevant spec in `specs/`
2. Check any referenced ADRs in `docs/adrs/`
3. Confirm the spec has no unresolved open questions
4. Use the spec's acceptance criteria as your definition of done

See [CONTRIBUTING.md](https://github.com/nbellowe/mortar/blob/main/CONTRIBUTING.md) for contribution guidelines.

## Key directories

| Path | What it contains |
|---|---|
| `cmd/server/` | Server entry point |
| `internal/api/` | HTTP handlers |
| `internal/plugins/` | Plugin registry and implementations |
| `app/` | Expo Router screens |
| `src/api/` | Frontend API client |
| `src/components/` | Shared UI components |
| `src/types/` | Shared TypeScript types |
| `specs/` | Feature and architecture specs |
| `docs/adrs/` | Architecture decision records |
