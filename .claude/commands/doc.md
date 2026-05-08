# Documentation workflow

Guides for the three types of documentation in this project. Usage: `/doc <type> <name>`

Arguments: $ARGUMENTS

Parse the first word of $ARGUMENTS as the type (`spec`, `session`, `plugin`). The remainder is the name or title.

---

## `spec` — write or update a feature spec

Usage: `/doc spec <feature-name>`

Feature specs live in `specs/features/<feature-name>.md`. All features begin as specs; implementation must not start until the spec has no unresolved open questions.

If the file **does not exist**, create it:

```
# <Feature Name>

**Status:** Draft
**Last updated:** <today>
**Related ADRs:** (fill in if applicable)

## Summary

<One paragraph: what this feature does and why it exists.>

## Audiences

<Which users does this feature serve: household users, homelab owner, or both?>

## Acceptance criteria

- [ ] <observable behavior 1>
- [ ] <observable behavior 2>

## Open questions

| Question | Status | Affects |
|----------|--------|---------|
| | | |

## Out of scope

<Explicit non-goals for this feature.>
```

If the file **already exists**, identify what changed (from `git diff` or the user's description), update the relevant sections, and change **Status** to `Updated` and **Last updated** to today.

Spec rules:
- Acceptance criteria are observable behaviors, not implementation steps
- Every open question must be resolved before implementation begins
- If a question is resolved, move it to a footnote or remove it; do not leave stale PROPOSED rows

---

## `session` — record a design session log

Usage: `/doc session <short-description>`

Session logs live in `docs/sessions/YYYY-MM-DD-<slug>.md`. They record *why* decisions were made, not what the current design is. Specs and ADRs are the authoritative source; session logs are context.

Create the file:

```
# <Date>: <Short description>

## What we decided

-

## Key constraints or context

-

## What we explicitly ruled out

-

## Open questions it raised

-
```

Keep entries short and factual. A session log should be readable in under two minutes.

---

## `plugin` — document a plugin

Usage: `/doc plugin <plugin-name>`

1. Read `specs/plugins/plugin-interface.md` to understand the capability flags.
2. Check whether `docs/plugins.md` exists. If not, create it with this header:

   ```
   # Mortar Plugins

   | Plugin | Type | Capabilities | Status |
   |--------|------|-------------|--------|
   ```

3. Add (or update) a row for the plugin:

   | Plugin | Type | Capabilities | Status |
   |--------|------|-------------|--------|
   | Name | `sonarr` / `radarr` / etc. | comma-separated capability flags | `planned` / `in-progress` / `complete` |

4. If the plugin implementation file at `internal/plugins/<type>/<type>.go` exists, verify the capabilities listed in `docs/plugins.md` match the `Capabilities()` method in the code.
