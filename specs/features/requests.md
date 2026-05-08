# Feature: Search & Request

## Metadata

- **Status:** `accepted`
- **Depends on:** [Plugin Interface](../plugins/plugin-interface.md), [ADR 0002](../../docs/adrs/0002-persistence-and-state.md), [ADR 0004](../../docs/adrs/0004-plugin-response-caching.md), [ADR 0006](../../docs/adrs/0006-request-routing-policy.md)
- **Last updated:** `2026-05-07`

## Goal

A household user can find any media — movie, show, audiobook, ebook — and request it, without knowing which underlying service handles each type.

## User flows

### Search

1. User types a query into the global search box.
2. Mortar fans the query out to all plugins with a `requests.*` capability simultaneously.
3. Results are normalized into `MediaItem` and deduplicated (same TMDB/IMDB/ASIN ID from multiple plugins = one result).
4. Results are grouped by type: Movies, Shows, Audiobooks, Ebooks, limited to the request capabilities available in the current install.
5. Each result shows: poster, title, year, type, and status badge (see below).

### Status badge

Each result displays one of:

| Badge | Meaning |
|---|---|
| Available | A matching library item was returned by `library.exists` |
| Requested | A pending request exists for this item |
| Request | Item is not available and not requested |

### Requesting

1. User clicks a result marked "Request".
2. A detail modal opens with: poster, title, overview, genres, year.
3. User confirms the request.
4. Mortar routes the request to the correct plugin based on media type and the configured routing policy.
5. Modal updates to show "Requested" with current status.

### Request status tracking

- All users can view all pending and historical requests.
- Mortar keeps durable local request snapshots for request history, duplicate-request checks, and faster views.
- Upstream request plugins remain the source of truth for current request status and review actions.
- Mortar does not provide an approve/deny UI. Admins who need to action a request are linked out to the upstream service (e.g., Jellyseerr).

## Acceptance criteria

- [ ] Search returns results within 3 seconds under normal network conditions.
- [ ] Search only fans out to plugins with a `requests.*` capability. Library plugins are not queried during search.
- [ ] When one search plugin times out or fails, partial results from successful plugins are shown with a non-blocking notice identifying the unavailable plugin.
- [ ] Results from multiple plugins are deduplicated by external ID.
- [ ] Media type routing follows the configured policy for each request capability.
- [ ] Ambiguous request routing is rejected at startup rather than resolved by plugin declaration order.
- [ ] Items already in the library show "Available" without a request option.
- [ ] A user cannot submit a duplicate request for an already-pending item.
- [ ] All users can see all pending and historical requests.
- [ ] Admins are linked to the upstream service for approve/deny actions. Mortar does not implement a review UI.

## Plugin dependencies

| Plugin type | Required capability |
|---|---|
| Jellyseerr | `requests.video` |
| AudioBookRequest | `requests.audio` |
| Shelfarr (optional) | `requests.ebook` |
| Jellyfin | `library.exists` (for status badges and library links) |

