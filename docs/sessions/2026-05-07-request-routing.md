# Request Routing Decision Session — 2026-05-07

## Why this session happened

The Search & Request flow needed a routing rule that was simple for common installs but still predictable once multiple request backends existed.

The project already allows multiple plugins of the same type, so relying on declaration order would have made request submission fragile and hard to reason about.

## Decisions made

### Core routing rule

- If exactly one plugin exposes a request capability, Mortar may route automatically
- If more than one plugin exposes the same request capability, explicit routing config is required
- If no plugin exposes a request capability, that request type is unavailable rather than fatal

### Config shape

- Routing lives in a top-level `routing` section
- `v1` routing is by request capability only:
  - `routing.requests.video`
  - `routing.requests.audio`
  - `routing.requests.ebook`

### Validation behavior

- Invalid routing references are startup errors
- Ambiguous multi-plugin request routing with no explicit config is a startup error
- Mortar should not guess based on declaration order

### Change application

- Routing is read and validated on startup
- Routing changes take effect on restart

## Why this direction won

- It keeps simple installs simple
- It prevents hidden behavior from config ordering
- It keeps capability declaration separate from routing policy
- It leaves room for richer routing later without forcing that complexity into `v1`

## Immediate follow-up chosen from this session

- Mark ADR 0006 as accepted
- Update architecture and request specs to remove declaration-order assumptions
