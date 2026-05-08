# Feature: Activity Feed

## Metadata

- **Status:** `accepted`
- **Depends on:** [Plugin Interface](../plugins/plugin-interface.md), [ADR 0002](../../docs/adrs/0002-persistence-and-state.md), [ADR 0003](../../docs/adrs/0003-realtime-updates.md), [ADR 0004](../../docs/adrs/0004-plugin-response-caching.md), [ADR 0005](../../docs/adrs/0005-upstream-user-identity-linking.md)
- **Last updated:** `2026-05-07`

## Goal

A single chronological feed showing what has happened across the entire media stack — downloads completed, items added to the library, requests made and fulfilled. Household members can see what's new without checking each service individually.

## User flows

### Viewing the feed

1. User opens the activity view.
2. Mortar fetches recent events from all plugins with `activity.read` capability.
3. Events are merged and sorted by timestamp descending.
4. Each event shows: source plugin icon, event type label, media item (with poster thumbnail), timestamp (relative: "2 hours ago").
5. While the view is active, Mortar refreshes the feed on a 30-second polling interval and may use incremental fetches where supported.

### Event types shown in feed

| Event | Source plugins | Shown to |
|---|---|---|
| Item added to library | Sonarr, Radarr | All users |
| Download completed | SABnzbd, qBittorrent | All users |
| Request submitted | Jellyseerr, AudioBookRequest | Admin + requester |
| Request approved | Jellyseerr, AudioBookRequest | Admin + requester |
| Request declined | Jellyseerr, AudioBookRequest | Admin + requester |
| Request failed | Jellyseerr, AudioBookRequest | Admin + requester |
| Item deleted | Sonarr, Radarr | Admin only |

### Filtering

Users can filter the feed by:
- Event type (added, requested, downloaded, etc.)
- Media type (movies, shows, audiobooks, etc.)

Admins additionally filter by:
- User (who triggered the event)
- Source plugin

## Acceptance criteria

- [ ] Feed loads within 2 seconds for the last 7 days of activity. The lookback window is fixed at 7 days and is not user-configurable.
- [ ] Events from all connected `activity.read` plugins are merged in correct timestamp order.
- [ ] While the view is active, the feed refreshes automatically on the shared polling model without requiring a page reload.
- [ ] Relative timestamps update without page reload.
- [ ] Filter controls narrow the feed immediately (client-side where possible).
- [ ] Regular users do not see request events from other users.
- [ ] Plugin failures (a service is down) degrade gracefully — show events from available plugins with a notice about unavailable ones.

## Plugin dependencies

| Plugin type | Required capability |
|---|---|
| Sonarr | `activity.read` |
| Radarr | `activity.read` |
| SABnzbd | `activity.read` |
| qBittorrent | `activity.read` |
| Jellyseerr | `activity.read` |
| AudioBookRequest | `activity.read` |

