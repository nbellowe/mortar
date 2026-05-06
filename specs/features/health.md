# Feature: Service Health

## Goal

The homelab owner can see at a glance whether all connected services are reachable, without opening each app individually.

## User flows

### Health dashboard

1. Admin opens the health view.
2. Mortar pings all configured plugins via their `health()` method.
3. Each plugin is shown as a card: name, type, URL, status indicator, latency, last checked timestamp.
4. Overall stack health is summarized at the top: "All services healthy" or "X of Y services unreachable."

### Status indicators

| Status | Meaning |
|---|---|
| Healthy | Reachable and responded within threshold |
| Degraded | Reachable but slow (latency > 2s) |
| Unreachable | Connection failed or timed out |
| Unknown | Not yet checked since startup |

### Home screen badge

The home screen navigation shows a badge if any service is currently unreachable. Clicking it goes to the health view. Regular users see the badge but cannot access the health detail view.

## Acceptance criteria

- [ ] All plugins are checked on a configurable interval (default: 60 seconds).
- [ ] Health check results are cached — the health view shows last-known state, not a live probe on page load.
- [ ] Latency is displayed in ms.
- [ ] Unreachable services are visually prominent (red indicator, not just a text label).
- [ ] The health view is accessible to admins only.
- [ ] Health check failures do not affect other Mortar functionality.

## Plugin dependencies

All plugins must implement `health.ping`. It is the only mandatory capability.

## Out of scope (v1)

- Alerting or notifications when a service goes down.
- Historical uptime graphs.
- Auto-restart or remediation actions.

## Open questions

- Should the health check interval be configurable per-plugin or globally?
- Should non-admin users see any health information at all, or just the badge?
