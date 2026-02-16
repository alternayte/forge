# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-16)

**Core value:** Schema is the single source of truth — define a resource once and everything is generated automatically with zero manual sync.
**Current focus:** Phase 2: Code Generation Engine

## Current Position

Phase: 2 of 8 (Code Generation Engine)
Plan: 5 of 5 in current phase
Status: In Progress
Last activity: 2026-02-16 — Completed 02-05-PLAN.md

Progress: [██░░░░░░░░] 20%

## Performance Metrics

**Velocity:**
- Total plans completed: 9
- Average duration: 3.6 minutes
- Total execution time: 0.54 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 | 5 | 18.9m | 3.8m |
| 02 | 4 | 13.7m | 3.4m |

**Recent Executions:**

| Phase | Plan | Duration | Tasks | Files |
|-------|------|----------|-------|-------|
| 01 | 05 | 4.0m | 2 | 7 |
| 02 | 01 | 3.2m | 2 | 8 |
| 02 | 02 | 4.5m | 2 | 10 |
| 02 | 03 | 2.9m | 1 | 2 |
| 02 | 05 | 3.1m | 2 | 7 |

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
- [Phase 02-05]: 300ms debounce for file changes to batch multi-file saves
- [Phase 02-05]: Skip Chmod events unconditionally (always noise from Spotlight/antivirus/editors)
- [Phase 02-05]: Watch parent directories not individual files (handles atomic writes correctly)
- [Phase 02-05]: Exclude gen/ paths from triggering regeneration (prevent infinite loops)

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-02-16T18:40:35Z
Stopped at: Completed 02-05-PLAN.md
Resume file: Phase 02 plan 05 complete - forge dev with file watching
