---
phase: 06-hypermedia-ui-generation
plan: 04
subsystem: ui
tags: [chi, datastar, datastar-go, templ, html, scaffold, handlers, hooks, sse]

# Dependency graph
requires:
  - phase: 06-01
    provides: html_sse.go.tmpl SSE helpers (MergeFragment, Redirect, RedirectTo) and datastar-go dependency
  - phase: 04-02
    provides: actions.go.tmpl with {Name}Actions interface and Registry pattern
  - phase: 05-03
    provides: api_register_all.go.tmpl pattern for registry-based route dispatcher

provides:
  - scaffold_handlers.go.tmpl: HTML handler scaffold with 7 routes calling action layer via Datastar SSE
  - scaffold_hooks.go.tmpl: Lifecycle hook stubs (Before/AfterCreate, Before/AfterUpdate, Before/AfterDelete)
  - html_register_all.go.tmpl: HTML route dispatcher template following api_register_all.go.tmpl pattern

affects:
  - 06-05 through 06-09: subsequent view templates (views.{Name}List, views.{Name}Form, views.{Name}Detail, views.{Name}Error)
  - generator.go orchestrator: must include scaffold_handlers and scaffold_hooks in GenerateScaffold call
  - html.go: must include html_register_all.go.tmpl in GenerateHTML call for gen/html/register_all.go

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Scaffold-once handlers: scaffold_handlers.go.tmpl goes to resources/<name>/handlers.go — developer-owned after generation"
    - "Scaffold-once hooks: scaffold_hooks.go.tmpl goes to resources/<name>/hooks.go — developer-owned lifecycle extension points"
    - "Generated-always dispatcher: html_register_all.go.tmpl regenerates on every forge generate — follows api_register_all.go.tmpl pattern"
    - "SSE mutation: Create/Update/Delete handlers use datastar.ReadSignals + datastar.NewSSE; GET handlers use templ.Render directly"

key-files:
  created:
    - internal/generator/templates/scaffold_handlers.go.tmpl
    - internal/generator/templates/scaffold_hooks.go.tmpl
    - internal/generator/templates/html_register_all.go.tmpl
  modified: []

key-decisions:
  - "scaffold_handlers.go.tmpl template data uses .Resource.Name (not .Name) to match planned struct {Resource parser.ResourceIR, ProjectModule string}"
  - "HandleDelete uses ssehelpers.MergeFragment with views.{Name}Error on failure — plan did not specify error view but handler needs error display"
  - "BeforeUpdate/BeforeDelete hooks accept interface{} id (not uuid.UUID) to avoid importing uuid in the hooks file — simpler generated code"
  - "html_register_all.go.tmpl uses {{snake .Name}} as Go import alias to avoid package name collisions between snake_case resource names"

patterns-established:
  - "Pattern: scaffold templates (scaffold_handlers, scaffold_hooks) use .Resource.Name; generation templates (html_register_all) use .Resources range loop"
  - "Pattern: Datastar SSE flow: datastar.ReadSignals reads signals -> acts.Method executes -> SSE redirects or re-renders form with field errors"

requirements-completed: []

# Metrics
duration: 2min
completed: 2026-02-17
---

# Phase 6 Plan 04: HTML Handler Scaffold Templates and HTML Route Registration Template Summary

**Chi-based HTML handler scaffold (7 CRUD routes calling action layer via Datastar SSE) and HTML route dispatcher template following the api_register_all.go.tmpl registry pattern**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-17T21:09:39Z
- **Completed:** 2026-02-17T21:11:39Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Created scaffold_handlers.go.tmpl generating resources/<name>/handlers.go with 7 HTTP handlers (List, New, Detail, Edit, Create, Update, Delete) routing under /{{kebab (plural .Resource.Name)}} via chi.Router
- Create/Update handlers read Datastar signals via datastar.ReadSignals and respond via datastar.NewSSE (SSE redirect on success, form re-render with field errors on failure); GET handlers render templ components directly
- Created scaffold_hooks.go.tmpl generating resources/<name>/hooks.go with 6 lifecycle hook stubs (BeforeCreate, AfterCreate, BeforeUpdate, AfterUpdate, BeforeDelete, AfterDelete) with no-op default implementations and developer-facing comments
- Created html_register_all.go.tmpl generating gen/html/register_all.go — RegisterAllHTMLRoutes dispatcher that iterates resources, type-asserts registry.Get to {Name}Actions, and calls per-resource Register{Name}HTMLRoutes

## Task Commits

Each task was committed atomically:

1. **Task 1: Create scaffold handlers and hooks templates** - `d3c80bf` (feat)
2. **Task 2: Create HTML route registration template** - `3f92949` (feat)

**Plan metadata:** `[pending]` (docs: complete plan)

## Files Created/Modified
- `internal/generator/templates/scaffold_handlers.go.tmpl` - HTML handler scaffold: Register{Name}HTMLRoutes with 7 chi routes; HandleList/New/Detail/Edit via templ.Render; HandleCreate/Update/Delete via Datastar SSE; toFieldErrors and parseUUID helpers
- `internal/generator/templates/scaffold_hooks.go.tmpl` - Lifecycle hook stubs: BeforeCreate, AfterCreate, BeforeUpdate, AfterUpdate, BeforeDelete, AfterDelete — all no-ops returning nil with developer customization comments
- `internal/generator/templates/html_register_all.go.tmpl` - HTML route dispatcher: RegisterAllHTMLRoutes iterates resources, type-asserts from registry.Get to {Name}Actions, calls {snake.Name}.Register{Name}HTMLRoutes

## Decisions Made
- Template data for scaffold_handlers.go.tmpl uses `.Resource.Name` (struct `{Resource parser.ResourceIR, ProjectModule string}`) consistent with plan spec — distinguishes from the generation template which uses `.Resources` (plural)
- `HandleDelete` uses `ssehelpers.MergeFragment(sse, views.{Name}Error(...))` on failure — plan omitted error display path for delete but handler requires it for correct UX
- Lifecycle hook `id` parameters typed as `interface{}` to avoid uuid import in the hooks file; callers in actions layer pass uuid.UUID values
- `html_register_all.go.tmpl` uses `{{snake .Name}}` as the Go import alias for each resource package to prevent naming conflicts between resources with similar names

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None — both tasks executed cleanly with `go build ./internal/generator/` passing immediately.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- scaffold_handlers.go.tmpl and scaffold_hooks.go.tmpl ready for generator.go to wire into GenerateScaffold call (resources/<name>/handlers.go and hooks.go)
- html_register_all.go.tmpl ready for html.go to render into gen/html/register_all.go
- View functions referenced (views.{Name}List, views.{Name}Form, views.{Name}Detail, views.{Name}Error) must be created in 06-05+ plans
- actions.{Name}Actions interface and registry patterns already established by 04-02 and wired by 05-03

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*
