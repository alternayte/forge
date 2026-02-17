---
phase: 06-hypermedia-ui-generation
plan: 01
subsystem: ui
tags: [templ, datastar, datastar-go, html, form-primitives, sse, tailwind]

# Dependency graph
requires:
  - phase: 05-api-generation
    provides: generator patterns (renderTemplate, writeGoFile, writeRawFile, ensureDir, funcmap)

provides:
  - html_primitives.templ.tmpl with FormField, TextInput, DecimalInput, SelectInput, RelationSelect
  - html_sse.go.tmpl with MergeFragment, Redirect, RedirectTo datastar SSE wrappers
  - GenerateHTML function that writes gen/html/primitives/ and gen/html/sse/
  - htmlInputType funcmap helper for field-type-to-HTML-input-type mapping

affects:
  - 06-02 through 06-09: subsequent plans import primitives/sse from gen/html/
  - generator orchestrator (generator.go): must wire in GenerateHTML call

# Tech tracking
tech-stack:
  added:
    - github.com/a-h/templ v0.3.977
    - github.com/starfederation/datastar-go v1.1.0
  patterns:
    - Semi-static generator pattern (GenerateHTML mirrors GenerateMiddleware — not per-resource, project-module-only data)
    - writeRawFile for .templ files (not Go formatted); writeGoFile for .go files
    - datastar.TemplComponent interface used in sse.go (no templ import needed in generated SSE package)

key-files:
  created:
    - internal/generator/templates/html_primitives.templ.tmpl
    - internal/generator/templates/html_sse.go.tmpl
    - internal/generator/html.go
    - internal/generator/html_test.go
  modified:
    - internal/generator/funcmap.go (added htmlInputType helper)
    - go.mod / go.sum (added templ + datastar-go dependencies)

key-decisions:
  - "Use datastar.TemplComponent interface in sse.go (not templ.Component import) — avoids templ dependency in generated SSE package while remaining compatible with templ components"
  - "SSE type is datastar.ServerSentEventGenerator (not datastar.SSE alias — no such alias exists in v1.1.0)"
  - "primitives.templ uses writeRawFile (not writeGoFile) because .templ files require templ compiler, not gofmt"
  - "RedirectTo implemented as sse.Redirect(fmt.Sprintf(format, args...)) — Redirectf exists on SDK but wrapping it provides a cleaner semantic name"

patterns-established:
  - "Pattern: Semi-static HTML generators follow middleware.go pattern (ProjectModule data, no per-resource loop)"
  - "Pattern: .templ template files generated with writeRawFile; .go files generated with writeGoFile"

requirements-completed: []

# Metrics
duration: 2min
completed: 2026-02-17
---

# Phase 6 Plan 01: Templ Primitives and SSE Helpers Summary

**Templ form primitives library (FormField, TextInput, DecimalInput, SelectInput, RelationSelect) and Datastar SSE helpers (MergeFragment, Redirect, RedirectTo) generated into gen/html/ via GenerateHTML function**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-17T21:03:35Z
- **Completed:** 2026-02-17T21:05:50Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Created html_primitives.templ.tmpl with 5 form components (FormField, TextInput, DecimalInput, SelectInput, RelationSelect) using Tailwind classes and Datastar data-bind attributes
- Created html_sse.go.tmpl with 3 semantic SSE wrappers (MergeFragment, Redirect, RedirectTo) around the datastar-go SDK
- Created GenerateHTML function following the GenerateMiddleware semi-static pattern
- Added htmlInputType funcmap helper mapping IR field types to HTML input types

## Task Commits

Each task was committed atomically:

1. **Task 1: Create templ primitives template and SSE helpers template** - `5e8b58e` (feat)
2. **Task 2: Create GenerateHTML function and tests** - `ab5c588` (feat)

**Plan metadata:** `[pending]` (docs: complete plan)

## Files Created/Modified
- `internal/generator/templates/html_primitives.templ.tmpl` - Templ component definitions for form primitives with Tailwind styling and Datastar data-bind
- `internal/generator/templates/html_sse.go.tmpl` - SSE helper functions wrapping datastar-go SDK (MergeFragment, Redirect, RedirectTo)
- `internal/generator/html.go` - GenerateHTML function; creates gen/html/primitives/ and gen/html/sse/ directories and renders templates
- `internal/generator/html_test.go` - TestGenerateHTML verifying all components and helpers are present in generated output
- `internal/generator/funcmap.go` - Added htmlInputType helper (String/Text->text, Int/BigInt/Decimal->number, Bool->checkbox, Email->email, URL->url, Date->date, DateTime->datetime-local)
- `go.mod / go.sum` - Added github.com/a-h/templ v0.3.977 and github.com/starfederation/datastar-go v1.1.0

## Decisions Made
- Used `datastar.TemplComponent` interface in sse.go (not `templ.Component`) — avoids importing templ into the generated SSE package while remaining compatible since both define `Render(ctx context.Context, w io.Writer) error`
- SSE type is `datastar.ServerSentEventGenerator` — confirmed no `datastar.SSE` alias exists in v1.1.0; plan reference was incorrect
- `primitives.templ` uses `writeRawFile` (not `writeGoFile`) because .templ files require the `templ` compiler, not `gofmt`
- `RedirectTo` wraps `fmt.Sprintf` + `sse.Redirect` (not `sse.Redirectf`) for explicit control and semantic clarity

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected datastar SSE type name**
- **Found during:** Task 1 (SSE helpers template)
- **Issue:** Plan specified `*datastar.SSE` but actual type is `*datastar.ServerSentEventGenerator` — no SSE alias exists
- **Fix:** Used `*datastar.ServerSentEventGenerator` in generated sse.go template
- **Files modified:** internal/generator/templates/html_sse.go.tmpl
- **Verification:** go build passes with correct type
- **Committed in:** 5e8b58e (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 bug - incorrect type name from plan)
**Impact on plan:** Minor correction; plan intent preserved. No scope creep.

## Issues Encountered
None — both tasks executed cleanly.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- GenerateHTML function ready to wire into generator.go orchestrator (not done in this plan — deferred to later plan)
- gen/html/primitives/ primitives.templ ready for scaffolded view templates to import
- gen/html/sse/ sse.go ready for HTML handlers to import
- templ and datastar-go dependencies added to go.mod

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*
