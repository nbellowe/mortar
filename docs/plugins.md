# Plugin Compatibility

This file tracks Mortar plugin support status and compatibility expectations.

## Support levels

- **Supported** — documented, compatibility-defined, tested, part of the reference stack, and release-blocking if broken
- **Experimental** — basic docs, minimal tests, known compatibility target, and not release-blocking

## Compatibility table

| Plugin type | Support level | Supported upstream versions | Capabilities | Notes |
|---|---|---|---|---|
| jellyfin | Experimental | 10.8+ | library.browse, library.exists, library.resume | Browse library, check availability, continue watching |
| radarr | Experimental | v3 API | activity.read | Movie download and import activity |
| sabnzbd | Experimental | 3.x | downloads.read | NZB download queue |
| sonarr | Experimental | v4.x | activity.read | TV show download and import activity |

Add one row per plugin type as implementations land.
