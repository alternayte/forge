# Phase 7: Advanced Data Features - Context

**Gathered:** 2026-02-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Schema annotations that control data access behavior: soft delete, field-level visibility/mutability, CRUD-level permissions, multi-tenancy with RLS, and audit logging. Developer declares these features on resources and Forge generates the correct queries, middleware, migrations, and UI/API behavior. Relationships (BelongsTo, HasMany, HasOne) with preloading are also in scope.

</domain>

<decisions>
## Implementation Decisions

### Soft delete behavior
- Deleted records disappear from normal UI — only admin role can see and restore via separate view
- No automatic cascade on soft delete — developer explicitly handles cascade per-relationship in custom action overrides (avoids implicit data loss bugs)
- No hard delete — soft delete is final state, records stay in DB permanently (developer writes raw SQL if needed)
- Partial unique indexes (WHERE deleted_at IS NULL) — unique among active records only, allowing re-creation after soft-delete

### Permission & role model
- Roles defined as schema-level constants (e.g., Role("admin"), Role("editor")) — compile-time known, generated into permission checks
- Permission denial is contextual: API returns 403 Forbidden, HTML redirects to 'not authorized' page or hides the action entirely
- Invisible fields omitted from response entirely (key doesn't appear in API JSON or HTML template) — response shape changes per role
- Permissions are role-only, not owner-aware — ownership-scoped logic is a custom action override if needed

### Tenant isolation model
- Tenant context propagated via middleware: extracts tenant from auth token/session, sets in context.Context — generated code reads from context
- Always single-tenant scoping — every query scoped to exactly one tenant, no cross-tenant admin access
- Defense in depth: application-level WHERE clause for performance + PostgreSQL RLS policies as safety net
- Binary resource model: TenantScoped() on or off — no shared-with-overrides pattern
- RLS context via SET LOCAL per-transaction — middleware runs SET LOCAL app.current_tenant = $1 at transaction start

### Audit trail design
- Full before/after diff stored as JSONB for every changed field — complete history
- Creation recorded as first audit entry with all initial field values — complete lifecycle from create through updates
- Single shared audit_logs table with resource_type and resource_id columns (not per-resource tables)
- Audit log surfaced via API endpoint only (GET /api/v1/resources/:id/audit) — no generated HTML view
- created_by/updated_by columns auto-populated from context user

### Claude's Discretion
- Relationship preloading strategy (Eager implementation details, N+1 prevention approach)
- Audit log retention/cleanup strategy
- RLS policy SQL generation specifics
- WithTrashed/OnlyTrashed query scope API design
- Exact audit_logs table schema (indexes, partitioning)

</decisions>

<specifics>
## Specific Ideas

- PRD Section 8.4 explicitly rejects automatic cascading soft deletes — "implicit cascading soft deletes are a source of data loss bugs"
- Phase 6 already established: role == '' || role == 'value' pattern for role guards, mutability modifier for editable-vs-readonly
- RLS SET LOCAL pattern ensures tenant context is transaction-scoped and automatically cleared on tx end

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 07-advanced-data-features*
*Context gathered: 2026-02-17*
