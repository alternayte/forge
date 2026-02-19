# Phase 9: Public API Surface & End-to-End Flow - Context

**Gathered:** 2026-02-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Make schema/ public (move from internal/), create forge runtime package with auth/sse/notify/jobs sub-packages, fix scaffold templates (go.mod dependency, schema import path, main.go server wiring), update all internal imports, change module path to github.com/alternayte/forge. End state: `forge init myapp && forge generate && forge dev` produces a running app with working CRUD.

</domain>

<decisions>
## Implementation Decisions

### Public package boundary
- `schema/` moves from `internal/schema/` to root-level `schema/` — clean import path: `github.com/alternayte/forge/schema`
- Two-package mental model for users: `schema` is what you write, `forge` is what you call
- Boundary rule: if generated apps or user code imports it, it's public. If only the forge CLI binary needs it, it's internal
- Stays internal: CLI (`internal/cli`), code generator (`internal/generator`), AST parser (`internal/parser`), template scaffolding (`internal/scaffold`), watcher, toolsync, UI, config, migrate
- Module path changes from `github.com/forge-framework/forge` to `github.com/alternayte/forge` — matches the real repository location
- `forgetest` lives at `forge/forgetest` as a sub-package

### Runtime package design
- Public `forge` package at repo root contains runtime types: Error, Transaction, App, TenantFromContext, RenderError
- Implementations move into the public package directly (not thin re-exports over internal/) — avoids Go anti-pattern of indirection, cleaner godoc, readable stack traces
- Auth helpers in `forge/auth` sub-package: HashPassword, RequireSession, LoginUser, etc.
- SSE helpers in `forge/sse` sub-package: MergeFragment, Redirect, limiter
- Notify hub in `forge/notify` sub-package: PostgreSQL LISTEN/NOTIFY fan-out
- Jobs in `forge/jobs` sub-package: River wrappers, worker registration
- Observability stays internal — users configure via forge.toml, framework auto-instruments. Custom spans use `go.opentelemetry.io/otel` directly
- Test helpers in `forge/forgetest` sub-package: NewTestDB, NewApp, PostDatastar

### Scaffold main.go wiring
- main.go is fully working out of the box — `go run main.go` starts serving after generate + DB exists
- Scaffold-once, user-owned — same pattern as resources/ handlers and views. User edits freely, `forge generate --diff` shows what a fresh main.go would look like
- Uses `forge.App` builder pattern: `app := forge.New(cfg); app.RegisterResources(...); app.Listen(":8080")` — framework manages lifecycle (graceful shutdown, SSE draining, River workers, DB pool, OTel flush)
- main.go scaffolded by `forge generate` (not `forge init`) — it imports gen/ packages which don't exist until generate runs
- forge init scaffolds: project structure, forge.toml, example schema (Post resource with Title, Body, Status)

### End-to-end golden path
- 3-step flow: `forge init myapp` -> (write schema or use example) -> `forge generate` -> `forge dev`
- `forge dev` auto-creates database if it doesn't exist (dev-mode convenience, not production behavior)
- `forge dev` auto-runs migrate diff + migrate up on schema changes, logging applied SQL
- Production deployments use explicit 4-step process with reviewed migrations
- forge generate output: every file listed with color-coding (green=new, yellow=updated, dim=unchanged) and icons. `--quiet` flag for summary-only
- Example Post schema included by forge init so generate works immediately — show don't tell, like Rails/Phoenix

### Claude's Discretion
- Exact forge.App method signatures and configuration API
- How to handle the forge init main.go stub (Quick-1 created one) — may need to be removed or made minimal since real main.go comes from forge generate
- Import alias strategy for internal packages that move to public
- Order of operations for the module path rename (go.mod, all import statements, templates)

</decisions>

<specifics>
## Specific Ideas

- "Two import patterns: schema for definitions, forge for everything runtime" — consistent with PRD Sections 5.1, 6, 8, 10, 17
- "Developers trust tools they can see working" — verbose file-by-file output as default, quiet flag for CI
- "forge dev should feel like Rails — auto-create DB, auto-migrate, just works"
- "forge generate --diff can show what a fresh main.go would look like" — same pattern as re-scaffolding views (PRD Section 5.6)
- Progressive complexity: scaffolded main.go works immediately, developer adds custom actions/middleware/jobs as they grow

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 09-public-api-surface-end-to-end-flow*
*Context gathered: 2026-02-19*
