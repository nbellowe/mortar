# Tech Stack Decision — 2026-05-07

## Decision

**Backend: Go**  
**Frontend: Expo (React Native, TypeScript)**

## Constraints that drove this

- Must deploy to web, iOS, Android, macOS, Windows via native app stores
- TV platforms (Apple TV, Android TV, Fire TV, Samsung Tizen, LG webOS) must be reachable post-v1 without a rewrite
- One codebase required; platform-specific UI code is acceptable
- Roku explicitly excluded (requires BrightScript — no shared codebase path exists)

## Why Expo

Expo is the only framework that satisfies all constraints simultaneously:

- **iOS, Android**: Native compilation via Expo/React Native — App Store and Google Play ✓
- **macOS**: react-native-macos (Microsoft-backed) — Mac App Store ✓
- **Windows**: react-native-windows (Microsoft-backed) — Microsoft Store ✓
- **Web**: React Native Web / Expo web target ✓
- **TV (post-v1)**: react-native-tvos supports Apple TV, Android TV, Fire TV with community support for Tizen and webOS

Platform-specific UI uses Expo's file extension convention (`.ios.tsx`, `.android.tsx`, `.web.tsx`) or `Platform.select()`. Business logic — API calls, state, types — is always shared.

### Flutter considered and rejected

Flutter is the main alternative. It has a slightly stronger unified codebase story and LG is officially investing in Flutter for webOS. It was rejected because:

- **Flutter Web uses canvas rendering, not DOM.** Text is not selectable. Safari has compatibility issues. SEO is broken. Since web is v1 for Mortar and the primary access point for household users, this is disqualifying.
- React Native TV is more mature today (react-native-tvos is in production on tvOS, Android TV, and Fire TV). Flutter's tvOS support is experimental.

If Flutter Web matures to DOM-based rendering, it would become worth reconsidering for the unified codebase benefit.

## Why Go for the backend

- The Mortar server is a long-running API proxy: it receives requests from clients, fans out to multiple upstream services concurrently, normalizes responses, and returns them. This is Go's ideal workload.
- Go's interface model maps cleanly to the plugin capability pattern — each plugin is a Go struct implementing one or more interfaces.
- Produces a single binary with no runtime dependency. Docker image is small.
- Since the frontend is Expo (not Next.js), there is no SSR benefit to a JavaScript backend. The only reason to use Next.js as a backend would be if it were also rendering the web client server-side — it isn't.

### Next.js (full-stack) considered and rejected

Next.js would have been appropriate if the web client were a Next.js app. Once Expo was chosen as the frontend framework, Next.js provides no backend advantage over Go, and Go is a better fit for the plugin runtime.

## Implications for the plugin interface spec

The plugin interface spec (`specs/plugins/plugin-interface.md`) is written in TypeScript. The canonical implementation is now Go. The TypeScript interfaces remain valid as a description of the contract; agents implementing plugins write Go code that satisfies those contracts. TypeScript types in `src/types/` mirror the Go types for use by the Expo frontend.
