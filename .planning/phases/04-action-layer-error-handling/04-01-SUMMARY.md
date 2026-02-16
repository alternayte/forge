---
phase: 04-action-layer-error-handling
plan: 01
subsystem: error-handling
tags: [errors, database, generation, foundation]
dependency_graph:
  requires: [generator-infrastructure, parser-ir]
  provides: [forge-error-type, db-error-mapping]
  affects: [action-generation, middleware, api-handlers]
tech_stack:
  added: [pgx-error-mapping]
  patterns: [structured-errors, error-wrapping, http-status-codes]
key_files:
  created:
    - internal/generator/templates/errors.go.tmpl
    - internal/generator/templates/errors_db_mapping.go.tmpl
    - internal/generator/errors.go
    - internal/generator/errors_test.go
  modified: []
decisions:
  - choice: "Error struct with 5 fields (Status, Code, Message, Detail, Err)"
    rationale: "Separates HTTP status, machine code, user message, developer detail, and wrapped error for comprehensive error handling"
    alternatives: ["Single error string", "HTTP-only errors"]
  - choice: "Static error templates (not per-resource)"
    rationale: "Error types are application-wide, not resource-specific. Follow validation_types.go.tmpl pattern for shared utilities"
    alternatives: ["Generate per-resource error types"]
  - choice: "NewValidationError accepts interface{} instead of ValidationErrors type"
    rationale: "Avoids circular import between gen/errors and gen/validation. Uses type assertion to call .Error() method"
    alternatives: ["Import validation package", "Generate validation bridge in validation package"]
  - choice: "MapDBError uses errors.As for pgconn.PgError detection"
    rationale: "Type-safe error unwrapping instead of string matching. Compatible with errors.Is/errors.As ecosystem"
    alternatives: ["String matching on error messages", "Direct type assertion"]
metrics:
  duration_seconds: 127
  completed_date: 2026-02-16T21:07:23Z
---

# Phase 04 Plan 01: Error Type and Database Error Mapping Summary

**One-liner:** Generated forge.Error type with HTTP status codes and PostgreSQL error mapping via pgconn.PgError type assertions

## What Was Built

Created the foundational error handling layer for the action system through generator templates and functions:

1. **Error Type Template** (`errors.go.tmpl`):
   - `Error` struct with Status (int), Code (string), Message (string), Detail (string), Err (error)
   - Implements `error` interface via `Error() string` method
   - Implements `Unwrap() error` for errors.Is/errors.As compatibility
   - 7 constructor functions: NotFound (404), UniqueViolation (409), ForeignKeyViolation (400), Unauthorized (401), Forbidden (403), BadRequest (400), InternalError (500)
   - `NewValidationError` function bridges validation errors to forge.Error (422)

2. **Database Error Mapping Template** (`errors_db_mapping.go.tmpl`):
   - 4 PostgreSQL error code constants (23505 unique, 23503 FK, 23502 not null, 23514 check)
   - `MapDBError` function using `errors.As(&pgconn.PgError)` for type-safe error detection
   - Maps DB errors to appropriate forge.Error instances with correct HTTP status codes
   - `IsNotFound` helper for sql.ErrNoRows detection

3. **Generator Function** (`errors.go`):
   - `GenerateErrors` function following validation.go pattern
   - Creates `gen/errors/` directory
   - Renders both templates with nil data (static files)
   - Uses existing `writeGoFile` for formatting and import management

4. **Comprehensive Tests** (`errors_test.go`):
   - End-to-end generation test verifying both files created
   - Content verification for Error struct, constructors, MapDBError, constants
   - File creation test confirming correct directory structure

## How It Works

**Error Generation Flow:**
1. `GenerateErrors` called with resources (unused, kept for consistency)
2. Creates `gen/errors/` directory
3. Renders `errors.go.tmpl` → `gen/errors/errors.go` (static, no template data)
4. Renders `errors_db_mapping.go.tmpl` → `gen/errors/db_mapping.go` (static)
5. Both files formatted via `writeGoFile` with automatic import management

**Runtime Error Mapping:**
```go
// In action layer (future):
err := db.Query(...)
if err != nil {
    return errors.MapDBError(err) // Maps pgx errors to forge.Error
}
```

**Database Error Detection:**
- Uses `errors.As(err, &pgErr)` to unwrap error chain and find `*pgconn.PgError`
- Type-safe, compatible with Go 1.13+ error wrapping
- Extracts constraint names, column names from pgErr for detailed error messages

## Deviations from Plan

None - plan executed exactly as written.

## Key Learnings

1. **Static Template Pattern**: Followed `validation_types.go.tmpl` pattern where templates generate shared utility files (not per-resource). Pass nil as template data since no dynamic content needed.

2. **Import Avoidance**: `NewValidationError` uses `interface{}` parameter instead of importing `gen/validation.ValidationErrors` to prevent circular dependencies. Type assertion to `error` interface allows calling `.Error()` method.

3. **Error Wrapping Best Practices**:
   - `Err` field stores wrapped error for `Unwrap()`
   - `Error()` method formats as "Message: wrapped" when Err present
   - Enables `errors.Is(err, sql.ErrNoRows)` and similar checks downstream

4. **PostgreSQL Error Codes**:
   - 23505: Unique violation (409 Conflict)
   - 23503: FK violation (400 Bad Request)
   - 23502: Not null violation (400 Bad Request)
   - 23514: Check constraint violation (400 Bad Request)

## Testing Evidence

```bash
$ go test ./internal/generator/ -run TestGenerateErrors -v
=== RUN   TestGenerateErrors
--- PASS: TestGenerateErrors (0.02s)
=== RUN   TestGenerateErrors_FileCreation
--- PASS: TestGenerateErrors_FileCreation (0.02s)
PASS
ok  	github.com/forge-framework/forge/internal/generator	0.480s

$ go build ./internal/generator/
Build successful

$ go vet ./internal/generator/
# No issues
```

**Test Coverage:**
- Error struct presence and all 5 fields
- All 8 constructor functions (7 + NewValidationError)
- MapDBError function and pgconn.PgError usage
- All 4 PostgreSQL error code constants
- IsNotFound helper function
- Both files created in correct locations
- DO NOT EDIT headers present

## Next Steps

**Plan 02** (Action Generation):
- Generate CRUD action functions using forge.Error return types
- Wire GenerateErrors into main generator pipeline
- Action functions will call MapDBError for database operations

**Plan 03** (Error Middleware):
- HTTP middleware to convert forge.Error to JSON responses
- Use error.Status for HTTP status code
- Use error.Code for machine-readable error codes in response

## Files Changed

**Created:**
- `internal/generator/templates/errors.go.tmpl` (121 lines) - Error type and constructors
- `internal/generator/templates/errors_db_mapping.go.tmpl` (70 lines) - DB error mapping
- `internal/generator/errors.go` (40 lines) - Generator function
- `internal/generator/errors_test.go` (152 lines) - Comprehensive tests

**Modified:** None

## Commits

- `72b6ad8`: feat(04-01): add error type and DB mapping templates
- `31c9959`: feat(04-01): add GenerateErrors function and tests

## Self-Check: PASSED

All generated files verified:
```bash
$ ls internal/generator/templates/errors*.tmpl
internal/generator/templates/errors.go.tmpl
internal/generator/templates/errors_db_mapping.go.tmpl

$ ls internal/generator/errors*
internal/generator/errors.go
internal/generator/errors_test.go
```

All commits verified:
```bash
$ git log --oneline --all | grep -E "(72b6ad8|31c9959)"
31c9959 feat(04-01): add GenerateErrors function and tests
72b6ad8 feat(04-01): add error type and DB mapping templates
```
