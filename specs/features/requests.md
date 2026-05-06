# Feature: Search & Request

## Goal

A household user can find any media — movie, show, audiobook, ebook — and request it, without knowing which underlying service handles each type.

## User flows

### Search

1. User types a query into the global search box.
2. Mortar fans the query out to all plugins with a `requests.*` capability simultaneously.
3. Results are normalized into `MediaItem` and deduplicated (same TMDB/IMDB/ASIN ID from multiple plugins = one result).
4. Results are grouped by type: Movies, Shows, Audiobooks, Ebooks.
5. Each result shows: poster, title, year, type, and status badge (see below).

### Status badge

Each result displays one of:

| Badge | Meaning |
|---|---|
| Available | Item exists in the library (`library.exists` check passed) |
| Requested | A pending request exists for this item |
| Request | Item is not available and not requested |

### Requesting

1. User clicks a result marked "Request".
2. A detail modal opens with: poster, title, overview, genres, year.
3. User confirms the request.
4. Mortar routes the request to the correct plugin based on media type.
5. Modal updates to show "Requested" with current status.

### Request status tracking

- Users can view their own request history.
- Admins can view all requests across all users.
- Status updates are pulled from the upstream plugin (Jellyseerr, AudioBookRequest, etc.) on demand — Mortar does not own request state.

## Acceptance criteria

- [ ] Search returns results within 3 seconds under normal network conditions.
- [ ] Results from multiple plugins are deduplicated by external ID.
- [ ] Media type routing is correct: video → first `requests.video` plugin, audio → first `requests.audio` plugin, ebook → first `requests.ebook` plugin.
- [ ] Items already in the library show "Available" without a request option.
- [ ] A user cannot submit a duplicate request for an already-pending item.
- [ ] Admins see a request management view with approve/deny actions (proxied to upstream plugin).

## Plugin dependencies

| Plugin type | Required capability |
|---|---|
| Jellyseerr | `requests.video` |
| AudioBookRequest | `requests.audio` |
| Shelfarr (optional) | `requests.ebook` |
| Jellyfin | `library.exists` (for status badges) |

## Open questions

- Should search also return results from the library (items already available) for discoverability?
- How do we handle search when one plugin times out — show partial results or surface an error?
- Should regular users see other users' pending requests, or only their own?
