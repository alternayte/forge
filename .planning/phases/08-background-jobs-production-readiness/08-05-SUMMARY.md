---
phase: 08-background-jobs-production-readiness
plan: 05
subsystem: infra
tags: [go, build, docker, cli, ldflags, embed, tailwind, templ, production]

# Dependency graph
requires:
  - phase: 08-02
    provides: OTel/admin server with /healthz endpoint on port 9090 (used in Dockerfile HEALTHCHECK)
provides:
  - forge build command running generate -> templ -> tailwind -> go build pipeline
  - Binary with -trimpath -ldflags="-s -w" + Version/Commit/Date injected via -X flags
  - forge deploy command generating multi-stage Dockerfile with alpine runtime
  - embed.go template for //go:embed migrations and static assets
  - embed.go written to project root as pre-build step
affects: [user-projects, deployment, production-packaging]

# Tech tracking
tech-stack:
  added: [text/template (deploy Dockerfile rendering)]
  patterns: [pipeline orchestration via sub-process steps, tool binary resolution (.forge/bin -> PATH fallback), ldflags version injection from git + time.Now()]

key-files:
  created:
    - internal/cli/build.go
    - internal/cli/deploy.go
    - internal/generator/templates/embed.go.tmpl
  modified:
    - internal/cli/root.go

key-decisions:
  - "forge build shells out to sub-processes (forge generate, templ, tailwindcss, go build) rather than calling internal functions directly — pipeline orchestrator pattern keeps separation clean"
  - "embed.go written by forge build as pre-step if missing, not by forge generate — embed.go belongs in project root (not gen/), so build owns it"
  - "Tailwind build step skipped gracefully if static/css/input.css doesn't exist — build works even without CSS assets"
  - "Binary size warning printed if > 30MB but not a hard error — provides visibility without blocking builds"
  - "deploy Dockerfile HEALTHCHECK uses wget on localhost:9090/healthz — matches admin server from 08-02"

patterns-established:
  - "resolveToolBinary: check .forge/bin first via toolsync.IsToolInstalled, fall back to exec.LookPath — consistent with generate_resource.go pattern"
  - "runStep: print step name then exec.Command with Dir=projectRoot, Stdout/Stderr piped through"
  - "buildLdflags: get git tag via git describe --tags --always, commit via git rev-parse --short HEAD, date from time.Now().UTC().Format(time.RFC3339)"

requirements-completed: [CLI-03, CLI-07, DEPLOY-02, DEPLOY-03]

# Metrics
duration: 2min
completed: 2026-02-19
---

# Phase 8 Plan 05: forge build and forge deploy CLI Commands Summary

**forge build runs generate -> templ -> tailwind -> go build pipeline with -trimpath -ldflags="-s -w" and git-injected Version/Commit/Date; forge deploy writes a multi-stage alpine Dockerfile with HEALTHCHECK on admin port 9090**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-02-19T17:30:08Z
- **Completed:** 2026-02-19T17:31:58Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- forge build command orchestrates full production build pipeline (generate -> templ generate ./... -> tailwind --minify -> go build)
- Version/Commit/Date injected via ldflags -X flags from git and time.Now() (DEPLOY-02, binary size < 30MB by -trimpath -s -w)
- forge deploy generates multi-stage Dockerfile with alpine:3.21 runtime, EXPOSE 8080 9090, HEALTHCHECK on /healthz (CLI-07)
- Both commands registered in root.go; embed.go.tmpl template created for //go:embed directives

## Task Commits

Each task was committed atomically:

1. **Task 1: Create forge build command with full pipeline and ldflags** - `8da0e48` (feat)
2. **Task 2: Create forge deploy command and register build/deploy in root** - `106153d` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `internal/cli/build.go` - forge build command: resolves tool binaries, runs pipeline steps, injects ldflags, writes embed.go if missing
- `internal/cli/deploy.go` - forge deploy command: renders multi-stage Dockerfile with Go text/template, writes to project root
- `internal/generator/templates/embed.go.tmpl` - Template for embed.go with //go:embed for migrations and static directories
- `internal/cli/root.go` - Registered newBuildCmd() and newDeployCmd()

## Decisions Made

- forge build shells out to sub-processes rather than calling internal functions — pipeline orchestrator pattern keeps separation clean and allows future pipeline extension
- embed.go written by forge build as pre-step (not forge generate) — embed.go belongs in project root, not gen/, so build command is the natural owner
- Tailwind step skipped gracefully if static/css/input.css missing — build works for projects without CSS assets
- Binary size warning printed if > 30MB but not fatal — provides observability without blocking CI builds

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - both tasks compiled clean on first attempt.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 8 is now complete: all 5 plans executed
- forge build and forge deploy provide production packaging pipeline
- forge version displays Version/Commit/Date injected at build time
- Production binary: single statically-linked binary with embedded migrations and static assets

## Self-Check: PASSED

- FOUND: internal/cli/build.go
- FOUND: internal/cli/deploy.go
- FOUND: internal/generator/templates/embed.go.tmpl
- FOUND: .planning/phases/08-background-jobs-production-readiness/08-05-SUMMARY.md
- FOUND: commit 8da0e48 (feat(08-05): add forge build command)
- FOUND: commit 106153d (feat(08-05): add forge deploy command)

---
*Phase: 08-background-jobs-production-readiness*
*Completed: 2026-02-19*
