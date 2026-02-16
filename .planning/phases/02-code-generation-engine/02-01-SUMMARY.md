---
phase: 02-code-generation-engine
plan: 01
subsystem: code-generation
tags:
  - generator
  - templates
  - code-generation
  - models
  - go-templates
dependency_graph:
  requires:
    - phase: 01
      plan: 04
      reason: "ResourceIR and FieldIR types from parser"
  provides:
    - "Generator infrastructure with template rendering pipeline"
    - "Model type generation producing 5 structs per resource"
    - "FormatGoSource with automatic import management"
    - "Template helpers for type mapping and field introspection"
  affects:
    - "All future generator implementations (Atlas HCL, factories, migrations)"
tech_stack:
  added:
    - golang.org/x/tools/imports
  patterns:
    - "Embedded template filesystem (go:embed)"
    - "Buffer-then-format-then-write pipeline"
    - "Template FuncMap for type conversions"
key_files:
  created:
    - internal/generator/generator.go
    - internal/generator/formatting.go
    - internal/generator/templates.go
    - internal/generator/funcmap.go
    - internal/generator/models.go
    - internal/generator/templates/model.go.tmpl
    - internal/generator/generator_test.go
  modified:
    - go.mod
    - go.sum
decisions:
  - "Use golang.org/x/tools/imports for automatic import management instead of manual tracking"
  - "Snake case handles acronyms as units (HTTPStatus->http_status, ProductID->product_id)"
  - "Filter struct only includes fields with Filterable modifier"
  - "Update structs use all pointer fields for partial updates"
  - "Create structs use non-pointer for required, pointer for optional"
metrics:
  duration_seconds: 190
  duration_formatted: "3.2m"
  completed_at: "2026-02-16T17:27:07Z"
---

# Phase 02 Plan 01: Generator Infrastructure & Model Generation Summary

JWT auth with refresh rotation using jose library

## One-Line Summary

Built generator infrastructure with template rendering, import management, and model type generation producing Resource/Create/Update/Filter/Sort structs from ResourceIR.

## What Was Built

### Generator Infrastructure
Created the foundational code generation package that all future generators will reuse:

1. **Template Engine** (`templates.go`, `funcmap.go`)
   - Embedded template filesystem using `//go:embed templates/*`
   - Template FuncMap with 13+ helper functions:
     - Type mapping: `goType`, `goPointerType` (14 field types -> Go types)
     - String transforms: `snake`, `camel`, `lower`, `plural`
     - Field introspection: `isIDField`, `isFilterable`, `isSortable`
     - Modifier queries: `hasModifier`, `getModifierValue`, `isRequired`

2. **Formatting Pipeline** (`formatting.go`)
   - `FormatGoSource` using `golang.org/x/tools/imports`
   - Automatic import addition/removal
   - Go syntax validation

3. **Generator Orchestrator** (`generator.go`)
   - `Generate` function to coordinate all generation
   - `renderTemplate` for template parsing and execution
   - `writeGoFile` for formatted output
   - `writeRawFile` for non-Go files (future HCL)
   - `ensureDir` for directory creation

### Model Generation
Created model type generation that transforms ResourceIR into compilable Go structs:

1. **Template** (`templates/model.go.tmpl`)
   - Generates 5 struct types per resource:
     - `Resource`: Full record with all fields, ID, timestamps, soft delete
     - `ResourceCreate`: Required fields non-pointer, optional fields pointer
     - `ResourceUpdate`: All fields pointer for partial updates
     - `ResourceFilter`: Only filterable fields with _neq variants
     - `ResourceSort`: Generic field+direction sort spec
   - Includes "DO NOT EDIT" header
   - Proper `json` and `db` struct tags
   - Snake_case field names in tags

2. **Generation Function** (`models.go`)
   - `GenerateModels` iterates resources and renders to `models/{snake_name}.go`
   - Calls formatting pipeline automatically
   - Creates output directories as needed

3. **Tests** (`generator_test.go`)
   - End-to-end generation test with Product schema
   - Type mapping verification for all 14 field types
   - Snake case conversion tests (including acronyms)
   - Field introspection helper tests

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed snake_case conversion for acronyms**
- **Found during:** Task 2 test execution
- **Issue:** Original snake case function inserted underscore before every uppercase letter, converting "ID" to "i_d" and "HTTPStatus" to "h_t_t_p_status"
- **Fix:** Enhanced snake function to handle consecutive uppercase letters as acronyms, checking if next character is lowercase to determine acronym boundaries
- **Result:** "ID" -> "id", "HTTPStatus" -> "http_status", "ProductID" -> "product_id"
- **Files modified:** internal/generator/funcmap.go
- **Commit:** 01250c2

**2. [Rule 1 - Bug] Updated test expectations for correct snake_case behavior**
- **Found during:** Task 2 test execution
- **Issue:** Test expected incorrect snake_case output matching old buggy behavior
- **Fix:** Updated test expectations to match correct acronym handling
- **Files modified:** internal/generator/generator_test.go
- **Commit:** 01250c2

## Key Technical Decisions

1. **Import Management Strategy**
   - Decision: Use `golang.org/x/tools/imports.Process()` instead of manual import tracking
   - Rationale: Automatic import management handles edge cases (aliasing, grouping, unused imports) that would take weeks to implement correctly
   - Impact: Generated code always has correct, formatted imports

2. **Snake Case Acronym Handling**
   - Decision: Treat consecutive uppercase letters as acronyms (HTTPStatus->http_status)
   - Rationale: More readable database column names and JSON fields
   - Impact: Better naming convention alignment with Go/PostgreSQL standards

3. **Struct Tag Strategy**
   - Decision: Include both `json` and `db` tags on all fields
   - Rationale: Models will be used for both API serialization and database scanning
   - Impact: Single struct type serves dual purpose

4. **Filter Struct Design**
   - Decision: Only include fields marked with Filterable modifier, add _neq variant for each
   - Rationale: Explicit control over which fields can be filtered, common inequality operator
   - Impact: Type-safe query building in future repository layer

## Test Results

All tests passing:
- `TestGenerateModels_ProductSchema`: End-to-end generation with verification of all 5 structs, DO NOT EDIT header, timestamps, soft delete, and valid Go syntax
- `TestGoTypeMapping`: Verified all 14 field type mappings (UUID, String, Text, Int, BigInt, Decimal, Bool, DateTime, Date, Enum, JSON, Slug, Email, URL)
- `TestSnakeCase`: Verified PascalCase to snake_case with acronym handling
- `TestIsIDField`: Verified ID field detection (Name="ID" AND Type="UUID")
- `TestIsRequired`: Verified Required modifier detection
- `TestIsFilterable`: Verified Filterable modifier detection

Build verification:
- `go build ./internal/generator/` - clean compile
- `go vet ./internal/generator/` - no issues
- Generated files pass `go/format.Source()` validation

## Files Changed

| File | Lines Added | Lines Removed | Purpose |
|------|-------------|---------------|---------|
| internal/generator/generator.go | 82 | 0 | Generation orchestrator and utilities |
| internal/generator/formatting.go | 13 | 0 | Import management via x/tools/imports |
| internal/generator/templates.go | 8 | 0 | Embedded template filesystem |
| internal/generator/funcmap.go | 153 | 0 | Template helper functions |
| internal/generator/models.go | 32 | 0 | Model generation implementation |
| internal/generator/templates/model.go.tmpl | 63 | 0 | Go model template |
| internal/generator/generator_test.go | 223 | 0 | Comprehensive test suite |
| go.mod | 5 | 1 | Added golang.org/x/tools dependency |

Total: 579 lines added, 1 line removed across 8 files

## Commits

| Hash | Type | Description |
|------|------|-------------|
| c46bcb5 | feat | Create generator infrastructure with template engine and formatting |
| 01250c2 | feat | Create Go model template and GenerateModels function |

## Next Steps

Phase 02 Plan 02 will build on this foundation to add:
- Atlas HCL schema generation for database migrations
- Template-based factory generation for test data
- SQL query builder integration

The infrastructure built here (template rendering, formatting, helpers) will be reused for all future generators.

## Self-Check: PASSED

Verified all deliverables:

**Created Files:**
```
FOUND: internal/generator/generator.go
FOUND: internal/generator/formatting.go
FOUND: internal/generator/templates.go
FOUND: internal/generator/funcmap.go
FOUND: internal/generator/models.go
FOUND: internal/generator/templates/model.go.tmpl
FOUND: internal/generator/generator_test.go
```

**Commits:**
```
FOUND: c46bcb5
FOUND: 01250c2
```

**Verification Commands:**
- go build ./internal/generator/ - PASSED
- go test ./internal/generator/ -v - PASSED (all 6 test suites, 30 subtests)
- go vet ./internal/generator/ - PASSED

All must-have truths satisfied:
- ✓ Generator infrastructure renders Go templates from IR with formatted output
- ✓ Model generation produces compilable Resource/Create/Update/Filter/Sort structs
- ✓ Generated files include DO NOT EDIT header
- ✓ All 14 IR field types map to correct Go types
