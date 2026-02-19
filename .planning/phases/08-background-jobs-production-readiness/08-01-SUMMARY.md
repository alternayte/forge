---
phase: 08-background-jobs-production-readiness
plan: 01
subsystem: schema
tags: [schema-dsl, parser, ast, ir, code-generation, river, hooks, lifecycle]

# Dependency graph
requires:
  - phase: 07-advanced-data-features
    provides: "Parser IR, extractor patterns, SchemaItem interface, PermissionsIR, funcmap helpers"
provides:
  - "Fixed Phase 7 generator regressions (4 tests green): FactoryTemplateData struct and atlas template scope"
  - "schema.WithHooks() DSL type with JobRef and Hooks struct types implementing SchemaItem interface"
  - "HooksIR and JobRefIR IR types in parser/ir.go; Hooks field on ResourceOptionsIR"
  - "extractHooks() AST extraction function in parser/extractor.go"
  - "hasHooks() and pascal() funcmap helpers registered in BuildFuncMap()"
affects:
  - "08-04 (River integration): depends on hasHooks funcmap and HooksIR in templates"
  - "08-02/08-05: any plan needing to check hooks in generated templates"

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "HooksItem implements SchemaItem marker interface (same pattern as PermissionItem)"
    - "isHooksType() gating function mirrors isPermissionType() pattern in extractor"
    - "Template variable capture: $resourceOptions := .Options saves outer range context for use inside inner range"

key-files:
  created:
    - internal/schema/hooks.go
  modified:
    - internal/generator/factories.go
    - internal/generator/funcmap.go
    - internal/generator/templates/atlas_schema.hcl.tmpl
    - internal/generator/atlas_test.go
    - internal/parser/ir.go
    - internal/parser/extractor.go

key-decisions:
  - "[Phase 08-01]: Template variable $resourceOptions captures ResourceOptionsIR in outer range so SoftDelete is accessible inside inner range .Fields — Go templates use $ for top-level data, not intermediate range context"
  - "[Phase 08-01]: TestGenerateAtlasSchema_ProductTable updated to assert products_sku_unique_active (partial WHERE deleted_at IS NULL) when SoftDelete=true — test was inconsistent with Phase 7 behavior"
  - "[Phase 08-01]: pascal() is a distinct funcmap helper from camel() — camel() only uppercases first char, pascal() splits on underscores for full snake_case conversion"

patterns-established:
  - "Range context capture: save outer context as $varName := .Field before entering inner range to preserve access inside nested loops"
  - "JobRef uses only string literal values (Kind/Queue) so it remains AST-parseable without schema package import at parse time"

requirements-completed: [SCHEMA-09]

# Metrics
duration: 5min
completed: 2026-02-19
---

# Phase 08 Plan 01: Schema Foundation and Generator Regression Fixes Summary

**Fixed 4 Phase 7 generator regressions (atlas template scope bug + FactoryTemplateData missing fields) and added schema.WithHooks() DSL, HooksIR parser types, and pascal/hasHooks funcmap helpers for River job lifecycle integration**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-19T17:20:29Z
- **Completed:** 2026-02-19T17:25:41Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Fixed 4 Phase 7 generator test regressions: added `Options` and `HasTimestamps` fields to `FactoryTemplateData` struct in `factories.go`, and fixed `atlas_schema.hcl.tmpl` scope bug where `$.Options.SoftDelete` evaluated at wrong level inside nested range loops
- Created `internal/schema/hooks.go` with `JobRef`, `Hooks`, and `HooksItem` types following the `PermissionItem` SchemaItem pattern, with `WithHooks()` constructor
- Extended `internal/parser/ir.go` with `JobRefIR`, `HooksIR` types and `Hooks HooksIR` field on `ResourceOptionsIR`; added `extractHooks()` AST extractor and `isHooksType()` gating in `extractor.go`
- Added `hasHooks()` and `pascal()` funcmap helpers registered in `BuildFuncMap()` for use in Phase 08 Plan 04 templates

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix Phase 7 generator regressions and add Hooks schema DSL** - `e722b6c` (feat)
2. **Task 2: Extend parser IR and AST extractor for Hooks** - `e375f3a` (feat)

**Plan metadata:** TBD (docs: complete plan)

## Files Created/Modified
- `internal/schema/hooks.go` - JobRef, Hooks, HooksItem types; WithHooks() SchemaItem constructor
- `internal/generator/factories.go` - Added Options and HasTimestamps fields to FactoryTemplateData struct
- `internal/generator/templates/atlas_schema.hcl.tmpl` - Captured $resourceOptions from outer range; fixed inner-range SoftDelete reference
- `internal/generator/atlas_test.go` - Fixed TestGenerateAtlasSchema_ProductTable to check for partial unique index (products_sku_unique_active) when SoftDelete=true
- `internal/parser/ir.go` - Added JobRefIR, HooksIR types; Hooks field on ResourceOptionsIR
- `internal/parser/extractor.go` - Added isHooksType(), extractHooks(); wired WithHooks branch into extractSchemaDefinition()
- `internal/generator/funcmap.go` - Added hasHooks() and pascal() helpers; registered both in BuildFuncMap()

## Decisions Made
- Used template variable capture (`$resourceOptions := .Options`) to preserve outer range context inside inner `{{range .Fields}}` loop — Go templates do not preserve the parent `.` when entering a nested range
- Updated `TestGenerateAtlasSchema_ProductTable` assertion to check for `products_sku_unique_active` (the correct Phase 7 soft-delete partial index behavior) since the test had `SoftDelete: true` but was checking for the non-partial index name; this was an inconsistency in the test
- `pascal()` is a separate funcmap from `camel()` because `camel()` only uppercases the first character and does not handle underscores; `pascal()` splits on `_` and title-cases each segment

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed TestGenerateAtlasSchema_ProductTable assertion for soft-delete unique index**
- **Found during:** Task 1 (Fix Phase 7 generator regressions)
- **Issue:** After fixing the `$resourceOptions.SoftDelete` scope bug, the atlas template correctly generated `products_sku_unique_active` (partial unique index) for a resource with `SoftDelete: true`. The test expected `products_sku_unique` (non-partial index), which is an inconsistency — the test resource has `SoftDelete: true`, so the correct output is the partial index.
- **Fix:** Updated the test to check for `products_sku_unique_active` and `where = "deleted_at IS NULL"`, which matches the actual Phase 7 design intent.
- **Files modified:** `internal/generator/atlas_test.go`
- **Verification:** Test passes, behavior is consistent with soft-delete partial index feature
- **Committed in:** `e722b6c` (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Auto-fix necessary for test correctness. No scope creep.

## Issues Encountered
- The initial fix to `atlas_schema.hcl.tmpl` changed `$.Options.SoftDelete` to `.Options.SoftDelete` but that also failed because inside `{{range .Fields}}`, `.` is a `FieldIR` not a `ResourceIR`. Solution: capture `$resourceOptions := .Options` before entering the fields range, then reference `$resourceOptions.SoftDelete` inside the loop.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Schema DSL foundation complete: `WithHooks()` available for schema definitions
- Parser IR ready: `HooksIR` and `JobRefIR` types extractable from AST
- Funcmap helpers ready: `hasHooks` and `pascal` available for Plan 04 (River integration) templates
- All 4 Phase 7 regressions resolved; entire generator and parser test suites pass

---
*Phase: 08-background-jobs-production-readiness*
*Completed: 2026-02-19*
