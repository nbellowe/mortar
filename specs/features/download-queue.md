# Feature: Download Queue

## Goal

A unified view of everything currently downloading, across all download clients. The homelab owner (and optionally regular users) can see queue status and progress at a glance.

## User flows

### Viewing the queue

1. User opens the downloads view.
2. Mortar fetches the current queue from all plugins with `downloads.read` capability.
3. Items are merged into a single list, sorted by: actively downloading first, then queued, then paused/failed.
4. Each item shows: name, progress bar, size, download speed, ETA, status, and source plugin badge.

### Status indicators

| Status | Meaning |
|---|---|
| Downloading | Actively transferring |
| Queued | Waiting to start |
| Paused | Manually or automatically paused |
| Processing | Post-download (unpacking, moving) |
| Failed | Error — see native app for details |

### What regular users see

Regular users see a simplified view: items in progress with name and ETA. No speed, no size, no source plugin badge. Gives them "is my request downloading?" without operational noise.

Admins see the full view.

## Acceptance criteria

- [ ] Queue refreshes automatically every 10 seconds without full page reload.
- [ ] Items from SABnzbd and qBittorrent are combined in one list.
- [ ] Progress bars update on each refresh cycle.
- [ ] Failed items are visually distinct and include a link to the native app.
- [ ] When the queue is empty, show a friendly empty state (not a blank list).
- [ ] Plugin failures degrade gracefully — show items from available plugins.

## Plugin dependencies

| Plugin type | Required capability |
|---|---|
| SABnzbd | `downloads.read` |
| qBittorrent | `downloads.read` |

## Out of scope (v1)

- Pause, resume, or cancel downloads from Mortar. Link to native app instead.
- Per-download logs or error details.

## Open questions

- Should regular users see the download queue at all, or only admins?
- Is 10-second polling acceptable, or should this use a WebSocket/SSE for real-time updates?
