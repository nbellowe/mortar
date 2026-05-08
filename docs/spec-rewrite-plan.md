# Mortar Spec Rewrite Plan

## Goal

Turn the current spec set into an implementation-ready plan for the Go backend and Expo frontend without forcing product or architecture decisions to be made ad hoc in code.

## Why this rewrite is needed

The current spec set has a strong product direction, but it still has three readiness gaps:

- Some accepted behaviors are not expressible in the plugin contract.
- Some cross-cutting architecture decisions still live as open questions instead of reviewable decisions.
- Some specs still reflect the earlier web-only mental model rather than the current Go + Expo stack.

## Rewrite principles

- Product specs define user-visible behavior, permissions, and outcomes.
- ADRs define cross-cutting technical decisions and tradeoffs.
- The plugin interface spec is the contract layer between product behavior and implementation.
- No accepted feature spec should contain unresolved open questions.
- Acceptance criteria should be testable and traceable to a contract or API surface.

## Workstreams

### 1. Spec governance and lifecycle

Purpose: make it obvious which specs are ready, blocked, or proposed.

Outputs:

- Add `Status` and `Depends on` metadata to every spec, with optional `Last updated` when useful.
- Distinguish between feature-level open questions and architecture-level ADRs.
- Adopt a simple status model: `proposed`, `accepted`, `blocked`, `implemented`.

### 2. Plugin contract v2

Purpose: make the contract capable of expressing the behaviors already promised by the product specs.

Outputs:

- Request review actions for admin approve / decline flows
- Library match resolution instead of boolean-only existence checks
- Optional continue-watching capability for personalized home screen rows
- Activity visibility and actor metadata for privacy-safe feeds
- Base health contract clarified as mandatory for all plugins
- Failure semantics documented for partial-degradation behavior

### 3. Architecture decision closure

Purpose: move unresolved platform decisions into reviewable ADRs with explicit defaults.

Outputs:

- ADR 0002: persistence and state model
- ADR 0003: real-time delivery model
- ADR 0004: plugin response caching
- ADR 0005: upstream user identity linking
- ADR 0006: request routing policy

### 4. Feature spec rewrites

Purpose: align feature specs with the new contract and accepted ADRs.

Priority order:

1. `requests.md`
2. `browse-play.md`
3. `activity-feed.md`
4. `download-queue.md`
5. `health.md`
6. New `home.md` spec for cross-cutting home-screen behavior

Required rewrites:

- Requests: duplicate prevention, timeout behavior, partial search failure UX, routing policy references
- Browse & Play: platform handoff behavior, account-link requirements, continue-watching source
- Activity Feed: visibility rules, actor metadata, default time window, partial-source failure UX
- Download Queue: audience scope, refresh model, simplified versus admin view contract
- Health: cached base health behavior and badge visibility
- Home: Recently Added, Continue Watching, and health badge requirements in one place

### 5. Traceability and testability

Purpose: make specs directly useful to future implementation and review work.

Outputs:

- Acceptance criteria that use concrete budgets and fallback behavior
- Contract test fixtures per capability
- Mapping from acceptance criterion -> contract surface -> automated test

## Recommended sequence

1. Review and accept or revise ADRs 0002-0006.
2. Freeze plugin interface v2.
3. Rewrite affected feature specs against the accepted ADRs.
4. Split home-screen behavior into its own spec.
5. Add contract-test and fixture planning before implementation starts.

## Done when

- Every spec has explicit status and dependencies.
- No accepted spec has unresolved open questions.
- The plugin interface can express every accepted feature behavior.
- The architecture spec matches the chosen Go + Expo deployment model.
- Each acceptance criterion is testable without inventing missing behavior during implementation.
