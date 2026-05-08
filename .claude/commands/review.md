# Review workflow

Review the current branch against `main`. Usage: `/review`

## Steps

1. **Understand the changes** — run `git log main..HEAD --oneline` and `git diff main...HEAD --stat` to get a full picture of what's on this branch.

2. **Spec alignment** — for every changed file under `internal/`, `app/`, or `src/`, verify the corresponding spec in `specs/` still accurately describes the behavior. Call out any drift.

3. **Plugin contract** — if any plugin implementation changed, verify:
   - Declared capabilities match implemented interfaces
   - Shared types (`MediaItem`, `Request`, etc.) are used exactly as defined — no added fields, no local variants
   - No upstream HTTP calls outside the plugin package

4. **Code quality**
   - `go vet ./...` passes (backend)
   - `npx tsc --noEmit` passes (frontend)
   - No hardcoded credentials or URLs that should be config
   - No unresolved open questions in specs being implemented against

5. **Session doc** — every PR must include a session log in `docs/sessions/YYYY-MM-DD-<slug>.md`. If one is missing, create it before the review is complete. The session log should record what was decided and why, what was ruled out, and any open questions raised. See `/doc session` for the format.

6. **Summary** — produce a short review summary:
   - What the branch does
   - Any spec drift or contract violations found
   - Whether a session doc exists (or was created)
   - Any open questions or suggested follow-ups
