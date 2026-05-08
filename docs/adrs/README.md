# Architecture Decision Records

This directory holds Mortar's cross-cutting technical decisions.

## Status values

- `Proposed` - drafted and ready for review
- `Accepted` - chosen and expected to guide implementation
- `Superseded` - replaced by a later ADR
- `Rejected` - considered and intentionally not adopted

## Current ADRs

| ID | Title | Status |
|---|---|---|
| [0001](0001-tech-stack.md) | Tech Stack | Accepted |
| [0002](0002-persistence-and-state.md) | Persistence and State Model | Proposed |
| [0003](0003-realtime-updates.md) | Real-Time Update Delivery | Proposed |
| [0004](0004-plugin-response-caching.md) | Plugin Response Caching | Proposed |
| [0005](0005-upstream-user-identity-linking.md) | Upstream User Identity Linking | Proposed |
| [0006](0006-request-routing-policy.md) | Request Routing Policy | Proposed |

## Usage

- Feature specs should reference ADR ids in their `Depends on` metadata.
- Architecture specs should summarize accepted decisions, not duplicate full ADR reasoning.
- If a decision changes, create a new ADR instead of rewriting history in place.
