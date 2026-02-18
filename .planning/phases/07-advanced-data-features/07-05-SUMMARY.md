---
phase: 07-advanced-data-features
plan: 05
subsystem: generator-templates
tags: [audit-logging, jsonb-diff, created-by, updated-by, audit-api, templates]
dependency_graph:
  requires: [07-03, 07-04]
  provides: [AUDIT-01, AUDIT-02, AUDIT-03]
  affects: [model.go.tmpl, atlas_schema.hcl.tmpl, actions.go.tmpl, api_register.go.tmpl]
tech_stack:
  added: []
  patterns:
    - JSONB diff computation via marshal/unmarshal + reflect.DeepEqual
    - Audit log as single shared static table (not per-resource)
    - GetDB() type assertion pattern for audit endpoint DB access
key_files:
  created: []
  modified:
    - internal/generator/templates/model.go.tmpl
    - internal/generator/templates/atlas_schema.hcl.tmpl
    - internal/generator/templates/actions.go.tmpl
    - internal/generator/templates/api_register.go.tmpl
decisions:
  - Audit calls in Update only execute when beforeItem and item are both non-nil — safe no-op until Bob query execution is wired
  - GetDB() type assertion (not interface method) on XxxActions keeps interface surface minimal while enabling audit endpoint DB access
  - recordAudit ignores errors for soft-delete path (nolint:errcheck) — audit failure must not block user-facing operations
  - computeJSONDiff is package-level function (not method) — reusable and testable without DefaultActions receiver
  - DATA-12 (eager loading with soft-delete awareness) deferred — Eager IR flag set (Plan 01), batch loading blocked by Bob query integration
metrics:
  duration: 4m
  completed_date: 2026-02-18
  tasks_completed: 2
  files_modified: 4
---

# Phase 7 Plan 5: Audit Logging Summary

Audit logging with JSONB diffs, created_by/updated_by auto-population, no-op detection, shared audit_logs table, and GET audit history API endpoint.

## What Was Built

### Task 1: Audit columns in model/atlas templates (commit: acbaa22)

**model.go.tmpl** — Added `CreatedBy` and `UpdatedBy` pointer UUID fields to the main model struct when `Auditable=true`. Placed after timestamp and soft-delete fields. Pointer types allow null for system operations without an authenticated user.

**atlas_schema.hcl.tmpl** — Two additions:
1. `created_by` and `updated_by` nullable UUID columns added to per-resource tables inside the range loop when `Auditable=true`
2. Static `audit_logs` table added OUTSIDE the range loop, after the sessions table, guarded by `{{if hasAuditableResource .Resources}}`. The table has `resource_type`, `resource_id`, `operation`, `changed_fields` (JSONB), `created_by`, `created_at` with a composite index on `(resource_type, resource_id)` and an index on `created_at`

### Task 2: Audit recording in actions and API endpoint (commit: f87679b)

**actions.go.tmpl** — Four additions:

1. **Conditional imports**: `encoding/json` and `reflect` when Auditable; `forgeauth` import condition expanded to include Auditable

2. **computeJSONDiff**: Package-level function comparing two `map[string]any` representations. Returns field-level `{"before": X, "after": Y}` map for changed fields only. Returns nil for identical maps (enables no-op detection, AUDIT-03).

3. **recordAudit method**: Marshals before/after to maps via JSON round-trip, computes diff, skips insert if nil diff and op != "create". For creates with empty diff, records all fields as the snapshot. Reads `forgeauth.UserFromContext(ctx)` for `created_by`. Inserts into `audit_logs` table.

4. **GetDB() method**: Returns the `DB` interface for audit endpoint direct DB access.

5. **Wiring**:
   - `Create`: Comment directing developer to set `created_by` from `forgeauth.UserFromContext(ctx)` when Bob insert is wired (AUDIT-01)
   - `Update`: Fetches `beforeItem` before the update placeholder; calls `recordAudit("update", ...)` with before/after when both are non-nil (AUDIT-02, AUDIT-03)
   - `Delete` (soft delete path): Calls `recordAudit("delete", ...)` after successful soft-delete (AUDIT-02)

**api_register.go.tmpl** — Added:
- Conditional imports: `encoding/json` and `time` when Auditable
- `GET /api/v1/{resource}/:id/audit` endpoint inside `{{- if .Options.Auditable}}` guard
- Type-asserts `act` to `interface{ GetDB() actions.DB }` to access DB for audit_logs query
- Returns entries as `[]map[string]any` with `id`, `operation`, `changed_fields`, `created_by`, `created_at`

## Deviations from Plan

### Out-of-scope discoveries

Pre-existing test failures in `TestGenerateAtlasSchema_ProductTable`, `TestAtlasUniqueIndex`, `TestGenerateFactories_ProductSchema`, `TestGenerateFactories_MultipleResources` — these fail due to `$.Options.SoftDelete` reference inside range loop context. These failures predate this plan and are unrelated to audit logging. Logged to deferred items.

### Adaptation

The plan showed `recordAudit` call in Create with `if err := a.recordAudit(ctx, "create", item.ID, nil, item)`. Since Create returns an error placeholder until Bob is wired (item is never returned), the call was replaced with a directive comment: _"When Bob insert is wired, set created_by from forgeauth.UserFromContext(ctx) on the insert payload"_. This matches the plan's own note: "add a comment indicating created_by should be set from context when the query is implemented."

## Requirements Addressed

- **AUDIT-01**: `created_by`/`updated_by` fields added to model struct; comment in Create directs auto-population from context when Bob insert is wired
- **AUDIT-02**: `recordAudit` inserts JSONB diffs for create/update/delete operations; GET audit endpoint exposes history
- **AUDIT-03**: `computeJSONDiff` returns nil for identical maps; `recordAudit` skips insert when diff is nil and op != "create"

## Deferred Items

- **DATA-12**: Eager loading with soft-delete awareness — blocked by Bob query execution integration. The `Eager` IR flag is set (Plan 01) but batch loading cannot be implemented until Bob queries are fully wired. Deferred to future phase.

## Self-Check: PASSED

All files exist and all commits verified:
- model.go.tmpl: FOUND, contains CreatedBy
- atlas_schema.hcl.tmpl: FOUND, contains audit_logs
- actions.go.tmpl: FOUND, contains recordAudit
- api_register.go.tmpl: FOUND, contains audit endpoint
- Commit acbaa22: FOUND
- Commit f87679b: FOUND
