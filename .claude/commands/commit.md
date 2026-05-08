# Commit workflow

Run the full commit procedure for the current staged changes.

## Steps

1. Run `git diff --staged --stat` and `git diff --staged` to understand what is changing.

2. **Spec drift check** — if any file under `internal/`, `app/`, or `src/` changed, verify the corresponding spec in `specs/` still accurately describes the behavior. If it does not, update the spec first, then proceed. Never commit implementation that diverges from the spec.

3. **Format the commit message** using [Conventional Commits](https://www.conventionalcommits.org/):

   | Type | When to use | Version bump |
   |------|-------------|--------------|
   | `feat` | new user-visible feature | minor |
   | `fix` | bug fix | patch |
   | `feat!` / `fix!` | breaking change (add `BREAKING CHANGE:` footer) | major |
   | `chore` | tooling, deps, config, scaffolding | none |
   | `docs` | spec, ADR, session log, README | none |
   | `test` | adding or updating tests | none |
   | `refactor` | internal restructuring, no behavior change | none |
   | `perf` | measurable performance improvement | patch |

   Scope is optional but encouraged: `feat(plugin-sonarr):`, `fix(activity-feed):`.

   Subject rules: imperative mood, lowercase, no trailing period, ≤72 characters.

   Multi-line body: use when the *why* is non-obvious. Blank line between subject and body.

4. **Update CHANGELOG.md** — only for `feat`, `fix`, and breaking changes (not `chore`, `docs`, `test`, `refactor`):
   - Add to the `## [Unreleased]` section (create it at the top if it doesn't exist)
   - Sub-sections: `### Added`, `### Changed`, `### Fixed`, `### Removed`
   - One line per entry: `- Short description (#issue or PR number if known)`
   - If CHANGELOG.md does not exist yet, create it with an `## [Unreleased]` section

5. **Stage any spec or CHANGELOG updates** made in steps 2–4 before committing.

6. Create the commit. Never use `--no-verify`. Never amend unless the user explicitly requested it.

## Versioning

This project uses [Semantic Versioning](https://semver.org/). The current version lives in `VERSION` at the repo root (create it with `0.0.0` if it doesn't exist yet).

When the user asks to cut a release:
1. Determine the next version from the commit types since the last tag (`feat` → minor, `fix` → patch, `BREAKING CHANGE` → major)
2. Update `VERSION`
3. Rename `## [Unreleased]` in CHANGELOG.md to `## [X.Y.Z] — YYYY-MM-DD`
4. Add a new empty `## [Unreleased]` above it
5. Commit with message `chore: release vX.Y.Z`
6. Tag: `git tag vX.Y.Z`
