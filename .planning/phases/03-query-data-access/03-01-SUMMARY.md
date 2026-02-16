---
phase: 03-query-data-access
plan: 01
subsystem: code-generation
tags: [validation, query-builder, bob, type-safety]

dependency_graph:
  requires:
    - 02-01 (model generation for input types)
    - 02-02 (factory generation for test patterns)
  provides:
    - Validation functions for create/update operations
    - Bob query mods for type-safe filtering and sorting
  affects:
    - Future action layer implementation (will consume validation)
    - Future query layer implementation (will consume query mods)

tech_stack:
  added:
    - Bob query builder mod generation
    - Validation function generation with FieldError types
  patterns:
    - Shared types.go file pattern (prevents type redeclaration)
    - Per-resource validation files with Create/Update variants
    - Type-safe query filters based on field types (EQ/NEQ/GTE/LTE/Contains)

key_files:
  created:
    - internal/generator/validation.go
    - internal/generator/validation_test.go
    - internal/generator/templates/validation.go.tmpl
    - internal/generator/templates/validation_types.go.tmpl
    - internal/generator/queries.go
    - internal/generator/queries_test.go
    - internal/generator/templates/queries.go.tmpl
  modified:
    - internal/generator/funcmap.go (added validation/query helpers)
    - internal/generator/generator.go (wired new generators)

decisions:
  - Separate types.go file prevents FieldError redeclaration across multiple resources
  - ValidationErrors implements error interface with joined message strings
  - Query mods use Bob's sm.QueryMod[*psql.SelectQuery] generic type
  - String fields get Contains (ILIKE) filter, numeric/date fields get GTE/LTE
  - Email validation uses basic "@" and "." presence check (not full RFC compliance)
  - SortMod validates field names via switch statement for type safety

metrics:
  duration: 3.98
  tasks_completed: 2
  files_created: 7
  files_modified: 2
  tests_added: 2
  completed_at: 2026-02-16T20:25:00Z
---

# Phase 03 Plan 01: Validation and Query Builder Generation Summary

**One-liner:** Type-safe validation functions with FieldError maps and Bob query mods with per-field-type filters (EQ/NEQ/GTE/LTE/Contains)

## What Was Built

Generated two core code generation subsystems:

1. **Validation Generator** - Produces `gen/validation/` files with:
   - Shared `types.go` containing `FieldError` struct and `ValidationErrors` map type
   - Per-resource validation files with `Validate{Resource}Create` and `Validate{Resource}Update` functions
   - Rule checks: Required (zero-value), MaxLen, MinLen, Enum membership, Email format
   - Create vs Update distinction: Update only validates non-nil pointer fields

2. **Query Builder Generator** - Produces `gen/queries/` files with:
   - `{Resource}Filters` struct with type-safe filter methods per filterable field
   - Filter methods vary by field type: all get EQ/NEQ, numeric/date get GTE/LTE, string types get Contains (ILIKE)
   - `FilterMods()` function converts `models.{Resource}Filter` to slice of Bob query mods
   - `SortMod()` function validates sortable fields and returns Bob ORDER BY mod

Both generators wired into `Generate()` orchestrator after factories generation.

## Task Breakdown

### Task 1: Create validation function generation
**Status:** Complete
**Commit:** ae1417f
**Files:** validation.go, validation_test.go, validation.go.tmpl, validation_types.go.tmpl, funcmap.go
**Duration:** ~2 minutes

Created validation generator with:
- Template for shared types (FieldError, ValidationErrors with Add/HasErrors/Error methods)
- Template for per-resource validation with Create/Update variants
- Helper functions: zeroValue, enumValues, hasMinLen, getMinLen, getMaxLen
- Tests verify generated code compiles and contains expected validation logic

### Task 2: Create query builder mod generation and wire orchestrator
**Status:** Complete
**Commit:** 2e11aa4
**Files:** queries.go, queries_test.go, queries.go.tmpl, generator.go
**Duration:** ~2 minutes

Created query builder generator with:
- Template generates filter methods based on field type (string → Contains, numeric → GTE/LTE)
- FilterMods function builds Bob mod slice from Filter struct
- SortMod function validates field name and builds ORDER BY mod
- Wired both GenerateValidation and GenerateQueries into orchestrator

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

All verification steps passed:

1. ✅ `go test ./internal/generator/ -v` - All tests pass (18 test suites)
2. ✅ `go build ./internal/generator/` - Clean compile
3. ✅ `go vet ./internal/generator/` - No issues
4. ✅ Generated validation files contain FieldError types, Validate{Resource}Create/Update
5. ✅ Generated query files contain FilterMods, SortMod, per-field EQ/NEQ/GTE/LTE/Contains
6. ✅ Generate() orchestrator calls all 5 generators: Models, Atlas, Factories, Validation, Queries

## Technical Decisions

**Shared types.go pattern:** Validation types (FieldError, ValidationErrors) generated once in separate file to prevent redeclaration when multiple resources exist. Cleaner than "only in first file" conditional logic.

**ValidationErrors as map:** `map[string][]FieldError` allows multiple errors per field. Implements error interface with joined message string for easy display.

**Bob generic types:** Used `sm.QueryMod[*psql.SelectQuery]` as return type for type-safe Bob mods. Verified against Bob v0.42.0 API.

**Type-specific filters:** String types get Contains (ILIKE) for partial match, numeric/date types get GTE/LTE for range queries. Matches common query patterns.

**Basic email validation:** Uses simple "@" and "." presence check rather than full RFC 5322 regex. Sufficient for generated code - users can add custom validators.

**SortMod field validation:** Switch statement on field name provides type safety and clear error path (invalid field returns no-op mod).

## Next Steps

Phase 03 Plan 02 will implement Bob query execution layer that consumes these query mods. Plan 03 will implement action layer that consumes these validation functions.

## Self-Check

Verifying created files exist:

```bash
[ -f "internal/generator/validation.go" ] && echo "FOUND: internal/generator/validation.go" || echo "MISSING: internal/generator/validation.go"
[ -f "internal/generator/validation_test.go" ] && echo "FOUND: internal/generator/validation_test.go" || echo "MISSING: internal/generator/validation_test.go"
[ -f "internal/generator/templates/validation.go.tmpl" ] && echo "FOUND: internal/generator/templates/validation.go.tmpl" || echo "MISSING: internal/generator/templates/validation.go.tmpl"
[ -f "internal/generator/templates/validation_types.go.tmpl" ] && echo "FOUND: internal/generator/templates/validation_types.go.tmpl" || echo "MISSING: internal/generator/templates/validation_types.go.tmpl"
[ -f "internal/generator/queries.go" ] && echo "FOUND: internal/generator/queries.go" || echo "MISSING: internal/generator/queries.go"
[ -f "internal/generator/queries_test.go" ] && echo "FOUND: internal/generator/queries_test.go" || echo "MISSING: internal/generator/queries_test.go"
[ -f "internal/generator/templates/queries.go.tmpl" ] && echo "FOUND: internal/generator/templates/queries.go.tmpl" || echo "MISSING: internal/generator/templates/queries.go.tmpl"
```

Verifying commits exist:

```bash
git log --oneline --all | grep -q "ae1417f" && echo "FOUND: ae1417f" || echo "MISSING: ae1417f"
git log --oneline --all | grep -q "2e11aa4" && echo "FOUND: 2e11aa4" || echo "MISSING: 2e11aa4"
```

**Results:**

All files verified: ✅
- FOUND: internal/generator/validation.go
- FOUND: internal/generator/validation_test.go
- FOUND: internal/generator/templates/validation.go.tmpl
- FOUND: internal/generator/templates/validation_types.go.tmpl
- FOUND: internal/generator/queries.go
- FOUND: internal/generator/queries_test.go
- FOUND: internal/generator/templates/queries.go.tmpl

All commits verified: ✅
- FOUND: ae1417f (Task 1: validation generation)
- FOUND: 2e11aa4 (Task 2: query builder generation)

## Self-Check: PASSED
