# Upstream Identity Linking Decision Session — 2026-05-07

## Why this session happened

Mortar owns its own authentication, but some plugin features need per-user upstream context to behave honestly.

The repo needed a clear answer for how Mortar users map to plugin-specific users, and what should happen when that mapping does not exist.

## Decisions made

### Core identity model

- Mortar remains the system of record for login and roles
- Upstream identity links are optional explicit records
- Links are scoped per Mortar user and per configured plugin instance

### Feature requirements

- Links are only required for features that need user-specific upstream context
- Request submission may still use a service account where the upstream system does not require end-user identity
- Playback and continue-watching require the relevant external account link

### Missing-link behavior

- When a feature requires a link and none exists, Mortar must fail explicitly with clear UX
- Mortar should not hide the dependency or guess its way through it
- Mortar should not use guest, shadow, or shared fallback accounts in `v1`

### Browse & Play posture

- For `v1`, the Browse & Play surface is link-gated for the relevant plugin
- We intentionally did not adopt a split model where browsing works through a shared plugin account but playback and resume require a personal link

### Provisioning approach

- `v1` starts with config-file or bootstrap seeding for upstream account links
- A dedicated admin UI or self-service linking flow can come later

## Why this direction won

- It keeps Mortar auth independent
- It avoids hidden impersonation and ambiguous shared-account behavior
- It keeps the initial model honest and supportable, even if it is stricter than the eventual product direction

## Immediate follow-up chosen from this session

- Mark ADR 0005 as accepted
- Update the architecture and Browse & Play specs to reflect explicit link-gated behavior
