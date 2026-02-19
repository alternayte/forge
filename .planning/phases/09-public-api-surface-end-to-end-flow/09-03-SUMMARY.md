---
phase: 09-public-api-surface-end-to-end-flow
plan: 03
subsystem: api
tags: [forge, public-api, app-builder, chi, huma, pgx, error-handling, transactions]

# Dependency graph
requires:
  - phase: 09-01
    provides: schema/ public package at repo root
  - phase: 09-02
    provides: forge/auth, forge/sse, forge/notify, forge/jobs, forge/forgetest public packages
provides:
  - forge.App builder with New(), RegisterAPIRoutes(), RegisterHTMLRoutes(), UseRecovery(), UseTokenStore(), UseAPIKeyStore(), Listen(), Pool()
  - forge.LoadConfig() reading forge.toml via internal/config
  - forge.Config wrapper with DatabaseURL() and ServerAddr() helpers
  - forge.Error struct matching errors.go.tmpl pattern with convenience constructors
  - forge.Transaction() and forge.TransactionWithJobs() wrapping pgx.BeginFunc
  - forge.DB interface for pgx-compatible query execution
  - forge.TransactionFunc type alias
  - internal/config.Config.API field (APIConfig) wired in Default()
affects: [09-04, 09-05, generated main.go assembly]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "forge.App builder pattern: New(cfg).RegisterAPIRoutes(fn).Listen(addr)"
    - "Config wrapper: public Config struct wraps *internal/config.Config so internal types never leak"
    - "Error type mirroring: forge.Error matches errors.go.tmpl for consistent error shapes across generated and public code"
    - "Transaction wrapping: forge.Transaction(ctx, pool, fn) adapts TransactionFunc signature to pgx.BeginFunc"

key-files:
  created:
    - forge/forge.go
    - forge/config.go
    - forge/errors.go
    - forge/transaction.go
  modified:
    - internal/config/config.go

key-decisions:
  - "forge.App stores *config.Config (internal) not forge.Config — Config wrapper is only for the public API surface at New() boundary"
  - "API APIConfig field added to internal/config/config.go so a.cfg.API works in Listen() without requiring SetupAPI callers to supply it separately"
  - "forge.Error mirrors errors.go.tmpl exactly including UniqueViolation and ForeignKeyViolation constructors, so generated gen/errors and forge.Error are structurally identical"
  - "forge.Transaction adapts TransactionFunc (ctx, tx) to pgx.BeginFunc's func(tx) signature via closure wrapper"
  - "pgconn.CommandTag imported as github.com/jackc/pgx/v5/pgconn (subdirectory of pgx v5 module, not separate pgconn module)"

patterns-established:
  - "Builder pattern: all App configuration methods return *App for chaining"
  - "Nil guard: recoveryMw defaults to passthrough if not set via UseRecovery()"
  - "Graceful shutdown: signal.NotifyContext + 10s shutdown timeout in Listen()"

requirements-completed: []

# Metrics
duration: 2min
completed: 2026-02-19
---

# Phase 9 Plan 03: forge App Builder, Config, Error, and Transaction Summary

**Public forge package runtime API: App builder with graceful shutdown, Config loader, Error type matching generated code, and Transaction wrappers using pgx.BeginFunc**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-19T21:10:23Z
- **Completed:** 2026-02-19T21:12:15Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Created `forge/forge.go`: App builder with full lifecycle management — DB pool, session manager, API+HTML route wiring, graceful shutdown on SIGINT/SIGTERM
- Created `forge/config.go`: Public Config wrapper that loads forge.toml via internal/config with DatabaseURL() and ServerAddr() helpers
- Created `forge/errors.go`: Error struct and convenience constructors exactly matching `errors.go.tmpl` so generated and public error types are structurally identical
- Created `forge/transaction.go`: DB interface, TransactionFunc, Transaction(), TransactionWithJobs() matching `transaction.go.tmpl` pattern
- Modified `internal/config/config.go`: Added `API APIConfig` field so `a.cfg.API` works in Listen() when calling SetupAPI

## Task Commits

Each task was committed atomically:

1. **Task 1: Create forge.App builder and Config loader** - `18f3275` (feat)
2. **Task 2: Create forge.Error and forge.Transaction types** - `bfedcf4` (feat)

**Plan metadata:** _(to be added in final commit)_

## Files Created/Modified
- `forge/forge.go` - App builder with New(), RegisterAPIRoutes(), RegisterHTMLRoutes(), UseRecovery(), UseTokenStore(), UseAPIKeyStore(), Listen(), Pool()
- `forge/config.go` - Config wrapper with LoadConfig(), DatabaseURL(), ServerAddr()
- `forge/errors.go` - Error struct, GetStatus(), Unwrap(), and constructors: NotFound, UniqueViolation, ForeignKeyViolation, Unauthorized, Forbidden, BadRequest, InternalError, NewValidationError
- `forge/transaction.go` - DB interface, TransactionFunc, Transaction(), TransactionWithJobs()
- `internal/config/config.go` - Added `API APIConfig \`toml:"api"\`` field and `API: DefaultAPIConfig()` in Default()

## Decisions Made
- `forge.App` stores `*config.Config` internally — the public `forge.Config` wrapper is only used at the `New()` call boundary to prevent internal type leakage
- `API APIConfig` field added to `internal/config/config.go` (not a separate lookup) so `a.cfg.API` just works in `Listen()` without additional wiring
- `forge.Error` mirrors `errors.go.tmpl` exactly, including `UniqueViolation` and `ForeignKeyViolation` constructors that were in the template but not in the plan spec
- `forge.Transaction` wraps `TransactionFunc(ctx, tx)` in a closure to match `pgx.BeginFunc`'s `func(pgx.Tx) error` signature
- `pgconn.CommandTag` uses import path `github.com/jackc/pgx/v5/pgconn` (it's a subdirectory of the pgx v5 module, not a standalone `pgconn` module)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed pgconn.CommandTag import path**
- **Found during:** Task 2 (forge/transaction.go compilation)
- **Issue:** Plan specified `github.com/jackc/pgconn` which is not in the module graph; pgconn is a subdirectory of pgx/v5
- **Fix:** Changed import to `github.com/jackc/pgx/v5/pgconn`
- **Files modified:** forge/transaction.go
- **Verification:** `go build ./forge/...` passes
- **Committed in:** bfedcf4 (Task 2 commit)

**2. [Rule 1 - Bug] Fixed TransactionFunc wrapping for pgx.BeginFunc**
- **Found during:** Task 2 (forge/transaction.go compilation)
- **Issue:** `pgx.BeginFunc` expects `func(pgx.Tx) error` but `TransactionFunc` is `func(ctx context.Context, tx pgx.Tx) error`; cannot pass directly
- **Fix:** Wrapped in closure `func(tx pgx.Tx) error { return fn(ctx, tx) }` for Transaction(); TransactionWithJobs already used inline closure
- **Files modified:** forge/transaction.go
- **Verification:** `go build ./forge/...` passes
- **Committed in:** bfedcf4 (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (2 bugs found at compile time)
**Impact on plan:** Both fixes necessary for compilation. No scope creep. The underlying API shapes are correct as specified.

## Issues Encountered
None beyond the auto-fixed compile errors above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- forge package now provides the complete runtime API: App, Config, Error, Transaction
- Ready for Phase 09-04 which will wire up generated code references to forge types
- `go build ./...` passes across the full module
- forge.App.Listen() requires a real database URL at runtime; test infrastructure uses forge/forgetest

## Self-Check: PASSED

All files created and commits verified:
- forge/forge.go: FOUND
- forge/config.go: FOUND
- forge/errors.go: FOUND
- forge/transaction.go: FOUND
- commit 18f3275: FOUND
- commit bfedcf4: FOUND

---
*Phase: 09-public-api-surface-end-to-end-flow*
*Completed: 2026-02-19*
