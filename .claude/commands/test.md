# Test suite

Run the full test suite and report results with coverage.

## Go backend

Run these in order — stop and report if any step fails:

```bash
# 1. Format check — output should be empty
gofmt -l .

# 2. Vet
go vet ./...

# 3. Tests with coverage
go test ./... -coverprofile=coverage.out -covermode=atomic

# 4. Per-package coverage report
go tool cover -func=coverage.out
```

**Coverage thresholds** (warn if below, do not fail the build):

| Path | Threshold |
|------|-----------|
| `internal/plugins/*/` | ≥ 80% |
| `internal/api/` | ≥ 70% |
| Overall | ≥ 60% |

**Integration tests** live in `internal/integration/` and require live plugin credentials. They are tagged `integration` and never run in CI:

```bash
go test -tags integration ./internal/integration/...
```

Run integration tests manually before cutting a release. They are not a gate on merging.

## TypeScript / Expo frontend

```bash
# 1. Type check — must pass with zero errors
npx tsc --noEmit

# 2. Tests with coverage
npx jest --coverage
```

**Coverage thresholds** (warn if below):

| Path | Threshold |
|------|-----------|
| `src/api/` | ≥ 80% |
| `src/components/` | ≥ 60% |
| Overall | ≥ 60% |

## CI expectations

A CI run (GitHub Actions or equivalent) must execute:
1. `gofmt -l .` — fail if output is non-empty
2. `go vet ./...`
3. `go test ./... -coverprofile=coverage.out`
4. `npx tsc --noEmit`
5. `npx jest`

CI never runs integration tests. CI runs on every push to `main` and on every PR targeting `main`.

If a `.github/workflows/ci.yml` does not exist yet, offer to create it after reporting test results.

## Reporting

After running all steps, summarize:
- Pass / fail for each step
- Any files failing the format check (list them)
- Per-package coverage (Go) and per-directory coverage (Jest)
- Highlight any package below its threshold with a warning
- Overall status: **PASS** (all steps green) or **FAIL** (any step failed)
