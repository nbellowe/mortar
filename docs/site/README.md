# Documentation Structure

Mortar documentation serves three audiences:

- **App users** — people using Mortar to request, browse, and track media
- **Operators** — people configuring and running Mortar
- **Contributors** — people changing Mortar itself

## Directory map

- [`users/`](users/) — task-oriented end-user docs
- [`operators/`](operators/) — installation, configuration, and operations docs

Contributor-facing docs live at the repo root rather than inside `docs/site/`:

- [`CONTRIBUTING.md`](../../CONTRIBUTING.md)
- [`SECURITY.md`](../../SECURITY.md)
- [`CODE_OF_CONDUCT.md`](../../CODE_OF_CONDUCT.md)

## Authoring rule of thumb

- If it helps someone use Mortar, it belongs in `docs/site/users/`
- If it helps someone run Mortar, it belongs in `docs/site/operators/`
- If it helps someone change Mortar, it belongs in the contributor docs at repo root
