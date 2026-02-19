---
phase: quick-1
plan: 01
subsystem: cli
tags: [scaffold, forge-init, forge-dev, ux, module-path]

requires: []
provides:
  - InferModule returns bare project name (no github.com prefix)
  - main.go scaffold uses run() error pattern with next-steps guidance
  - forge dev --help clearly describes file-watching regeneration, not app server
affects: [scaffold, cli]

tech-stack:
  added: []
  patterns:
    - "InferModule is a pure pass-through — user sets their own module path post-init"
    - "Scaffolded main.go uses run() error pattern as foundation for server wiring"

key-files:
  created: []
  modified:
    - internal/scaffold/scaffold.go
    - internal/scaffold/templates/main.go.tmpl
    - internal/cli/dev.go

key-decisions:
  - "InferModule returns projectName as-is — user owns module path, not forge developer's git config"
  - "main.go scaffold uses run() error pattern with commented-out server wiring example for guidance after forge generate"
  - "forge dev Short/Long updated to remove 'development server' confusion and explicitly state it does NOT run the app"

patterns-established: []

requirements-completed: [UX-01, UX-02, UX-03]

duration: 4min
completed: 2026-02-19
---

# Quick Task 1: Fix forge init UX — Module Path, main.go, and Dev Help Text

**InferModule stripped of git-config detection, main.go scaffold replaced with run() error pattern and next-steps guidance, forge dev help text updated to eliminate "development server" confusion.**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-02-19T00:00:00Z
- **Completed:** 2026-02-19T00:04:00Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- `InferModule` now returns the bare project name — `forge init myapp` produces `module myapp` in go.mod, not `module github.com/git-user/myapp`
- `main.go.tmpl` replaced with a compilable `run() error` pattern scaffold that prints next-step guidance and includes a commented-out server wiring example using `{{.Module}}` import paths
- `forge dev --help` now clearly states it is a file watcher that re-runs `forge generate`, explicitly notes it does NOT run the application, and directs users to `go run .`

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix module path inference and main.go scaffold template** - `8ba2ae5` (fix)
2. **Task 2: Clarify forge dev help text as code regeneration watcher** - `a6380d0` (fix)

**Plan metadata:** (see final docs commit)

## Files Created/Modified

- `internal/scaffold/scaffold.go` - InferModule simplified to `return projectName`; `os/exec` import removed
- `internal/scaffold/templates/main.go.tmpl` - Replaced print+exit stub with run() error pattern and next-steps output
- `internal/cli/dev.go` - Updated Short and Long descriptions to clarify watcher-only behavior

## Decisions Made

- InferModule reads git config to infer the user's GitHub username — this is wrong because the user's git name has no reliable relationship to their Go module path. Bare project name is always a valid Go module path and is the correct minimal default.
- main.go template needed the `run() error` pattern so users have the correct structural foundation to build on; the commented-out import block uses `{{.Module}}` so the paths are correct once the user has run `forge generate`.
- "development server" in forge dev misleads users into thinking it runs their application (like `npm run dev`). The corrected help text removes all ambiguity.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed unused os/exec import after InferModule simplification**
- **Found during:** Task 1 (InferModule rewrite)
- **Issue:** Removing the git-config exec.Command call left `os/exec` as an unused import, which is a compile error in Go
- **Fix:** Removed `os/exec` from the import block
- **Files modified:** internal/scaffold/scaffold.go
- **Verification:** `go build ./internal/scaffold/...` passes
- **Committed in:** 8ba2ae5 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - bug/compile error)
**Impact on plan:** Necessary for compilation. No scope creep.

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `forge init` now produces a correct minimal go.mod and a useful main.go scaffold
- Users running `forge dev` will no longer be confused about whether it runs their app
- All existing tests pass (`go test ./...` clean)

---
*Phase: quick-1*
*Completed: 2026-02-19*

## Self-Check: PASSED

- FOUND: internal/scaffold/scaffold.go
- FOUND: internal/scaffold/templates/main.go.tmpl
- FOUND: internal/cli/dev.go
- FOUND: .planning/quick/1-fix-forge-init-ux-module-path-main-go-sc/1-SUMMARY.md
- FOUND commit: 8ba2ae5 (Task 1)
- FOUND commit: a6380d0 (Task 2)
