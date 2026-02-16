---
phase: 02-code-generation-engine
plan: 02
subsystem: code-generation
tags: [atlas, hcl, postgresql, test-factories, builder-pattern, code-generation]

# Dependency graph
requires:
  - phase: 02-01
    provides: Generator infrastructure, model generation, funcmap helpers
provides:
  - Atlas HCL schema generation with PostgreSQL type mapping for all 14 field types
  - Test factory generation with builder pattern for rapid test data creation
  - Complete Generate orchestrator calling models + atlas + factories
affects: [02-03, 02-04, 02-05, database-migrations, testing]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Atlas HCL type mapping via atlasType/atlasTypeWithModifiers helper functions"
    - "Test factory builder pattern with New{Resource}() and Build{Resource}() functions"
    - "Template-driven HCL generation using writeRawFile (no Go formatting for HCL)"

key-files:
  created:
    - internal/generator/atlas.go
    - internal/generator/atlas_test.go
    - internal/generator/factories.go
    - internal/generator/factories_test.go
    - internal/generator/templates/atlas_schema.hcl.tmpl
    - internal/generator/templates/factory.go.tmpl
  modified:
    - internal/generator/funcmap.go
    - internal/generator/generator.go

key-decisions:
  - "Enum types map to text with CHECK constraints (not PostgreSQL enum type) for simpler migration story"
  - "ID field excluded from factory builders (auto-generated via gen_random_uuid())"
  - "MaxLen modifier overrides default varchar(255) length for String/Email/Slug types"
  - "Factory imports use projectModule for gen/models import path construction"

patterns-established:
  - "Atlas helper functions follow atlasType/atlasNull/atlasDefault naming convention"
  - "HCL templates use writeRawFile (not writeGoFile) to avoid Go formatting on non-Go files"
  - "Test factories provide both convenience (New{Resource}) and builder (Build{Resource}) APIs"

# Metrics
duration: 4.5min
completed: 2026-02-16
---

# Phase 02 Plan 02: Atlas Schema & Factory Generation Summary

**Atlas HCL schema generation with complete PostgreSQL type mapping and test factory builders with fluent API for all resources**

## Performance

- **Duration:** 4m 34s
- **Started:** 2026-02-16T17:29:34Z
- **Completed:** 2026-02-16T17:34:08Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- Atlas HCL schema generation producing single schema.hcl with all resources, tables, columns, indexes, and constraints
- Complete PostgreSQL type mapping for all 14 field types (UUID→uuid, String→varchar, Decimal→numeric, etc.)
- Test factory generation with builder pattern providing New{Resource}() convenience and Build{Resource}() fluent API
- Generate orchestrator now calls all three generators: models, atlas, factories

## Task Commits

Each task was committed atomically:

1. **Task 1: Create Atlas HCL schema generation with PostgreSQL type mapping** - `bd4f896` (feat)
2. **Task 2: Create test factory generation and wire Generate orchestrator** - `095f228` (feat)

## Files Created/Modified

- `internal/generator/funcmap.go` - Added atlasType, atlasTypeWithModifiers, atlasNull, atlasDefault, hasDefault, defaultTestValue helpers
- `internal/generator/atlas.go` - GenerateAtlasSchema function producing schema.hcl from ResourceIR
- `internal/generator/atlas_test.go` - Comprehensive tests for Atlas generation, type mapping, indexes, constraints
- `internal/generator/templates/atlas_schema.hcl.tmpl` - HCL template with table/column/index/primary key definitions
- `internal/generator/factories.go` - GenerateFactories function producing per-resource factory files
- `internal/generator/factories_test.go` - Tests for factory generation, builder pattern, default test values
- `internal/generator/templates/factory.go.tmpl` - Factory template with New{Resource}() and {Resource}Builder pattern
- `internal/generator/generator.go` - Updated Generate orchestrator to call all three generators

## Decisions Made

- **Enum to text mapping:** Enum types map to PostgreSQL `text` (not native enum type) for simpler migration story. CHECK constraints can be added later if needed for validation.
- **ID field exclusion:** ID field excluded from factory builders since it's auto-generated via `gen_random_uuid()` at database level.
- **MaxLen modifier:** MaxLen modifier overrides default varchar(255) length for String/Email/Slug types, enabling custom length constraints.
- **Default test values:** Test factories use reasonable defaults (uuid.New(), "test-{field}", 42, true, time.Now()) for rapid test data creation.

## Deviations from Plan

None - plan executed exactly as written. All tests pass, go vet reports no issues.

## Issues Encountered

**Template variable scoping:** Initial template used `$.Name` within nested range which referred to wrong context. Fixed by introducing `{{$resourceName := .Name}}` variable to capture resource name before nested range. Test failure immediately caught the issue.

## Next Phase Readiness

- Atlas HCL schema generation ready for Plan 03 (migration generation)
- Test factories ready for Plan 04 (action layer and CRUD logic)
- Generate orchestrator complete, ready to add more generators in future plans
- All 14 field types map correctly to PostgreSQL types, verified via tests

## Self-Check: PASSED

All created files verified:
- ✓ internal/generator/atlas.go
- ✓ internal/generator/atlas_test.go
- ✓ internal/generator/factories.go
- ✓ internal/generator/factories_test.go
- ✓ internal/generator/templates/atlas_schema.hcl.tmpl
- ✓ internal/generator/templates/factory.go.tmpl

All commits verified:
- ✓ bd4f896 (Task 1: Atlas HCL schema generation)
- ✓ 095f228 (Task 2: Factory generation and orchestrator)

---
*Phase: 02-code-generation-engine*
*Completed: 2026-02-16*
