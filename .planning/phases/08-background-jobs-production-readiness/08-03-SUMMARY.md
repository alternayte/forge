---
phase: 08-background-jobs-production-readiness
plan: 03
subsystem: infra
tags: [sse, notify, postgresql, listen-notify, pgxlisten, backpressure, atomic]

requires:
  - phase: 06-html-layer-datastar-integration
    provides: SSE pattern established (datastar.ServerSentEventGenerator); context-cancelled shutdown model

provides:
  - SSELimiter with global + per-user atomic connection counters and Acquire/release pattern
  - NotifyHub interface (Subscribe/Publish/Start) for swappable pub/sub backends
  - PostgresHub implementation using pgxlisten for single-connection LISTEN/NOTIFY with auto-reconnect
  - Subscription type with Events read-only channel and Close() cancel function
  - Event type with Type field; CloseEvent sentinel for graceful shutdown signalling

affects:
  - 08-04 (SSE HTTP handler wiring)
  - Any phase adding SSE endpoints

tech-stack:
  added:
    - github.com/jackc/pgxlisten v0.0.0-20250802141604-12b92425684c
  patterns:
    - Acquire/release pattern for resource-bounded connections (atomic counters, sync.Map)
    - Single LISTEN connection with in-payload routing (channel + tenantID JSON envelope)
    - Non-blocking fan-out with backpressure (drop + refresh signal on full buffer)
    - Executor interface for pg_notify calls (avoids circular imports with gen/actions)
    - Context cancellation as graceful shutdown mechanism for long-lived connections

key-files:
  created:
    - internal/sse/limiter.go
    - internal/notify/hub.go
    - internal/notify/subscription.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "PostgresHub accepts Executor interface (not *pgxpool.Pool) for Publish — satisfied by pool, conn, or tx without importing generated actions package"
  - "Single forge_events PostgreSQL LISTEN channel with channel+tenantID routing in JSON payload — keeps LISTEN count to 1 regardless of resource count"
  - "bufferSize defaults to 32 when 0 — balances memory use and backpressure resistance"
  - "Reconnect errors that match context.Canceled are not logged — avoids shutdown noise from expected cancellation"
  - "unsubscribe uses swap-with-last removal — O(1) slice removal without preserving order (fan-out order is irrelevant)"

patterns-established:
  - "Pattern: SSELimiter.Acquire returns (release func, error) — release must be called (typically via defer) when the connection ends"
  - "Pattern: Subscription.Close() is the consumer's cancel — named consistently with io.Closer conventions"
  - "Pattern: Non-blocking fan-out: select { case ch <- event: default: send refresh } — subscriber lag never blocks the dispatch goroutine"

requirements-completed:
  - SSE-01
  - SSE-02
  - SSE-03
  - SSE-04
  - SSE-05
  - SSE-06

duration: 2min
completed: 2026-02-19
---

# Phase 8 Plan 03: SSE Connection Limiter and PostgreSQL LISTEN/NOTIFY Hub Summary

**SSELimiter with atomic global/per-user caps and PostgresHub using pgxlisten for single-connection fan-out with non-blocking backpressure**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-02-19T17:20:41Z
- **Completed:** 2026-02-19T17:22:57Z
- **Tasks:** 2
- **Files modified:** 5 (3 created, 2 updated: go.mod, go.sum)

## Accomplishments

- SSELimiter enforces global cap (SSE-01) and per-user cap (SSE-02) using atomic counters with sync.Map for zero-allocation per-user tracking
- PostgresHub implements single shared LISTEN connection (SSE-04) via pgxlisten with 5s auto-reconnect delay
- Non-blocking fan-out with backpressure: full subscriber buffers drop events and send a "refresh" signal (SSE-05)
- Context cancellation propagates shutdown to the LISTEN goroutine — Start() blocks until ctx is cancelled (SSE-03)
- NotifyHub interface with swappable backend (SSE-06) and Executor interface for Publish to avoid circular imports

## Task Commits

1. **Task 1: Create SSE connection limiter** - `798dff1` (feat)
2. **Task 2: Create NotifyHub interface and PostgresHub** - `40b89e9` (feat)

## Files Created/Modified

- `internal/sse/limiter.go` - SSELimiter with Acquire/release, ErrTooManyConnections, ErrTooManyConnectionsForUser, ActiveConnections(), ActiveForUser()
- `internal/notify/subscription.go` - Event type with Type field, CloseEvent sentinel, Subscription with Events channel and Close()
- `internal/notify/hub.go` - NotifyHub interface, Executor interface, PostgresHub implementation, notifyMessage envelope, handleNotification fan-out
- `go.mod` / `go.sum` - Added github.com/jackc/pgxlisten dependency

## Decisions Made

- `PostgresHub.Publish` accepts an `Executor` interface (not a concrete pool type) so callers can pass a pool, a connection, or a transaction without the notify package needing to import generated code.
- Single `forge_events` PostgreSQL NOTIFY channel with channel+tenantID routing inside the JSON payload keeps the LISTEN count to 1 regardless of how many resource types or tenants exist.
- Buffer size defaults to 32 (configurable via `NewPostgresHub` bufferSize arg); aligns with SSEConfig.BufferSize from config.go (added in a sibling plan).
- Reconnect errors that match `context.Canceled` are not logged — avoids spurious warnings during graceful server shutdown.
- `unsubscribe` uses swap-with-last-element for O(1) removal since fan-out order is irrelevant.

## Deviations from Plan

None - plan executed exactly as written.

The plan specified an `Executor` interface design for `Publish`; this was implemented as a separate exported `Executor` interface in `hub.go` rather than referencing `actions.DB` directly, which the plan explicitly noted as a design decision.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- SSELimiter ready for wiring into SSE HTTP handler (08-04 or later)
- PostgresHub.Subscribe ready for SSE handler; PostgresHub.Start() ready for server startup goroutine
- NotifyHub.Publish ready for action layer integration (pg_notify after successful writes)
- pgxlisten dependency added to go.mod; no additional setup required

---
*Phase: 08-background-jobs-production-readiness*
*Completed: 2026-02-19*
