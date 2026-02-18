---
phase: 07-advanced-data-features
plan: 01
subsystem: schema
tags: [schema-dsl, parser, ir, funcmap, visibility, mutability, permissions, eager-loading]

requires:
  - phase: 06-hypermedia-ui-generation
    provides: established schema DSL, parser IR, and funcmap patterns this plan extends

provides:
  - ModVisibility and ModMutability modifier type constants in modifier.go
  - Field.Visibility() and Field.Mutability() fluent methods in field.go
  - PermissionItem type with Permission() constructor in permission.go
  - PermissionsIR type (map[string][]string) and Permissions field on ResourceOptionsIR in ir.go
  - Eager bool field on RelationshipIR in ir.go
  - Visibility, Mutability, Eager recognized as modifiers in extractor.go isModifierMethod()
  - isPermissionType() and extractPermission() functions in extractor.go
  - Permission branch in extractSchemaDefinition() populating PermissionsIR
  - Eager handling in extractRelationship() setting RelationshipIR.Eager
  - hasPermission, permissionRoles, hasAnyVisibility, hasAnyPermission, hasAuditableResource, hasTenantScopedResource funcmap helpers

affects:
  - 07-02 soft delete generation (uses SoftDelete option pattern)
  - 07-03 tenant scoping generation (uses TenantScoped + PermissionsIR)
  - 07-04 permissions/audit generation (uses PermissionsIR, Auditable, hasPermission funcmap helpers)
  - 07-05 eager loading generation (uses RelationshipIR.Eager)

tech-stack:
  added: []
  patterns:
    - "PermissionsIR as map[string][]string — operation name keys, role string slices as values"
    - "Permission() as standalone SchemaItem (not an Option) passed directly to schema.Define()"
    - "Eager() as a modifier on relationship chains (not a field modifier)"
    - "extractPermission() extracts variadic role strings from schema.Permission() AST call"

key-files:
  created:
    - internal/schema/permission.go
  modified:
    - internal/schema/modifier.go
    - internal/schema/field.go
    - internal/parser/ir.go
    - internal/parser/extractor.go
    - internal/generator/funcmap.go

key-decisions:
  - "Permission() is a SchemaItem (not an Option) — kept separate so permission rules don't pollute resource-level boolean flags"
  - "PermissionsIR is map[string][]string — operation-keyed for O(1) lookup in templates; multiple permissions per resource supported"
  - "Eager is a modifier on RelationshipIR (not a standalone type) — consistent with Optional and OnDelete modifier pattern"

patterns-established:
  - "Phase 7 modifier pattern: add constant to modifier.go, add method to field.go, add to isModifierMethod() map, handle in extractRelationship() or extractField() as needed"
  - "Phase 7 SchemaItem pattern: new file in schema/, implement schemaItem() interface, add isXType() + extractX() to extractor.go, add branch in extractSchemaDefinition()"
  - "Phase 7 funcmap pattern: add helpers under // Phase 7 comment section in funcmap.go, register in BuildFuncMap()"

requirements-completed: [SCHEMA-07, SCHEMA-08]

duration: 2min
completed: 2026-02-18
---

# Phase 7 Plan 01: Advanced Data Features Foundation Summary

**Visibility/Mutability field modifiers, Permission resource-level rules, and Eager relationship loading wired through schema DSL, parser IR, extractor, and template funcmap**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-18T07:31:51Z
- **Completed:** 2026-02-18T07:33:46Z
- **Tasks:** 2
- **Files modified:** 6 (5 modified, 1 created)

## Accomplishments
- Extended schema DSL with Visibility() and Mutability() field methods and ModVisibility/ModMutability constants
- Created permission.go with Permission() SchemaItem constructor that maps operations to allowed roles
- Extended IR with PermissionsIR type, Permissions on ResourceOptionsIR, and Eager on RelationshipIR
- Wired parser extraction: Visibility/Mutability/Eager recognized as modifiers; Permission() extracted as PermissionsIR
- Added 6 Phase 7 template funcmap helpers: hasPermission, permissionRoles, hasAnyVisibility, hasAnyPermission, hasAuditableResource, hasTenantScopedResource
- All 12 existing parser tests pass without regression

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Visibility, Mutability, Permission, and Eager to schema DSL and IR** - `6b2cd5f` (feat)
2. **Task 2: Wire parser extraction and add funcmap template helpers** - `57f5439` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `internal/schema/modifier.go` - Added ModVisibility and ModMutability constants
- `internal/schema/field.go` - Added Visibility() and Mutability() fluent methods
- `internal/schema/permission.go` - NEW: PermissionItem struct and Permission() constructor
- `internal/parser/ir.go` - Added PermissionsIR type, Permissions on ResourceOptionsIR, Eager on RelationshipIR
- `internal/parser/extractor.go` - Added isPermissionType(), extractPermission(), Permission branch in extractSchemaDefinition(), Eager in extractRelationship(), Visibility/Mutability/Eager in isModifierMethod()
- `internal/generator/funcmap.go` - Added hasPermission, permissionRoles, hasAnyVisibility, hasAnyPermission, hasAuditableResource, hasTenantScopedResource helpers

## Decisions Made
- Permission() is a SchemaItem (not an Option) — permission rules are distinct from boolean feature flags like SoftDelete/Auditable
- PermissionsIR is map[string][]string — operation-keyed for O(1) lookup in templates; supports multiple permission rules per resource
- Eager is a modifier on RelationshipIR (not a standalone type) — consistent with Optional and OnDelete modifier pattern

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- IR foundation complete: all Phase 7 features (soft delete, tenant scoping, permissions, audit, eager loading) can now read Visibility/Mutability/Eager modifiers and PermissionsIR from the IR
- Downstream plans (07-02 through 07-05) can reference these IR fields in template conditionals
- hasAuditableResource and hasTenantScopedResource funcmap helpers ready for cross-resource template logic

---
*Phase: 07-advanced-data-features*
*Completed: 2026-02-18*

## Self-Check: PASSED

All files verified:
- FOUND: internal/schema/permission.go
- FOUND: internal/schema/modifier.go (ModVisibility present)
- FOUND: internal/schema/field.go
- FOUND: internal/parser/ir.go (PermissionsIR present)
- FOUND: internal/parser/extractor.go (Visibility in isModifierMethod present)
- FOUND: internal/generator/funcmap.go (hasPermission present)

All commits verified:
- FOUND: 6b2cd5f (Task 1)
- FOUND: 57f5439 (Task 2)
