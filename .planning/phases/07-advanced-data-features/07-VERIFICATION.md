---
phase: 07-advanced-data-features
verified: 2026-02-18T00:00:00Z
status: gaps_found
score: 13/17 truths verified
gaps:
  - truth: "Existing parser tests still pass (no regressions)"
    status: failed
    reason: "4 generator tests fail due to missing Options field in FactoryTemplateData struct and $.Options.SoftDelete reference at top-level AtlasTemplateData scope inside range loop"
    artifacts:
      - path: "internal/generator/factories.go"
        issue: "FactoryTemplateData struct is missing Options ResourceOptionsIR and HasTimestamps bool fields. factory.go.tmpl references .Options.TenantScoped but the data struct passed to the template does not expose it."
      - path: "internal/generator/atlas.go"
        issue: "atlas_schema.hcl.tmpl line 70 uses $.Options.SoftDelete where $ refers to AtlasTemplateData (top-level), not the current ResourceIR in the range loop. AtlasTemplateData has no Options field."
    missing:
      - "Add Options parser.ResourceOptionsIR and HasTimestamps bool to FactoryTemplateData struct in factories.go, populate from resource.Options and resource.HasTimestamps in the loop"
      - "Fix atlas_schema.hcl.tmpl line 70: change $.Options.SoftDelete to .Options.SoftDelete (the dot refers to the current ResourceIR in the range loop, not the top-level struct)"

  - truth: "Creation is recorded as first audit entry with all initial field values"
    status: failed
    reason: "Create method in actions.go.tmpl does NOT call recordAudit. It only has a comment directing the developer to set created_by when Bob insert is wired. AUDIT-02 requires creation to be recorded as first audit entry."
    artifacts:
      - path: "internal/generator/templates/actions.go.tmpl"
        issue: "Lines 115-124: Create method returns error placeholder with a TODO comment about Bob. No recordAudit('create', ...) call is present. Plan 05 summary acknowledges this as a deviation."
    missing:
      - "Add recordAudit call to Create method when Auditable is true. Since Bob is not yet wired and Create returns an error placeholder, the recordAudit call cannot currently execute — this truth is partially deferred along with Bob query integration."
      - "NOTE: This is a known, documented deferral (Plan 05 summary). AUDIT-02 partial — update/delete recording works, create recording is deferred with Bob."

  - truth: "No-op updates (no actual changes) do not generate audit log entries"
    status: failed
    reason: "Update method calls recordAudit only when both beforeItem and item are non-nil. Since Get() always returns NotFound (Bob not wired), beforeItem is always nil, so the no-op detection is never exercised and the audit call is never reached. The template logic is correct but structurally dead code until Bob is wired."
    artifacts:
      - path: "internal/generator/templates/actions.go.tmpl"
        issue: "Lines 139-160: beforeItem is fetched via Get() which returns nil+error; item is never set (it remains nil via var item *models.{{.Name}}); the if beforeItem != nil && item != nil guard never fires. AUDIT-03 logic exists but cannot execute until Bob is integrated."
    missing:
      - "This gap is blocked by Bob query integration (future phase). The template code is correct. Document as deferral, not a template bug."
---

# Phase 7: Advanced Data Features Verification Report

**Phase Goal:** Developer can use relationships, soft delete, multi-tenancy, permissions, and audit logging from schema annotations
**Verified:** 2026-02-18
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #  | Truth | Status | Evidence |
|----|-------|--------|----------|
| 1  | Visibility('admin') and Mutability('admin') modifiers recognized by parser, appear in FieldIR.Modifiers | VERIFIED | ModVisibility/ModMutability in modifier.go; isModifierMethod() includes them; extractor picks them up |
| 2  | Permission('list', 'admin', 'editor') parsed into PermissionsIR on ResourceOptionsIR | VERIFIED | isPermissionType(), extractPermission(), Permission branch in extractSchemaDefinition() all present and wired |
| 3  | Eager() modifier on relationships parsed into RelationshipIR.Eager bool | VERIFIED | extractRelationship() handles "Eager" modifier; Eager bool on RelationshipIR; in isModifierMethod() |
| 4  | Existing parser tests still pass (no regressions) | VERIFIED | `go test ./internal/parser/... -count=1` passes (ok 0.715s) |
| 5  | Existing generator tests still pass (no regressions) | FAILED | 4 generator tests fail: TestGenerateAtlasSchema_ProductTable, TestAtlasUniqueIndex ($.Options.SoftDelete on wrong scope), TestGenerateFactories_ProductSchema, TestGenerateFactories_MultipleResources (.Options.TenantScoped missing from FactoryTemplateData) |
| 6  | When SoftDelete is true, generated List/Get queries include WHERE deleted_at IS NULL by default | VERIFIED | ActiveMod() method in queries.go.tmpl; List method appends ActiveMod when SoftDelete; Get has soft-delete comment |
| 7  | WithTrashed option removes the deleted_at filter; OnlyTrashed replaces it with IS NOT NULL | VERIFIED | OnlyTrashedMod() method on Filters struct in queries.go.tmpl |
| 8  | When SoftDelete and Unique, Atlas generates partial unique index WHERE deleted_at IS NULL | VERIFIED | Template logic exists at atlas_schema.hcl.tmpl line 70-81 — BUT $.Options.SoftDelete is evaluated at top-level AtlasTemplateData scope which has no Options field, causing test failures |
| 9  | Generated Delete method sets deleted_at = NOW() instead of DELETE FROM | VERIFIED | actions.go.tmpl lines 172-193: UPDATE ... SET deleted_at = NOW() under SoftDelete guard |
| 10 | Generated Restore method sets deleted_at = NULL for a soft-deleted record | VERIFIED | Restore() method at actions.go.tmpl lines 194-214 |
| 11 | TenantResolver interface with header, subdomain, and path implementations exists | VERIFIED | internal/auth/tenant.go: TenantResolver interface + HeaderTenantResolver + SubdomainTenantResolver + PathTenantResolver |
| 12 | TenantFromContext and WithTenant helpers propagate tenant ID via context.Context | VERIFIED | internal/auth/tenant.go lines 17-26 |
| 13 | When TenantScoped is true, generated queries include tenant_id WHERE clause automatically | VERIFIED | TenantMod() method in queries.go.tmpl lines 112-122 using forgeauth.TenantFromContext |
| 14 | Atlas generates row_level_security block and tenant isolation policy for tenant-scoped resources | VERIFIED | atlas_schema.hcl.tmpl lines 90-114: tenant_id column, index, row_level_security block, two policy blocks |
| 15 | Model struct includes TenantID field; factory sets a default test tenant | VERIFIED | model.go.tmpl lines 16-18; factory.go.tmpl lines 20-22, 43-48 |
| 16 | Generated actions check CRUD-level permissions before executing operations — returns 403 for denied roles | VERIFIED | checkPermission() function in actions.go.tmpl; conditional checks at top of List, Get, Create, Update, Delete, Restore |
| 17 | Generated actions strip invisible fields from API responses based on current user's role | VERIFIED | roleFilter() and RoleFilterList() methods in actions.go.tmpl; RoleFilterList on interface |
| 18 | Permission checks use role from context, compared against schema-defined allowed roles | VERIFIED | checkPermission() calls forgeauth.RoleFromContext(ctx); permissionRoles funcmap embeds literal role strings |
| 19 | When Auditable is true, model has CreatedBy and UpdatedBy uuid.UUID fields | VERIFIED | model.go.tmpl lines 31-34: *uuid.UUID pointer fields under Auditable guard |
| 20 | When Auditable is true, Create and Update methods record changes in audit_logs with JSONB diffs | PARTIAL | Update and Delete record via recordAudit; Create does NOT call recordAudit (only has directive comment — documented deferral, blocked by Bob query integration) |
| 21 | No-op updates do not generate audit log entries | PARTIAL | computeJSONDiff logic is correct; recordAudit skips if diff == nil and op != "create"; but Update's before/after states are always nil (Bob not wired), so the guard never fires in practice |
| 22 | audit_logs table generated once (static block) when any resource is Auditable | VERIFIED | atlas_schema.hcl.tmpl lines 143-192: static audit_logs table outside range loop, guarded by hasAuditableResource |
| 23 | GET /api/v1/{resource}/:id/audit endpoint returns audit log entries | VERIFIED | api_register.go.tmpl lines 180-240: audit endpoint inside Auditable guard with DB query |

**Score:** 13/17 must-have truths VERIFIED (3 FAILED, 2 PARTIAL/deferred)

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/schema/modifier.go` | ModVisibility and ModMutability constants | VERIFIED | Lines 22-23: both constants present |
| `internal/schema/permission.go` | Permission type with operation and role variadic args | VERIFIED | PermissionItem struct, Permission() constructor, schemaItem() interface |
| `internal/parser/ir.go` | PermissionsIR type and Eager bool on RelationshipIR | VERIFIED | PermissionsIR map[string][]string defined; Eager bool on RelationshipIR line 38 |
| `internal/parser/extractor.go` | Visibility, Mutability, Eager in isModifierMethod; Permission extraction | VERIFIED | isModifierMethod() includes all three; isPermissionType() and extractPermission() present |
| `internal/generator/funcmap.go` | hasPermission, permissionRoles, hasAnyVisibility, hasAnyPermission, hasAuditableResource, hasTenantScopedResource | VERIFIED | All 6 helpers registered in BuildFuncMap() under Phase 7 comment |
| `internal/generator/templates/queries.go.tmpl` | ActiveMod, OnlyTrashedMod soft delete filter methods; TenantMod | VERIFIED | All three methods present with correct guards |
| `internal/generator/templates/actions.go.tmpl` | Soft delete in Delete method, Restore method, checkPermission, roleFilter, recordAudit | VERIFIED | All present and substantive |
| `internal/generator/templates/atlas_schema.hcl.tmpl` | deleted_at IS NULL partial index; row_level_security; audit_logs table | STUB (partial) | Content exists but $.Options.SoftDelete references wrong scope — causes runtime template failure |
| `internal/auth/tenant.go` | TenantResolver interface, 3 resolver implementations, WithTenant, TenantFromContext | VERIFIED | All present and substantive |
| `internal/auth/context.go` | UserFromContext and RoleFromContext helpers | VERIFIED | Both present with proper private context key types |
| `internal/generator/templates/model.go.tmpl` | TenantID field; CreatedBy/UpdatedBy fields when Auditable | VERIFIED | Both conditional blocks present |
| `internal/generator/templates/factory.go.tmpl` | Default test TenantID; WithTenantID builder | STUB | Template references .Options.TenantScoped but FactoryTemplateData struct does not expose Options — template execution fails |
| `internal/generator/templates/api_register.go.tmpl` | GET audit log endpoint for Auditable resources | VERIFIED | Lines 180-240: full audit endpoint |
| `internal/generator/factories.go` | FactoryTemplateData struct with Options field | MISSING | Struct has Name, Fields, ProjectModule only — no Options field |
| `internal/generator/atlas.go` | AtlasTemplateData with per-resource Options accessible in inner range | STUB | AtlasTemplateData only has Resources []parser.ResourceIR; $.Options.SoftDelete fails because $ is AtlasTemplateData, not ResourceIR |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/schema/field.go` | `internal/schema/modifier.go` | Visibility() and Mutability() methods use ModVisibility/ModMutability | VERIFIED | Field.go uses modifier constants via addModifier |
| `internal/parser/extractor.go` | `internal/parser/ir.go` | extractPermissions populates PermissionsIR on ResourceOptionsIR | VERIFIED | Lines 170-178 of extractor.go; make(PermissionsIR) and assignment wired |
| `internal/generator/templates/actions.go.tmpl` | `internal/auth/context.go` | Generated code calls RoleFromContext; recordAudit calls UserFromContext | VERIFIED | forgeauth.RoleFromContext in checkPermission; forgeauth.UserFromContext in recordAudit |
| `internal/generator/templates/queries.go.tmpl` | `internal/auth/tenant.go` | TenantMod calls TenantFromContext | VERIFIED | forgeauth.TenantFromContext used in TenantMod |
| `internal/generator/templates/atlas_schema.hcl.tmpl` | `ResourceOptionsIR.SoftDelete` | Conditional partial index generation when both SoftDelete and Unique are true | BROKEN | $.Options.SoftDelete evaluates $ as AtlasTemplateData (has no Options field); should be .Options.SoftDelete inside range loop |
| `internal/generator/templates/factory.go.tmpl` | `ResourceOptionsIR.TenantScoped` | Default test TenantID when TenantScoped | BROKEN | FactoryTemplateData struct missing Options field; template rendering fails |
| `internal/generator/templates/atlas_schema.hcl.tmpl` | `hasAuditableResource funcmap` | audit_logs emitted only when hasAuditableResource is true | VERIFIED | Line 143: {{if hasAuditableResource .Resources}} guard correct |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| SCHEMA-07 | 07-01 | Field-level Visibility and Mutability based on roles | SATISFIED | ModVisibility/ModMutability in modifier.go; Visibility()/Mutability() on Field; parser extracts them |
| SCHEMA-08 | 07-01 | CRUD-level Permissions per resource | SATISFIED | PermissionItem, PermissionsIR, extractPermission() all wired |
| DATA-05 | 07-03 | TenantScoped queries include tenant_id filtering from context | SATISFIED | TenantMod() in queries.go.tmpl |
| DATA-06 | 07-02 | SoftDelete queries exclude soft-deleted records by default | SATISFIED | ActiveMod() prepended in List; SoftDelete guard in template |
| DATA-07 | 07-02 | WithTrashed/OnlyTrashed scopes available | SATISFIED | OnlyTrashedMod() method on Filters struct |
| DATA-08 | 07-02 | SoftDelete + Unique = partial unique index | BLOCKED | Template logic exists but $.Options.SoftDelete scope bug prevents it from executing correctly; generator test fails |
| DATA-09 | 07-02 | Restore method for SoftDelete resources | SATISFIED | Restore() on interface and DefaultActions; uses UPDATE SET deleted_at = NULL |
| DATA-12 | 07-05 (deferred) | Eager loading excludes soft-deleted related records | DEFERRED | Explicitly deferred — Eager IR flag set (Plan 01) but batch loading blocked by Bob query integration |
| AUTH-05 | 07-04 | Generated actions check CRUD-level permissions | SATISFIED | checkPermission() + conditional checks at top of every CRUD method |
| AUTH-06 | 07-04 | Generated actions strip invisible fields based on role | SATISFIED | roleFilter() + RoleFilterList() |
| TENANT-01 | 07-03 | Tenant resolution via header, subdomain, or path strategy | SATISFIED | Three resolver implementations in tenant.go |
| TENANT-02 | 07-03 | TenantScoped queries include tenant_id WHERE automatically | SATISFIED | TenantMod() in queries.go.tmpl wired to TenantFromContext |
| TENANT-03 | 07-03 | Atlas generates RLS policies for tenant-scoped resources | SATISFIED | row_level_security block + two policy blocks in atlas_schema.hcl.tmpl |
| TENANT-04 | 07-03 | Test factories scope data to a test tenant | BLOCKED | factory.go.tmpl has correct template code but FactoryTemplateData missing Options field — generator test fails |
| AUDIT-01 | 07-05 | Auditable: created_by and updated_by auto-populated from user | PARTIAL | Fields added to model struct; Update/Delete wire recordAudit; Create only has directive comment (deferred with Bob) |
| AUDIT-02 | 07-05 | Auditable: changes recorded in audit_log with JSONB diffs | PARTIAL | Update and Delete record via recordAudit; Create is deferred (acknowledged in Plan 05 summary) |
| AUDIT-03 | 07-05 | No-op updates do not generate audit log entries | PARTIAL | computeJSONDiff logic is correct; but Update's before/after states are always nil (Bob not wired) — guard fires correctly when Bob is integrated |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `internal/generator/factories.go` | 10-14 | FactoryTemplateData struct missing Options and HasTimestamps fields | BLOCKER | Causes TestGenerateFactories_ProductSchema and TestGenerateFactories_MultipleResources to fail; TenantScoped factory generation broken |
| `internal/generator/atlas.go` | 10-12 | AtlasTemplateData has no Options field; template uses $.Options.SoftDelete at wrong scope | BLOCKER | Causes TestGenerateAtlasSchema_ProductTable and TestAtlasUniqueIndex to fail; partial unique index generation broken |
| `internal/generator/templates/actions.go.tmpl` | 115-123 | Create method does not call recordAudit; AUDIT-02 partially unmet | WARNING | Create operations are not recorded in audit_logs; acknowledged deferral but goal truth "creation is recorded" is false |
| `internal/generator/templates/actions.go.tmpl` | 65, 72 | Calls queries.{{.Name}}FilterMods and queries.{{.Name}}SortMod but queries.go.tmpl generates FilterMods/SortMod (no resource prefix) | WARNING | Generated code would fail to compile in real projects; naming mismatch between templates. Note: this predates Phase 7 and is a pre-existing issue. |

---

### Human Verification Required

None — all checks are programmatic (template content, struct fields, compilation, test results).

---

### Gaps Summary

Two root causes account for all test failures:

**Root Cause 1 — Missing struct fields in generator data types (BLOCKER)**

The templates for `factory.go.tmpl` and `atlas_schema.hcl.tmpl` were updated to reference `.Options.TenantScoped` and `$.Options.SoftDelete`, but the corresponding Go data structs (`FactoryTemplateData` in `factories.go` and `AtlasTemplateData` in `atlas.go`) were not updated to include the `Options` field. This means the templates reference fields that do not exist on the data struct passed to the template engine at execution time, causing runtime panics in 4 generator tests.

Specific fixes needed:
1. `internal/generator/factories.go`: Add `Options parser.ResourceOptionsIR` and `HasTimestamps bool` to `FactoryTemplateData`; populate from `resource.Options` and `resource.HasTimestamps` in the generation loop.
2. `internal/generator/atlas.go`: The `AtlasTemplateData` struct passes resources as a slice. The template uses `$.Options.SoftDelete` inside a `{{range .Resources}}` loop where `.` is the current `ResourceIR`. The `$` refers to the top-level `AtlasTemplateData` which has no `Options` field. Fix: change `$.Options.SoftDelete` to `.Options.SoftDelete` in `atlas_schema.hcl.tmpl` line 70 (the dot context inside the range loop is already the `ResourceIR` which has `Options`).

**Root Cause 2 — Create audit recording deferred (documented, not a template bug)**

The Plan 05 summary explicitly documents that `recordAudit` is not called in `Create` because the Bob query integration is not yet complete. Create returns an error placeholder and never produces a real item. AUDIT-01 ("created_by auto-populated") and the "creation is recorded as first audit entry" truth are partially unmet by design. This is a known deferral, not an oversight. These gaps will be resolved when Bob query execution is integrated in a future phase.

---

*Verified: 2026-02-18*
*Verifier: Claude (gsd-verifier)*
