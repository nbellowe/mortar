# Feature: Home

## Metadata

- **Status:** `accepted`
- **Depends on:** [Plugin Interface](../plugins/plugin-interface.md), [ADR 0002](../../docs/adrs/0002-persistence-and-state.md), [ADR 0003](../../docs/adrs/0003-realtime-updates.md), [ADR 0004](../../docs/adrs/0004-plugin-response-caching.md), [ADR 0005](../../docs/adrs/0005-upstream-user-identity-linking.md)
- **Last updated:** `2026-05-07`

## Goal

The home screen gives household users a fast at-a-glance starting point: what is new, what they can resume, and whether the stack currently needs attention.

## User flows

### Opening Home

1. User opens the Mortar home screen.
2. Mortar loads each home surface independently using the configured library and health sources.
3. Home MAY show up to three surfaces:
   - **Recently Added**
   - **Continue Watching**
   - **Service health badge**
4. A failure in one surface MUST NOT blank the entire home screen.

### Recently Added

1. Mortar reads recent library additions from plugins with `library.browse` capability, typically Jellyfin in the supported `v1` stack.
2. If browsing that library requires an external account link and the user does not have one, Mortar MUST show a clear linked-account-required state for the row rather than shared or anonymous data.
3. If data is available, Mortar shows the 10 most recent additions as horizontal cards with poster, title, year, and type.
4. Selecting a card opens the same item detail flow defined by [Browse & Play](browse-play.md).

### Continue Watching

1. If a library plugin exposes `library.resume`, Mortar requests in-progress items for the current user.
2. If the required external account link is missing, Mortar MUST show a clear linked-account-required state instead of personalized playback data.
3. If the capability is not available, Mortar omits the row entirely.
4. Selecting a card opens the same detail or playback handoff flow defined by [Browse & Play](browse-play.md).

### Service Health Badge

1. Mortar reads the last-known health snapshot summary rather than blocking on a fresh live probe.
2. If any configured plugin is currently unreachable, Mortar shows a badge in home or primary navigation indicating degraded stack health.
3. Admins can use the badge as an entry point to the detailed [Service Health](health.md) view.
4. Regular users MAY see the badge, but MUST NOT gain access to the health detail view.

## Acceptance criteria

- [ ] Home loads useful content without waiting on a fresh live health probe.
- [ ] Recently Added shows the 10 most recent additions from the supported library source.
- [ ] Continue Watching is shown only when the configured library plugin supports `library.resume`.
- [ ] Selecting a home card opens the same item detail or playback flow used by Browse & Play.
- [ ] When a required external account link is missing, affected home rows fail explicitly with clear UX rather than showing shared or anonymous fallback data.
- [ ] The health badge appears when any configured plugin is unreachable based on last-known health state.
- [ ] A failure in one home surface degrades gracefully without blanking the entire home screen.

## Plugin dependencies

| Plugin type | Required capability |
|---|---|
| Jellyfin | `library.browse` |
| Jellyfin (optional) | `library.resume` |
| All plugins | Base `health()` method for badge summary input |

## Documentation impact

- **App users** — document what appears on Home, what Continue Watching requires, and what the health badge means.
- **Operators** — document that Home depends on a configured library plugin, linked upstream accounts for personalized rows, and last-known health snapshots.
- **Contributors** — treat this spec as the canonical source for home-screen behavior instead of duplicating row or badge rules across other feature specs.

## Out of scope (v1)

- Customizable or reorderable home widgets.
- Recommendations or personalized discovery beyond continue-watching state.
- Shared-account or guest fallbacks for library rows that require a linked upstream account.
