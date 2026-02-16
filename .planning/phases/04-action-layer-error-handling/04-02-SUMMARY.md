---
phase: 04-action-layer-error-handling
plan: 02
subsystem: action-layer
tags: [actions, business-logic, crud, generation, interfaces]
dependency_graph:
  requires: [error-handling, validation-generation, query-generation, model-generation]
  provides: [action-interfaces, default-actions, action-registry]
  affects: [html-handlers, api-handlers, handler-generation]
tech_stack:
  added: [action-pattern, embedding-pattern]
  patterns: [interface-segregation, dependency-injection, explicit-registration]
key_files:
  created:
    - internal/generator/templates/actions_types.go.tmpl
    - internal/generator/templates/actions.go.tmpl
    - internal/generator/actions.go
    - internal/generator/actions_test.go
  modified: []
decisions:
  - choice: "Per-resource Actions interface with 5 CRUD methods (List, Get, Create, Update, Delete)"
    rationale: "Defines clear contract that both HTML and API handlers call. Prevents business logic duplication between dual interfaces"
    alternatives: ["Separate interfaces per operation", "Single generic CRUD interface"]
  - choice: "DefaultActions implementation calls generated validation directly (not validator interfaces)"
    rationale: "Simpler embedding override pattern. Developers embed DefaultActions and override entire methods for custom behavior. Validator interfaces kept as optional advanced pattern documented in comments"
    alternatives: ["Type assertion to check for validator interfaces in DefaultActions methods"]
  - choice: "Placeholder TODO comments for Bob query execution"
    rationale: "Action interface contracts and wiring are ready. Actual query execution depends on Bob integration layer that will be built later. Generated code compiles and maintains correct type signatures"
    alternatives: ["Wait for Bob integration before generating actions", "Generate with stub implementations"]
  - choice: "Registry uses explicit Register/Get methods (no init() magic)"
    rationale: "Application explicitly wires actions at startup. Visible, testable, no hidden global state. Follows Go best practices over framework magic"
    alternatives: ["Auto-registration via init()", "Reflection-based discovery"]
  - choice: "DB interface with pgx types (not generic interface{})"
    rationale: "Generated code always runs with pgx. Using concrete pgx types provides better compile-time safety and avoids type assertions. No coupling concern since this is generated output, not library code"
    alternatives: ["Generic interface{} for args parameter", "sql.DB interface"]
metrics:
  duration_seconds: 111
  completed_date: 2026-02-16T21:12:07Z
---

# Phase 04 Plan 02: Action Interface and Default Implementation Generation Summary

**One-liner:** Generated per-resource Actions interfaces and DefaultActions implementations with validation wiring, error mapping, and explicit Registry for shared business logic layer

## What Was Built

Created the action generation layer that produces CRUD interfaces and default implementations for each resource:

1. **Action Types Template** (`actions_types.go.tmpl`):
   - `DB` interface with pgx methods: QueryRow, Query, Exec (matches pgx pool interface)
   - `CreateValidator` interface for optional custom create validation
   - `UpdateValidator` interface for optional custom update validation
   - `Registry` struct with `actions map[string]any` for explicit action registration
   - `NewRegistry()` constructor initializing empty registry
   - `Register(name, action)` method for storing action implementations
   - `Get(name)` method returning action by resource name
   - `GetTyped[T any](registry, name)` generic helper for type-safe retrieval

2. **Per-Resource Actions Template** (`actions.go.tmpl`):
   - `{Resource}Actions` interface with 5 methods:
     - `List(ctx, filter, sort, page, pageSize)` - retrieves filtered/sorted/paginated results
     - `Get(ctx, id)` - retrieves single record by UUID
     - `Create(ctx, input)` - creates after validation
     - `Update(ctx, id, input)` - updates after validation
     - `Delete(ctx, id)` - removes by UUID
   - `Default{Resource}Actions` struct with `DB` field
   - Implementation methods that:
     - Call generated validation functions (`validation.Validate{Resource}Create/Update`)
     - Wrap validation errors via `errors.NewValidationError`
     - Call query helper functions (`queries.{Resource}FilterMods`, `queries.{Resource}SortMod`)
     - Return forge.Error types (`errors.NotFound`, `errors.InternalError`)
     - Include TODO placeholders for Bob query execution (interface ready, execution pending)

3. **Generator Function** (`actions.go`):
   - `GenerateActions(resources, outputDir, projectModule)` following validation.go pattern
   - Creates `gen/actions/` directory
   - Renders `actions_types.go.tmpl` with ProjectModule data → `gen/actions/types.go`
   - Iterates resources, renders `actions.go.tmpl` per resource → `gen/actions/{snake_name}.go`
   - Uses existing `writeGoFile` for formatting and automatic import management
   - Follows exact structure of `GenerateValidation` for consistency

4. **Comprehensive Tests** (`actions_test.go`):
   - `TestGenerateActions` - end-to-end test verifying:
     - types.go contains: DB interface, CreateValidator, UpdateValidator, Registry, NewRegistry, Register, Get, GetTyped
     - product.go contains: ProductActions interface, all 5 method signatures, DefaultProductActions struct, all method implementations
     - Validation wiring: `validation.ValidateProductCreate(input)`, `errors.NewValidationError(valErrs)`
     - Error wiring: `errors.NotFound`, `errors.InternalError`
     - Query wiring: `queries.ProductFilterMods(filter)`, `queries.ProductSortMod(sort)`
   - `TestGenerateActions_MultipleResources` - verifies separate file per resource

## How It Works

**Generation Flow:**
1. `GenerateActions` called with parsed resources, output directory, project module
2. Creates `gen/actions/` directory via `ensureDir`
3. Renders `actions_types.go.tmpl` with ProjectModule → writes `types.go` (shared types)
4. For each resource:
   - Embeds ResourceIR + ProjectModule in template data struct
   - Renders `actions.go.tmpl` → writes `{snake_name}.go`
   - Uses snake case for filename (Product → product.go, ProductCategory → product_category.go)

**Runtime Usage Pattern:**
```go
// App startup - explicit registration
registry := actions.NewRegistry()
registry.Register("product", &actions.DefaultProductActions{DB: pool})
registry.Register("category", &actions.DefaultCategoryActions{DB: pool})

// Handler retrieves actions
productActions, ok := actions.GetTyped[actions.ProductActions](registry, "product")

// Call business logic
product, err := productActions.Create(ctx, models.ProductCreate{Name: "Laptop", Price: 999.99})
if err != nil {
    // err is *errors.Error with Status, Code, Message, Detail, Err
    return err
}

// Custom actions via embedding
type CustomProductActions struct {
    actions.DefaultProductActions
    Cache *redis.Client
}

func (a *CustomProductActions) Get(ctx context.Context, id uuid.UUID) (*models.Product, error) {
    // Check cache first
    if cached := a.Cache.Get(ctx, id); cached != nil {
        return cached, nil
    }
    // Fall back to default
    return a.DefaultProductActions.Get(ctx, id)
}
```

**Validation Integration:**
- Create/Update methods call `validation.Validate{Resource}Create/Update(input)`
- Check `valErrs.HasErrors()` → wrap with `errors.NewValidationError(valErrs)` (returns 422)
- Validation errors flow through forge.Error type with consistent structure

**Query Integration:**
- List method calls `queries.{Resource}FilterMods(filter)` and `queries.{Resource}SortMod(sort)`
- Query mods ready to pass to Bob query builder (actual execution pending Bob integration)
- Placeholder returns empty slice + 0 count to maintain compilable interface

## Deviations from Plan

None - plan executed exactly as written.

## Key Learnings

1. **Static + Dynamic Template Pattern**:
   - `actions_types.go.tmpl` is semi-static (takes ProjectModule but no resource-specific data)
   - `actions.go.tmpl` is per-resource (takes ResourceIR + ProjectModule)
   - Same pattern as validation: types file generated once, resource files generated per resource

2. **Embedding Pattern Over Type Assertions**:
   - Original plan considered checking `any(a).(CreateValidator)` inside DefaultActions methods
   - Problem: DefaultActions methods receive `*Default{Resource}Actions`, not the embedding type
   - Solution: DefaultActions always calls generated validation. Developers override entire methods for custom behavior
   - CreateValidator/UpdateValidator interfaces kept in types.go as optional advanced pattern for documentation

3. **Interface Contracts Before Implementation**:
   - Generated methods return correct types with TODO placeholders for actual DB operations
   - Enables downstream development (handler generation) without blocking on Bob integration
   - `_ = filterMods` and similar patterns prevent unused variable errors while maintaining compilable code

4. **Explicit vs Magic Registration**:
   - Registry pattern with explicit Register/Get calls at app startup
   - No init() functions, no reflection-based auto-discovery
   - Clear, testable, visible wiring that developers control

5. **Generic Type-Safe Retrieval**:
   - `GetTyped[T any](registry, name)` provides compile-time type safety
   - Returns zero value if resource not found or type doesn't match
   - Enables: `actions := actions.GetTyped[actions.ProductActions](reg, "product")`

## Testing Evidence

```bash
$ go test ./internal/generator/ -run TestGenerateActions -v
=== RUN   TestGenerateActions
--- PASS: TestGenerateActions (0.02s)
=== RUN   TestGenerateActions_MultipleResources
--- PASS: TestGenerateActions_MultipleResources (0.03s)
PASS
ok  	github.com/forge-framework/forge/internal/generator	0.521s

$ go build ./internal/generator/
# Build successful

$ go vet ./internal/generator/
# No issues
```

**Test Coverage:**
- types.go generation: DB interface with all 3 methods, CreateValidator, UpdateValidator, Registry struct, all registry methods
- Per-resource generation: Actions interface, all 5 method signatures, DefaultActions struct with DB field
- Method implementations: List, Get, Create, Update, Delete all present
- Validation wiring: calls to Validate{Resource}Create and Validate{Resource}Update
- Error wiring: errors.NotFound, errors.NewValidationError, errors.InternalError
- Query wiring: queries.{Resource}FilterMods and queries.{Resource}SortMod calls
- Multiple resource handling: separate files for product.go and category.go

## Next Steps

**Plan 03** (Wire Actions into Main Generator):
- Add `GenerateActions` call to main `Generate()` function in generator.go
- Ensure actions generation runs after models, validation, queries, errors
- Update `forge generate` command output to show actions/ directory statistics

**Phase 05** (HTML Handler Generation):
- HTML handlers call `{Resource}Actions` interface for business logic
- No direct DB queries in handlers - all through action layer
- Datastar templates render results from action method calls

**Phase 06** (API Handler Generation):
- Huma API handlers call same `{Resource}Actions` interface
- Single business logic implementation shared by both HTML and API
- Demonstrates core value: action layer prevents duplication

## Files Changed

**Created:**
- `internal/generator/templates/actions_types.go.tmpl` (79 lines) - Shared types: DB interface, validators, Registry
- `internal/generator/templates/actions.go.tmpl` (107 lines) - Per-resource Actions interface + DefaultActions
- `internal/generator/actions.go` (59 lines) - GenerateActions function following validation.go pattern
- `internal/generator/actions_test.go` (293 lines) - Comprehensive tests for types.go and per-resource generation

**Modified:** None

## Commits

- `5cffed2`: feat(04-02): add action type templates
- `ffacf0a`: feat(04-02): add GenerateActions function and tests

## Self-Check: PASSED

All generated files verified:
```bash
$ ls internal/generator/templates/actions*.tmpl
internal/generator/templates/actions.go.tmpl
internal/generator/templates/actions_types.go.tmpl

$ ls internal/generator/actions*
internal/generator/actions.go
internal/generator/actions_test.go
```

All commits verified:
```bash
$ git log --oneline --all | grep -E "(5cffed2|ffacf0a)"
ffacf0a feat(04-02): add GenerateActions function and tests
5cffed2 feat(04-02): add action type templates
```
