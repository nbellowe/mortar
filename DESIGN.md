# Mortar Design Contract

## Purpose

This file is the canonical source of truth for Mortar's frontend design rules.

Google Stitch is the preferred tool for ideation, screen exploration, and flow prototyping. Stitch outputs are proposals. This file defines what becomes implementation truth in the repository.

## Product posture

Mortar is the front door for a household media stack, not an operator dashboard.

The UI should feel:

- warm, calm, and welcoming rather than neon, gamer, or terminal-like
- confident and clear rather than sparse and clinical
- media-forward for household workflows
- operationally legible for admin workflows without turning the whole app into infrastructure software

## Audience rules

### Household views

- Use plain-language actions: `Request`, `Watch`, `Browse`, `Downloading`
- Hide plugin and service names by default unless they are necessary to complete the task
- Prefer posters, titles, and status labels over system terminology

### Admin views

- Plugin names, health details, and source attribution may be visible
- Operational detail should be available, but the layout should still share the same visual system as household views

## Visual direction

Design reference: a comfortable media room in the evening — warm, dim, cinema-like.

The app uses a **dark theme by default**. The palette is warm brown-black (not cool grey or neon), consistent with the "Digital Hearth" color system adopted in the initial frontend implementation.

Avoid:

- cold, blue-tinted dark UIs
- bright purple SaaS gradients
- raw server-admin aesthetics
- overuse of glassmorphism or chrome-heavy cards
- neon or high-saturation cyberpunk accents

Prefer:

- warm dark backgrounds (deep brown-black, not grey)
- earthy, muted accent colors with warm orange-peach primaries
- strong poster imagery
- rounded, touch-friendly surfaces
- restrained motion

## Design tokens

These names are stable even if individual values are tuned later.

### Color

The canonical token values are defined in `src/frontend/theme/tokens.ts`. The table below maps semantic names to their current dark-theme values.

| Token | Value | Use |
|---|---|---|
| `color.bg.canvas` | `#1b110d` | App background |
| `color.bg.surface` | `#281d19` | Cards, sheets, panels |
| `color.bg.elevated` | `#3e322e` | Raised dialogs and focused surfaces |
| `color.fg.default` | `#f3ded8` | Primary text |
| `color.fg.muted` | `#ddc0b6` | Secondary text |
| `color.border.subtle` | `#56423b` | Dividers, borders |
| `color.brand.primary` | `#ffb599` | Primary actions, active states |
| `color.brand.strong` | `#e17141` | Pressed states, emphasis |
| `color.info` | `#74d2f6` | Informational accents |
| `color.success` | `#81c784` | Available, healthy |
| `color.warning` | `#ffb74d` | Queued, degraded |
| `color.danger` | `#e57373` | Failed, unreachable |

### Typography

- Primary UI family: `Atkinson Hyperlegible Next`
- Optional display accent: `Space Grotesk`

Rules:

- Use the primary family for body, navigation, buttons, forms, and status labels
- Use the display accent sparingly for large section headers or featured hero text
- Do not mix more than two font families in the app

Suggested scale:

| Token | Size |
|---|---|
| `type.hero` | `32` |
| `type.title` | `24` |
| `type.section` | `20` |
| `type.body` | `16` |
| `type.bodySmall` | `14` |
| `type.meta` | `12` |

### Spacing

Base spacing unit: `4`

| Token | Value |
|---|---|
| `space.1` | `4` |
| `space.2` | `8` |
| `space.3` | `12` |
| `space.4` | `16` |
| `space.5` | `20` |
| `space.6` | `24` |
| `space.8` | `32` |
| `space.10` | `40` |

### Radius

| Token | Value | Use |
|---|---|---|
| `radius.sm` | `10` | Inputs, badges |
| `radius.md` | `16` | Buttons, cards |
| `radius.lg` | `22` | Sheets, feature panels |
| `radius.round` | `999` | Pills, avatars |

### Motion

- Standard transitions: `160-220ms`
- Emphasized transitions: `240-320ms`
- Favor fade, slide, and scale combinations with ease-out timing
- Motion should explain state changes, not decorate every interaction

## Layout rules

- Favor clear vertical flow over dense dashboard grids on smaller screens
- Use poster-led cards for media-heavy surfaces
- Keep primary actions visible without requiring long explanatory text
- Prefer one strong focal area per screen
- Do not overload the home screen with equal-weight widgets

## Component rules

### Search

- Search is a primary action, not a tiny utility control
- Result rows or cards must clearly communicate title, type, art, and request state
- Status should never rely on color alone

### Cards

- Media cards should prioritize poster art first, metadata second
- Operational cards should prioritize state first, explanation second
- Do not mix multiple visual card styles on the same screen without a clear reason

### Buttons

- One primary action per context
- Secondary actions should be quieter but still obvious
- Destructive actions must use `color.danger`

### Badges and status chips

- Use plain language
- Keep labels short
- Always pair color with text and, where helpful, iconography

### Empty states

- Explain what the user is looking at
- Say what to do next
- Keep the tone warm and matter-of-fact

## Feature guidance

### Home

- Prioritize `Continue Watching`, `Recently Added`, and request entry points
- Health badges and operator cues should stay present but visually secondary for household users

### Search & Request

- Feels fast, forgiving, and obvious
- Request state should be readable at a glance
- Available items should naturally link into library details

### Browse & Play

- Library surfaces should feel browseable and rich, not spreadsheet-like
- Deep-link handoff should feel intentional and platform-native

### Activity Feed

- Feels like a readable timeline, not a raw event log
- Household users should primarily see meaningful arrivals and request outcomes

### Download Queue

- Admin view may expose more operational detail
- Household view should answer "is it downloading?" with minimal noise

### Health

- Make unhealthy states unmistakable
- Avoid turning the screen into a wall of identical status boxes

## Cross-platform rules

- Information architecture should remain consistent across web and native clients
- Layout may adapt by form factor, but terminology and core flows should not change by platform
- Web may use browser tabs for handoff; native clients may open the target app or browser using platform conventions

## Accessibility

- Minimum touch target: `44x44`
- Text and status indicators must meet accessible contrast expectations
- Never rely on color alone to communicate meaning
- Typography must remain readable for non-technical household users at arm's-length viewing distances

## Stitch workflow

When using Google Stitch:

1. Start from the rules in this file
2. Generate screens and flows in Stitch
3. Review the output against Mortar's audience rules and token system
4. Normalize accepted designs back into repo-owned tokens, component rules, and implementation plans

Do not:

- treat raw exported styles as canonical
- commit one-off colors or spacing values without mapping them to tokens
- let Stitch outputs override product language or accessibility rules

## Planned implementation targets

As the app is scaffolded, these rules should be mirrored into code-level artifacts such as:

- `src/theme/tokens.ts`
- `src/theme/typography.ts`
- shared primitives in `src/components/`
- screen-specific layout patterns in `app/`

## Change policy

If a Stitch prototype and this file disagree, update this file first or reject the prototype. The repo-owned design contract wins.
