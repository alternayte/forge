---
phase: 09-public-api-surface-end-to-end-flow
plan: 02
subsystem: api
tags: [auth, sse, notify, jobs, forgetest, public-api, package-structure]

# Dependency graph
requires:
  - phase: 09-01
    provides: Module rename to github.com/alternayte/forge, schema/ at repo root

provides:
  - forge/auth/ public package (auth, sessions, OAuth, tokens, API keys, tenant)
  - forge/sse/ public package (SSE connection limiter)
  - forge/notify/ public package (PostgreSQL LISTEN/NOTIFY hub)
  - forge/jobs/ public package with own Config struct
  - forge/forgetest/ public package with corrected runtime.Caller depth

affects:
  - 09-03
  - 09-04
  - 09-05
  - generated application code that imports forge/auth, forge/sse, forge/notify, forge/jobs, forge/forgetest

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Direct move of implementations (no thin re-export wrappers)
    - Public packages have own types (jobs.Config owns its fields, no internal type leakage)
    - runtime.Caller depth updated when file moves depth in tree

key-files:
  created:
    - forge/auth/apikey.go
    - forge/auth/context.go
    - forge/auth/html_middleware.go
    - forge/auth/oauth.go
    - forge/auth/password.go
    - forge/auth/session.go
    - forge/auth/tenant.go
    - forge/auth/token.go
    - forge/sse/limiter.go
    - forge/notify/hub.go
    - forge/notify/subscription.go
    - forge/jobs/client.go
    - forge/forgetest/app.go
    - forge/forgetest/datastar.go
    - forge/forgetest/db.go
  modified:
    - internal/api/server.go
    - internal/api/html_server.go
    - internal/api/middleware/auth.go
    - internal/generator/templates/actions.go.tmpl
    - internal/generator/templates/queries.go.tmpl
    - internal/generator/templates/scaffold_jobs.go.tmpl

key-decisions:
  - "forge/jobs has own Config struct (Enabled bool, Queues map[string]int) — public API cannot expose internal/config.JobsConfig"
  - "forge/forgetest/db.go uses ../../.. (3 levels) for repo root — one more level than internal/forgetest which used ../.."
  - "All implementations moved directly (no indirection wrappers) — locked decision from plan"

patterns-established:
  - "When moving a package, update runtime.Caller depth to match new tree depth"
  - "Public packages define their own Config types rather than re-exporting internal types"

requirements-completed: []

# Metrics
duration: 3min
completed: 2026-02-19
---

# Phase 09 Plan 02: Move Internal Packages to Public forge/ API Summary

**Five runtime packages (auth, sse, notify, jobs, forgetest) moved from internal/ to public forge/ sub-packages with corrected import paths across all callers and generator templates**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-19T21:04:20Z
- **Completed:** 2026-02-19T21:07:54Z
- **Tasks:** 2
- **Files modified:** 21

## Accomplishments

- Moved 8 auth files (apikey, context, html_middleware, oauth, password, session, tenant, token) to forge/auth/
- Moved sse/limiter.go to forge/sse/ and notify/ (hub, subscription) to forge/notify/
- Moved jobs/client.go to forge/jobs/ with own Config struct replacing internal/config.JobsConfig dependency
- Moved forgetest/ (app, datastar, db) to forge/forgetest/ with corrected runtime.Caller path depth (3 levels up)
- Updated all callers: internal/api/server.go, html_server.go, middleware/auth.go
- Updated generator templates (actions, queries, scaffold_jobs) to emit forge/auth import paths
- go build ./... and go vet ./... both pass with zero errors

## Task Commits

Each task was committed atomically:

1. **Task 1: Move auth, sse, and notify packages to forge/** - `dd374b8` (feat)
2. **Task 2: Move jobs and forgetest packages to forge/** - `168c80f` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `forge/auth/*.go` (8 files) - Public auth package: sessions, OAuth, passwords, tokens, API keys, tenant context
- `forge/sse/limiter.go` - Public SSE connection limiter (global + per-user caps)
- `forge/notify/hub.go` - Public PostgreSQL LISTEN/NOTIFY hub
- `forge/notify/subscription.go` - Public Subscription type and Event definitions
- `forge/jobs/client.go` - Public jobs package with own Config struct, NewRiverClient, RunRiverMigrations
- `forge/forgetest/app.go` - Public test app server helper
- `forge/forgetest/datastar.go` - Public Datastar SSE test helpers
- `forge/forgetest/db.go` - Public test DB helper with corrected ../../.. repo root path
- `internal/api/server.go` - Updated import to forge/auth
- `internal/api/html_server.go` - Updated import to forge/auth
- `internal/api/middleware/auth.go` - Updated import to forge/auth
- `internal/generator/templates/actions.go.tmpl` - Updated to emit forge/auth path
- `internal/generator/templates/queries.go.tmpl` - Updated to emit forge/auth path
- `internal/generator/templates/scaffold_jobs.go.tmpl` - Updated to emit forge/auth path

## Decisions Made

- `forge/jobs` defines its own `Config struct { Enabled bool; Queues map[string]int }` — public packages cannot expose internal types in their API surface
- `forge/forgetest/db.go` uses `../../..` for repo root — the file is now 3 directories deep (forge/forgetest/db.go) vs the old 2 (internal/forgetest/db.go)
- All implementations moved directly per the locked plan decision — no thin re-export wrappers

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All five packages are now part of the public forge/ API surface
- Generated apps can import github.com/alternayte/forge/forge/auth, forge/sse, forge/notify, forge/jobs, forge/forgetest directly
- Generator templates emit correct forge/auth import paths
- Ready for Plan 03: internal/api package exposure and server assembly

---
*Phase: 09-public-api-surface-end-to-end-flow*
*Completed: 2026-02-19*
