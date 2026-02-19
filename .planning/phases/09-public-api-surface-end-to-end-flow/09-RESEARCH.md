# Phase 9: Public API Surface & End-to-End Flow - Research

**Researched:** 2026-02-19
**Domain:** Go module path rename, package restructuring, public API surface design, scaffold template fixes, forge.App builder pattern
**Confidence:** HIGH — based on direct codebase inspection, no third-party library unknowns

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Public package boundary**
- `schema/` moves from `internal/schema/` to root-level `schema/` — clean import path: `github.com/alternayte/forge/schema`
- Two-package mental model for users: `schema` is what you write, `forge` is what you call
- Boundary rule: if generated apps or user code imports it, it's public. If only the forge CLI binary needs it, it's internal
- Stays internal: CLI (`internal/cli`), code generator (`internal/generator`), AST parser (`internal/parser`), template scaffolding (`internal/scaffold`), watcher, toolsync, UI, config, migrate
- Module path changes from `github.com/forge-framework/forge` to `github.com/alternayte/forge` — matches the real repository location
- `forgetest` lives at `forge/forgetest` as a sub-package

**Runtime package design**
- Public `forge` package at repo root contains runtime types: Error, Transaction, App, TenantFromContext, RenderError
- Implementations move into the public package directly (not thin re-exports over internal/) — avoids Go anti-pattern of indirection, cleaner godoc, readable stack traces
- Auth helpers in `forge/auth` sub-package: HashPassword, RequireSession, LoginUser, etc.
- SSE helpers in `forge/sse` sub-package: MergeFragment, Redirect, limiter
- Notify hub in `forge/notify` sub-package: PostgreSQL LISTEN/NOTIFY fan-out
- Jobs in `forge/jobs` sub-package: River wrappers, worker registration
- Observability stays internal — users configure via forge.toml, framework auto-instruments. Custom spans use `go.opentelemetry.io/otel` directly
- Test helpers in `forge/forgetest` sub-package: NewTestDB, NewApp, PostDatastar

**Scaffold main.go wiring**
- main.go is fully working out of the box — `go run main.go` starts serving after generate + DB exists
- Scaffold-once, user-owned — same pattern as resources/ handlers and views. User edits freely, `forge generate --diff` shows what a fresh main.go would look like
- Uses `forge.App` builder pattern: `app := forge.New(cfg); app.RegisterResources(...); app.Listen(":8080")` — framework manages lifecycle (graceful shutdown, SSE draining, River workers, DB pool, OTel flush)
- main.go scaffolded by `forge generate` (not `forge init`) — it imports gen/ packages which don't exist until generate runs
- forge init scaffolds: project structure, forge.toml, example schema (Post resource with Title, Body, Status)

**End-to-end golden path**
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

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

## Summary

Phase 9 is a structural refactor with no new external library dependencies. The work divides into five distinct operations: (1) rename the Go module path in go.mod and all import statements, (2) physically move `internal/schema/` to `schema/` at the repo root, (3) create the public `forge` package at the repo root with sub-packages (`forge/auth`, `forge/sse`, `forge/notify`, `forge/jobs`, `forge/forgetest`), (4) update generator templates to emit the new `github.com/alternayte/forge/*` import paths, and (5) fix scaffold templates so `forge init` + `forge generate` produces a compilable, runnable main.go.

The critical insight from codebase inspection is that there are currently **three categories of `forge-framework/forge/internal` references** that all need updating in a single coordinated pass: (a) Go source files in the forge CLI itself (78+ import statements across ~45 files), (b) generator templates embedded in `internal/generator/templates/` (3 templates directly import `forge-framework/forge/internal/auth`), and (c) scaffold templates in `internal/scaffold/templates/` (main.go.tmpl references the old module path in comments). The module rename and package moves must happen atomically or the codebase will fail to build at every intermediate step.

The `forge.App` builder pattern is the most design-sensitive part of this phase. It must wrap the existing `internal/api` server setup (`SetupAPI`, `SetupHTML`) and lifecycle management that's currently scattered. The existing code already has clear seams: `api.SetupAPI()` takes a chi router + config + token/API key stores + middleware funcs, and `api.SetupHTML()` takes a session manager + route registration func. The `forge.App` type needs to collect all those dependencies and call them in the correct order.

**Primary recommendation:** Execute the module rename first (go.mod + all .go imports + all .tmpl files in one `sed`/scripted pass), then move the schema package, then create public packages by physically moving implementations rather than re-exporting, then update generator templates, and finally add `forge generate` main.go scaffolding and the `forge dev` auto-DB/auto-migrate behavior.

---

## Standard Stack

### Core
| Component | Current Location | Purpose | Move To |
|-----------|-----------------|---------|---------|
| Module path | `go.mod` | Go module identity | Change to `github.com/alternayte/forge` |
| schema package | `internal/schema/` | DSL types (Define, Field, etc.) | `schema/` (repo root) |
| forge package | (doesn't exist yet) | Runtime: App, Error, Transaction | `forge/` (repo root) |
| auth sub-package | `internal/auth/` | HashPassword, RequireSession, etc. | `forge/auth/` |
| sse sub-package | `internal/sse/` | SSELimiter | `forge/sse/` |
| notify sub-package | `internal/notify/` | PostgresHub, NotifyHub interface | `forge/notify/` |
| jobs sub-package | `internal/jobs/` | NewRiverClient, RunRiverMigrations | `forge/jobs/` |
| forgetest sub-package | `internal/forgetest/` | NewTestDB, NewTestPool, PostDatastar | `forge/forgetest/` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `go.mod` tooling | stdlib | Module rename | Single file edit to change module path |
| `golang.org/x/tools/cmd/goimports` | already in go.mod | Fix imports post-move | During import path rewrite |

### No New External Dependencies
This phase introduces no new third-party libraries. All packages being made public already exist in `internal/` with their dependencies. The move is purely structural.

---

## Architecture Patterns

### Recommended Directory Structure After Phase 9

```
github.com/alternayte/forge   (repo root)
├── schema/              # PUBLIC: DSL types (moved from internal/schema/)
│   ├── schema.go        # SchemaItem interface, FieldType, OnDeleteAction
│   ├── definition.go    # Definition, Define()
│   ├── field.go         # Field, field modifiers
│   ├── field_type.go    # FieldType constants
│   ├── hooks.go         # Hooks types
│   ├── modifier.go      # Modifier types
│   ├── option.go        # Option types
│   ├── permission.go    # Permission types
│   └── relationship.go  # Relationship types
├── forge/               # PUBLIC: Runtime package
│   ├── forge.go         # App builder: New(), RegisterResources(), Listen()
│   ├── errors.go        # Error, RenderError types
│   ├── transaction.go   # Transaction wrapper
│   ├── auth/            # Auth helpers (moved from internal/auth/)
│   │   ├── session.go   # NewSessionManager, SessionMiddleware, LoginUser, etc.
│   │   ├── password.go  # HashPassword, CheckPassword
│   │   ├── context.go   # UserFromContext, RoleFromContext, WithUserRole
│   │   ├── tenant.go    # TenantFromContext, WithTenant, TenantResolver impls
│   │   ├── html_middleware.go  # RequireSession, GetSession*
│   │   ├── apikey.go    # APIKeyStore interface
│   │   ├── token.go     # TokenStore interface
│   │   └── oauth.go     # OAuth helpers
│   ├── sse/             # SSE helpers (moved from internal/sse/)
│   │   └── limiter.go   # SSELimiter
│   ├── notify/          # Notify hub (moved from internal/notify/)
│   │   ├── hub.go       # PostgresHub, NotifyHub interface
│   │   └── subscription.go # Subscription, Event types
│   ├── jobs/            # Jobs (moved from internal/jobs/)
│   │   └── client.go    # NewRiverClient, RunRiverMigrations
│   └── forgetest/       # Test helpers (moved from internal/forgetest/)
│       ├── app.go       # NewApp, AppURL
│       ├── db.go        # NewTestDB, NewTestPool
│       └── datastar.go  # PostDatastar, ReadSSEEvents
├── internal/            # CLI only — never imported by generated apps
│   ├── api/             # SetupAPI, SetupHTML — used by forge.App internally
│   ├── cli/             # Cobra commands
│   ├── config/          # forge.toml parsing
│   ├── errors/          # CLI error formatting
│   ├── generator/       # Code generation engine
│   ├── migrate/         # Atlas wrapper
│   ├── observe/         # OTel/slog setup
│   ├── parser/          # go/ast schema parser
│   ├── scaffold/        # forge init templates
│   ├── toolsync/        # Tool binary downloads
│   ├── ui/              # Terminal styling
│   └── watcher/         # File watcher + dev server
├── main.go              # forge CLI binary entry point
├── go.mod               # module github.com/alternayte/forge
└── go.sum
```

### Pattern 1: Move Implementation Directly (Not Re-export)

**What:** Copy package contents from `internal/X/` to `forge/X/`, change the `package` declaration, delete the internal package, update all callers.

**Why:** The decision is to avoid thin re-export wrappers (the Go anti-pattern of `package auth` that just calls `internal/auth`). Callers see clean stack traces and godoc.

**Example — moving internal/auth/password.go:**
```go
// Before: internal/auth/password.go — package auth
// After:  forge/auth/password.go  — package auth (same package name, new path)

package auth

import (
    "errors"
    "golang.org/x/crypto/bcrypt"
)

// HashPassword hashes plaintext using bcrypt with BcryptCost.
func HashPassword(plaintext string) (string, error) { ... }
```

The package name `auth` stays the same; only the import path changes from `github.com/alternayte/forge/internal/auth` to `github.com/alternayte/forge/forge/auth`.

### Pattern 2: forge.App Builder

**What:** A `New(cfg Config) *App` constructor + method chain that collects all wiring dependencies and starts them on `Listen()`.

**Design from existing seams in internal/api/:**

The existing `api.SetupAPI()` takes:
- `chi.Router`
- `config.APIConfig`
- `auth.TokenStore`
- `auth.APIKeyStore`
- `recoveryMiddleware func(http.Handler) http.Handler`
- `registerRoutes func(huma.API)`

The existing `api.SetupHTML()` takes:
- `chi.Router`
- `HTMLServerConfig{SessionManager, RegisterRoutes}`

**Proposed forge.App API (Claude's discretion):**

```go
package forge

// Config is the runtime configuration loaded from forge.toml.
// Generated apps receive this from forge.New(cfg).
type Config struct {
    internal.Config // embed internal/config.Config — keeps config internal
}

// App is the forge application lifecycle manager.
type App struct {
    cfg      *internal.Config
    router   chi.Router
    pool     *pgxpool.Pool
    sm       *scs.SessionManager
    // ... other managed resources
}

// New creates a new App from a forge.toml Config.
// Use forge.LoadConfig("forge.toml") to load from disk, or
// pass a programmatic config for testing.
func New(cfg Config) *App { ... }

// RegisterResources wires all generated resource routes onto both the
// HTML router and the Huma API. Call before Listen().
//
//     app.RegisterResources(
//         genapi.RegisterAllRoutes,
//         genhtml.RegisterAllHTMLRoutes,
//         gen_actions.NewRegistry(),
//     )
func (a *App) RegisterResources(
    apiRoutes func(huma.API, *gen_actions.Registry),
    htmlRoutes func(chi.Router, *gen_actions.Registry),
    registry *gen_actions.Registry,
) *App { ... }

// Listen starts the HTTP server on the given address and blocks until
// the process receives SIGTERM/SIGINT. On shutdown: drains SSE connections,
// stops River workers, flushes OTel spans, closes DB pool.
func (a *App) Listen(addr string) error { ... }
```

**Challenge:** The `forge` package cannot import `internal/api` (internal packages can only be imported by packages within the same module with the same import path prefix). Since `forge/` is a sub-directory of the module root, it CAN import `internal/api`. So `forge.App.Listen()` can call `internal/api.SetupAPI()` and `internal/api.SetupHTML()` — this is the correct layering.

**Important:** `gen_actions.Registry` is generated into the project's `gen/` directory — `forge.App` cannot know its type at compile time. The `RegisterResources` signature must use `any` or an interface, not the concrete generated type. The most pragmatic approach: accept a registration function that receives the chi.Router and huma.API directly, letting the generated code do the wiring.

```go
// Simpler API that avoids needing to know about gen/ types:
func (a *App) RegisterAPIRoutes(fn func(huma.API)) *App
func (a *App) RegisterHTMLRoutes(fn func(chi.Router)) *App
```

Generated main.go would then look like:

```go
package main

import (
    "log"

    "github.com/alternayte/forge/forge"
    genapi  "myapp/gen/api"
    genhtml "myapp/gen/html"
    genmid  "myapp/gen/middleware"

    "myapp/resources/post"
)

func main() {
    cfg, err := forge.LoadConfig("forge.toml")
    if err != nil {
        log.Fatal(err)
    }

    registry := forge.NewRegistry()
    registry.Register("post", &post.PostActions{DB: cfg.DB()})

    app := forge.New(cfg).
        UseRecovery(genmid.Recovery).
        RegisterAPIRoutes(func(api huma.API) {
            genapi.RegisterAllRoutes(api, registry)
        }).
        RegisterHTMLRoutes(func(r chi.Router) {
            genhtml.RegisterAllHTMLRoutes(r, registry)
        })

    log.Fatal(app.Listen(":8080"))
}
```

### Pattern 3: Module Path Rename (Order of Operations)

**What:** Changing `github.com/forge-framework/forge` to `github.com/alternayte/forge` across 45+ files.

**Correct order:**
1. Edit `go.mod` — change module line
2. Run `find . -name "*.go" -o -name "*.tmpl" | xargs sed -i 's|github.com/forge-framework/forge|github.com/alternayte/forge|g'`
3. Move `internal/schema/` → `schema/` (update package paths in files)
4. Move `internal/auth/`, etc. → `forge/auth/`, etc.
5. Update generator templates that reference the old auth import path
6. Run `go build ./...` to verify

**Why this order matters:** `go.mod` must change before any internal `go build` can work. The rename is a global find-replace so must be done before package moves (otherwise you'd need to update twice).

### Pattern 4: Scaffold Template Changes

**Current state (from codebase inspection):**

`internal/scaffold/templates/schema.go.tmpl` currently generates:
```go
import "{{.Module}}/gen/schema"
```
This is wrong — after Phase 9, the schema package import is `github.com/alternayte/forge/schema`, not a path relative to the user's module.

`internal/scaffold/templates/main.go.tmpl` currently creates a placeholder main.go with commented wiring. This will be replaced by `forge generate` scaffolding.

`internal/scaffold/templates/go.mod.tmpl` needs to include `github.com/alternayte/forge` as a `require` directive.

**The Quick-1 conflict:** `forge init` currently scaffolds a `main.go` via `internal/scaffold/templates/main.go.tmpl`. Per the Phase 9 decision, the real main.go comes from `forge generate`. Resolution: `forge init` should scaffold a minimal placeholder main.go that prints "Run forge generate to wire your app" — similar to what the current template already does. The `forge generate` command then adds a real main.go (scaffold-once) alongside the gen/ code.

**Fixed schema.go.tmpl after Phase 9:**
```go
package {{.ExampleResource}}

import "github.com/alternayte/forge/schema"

var {{.ExampleResourceTitle}} = schema.Define("{{.ExampleResourceTitle}}",
    schema.String("Title").Required().MaxLen(200),
    schema.Text("Body"),
    schema.String("Status").Required().Default("draft"),
    schema.Timestamps(),
)
```

Note: The example resource is changing from "Product" to "Post" (with Title, Body, Status fields) per the decisions.

### Pattern 5: forge dev Auto-DB / Auto-Migrate

**What:** `forge dev` currently only watches files and re-runs generate. Per the decisions, it needs to additionally:
1. Auto-create the database if it doesn't exist (dev-mode only)
2. Auto-run `migrate diff` + `migrate up` on schema changes

**Current `internal/watcher/dev.go` structure:** `runGeneration()` method runs parse + generate. It needs a post-generate hook that runs migration.

**Auto-create DB:** Use `createdb` command or a pgx connection attempt to `postgres` database with `CREATE DATABASE IF NOT EXISTS`. The database URL comes from `forge.toml`. Need to parse host/port/user from the URL and connect to the `postgres` maintenance database to issue CREATE DATABASE.

**Auto-migrate flow in dev:**
```
1. runGeneration() — as today
2. detectSchemaChange() — compare gen/atlas/schema.hcl hash before/after
3. If changed: migrate.Diff(cfg, "auto", true) — generate migration
4. migrate.Up(cfg) — apply migration
5. Log applied SQL to terminal
```

### Anti-Patterns to Avoid

- **Re-exporting via thin wrappers:** Don't create `forge/auth/auth.go` that just calls `internal/auth.HashPassword()`. Move the implementation directly.
- **Circular imports:** The `forge` package (runtime) must NOT import `internal/config` or `internal/api` types in its public API surface. Use interfaces and the config struct embedded/copied. The `forge` package CAN call `internal/api` functions — the import direction is fine (public importing internal is allowed in Go within the same module).
- **Breaking the registry type:** The `forge.App` API should not import generated `gen/` packages since those are in the user's project, not the framework. Use function callbacks that receive `huma.API` and `chi.Router`.
- **Moving internal/api to public:** The `internal/api` package (`SetupAPI`, `SetupHTML`) stays internal — it's called by `forge.App` but is implementation detail, not user API.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Module path rename across 45+ files | Custom Go AST rewriter | `sed -i` or `gofmt -w` + regex | Standard refactoring pattern; sed is sufficient for a full string replacement |
| Package move tracking | Complex diff/merge | `git mv` for directory moves + manual import updates | Go tooling (`goimports`) handles import cleanup; git mv preserves history |
| Auto-create database | Custom PG driver code | `exec.Command("createdb", ...)` or raw `pgx` connection to `postgres` DB | Existing `internal/cli/db.go` already has database creation logic to reference |
| forge.App lifecycle | Custom signal handling | `signal.NotifyContext` (already used in `internal/cli/dev.go`) | Pattern already established in codebase |

**Key insight:** This phase is 80% mechanical file moves and string replacements. The design work is concentrated in (a) the `forge.App` API signature and (b) what `forge generate` puts in main.go.

---

## Common Pitfalls

### Pitfall 1: Partial Module Rename Leaves Uncompilable State
**What goes wrong:** If the module rename is done file-by-file over multiple commits, every intermediate commit fails `go build ./...`. CI breaks for the entire duration.
**Why it happens:** Go requires all import paths to be self-consistent — one file referencing the old module path makes the whole thing fail to resolve.
**How to avoid:** Do the module rename in a single commit. Use a script: `find . -name "*.go" | xargs sed -i 's|github.com/forge-framework/forge|github.com/alternayte/forge|g'` then update go.mod, then `go build ./...` before committing.
**Warning signs:** `go build ./...` returns "cannot find module" after a partial rename.

### Pitfall 2: Generator Templates Still Reference Old Auth Import Path
**What goes wrong:** `forge generate` produces code with `github.com/forge-framework/forge/internal/auth` imports in generated files — those generated files won't compile because the module no longer exists at that path.
**Why it happens:** Three template files embed the hard-coded old module path: `actions.go.tmpl`, `queries.go.tmpl`, `scaffold_jobs.go.tmpl`. These won't be caught by a simple find-replace on `.go` files if `.tmpl` files are excluded.
**How to avoid:** The module rename script MUST include `*.tmpl` files. The new import path will be `github.com/alternayte/forge/forge/auth` (note: `forge/auth` not `internal/auth`).
**Warning signs:** Generated code doesn't compile after module rename; error is "cannot find module `github.com/forge-framework/forge/internal/auth`".

### Pitfall 3: forge.App Cannot Import gen/ Package Types
**What goes wrong:** If `forge.App.RegisterResources()` takes `*gen_actions.Registry` as a parameter type, the framework package would need to import the user's generated code — an import cycle that Go will reject.
**Why it happens:** Generated code lives in the user's module (`myapp/gen/actions`), not in the forge framework. Framework packages cannot import per-project generated code.
**How to avoid:** `forge.App` registration methods accept callback functions (`func(huma.API)`, `func(chi.Router)`) rather than concrete generated types. The generated main.go calls those callbacks with the generated registrations. The Registry type itself can be defined in `forge` (a simple `map[string]any`) if needed.
**Warning signs:** `import cycle not allowed` compiler error; or generated code referencing framework types that reference generated types.

### Pitfall 4: schema.go.tmpl Generates Wrong Import Path
**What goes wrong:** `forge init` generates `resources/post/schema.go` with `import "myapp/gen/schema"` — this was the old broken import pattern. The schema package is now `github.com/alternayte/forge/schema`, not a path under the user's module.
**Why it happens:** The existing `schema.go.tmpl` template has `import "{{.Module}}/gen/schema"` hardcoded. After Phase 9, the schema package moves to the framework's public API.
**How to avoid:** Update `internal/scaffold/templates/schema.go.tmpl` to emit `import "github.com/alternayte/forge/schema"` (no template variable — it's always the framework's package).
**Warning signs:** `forge init` + `forge generate` produces schema files that don't compile because the import doesn't resolve.

### Pitfall 5: go.mod in Scaffolded Projects Missing forge Dependency
**What goes wrong:** `forge init myapp` generates a `go.mod` with just the project module — no `require github.com/alternayte/forge` line. The generated main.go and resources/*/schema.go files import forge packages that aren't in go.mod.
**Why it happens:** `internal/scaffold/templates/go.mod.tmpl` currently generates a minimal `module X\n\ngo Y.Z` with no require directives.
**How to avoid:** Update `go.mod.tmpl` to include `require github.com/alternayte/forge v0.X.Y` (with a pinned version or `latest`). The template needs a `ForgeVersion` field in `ProjectData`. For local development, the template should use `replace github.com/alternayte/forge => /path/to/local/forge` or the developer runs `go get`.

**Practical approach:** `go.mod.tmpl` generates with `require github.com/alternayte/forge latest` and instructs developer to run `go mod tidy` as part of the next-steps output.

### Pitfall 6: forge dev Auto-Migrate Destroys Data on Schema Changes
**What goes wrong:** `forge dev` auto-runs `migrate diff + up` on every schema change. If a destructive migration is generated (e.g., column rename → drop+add), data is silently lost in dev.
**Why it happens:** The existing `migrate.Diff()` has a `force bool` parameter — when `force=false`, it checks for and rejects destructive changes. But in dev mode we want convenience.
**How to avoid:** In dev mode, `migrate.Diff()` should be called with `force=true` but a warning printed to terminal showing the generated SQL before applying. Developer sees "WARNING: destructive change detected: DROP COLUMN old_name" before it runs. This matches the Rails dev experience.
**Warning signs:** Developers losing column data during development without warning.

### Pitfall 7: forge.App Lifecycle Order
**What goes wrong:** `app.Listen()` starts the HTTP server before the DB pool is connected, or shuts down the HTTP server after the DB pool is closed (leaving in-flight requests unable to complete queries).
**Why it happens:** Lifecycle order is subtle: start DB pool → start River workers → start HTTP server → on shutdown: stop accepting → drain in-flight → stop workers → close pool.
**How to avoid:** Use `errgroup` or explicit sequencing in `Listen()`. The shutdown sequence must respect the dependency order. Context cancellation propagates through: `signal.NotifyContext` → HTTP server graceful shutdown → River client stop → pool close.
**Warning signs:** "pool is closed" errors appearing during graceful shutdown.

---

## Code Examples

Verified patterns from codebase inspection:

### Current packages being made public — what they export

**internal/auth (→ forge/auth):**
```go
// session.go
func NewSessionManager(pool *pgxpool.Pool, isDev bool) *scs.SessionManager
func SessionMiddleware(sm *scs.SessionManager) func(http.Handler) http.Handler
// html_middleware.go
func RequireSession(sm *scs.SessionManager) func(http.Handler) http.Handler
func LoginUser(sm *scs.SessionManager, r *http.Request, userID, email string) error
func LogoutUser(sm *scs.SessionManager, r *http.Request) error
// password.go
func HashPassword(plaintext string) (string, error)
func CheckPassword(plaintext, hash string) error
// context.go
func UserFromContext(ctx context.Context) uuid.UUID
func RoleFromContext(ctx context.Context) string
func WithUserRole(ctx context.Context, userID uuid.UUID, role string) context.Context
// tenant.go
func TenantFromContext(ctx context.Context) (uuid.UUID, bool)
func WithTenant(ctx context.Context, id uuid.UUID) context.Context
type TenantResolver interface { Resolve(r *http.Request) (uuid.UUID, error) }
type HeaderTenantResolver struct{ Header string }
type SubdomainTenantResolver struct{}
type PathTenantResolver struct{}
func TenantMiddleware(resolver TenantResolver) func(http.Handler) http.Handler
```

**internal/notify (→ forge/notify):**
```go
type NotifyHub interface {
    Subscribe(channel string, tenantID uuid.UUID) *Subscription
    Publish(ctx context.Context, channel string, tenantID uuid.UUID, payload []byte) error
    Start(ctx context.Context) error
}
type PostgresHub struct { ... }
func NewPostgresHub(connConfig *pgx.ConnConfig, db Executor, bufferSize int) *PostgresHub
type Subscription struct { Events <-chan Event; cancel func() }
type Event struct { Channel string; Payload json.RawMessage }
```

**internal/sse (→ forge/sse):**
```go
type SSELimiter struct { ... }
func NewSSELimiter(maxTotal, maxPerUser int) *SSELimiter
func (l *SSELimiter) Acquire(userID string) (release func(), err error)
```

**internal/jobs (→ forge/jobs):**
```go
func NewRiverClient(pool *pgxpool.Pool, cfg config.JobsConfig, workers *river.Workers) (*river.Client[pgx.Tx], error)
func RunRiverMigrations(ctx context.Context, pool *pgxpool.Pool) error
```
Note: `NewRiverClient` currently takes `config.JobsConfig` (internal type). When moved to `forge/jobs`, either (a) define a `jobs.Config` struct in the public package, or (b) accept the fields directly. Option (a) is cleaner.

**internal/forgetest (→ forge/forgetest):**
```go
func NewApp(t *testing.T, handler http.Handler) *httptest.Server
func AppURL(srv *httptest.Server, path string) string
func NewTestDB(t *testing.T, opts ...func(*TestDBConfig)) *sql.DB
func NewTestPool(t *testing.T, opts ...func(*TestDBConfig)) *pgxpool.Pool
func PostDatastar(t *testing.T, srv *httptest.Server, path string, signals any) *http.Response
func ReadSSEEvents(t *testing.T, resp *http.Response) []SSEEvent
```

### Generator templates that need auth import path update

Three templates currently hardcode `github.com/forge-framework/forge/internal/auth`:

1. `internal/generator/templates/actions.go.tmpl` line 12:
   ```
   forgeauth "github.com/forge-framework/forge/internal/auth"
   ```
   Must become: `forgeauth "github.com/alternayte/forge/forge/auth"`

2. `internal/generator/templates/queries.go.tmpl` line 9:
   ```
   forgeauth "github.com/forge-framework/forge/internal/auth"
   ```
   Must become: `forgeauth "github.com/alternayte/forge/forge/auth"`

3. `internal/generator/templates/scaffold_jobs.go.tmpl` line 7:
   ```
   forgeauth "github.com/forge-framework/forge/internal/auth"
   ```
   Must become: `forgeauth "github.com/alternayte/forge/forge/auth"`

### forge generate output with color-coding

The existing `internal/cli/generate.go` uses `ui.Success()` for output. The new verbose file-by-file output with color coding (green=new, yellow=updated, dim=unchanged) should use the existing `ui` package styles. Reference from `internal/ui/styles.go` (not shown but referenced by CLI files).

### forge dev auto-migrate logic (new in Phase 9)

In `internal/watcher/dev.go`, `runGeneration()` needs to:
```go
func (d *DevServer) runGenerationAndMigrate() error {
    // 1. Hash gen/atlas/schema.hcl before
    beforeHash := hashSchemaFile(d.ProjectRoot)

    // 2. Generate as today
    if err := d.runGeneration(); err != nil {
        return err
    }

    // 3. Hash after
    afterHash := hashSchemaFile(d.ProjectRoot)
    if beforeHash == afterHash {
        return nil // No schema changes, skip migration
    }

    // 4. Auto-create DB if missing
    if err := d.ensureDatabase(); err != nil {
        fmt.Println(ui.Warn("Could not auto-create database: " + err.Error()))
        return nil // Non-fatal
    }

    // 5. Run migrate diff (force=true in dev, but show warning)
    migFile, err := migrate.Diff(migCfg, "auto", true /*force*/)
    if err != nil {
        return err
    }
    // Show SQL before applying
    sql, _ := os.ReadFile(migFile)
    fmt.Println(ui.DimStyle.Render("Migration: " + string(sql)))

    // 6. Apply
    output, err := migrate.Up(migCfg)
    fmt.Println(output)
    return err
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `github.com/forge-framework/forge` | `github.com/alternayte/forge` | Phase 9 | All internal imports update; all generated code imports update |
| `internal/schema/` | `schema/` (root) | Phase 9 | User schemas import `github.com/alternayte/forge/schema` |
| Scattered runtime in `internal/` | `forge/` public sub-packages | Phase 9 | Users can import auth, sse, notify, jobs helpers directly |
| `forge init` creates placeholder main.go | `forge generate` creates real main.go | Phase 9 | `go run main.go` works immediately after generate |
| `forge dev` = watch + generate only | `forge dev` = watch + generate + auto-migrate + auto-create-DB | Phase 9 | One command gets app running |

---

## Open Questions

1. **forge.App `jobs.Config` type**
   - What we know: `NewRiverClient` currently takes `config.JobsConfig` (internal type). When `forge/jobs` is made public, it cannot import `internal/config`.
   - What's unclear: Should `forge/jobs` define its own `Config` struct (duplicating fields), or should `forge.App` pass the fields as primitives?
   - Recommendation: Define `jobs.Config{ Enabled bool; Queues map[string]int }` in `forge/jobs` — it's small and already matches the TOML structure. `forge.App` internally bridges between `internal/config.JobsConfig` and `jobs.Config`.

2. **forge init main.go stub vs forge generate main.go**
   - What we know: Quick-1 created a main.go template in `internal/scaffold/templates/main.go.tmpl` that currently emits a placeholder. `forge generate` needs to scaffold main.go when it doesn't exist.
   - What's unclear: Should `forge init` stop scaffolding main.go entirely, or keep a minimal stub?
   - Recommendation: Keep `forge init` creating a minimal main.go stub (shows "run forge generate") so `go run main.go` compiles immediately after init. `forge generate` then overwrites it with the real wiring (scaffold-once: if main.go exists, skip unless `--diff`). This matches the stated "scaffold-once, user-owned" pattern.

3. **forge generate main.go template placement**
   - What we know: `forge generate` currently does NOT scaffold main.go. The generator output is all in `gen/`. Main.go would need to be written into the project root, not `gen/`.
   - What's unclear: Where does the main.go template live — in `internal/scaffold/templates/` or `internal/generator/templates/`?
   - Recommendation: Add `main.go.tmpl` to `internal/generator/templates/` (alongside other generator templates) and call it from `generator.Generate()` with the scaffold-once pattern (check if file exists before writing). This keeps generate templates alongside generate logic.

4. **Example resource: Post vs Product**
   - What we know: Current `forge init` scaffolds a "product" resource. Decisions say "Post resource with Title, Body, Status".
   - What's unclear: Does the scaffold also need to update the `ExampleResource` / `ExampleResourceTitle` in `ProjectData`?
   - Recommendation: Change `internal/cli/init.go` to use `ExampleResource: "post"`, `ExampleResourceTitle: "Post"` and update the schema template accordingly.

5. **forgetest dependency path hack**
   - What we know: `internal/forgetest/db.go` uses `runtime.Caller(0)` to resolve paths relative to its own source file location. After the move to `forge/forgetest/db.go`, the `repoRoot` calculation will be wrong (was two levels up, now three).
   - What's unclear: Whether to keep the `runtime.Caller` hack or replace with a better approach.
   - Recommendation: Update the path calculation in `DefaultTestDBConfig()` to account for the new depth (`filepath.Dir(filename), "../../.."` instead of `"../.."`). Long-term, accept opts to make this configurable — which is already supported via `WithMigrationDir` and `WithAtlasBin`.

---

## Sources

### Primary (HIGH confidence)
- Direct inspection of `/Users/nathananderson-tennant/Development/forge-go/` — all package contents, templates, import paths
- Go language specification on internal packages: packages under `internal/` can only be imported by code with the same import path prefix
- Go module documentation: `go.mod` module directive controls the module path; `go build ./...` validates all imports

### Secondary (MEDIUM confidence)
- Go community convention: public framework packages should not re-export internal packages via thin wrappers; implementations should live directly in the public package
- River library usage patterns: River client requires `pgx.Tx` generic type — public `forge/jobs` package will expose this type to users

### Tertiary (LOW confidence)
- Exact `forge.App` API ergonomics — no prior art in this codebase; design based on analogous frameworks (chi, echo, gin builder patterns)

---

## Metadata

**Confidence breakdown:**
- Module rename mechanics: HIGH — standard Go operation, fully verified by inspecting all import sites
- Package move mechanics: HIGH — direct inspection of all packages being moved; no hidden dependencies
- forge.App design: MEDIUM — no existing implementation; design is reasoned from existing seams but will need iteration
- Generator template fixes: HIGH — all three affected template lines identified and verified
- scaffold template fixes: HIGH — all affected templates inspected; changes are clear
- forge dev auto-migrate: MEDIUM — new behavior; integration with Atlas CLI is understood but the exact auto-create-DB approach needs validation against the existing `internal/cli/db.go` patterns

**Research date:** 2026-02-19
**Valid until:** 2026-03-19 (stable Go tooling; no fast-moving dependencies involved)
