---
phase: 06-hypermedia-ui-generation
plan: 09
subsystem: ui
tags: [cli, scaffold, tailwind, html-server, session, chi, cobra]

# Dependency graph
requires:
  - phase: 06-05
    provides: generator.ScaffoldResource and generator.DiffResource
  - phase: 06-06
    provides: ScaffoldResource and DiffResource implementations
  - phase: 06-08
    provides: GenerateHTML orchestrator + htmlRoutes() in forge routes

provides:
  - internal/cli/generate_resource.go: forge generate resource <name> [--diff] CLI command
  - internal/api/html_server.go: SetupHTML wires session middleware + public/protected route groups
  - internal/watcher/tailwind.go: RunTailwind, RunTailwindWatch, ScaffoldTailwindInput

affects:
  - forge generate resource <name>: wires ScaffoldResource into user-facing CLI
  - forge generate resource <name> --diff: wires DiffResource into user-facing CLI
  - HTML server setup: SetupHTML used by generated app main.go for session+auth wiring
  - forge dev: RunTailwindWatch provides --watch mode for continuous CSS compilation

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "forge generate resource as subcommand of forge generate (generateCmd.AddCommand pattern)"
    - "Case-insensitive resource name matching with available-list error message"
    - "runTemplGenerate falls back from .forge/bin/templ to PATH for templ binary lookup"
    - "SetupHTML uses two route groups: public (no auth) + protected (RequireSession) with LoadAndSave wrapping both"
    - "Tailwind binary at .forge/bin/tailwindcss (zero npm); scaffold-once for input.css + tailwind.config.js"

key-files:
  created:
    - internal/cli/generate_resource.go
    - internal/api/html_server.go
    - internal/watcher/tailwind.go
  modified:
    - internal/cli/root.go

key-decisions:
  - "runTemplGenerate tries .forge/bin/templ first then falls back to PATH — forge tool sync is preferred but not required to use the command"
  - "SetupHTML public group is intentionally empty — app registers OAuth routes via RegisterOAuthRoutes; forge framework cannot import generated auth code"
  - "RunTailwindWatch uses cmd.Start (not cmd.Run) so caller controls process lifetime — aligns with forge dev's start/stop lifecycle"
  - "ScaffoldTailwindInput uses tailwindInputCSS const with v3 @tailwind directives (not v4 @import) — matches toolsync registry v3.4.17 pin"

# Metrics
duration: 3min
completed: 2026-02-17
---

# Phase 6 Plan 09: CLI Wiring, HTML Server Setup, and Tailwind Compilation Summary

**forge generate resource CLI command wiring ScaffoldResource/DiffResource, SetupHTML session middleware with public/protected route groups, and Tailwind CSS standalone compilation via .forge/bin/tailwindcss**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-17T21:24:09Z
- **Completed:** 2026-02-17T21:27:09Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Created `generate_resource.go` with `newGenerateResourceCmd()` (cobra.ExactArgs(1), --diff flag) and `runGenerateResource()` that finds resource by case-insensitive name, calls DiffResource or ScaffoldResource, then runs templ generate on scaffolded views
- Updated `root.go` to wire `newGenerateResourceCmd()` as a subcommand of `newGenerateCmd()` making the full command `forge generate resource <name>`
- Created `html_server.go` with `HTMLServerConfig` struct and `SetupHTML()` that applies LoadAndSave to all routes, creates a public group (auth/OAuth routes), and a protected group with RequireSession + RegisterRoutes
- Created `tailwind.go` with `RunTailwind` (single-shot compilation), `RunTailwindWatch` (--watch mode for forge dev), and `ScaffoldTailwindInput` (scaffold-once input.css + tailwind.config.js)

## Task Commits

Each task was committed atomically:

1. **Task 1: Create forge generate resource CLI command with --diff flag** - `78a52ff` (feat)
2. **Task 2: Create HTML server setup and Tailwind compilation helper** - `b44c414` (feat)

## Files Created/Modified
- `internal/cli/generate_resource.go` - newGenerateResourceCmd (--diff flag, cobra.ExactArgs(1)), runGenerateResource (case-insensitive resource lookup, DiffResource/ScaffoldResource dispatch, templ generate on views), runTemplGenerate helper (forge bin + PATH fallback)
- `internal/cli/root.go` - generateCmd stored as variable, newGenerateResourceCmd() added as subcommand before rootCmd.AddCommand(generateCmd)
- `internal/api/html_server.go` - HTMLServerConfig struct (SessionManager, RegisterRoutes func), SetupHTML wiring LoadAndSave + public group + protected RequireSession group
- `internal/watcher/tailwind.go` - RunTailwind (single-shot), RunTailwindWatch (--watch + cmd.Start), ScaffoldTailwindInput (input.css + tailwind.config.js scaffold-once), tailwindBinPath helper

## Decisions Made
- `runTemplGenerate` falls back from `.forge/bin/templ` to PATH using `exec.LookPath` — the forge-managed binary is preferred but developers can use a system-installed templ without running `forge tool sync`
- `SetupHTML` public group is intentionally left empty of route registrations — the generated application registers OAuth/login/logout routes separately; SetupHTML provides the group structure without importing generated code
- `RunTailwindWatch` uses `cmd.Start` (not `cmd.Run`) and returns the `*exec.Cmd` to the caller — this matches the forge dev server pattern where the caller manages the process lifetime and can kill it on shutdown
- Tailwind v3.4.17 @tailwind directives retained in `ScaffoldTailwindInput` (not v4 @import syntax) — consistent with toolsync registry pin

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 6 complete: all 9 plans executed
- forge generate resource <name> wires the full scaffold pipeline end-to-end
- forge generate resource <name> --diff provides safe preview before writing
- SetupHTML provides the session+auth wiring entry point for generated app main.go
- RunTailwindWatch available for forge dev CSS watching integration

## Self-Check: PASSED

- FOUND: internal/cli/generate_resource.go
- FOUND: internal/cli/root.go (modified)
- FOUND: internal/api/html_server.go
- FOUND: internal/watcher/tailwind.go
- FOUND: .planning/phases/06-hypermedia-ui-generation/06-09-SUMMARY.md
- FOUND commit 78a52ff (Task 1)
- FOUND commit b44c414 (Task 2)

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*
