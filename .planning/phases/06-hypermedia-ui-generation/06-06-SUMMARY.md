---
phase: 06-hypermedia-ui-generation
plan: 06
subsystem: ui
tags: [scaffold, generator, diffmatchpatch, go-diff, skip-if-exists, testing]

# Dependency graph
requires:
  - phase: 06-03
    provides: scaffold_form.templ.tmpl, scaffold_list.templ.tmpl, scaffold_detail.templ.tmpl
  - phase: 06-04
    provides: scaffold_handlers.go.tmpl, scaffold_hooks.go.tmpl

provides:
  - internal/generator/scaffold.go: ScaffoldResource and DiffResource functions
  - ScaffoldResource: writes 5 scaffold-once files into resources/<name>/ with skip-if-exists protection
  - DiffResource: produces unified diff between on-disk and freshly-rendered scaffold files

affects:
  - CLI: forge generate resource <name> command wires to ScaffoldResource
  - CLI: forge generate resource <name> --diff wires to DiffResource
  - 06-07+: subsequent plans can call ScaffoldResource to test the full scaffold pipeline end-to-end

# Tech tracking
tech-stack:
  added:
    - github.com/sergi/go-diff v1.4.0 (diffmatchpatch for unified diff output)
  patterns:
    - "renderScaffoldToMap shared between ScaffoldResource and DiffResource — single render path, two consumers"
    - "scaffoldFile struct with IsTempl bool — routes templ files to writeRawFile and Go files to writeGoFile"
    - "Skip-if-exists pattern: os.Stat check before write prevents overwriting developer customizations"

key-files:
  created:
    - internal/generator/scaffold.go
    - internal/generator/scaffold_test.go
  modified:
    - go.mod (added github.com/sergi/go-diff)
    - go.sum

key-decisions:
  - "DiffResource uses dmp.DiffPrettyText for human-readable output (not patch format) — intended for CLI display, not programmatic patching"
  - "sampleProduct includes Quantity (Int) field in tests — required to exercise data-bind path in form template"
  - "DiffResource reports no-change files with 'No changes.' rather than omitting them — gives complete picture of all scaffold files"

patterns-established:
  - "Pattern: ScaffoldResult.Created/Skipped slices use relative OutputPath (not absolute) — consistent with scaffoldFile struct definition"
  - "Pattern: renderScaffoldToMap returns map[outputPath]bytes — callers use scaffoldFiles() iteration order to maintain predictable ordering"

requirements-completed: []

# Metrics
duration: 2min
completed: 2026-02-17
---

# Phase 6 Plan 06: Scaffold Generator with Skip-If-Exists and Diff Capability Summary

**ScaffoldResource and DiffResource functions using go-diff for unified diff output, with skip-if-exists protection preventing overwrite of developer customizations across 5 scaffold files**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-17T21:17:50Z
- **Completed:** 2026-02-17T21:19:50Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Created `scaffold.go` with `ScaffoldResource` (writes 5 scaffold files into `resources/<name>/` skipping existing files) and `DiffResource` (produces unified diff between on-disk and freshly-rendered scaffolds via `diffmatchpatch`)
- Created `renderScaffoldToMap` shared helper that renders all 5 scaffold templates to a `map[outputPath]bytes` — consumed by both ScaffoldResource and DiffResource
- Added `scaffold_test.go` with 5 test cases covering fresh scaffold, skip-existing protection, multiple resources (no cross-contamination), diff with changes, and diff with no existing files
- Added `github.com/sergi/go-diff v1.4.0` dependency

## Task Commits

Each task was committed atomically:

1. **Task 1: Create ScaffoldResource function with skip-if-exists logic** - `eac17c0` (feat)
2. **Task 2: Create comprehensive scaffold tests** - `85333a0` (feat)

**Plan metadata:** `[pending]` (docs: complete plan)

## Files Created/Modified
- `internal/generator/scaffold.go` - ScaffoldResource (skip-if-exists file writer), DiffResource (unified diff via diffmatchpatch), renderScaffoldToMap (shared render helper), scaffoldFile/ScaffoldResult types
- `internal/generator/scaffold_test.go` - 5 tests: TestScaffoldResource, TestScaffoldResource_SkipsExisting, TestScaffoldResource_MultipleResources, TestDiffResource, TestDiffResource_NoExistingFiles
- `go.mod` - Added github.com/sergi/go-diff v1.4.0
- `go.sum` - Updated checksums

## Decisions Made
- `DiffResource` uses `dmp.DiffPrettyText` for human-readable colored output (not patch format) — intended for CLI display in the terminal, not machine-readable patching
- `sampleProduct()` in tests includes a `Quantity` (Int) field in addition to Name and Price — the Int type generates an inline `data-bind` input whereas String/Decimal delegate to primitives; this exercises the data-bind code path
- `DiffResource` reports all 5 files including unchanged ones (with "No changes.") — gives a complete picture of all scaffold files rather than only showing diffs

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed test data to include Int field for data-bind coverage**
- **Found during:** Task 2 (create comprehensive scaffold tests)
- **Issue:** `sampleProduct()` only had String/Decimal fields; these render via primitives.TextInput/DecimalInput which don't have inline `data-bind`. Test assertion `form.templ contains "data-bind"` failed.
- **Fix:** Added `Quantity (Int)` field to sampleProduct() so the Int branch generates an inline `<input data-bind="quantity">` element
- **Files modified:** internal/generator/scaffold_test.go
- **Verification:** All 5 scaffold tests pass
- **Committed in:** 85333a0 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug in test data)
**Impact on plan:** Auto-fix corrected test data to match actual template behavior. No scope creep.

## Issues Encountered
Test assertion for `data-bind` in form.templ failed because the sample resource lacked an Int/Bool field. Fixed by adding a Quantity (Int) field to the sample product.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- `ScaffoldResource` ready to be wired into `forge generate resource <name>` CLI command
- `DiffResource` ready to be wired into `forge generate resource <name> --diff` CLI flag
- Skip-if-exists protection fully tested — safe to call on existing projects
- All generator tests pass with no regressions (including TestScaffoldTemplates from plan 06-03)

## Self-Check: PASSED

- FOUND: internal/generator/scaffold.go
- FOUND: internal/generator/scaffold_test.go
- FOUND: .planning/phases/06-hypermedia-ui-generation/06-06-SUMMARY.md
- FOUND commit eac17c0 (Task 1)
- FOUND commit 85333a0 (Task 2)

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*
