---
phase: 05-rest-api-generation
plan: 04
subsystem: cli
tags: [cli, cobra, openapi, routes, lipgloss, huma]

# Dependency graph
requires:
  - phase: 05-rest-api-generation
    plan: 01
    provides: funcmap helpers (kebab, lowerCamel, plural) patterns used in CLI route derivation
  - phase: 05-rest-api-generation
    plan: 02
    provides: CLI structure and forge project patterns to follow
  - phase: 02-code-generation-engine
    provides: parser.ParseDir() and parser.ResourceIR for schema-driven CLI commands

provides:
  - forge routes CLI command listing 5 CRUD API routes per resource, grouped, color-coded
  - forge openapi export CLI command exporting OpenAPI 3.1 spec from parsed IR
  - buildSpecFromIR() helper generating a complete huma.OpenAPI document from resource IR

affects:
  - internal/cli/root.go (added routes and openapi subcommands)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - CLI commands derive routes from IR using same kebab/plural logic as generator templates
    - OpenAPI spec built from huma.OpenAPI struct directly (no HTTP adapter needed for spec export)
    - lipgloss styles for method color-coding (GET=green, POST=blue, PUT=yellow, DELETE=red)
    - apiRoutes() and htmlRoutes() separation pattern for Phase 6 HTML route addition

key-files:
  created:
    - internal/cli/routes.go
    - internal/cli/openapi.go
  modified:
    - internal/cli/root.go

key-decisions:
  - "Build huma.OpenAPI struct directly from IR (no HTTP adapter needed) for spec export — avoids adding chi or net/http dependencies"
  - "apiRoutes() separated from runRoutes() so Phase 6 can add htmlRoutes() to same output"
  - "routeKebab/routePlural/routeLowerCamel duplicated in cli package (not imported from generator) to keep packages independent"
  - "Spec operations use explicit Responses map (not Huma Register) to avoid adapter requirement for CLI-only spec generation"

patterns-established:
  - "CLI route commands parse IR with same flow as forge generate (findProjectRoot + config.Load + parser.ParseDir)"
  - "OpenAPI spec builds from huma.OpenAPI.AddOperation() — no handler code needed, just metadata"

# Metrics
duration: 4min
completed: 2026-02-17
---

# Phase 5 Plan 04: CLI Routes and OpenAPI Export Summary

**forge routes command displaying API routes grouped by resource with lipgloss color-coding, and forge openapi export building OpenAPI 3.1 spec from IR using huma.OpenAPI struct**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-17T19:07:42Z
- **Completed:** 2026-02-17T19:11:13Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Created `internal/cli/routes.go` with `routesCmd` that parses schemas via `parser.ParseDir()`, derives 5 CRUD routes per resource, and displays grouped output with lipgloss color-coded methods and column-aligned paths
- Created `internal/cli/openapi.go` with `openapiCmd` parent and `openapiExportCmd` subcommand supporting `--format json|yaml` and optional output filename
- `buildSpecFromIR()` builds a complete `huma.OpenAPI` struct from IR — every operation has explicit `operationId`, `tags`, and `summary` for SDK and spectral lint readiness
- Separated `apiRoutes()` function from display logic so Phase 6 can cleanly add `htmlRoutes()` alongside
- All verification checks pass: `go build`, `go vet`, all `--help` outputs correct, exported spec passes `operationId` verification

## Task Commits

Each task was committed atomically:

1. **Task 1: Create forge routes command with grouped output** - `9ea95f4` (feat)
2. **Task 2: Create forge openapi export command with JSON/YAML format flag** - `5a6b221` (feat)

## Files Created/Modified

- `internal/cli/routes.go` - routesCmd with lipgloss-styled grouped resource output and apiRoutes() helper
- `internal/cli/openapi.go` - openapiCmd + openapiExportCmd with buildSpecFromIR() and resourceOperations()
- `internal/cli/root.go` - Registered routesCmd and openapiCmd (with openapiExportCmd subcommand)

## Decisions Made

- **huma.OpenAPI struct directly**: Built the spec using `huma.OpenAPI.AddOperation()` without any HTTP adapter — cleaner than needing chi or net/http mux for a CLI-only operation that just needs spec bytes
- **apiRoutes() separation**: Isolated API route derivation into a standalone function so Phase 6 can add `htmlRoutes()` to the same `forge routes` output without refactoring
- **CLI-local helpers**: Duplicated `routeKebab`, `routePlural`, `routeLowerCamel` in the cli package rather than importing from generator — keeps packages independent (cli should not depend on generator internals)
- **Direct Responses map**: Used `huma.Operation.Responses` map directly rather than Huma's Register path — avoids needing an adapter and keeps the spec generation self-contained in the CLI

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed missing return after unreachable blank identifier statements**
- **Found during:** Task 1 compilation
- **Issue:** Initial `apiRoutes()` had `_ = lowerName` and `_ = lowerPluralName` after `return` statement, causing `missing return` compiler error
- **Fix:** Removed the unused `lowerName` and `lowerPluralName` variables since they weren't needed in the return value
- **Files modified:** internal/cli/routes.go
- **Commit:** 9ea95f4 (fixed before commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Minor compilation fix during initial write. No scope impact.

## Verification Results

All plan verification criteria passed:
1. `go build ./` — clean compile
2. `go vet ./internal/cli/` — no issues
3. `forge routes --help` — shows "Display all HTML and API routes grouped by resource"
4. `forge openapi --help` — shows export subcommand
5. `forge openapi export --help` — shows --format flag with json/yaml default
6. Routes listed grouped by resource per user decision (rails routes style)
7. Routes show correct `/api/v1/{kebab-plural}` paths with camelCase operationIds
8. Exported JSON spec verified: all 5 operations have operationId, tags, and summary
9. YAML export verified: `forge openapi export --format yaml` produces valid YAML

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `forge routes` and `forge openapi export` are production-ready CLI commands
- Phase 6 (HTML generation) can extend `apiRoutes()` with an `htmlRoutes()` counterpart in routes.go
- The `buildSpecFromIR()` function can be enhanced in Phase 6 to include actual field schemas when typed resources are available

---
*Phase: 05-rest-api-generation*
*Completed: 2026-02-17*

## Self-Check: PASSED

All files verified present:
- FOUND: internal/cli/routes.go
- FOUND: internal/cli/openapi.go
- FOUND: .planning/phases/05-rest-api-generation/05-04-SUMMARY.md

All commits verified:
- FOUND: 9ea95f4 (Task 1 commit)
- FOUND: 5a6b221 (Task 2 commit)
