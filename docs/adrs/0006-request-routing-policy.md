# ADR 0006: Request Routing Policy

## Status

Accepted

## Date

2026-05-07

## Context

Mortar routes request submissions by media type, but multiple plugins of the same type are allowed and multiple plugins may expose the same request capability. The current "first declared plugin wins" idea is simple, but it is brittle and easy to misconfigure in real homelabs.

## Decision

Use a top-level per-capability `routing` section for request submission when there is more than one candidate plugin. If exactly one plugin exposes a given request capability, Mortar may route to it automatically.

## Policy

- Routing is by request capability only in `v1`:
  - `routing.requests.video`
  - `routing.requests.audio`
  - `routing.requests.ebook`
- If zero plugins expose a request capability, that request type is unavailable.
- If exactly one plugin exposes a request capability, Mortar routes to it automatically.
- If more than one plugin exposes a request capability, Mortar requires explicit routing config for that capability.
- If routing config references a plugin ID that does not exist, startup must fail with a clear validation error.
- If routing config references a plugin that does not expose the required capability, startup must fail with a clear validation error.
- If more than one compatible plugin exists and no explicit route is configured, startup must fail with a clear validation error.
- Routing is read and validated at startup; config changes take effect on restart.

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
- A top-level routing section keeps capability declaration separate from policy.
- It scales to future use cases like separate family libraries or multiple request backends without forcing those rules into `v1`.

## Consequences

- The config schema must add an optional `routing` section.
- Feature specs should refer to configured routing policy rather than declaration order.
- Startup validation becomes more important, but it catches misconfiguration earlier.
- Operators can omit routing config in the simple single-candidate case.
- Request capability availability in `v1` is determined by plugin capabilities plus validated routing configuration.

## Related

- [Request Routing Decision Session](../sessions/2026-05-07-request-routing.md)
- [Search & Request Spec](../../specs/features/requests.md)
