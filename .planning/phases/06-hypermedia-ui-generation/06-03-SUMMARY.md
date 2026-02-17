---
phase: 06-hypermedia-ui-generation
plan: 03
subsystem: ui
tags: [templ, datastar, scaffold, form, list, detail, role-based-visibility, tailwind]

# Dependency graph
requires:
  - phase: 06-01
    provides: gen/html/primitives (FormField, TextInput, DecimalInput, SelectInput) and funcmap helpers (hasModifier, getModifierValue, isFilterable, isSortable, filterableFields)

provides:
  - scaffold_form.templ.tmpl: Datastar-native form with role-based visibility/mutability guards
  - scaffold_list.templ.tmpl: Paginated table with sortable headers, filter controls, SSE interaction
  - scaffold_detail.templ.tmpl: Read-only dl/dt/dd field display with Edit/Back action links

affects:
  - 06-04: scaffold generator will render these templates into resources/<name>/views/
  - Phase 7: role guards use simple equality check now; hierarchical roles deferred

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Scaffold-once template pattern (developer owns after generation — different from always-regenerated gen/ files)
    - Role-based if-guards in Templ via simple equality: if role == "" || role == "rolename"
    - Mutability modifier generates editable/read-only conditional with fmt.Sprint fallback for non-matching role
    - filterableFields funcmap helper used to conditionally render filter section only when resource has Filterable fields
    - templ.SafeURL + fmt.Sprintf for type-safe dynamic href construction in list action links

key-files:
  created:
    - internal/generator/templates/scaffold_form.templ.tmpl
    - internal/generator/templates/scaffold_detail.templ.tmpl
    - internal/generator/templates/scaffold_list.templ.tmpl
    - internal/generator/scaffold_templates_test.go
  modified: []

key-decisions:
  - "Role guard uses role == '' || role == 'value' pattern — empty role means no auth context (admin/dev views), explicit role enables field-level access control"
  - "Mutability modifier generates editable-vs-readonly conditional rather than hiding the field — ensures data remains visible even for restricted roles"
  - "Filter section only rendered when resource has Filterable fields — uses filterableFields funcmap helper to avoid empty filter UI"
  - "templ.SafeURL used for dynamic action URLs in list rows — required for Templ's XSS-safe href attribute rendering"
  - "Inline inputs (Bool/Int/Date/DateTime) have explicit data-bind; string/decimal fields delegate to primitives.TextInput/DecimalInput which carry data-bind internally"

patterns-established:
  - "Pattern: Scaffold templ templates use $.Resource.Name to access parent context within range .Resource.Fields loops"
  - "Pattern: Generated Templ conditionals (if/for) are output as literal text — Go text/template only processes {{ }} delimiters"

requirements-completed: []

# Metrics
duration: 5min
completed: 2026-02-17
---

# Phase 6 Plan 03: Scaffold Form, List, and Detail Templ Templates Summary

**Three Go text/template files that generate Datastar-native Templ scaffold views (form, list, detail) with role-based field visibility/mutability guards and full field type coverage**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-17T21:09:36Z
- **Completed:** 2026-02-17T21:14:58Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Created `scaffold_form.templ.tmpl` generating Datastar-native forms with `data-on:submit__prevent`, `data-signals` (all non-ID fields), and field-level inputs using the primitives library; Visibility modifier generates `if role == "" || role == "value"` guards; Mutability modifier generates editable/read-only conditionals
- Created `scaffold_detail.templ.tmpl` generating read-only dl/dt/dd field displays with Edit and Back-to-list action buttons
- Created `scaffold_list.templ.tmpl` generating paginated tables with Datastar sort headers (for Sortable fields), filter input controls (for Filterable fields, conditionally rendered), and offset-based Previous/Next pagination
- Added `scaffold_templates_test.go` with `TestScaffoldTemplates` verifying all three templates render with correct content for a sample resource with varied field types and modifiers

## Task Commits

Each task was committed atomically:

1. **Task 1: Create scaffold form and detail templ templates** - `b6a9ca5` (feat)
2. **Task 2: Create scaffold list templ template** - `541c514` (feat)

**Plan metadata:** `[pending]` (docs: complete plan)

## Files Created/Modified
- `internal/generator/templates/scaffold_form.templ.tmpl` - Go text/template producing form.templ with Datastar SSE submission, data-signals initialization, type-specific inputs, FormField/primitives usage, and role-based Visibility/Mutability guards
- `internal/generator/templates/scaffold_detail.templ.tmpl` - Go text/template producing detail.templ with dl/dt/dd read-only layout and action buttons
- `internal/generator/templates/scaffold_list.templ.tmpl` - Go text/template producing list.templ with sortable table headers, filterable field inputs, for-range tbody, and offset pagination
- `internal/generator/scaffold_templates_test.go` - TestScaffoldTemplates covering form (17 checks), detail (9 checks), and list (15 checks) template rendering

## Decisions Made
- Role guard uses `role == "" || role == "value"` — empty role means no auth context (admin/dev views), explicit role restricts field to matching users
- Mutability modifier generates editable-vs-readonly conditional rather than hiding — data remains visible to all but only editable by matching role
- Filter section only rendered when resource has Filterable fields, using the `filterableFields` funcmap helper
- `templ.SafeURL` used for dynamic action URLs in list rows — required for Templ's XSS-safe href rendering
- Inline inputs (Bool, Int, BigInt, Date, DateTime) have explicit `data-bind` attributes; String/Decimal/Enum delegate to primitives which carry binding internally

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered
None — all templates rendered correctly on first attempt. The Go text/template engine properly passes through Templ's `{ expr }` syntax since it only processes `{{ }}` delimiters.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Three scaffold templates ready for the scaffold generator (plan 06-04) to render into `resources/<name>/views/` directories
- Role guards are simple equality checks; Phase 7 permissions system can evolve these when hierarchical roles are implemented
- All templates verified via `go build ./internal/generator/` and `TestScaffoldTemplates`

## Self-Check: PASSED

- FOUND: internal/generator/templates/scaffold_form.templ.tmpl
- FOUND: internal/generator/templates/scaffold_detail.templ.tmpl
- FOUND: internal/generator/templates/scaffold_list.templ.tmpl
- FOUND: internal/generator/scaffold_templates_test.go
- FOUND: .planning/phases/06-hypermedia-ui-generation/06-03-SUMMARY.md
- FOUND commit: b6a9ca5 (Task 1)
- FOUND commit: 541c514 (Task 2)

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*
