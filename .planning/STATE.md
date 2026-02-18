# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-16)

**Core value:** Schema is the single source of truth — define a resource once and everything is generated automatically with zero manual sync.
**Current focus:** Phase 2: Code Generation Engine

## Current Position

Phase: 7 of 8 (Advanced Data Features)
Plan: 2 of 5 in current phase
Status: In Progress
Last activity: 2026-02-18 — Completed 07-02-PLAN.md (Soft delete query mods, partial unique indexes, Delete/Restore actions)

Progress: [████████░░] 66%

## Performance Metrics

**Velocity:**
- Total plans completed: 18
- Average duration: 3.3 minutes
- Total execution time: 0.95 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 | 5 | 18.9m | 3.8m |
| 02 | 5 | 18.0m | 3.6m |
| 03 | 3 | 10.2m | 3.4m |
| 04 | 3 | 6.1m | 2.0m |
| 05 | 4 (so far) | ~15m | ~3.8m |

**Recent Executions:**

| Phase | Plan | Duration | Tasks | Files |
|-------|------|----------|-------|-------|
| 07 | 01 | 2m | 2 | 6 |
| 06 | 09 | 3m | 2 | 4 |
| 06 | 07 | 4m | 2 | 5 |
| 06 | 04 | 2m | 2 | 3 |
| 06 | 02 | 2m | 2 | 5 |
| 06 | 01 | 2m | 2 | 6 |
| 05 | 04 | 4m | 2 | 3 |
| 05 | 03 | 5m | 2 | 6 |
| 05 | 01 | 5m | 2 | 8 |
| 05 | 02 | 3m | 2 | 6 |
| 04 | 03 | 2.1m | 2 | 5 |
| 04 | 02 | 1.9m | 2 | 4 |
| 04 | 01 | 2.1m | 2 | 4 |
| 03 | 02 | 3.7m | 2 | 10 |
| Phase 06 P05 | 2 | 2 tasks | 4 files |
| Phase 06 P03 | 5 | 2 tasks | 4 files |
| Phase 06 P06 | 2 | 2 tasks | 4 files |
| Phase 06 P08 | 2 | 2 tasks | 4 files |
| Phase 07 P02 | 2m | 2 tasks | 3 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Huma for OpenAPI (not custom generation) — Best OpenAPI 3.1 from code in Go; avoids months of spec compliance work
- Atlas for migrations (not custom diffing) — Declarative schema diffing is multi-year effort; Atlas handles edge cases
- go/ast parsing (not compile-and-execute) — Solves bootstrapping; schemas parseable before gen/ exists
- Action layer shared by HTML + API — Prevents business logic duplication between Datastar and Huma handlers
- Scaffolded-once views (resources/) vs always-regenerated (gen/) — Clear ownership boundary
- [Phase 01]: Fluent method chaining API for schema definition
- [Phase 01]: SchemaItem marker interface for variadic Define() args
- [Phase 01-foundation-schema-dsl]: IR uses strings (not enums) for types to decouple from schema package
- [Phase 01-foundation-schema-dsl]: Error codes use E0xx/E1xx ranges (parser/validation)
- [Phase 01-foundation-schema-dsl]: Help links use CLI format (forge help E001) for offline-first design
- [Phase 01-foundation-schema-dsl]: Method chains traverse DOWN from outermost call to find root constructor
- [Phase 01-foundation-schema-dsl]: Parser collects all errors in single pass for batch reporting
- [Phase 01]: forge init supports both new directory and current directory initialization modes
- [Phase 01]: forge.toml uses commented template style with all options visible
- [Phase 01-foundation-schema-dsl]: Standalone Tailwind CLI binary (zero npm) from GitHub releases
- [Phase 01-foundation-schema-dsl]: On-demand tool sync (tools downloaded when needed, not upfront)
- [Phase 01-foundation-schema-dsl]: Memory buffer download for checksum verification before disk write
- [Phase 02-01]: Use golang.org/x/tools/imports for automatic import management instead of manual tracking
- [Phase 02-01]: Snake case handles acronyms as units (HTTPStatus->http_status, ProductID->product_id)
- [Phase 02-01]: Filter struct only includes fields with Filterable modifier
- [Phase 02-01]: Update structs use all pointer fields for partial updates
- [Phase 02-01]: Create structs use non-pointer for required, pointer for optional
- [Phase 02-02]: Enum types map to text with CHECK constraints (not PostgreSQL enum type) for simpler migration story
- [Phase 02-02]: ID field excluded from factory builders (auto-generated via gen_random_uuid())
- [Phase 02-02]: MaxLen modifier overrides default varchar(255) length for String/Email/Slug types
- [Phase 02-03]: Only display generated directories that exist (no "0 files" noise)
- [Phase 02-03]: Clean gen/ directory before generation for idempotent output
- [Phase 02-04]: Use same database URL for dev-url in Atlas diff (Atlas creates temporary schemas)
- [Phase 02-04]: Delete rejected migration files on destructive warning (Atlas creates file before we can check it)
- [Phase 02-04]: Regex-based destructive detection instead of SQL parsing (simple patterns cover 95% of cases)
- [Phase 02-04]: Include line numbers in destructive change warnings (easy to locate problematic SQL)
- [Phase 02-05]: 300ms debounce for file changes to batch multi-file saves
- [Phase 02-05]: Skip Chmod events unconditionally (always noise from Spotlight/antivirus/editors)
- [Phase 02-05]: Watch parent directories not individual files (handles atomic writes correctly)
- [Phase 02-05]: Exclude gen/ paths from triggering regeneration (prevent infinite loops)
- [Phase 03-03]: Reuse migrate.Up() directly instead of shelling out (avoids recursive binary invocation)
- [Phase 03-03]: Parse host/port/dbname from database URL for PostgreSQL client tool flags
- [Phase 03-03]: Treat 'already exists' as no-op for idempotent database create command
- [Phase 03-03]: Pass DATABASE_URL env var to seed file for database connection
- [Phase 03-01]: Separate types.go file prevents FieldError redeclaration across multiple resources
- [Phase 03-01]: Bob query mods use sm.QueryMod[*psql.SelectQuery] generic type for type safety
- [Phase 03-01]: Type-specific query filters: string → Contains (ILIKE), numeric/date → GTE/LTE
- [Phase 03-02]: Pagination utilities generated once (not per-resource) since logic is generic
- [Phase 03-02]: Cursor includes ID + sort field + sort value for uniqueness and stability
- [Phase 03-02]: Page size capped at 100 to prevent accidental large queries
- [Phase 03-02]: ProjectRoot added to GenerateConfig for files outside gen/ directory
- [Phase 04-01]: Error struct with 5 fields (Status, Code, Message, Detail, Err) for comprehensive error handling
- [Phase 04-01]: Static error templates (not per-resource) following validation_types.go.tmpl pattern
- [Phase 04-01]: NewValidationError accepts interface{} to avoid circular import with gen/validation
- [Phase 04-01]: MapDBError uses errors.As for type-safe pgconn.PgError detection
- [Phase 04-02]: Per-resource Actions interface with 5 CRUD methods (List, Get, Create, Update, Delete)
- [Phase 04-02]: DefaultActions implementation calls generated validation directly (not validator interfaces)
- [Phase 04-02]: Placeholder TODO comments for Bob query execution (interface ready, execution pending)
- [Phase 04-02]: Registry uses explicit Register/Get methods (no init() magic)
- [Phase 04-02]: DB interface with pgx types (not generic interface{}) for generated code
- [Phase 04-03]: SSE errors use HTTP 200 with datastar-merge-fragments event for toast notifications
- [Phase 04-03]: JSON errors follow RFC 9457 shape using simple fmt.Fprintf (no json.Marshal dependency)
- [Phase 04-03]: HTML errors use inline template string (templ templates come in Phase 6)
- [Phase 04-03]: Import stdlib errors as stderrors to avoid conflict with gen/errors package
- [Phase 05-01]: List output structs use header:"Link" tag field (not huma.SetHeader) for RFC 8288 Link header — canonical Huma v2 pattern
- [Phase 05-01]: buildAPILinkHeader function generated in types.go (not as a funcmap helper) — link header logic lives inside generated code
- [Phase 05-01]: humaValidationTag builds from MinLen, MaxLen, Min, Max, and Enum modifiers in one function
- [Phase 05-01]: funcmap not/join helpers added as template primitives for boolean negation and string joining
- [Phase 05-02]: huma.API passed to AuthMiddleware constructor so WriteErr can produce structured 401 responses
- [Phase 05-02]: validateBearerToken/validateAPIKey return updated huma.Context (not mutate) to thread context through middleware chain
- [Phase 05-02]: Phase 5 rate limiting uses Default tier for all requests — tiered enforcement deferred to Plan 03 server assembly
- [Phase 05-02]: CORSMiddleware logs warning and disables credentials when wildcard origin combined with AllowCredentials (CORS spec violation guard)
- [Phase 05-03]: SetupAPI accepts func(huma.API) for route registration (not func(huma.API, *actions.Registry)) — gen/actions is user-generated code, forge tool cannot import it; caller uses a closure
- [Phase 05-03]: wrapHTTPMiddleware uses humachi.Unwrap to bridge http.Handler middlewares (CORS, rate limit) into Huma middleware chain
- [Phase 05-03]: api_register_all.go.tmpl uses registry.Get(name) with type assertion to {Name}Actions — generated dispatcher, not per-resource
- [Phase 05-04]: Build huma.OpenAPI struct directly from IR (no HTTP adapter) for spec export — avoids adding chi or net/http dependencies to CLI
- [Phase 05-04]: apiRoutes() separated from runRoutes() so Phase 6 can add htmlRoutes() to same forge routes output without refactoring
- [Phase 05-04]: routeKebab/routePlural/routeLowerCamel duplicated in cli package (not imported from generator) to keep packages independent
- [Phase 06]: SCS session manager uses pgxstore for PostgreSQL-backed sessions (no Redis) — AUTH-03 compliance
- [Phase 06]: RequireSession redirects to /auth/login with 302 (HTML UX) vs API auth returning 401 JSON — intentional divergence
- [Phase 06]: LoginUser calls sm.RenewToken before storing credentials to prevent session fixation attacks
- [Phase 06]: BcryptCost=12 (not default 10) and 72-byte length guard in HashPassword for 2026 hardware and bcrypt truncation safety
- [Phase 06]: Use datastar.TemplComponent interface in sse.go (not templ.Component) — avoids templ dependency in generated SSE package while remaining compatible
- [Phase 06]: primitives.templ uses writeRawFile (not writeGoFile) — .templ files need templ compiler, not gofmt
- [Phase 06]: SSE type is datastar.ServerSentEventGenerator (no datastar.SSE alias exists in v1.1.0)
- [Phase 06]: Use pgtestdb.Custom over pgtestdb.New for NewTestPool to avoid open sql.DB connection interfering with pgxpool
- [Phase 06]: Custom atlasMigrator uses Atlas CLI shell-out (not Go library) — consistent with project's existing approach in migrate package
- [Phase 06]: runtime.Caller(0) in DefaultTestDBConfig resolves repo root from internal/forgetest/db.go source file path
- [Phase 06]: scaffold_handlers template uses .Resource.Name for per-resource data access; html_register_all iterates .Resources — distinguishes scaffold-once vs generated-always templates
- [Phase 06]: OAuth temp state in gorilla sessions signed cookie (gothic.Store), completed auth sessions in PostgreSQL SCS+pgxstore (AUTH-03)
- [Phase 06]: RegisterOAuthRoutes places auth routes outside RequireSession middleware to prevent infinite redirect loops
- [Phase 06]: UserFinder/PasswordAuthenticator as function types (not interfaces) so generated app passes closures without forge importing gen/
- [Phase 06]: Sessions table is static block in Atlas HCL template (not conditional) — infrastructure required by pgxstore, not user-defined schema
- [Phase 06]: Role guard uses role == '' || role == 'value' pattern — empty role ensures fields visible in admin/dev contexts
- [Phase 06]: Mutability modifier generates editable-vs-readonly conditional rather than hiding — data visible to all, editable only by matching role
- [Phase 06]: Filter section only rendered when filterableFields helper returns non-empty slice — avoids empty filter UI
- [Phase 06]: DiffResource uses dmp.DiffPrettyText for human-readable output (not patch format) — intended for CLI display
- [Phase 06]: htmlRoutes() returns 7 routes per resource at root path (no /api/v1/ prefix) with GET/POST/PUT/DELETE for Datastar SSE mutations
- [Phase 06]: forge routes displays API and HTML sub-sections per resource with combined total route count
- [Phase 06]: runTemplGenerate tries .forge/bin/templ first then falls back to PATH — forge tool sync preferred but not required
- [Phase 06]: SetupHTML public group intentionally empty — app registers OAuth routes separately; forge cannot import generated auth code
- [Phase 06]: RunTailwindWatch uses cmd.Start (not cmd.Run) and returns *exec.Cmd so caller controls process lifetime for forge dev
- [Phase 06]: ScaffoldTailwindInput uses Tailwind v3 @tailwind directives (not v4 @import) — matches toolsync registry v3.4.17 pin
- [Phase 07-01]: Permission() is a SchemaItem (not an Option) — permission rules kept separate from boolean feature flags like SoftDelete/Auditable
- [Phase 07-01]: PermissionsIR is map[string][]string — operation-keyed for O(1) lookup in templates; supports multiple permission rules per resource
- [Phase 07-01]: Eager is a modifier on RelationshipIR (not a standalone type) — consistent with Optional and OnDelete modifier pattern
- [Phase 07]: No hard DELETE generated for SoftDelete resources — soft delete is final state; raw SQL available if truly needed
- [Phase 07]: Restore method on {{resource}}Actions interface (not DefaultActions only) — contract preserved for custom implementations
- [Phase 07]: ActiveMod prepended to filterMods (not appended) in List — soft delete exclusion established before user-supplied filters

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-18
Stopped at: Completed 07-02-PLAN.md (Soft delete query mods, partial unique indexes, Delete/Restore actions)
Resume file: .planning/phases/07-advanced-data-features/07-03-PLAN.md
