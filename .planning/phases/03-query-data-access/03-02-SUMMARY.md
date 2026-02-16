---
phase: 03-query-data-access
plan: 02
subsystem: code-generation
tags: [pagination, sqlc, transactions, river, bob, cursor-pagination]

dependency_graph:
  requires:
    - 03-01 (validation and query builder generation)
    - 02-01 (model generation for type infrastructure)
  provides:
    - Pagination utilities for cursor and offset-based pagination
    - SQLC configuration for custom SQL escape hatch
    - Transaction wrappers with River job support
  affects:
    - Future API implementation (will use cursor pagination)
    - Future action layer (will use transaction wrappers)
    - Developer workflow (can write custom SQL via SQLC)

tech_stack:
  added:
    - Pagination utility generation (cursor + offset)
    - SQLC configuration generation (v2 format)
    - Transaction wrapper generation (pgx + River)
  patterns:
    - Base64 cursor encoding with JSON for opaque tokens
    - Row value comparison for cursor pagination (PostgreSQL tuple syntax)
    - Page size capping at 100 for safety
    - BeginFunc pattern for automatic commit/rollback
    - InsertTx for transactional job enqueueing

key_files:
  created:
    - internal/generator/pagination.go
    - internal/generator/pagination_test.go
    - internal/generator/templates/pagination.go.tmpl
    - internal/generator/sqlc.go
    - internal/generator/sqlc_test.go
    - internal/generator/templates/sqlc.yaml.tmpl
    - internal/generator/templates/transaction.go.tmpl
  modified:
    - internal/generator/generator.go (added ProjectRoot, wired 3 new generators)
    - internal/cli/generate.go (added ProjectRoot to config)
    - internal/watcher/dev.go (added ProjectRoot to config)

decisions:
  - Pagination utilities generated once (not per-resource) since logic is generic
  - Cursor includes ID + sort field + sort value for uniqueness and stability
  - Base64 URL encoding for cursor tokens (standard, debuggable, no encryption)
  - Page size capped at 100 to prevent accidental large queries
  - SQLC config placed at project root (standard location for sqlc.yaml)
  - Transaction wrapper in gen/forge/ package (runtime utilities for user code)
  - ProjectRoot added to GenerateConfig for files outside gen/ directory
  - All 8 generators called from Generate() orchestrator in logical order

metrics:
  duration: 3.73
  tasks_completed: 2
  files_created: 7
  files_modified: 3
  tests_added: 2
  completed_at: 2026-02-16T20:32:07Z
---

# Phase 03 Plan 02: Pagination, SQLC Config, and Transaction Generation Summary

**One-liner:** Generic pagination utilities with cursor/offset support, SQLC escape hatch configuration, and pgx+River transaction wrappers

## What Was Built

Generated three core infrastructure subsystems:

1. **Pagination Generator** - Produces single `gen/queries/pagination.go` file with:
   - `PageInfo` struct with hasNext/hasPrev, cursors, and optional totalCount
   - `EncodeCursor`/`DecodeCursor` for base64(JSON) opaque tokens
   - `OffsetPaginationMods(page, pageSize)` with page size capping at 100
   - `CursorPaginationMods(cursor, pageSize, sortDir)` with row value comparison using `(sort_field, id) > (?, ?)`
   - Cursor struct includes ID + sort field + sort value for stable ordering

2. **SQLC Config Generator** - Produces `sqlc.yaml` at project root with:
   - SQLC v2 configuration format
   - Points to `queries/custom/` for developer-written SQL files
   - Uses `gen/atlas/` as schema source
   - Configures pgx/v5 driver with JSON tags and pointer result structs
   - Provides escape hatch for complex queries Bob can't express

3. **Transaction Wrapper Generator** - Produces `gen/forge/transaction.go` with:
   - `DB` interface compatible with pgx transaction types
   - `Transaction(ctx, pool, fn)` wrapping pgx.BeginFunc
   - `TransactionWithJobs(ctx, pool, client, fn)` for atomic job enqueueing with River
   - Clean API surface with automatic commit/rollback handling

All three generators wired into `Generate()` orchestrator. Added `ProjectRoot` field to `GenerateConfig` for files that live outside gen/.

## Task Breakdown

### Task 1: Create pagination utility generation and SQLC config generation
**Status:** Complete
**Commit:** dc93e54
**Files:** pagination.go, pagination_test.go, pagination.go.tmpl, sqlc.go, sqlc_test.go, sqlc.yaml.tmpl
**Duration:** ~2 minutes

Created pagination and SQLC generators with:
- Pagination template generates cursor and offset utilities in single file
- Cursor encoding uses base64(JSON) with ID + sort field + sort value
- Offset pagination validates and caps page size at 100
- Cursor pagination uses PostgreSQL row value comparison `(col, id) > (?, ?)`
- SQLC config template generates static YAML pointing to queries/custom/
- Tests verify generated code contains expected functions and compiles

### Task 2: Create transaction wrapper generation and wire all new generators
**Status:** Complete
**Commit:** f68bcdf
**Files:** transaction.go.tmpl, sqlc.go (updated), sqlc_test.go (updated), generator.go, generate.go, dev.go
**Duration:** ~1.5 minutes

Created transaction generator and orchestrator wiring:
- Transaction template generates DB interface and wrapper functions
- Transaction() uses pgx.BeginFunc for clean commit/rollback
- TransactionWithJobs() passes River client for InsertTx calls
- Added ProjectRoot to GenerateConfig for sqlc.yaml placement
- Wired 8 generators in Generate(): Models, Atlas, Factories, Validation, Queries, Pagination, Transaction, SQLCConfig
- Updated CLI generate and dev watcher to pass ProjectRoot

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

All verification steps passed:

1. ✅ `go test ./internal/generator/ -v` - All tests pass (20 test suites including new pagination/sqlc/transaction)
2. ✅ `go build ./internal/generator/` - Clean compile
3. ✅ `go vet ./internal/generator/` - No issues
4. ✅ Generated pagination.go contains EncodeCursor, DecodeCursor, OffsetPaginationMods, CursorPaginationMods
5. ✅ Generated sqlc.yaml points to queries/custom/ with pgx/v5 driver
6. ✅ Generated transaction.go contains Transaction and TransactionWithJobs with pgx.BeginFunc
7. ✅ Generate() calls all 8 generators in order

## Technical Decisions

**Generic pagination file:** Pagination logic is not resource-specific, so generates single file `gen/queries/pagination.go` rather than per-resource files. Reduces duplication and maintains DRY principle.

**Cursor structure:** Cursor includes three components (ID, sort field, sort value) to ensure stable, unique ordering. ID alone is unique but doesn't respect sort order; sort field alone may not be unique. Combination guarantees correctness.

**Base64 URL encoding:** Used `base64.URLEncoding` instead of `StdEncoding` for URL-safe cursors. JSON serialization provides debuggability (cursors can be decoded to inspect position). No encryption/signing needed initially - can add HMAC later if manipulation becomes concern.

**Page size capping:** Hard cap at 100 records per page prevents accidental large queries that could overwhelm database or client. Users can manually batch if needed, but default is safe.

**Row value comparison:** PostgreSQL tuple comparison `(col, id) > (?, ?)` is efficient and handles multi-column sorts correctly. More efficient than separate WHERE clauses with OR logic.

**SQLC at project root:** Follows SQLC convention of placing `sqlc.yaml` at project root. Required adding ProjectRoot to GenerateConfig since OutputDir points to gen/.

**Transaction in gen/forge/:** Runtime utilities for user code belong in gen/forge/ package, not gen/queries/. Separates query-time utilities (pagination, filters) from transaction-time utilities.

**Generator ordering:** Models → Atlas → Factories → Validation → Queries → Pagination → Transaction → SQLCConfig. Logical dependency order, though most are independent. SQLCConfig last since it's a one-time setup file.

## Next Steps

Phase 03 Plan 03 will implement database CLI commands (create, drop, seed, console, reset) that consume the transaction wrappers and migrations from previous plans.

## Self-Check

Verifying created files exist:

```bash
[ -f "internal/generator/pagination.go" ] && echo "FOUND: internal/generator/pagination.go" || echo "MISSING: internal/generator/pagination.go"
[ -f "internal/generator/pagination_test.go" ] && echo "FOUND: internal/generator/pagination_test.go" || echo "MISSING: internal/generator/pagination_test.go"
[ -f "internal/generator/templates/pagination.go.tmpl" ] && echo "FOUND: internal/generator/templates/pagination.go.tmpl" || echo "MISSING: internal/generator/templates/pagination.go.tmpl"
[ -f "internal/generator/sqlc.go" ] && echo "FOUND: internal/generator/sqlc.go" || echo "MISSING: internal/generator/sqlc.go"
[ -f "internal/generator/sqlc_test.go" ] && echo "FOUND: internal/generator/sqlc_test.go" || echo "MISSING: internal/generator/sqlc_test.go"
[ -f "internal/generator/templates/sqlc.yaml.tmpl" ] && echo "FOUND: internal/generator/templates/sqlc.yaml.tmpl" || echo "MISSING: internal/generator/templates/sqlc.yaml.tmpl"
[ -f "internal/generator/templates/transaction.go.tmpl" ] && echo "FOUND: internal/generator/templates/transaction.go.tmpl" || echo "MISSING: internal/generator/templates/transaction.go.tmpl"
```

Verifying commits exist:

```bash
git log --oneline --all | grep -q "dc93e54" && echo "FOUND: dc93e54" || echo "MISSING: dc93e54"
git log --oneline --all | grep -q "f68bcdf" && echo "FOUND: f68bcdf" || echo "MISSING: f68bcdf"
```

**Results:**

All files verified: ✅
- FOUND: internal/generator/pagination.go
- FOUND: internal/generator/pagination_test.go
- FOUND: internal/generator/templates/pagination.go.tmpl
- FOUND: internal/generator/sqlc.go
- FOUND: internal/generator/sqlc_test.go
- FOUND: internal/generator/templates/sqlc.yaml.tmpl
- FOUND: internal/generator/templates/transaction.go.tmpl

All commits verified: ✅
- FOUND: dc93e54 (Task 1: pagination and SQLC config)
- FOUND: f68bcdf (Task 2: transaction wrapper and orchestrator wiring)

## Self-Check: PASSED
