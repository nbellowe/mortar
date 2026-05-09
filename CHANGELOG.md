# Changelog

All notable changes to Mortar will be documented in this file.

## Unreleased

### Added

- `external_url` plugin config field for Jellyfin and Jellyseerr — used for poster images, play deep-links, and admin review links when internal and external hostnames differ
- Activity and Downloads screens added to mobile tab bar (previously unreachable on narrow screens)
- Role-based download queue filtering: regular users see name/progress/ETA only; admins see full detail including size, speed, and source plugin

### Fixed

- Jellyfin library screen (502) and "Recently Added" (empty) when `username` not set in Jellyfin plugin config
- Jellyfin poster images and play URLs used internal cluster hostname, causing browser failures
- Jellyseerr "Review upstream" admin links pointed to internal cluster URL
- Sonarr and Radarr activity events showed empty item titles (missing `includeSeries`/`includeMovie` query params)
- Nil pointer dereference in `handleLibraryPlay` and `handleSubmitRequest` if session context is empty despite `requireAuth`
- Login errors silently swallowed in `AuthContext`; error message now surfaces to login screen
- `AbortController` signal not forwarded to `getActivity()` and `getDownloads()`, preventing request cancellation on unmount
- Activity feed plugin goroutines could hang indefinitely; 10-second fan-out timeout added
- Optimistic request update used a fake constructed ID; now uses the `Request` object returned by the server
- Request button could be tapped multiple times in parallel; now disabled (spinner shown) while in-flight
- Failed/declined requests shown as "Request" with no indication of prior failure; now labeled "Retry"
- Duplicate "Sign out" button on home screen when sidebar is visible on desktop

- Initial OSS project scaffolding for license, roadmap, and documentation structure

## Changelog conventions

Release entries should use the sections that apply:

- `Breaking`
- `Upgrade Notes`
- `Added`
- `Changed`
- `Fixed`
- `Known Issues`

Optional when relevant:

- `Security`
- `Deprecated`
- `Removed`
