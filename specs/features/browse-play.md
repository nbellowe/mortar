# Feature: Browse & Play

## Metadata

- **Status:** `accepted`
- **Depends on:** [Plugin Interface](../plugins/plugin-interface.md), [ADR 0004](../../docs/adrs/0004-plugin-response-caching.md), [ADR 0005](../../docs/adrs/0005-upstream-user-identity-linking.md)
- **Last updated:** `2026-05-07`

## Goal

Household members can browse what's already available in the library and launch playback, without leaving Mortar or knowing which media server holds what.

## User flows

### Browsing

1. User opens the library view.
2. Mortar checks whether the user has the required external account link for the target library plugin.
3. If the required link is missing, Mortar shows a clear message that browsing this library requires a linked account for that service.
4. If the link exists, Mortar fetches content from all plugins with `library.browse` capability (typically Jellyfin).
5. Content is displayed in a grid: poster, title, year, type badge.
6. Users can filter by: type (movies, shows, audiobooks), genre, and sort by (recently added, title, year).
7. Pagination or infinite scroll handles large libraries.

### Playing

1. User clicks a library item.
2. A detail view opens: full metadata, poster, overview, genres, cast (where available).
3. If the required external account link for the target plugin is missing, Mortar shows a clear linked-account requirement instead of attempting playback.
4. If the link exists, a "Play" button generates a deep link URI via the plugin's `getPlayUrl()` and hands it off to the OS.
5. The OS opens the URI in the appropriate external app (e.g., the Jellyfin app, Infuse, VLC). Mortar never plays media inline.

Mortar does not implement a video or audio player. Playback is always delegated to an external app via a deep link URI on all platforms.

### Home entry points

The home-screen behavior for Recently Added, Continue Watching, and the health badge is specified in [Home](home.md).

When a user opens a library item from Home, Mortar MUST route to the same detail and playback handoff flows described in this spec.

## Acceptance criteria

- [ ] When the required external account link is missing, the Browse & Play surface fails explicitly with clear UX rather than using a guest or shared fallback.
- [ ] Library grid loads within 3 seconds for libraries up to 5,000 items.
- [ ] Genre and type filters apply without a full page reload.
- [ ] "Play" hands off to an external app via the deep link URI returned by the target library plugin. Mortar does not play inline.
- [ ] Items opened from Home use the same detail and playback flows as items opened from the library browse view.
- [ ] Items marked as "Available" in search link through to the library detail view.

## Plugin dependencies

| Plugin type | Required capability |
|---|---|
| Jellyfin | `library.browse`, `library.exists` |

## Out of scope (v1)

- In-Mortar video or audio playback.
- Managing watchlists or favorites within Mortar.
- Multiple library plugin sources (one Jellyfin instance is sufficient for v1).

