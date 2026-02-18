---
phase: 07-advanced-data-features
plan: 02
subsystem: database
tags: [soft-delete, templates, code-generation, queries, actions, atlas, partial-index]

requires:
  - phase: 07-advanced-data-features
    provides: SoftDelete option on ResourceOptionsIR established in 07-01 foundation plan

provides:
  - ActiveMod() and OnlyTrashedMod() methods on {{resource}}Filters for soft delete scoping
  - Partial unique index generation (WHERE deleted_at IS NULL) in Atlas schema when SoftDelete + Unique
  - Soft delete behavior in Delete method (UPDATE SET deleted_at = NOW() instead of hard DELETE)
  - Restore method on {{resource}}Actions interface and DefaultActions implementation
  - List method auto-prepends ActiveMod() when SoftDelete enabled (invisible by default)

affects:
  - 07-03 tenant scoping (List/Get filtering pattern established here)
  - 07-04 permissions generation (Restore requires admin permission enforcement)
  - 07-05 eager loading (query mod pattern reused)

tech-stack:
  added: []
  patterns:
    - "SoftDelete conditional blocks in templates use {{- if .Options.SoftDelete}} guards"
    - "Partial unique index uses $.Options.SoftDelete ($ for outer range scope access)"
    - "Restore method placed outside Delete function, in its own guarded block at end of file"
    - "ActiveMod prepended to filterMods slice (not appended) for visibility as default behavior"

key-files:
  created: []
  modified:
    - internal/generator/templates/queries.go.tmpl
    - internal/generator/templates/actions.go.tmpl
    - internal/generator/templates/atlas_schema.hcl.tmpl

key-decisions:
  - "No hard DELETE generated for SoftDelete resources — soft delete is final state, raw SQL available for developer if truly needed"
  - "Restore method is on the interface (not DefaultActions only) — callers can override behavior while maintaining the contract"
  - "ActiveMod prepended in List (not appended) — soft delete filter established before user filters, semantically clearest ordering"

patterns-established:
  - "Soft delete scoping pattern: ActiveMod() / OnlyTrashedMod() as named methods on Filters struct, not inline SQL strings"
  - "Template conditional soft delete: {{- if .Options.SoftDelete}} wraps entire method body or method definition"
  - "Partial index HCL pattern: $.Options.SoftDelete accesses outer range context, generates _unique_active suffix"

requirements-completed: [DATA-06, DATA-07, DATA-08, DATA-09]

duration: 2min
completed: 2026-02-18
---

# Phase 7 Plan 02: Soft Delete Generation Summary

**Soft delete wired across queries, actions, and Atlas templates: invisible by default, restorable via Restore method, partial unique indexes for active records**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-18T07:36:35Z
- **Completed:** 2026-02-18T07:38:19Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Added ActiveMod() and OnlyTrashedMod() filter methods on {{resource}}Filters, guarded by SoftDelete option
- Updated Atlas HCL template to generate partial unique index (WHERE deleted_at IS NULL) when SoftDelete + Unique; standard unique index otherwise (no regression)
- Modified Delete method to use UPDATE SET deleted_at = NOW() (no hard DELETE for SoftDelete resources)
- Added Restore method to both interface and DefaultActions implementation, using UPDATE SET deleted_at = NULL
- List method auto-prepends ActiveMod() filter when SoftDelete enabled — soft-deleted records invisible by default

## Task Commits

Each task was committed atomically:

1. **Task 1: Add soft delete query mods and partial unique indexes** - `d15d281` (feat)
2. **Task 2: Add soft delete behavior to actions (Delete, Restore, List)** - `b85f5fe` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `internal/generator/templates/queries.go.tmpl` - Added ActiveMod() and OnlyTrashedMod() on {{resource}}Filters inside SoftDelete guard
- `internal/generator/templates/atlas_schema.hcl.tmpl` - Updated unique index generation: partial (WHERE deleted_at IS NULL) for SoftDelete, standard otherwise
- `internal/generator/templates/actions.go.tmpl` - Delete uses soft delete UPDATE, Restore method added, List prepends ActiveMod, Get has soft delete doc note

## Decisions Made
- No hard DELETE generated for SoftDelete resources — soft delete is final state; developers use raw SQL if permanent removal needed
- Restore method on interface (not DefaultActions only) — preserves contract for custom implementations
- ActiveMod prepended to filterMods (not appended) — default exclusion established before user-supplied filters

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Soft delete templates fully wired: queries filter invisible by default, actions use UPDATE not DELETE, Atlas generates partial indexes
- WithTrashed/OnlyTrashed scopes available via OnlyTrashedMod() on Filters struct
- Restore method ready for permission guard wiring in 07-04
- Pattern established for other conditional features (TenantScoped, Auditable) in 07-03 and 07-04

---
*Phase: 07-advanced-data-features*
*Completed: 2026-02-18*

## Self-Check: PASSED

All files verified:
- FOUND: internal/generator/templates/queries.go.tmpl (ActiveMod present)
- FOUND: internal/generator/templates/actions.go.tmpl (Restore present)
- FOUND: internal/generator/templates/atlas_schema.hcl.tmpl (deleted_at IS NULL present)
- FOUND: .planning/phases/07-advanced-data-features/07-02-SUMMARY.md

All commits verified:
- FOUND: d15d281 (Task 1: query mods + atlas partial index)
- FOUND: b85f5fe (Task 2: actions soft delete + restore)
