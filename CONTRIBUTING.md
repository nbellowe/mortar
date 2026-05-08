# Contributing

Thanks for your interest in Mortar.

Mortar is still in an early spec-first phase, so the best contributions right now are usually one of these:

- refining specs
- resolving ADRs
- improving project docs
- implementing work that is clearly unblocked by the current specs

## Before you change code

1. Read the relevant spec in `specs/`
2. Check any referenced ADRs in `docs/adrs/`
3. Confirm the spec is not blocked by unresolved open questions
4. Align your change with the current roadmap in `ROADMAP.md`

## Contribution guidelines

- Keep changes scoped and explain the why, not just the diff
- Do not implement around unresolved spec questions without first resolving the spec
- Prefer updating docs when a change affects users, operators, or contributors
- Never include secrets, API keys, or private service data in commits

## Docs audiences

Mortar maintains docs for three audiences:

- **App users** — people using Mortar day to day
- **Operators** — people deploying and configuring Mortar
- **Contributors** — people changing Mortar itself

Use these homes:

- `docs/site/users/`
- `docs/site/operators/`
- repo-root contributor docs like this file

## Development status

Implementation has not started yet. If you want to help early, documentation and spec work are especially valuable.
