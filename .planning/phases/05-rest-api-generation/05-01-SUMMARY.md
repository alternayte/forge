---
phase: 05-rest-api-generation
plan: 01
subsystem: api
tags: [huma, openapi, code-generation, rest, pagination, rfc8288]

# Dependency graph
requires:
  - phase: 04-action-layer-error-handling
    provides: Action interfaces, error types, and Registry for generated API handlers to call
  - phase: 03-query-layer
    provides: models.Filter, models.Sort structs referenced in generated route handlers
  - phase: 02-code-generation-engine
    provides: Generator patterns (renderTemplate, writeGoFile, BuildFuncMap) and model types

provides:
  - GenerateAPI function producing gen/api/ directory with inputs, outputs, routes, and types files
  - api_inputs.go.tmpl generating per-resource Huma Input structs with validation tags
  - api_outputs.go.tmpl generating per-resource response envelopes with data/pagination wrapper
  - api_register.go.tmpl generating 5 CRUD route registrations per resource with Huma operationIds
  - api_types.go.tmpl generating shared PaginationMeta, toHumaError, buildAPILinkHeader
  - funcmap helpers: kebab, lowerCamel, humaValidationTag, sortableFieldNames, filterableFields, buildLinkHeader, not, join

affects:
  - 05-rest-api-generation (subsequent plans building on api layer)
  - generator.go (needs GenerateAPI added to main Generate orchestrator)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - GenerateAPI follows existing generator pattern (renderTemplate + writeGoFile)
    - Output structs use header field tags for RFC 8288 Link header (not huma.SetHeader)
    - toHumaError maps forge.Error status codes to Huma HTTP error helpers
    - buildAPILinkHeader generates RFC 8288 Link header strings

key-files:
  created:
    - internal/generator/api.go
    - internal/generator/api_test.go
    - internal/generator/templates/api_inputs.go.tmpl
    - internal/generator/templates/api_outputs.go.tmpl
    - internal/generator/templates/api_register.go.tmpl
    - internal/generator/templates/api_types.go.tmpl
  modified:
    - internal/generator/funcmap.go (added 8 new helpers)
    - go.mod, go.sum (go mod tidy for missing entries)

key-decisions:
  - "List output structs include a Link string header field (header:\"Link\" tag) rather than huma.SetHeader call — cleaner Huma v2 pattern"
  - "buildAPILinkHeader function generated in types.go (not a template helper) — keeps link header logic inside generated code"
  - "toHumaError maps forge.Error.Detail (string) to huma.ErrorDetail.Message for 422 validation errors"
  - "funcmap not/join helpers added to support template boolean negation and string joining"

patterns-established:
  - "API generation follows same renderTemplate + writeGoFile pattern as all other generators"
  - "Route handlers are thin — parse input, call action, wrap output, return"
  - "RFC 8288 Link header set via struct field with header tag on ListOutput, not via context"

# Metrics
duration: 5min
completed: 2026-02-17
---

# Phase 5 Plan 01: REST API Generation Summary

**Three Go templates and GenerateAPI function producing Huma Input/Output structs and CRUD route registrations with RFC 8288 pagination Link headers per resource**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-17T18:58:13Z
- **Completed:** 2026-02-17T19:03:29Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Created `api_inputs.go.tmpl` generating per-resource Huma Input structs (List, Get, Create, Update, Delete) with validation tags derived from schema modifiers (MinLen, MaxLen, Enum)
- Created `api_outputs.go.tmpl` generating response envelopes with `{"data": ...}` and `{"data": [...], "pagination": {...}}` wrappers, plus `Link` header field for RFC 8288 pagination
- Created `api_register.go.tmpl` generating 5 CRUD Huma endpoints per resource with camelCase operationIds (listProducts, getProduct, etc.) and kebab-case URL paths (/api/v1/products)
- Created `api_types.go.tmpl` with shared `PaginationMeta`, `toHumaError`, and `buildAPILinkHeader`
- Added 8 new funcmap helpers: `kebab`, `lowerCamel`, `humaValidationTag`, `sortableFieldNames`, `filterableFields`, `buildLinkHeader`, `not`, `join`
- Created `GenerateAPI` function following existing generator pattern
- All 4 tests pass: TestGenerateAPI, TestGenerateAPI_MultipleResources, TestGenerateAPI_OperationIDs, TestGenerateAPI_LinkHeader

## Task Commits

Each task was committed atomically:

1. **Task 1: Create Huma Input/Output struct templates and new funcmap helpers** - `5bda9b3` (feat)
2. **Task 2: Create route registration template, GenerateAPI function, and tests** - `f01d83d` (feat)

## Files Created/Modified

- `internal/generator/api.go` - GenerateAPI function following actions.go pattern
- `internal/generator/api_test.go` - Comprehensive tests for all API generation scenarios
- `internal/generator/funcmap.go` - Added 8 new template helpers for API generation
- `internal/generator/templates/api_inputs.go.tmpl` - Per-resource Huma Input struct template
- `internal/generator/templates/api_outputs.go.tmpl` - Per-resource response envelope template with Link header field
- `internal/generator/templates/api_register.go.tmpl` - Per-resource route registration template with 5 CRUD endpoints
- `internal/generator/templates/api_types.go.tmpl` - Shared types (PaginationMeta, toHumaError, buildAPILinkHeader)
- `go.mod`, `go.sum` - Updated via go mod tidy to fix missing dependency entries

## Decisions Made

- **Link header via struct field**: Used `Link string \`header:"Link"\`` on `List{{Resource}}Output` instead of `huma.SetHeader(ctx, ...)` — cleaner Huma v2 pattern that keeps output self-contained
- **buildAPILinkHeader in types.go**: Link header building logic lives in generated `types.go` (not as a funcmap helper) — keeps the link header formatting inside the generated code where it belongs
- **humaValidationTag handles all modifier types**: Single function builds validation tags from MinLen, MaxLen, Min, Max, and Enum modifiers — consistent tag generation across all field types
- **funcmap not/join as primitives**: Added `not` (bool negation) and `join` (string slice join) as simple template utilities to enable complex conditional expressions in templates

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed missing go.sum entries**
- **Found during:** Task 1 (verifying funcmap.go compilation)
- **Issue:** `go build ./internal/generator/` failed with missing go.sum entry for `github.com/mattn/go-runewidth`
- **Fix:** Ran `go mod tidy` to update go.sum with all missing transitive dependency checksums
- **Files modified:** go.mod, go.sum
- **Verification:** `go build ./internal/generator/` and `go vet ./internal/generator/` both clean
- **Committed in:** 5bda9b3 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Pre-existing go.sum issue unrelated to this plan's changes. Fix necessary to verify compilation. No scope creep.

## Issues Encountered

- `huma.SetHeader(ctx, ...)` was the initially considered approach for the List handler Link header. Switched to the output struct `header:` tag field approach as it is the canonical Huma v2 pattern and avoids needing the Huma context accessor.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- GenerateAPI function ready to be integrated into the main `Generate()` orchestrator in generator.go
- All three templates produce valid compilable Go code (verified via goimports formatting)
- Tests confirm operationId format, path format, action delegation, and Link header generation
- Phase 05-02 can build on this to add the API to the main generation pipeline and test with real schema files

---
*Phase: 05-rest-api-generation*
*Completed: 2026-02-17*

## Self-Check: PASSED

All files verified present:
- FOUND: internal/generator/api.go
- FOUND: internal/generator/api_test.go
- FOUND: internal/generator/templates/api_inputs.go.tmpl
- FOUND: internal/generator/templates/api_outputs.go.tmpl
- FOUND: internal/generator/templates/api_register.go.tmpl
- FOUND: internal/generator/templates/api_types.go.tmpl

All commits verified:
- FOUND: 5bda9b3 (Task 1 commit)
- FOUND: f01d83d (Task 2 commit)
