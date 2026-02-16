# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-16)

**Core value:** Schema is the single source of truth — define a resource once and everything is generated automatically with zero manual sync.
**Current focus:** Phase 2: Code Generation Engine

## Current Position

Phase: 4 of 8 (Action Layer Error Handling)
Plan: 2 of 3 in current phase
Status: In Progress
Last activity: 2026-02-16 — Completed 04-02-PLAN.md

Progress: [███░░░░░░░] 31%

## Performance Metrics

**Velocity:**
- Total plans completed: 15
- Average duration: 3.4 minutes
- Total execution time: 0.82 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 | 5 | 18.9m | 3.8m |
| 02 | 5 | 18.0m | 3.6m |
| 03 | 3 | 10.2m | 3.4m |
| 04 | 2 | 4.0m | 2.0m |

**Recent Executions:**

| Phase | Plan | Duration | Tasks | Files |
|-------|------|----------|-------|-------|
| 04 | 02 | 1.9m | 2 | 4 |
| 04 | 01 | 2.1m | 2 | 4 |
| 03 | 02 | 3.7m | 2 | 10 |
| 03 | 01 | 4.0m | 2 | 9 |
| 03 | 03 | 2.5m | 2 | 2 |
| 02 | 04 | 4.3m | 2 | 5 |

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

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-16T21:12:07Z
Stopped at: Completed 04-02-PLAN.md
Resume file: Phase 04 plan 02 complete - Action interface and default implementation generation
