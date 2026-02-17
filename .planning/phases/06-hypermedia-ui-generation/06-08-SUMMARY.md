---
phase: 06-hypermedia-ui-generation
plan: 08
subsystem: ui
tags: [generator, html, routes, datastar, chi, templ]

# Dependency graph
requires:
  - phase: 06-01
    provides: GenerateHTML foundation (primitives.templ + sse.go generation)
  - phase: 06-04
    provides: html_register_all.go.tmpl template for HTML route dispatcher
provides:
  - GenerateHTML now produces gen/html/register_all.go with RegisterAllHTMLRoutes dispatcher
  - forge generate orchestrator includes GenerateHTML as 13th generator (after GenerateAPI)
  - forge routes displays both API routes (/api/v1/...) and HTML routes (/...) per resource
affects:
  - 06-09

# Tech tracking
tech-stack:
  added: []
  patterns:
    - htmlRoutes() separated from apiRoutes() following same pattern as apiRoutes() was separated from runRoutes() in Phase 5
    - Routes command displays multiple sections (API + HTML) under each resource header
    - GenerateHTML generates 3 files: primitives.templ, sse.go, register_all.go

key-files:
  created: []
  modified:
    - internal/generator/html.go
    - internal/generator/html_test.go
    - internal/generator/generator.go
    - internal/cli/routes.go

key-decisions:
  - "htmlRoutes() returns 7 routes (list, new, detail, edit, create, update, delete) at root path (no /api/v1/ prefix)"
  - "forge routes displays API and HTML sections per resource with combined total count"

patterns-established:
  - "Route display pattern: resource header, then API sub-section, then HTML sub-section, each with lipgloss method colors"

requirements-completed: []

# Metrics
duration: 2min
completed: 2026-02-17
---

# Phase 6 Plan 8: Wire HTML Generation and Routes Summary

**html_register_all.go dispatcher wired into orchestrator as 13th generator, and forge routes extended with 7 HTML routes per resource in API+HTML sectioned display**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-17T21:18:06Z
- **Completed:** 2026-02-17T21:20:00Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- GenerateHTML now produces a third output file: gen/html/register_all.go containing the RegisterAllHTMLRoutes chi dispatcher
- forge generate orchestrator calls GenerateHTML as the 13th generator in the pipeline after GenerateAPI
- forge routes command shows both API Routes (5 per resource, /api/v1/ prefix) and HTML Routes (7 per resource, root path) in sectioned display
- TestGenerateHTML updated to use a sample resource and assert all 3 generated files including register_all.go with registry.Get

## Task Commits

Each task was committed atomically:

1. **Task 1: Update GenerateHTML to also generate html_register_all.go, wire into orchestrator** - `7ada7f1` (feat)
2. **Task 2: Add htmlRoutes() to forge routes command** - `9fe99c9` (feat)

**Plan metadata:** (docs commit to follow)

## Files Created/Modified
- `internal/generator/html.go` - Added third generation step rendering html_register_all.go.tmpl and writing gen/html/register_all.go
- `internal/generator/html_test.go` - Updated to use a sample resource; added assertions for register_all.go (RegisterAllHTMLRoutes, package html, registry.Get)
- `internal/generator/generator.go` - Added GenerateHTML as 13th generator after GenerateAPI in Generate() orchestrator pipeline
- `internal/cli/routes.go` - Added htmlRoutes() returning 7 HTML routes; updated runRoutes() to display API and HTML sections per resource with combined totals

## Decisions Made
- htmlRoutes() returns 7 routes per resource at root path (no /api/v1/ prefix): GET /{plural}, GET /{plural}/new, GET /{plural}/{id}, GET /{plural}/{id}/edit, POST /{plural}, PUT /{plural}/{id}, DELETE /{plural}/{id}
- forge routes displays separate "API" and "HTML" sub-sections under each resource header using sectionStyle (bold dim white) distinct from the resource header bold style
- TestGenerateHTML changed from nil resources to []ResourceIR{{Name: "Product"}} so the range block in html_register_all.go.tmpl emits registry.Get content

## Deviations from Plan

None - plan executed exactly as written. The only adjustment was updating the test to use a real resource (instead of nil) so that registry.Get appears in the generated register_all.go output when the range block executes.

## Issues Encountered
- Initial test call used nil resources, so registry.Get was absent from register_all.go output (range block skipped). Updated test to pass a sample ResourceIR so the assertion could be met as specified in the plan.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- HTML generation pipeline complete: primitives.templ + sse.go + register_all.go all generated
- forge routes shows unified view of API + HTML routes
- Ready for Phase 06-09 (final integration / wiring)

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*

## Self-Check: PASSED

- FOUND: internal/generator/html.go
- FOUND: internal/generator/html_test.go
- FOUND: internal/generator/generator.go
- FOUND: internal/cli/routes.go
- FOUND: 06-08-SUMMARY.md
- FOUND: commit 7ada7f1 (Task 1)
- FOUND: commit 9fe99c9 (Task 2)
