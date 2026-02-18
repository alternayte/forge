---
phase: 07-advanced-data-features
plan: 03
subsystem: auth
tags: [tenant, multi-tenant, rls, row-level-security, context, middleware, postgres, atlas, bob]

# Dependency graph
requires:
  - phase: 07-02
    provides: soft delete query mods, partial unique indexes pattern for templates
  - phase: 05-02
    provides: auth middleware pattern used as model for TenantMiddleware
provides:
  - TenantResolver interface with three HTTP-level extraction strategies
  - WithTenant/TenantFromContext context helpers for tenant propagation
  - TenantMiddleware http.Handler wrapper for Chi/stdlib integration
  - TenantMod query method for automatic tenant scoping via Bob
  - Atlas HCL tenant_id column, index, row_level_security block, and tenant isolation policies
  - TenantID field on generated model structs
  - Default test tenant UUID in generated factories with WithTenantID builder
affects:
  - 07-04-PLAN.md
  - 07-05-PLAN.md

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Private struct{} context key type prevents cross-package key collisions"
    - "TenantResolver interface + three strategies (header/subdomain/path) — developer chooses at wire-up"
    - "TenantMiddleware follows same constructor-returns-middleware pattern as auth middlewares"
    - "TenantMod panics (not errors) when tenant missing from context — fast fail, middleware misconfiguration"
    - "RLS policy uses current_setting('app.current_tenant')::uuid — SET LOCAL per transaction"
    - "Application WHERE clause (TenantMod) is primary isolation; RLS is defense-in-depth safety net"

key-files:
  created:
    - internal/auth/tenant.go
  modified:
    - internal/generator/templates/model.go.tmpl
    - internal/generator/templates/factory.go.tmpl
    - internal/generator/templates/queries.go.tmpl
    - internal/generator/templates/atlas_schema.hcl.tmpl

key-decisions:
  - "TenantMiddleware panics on missing context tenant rather than returning error — fast fail signals middleware misconfiguration clearly"
  - "PathTenantResolver supports both /tenants/{uuid}/... and /{uuid}/... path patterns for flexibility"
  - "RLS policy uses current_setting('app.current_tenant')::uuid — middleware runs SET LOCAL app.current_tenant at transaction start"
  - "forgeauth alias used for auth import in queries template to avoid collision with gen/auth package"

patterns-established:
  - "Tenant isolation: two-layer defense — application WHERE via TenantMod (primary, uses index) + PostgreSQL RLS (safety net)"
  - "Template conditional pattern: {{- if .Options.TenantScoped}} guards all tenant-specific generated code"

requirements-completed: [TENANT-01, DATA-05, TENANT-02, TENANT-03, TENANT-04]

# Metrics
duration: 6min
completed: 2026-02-18
---

# Phase 7 Plan 03: Tenant Isolation Summary

**Multi-tenant isolation via TenantResolver/TenantMiddleware runtime + automatic tenant_id WHERE clause (TenantMod) + PostgreSQL RLS policies generated for TenantScoped Atlas schemas**

## Performance

- **Duration:** 6 min
- **Started:** 2026-02-18T07:40:40Z
- **Completed:** 2026-02-18T07:46:50Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Created `internal/auth/tenant.go` with TenantResolver interface, three resolver implementations (header/subdomain/path), TenantMiddleware, and WithTenant/TenantFromContext context helpers
- Added TenantMod query method to queries.go.tmpl that reads tenant from context via forgeauth.TenantFromContext and applies WHERE tenant_id = $1
- Added tenant_id column, tenant_id index, row_level_security block (enabled + enforced), and two RLS policy blocks to atlas_schema.hcl.tmpl for TenantScoped resources
- Added TenantID field to generated model structs and default test tenant UUID plus WithTenantID builder to generated factories

## Task Commits

Each task was committed atomically:

1. **Task 1: Create tenant runtime and update model/factory templates** - `f47fe28` (feat)
2. **Task 2: Add tenant query scoping and Atlas RLS policies** - `acf63dd` (feat)

**Plan metadata:** `[pending]` (docs: complete plan)

## Files Created/Modified
- `internal/auth/tenant.go` - TenantResolver interface, HeaderTenantResolver, SubdomainTenantResolver, PathTenantResolver, TenantMiddleware, WithTenant, TenantFromContext
- `internal/generator/templates/model.go.tmpl` - TenantID uuid.UUID field when TenantScoped (after ID, before range .Fields)
- `internal/generator/templates/factory.go.tmpl` - Default test TenantID (00000000-0000-0000-0000-000000000001) and WithTenantID builder method when TenantScoped
- `internal/generator/templates/queries.go.tmpl` - TenantMod method with forgeauth import guard, context import guard
- `internal/generator/templates/atlas_schema.hcl.tmpl` - tenant_id column, tenant_id index, row_level_security block, tenant_isolation SELECT policy, tenant_isolation_mod ALL policy

## Decisions Made
- TenantMod panics (not returns error) when tenant missing from context — panic is appropriate for developer misconfiguration, not runtime user error
- PathTenantResolver supports both /tenants/{uuid}/... and /{uuid}/... to cover common URL patterns
- Used `forgeauth` alias for the auth import in queries template to prevent collision with any gen/auth package in the generated project
- RLS policy uses `current_setting('app.current_tenant')::uuid` — caller middleware runs `SET LOCAL app.current_tenant = $1` at transaction start (per prior user decision)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Tenant isolation fully wired: runtime helpers, query scoping, Atlas schema generation, model/factory generation
- Plan 04 can build on TenantScoped IR field and generated tenant_id column for relationship or auditing features
- Plan 05 can use TenantMiddleware in the assembled HTTP server setup

---
*Phase: 07-advanced-data-features*
*Completed: 2026-02-18*
