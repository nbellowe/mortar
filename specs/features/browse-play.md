# Feature: Browse & Play

## Goal

Household members can browse what's already available in the library and launch playback, without leaving Mortar or knowing which media server holds what.

## User flows

### Browsing

1. User opens the library view.
2. Mortar fetches content from all plugins with `library.browse` capability (typically Jellyfin).
3. Content is displayed in a grid: poster, title, year, type badge.
4. Users can filter by: type (movies, shows, audiobooks), genre, and sort by (recently added, title, year).
5. Pagination or infinite scroll handles large libraries.

### Playing

1. User clicks a library item.
2. A detail view opens: full metadata, poster, overview, genres, cast (where available).
3. A "Play" button generates a deep link via the plugin's `getPlayUrl()` and opens it.
4. The player opens in a new tab (Jellyfin web player, or the Jellyfin app if configured).

Mortar does not implement a video player. Playback is always delegated to the native app.

### Continue watching / recently added (home screen)

The home screen surfaces two rows:
- **Recently Added** — last 10 items added to the library
- **Continue Watching** — items with in-progress playback (if the plugin supports it)

These are sourced from the `library.browse` plugin with appropriate sort/filter options.

## Acceptance criteria

- [ ] Library grid loads within 3 seconds for libraries up to 5,000 items.
- [ ] Genre and type filters apply without a full page reload.
- [ ] "Play" deep link opens the correct item in Jellyfin.
- [ ] Recently Added row on home screen reflects last 10 additions.
- [ ] Items marked as "Available" in search link through to the library detail view.

## Plugin dependencies

| Plugin type | Required capability |
|---|---|
| Jellyfin | `library.browse`, `library.exists` |

## Out of scope (v1)

- In-Mortar video or audio playback.
- Managing watchlists or favorites within Mortar.
- Multiple library plugin sources (one Jellyfin instance is sufficient for v1).

## Open questions

- Should Mortar show Audiobookshelf content in the library browse view, or is that a separate section?
- How do we handle users who are not registered in Jellyfin — can we generate a guest play link, or do they need a Jellyfin account?
