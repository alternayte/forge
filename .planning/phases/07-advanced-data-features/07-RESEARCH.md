# Phase 7: Advanced Data Features - Research

**Researched:** 2026-02-17
**Domain:** Schema DSL extensions, query modification, Atlas HCL generation, PostgreSQL RLS, JSONB audit diffs, context propagation
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Soft delete behavior
- Deleted records disappear from normal UI — only admin role can see and restore via separate view
- No automatic cascade on soft delete — developer explicitly handles cascade per-relationship in custom action overrides (avoids implicit data loss bugs)
- No hard delete — soft delete is final state, records stay in DB permanently (developer writes raw SQL if needed)
- Partial unique indexes (WHERE deleted_at IS NULL) — unique among active records only, allowing re-creation after soft-delete

#### Permission & role model
- Roles defined as schema-level constants (e.g., Role("admin"), Role("editor")) — compile-time known, generated into permission checks
- Permission denial is contextual: API returns 403 Forbidden, HTML redirects to 'not authorized' page or hides the action entirely
- Invisible fields omitted from response entirely (key doesn't appear in API JSON or HTML template) — response shape changes per role
- Permissions are role-only, not owner-aware — ownership-scoped logic is a custom action override if needed

#### Tenant isolation model
- Tenant context propagated via middleware: extracts tenant from auth token/session, sets in context.Context — generated code reads from context
- Always single-tenant scoping — every query scoped to exactly one tenant, no cross-tenant admin access
- Defense in depth: application-level WHERE clause for performance + PostgreSQL RLS policies as safety net
- Binary resource model: TenantScoped() on or off — no shared-with-overrides pattern
- RLS context via SET LOCAL per-transaction — middleware runs SET LOCAL app.current_tenant = $1 at transaction start

#### Audit trail design
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

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SCHEMA-07 | Developer can define field-level Visibility and Mutability based on roles | DSL: Visibility/Mutability modifiers already detected by scaffold_form template but NOT yet in schema/modifier.go or parser/extractor.go — Phase 7 must add them to schema DSL, ModifierType constants, and isModifierMethod() |
| SCHEMA-08 | Developer can define CRUD-level Permissions (List, Read, Create, Update, Delete) per resource | DSL: New Permission struct type needed in schema package; parser needs isPermissionType(); IR needs PermissionsIR field on ResourceOptionsIR |
| DATA-05 | When TenantScoped is true, all generated queries automatically include tenant_id filtering from context | Query template: add sm.Where(psql.Quote("tenant_id").EQ(psql.Arg(tenantFromCtx(ctx)))) mod in List/Get/Update/Delete query builders; tenantFromCtx helper reads from context.Context |
| DATA-06 | When SoftDelete is true, all queries exclude soft-deleted records by default | Query template: add sm.Where(psql.Quote("deleted_at").IsNull()) mod to all query builders when resource.Options.SoftDelete |
| DATA-07 | Developer can use WithTrashed() and OnlyTrashed() to include/only-show soft-deleted records | Query template: WithTrashed removes the deleted_at IS NULL filter; OnlyTrashed replaces it with deleted_at IS NOT NULL |
| DATA-08 | When a field is Unique on a SoftDelete resource, Atlas generates a partial unique index (WHERE deleted_at IS NULL) | Atlas template: current index block generates `unique = true`; when resource.Options.SoftDelete, add `where = "deleted_at IS NULL"` to unique index block |
| DATA-09 | Generated actions include a Restore method for SoftDelete resources | Actions template: add Restore(ctx, id) method to interface and DefaultActions — sets deleted_at = NULL via UPDATE |
| DATA-12 | Relationship preloading (Eager) automatically excludes soft-deleted related records | Relationship IR needs Eager bool; query template generates JOIN/subquery with deleted_at IS NULL filter on related table |
| AUTH-05 | Generated actions check CRUD-level permissions before executing operations | Actions template: inject permission check at top of each CRUD method — if !hasPermission(ctx, role, "list") return 403 |
| AUTH-06 | Generated actions strip invisible fields based on current user's role before returning data | Actions template OR model template: roleFilteredResponse() omits fields where Visibility modifier role doesn't match current user role |
| TENANT-01 | Developer can configure tenant resolution via header, subdomain, or path strategy | New internal/auth/tenant.go with TenantResolver interface and header/subdomain/path implementations; middleware sets ctx key |
| TENANT-02 | When TenantScoped, all generated queries include tenant_id WHERE clause automatically | Same as DATA-05 — query template conditional on resource.Options.TenantScoped |
| TENANT-03 | Atlas generates row-level security (RLS) policies for tenant-scoped resources | Atlas template: add RLS policy block when resource.Options.TenantScoped; SET LOCAL at transaction start |
| TENANT-04 | Test factories scope data to a test tenant | factory.go.tmpl: when resource.Options.TenantScoped, add TenantID field with default test tenant UUID constant |
| AUDIT-01 | When Auditable is true, created_by and updated_by columns are auto-populated from authenticated user | Actions template: read userID from context in Create/Update; pass to DB query; model.go.tmpl adds CreatedBy/UpdatedBy uuid.UUID fields |
| AUDIT-02 | When Auditable is true, changes are recorded in audit_log table with JSONB diffs of old/new values | Actions template: fetch before-state in Update, compute JSONB diff after update, INSERT into audit_logs; single shared template (not per-resource) |
| AUDIT-03 | No-op updates (no actual changes) do not generate audit log entries | Actions template: compute diff before/after; skip INSERT into audit_logs when diff is empty/null |
</phase_requirements>

## Summary

Phase 7 is a schema annotation enforcement phase — the schema DSL already declares SoftDelete, Auditable, and TenantScoped (they're parsed into ResourceOptionsIR), but the generators completely ignore them. The Visibility and Mutability modifiers are recognized by the scaffold form template but do NOT exist in the schema DSL or parser yet. Phase 7 closes all these gaps.

The work has three tiers. First, DSL expansion: add Visibility/Mutability to schema/modifier.go and parser/extractor.go, add a Permission type to schema package and ResourceOptionsIR. Second, generator modifications: every existing generator (queries, actions, models, atlas, factories) must be updated to conditionally generate advanced behavior based on the options already in ResourceOptionsIR. Third, new runtime code: tenant middleware (internal/auth/tenant.go), audit_logs table generation, and context keys for tenant/user propagation.

A critical architectural insight: the data flow in Phase 7 is entirely within the existing generation pipeline. No new generator files are needed — every feature is a conditional addition inside existing templates (queries.go.tmpl, actions.go.tmpl, atlas_schema.hcl.tmpl, model.go.tmpl, factory.go.tmpl). The only new source files are internal/auth/tenant.go (tenant resolution middleware) and the audit_logs table entry in the atlas template (static, like the sessions table already present).

**Primary recommendation:** Treat Phase 7 as a template-and-IR modification phase. Expand the IR with Permission and Eager fields, wire Visibility/Mutability through the schema DSL, then systematically update each generator template to emit conditional logic for SoftDelete, TenantScoped, Auditable, and Permissions. Keep all generated logic purely additive (no behavioral changes to non-annotated resources).

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/jackc/pgx/v5 | v5.7.1 | PostgreSQL driver | Already in project; pgx.Tx.Exec used for SET LOCAL app.current_tenant |
| github.com/google/uuid | v1.6.0 | UUID types | Already in project; used for UserID/TenantID in context keys |
| context (stdlib) | stdlib | Context propagation for tenant/user | Standard Go pattern; context.WithValue for typed keys |
| encoding/json (stdlib) | stdlib | JSONB diff encoding for audit logs | json.Marshal for before/after audit snapshots |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/stephenafamo/bob | (in go.sum) | Query mod building for soft delete / tenant filters | Adds sm.Where mods to existing query builder pattern |
| reflect (stdlib) | stdlib | Computing JSONB diffs for audit logging | reflect.DeepEqual for no-op detection; struct field comparison |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| reflect-based JSONB diff | json.Marshal + compare | Reflect is direct but error-prone; json.Marshal both structs to map[string]any then diff is simpler and handles nested types correctly |
| context.WithValue with typed key | string key in context | Typed key (private struct) prevents collisions; string keys can clash across packages |
| SET LOCAL per-transaction | SET SESSION | SET LOCAL is transaction-scoped, automatically cleared on COMMIT/ROLLBACK — exactly what we want; SET SESSION leaks across connection pool |

## Architecture Patterns

### Recommended Project Structure

No new directories needed. Changes are:

```
internal/
├── schema/
│   ├── modifier.go        # ADD: ModVisibility, ModMutability constants
│   └── permission.go      # NEW: Permission type, PermissionType constants (List/Read/Create/Update/Delete)
├── parser/
│   ├── ir.go              # ADD: PermissionsIR to ResourceOptionsIR; Eager bool to RelationshipIR
│   └── extractor.go       # ADD: isPermissionType(), extractPermissions(); ADD Visibility/Mutability to isModifierMethod(); ADD Eager to isModifierMethod() (relationships)
├── auth/
│   └── tenant.go          # NEW: TenantResolver interface + header/subdomain/path implementations
└── generator/
    ├── funcmap.go          # ADD: hasVisibility, getVisibilityRole, hasPermission helpers
    └── templates/
        ├── queries.go.tmpl         # ADD: soft delete WHERE, tenant WHERE, WithTrashed/OnlyTrashed
        ├── actions.go.tmpl         # ADD: permission checks, field stripping, Restore method, audit logging
        ├── model.go.tmpl           # ADD: CreatedBy/UpdatedBy fields when Auditable; TenantID when TenantScoped
        ├── atlas_schema.hcl.tmpl   # ADD: partial unique indexes, RLS policies, audit_logs table, tenant columns
        └── factory.go.tmpl         # ADD: TenantID default when TenantScoped
```

### Pattern 1: Soft Delete Query Filtering

**What:** When `resource.Options.SoftDelete` is true, inject `WHERE deleted_at IS NULL` into every SELECT query mod. Expose `WithTrashed()` and `OnlyTrashed()` as additional query scope functions in the queries file.

**When to use:** Conditionally in queries.go.tmpl, guarded by `{{- if .Options.SoftDelete}}`

**Example (queries.go.tmpl addition):**
```go
// Source: pattern consistent with existing sm.Where mods in queries.go.tmpl
{{- if .Options.SoftDelete}}
// ActiveMod returns a query mod that excludes soft-deleted records.
// Applied by default to all List/Get queries.
func ActiveMod() sm.QueryMod[*psql.SelectQuery] {
    return sm.Where(psql.Quote("deleted_at").IsNull())
}

// WithTrashedMod returns a query mod that includes soft-deleted records.
// Use to show all records regardless of deletion state.
func WithTrashedMod() sm.QueryMod[*psql.SelectQuery] {
    // Returns a no-op mod — caller simply omits ActiveMod
    return sm.QueryMod[*psql.SelectQuery]{}
}

// OnlyTrashedMod returns a query mod that shows only soft-deleted records.
func OnlyTrashedMod() sm.QueryMod[*psql.SelectQuery] {
    return sm.Where(psql.Quote("deleted_at").IsNotNull())
}
{{- end}}
```

The `List` method in actions.go.tmpl prepends `ActiveMod()` to filterMods when `SoftDelete` is true (by default). Callers using custom action overrides can swap to `WithTrashedMod()` or `OnlyTrashedMod()`.

### Pattern 2: Tenant Scoping via Context

**What:** A private context key type plus helper functions read and write the current tenant ID. Tenant middleware sets the value; generated query code reads it.

**When to use:** Tenant middleware sets the context key. Generated query builders read from context.

**Example (internal/auth/tenant.go):**
```go
// Source: standard Go context key pattern — avoids package collisions
type tenantContextKey struct{}

func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
    id, ok := ctx.Value(tenantContextKey{}).(uuid.UUID)
    return id, ok
}

func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
    return context.WithValue(ctx, tenantContextKey{}, tenantID)
}

// TenantResolver extracts tenant ID from an HTTP request.
type TenantResolver interface {
    Resolve(r *http.Request) (uuid.UUID, error)
}

// HeaderTenantResolver reads tenant ID from X-Tenant-ID header.
type HeaderTenantResolver struct{ Header string }
func (h HeaderTenantResolver) Resolve(r *http.Request) (uuid.UUID, error) {
    return uuid.Parse(r.Header.Get(h.Header))
}
```

**Example (queries.go.tmpl addition):**
```go
{{- if .Options.TenantScoped}}
// TenantMod returns a query mod scoping results to the current tenant.
// Panics if tenant ID is missing from context — tenant middleware is required.
func TenantMod(ctx context.Context) sm.QueryMod[*psql.SelectQuery] {
    tenantID, ok := auth.TenantFromContext(ctx)
    if !ok {
        panic("tenant ID missing from context — ensure tenant middleware is applied")
    }
    return sm.Where(psql.Quote("tenant_id").EQ(psql.Arg(tenantID)))
}
{{- end}}
```

### Pattern 3: RLS Policy in Atlas HCL

**What:** When `TenantScoped` is true, Atlas HCL template generates a PostgreSQL Row Level Security policy on the table. The RLS policy uses `current_setting('app.current_tenant')` to match against the `tenant_id` column.

**When to use:** atlas_schema.hcl.tmpl, inside the table block, guarded by `{{if .Options.TenantScoped}}`

**Example:**
```hcl
# Source: PostgreSQL RLS documentation pattern; SET LOCAL per transaction
{{if .Options.TenantScoped}}
policy "tenant_isolation" {
  on     = table.{{plural (snake .Name)}}
  for    = "ALL"
  using  = "tenant_id = current_setting('app.current_tenant')::uuid"
}
{{end}}
```

The generated transaction wrapper (transaction.go.tmpl) must also be updated to run `SET LOCAL app.current_tenant = $1` at the start of each transaction when a tenant-scoped resource is involved. This is best done in the tenant middleware before the handler executes — the middleware begins a transaction, runs SET LOCAL, then calls next.

### Pattern 4: Audit Log JSONB Diff

**What:** On Update, fetch the before-state, marshal both before/after to `map[string]any`, diff them, and INSERT into audit_logs if the diff is non-empty. On Create, record all initial field values as the "after" with no "before".

**When to use:** actions.go.tmpl, inside the Update and Create methods, guarded by `{{if .Options.Auditable}}`

**Example (actions.go.tmpl addition):**
```go
{{- if .Options.Auditable}}
// recordAudit records a change to the audit log.
// before is nil for creates. If before and after are identical, no entry is recorded.
func (a *Default{{.Name}}Actions) recordAudit(ctx context.Context, op string, resourceID uuid.UUID, before, after interface{}) error {
    // Marshal to map[string]any for field-level diff
    var beforeMap, afterMap map[string]any
    if before != nil {
        b, _ := json.Marshal(before)
        json.Unmarshal(b, &beforeMap)
    }
    a2, _ := json.Marshal(after)
    json.Unmarshal(a2, &afterMap)

    // Compute diff — only fields that changed
    diff := computeJSONDiff(beforeMap, afterMap)
    if len(diff) == 0 {
        return nil // AUDIT-03: no-op updates produce no audit entry
    }

    userID := auth.UserFromContext(ctx) // returns uuid.UUID or uuid.Nil
    diffJSON, _ := json.Marshal(diff)

    _, err := a.DB.Exec(ctx,
        `INSERT INTO audit_logs (resource_type, resource_id, operation, changed_fields, created_by)
         VALUES ($1, $2, $3, $4, $5)`,
        "{{snake .Name}}", resourceID, op, diffJSON, userID,
    )
    return err
}

func computeJSONDiff(before, after map[string]any) map[string]any {
    diff := make(map[string]any)
    for k, afterVal := range after {
        beforeVal := before[k]
        if !reflect.DeepEqual(beforeVal, afterVal) {
            diff[k] = map[string]any{"before": beforeVal, "after": afterVal}
        }
    }
    return diff
}
{{- end}}
```

### Pattern 5: Permission Checks in Actions

**What:** Each CRUD method checks CRUD-level permissions before executing. Roles are defined in the schema as `Permission("list", Role("admin"), Role("editor"))`. Generated code reads the user's role from context and compares against the permission set.

**When to use:** actions.go.tmpl, first line of each CRUD method, guarded by `{{if .Options.Permissions}}`

**Example (actions.go.tmpl addition):**
```go
func (a *Default{{.Name}}Actions) List(ctx context.Context, ...) (...) {
    {{- if hasPermission .Options "List"}}
    if err := checkPermission(ctx, {{permissionRoles .Options "List"}}); err != nil {
        return nil, 0, err
    }
    {{- end}}
    // ... existing logic
}

// checkPermission checks if the user's role is in the allowed roles list.
// Returns errors.Forbidden if denied.
func checkPermission(ctx context.Context, allowedRoles ...string) error {
    userRole := auth.RoleFromContext(ctx)
    for _, r := range allowedRoles {
        if r == userRole {
            return nil
        }
    }
    return errors.Forbidden("insufficient permissions")
}
```

### Pattern 6: Field Visibility Stripping in API Responses

**What:** When a field has `Visibility("admin")` modifier, the API action strips that field from the response struct if the current user is not in the matching role. The field key does not appear in JSON response.

**When to use:** A `roleFilteredResponse` helper in the generated actions file. Returns a map[string]any with invisible fields omitted.

**Example:**
```go
{{- if hasAnyVisibility .Fields}}
func (a *Default{{.Name}}Actions) roleFilter(role string, item models.{{.Name}}) map[string]any {
    result := map[string]any{
        "id": item.ID,
        {{range .Fields}}
        {{- if not (isIDField .)}}
        {{- if hasModifier .Modifiers "Visibility"}}
        // Field {{.Name}} only visible for role "{{getModifierValue .Modifiers "Visibility"}}"
        // Omitted if role doesn't match
        {{- else}}
        "{{snake .Name}}": item.{{.Name}},
        {{- end}}
        {{- end}}
        {{- end}}
    }
    {{range .Fields}}
    {{- if and (not (isIDField .)) (hasModifier .Modifiers "Visibility")}}
    if role == "" || role == "{{getModifierValue .Modifiers "Visibility"}}" {
        result["{{snake .Name}}"] = item.{{.Name}}
    }
    {{- end}}
    {{- end}}
    return result
}
{{- end}}
```

### Pattern 7: Relationship Eager Loading with Soft Delete

**What:** When a RelationshipIR has `Eager: true`, the generated action's Get/List methods perform a secondary query (not a JOIN, to avoid N+1 complexity at the SQL level) to fetch related records. If the related resource has SoftDelete enabled, the secondary query adds `WHERE deleted_at IS NULL`.

**Recommendation:** Use a separate query (SELECT WHERE foreign_key IN (...)) after the primary query. This avoids JOIN complexity while being more efficient than N+1 individual queries. Known as "batch loading."

**When to use:** actions.go.tmpl, after primary query, guarded by `{{range .Relationships}}{{if .Eager}}`

**Example:**
```go
{{- range .Relationships}}
{{- if .Eager}}
// Batch-load {{.Name}} for each fetched {{$.Name}}
// Related table: {{.Table}}; soft-delete aware if target resource has SoftDelete
ids := make([]uuid.UUID, len(items))
for i, item := range items {
    ids[i] = item.ID
}
// SELECT * FROM {{.Table}} WHERE {{snake $.Name}}_id = ANY($1) AND deleted_at IS NULL
{{- end}}
{{- end}}
```

### Anti-Patterns to Avoid
- **CASCADE soft delete in generator:** The decisions explicitly reject this — never emit cascade soft-delete logic in generated code. Developer handles cascades in custom action overrides.
- **Hard delete via generated code:** No DELETE FROM statement should appear in generated actions when SoftDelete is enabled. Only UPDATE ... SET deleted_at = NOW().
- **Cross-tenant queries:** Generated code never allows querying without tenant filter when TenantScoped is true. The `TenantMod()` panics on missing context to catch misconfiguration early.
- **Audit log on every field touch:** Compute the diff first; only insert when diff is non-empty (AUDIT-03).
- **RLS as the only isolation layer:** RLS is a safety net. Application-level WHERE clauses run first for performance (index usage). RLS is defense-in-depth.
- **JSON marshal in hot path without caching:** For field-level visibility filtering, don't marshal on every request — return the typed struct and let the JSON encoder skip fields via struct tags when possible. The roleFilter approach with map[string]any is acceptable given the phase's design constraints.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSONB diff computation | Custom recursive differ | json.Marshal to map[string]any + key comparison | Field-level diff is simple equality; full JSONB differ is unnecessary complexity for audit logs |
| PostgreSQL RLS policy management | Custom policy sync | Atlas HCL policy blocks (rendered by template) | Atlas already owns the schema state; RLS policies belong in the HCL file alongside tables |
| Relationship loading | ORM-style lazy loading | Batch secondary queries (SELECT ... WHERE id IN (...)) | Lazy loading requires proxy objects; batch loading is explicit, fast, and simple to understand |
| Permission middleware | Custom RBAC library | Inline role comparison in generated actions | The permission model is intentionally simple (role string equality) — no library needed |
| Tenant context propagation | Middleware framework | context.WithValue with private key type | One function + one key struct is sufficient; frameworks add unnecessary complexity |

**Key insight:** Phase 7 is mostly template modification, not new libraries. The hard work of code generation infrastructure is done. Adding conditional blocks to existing templates is the primary task.

## Common Pitfalls

### Pitfall 1: Visibility/Mutability Not in Schema DSL Yet

**What goes wrong:** The scaffold_form.templ.tmpl already uses `hasModifier .Modifiers "Visibility"` and `hasModifier .Modifiers "Mutability"` to generate role-guarded HTML. However, these modifiers are NOT in `internal/schema/modifier.go` (no `ModVisibility`, `ModMutability` constants) and NOT in `internal/parser/extractor.go`'s `isModifierMethod()` map. If a developer writes `schema.String("Name").Visibility("admin")` today, the parser silently ignores it.

**Why it happens:** Phase 6 designed the template to accept these modifiers (anticipating Phase 7) but left the DSL implementation for Phase 7.

**How to avoid:** Phase 7 MUST add ModVisibility and ModMutability to modifier.go AND add "Visibility" and "Mutability" to isModifierMethod() in extractor.go. Without this, SCHEMA-07 and AUTH-06 cannot be satisfied.

**Warning signs:** Parser tests pass but Visibility modifier is always empty in generated code.

### Pitfall 2: Permission IR Gap

**What goes wrong:** ResourceOptionsIR has `SoftDelete bool`, `Auditable bool`, `TenantScoped bool` but NO permissions field. The schema package has no Permission type. SCHEMA-08 requires `Permissions(List("admin"), Create("admin"))` syntax in schema definitions.

**Why it happens:** Permissions were planned for Phase 7 from the start (not in Phase 1 DSL).

**How to avoid:** Add a `PermissionsIR` struct to ir.go (map of operation → []string roles), add `Permissions PermissionsIR` to ResourceOptionsIR, create internal/schema/permission.go, and wire the parser to extract permission definitions.

**Warning signs:** AUTH-05 remains unimplemented; generated actions have no permission checks.

### Pitfall 3: Eager Not in RelationshipIR

**What goes wrong:** RelationshipIR has Name, Type, Table, OnDelete, Optional but no `Eager bool`. DATA-12 requires `Eager` annotation on relationships. If developer writes `schema.HasMany("Tags", "tags").Eager()`, the parser ignores it.

**Why it happens:** Eager loading relationship modifier was not needed in Phase 3-6.

**How to avoid:** Add `Eager bool` to RelationshipIR and add "Eager" as a recognized modifier in extractRelationship (NOT isModifierMethod — relationship modifiers have their own logic in extractRelationship).

**Warning signs:** Relationship queries never include eager-loaded data.

### Pitfall 4: Partial Unique Index in Atlas — Wrong Syntax

**What goes wrong:** Atlas HCL uses a different syntax for partial indexes than standard SQL. The `where` attribute in Atlas index block generates a PostgreSQL partial index. Getting the HCL syntax wrong causes Atlas parse errors.

**Why it happens:** Atlas HCL has its own DSL; the WHERE clause is an attribute, not SQL inline.

**How to avoid:** Use `where = "deleted_at IS NULL"` (not `WHERE deleted_at IS NULL`). The current atlas template generates:
```hcl
index "..." {
  columns = [column.field]
  unique  = true
}
```
For DATA-08, add:
```hcl
{{if and (hasModifier .Modifiers "Unique") $.Options.SoftDelete}}
index "..." {
  columns = [column.field]
  unique  = true
  where   = "deleted_at IS NULL"
}
{{end}}
```

**Warning signs:** `forge migrate diff` produces Atlas parse error when SoftDelete + Unique combination used.

### Pitfall 5: SET LOCAL Must Be Inside a Transaction

**What goes wrong:** `SET LOCAL app.current_tenant = $1` only takes effect within the current transaction. If called on a pool connection (not inside BEGIN...COMMIT), it silently has no effect — or it may error with "SET LOCAL can only be used in transaction blocks."

**Why it happens:** SET LOCAL is transaction-scoped by design (the feature we want), but requires an active transaction.

**How to avoid:** The tenant middleware must always wrap the request handler in a database transaction when TenantScoped resources are in use. The generated transaction wrapper (transaction.go.tmpl already exists) should be used. The pattern is: begin transaction → SET LOCAL → run handler → commit.

**Warning signs:** RLS policies don't block cross-tenant access; no PostgreSQL error appears but tenant isolation is absent.

### Pitfall 6: Audit Log on No-Op Update (AUDIT-03)

**What goes wrong:** Computing diff after update, not before. If you only know the Update input, you can't tell if the DB value actually changed (e.g., updating Name to the same value). The before-state must be fetched BEFORE the UPDATE query runs.

**Why it happens:** Intuitive to record after the fact; before-state fetch adds an extra SELECT.

**How to avoid:** Generated Update method fetches current record first (via Get), runs the UPDATE, then computes diff between before and after. If diff is empty, no audit log INSERT.

**Warning signs:** audit_logs fills with identical before/after entries for no-op updates.

### Pitfall 7: audit_logs Table Must Be Static in Atlas HCL

**What goes wrong:** If the audit_logs table is generated inside the `{{range .Resources}}` loop, it gets emitted once per resource — creating duplicate table definitions that cause Atlas errors.

**Why it happens:** The sessions table (already in atlas_schema.hcl.tmpl) is correctly placed OUTSIDE the range loop as a static block. Audit logs must follow the same pattern.

**How to avoid:** Place the audit_logs table definition OUTSIDE the `{{range .Resources}}` loop, as a static block at the end of atlas_schema.hcl.tmpl, with a `{{if hasAuditableResource .Resources}}` guard so it's only emitted when at least one resource is Auditable.

**Warning signs:** Atlas reports "table audit_logs redefined" or generates N duplicate table blocks.

## Code Examples

Verified patterns from existing codebase analysis:

### Soft Delete in Queries Template (addition to existing queries.go.tmpl)
```go
// Source: extends existing queries.go.tmpl pattern (sm.Where already used)
{{- if .Options.SoftDelete}}

// ActiveMod filters out soft-deleted records. Applied by default in List/Get.
func (f {{$resource}}Filters) ActiveMod() sm.QueryMod[*psql.SelectQuery] {
    return sm.Where(psql.Quote("deleted_at").IsNull())
}

// OnlyTrashedMod shows only soft-deleted records.
func (f {{$resource}}Filters) OnlyTrashedMod() sm.QueryMod[*psql.SelectQuery] {
    return sm.Where(psql.Quote("deleted_at").IsNotNull())
}
{{- end}}
```

### Tenant Scoping in Queries Template
```go
// Source: mirrors existing sm.Where pattern in queries.go.tmpl
{{- if .Options.TenantScoped}}

// TenantMod scopes the query to the current tenant from context.
func (f {{$resource}}Filters) TenantMod(ctx context.Context) sm.QueryMod[*psql.SelectQuery] {
    tenantID, ok := forgeauth.TenantFromContext(ctx)
    if !ok {
        panic("{{.Name}}: tenant ID required in context but not found — is TenantMiddleware applied?")
    }
    return sm.Where(psql.Quote("tenant_id").EQ(psql.Arg(tenantID)))
}
{{- end}}
```

### Restore Method in Actions Template
```go
// Source: extends existing actions.go.tmpl — new method on DefaultActions
{{- if .Options.SoftDelete}}
// Restore restores a soft-deleted {{.Name}} by clearing deleted_at.
func (a *Default{{.Name}}Actions) Restore(ctx context.Context, id uuid.UUID) (*models.{{.Name}}, error) {
    _, err := a.DB.Exec(ctx,
        `UPDATE {{plural (snake .Name)}} SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`,
        id,
    )
    if err != nil {
        return nil, errors.InternalError("failed to restore {{.Name}}")
    }
    return a.Get(ctx, id)
}
{{- end}}
```

### Partial Unique Index in Atlas HCL
```hcl
# Source: extends existing index block in atlas_schema.hcl.tmpl
# Current template generates standard unique index; this adds WHERE clause for SoftDelete resources
{{range .Fields}}
{{if hasModifier .Modifiers "Unique"}}
  {{if $.Options.SoftDelete}}
  index "{{plural (snake $resourceName)}}_{{snake .Name}}_unique_active" {
    columns = [column.{{snake .Name}}]
    unique  = true
    where   = "deleted_at IS NULL"
  }
  {{else}}
  index "{{plural (snake $resourceName)}}_{{snake .Name}}_unique" {
    columns = [column.{{snake .Name}}]
    unique  = true
  }
  {{end}}
{{end}}
{{end}}
```

### Auditable Columns in Atlas HCL
```hcl
# Source: follows same pattern as existing SoftDelete and Timestamps conditionals
{{if .Options.Auditable}}
  column "created_by" {
    type = uuid
    null = true
  }
  column "updated_by" {
    type = uuid
    null = true
  }
{{end}}
```

### audit_logs Table (static block, outside range loop)
```hcl
# Source: follows pattern of existing sessions table (static, outside range)
{{if hasAuditableResource .Resources}}
table "audit_logs" {
  schema = schema.public

  column "id" {
    type    = uuid
    default = sql("gen_random_uuid()")
    null    = false
  }
  column "resource_type" {
    type = varchar(255)
    null = false
  }
  column "resource_id" {
    type = uuid
    null = false
  }
  column "operation" {
    type = varchar(50)
    null = false
  }
  column "changed_fields" {
    type = jsonb
    null = true
  }
  column "created_by" {
    type = uuid
    null = true
  }
  column "created_at" {
    type    = timestamptz
    default = sql("now()")
    null    = false
  }

  primary_key {
    columns = [column.id]
  }

  index "audit_logs_resource_idx" {
    columns = [column.resource_type, column.resource_id]
  }
  index "audit_logs_created_at_idx" {
    columns = [column.created_at]
  }
}
{{end}}
```

### Context Key Pattern (internal/auth/tenant.go)
```go
// Source: standard Go private context key pattern
package auth

type tenantContextKey struct{}
type userContextKey struct{}
type roleContextKey struct{}

func WithTenant(ctx context.Context, id uuid.UUID) context.Context {
    return context.WithValue(ctx, tenantContextKey{}, id)
}

func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
    id, ok := ctx.Value(tenantContextKey{}).(uuid.UUID)
    return id, ok
}

func WithUserRole(ctx context.Context, userID uuid.UUID, role string) context.Context {
    ctx = context.WithValue(ctx, userContextKey{}, userID)
    return context.WithValue(ctx, roleContextKey{}, role)
}

func UserFromContext(ctx context.Context) uuid.UUID {
    id, _ := ctx.Value(userContextKey{}).(uuid.UUID)
    return id
}

func RoleFromContext(ctx context.Context) string {
    role, _ := ctx.Value(roleContextKey{}).(string)
    return role
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| SoftDelete option parsed but ignored by generators | SoftDelete flag read and all generators emit conditional behavior | Phase 7 | Enables DATA-06, DATA-07, DATA-08, DATA-09, DATA-12 |
| TenantScoped parsed but ignored | TenantScoped emits tenant_id WHERE clauses + RLS policies | Phase 7 | Enables TENANT-01 through TENANT-04 |
| Auditable parsed but ignored | Auditable emits created_by/updated_by columns + audit_logs writes | Phase 7 | Enables AUDIT-01 through AUDIT-03 |
| Visibility/Mutability recognized in scaffold template only | Visibility/Mutability in schema DSL + parser + actions API response stripping | Phase 7 | Enables SCHEMA-07 and AUTH-06 end-to-end |
| No CRUD-level permissions | Permissions declared in schema, generated into action method guards | Phase 7 | Enables SCHEMA-08 and AUTH-05 |

**Deprecated/outdated:**
- Standard unique index on SoftDelete + Unique fields: replaced by partial index `WHERE deleted_at IS NULL`
- Ignoring ResourceOptionsIR.SoftDelete/Auditable/TenantScoped in generator: all three now drive conditional generation

## Open Questions

1. **TenantScoped + transaction lifecycle**
   - What we know: SET LOCAL must run inside a transaction; the generated transaction helper exists but is for background job use; the tenant middleware needs to manage the transaction boundary
   - What's unclear: Does the tenant middleware always wrap every request in a transaction, or only for mutation endpoints? Read-only LIST queries still need tenant isolation.
   - Recommendation: Middleware always applies application-level WHERE clause (via TenantMod in queries); transaction with SET LOCAL only required for write operations where RLS is the safety net. Application-level WHERE runs on pool connections (no transaction needed). Document this clearly in generated code comments.

2. **Permission IR representation: schema-level syntax**
   - What we know: Roles are `Role("admin")` constants; permissions map operations to allowed roles; resource-level option (not field-level)
   - What's unclear: How does the developer write this in schema.go? Option `Permissions(List("admin"), Create("admin"))` or `Permission("list", Role("admin"), Role("editor"))`?
   - Recommendation: Use `Permission("list", "admin", "editor")` (operation string + variadic role strings) as it is the simplest to parse via go/ast (all string literals). This aligns with how the role guard pattern already works in generated code (string equality checks).

3. **Audit log retention strategy**
   - What we know: audit_logs grows unboundedly; CONTEXT.md marks retention as Claude's discretion
   - What's unclear: Should Forge generate a cleanup SQL migration or scheduled job?
   - Recommendation: Generate the audit_logs table with a `created_at` index (already in the schema above) and add a comment in generated code referencing a standard PostgreSQL partitioning approach. Do NOT generate automatic cleanup — it's application-specific policy. Document that developers can add `pg_partman` for time-based partitioning or a cron-based DELETE if needed.

4. **Audit GET /api/v1/resources/:id/audit endpoint**
   - What we know: Decision is API endpoint only, no HTML view
   - What's unclear: Does this register via the existing api_register.go.tmpl or require a new template?
   - Recommendation: Add the audit endpoint directly to api_register.go.tmpl inside a `{{if .Options.Auditable}}` guard. The endpoint does a simple SELECT FROM audit_logs WHERE resource_type = $1 AND resource_id = $2 ORDER BY created_at DESC. No separate template needed.

5. **Eager loading IR and Relationships**
   - What we know: Eager bool needs to be on RelationshipIR
   - What's unclear: How does the batch loader know if the RELATED resource has SoftDelete enabled? The generator only has the current resource's IR, not the target resource's options.
   - Recommendation: The Eager preloading in generated code should always add `WHERE deleted_at IS NULL` on the batch query. If the target table doesn't have a deleted_at column, this is a query error at runtime (not compile time). Mitigate by generating a code comment warning: "If related resource does not have SoftDelete enabled, remove the deleted_at IS NULL filter." Alternatively, pass all resources to the generator and cross-reference. The latter is cleaner but requires passing the full resource list to per-resource template data — a generator.go change.

## Sources

### Primary (HIGH confidence)
- Direct code analysis: `/Users/nathananderson-tennant/Development/forge-go/internal/schema/` — confirmed current modifier and option types
- Direct code analysis: `/Users/nathananderson-tennant/Development/forge-go/internal/parser/extractor.go` — confirmed isModifierMethod() gap for Visibility/Mutability
- Direct code analysis: `/Users/nathananderson-tennant/Development/forge-go/internal/parser/ir.go` — confirmed ResourceOptionsIR structure and missing Permission/Eager fields
- Direct code analysis: `/Users/nathananderson-tennant/Development/forge-go/internal/generator/templates/atlas_schema.hcl.tmpl` — confirmed SoftDelete column present, partial index absent
- Direct code analysis: `/Users/nathananderson-tennant/Development/forge-go/internal/generator/templates/actions.go.tmpl` — confirmed no permission checks, no audit logging, no soft-delete handling
- Direct code analysis: `/Users/nathananderson-tennant/Development/forge-go/.planning/STATE.md` — confirmed Phase 6 role guard pattern (role == '' || role == 'value')

### Secondary (MEDIUM confidence)
- PostgreSQL documentation: SET LOCAL is transaction-scoped, SET SESSION is session-scoped — confirmed through standard PostgreSQL RLS documentation pattern
- Atlas HCL syntax: partial indexes use `where = "..."` attribute in index block — standard Atlas schema pattern

### Tertiary (LOW confidence)
- Batch relationship loading vs. JOIN: batch secondary queries recommended over JOINs for N+1 prevention without ORM proxy objects — common Go pattern, not verified against specific benchmark

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries are already in the project; no new dependencies needed
- Architecture: HIGH — all patterns are extensions of existing verified patterns (sm.Where, template conditionals, Atlas HCL blocks)
- Pitfalls: HIGH — all pitfalls identified from direct codebase analysis (missing DSL entries, missing IR fields, Atlas HCL partial index syntax)

**Research date:** 2026-02-17
**Valid until:** 2026-03-19 (30 days — stable codebase, no external dependencies changing)
