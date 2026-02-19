---
phase: 09-public-api-surface-end-to-end-flow
plan: 05
subsystem: generator
tags: [generator, scaffold, dev-server, auto-migrate, pgx, atlas]

# Dependency graph
requires:
  - phase: 09-04
    provides: scaffold templates for Post resource and forge/schema import
  - phase: 09-01
    provides: forge package public API (forge.New, forge.LoadConfig, forge.App)
  - phase: 09-03
    provides: SetupAPI and RegisterAPIRoutes accepting func(huma.API) closure
provides:
  - "main.go generator template using forge.New/RegisterAPIRoutes/RegisterHTMLRoutes/Listen"
  - "GenerateMain scaffold-once logic (skip if forge.New already present)"
  - "forge dev auto-creates PostgreSQL database on start (ensureDatabase via pgx)"
  - "forge dev auto-runs migrate diff+up when gen/atlas/schema.hcl changes"
  - "hashSchemaFile SHA-256 change detection for schema drift"
  - "complete golden path: forge init -> forge generate -> forge dev"
affects: [end-to-end testing, user documentation, golden path demo]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "scaffold-once: write file only if target lacks forge.New/forge.LoadConfig marker strings"
    - "schema change detection: SHA-256 hash gen/atlas/schema.hcl before/after generation"
    - "non-fatal dev auto-migrate: warn on migrate failures rather than crashing the watcher"
    - "pgx maintenance DB connection: connect to /postgres to CREATE DATABASE app-db"

key-files:
  created:
    - internal/generator/templates/main.go.tmpl
    - internal/generator/generate_main.go
  modified:
    - internal/generator/generator.go
    - internal/watcher/dev.go

key-decisions:
  - "Scaffold-once detection uses bytes.Contains for 'forge.New' or 'forge.LoadConfig' markers — simple, fast, correct"
  - "ensureDatabase uses pgx directly (not exec createdb) — avoids external tool dependency in watcher"
  - "Auto-migrate is non-fatal in dev mode — warn on failure and continue watching"
  - "DevURL uses same database URL as DatabaseURL (Atlas creates temporary schemas for diffing)"
  - "hashSchemaFile returns empty string on missing file — triggers migrate on first run when schema.hcl doesn't exist yet"
  - "GenerateMain guarded by cfg.ProjectRoot != '' in orchestrator — safe for callers without project root"

patterns-established:
  - "Scaffold-once pattern: check for marker strings before overwriting user-editable files"
  - "Dev-server auto-migrate: hash-before/hash-after with non-fatal error path for resilient dev loop"

requirements-completed: []

# Metrics
duration: 2min
completed: 2026-02-19
---

# Phase 9 Plan 05: Generator main.go Scaffolding and Dev Auto-Migrate Summary

**forge generate scaffold-once main.go using forge.New/RegisterAllRoutes(api, registry), and forge dev auto-creates DB + auto-runs atlas migrate diff+up on schema.hcl changes**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-19T21:18:18Z
- **Completed:** 2026-02-19T21:20:30Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Generator template `main.go.tmpl` produces a wiring-complete main.go with forge.New, UseRecovery, RegisterAPIRoutes (capturing registry), RegisterHTMLRoutes, and Listen
- scaffold-once logic in `generate_main.go` skips overwriting if "forge.New" or "forge.LoadConfig" already present in main.go
- `generator.go` orchestrator calls `GenerateMain` after all gen/ generation when ProjectRoot is set
- `dev.go` enhanced with `ensureDatabase` (pgx to postgres maintenance DB), `hashSchemaFile` (SHA-256 change detection), and `runGenerationAndMigrate` (auto migrate diff+up on schema change)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add forge generate main.go scaffolding** - `2efcb18` (feat)
2. **Task 2: Add forge dev auto-DB-create and auto-migrate** - `702d4b4` (feat)

**Plan metadata:** (docs commit pending)

## Files Created/Modified
- `internal/generator/templates/main.go.tmpl` - main.go template with forge.New/RegisterAPIRoutes/RegisterHTMLRoutes/Listen wiring
- `internal/generator/generate_main.go` - GenerateMain scaffold-once function
- `internal/generator/generator.go` - Added GenerateMain call to Generate orchestrator
- `internal/watcher/dev.go` - Added ensureDatabase, hashSchemaFile, runGenerationAndMigrate; updated Start/onFileChange

## Decisions Made
- Scaffold-once detection uses `bytes.Contains` for "forge.New" / "forge.LoadConfig" — both are unique markers that the scaffold main.go placeholder lacks
- `ensureDatabase` connects to `/postgres` maintenance database via pgx (not exec.Command with `createdb`) to avoid external tool dependency in the watcher
- Auto-migrate is fully non-fatal in dev mode: failures print a warning and the watcher continues
- DevURL set to same database URL as DatabaseURL (Atlas creates temporary schemas internally for diffing)
- `hashSchemaFile` returns empty string when file doesn't exist, which triggers the migrate path on first run (correct behavior — fresh project needs initial migration)
- `GenerateMain` call in orchestrator is guarded by `cfg.ProjectRoot != ""` for backward compatibility

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 9 Plan 05 completes the final plan of Phase 9 (Public API Surface & End-to-End Flow)
- The golden path is complete: `forge init myapp` -> `forge generate` -> `forge dev`
- forge generate produces a fully wired main.go on first run (scaffold-once)
- forge dev auto-creates the database and auto-migrates on schema changes
- All forge/ public API packages are in place; the framework is ready for integration testing

---
*Phase: 09-public-api-surface-end-to-end-flow*
*Completed: 2026-02-19*
