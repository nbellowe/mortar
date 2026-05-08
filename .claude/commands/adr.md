# Create an ADR

Create a new Architecture Decision Record for a cross-cutting technical decision.

Usage: `/adr <title>`

Arguments: $ARGUMENTS

## Steps

1. List `docs/adrs/` to find the next sequential number. If the highest existing file is `0006-*.md`, the next number is `0007`. Zero-pad to 4 digits.

2. Slugify the title: lowercase, spaces → hyphens, strip punctuation. Example: "Plugin Response Caching" → `plugin-response-caching`.

3. Create `docs/adrs/<NNNN>-<slug>.md` with this structure:

```
# ADR-NNNN: <Title>

**Status:** Proposed
**Date:** <today's date, YYYY-MM-DD>
**Affects:** <list the features, components, or specs this decision touches>

## Context

<Describe the problem or situation. What constraints or forces are in play? Why does a decision need to be made now?>

## Decision

<State the decision in 1–2 sentences. Be direct: "We will…">

## Consequences

**Good:**
-

**Bad / trade-offs:**
-

## Alternatives considered

| Option | Why rejected |
|--------|-------------|
|        |              |
```

4. Add a row to the **Open questions** table in both `CLAUDE.md` and `AGENTS.md`:

   ```
   | <title> | **PROPOSED** — see `docs/adrs/<NNNN>-<slug>.md` | <affected areas> |
   ```

5. Report the file path created. Remind: change the status from **Proposed** to **Accepted** or **Rejected** (with a brief rationale note) before the decision unblocks any dependent work.

## What belongs in an ADR

ADRs are for **cross-cutting technical decisions** — choices that affect multiple features, establish a project-wide pattern, or involve significant trade-offs. Examples: persistence strategy, real-time delivery model, auth approach, caching policy.

ADRs are **not** for feature design (that belongs in `specs/features/`) or per-plugin implementation choices (document those in the plugin spec or code comments).
