# ADR 0005: Upstream User Identity Linking

## Status

Accepted

## Date

2026-05-07

## Context

Mortar owns its own login and roles, but some features are user-specific in upstream systems:

- Playback handoff may need a Jellyfin account
- Continue Watching requires per-user playback history
- Some activity events are only visible to the requester

Without a mapping model between a Mortar user and an upstream account, the contract cannot reliably support per-user behavior.

## Decision

Mortar remains the system of record for authentication. Per-plugin external account links are optional, explicit records that connect a Mortar user to a specific configured plugin's upstream user id when a feature requires it.

## Policy

- Store external account links in Mortar's local database.
- Plugins that need user-specific upstream context receive the matching external account link through `MortarUser.external_accounts`.
- External account links are optional overall and are only required for features that need user-specific upstream context.
- Request submission may still use a service account where the upstream system does not require an end-user login.
- Playback and personalized resume features require an external account link for the target plugin.
- If a feature requires an external account link and none exists, Mortar must fail explicitly with clear UX rather than hiding the behavior or guessing.
- In `v1`, the Browse & Play surface is link-gated for the relevant plugin rather than split between anonymous browse and linked playback.
- `v1` starts with config-file or bootstrap seeding for external account links rather than a dedicated admin UI.
- Mortar does not auto-provision guest or shadow accounts in v1.

## Rationale

- This keeps Mortar auth independent while still allowing personalized upstream behavior.
- It makes missing-account behavior explicit instead of relying on hidden fallback rules.
- It avoids silently impersonating one shared account for every user.
- It avoids introducing a split "shared browse, personalized play" model before the product has clearer operator and user UX around linking.

## Consequences

- Browse & Play must define the UX when a required external account link is missing.
- Admin setup needs a way to manage external account links, even if v1 starts with config-file seeding instead of a UI.
- Future SSO work can build on this model instead of replacing it.
- The architecture and feature specs should distinguish between features that can run on Mortar identity alone and features that require explicit upstream identity links.

## Related

- [Upstream Identity Linking Decision Session](../sessions/2026-05-07-upstream-identity-linking.md)
- [Browse & Play Spec](../../specs/features/browse-play.md)
