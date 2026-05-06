# Vision

## Problem

A typical homelab media stack runs 10–20 separate services, each with its own UI, auth, and design language. The homelab owner navigates this fluently. Household members — a spouse, kids, roommates — face a confusing array of apps just to request and watch something.

The ecosystem is intentionally fractured: each tool does one thing well, and they interoperate via APIs. That's a strength for the homelab owner. It's a barrier for everyone else.

## What Mortar is

Mortar is a unified frontend for homelab media stacks. It aggregates the daily-use workflows from multiple backend services into a single, consistent UI.

It is **the front door** — the one app household members need to know about.

## Who it is for

Two audiences:

**Household users** — people who want to request, browse, and track media. They should never need to know what Sonarr, SABnzbd, or Prowlarr are. Mortar speaks their language: search, request, watch.

**The homelab owner** — the person who configures and maintains the stack. They configure which plugins are enabled, manage user accounts, and can see operational views (download queue, service health) that regular users don't need.

## What Mortar is NOT

- Not a replacement for any underlying service. Sonarr, Jellyseerr, SABnzbd, Jellyfin all continue to run and own their domains.
- Not a power-user tool. Complex configuration (quality profiles, indexer management, custom formats) belongs in the native UIs.
- Not an iframe wrapper. Mortar builds its own UI against service APIs — consistent design, not embedded windows.
- Not a self-contained media server. It has no opinion on storage, transcoding, or download logic.

## Core user flows

Mortar owns exactly five flows:

1. **Search & request** — find media, request it, track its status
2. **Activity feed** — see what recently arrived across all services
3. **Download queue** — see what's currently downloading and its progress
4. **Browse & play** — surface library content and hand off to a player
5. **Service health** — confirm services are reachable (owner-facing)

Everything else is a link to the native app.

## Non-goals (v1)

- User-facing notifications / push alerts
- Playback within Mortar (deep link to Jellyfin instead)
- Admin configuration UI (config file is sufficient for v1)
- Mobile app (responsive web is good enough)
- Multi-server / multi-instance support
