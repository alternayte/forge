---
phase: 01-foundation-schema-dsl
plan: 01
subsystem: schema-dsl
tags: [foundation, dsl, schema, api-design]
dependencies:
  requires: []
  provides: [schema-definition-api]
  affects: []
tech_stack:
  added: [go-1.23]
  patterns: [fluent-api, builder-pattern, marker-interface]
key_files:
  created:
    - internal/schema/schema.go
    - internal/schema/field_type.go
    - internal/schema/modifier.go
    - internal/schema/field.go
    - internal/schema/relationship.go
    - internal/schema/option.go
    - internal/schema/definition.go
    - internal/schema/schema_test.go
    - go.mod
    - go.sum
  modified: []
decisions:
  - title: "Fluent method chaining API for schema definition"
    rationale: "Provides intuitive developer experience with type-safe chaining like schema.String('Title').Required().MaxLen(200)"
    alternatives: ["struct-based config", "functional options"]
    chosen: "fluent-chaining"
  - title: "SchemaItem marker interface for variadic Define() args"
    rationale: "Allows flat list of fields, relationships, options at same level while maintaining type safety"
    alternatives: ["separate parameters", "builder methods"]
    chosen: "marker-interface"
  - title: "Unexported schemaItem() marker method"
    rationale: "Prevents external types from implementing SchemaItem, maintaining package control"
    alternatives: ["exported interface", "no interface"]
    chosen: "unexported-marker"
metrics:
  duration_minutes: 2.5
  completed_at: "2026-02-16T15:33:24Z"
  tasks_completed: 2
  files_created: 10
  tests_added: 5
---

# Phase 01 Plan 01: Schema Definition DSL Summary

**One-liner:** Fluent Go API for defining resource schemas with 14 field types, relationships, options, and timestamps

## Objective Achievement

Created the complete schema definition DSL package (`internal/schema/`) that developers use to define resource schemas with fluent method chaining. This is the foundation of Forge's "schema as single source of truth" philosophy.

**Must-haves delivered:**
- ✅ Developer can define resources with all 14 field types using fluent chaining
- ✅ Developer can apply all field modifiers (Required, MaxLen, MinLen, etc.) via chaining
- ✅ Developer can define BelongsTo, HasMany, HasOne, ManyToMany relationships inline
- ✅ Developer can enable SoftDelete, Auditable, TenantScoped, Searchable options per resource
- ✅ Developer can call Timestamps() to auto-add created_at and updated_at
- ✅ Developer can define Enum fields with specific allowed values and Default
- ✅ schema.Define() takes variadic args with fields, relationships, options at same level

## Tasks Completed

### Task 1: Initialize Go module and create schema type foundations

**Commit:** `34611b6`

**What was built:**
- Initialized Go module as `github.com/forge-framework/forge`
- Created `SchemaItem` interface with unexported `schemaItem()` marker method
- Defined `FieldType` enum with all 14 types (UUID, String, Text, Int, BigInt, Decimal, Bool, DateTime, Date, Enum, JSON, Slug, Email, URL)
- Defined `OnDeleteAction` enum (Cascade, SetNull, Restrict, NoAction)
- Created `Field` struct with fluent modifier methods
- Created all 14 field type constructor functions
- Defined `ModifierType` enum and `Modifier` struct with 15 modifiers
- Implemented fluent chaining for Required(), Optional(), PrimaryKey(), MaxLen(), MinLen(), Sortable(), Filterable(), Searchable(), Unique(), Index(), Default(), Immutable(), Label(), Placeholder(), Help()

**Key files:**
- `internal/schema/schema.go` - Package entry point with interfaces and enums
- `internal/schema/field_type.go` - Constructor functions for all 14 field types
- `internal/schema/modifier.go` - Modifier types and constants
- `internal/schema/field.go` - Field struct with fluent API methods
- `go.mod` - Go module initialization

**Verification:** `go build ./internal/schema/` compiled with zero errors

### Task 2: Create relationships, options, timestamps, and Define function

**Commit:** `44e697b`

**What was built:**
- Created `Relationship` struct with `RelationType` enum (BelongsTo, HasMany, HasOne, ManyToMany)
- Added fluent `Optional()` and `OnDelete()` methods on Relationship
- Created `Option` struct with `OptionType` enum (SoftDelete, Auditable, TenantScoped, Searchable)
- Added `Timestamps()` function returning `TimestampsItem` marker
- Created `Definition` struct with getters for name, fields, relationships, options, hasTimestamps
- Implemented `Define()` function with variadic `SchemaItem` args that categorizes items by type
- Added comprehensive tests validating Product schema example, all field types, all relationships, all options, and fluent chaining

**Key files:**
- `internal/schema/relationship.go` - Relationship types with fluent API
- `internal/schema/option.go` - Resource-level options and Timestamps marker
- `internal/schema/definition.go` - Define function and Definition type
- `internal/schema/schema_test.go` - Comprehensive test suite

**Verification:** `go test ./internal/schema/` passed with 5 test functions covering:
- Product schema example from plan (6 fields, 2 relationships, 1 option, timestamps)
- All 14 field types
- All 4 relationship types
- All 4 options
- Fluent chaining behavior

## Deviations from Plan

None - plan executed exactly as written.

## Example Usage

The complete API now supports the exact usage pattern from the plan:

```go
var Product = schema.Define("Product",
    schema.SoftDelete(),
    schema.UUID("ID").PrimaryKey(),
    schema.String("Title").Required().MaxLen(200).Label("Product Title"),
    schema.Text("Description").Help("Full product description"),
    schema.Decimal("Price").Required().Filterable().Sortable(),
    schema.Enum("Status", "draft", "published", "archived").Default("draft"),
    schema.Bool("Featured").Default(false),
    schema.BelongsTo("Category", "categories").Optional().OnDelete(schema.SetNull),
    schema.HasMany("Reviews", "reviews"),
    schema.Timestamps(),
)
```

## Technical Decisions

**1. Fluent method chaining API**
- Chosen pattern: Each modifier method returns `*Field` for continued chaining
- Enables intuitive, readable schema definitions
- Type-safe at compile time
- Example: `schema.String("Title").Required().MaxLen(200).Searchable()`

**2. SchemaItem marker interface**
- Unexported `schemaItem()` method prevents external implementations
- Allows `Define()` to accept variadic args of different types (Field, Relationship, Option, TimestampsItem)
- Enables flat, order-independent item list instead of nested builder calls
- Type switch inside `Define()` categorizes items correctly

**3. Enum values stored on Field**
- `Enum()` constructor takes variadic `values` and stores on `Field.enumValues`
- Makes enum values available for AST parsing and validation generation
- Accessed via `Field.EnumValues()` getter

**4. Default OnDeleteAction**
- All relationships default to `NoAction` if not specified
- Developer must explicitly call `.OnDelete(schema.Cascade)` if desired
- Prevents accidental cascading deletes

## Key Artifacts Delivered

| Artifact | Purpose | Interface |
|----------|---------|-----------|
| `SchemaItem` interface | Common type for Define() args | `schemaItem()` marker |
| `Field` struct | Field definition with modifiers | 14 constructors + 15 fluent methods |
| `Relationship` struct | Relationship definition | 4 constructors + Optional/OnDelete |
| `Option` struct | Resource-level options | 4 constructors (SoftDelete, Auditable, TenantScoped, Searchable) |
| `TimestampsItem` | Timestamps marker | `Timestamps()` constructor |
| `Definition` struct | Complete schema definition | `Define()` function + 5 getters |

## Verification Results

**Build:** ✅ `go build ./internal/schema/` - zero errors

**Tests:** ✅ `go test ./internal/schema/` - all 5 test functions passed
- TestProductSchema: Validates complete Product example
- TestAllFieldTypes: Confirms all 14 field types work
- TestAllRelationshipTypes: Confirms all 4 relationship types work
- TestAllOptions: Confirms all 4 options work
- TestFluentChaining: Validates modifier chaining behavior

**Coverage:**
- 14/14 field type constructors tested
- 4/4 relationship type constructors tested
- 4/4 option constructors tested
- Timestamps() marker tested
- Define() categorization tested
- Fluent chaining tested

## Next Steps

This plan delivers the schema definition API. The next plan (01-02) should implement the AST parser that reads schema.go files and extracts these definitions without executing Define().

**Dependencies satisfied for:**
- AST parsing (can parse Define() calls with SchemaItem args)
- Code generation (Definition provides all metadata needed)
- Migration generation (field types and modifiers map to SQL types and constraints)

## Self-Check: PASSED

**Files verified:**
```
FOUND: internal/schema/schema.go
FOUND: internal/schema/field_type.go
FOUND: internal/schema/modifier.go
FOUND: internal/schema/field.go
FOUND: internal/schema/relationship.go
FOUND: internal/schema/option.go
FOUND: internal/schema/definition.go
FOUND: internal/schema/schema_test.go
FOUND: go.mod
FOUND: go.sum
```

**Commits verified:**
```
FOUND: 34611b6
FOUND: 44e697b
```

All claimed files and commits exist.
