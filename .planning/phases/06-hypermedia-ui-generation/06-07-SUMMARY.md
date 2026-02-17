---
phase: 06-hypermedia-ui-generation
plan: 07
subsystem: testing
tags: [pgtestdb, pgxpool, atlas, httptest, sse, datastar, postgresql]

# Dependency graph
requires:
  - phase: 04-action-layer-error-handling
    provides: "pgxpool-based Actions interface that forgetest.NewTestPool enables"
  - phase: 02-code-generation-engine
    provides: "Atlas HCL schema and migrations/ directory forged apps will use"
provides:
  - "forgetest.NewTestDB — isolated PostgreSQL schema per test with pgtestdb + Atlas CLI migrator"
  - "forgetest.NewTestPool — pgxpool.Pool wrapper for project-compatible database tests"
  - "forgetest.NewApp — httptest.Server wrapper with auto-cleanup"
  - "forgetest.PostDatastar — SSE form submission helper with JSON signals"
  - "forgetest.GetDatastar — SSE GET request helper"
  - "forgetest.ReadSSEEvents — SSE event parser returning []SSEEvent"
affects:
  - 06-hypermedia-ui-generation
  - any future integration tests in generated forge apps

# Tech tracking
tech-stack:
  added:
    - "github.com/peterldowns/pgtestdb v0.1.1 — isolated PostgreSQL test schemas"
  patterns:
    - "Custom pgtestdb.Migrator using Atlas CLI (atlas migrate apply) for migration execution"
    - "runtime.Caller(0) for repo-root-relative path resolution in test helpers"
    - "pgtestdb.Custom (not pgtestdb.New) used to avoid sql.DB interference with pgxpool"
    - "t.Helper() + t.Cleanup() pattern consistently applied across all helpers"

key-files:
  created:
    - internal/forgetest/db.go
    - internal/forgetest/app.go
    - internal/forgetest/datastar.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "Use pgtestdb.Custom (not New) for NewTestPool — avoids open sql.DB connection interfering with pgxpool"
  - "Implement custom atlasMigrator using Atlas CLI shell-out (not a Go library) — consistent with project's existing Atlas CLI approach"
  - "runtime.Caller(0) for repo root resolution — package path is stable (internal/forgetest/db.go), reliable across all callers"
  - "NoopMigrator fallback not implemented — migrations dir missing is surfaced as error, not silently skipped"

patterns-established:
  - "forgetest package: thin wrappers with t.Helper() + t.Cleanup() for all test infrastructure"
  - "atlasMigrator: Hash() uses common.HashDir for content-based cache invalidation"

requirements-completed: []

# Metrics
duration: 4min
completed: 2026-02-17
---

# Phase 06 Plan 07: Forgetest Infrastructure Summary

**forgetest package with isolated PostgreSQL test schemas (pgtestdb + Atlas CLI migrator), httptest.Server wrapper, and Datastar SSE test helpers**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-17T21:03:22Z
- **Completed:** 2026-02-17T21:06:41Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Created `internal/forgetest/db.go` providing `NewTestDB` (isolated PostgreSQL schema via pgtestdb) and `NewTestPool` (pgxpool wrapper) with Atlas CLI migration support
- Created `internal/forgetest/app.go` with `NewApp` (httptest.Server with auto-cleanup) and `AppURL` convenience helper
- Created `internal/forgetest/datastar.go` with `PostDatastar`, `GetDatastar`, `ReadSSEEvents`, and `SSEEvent` for testing Datastar SSE endpoints
- Added `github.com/peterldowns/pgtestdb v0.1.1` as dependency

## Task Commits

Each task was committed atomically:

1. **Task 1: Create NewTestDB with pgtestdb and atlas migrator** - `7ca18a1` (feat)
2. **Task 2: Create NewApp HTTP test server and PostDatastar helper** - `31d18d0` (feat)

## Files Created/Modified

- `internal/forgetest/db.go` — TestDBConfig, DefaultTestDBConfig, atlasMigrator, NewTestDB, NewTestPool, option functions
- `internal/forgetest/app.go` — NewApp, AppURL
- `internal/forgetest/datastar.go` — PostDatastar, GetDatastar, ReadSSEEvents, SSEEvent
- `go.mod` / `go.sum` — added github.com/peterldowns/pgtestdb v0.1.1

## Decisions Made

- **pgtestdb.Custom over pgtestdb.New for NewTestPool:** `pgtestdb.New` returns a `*sql.DB` connection that stays open; using `pgtestdb.Custom` gives us just the database config so pgxpool can open its own connections cleanly without interference.
- **Custom atlasMigrator using Atlas CLI shell-out:** The project already uses Atlas CLI (not a Go library) for migrations. Implementing a `pgtestdb.Migrator` that shells out to `atlas migrate apply` keeps the approach consistent.
- **runtime.Caller(0) for path resolution:** The file `internal/forgetest/db.go` is at a known depth from the repo root. Using `runtime.Caller(0)` gives the absolute path of this source file at compile time, so the relative resolution to `../../` is stable regardless of test working directory.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Used pgtestdb.Custom (not pgtestdb.New) for NewTestPool**
- **Found during:** Task 1 (NewTestPool implementation)
- **Issue:** Plan specified "Extract connection string from the *sql.DB driver" but pgx stdlib's sql.DB doesn't expose the DSN publicly; the cleaner approach is pgtestdb.Custom which returns the Config directly
- **Fix:** Used `pgtestdb.Custom` which returns `*pgtestdb.Config` with URL() method, then passed that URL directly to `pgxpool.New`
- **Files modified:** internal/forgetest/db.go
- **Verification:** `go build ./internal/forgetest/` passes
- **Committed in:** 7ca18a1 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 3 - blocking implementation approach)
**Impact on plan:** Cleaner implementation than plan specified. No scope changes.

## Issues Encountered

- `github.com/peterldowns/testdb` (the module import in the plan) is the old module name. The correct package is `github.com/peterldowns/pgtestdb`. `go get` caught this immediately and corrected it.

## User Setup Required

None - no external service configuration required. Tests require a running PostgreSQL instance but no extra forge configuration.

## Next Phase Readiness

- `forgetest.NewTestDB` and `forgetest.NewTestPool` ready for use in Phase 6 integration tests
- `forgetest.NewApp` + `forgetest.PostDatastar` + `forgetest.ReadSSEEvents` ready for Datastar endpoint testing
- Phase 6 Plans 8-9 can import this package for end-to-end integration tests

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*
