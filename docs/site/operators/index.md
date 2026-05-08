# Getting started

Mortar is a lightweight server you run alongside your existing homelab services. It connects to them via their APIs and exposes a single web app to your household.

## What Mortar is not

- Not a replacement for Jellyfin, Jellyseerr, Sonarr, etc.
- Not a media server or download client
- Not a mobile app (v1.0 is web-only; native builds are experimental)

## Prerequisites

- A machine to run the Mortar server (any Linux/macOS/Windows host, or a container)
- At least one supported upstream service already running ([see Plugins](./plugins))
- Go 1.21+ **or** Docker, for running the server

## Reference stack

Mortar v1.0 is designed around:

| Service | Purpose |
|---|---|
| Jellyfin | Library browsing, playback handoff, continue watching |
| Jellyseerr | Movie and TV requests |
| Sonarr | TV download activity |
| Radarr | Movie download activity |
| SABnzbd | NZB download queue |

You don't need all of them. Mortar degrades gracefully — if a service isn't configured, the features that depend on it simply don't appear.

## Next steps

1. [Install the server](./installation)
2. [Write your config file](./configuration)
3. [Configure your plugins](./plugins)
