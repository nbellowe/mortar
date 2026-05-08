# ADR 0006: Request Routing Policy

## Status

Proposed

## Date

2026-05-07

## Context

Mortar routes request submissions by media type, but multiple plugins of the same type are allowed and multiple plugins may expose the same request capability. The current "first declared plugin wins" idea is simple, but it is brittle and easy to misconfigure in real homelabs.

## Decision

Use explicit per-capability routing when there is more than one candidate plugin. If exactly one plugin exposes a given request capability, Mortar may route to it automatically.

## Policy

- Zero compatible plugins for a request capability: startup error
- One compatible plugin: route automatically
- More than one compatible plugin: require explicit config

## Example

```yaml
plugins:
  - id: jellyseerr
    type: jellyseerr
    url: http://jellyseerr:5055

  - id: audiobookrequest
    type: audiobookrequest
    url: http://audiobookrequest:8080

routing:
  requests:
    video: jellyseerr
    audio: audiobookrequest
```

## Rationale

- Explicit routing removes accidental behavior from config ordering.
- It keeps the simple case simple while making the ambiguous case visible and reviewable.
- It scales to future use cases like separate family libraries or multiple request backends.

## Consequences

- The config schema must add an optional `routing` section.
- Feature specs should refer to configured routing policy rather than declaration order.
- Startup validation becomes more important, but it catches misconfiguration earlier.
