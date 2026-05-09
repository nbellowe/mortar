# 2026-05-08: Frontend UI scaffold — dark theme, navigation, full-stack test

## What we decided

- Adopted the "Digital Hearth" dark colour system from Stitch mockup output as the canonical token set. Extracted it into `src/frontend/theme/tokens.ts` using Material Design 3 naming (`surfaceContainer`, `onPrimaryFixed`, etc.).
- Implemented a responsive navigation shell: sidebar (≥768 px, web only) + bottom tabs (mobile). Desktop hides tabs via `tabBar={() => null}` on the `<Tabs>` navigator prop rather than `tabBarStyle: { display: 'none' }`.
- Made the Requests screen a first-class navigation destination (sidebar + mobile tab bar). It was registered as a route but unreachable through the UI.
- Restyled all five data screens (Home, Search, Health, Requests, Library) and added stub screens for Activity and Downloads.
- Fixed a type mismatch: Go `*string` marshals as JSON `null`; changed `PluginHealth.detail` from `string | undefined` to `string | null`.

## Key constraints or context

- `<Link asChild>` in expo-router clones the child and merges its own `style` prop. If the child already has a style array, the result is a nested array that expo-router's `<Slot>` rejects at runtime. Fix: `StyleSheet.flatten` before passing the style.
- `tabBarStyle: { display: 'none' }` causes React Navigation to merge our style object with its own defaults into an array, which also triggers the Slot rejection. Fix: use the `tabBar` navigator-level prop instead.
- Backend (Go, port 3000) and frontend (Expo web, port 8081) are separate processes. CORS is configured with `AllowedOrigins: ["*"]` so cross-origin fetch works in development.
- The backend has real plugin instances configured (Jellyfin, Jellyseerr, Sonarr, Radarr, SABnzbd) all reachable locally. All five were healthy at the time of testing.

## What we explicitly ruled out

- Interacting with Stitch before implementing (the user asked about it mid-session; we agreed to proceed with the existing mockup output since 4 screens were already done).
- Putting Activity and Downloads in the mobile tab bar — both are secondary views with no live data yet; sidebar-only access is acceptable for now.
- Adding a Library tab item on mobile for the scaffold — swapped it for Requests since Library is a stub and Requests has live data.

## Open questions it raised

- DESIGN.md still specifies a light theme (`#F3EEE7` background, `#C35A2C` primary). It now conflicts with the implemented dark tokens. DESIGN.md should be updated to reflect the accepted dark palette, or a formal dark/light toggle decision should be made.
- The Stitch product brief (`STITCH.md`) was not used in this session. If the user wants to iterate on designs interactively before the next feature, the `/stitch` command should be the entry point.
- `formatCheckedAt` returns `NaNh ago` for empty `checked_at` strings (only occurs if a plugin's `Health()` returns a non-nil error, which current plugins never do). Low priority but worth hardening.
