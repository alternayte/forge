---
phase: 08-background-jobs-production-readiness
plan: 04
subsystem: jobs
tags: [river, background-jobs, code-generation, pgx, otel, transactional-enqueueing]

# Dependency graph
requires:
  - phase: 08-background-jobs-production-readiness
    plan: 01
    provides: "HooksIR parser types, hasHooks/pascal funcmap helpers"
  - phase: 07-advanced-data-features
    provides: "Parser IR, ResourceOptionsIR with Hooks field, auth package with TenantFromContext"
provides:
  - "internal/jobs package with NewRiverClient, RunRiverMigrations, riverErrorHandler"
  - "actions.go.tmpl generates transactional InsertTx calls inside pgx.BeginFunc for AfterCreate/AfterUpdate hooks"
  - "actions.go.tmpl generates inline Args structs (Kind/ResourceID/TenantID) avoiding import cycles"
  - "scaffold_jobs.go.tmpl generates per-hook Worker stubs referencing actions.{Kind}Args"
  - "scaffoldFiles() conditionally includes jobs.go for resources with Hooks"
  - "DB interface extended with Begin() for pgx.BeginFunc compatibility"
affects:
  - "08-05 (forge build/deploy): NewRiverClient wiring into app startup"
  - "Any plan that generates or tests actions.go output with Hooks resources"

# Tech tracking
tech-stack:
  added:
    - "github.com/riverqueue/river v0.30.2 — background job processing"
    - "github.com/riverqueue/river/riverdriver/riverpgxv5 v0.30.2 — pgx v5 River driver"
    - "github.com/riverqueue/river/rivermigrate v0.30.2 — River schema migrations"
    - "github.com/riverqueue/rivercontrib/otelriver v0.7.0 — OpenTelemetry middleware for River"
  patterns:
    - "Inline Args structs in generated actions.go avoid import cycles (generated package cannot import user-scaffolded resource packages)"
    - "pgx.BeginFunc wraps Bob insert/update + InsertTx calls so job enqueue is atomic with record mutation (JOBS-02)"
    - "if result == nil guard makes InsertTx calls structurally present but safely unreachable until Bob is wired"
    - "scaffoldFiles(resource) accepts ResourceIR to conditionally append jobs.go — same pattern used by ScaffoldResource, renderScaffoldToMap, DiffResource"

key-files:
  created:
    - internal/jobs/client.go
    - internal/generator/templates/scaffold_jobs.go.tmpl
  modified:
    - internal/generator/templates/actions.go.tmpl
    - internal/generator/templates/actions_types.go.tmpl
    - internal/generator/scaffold.go
    - go.mod
    - go.sum

key-decisions:
  - "[Phase 08-04]: River client type is *river.Client[pgx.Tx] not *river.Client[pgxpool.Pool] — riverpgxv5.New(pool) driver's transaction type is pgx.Tx"
  - "[Phase 08-04]: DB interface extended with Begin(ctx) to satisfy pgx.BeginFunc's interface requirement — pgxpool.Pool already implements this; no impact on existing callers"
  - "[Phase 08-04]: Inline Args structs in actions.go (not in resources/ scaffold) avoid import cycles — generated gen/actions cannot import user-scaffolded resources/{name}/"
  - "[Phase 08-04]: if result == nil guard before InsertTx calls makes transactional job enqueueing structurally present and vet-clean before Bob queries are wired"
  - "[Phase 08-04]: TenantFromContext returns (uuid.UUID, bool); blank identifier discards the bool — tenant ID defaults to uuid.Nil when context missing"

patterns-established:
  - "River client wires otelriver middleware globally — all River jobs get OTel spans without per-worker instrumentation"
  - "RunRiverMigrations uses rivermigrate.New(riverpgxv5.New(pool), nil) — call at startup before river.Client.Start()"
  - "riverErrorHandler implements river.ErrorHandler with slog.WarnContext (error) / slog.ErrorContext (panic) — consistent with project slog pattern"

requirements-completed: [JOBS-01, JOBS-02, JOBS-03, JOBS-04]

# Metrics
duration: 6min
completed: 2026-02-19
---

# Phase 08 Plan 04: River Integration and Job Worker Scaffold Summary

**River client with otelriver middleware, transactional InsertTx in generated actions (pgx.BeginFunc + inline Args structs), and scaffold_jobs.go.tmpl generating per-hook Worker stubs for resources with Hooks**

## Performance

- **Duration:** 6 min
- **Started:** 2026-02-19T17:29:53Z
- **Completed:** 2026-02-19T17:36:41Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Created `internal/jobs/client.go` with `NewRiverClient` (queues from forge.toml + otelriver middleware), `RunRiverMigrations` (rivermigrate up), and `riverErrorHandler` (slog-based error/panic logging)
- Updated `actions.go.tmpl` to generate transactional InsertTx calls inside `pgx.BeginFunc` for resources with AfterCreate/AfterUpdate hooks; inline Args structs avoid import cycles
- Created `scaffold_jobs.go.tmpl` generating per-hook Worker structs with `river.WorkerDefaults[actions.{Kind}Args]` and `Work()` stubs that restore tenant context (JOBS-03)
- Updated `scaffoldFiles()` in scaffold.go to accept `parser.ResourceIR` and conditionally include `jobs.go` for resources with Hooks
- Extended `DB` interface with `Begin(ctx)` method to satisfy `pgx.BeginFunc` requirements without breaking any existing callers

## Task Commits

Each task was committed atomically:

1. **Task 1: Create River client package and job worker scaffold template** - `c27620d` (feat)
2. **Task 2: Update actions template for transactional InsertTx and scaffold integration** - `a02f447` (feat)

**Plan metadata:** TBD (docs: complete plan)

## Files Created/Modified
- `internal/jobs/client.go` - NewRiverClient, RunRiverMigrations, riverErrorHandler
- `internal/generator/templates/scaffold_jobs.go.tmpl` - Per-hook Worker scaffold with WorkerDefaults and Work() stub
- `internal/generator/templates/actions.go.tmpl` - River imports, inline Args structs, River field on DefaultActions, pgx.BeginFunc transactional blocks for AfterCreate/AfterUpdate
- `internal/generator/templates/actions_types.go.tmpl` - Added Begin(ctx) to DB interface
- `internal/generator/scaffold.go` - scaffoldFiles(resource) accepts ResourceIR; conditionally appends jobs.go
- `go.mod` / `go.sum` - river, riverpgxv5, rivermigrate, otelriver dependencies

## Decisions Made
- River client generic type is `*river.Client[pgx.Tx]` (not `pgxpool.Pool`) — riverpgxv5.New() returns a driver whose transaction type is `pgx.Tx`, which is what River's type parameter captures
- `DB` interface required `Begin(ctx)` to satisfy `pgx.BeginFunc`'s inline interface (`interface{ Begin(ctx) (Tx, error) }`). Extended the generated `actions_types.go.tmpl`; pgxpool.Pool already implements this so no callers break
- Inline Args structs generated in `gen/actions/{resource}.go` (not in `resources/{name}/jobs.go`) — avoids circular imports since the generated actions package cannot import user-scaffolded resource packages
- Used `if result == nil` guard in `pgx.BeginFunc` closure to make InsertTx calls structurally present but not unreachable (go vet clean) until Bob insert queries are wired

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Corrected River sub-module import paths**
- **Found during:** Task 1 (installing dependencies)
- **Issue:** Plan specified `go get github.com/riverqueue/river/riverpgxv5` but the actual module path is `github.com/riverqueue/river/riverdriver/riverpgxv5` (sub-module under `riverdriver/`)
- **Fix:** Used correct import path `github.com/riverqueue/river/riverdriver/riverpgxv5`
- **Files modified:** go.mod, go.sum
- **Verification:** `go build ./internal/jobs/...` compiles
- **Committed in:** `c27620d` (Task 1 commit)

**2. [Rule 1 - Bug] Added Begin() to DB interface for pgx.BeginFunc compatibility**
- **Found during:** Task 2 (updating actions template)
- **Issue:** `pgx.BeginFunc(ctx, a.DB, ...)` requires the second argument to implement `interface{ Begin(ctx) (pgx.Tx, error) }`. The existing `DB` interface only had `QueryRow`, `Query`, `Exec` — missing `Begin`
- **Fix:** Added `Begin(ctx context.Context) (pgx.Tx, error)` to DB interface in `actions_types.go.tmpl`. `pgxpool.Pool` already implements this method so no existing code breaks
- **Files modified:** `internal/generator/templates/actions_types.go.tmpl`
- **Verification:** `go build ./internal/generator/...` and `go test ./internal/generator/...` both pass
- **Committed in:** `a02f447` (Task 2 commit)

**3. [Rule 1 - Bug] Used if-guard pattern instead of unreachable code after return**
- **Found during:** Task 2 (writing InsertTx template block)
- **Issue:** Plan's template pseudocode showed `return errors.InternalError(...)` before the `{{range .Options.Hooks.AfterCreate}}` InsertTx calls, which would make the code unreachable (go vet error)
- **Fix:** Used `if result == nil { return errors.InternalError(...) }` guard so InsertTx calls appear AFTER the guard but are conditionally reachable (go vet clean)
- **Files modified:** `internal/generator/templates/actions.go.tmpl`
- **Verification:** `go vet ./...` passes with no issues
- **Committed in:** `a02f447` (Task 2 commit)

---

**Total deviations:** 3 auto-fixed (1 blocking import path, 2 Rule 1 bugs)
**Impact on plan:** All fixes required for correctness and compilation. No scope creep.

## Issues Encountered
- River uses sub-modules with non-obvious paths: the driver package is `github.com/riverqueue/river/riverdriver/riverpgxv5`, not `github.com/riverqueue/river/riverpgxv5`. Discovered via `go list -m all | grep river` after initial `go get` failed.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- River client ready for wiring into application startup in Phase 08-05
- `internal/jobs` package provides `NewRiverClient`, `RunRiverMigrations`, and `riverErrorHandler` for app assembly
- Generated actions.go will emit transactional job enqueueing for any resource with `WithHooks()` in schema
- Scaffold pattern complete: `forge generate resource` creates `resources/{name}/jobs.go` with Worker stubs for each hook kind
- All existing generator tests pass; go vet clean across the project

---
*Phase: 08-background-jobs-production-readiness*
*Completed: 2026-02-19*
