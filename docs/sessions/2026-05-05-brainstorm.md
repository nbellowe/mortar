# Brainstorm: 2026-05-05

Initial ideation session that produced the Mortar concept.

## Starting point

Nathan runs a mature single-node k3s homelab (Ansible + Argo CD, GitOps) with a full media stack: Jellyfin, Sonarr/Radarr/Lidarr, Prowlarr, SABnzbd/qBittorrent, Jellyseerr, Bazarr, Recyclarr, Audiobookshelf, Calibre + Calibre-Web, Storyteller, Home Assistant, and Ollama with Nvidia GPU passthrough.

He wanted to brainstorm a fun upstream open-source contribution to the homelab/media server community.

## Ideas considered

### 1. Declarative stack wiring tool

A CLI/container that reads a config declaring service URLs and desired inter-service connections, then idempotently wires them together via each app's API (e.g. Prowlarr → Sonarr, Sonarr → SABnzbd).

**Research finding:** Configarr (592 stars, active) covers quality profiles and custom formats but explicitly does NOT handle inter-service wiring. GitHub issue #264 is an open request for this. The gap is real.

**Why we moved on:** This is more of an infrastructure/devops tool. The unified UI idea was more exciting.

### 2. Book/audiobook request portal

An Overseerr-equivalent for ebooks and audiobooks.

**Research finding:** AudioBookRequest (645 stars, active) covers audiobooks via Prowlarr. Shelfarr (131 stars, early) is attempting a unified books + audiobooks portal. Libreseerr exists but is low-activity. The space is more crowded than expected.

**Why we moved on:** First-mover advantage is gone. Contributing to Shelfarr might be better than building from scratch.

### 3. Unified activity feed

A webhook aggregator that receives events from all *arr apps and presents a normalized timeline.

**Status:** Good idea, relatively low effort, but felt like a smaller project. Kept as a feature within Mortar rather than a standalone project.

### 4. Mortar (chosen direction)

A unified frontend for the whole homelab media stack — one UI, consistent design, for household members who don't know what Sonarr is.

## Key design decisions

### Plugin-first, not *arr-first

Mortar is not an *arr tool. Every integration is a plugin with standard capability interfaces. This keeps it ecosystem-agnostic and extensible.

### Build on top of Seerr, not around it

Initially considered going directly to Sonarr/Radarr for requests. Decision: use Jellyseerr's API for video requests instead. Jellyseerr handles approval logic, user quotas, "already exists" detection, and notifications. Mortar doesn't need to rewrite that.

This also meant the architecture is naturally multi-backend for requests: Jellyseerr for video, AudioBookRequest for audiobooks, Shelfarr for ebooks — each a plugin.

### Front door, not replacement

Mortar is not a replacement for any service. Complex workflows stay in native apps. Mortar owns exactly five user flows: search/request, activity feed, download queue, browse/play, health.

### Target audience: household users, not power users

The differentiator from Homarr/Homepage (which are link aggregators for the homelab owner) is that Mortar targets household members who want to request and watch media without understanding the underlying stack.

## What Mortar is NOT

- Not an iframe wrapper (Organizr tried this, bad UX)
- Not a replacement for Sonarr, Radarr, Jellyfin, etc.
- Not a power-user tool for complex configuration
- Not limited to the *arr ecosystem (hence the name "Mortar" not "Somethingarr")

## Name

"Mortar" — the material that holds bricks together. Chosen because:
- It's the connective layer between services
- Not tied to any ecosystem naming convention
- Short, memorable, one word, easily Googleable
- Evokes binding/aggregation without being prescriptive

Rejected names: Stitch, Rivet, Weave (HashiCorp conflict), Plumbr, Linkarr.

## Open questions at close of session

- Tech stack: TypeScript/Next.js full-stack vs. Go backend + React frontend
- Database: SQLite for user accounts and request cache?
- Real-time updates: WebSocket/SSE vs. polling for downloads and activity
- Caching strategy for upstream service responses
