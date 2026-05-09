# Mortar Google Stitch Brief

This file is a paste-ready product brief for Google Stitch.

Use it together with [DESIGN.md](DESIGN.md), which remains the canonical design contract for implementation.

## Prompt

```text
Design a responsive web app for a product called Mortar.

Mortar is the front door for a household homelab media stack. It unifies the daily-use workflows from multiple self-hosted backend services into one calm, friendly interface so non-technical household members can search, request, browse, and watch media without needing to know what Jellyfin, Sonarr, Radarr, SABnzbd, or other service names mean.

This is not an operator dashboard, not an iframe wrapper, and not a replacement for the underlying services. Mortar owns the everyday 80% flows and links out to native apps for complex or service-specific tasks.

Primary audiences:
- Household users: spouse, kids, roommates, guests. They want to request media, browse what is available, and keep up with what is new.
- Homelab owner / admin: configures the stack and can access operational views like service health and a more detailed download queue.

Product posture:
- Warm, calm, welcoming, media-forward
- Clear and confident rather than clinical
- Easy for non-technical users
- Operationally legible for admins without making the whole product feel like infrastructure software

Platform and release posture:
- Design for the supported v1 web client first
- Make it responsive so the same information architecture could later map to native clients
- Avoid hover-only interactions and dense desktop-only dashboard patterns

Design goals:
- Make Search a primary action
- Make Home feel like a useful starting point, not a wall of widgets
- Make library and media browsing feel rich and poster-led
- Make admin states understandable without overwhelming household users
- Keep the whole app in one consistent visual language

Core app flows Mortar owns:
1. Search & request
2. Activity feed
3. Download queue
4. Browse & play
5. Service health

Suggested information architecture:
- Home
- Search
- Library
- Activity
- Downloads
- Health

Role rules:
- Household users should mostly see plain-language actions and statuses: Request, Watch, Browse, Downloading
- Hide plugin and service names by default in household-facing surfaces unless they are necessary to explain an error or diagnostic state
- Admin views may show plugin names, source attribution, URLs, latency, and richer operational detail
- Health detail is admin-only
- Downloads may have a simplified household view and a richer admin view

Design the following screens and flows:

1. App shell and navigation
- Show a responsive navigation pattern for web
- Make Search feel central and obvious
- Accommodate a subtle but visible stack-health badge when the system is degraded

2. Home
- Home is an at-a-glance starting point
- It may show up to three surfaces:
  - Continue Watching
  - Recently Added
  - Service health badge
- Continue Watching is personalized and may require a linked upstream account
- Recently Added should show the 10 most recent additions as horizontal poster-led cards
- One failing surface must not blank the whole screen
- Design for partial loading, empty states, and linked-account-required states

3. Search & Request
- Global search fans out across request-capable services and returns normalized results
- Results are grouped by media type: Movies, Shows, Audiobooks, Ebooks
- Each result should clearly show poster, title, year, media type, and one of these exact states:
  - Available
  - Requested
  - Request
- If an item is requestable, clicking it opens a request-confirmation modal with poster, title, overview, genres, and year
- After confirmation, the modal should update to a requested state
- Search must gracefully support partial service failure with a non-blocking notice naming the unavailable source
- Items already available in the library should feel like a natural path into Watch / library handoff

4. Library / Browse & Play
- Browsing is for content already available in the library
- Use a poster grid with filters for type and genre and sorting for recently added, title, and year
- The library should feel browseable and rich, not like a spreadsheet
- Clicking an available item should hand off directly to the Jellyfin web UI; do not design an in-app player
- If the required linked account is missing, show a clear linked-account-required state instead of anonymous or guest access

5. Activity Feed
- A single merged timeline of activity across services
- Prioritize meaningful household events like newly added items and completed downloads
- Show event type, media item, timestamp, and subtle source attribution
- Support filters for event type and media type
- Admins can also filter by user and source plugin
- Design this as a readable timeline, not a raw log viewer
- Support partial service outage notices without collapsing the feed

6. Downloads
- Unified queue across download services
- Admin view should show name, progress, size, speed, ETA, status, and source plugin
- Household view should answer "is it downloading?" with minimal noise: name, progress, ETA
- Use these exact status labels:
  - Downloading
  - Queued
  - Paused
  - Processing
  - Failed
- Failed items should be visually distinct and clearly link out to the native app for details
- Include a friendly empty state

7. Service Health
- Admin-only view
- Show overall stack summary plus per-plugin cards
- Each card should show name, type, URL, status, latency, and last checked time
- Use these exact health states:
  - Healthy
  - Degraded
  - Unreachable
  - Unknown
- Unhealthy states must be unmistakable, but avoid making the whole screen a wall of identical status tiles
- The page should use last-known status, not a blocking live probe

8. Important states to include across the system
- Linked account required
- Partial service outage / some plugins unavailable
- No results
- Empty queue
- Unknown health state
- Failed request or failed download
- Duplicate request prevented
- Loading states that work per-surface, not just full-screen

Interaction rules and product constraints:
- Mortar never exposes service API keys or other credentials
- Mortar never plays media inline; it hands off to Jellyfin web
- Mortar does not provide an approve / deny request UI; admins link out to the upstream service
- Mortar does not include admin configuration UI in v1
- Mortar should not force users to know backend product names for normal usage
- Do not invent extra product areas like recommendations, social features, notifications, chat, watch parties, or a settings-heavy control center
- Do not design native deep-link patterns as the primary playback flow; browser handoff is the current v1 reality

Visual direction:
- Think "comfortable media room with good lighting, tactile surfaces, clear signage"
- Avoid dark cyberpunk dashboards
- Avoid bright purple SaaS gradients
- Avoid terminal/admin-console aesthetics
- Avoid overdone glassmorphism
- Prefer warm neutral backgrounds, earthy accents, rounded touch-friendly surfaces, strong poster imagery, and restrained motion

Use this design system as the starting point:
- Background canvas: #F3EEE7
- Surface: #FBF7F1
- Elevated surface: #FFFFFF
- Primary text: #1F1A17
- Muted text: #6C625A
- Subtle border: #D8CEC3
- Brand primary: #C35A2C
- Brand strong: #8F3E1D
- Info: #356D8C
- Success: #2E7A57
- Warning: #A56A12
- Danger: #B13C2E
- Primary typeface: Atkinson Hyperlegible Next
- Optional display accent: Space Grotesk
- Base spacing unit: 4
- Corner radii should feel soft and welcoming rather than sharp

Layout and component guidance:
- Favor strong vertical flow over overly dense grids, especially on smaller screens
- Keep one clear focal area per screen
- Use poster-led media cards
- Use distinct but compatible visual treatments for media cards vs operational cards
- One primary action per context
- Status should never rely on color alone
- Minimum touch target should feel mobile-friendly even on web

Accessibility and tone:
- Prioritize readability for non-technical users at arm's length
- Maintain accessible contrast
- Pair all status colors with text labels and, where helpful, icons
- Use warm, plain-language empty states and guidance copy

What I want back:
- A coherent app shell and navigation concept
- High-fidelity concepts for Home, Search & Request, Library, Activity, Downloads, and Health
- Clear role-aware behavior for household users vs admins
- Empty, loading, degraded, and linked-account-required states
- A small reusable component language that could realistically become a design system
- A direction that feels intentional and memorable, but still practical to build in Expo / React Native Web
```

## Source coverage

This brief consolidates:

- [Vision](specs/vision.md)
- [Architecture](specs/architecture.md)
- [Plugin Interface](specs/plugins/plugin-interface.md)
- [Home](specs/features/home.md)
- [Search & Request](specs/features/requests.md)
- [Browse & Play](specs/features/browse-play.md)
- [Activity Feed](specs/features/activity-feed.md)
- [Download Queue](specs/features/download-queue.md)
- [Service Health](specs/features/health.md)
- [Design Contract](DESIGN.md)
