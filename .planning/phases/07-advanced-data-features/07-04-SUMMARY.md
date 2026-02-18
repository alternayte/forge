---
phase: 07-advanced-data-features
plan: 04
subsystem: auth
tags: [rbac, permissions, field-visibility, actions, context, code-generation]

# Dependency graph
requires:
  - phase: 07-advanced-data-features plan 01
    provides: hasPermission, permissionRoles, hasAnyVisibility, hasAnyPermission funcmap helpers registered in BuildFuncMap
  - phase: 04-action-layer-error-handling
    provides: errors.Forbidden and DefaultActions pattern that actions.go.tmpl extends

provides:
  - internal/auth/context.go with WithUserRole, UserFromContext, RoleFromContext context helpers
  - actions.go.tmpl: checkPermission helper function for CRUD-level role enforcement
  - actions.go.tmpl: conditional permission checks at top of all CRUD methods (List, Get, Create, Update, Delete, Restore)
  - actions.go.tmpl: roleFilter method for field-level visibility stripping based on user role
  - actions.go.tmpl: RoleFilterList method and interface entry for handler use
affects:
  - phase: 07-advanced-data-features plan 05
  - api handlers (gen/handlers) that call RoleFilterList before returning responses
  - html handlers (gen/html) that call RoleFilterList for view data

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Role-based permission checks at action layer (not handler layer) — single enforcement point for HTML and API
    - Field visibility as map[string]any stripping (not struct tags) — key omission is the contract
    - Empty role == visible — admin/dev context sees all fields; role-constrained context sees filtered set

key-files:
  created:
    - internal/auth/context.go
  modified:
    - internal/generator/templates/actions.go.tmpl

key-decisions:
  - "Permission checks live in generated action methods, not handlers — enforced at action layer for both HTML and API handlers"
  - "roleFilter returns map[string]any (not a filtered struct) so invisible field keys are truly absent from JSON"
  - "checkPermission is generated per-resource-file (not a shared helper) to keep generated code self-contained"
  - "Restore method uses delete permission — inverse operation shares the same authorization gate"

patterns-established:
  - "Template conditional guards: {{- if hasAnyPermission .Options}} wraps checkPermission function; {{- if hasPermission .Options 'operation'}} wraps each method's check"
  - "Context helpers use private struct key types (userContextKey, roleContextKey) to prevent collisions"
  - "forgeauth import is conditional on or (hasAnyPermission .Options) (hasAnyVisibility .Fields) — zero overhead for resources without auth features"

requirements-completed: [AUTH-05, AUTH-06]

# Metrics
duration: 2min
completed: 2026-02-18
---

# Phase 07 Plan 04: Permission Checks and Field Visibility Stripping Summary

**Role-based CRUD permission enforcement and field-level visibility filtering in generated actions, using context-stored role against schema-defined allowed-role lists**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-18T07:40:43Z
- **Completed:** 2026-02-18T07:42:41Z
- **Tasks:** 2
- **Files modified:** 2 (1 created, 1 modified)

## Accomplishments
- Created `internal/auth/context.go` with `WithUserRole`, `UserFromContext`, `RoleFromContext` for clean role extraction from request context
- Updated `actions.go.tmpl` with conditional `checkPermission` helper that reads role from context and returns `errors.Forbidden` for unauthorized roles
- Added conditional permission checks at the top of every CRUD method (List, Get, Create, Update, Delete, Restore) using `hasPermission` funcmap
- Added `roleFilter` method that returns `map[string]any` with `Visibility`-guarded fields omitted when role does not match
- Added `RoleFilterList` convenience method and interface entry for handler consumption
- Verified all four required funcmap helpers (`hasPermission`, `permissionRoles`, `hasAnyVisibility`, `hasAnyPermission`) present from Plan 01

## Task Commits

Each task was committed atomically:

1. **Task 1: Create auth context helpers and add permission checks to actions** - `3c2944d` (feat)
2. **Task 2: Add field-level visibility stripping to actions** - `3c2944d` (feat, included in Task 1 commit — both tasks modify actions.go.tmpl and were implemented as a complete unit)

**Plan metadata:** (pending docs commit)

## Files Created/Modified
- `internal/auth/context.go` - WithUserRole/UserFromContext/RoleFromContext context helpers for reading auth state from Go context
- `internal/generator/templates/actions.go.tmpl` - Permission checks in all CRUD methods; checkPermission helper; roleFilter and RoleFilterList for field-level visibility stripping

## Decisions Made
- Permission checks live in generated action methods, not handlers — enforced at action layer for both HTML and API handlers without duplication
- `roleFilter` returns `map[string]any` so invisible field keys are truly absent from the JSON response (not just zero-valued)
- `checkPermission` is generated per-resource-file (not a shared forge helper) to keep generated code self-contained and importable without forge dependencies
- Restore method uses "delete" permission gate — the inverse operation shares the same authorization requirement

## Deviations from Plan

None - plan executed exactly as written. All four funcmap helpers were confirmed present from Plan 01 as expected. Both tasks implemented in a single template write since they both modify `actions.go.tmpl`.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Auth context helpers ready for middleware to populate (WithUserRole called after bearer token validation)
- Permission checks in generated actions will enforce AUTH-05 once resources define Permission() in their schema
- Field visibility stripping (AUTH-06) ready for handler integration — call `actions.RoleFilterList(role, items)` before returning list responses
- Plan 05 (relationship loading) can proceed without dependencies on this plan

---
*Phase: 07-advanced-data-features*
*Completed: 2026-02-18*
