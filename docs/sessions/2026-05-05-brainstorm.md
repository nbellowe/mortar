# Mortar: Initial Design Session — 2026-05-05

## What Mortar is

A unified frontend for homelab media stacks. One UI, consistent design, for household members who don't know what Sonarr is. Mortar is the front door — it connects to existing services via their APIs and presents the daily-use workflows in one place, without replacing anything.

## Problem it solves

A typical homelab media stack runs 10–20 services, each with its own UI, auth, and design language. The homelab owner navigates this fluently. Everyone else is lost. No existing tool addresses this: Homarr and Homepage are link aggregators for the owner, not usable interfaces for the household.

## Core design decisions

### Plugin-first, not ecosystem-first

Every integration is a plugin with a standard capability interface. Mortar is not an *arr tool and is not built on the Servarr ecosystem. The *arr apps (Sonarr, Radarr, Lidarr) are important plugins but not privileged ones. This was a deliberate choice:

- Servarr apps are backends, not a framework — there is nothing to actually build on
- Mortar's stack includes Jellyfin, Audiobookshelf, SABnzbd, and Jellyseerr, none of which are Servarr projects
- Centering on Servarr would signal the wrong audience and create coupling without leverage
- The plugin model works whether you use the full *arr stack, a partial one, or none at all

### Build on top of Jellyseerr for video requests, not around it

Considered going directly to Sonarr/Radarr for the request flow. Decision: use Jellyseerr's API instead. Jellyseerr already handles approval logic, user quotas, "already exists" detection, and notifications. Mortar does not rewrite that.

This naturally extends to other request types: AudioBookRequest for audiobooks, Shelfarr for ebooks. Each is a plugin. Jellyseerr is not special — it is the first `requests.video` plugin.

### Front door, not replacement

Mortar owns exactly five user flows: search/request, activity feed, download queue, browse/play, health. Everything else is a link to the native app. Complex operations (quality profiles, indexer management, download client config) stay in Sonarr, Radarr, SABnzbd, etc.

### Target audience: household users, not the homelab owner

The differentiator from Homarr/Homepage is audience. Mortar is built for a spouse, kids, or roommates who want to request and watch media — not for the person who built the stack. Non-technical users should never need to know what Sonarr is.

## What Mortar is NOT

- Not an iframe wrapper — Organizr proved this approach produces bad UX
- Not a replacement for any underlying service
- Not a power-user configuration tool
- Not tied to the *arr ecosystem (hence "Mortar", not "Somethingarr")
- Not a media server — no transcoding, no storage opinions

## Name

"Mortar" — the material that holds bricks together. The connective layer between services. Short, memorable, not tied to any ecosystem naming convention.

Rejected: Stitch, Rivet, Weave (HashiCorp conflict), Plumbr, Linkarr.

## Competitive landscape

| Tool | What it does | Why Mortar is different |
|---|---|---|
| Homarr / Homepage | Link aggregator with status widgets | Owner-facing, no interactive workflows |
| Organizr | iframe tab wrapper | Inconsistent UX, auth problems, mobile broken |
| Jellyseerr | Request portal for video only | Single-service, not a unified frontend |
| AudioBookRequest | Request portal for audiobooks only | Single-service |
| Overseerr | Request portal for video only | No book support, not planned |

## Open questions

- **Tech stack:** TypeScript/Next.js full-stack vs. Go backend + React frontend
- **Database:** SQLite for user accounts and request status cache?
- **Real-time updates:** WebSocket/SSE vs. polling for downloads and activity feed
- **Caching:** How aggressively to cache upstream plugin responses
- **Audiobookshelf in browse view:** Separate section or merged with Jellyfin library?
- **Guest playback:** Can non-Jellyfin users get a play link, or do they need a Jellyfin account?
