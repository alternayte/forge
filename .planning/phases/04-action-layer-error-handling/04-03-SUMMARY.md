---
phase: 04-action-layer-error-handling
plan: 03
type: execute
subsystem: generator
tags: [middleware, error-handling, panic-recovery, sse, orchestrator]
dependency_graph:
  requires: [errors-types, actions-layer]
  provides: [middleware-recovery, middleware-errors, complete-orchestrator]
  affects: [generator-orchestrator]
tech_stack:
  added: []
  patterns: [panic-recovery, context-aware-error-rendering, sse-fragments]
key_files:
  created:
    - internal/generator/templates/middleware_recovery.go.tmpl
    - internal/generator/templates/middleware_errors.go.tmpl
    - internal/generator/middleware.go
    - internal/generator/middleware_test.go
  modified:
    - internal/generator/generator.go
decisions:
  - decision: "SSE errors use HTTP 200 with datastar-merge-fragments event for toast notifications"
    rationale: "SSE protocol requires 200 status; error details conveyed in event data"
    alternatives: ["Use error event type", "Return non-200 status"]
    trade_offs: "200 status is less RESTful but required for SSE streams to remain open"
  - decision: "JSON errors follow RFC 9457 shape using simple fmt.Fprintf"
    rationale: "Provides Huma-compatible structure without json.Marshal dependency"
    alternatives: ["Use json.Marshal", "Custom JSON serialization"]
    trade_offs: "Simple string formatting limits complex error structures but adequate for Phase 4"
  - decision: "HTML errors use inline template string instead of templ"
    rationale: "Minimal fallback for direct browser access; templ templates come in Phase 6"
    alternatives: ["Use templ now", "Return plain text"]
    trade_offs: "Simple HTML is less maintainable but avoids premature templ integration"
  - decision: "Import stdlib errors as stderrors to avoid conflict with gen/errors"
    rationale: "Cleaner to use unqualified 'errors' for forge errors since most references are to forge types"
    alternatives: ["Import gen/errors as forgeerrors", "Use fully qualified paths"]
    trade_offs: "Stdlib alias is slightly less intuitive but reads better in forge-heavy code"
metrics:
  duration_minutes: 2.1
  tasks_completed: 2
  files_created: 4
  files_modified: 1
  tests_added: 1
  commit_count: 2
  completed_date: 2026-02-16
---

# Phase 04 Plan 03: Middleware Generation and Orchestrator Summary

**One-liner:** Panic recovery middleware with slog stack traces, context-aware error rendering (SSE/JSON/HTML), and 11-generator orchestrator

## Overview

Completed the Phase 4 error handling story by generating panic recovery middleware and error rendering utilities, then wiring all Phase 4 generators (errors, actions, middleware) into the forge generate orchestrator. The middleware prevents panics from crashing the server and renders errors appropriately based on request context (SSE for Datastar, JSON for APIs, HTML for browsers).

## Tasks Completed

### Task 1: Create middleware templates for panic recovery and error rendering
**Status:** Complete
**Commit:** e4e121a

Created two Go templates that generate `gen/middleware/recovery.go` and `gen/middleware/errors.go`:

**middleware_recovery.go.tmpl:**
- Recovery middleware wraps handlers with defer/recover
- Logs panics with full stack trace via slog.Error (error, stack, method, path)
- Creates generic 500 error with "internal_error" code and "An unexpected error occurred" message
- Never exposes panic details to clients (security requirement)
- Calls ErrorResponder to render error appropriately

**middleware_errors.go.tmpl:**
- ErrorResponder inspects Accept header and path to determine response format
- Extracts forge.Error from error or wraps with InternalError if not a forge error
- Routes to writeSSEError, writeJSONError, or writeHTMLError based on context

**writeSSEError:**
- Returns HTTP 200 with text/event-stream content type (SSE protocol requirement)
- Sends datastar-merge-fragments event with toast div containing error message
- Format: `event: datastar-merge-fragments\ndata: fragments <div id="toast-container">...</div>\n\n`

**writeJSONError:**
- Returns JSON with RFC 9457 problem details shape: status, title, detail, code
- Uses fmt.Fprintf for simple serialization (no json.Marshal dependency)
- Huma will enhance this in Phase 5 with full RFC 9457 compliance

**writeHTMLError:**
- Returns minimal HTML error page with status code and message
- Inline CSS for basic styling (no templ yet - comes in Phase 6)
- Fallback for direct browser access outside Datastar/API contexts

### Task 2: Create middleware generator, tests, and wire all Phase 4 generators into orchestrator
**Status:** Complete
**Commit:** 4a26dbb

**middleware.go:**
- GenerateMiddleware function creates gen/middleware/ directory
- Renders both templates with project module for imports
- Writes recovery.go and errors.go using standard writeGoFile pipeline

**middleware_test.go:**
- TestGenerateMiddleware validates both generated files
- Checks for required elements: Recovery function, recover(), debug.Stack(), slog.Error, ErrorResponder
- Verifies error rendering helpers: writeSSEError, writeJSONError, writeHTMLError
- Confirms Datastar event format and content type headers

**generator.go updates:**
- Added three new generator calls after existing 8 generators:
  - GenerateErrors (9th) - error types and DB error mapping
  - GenerateActions (10th) - action interfaces and default implementations
  - GenerateMiddleware (11th) - panic recovery and error rendering
- Each with standard error check and return pattern
- Orchestrator now produces all Phase 4 artifacts alongside existing ones

## Verification Results

All verification criteria met:

1. All generator tests pass (40+ tests) - PASSED
2. go build ./internal/generator/ compiles cleanly - PASSED
3. go vet ./internal/generator/ reports no issues - PASSED
4. Generated recovery.go has Recovery middleware with defer/recover and slog stack trace logging - VERIFIED
5. Generated errors.go has ErrorResponder with SSE/JSON/HTML rendering based on Accept header - VERIFIED
6. SSE errors use 200 status with datastar-merge-fragments event format - VERIFIED
7. JSON errors follow RFC 9457 shape (status, title, detail, code) - VERIFIED
8. Generate() orchestrator calls all 11 generators in order - VERIFIED
9. Panic recovery does NOT expose panic details in response (Message is generic, Detail is empty) - VERIFIED

## Success Criteria

- [x] Panic recovery middleware catches panics, logs full stack trace via slog, returns generic 500
- [x] SSE context (Accept: text/event-stream) gets 200 status with toast fragment via Datastar event
- [x] JSON context (Accept: application/json or /api/ path) gets RFC 9457-shaped error response
- [x] HTML context gets minimal error page with status code and message
- [x] Panic details never exposed to client (security requirement)
- [x] Generate() orchestrator produces all Phase 4 artifacts alongside existing ones
- [x] All 11 generators called in correct order
- [x] All existing tests continue to pass (no regressions)

## Key Technical Decisions

1. **SSE errors use HTTP 200 status:** SSE protocol requires 200 status for streams to remain open. Error details conveyed in event data using datastar-merge-fragments event with toast div.

2. **Simple fmt.Fprintf for JSON errors:** Provides RFC 9457-compatible shape without json.Marshal dependency. Adequate for Phase 4; Huma will enhance in Phase 5.

3. **Inline HTML template:** Minimal fallback for direct browser access. Avoids premature templ integration (comes in Phase 6).

4. **Import stdlib errors as stderrors:** Avoids conflict with gen/errors package. Unqualified 'errors' refers to forge errors since most references are to forge types.

5. **11-generator orchestrator order:** Phase 4 generators (errors, actions, middleware) called after existing 8 to ensure dependencies exist (models, validation, queries).

## Deviations from Plan

None - plan executed exactly as written.

## Files Generated

When forge generate runs, produces:
- `gen/middleware/recovery.go` - Recovery middleware with panic handling
- `gen/middleware/errors.go` - ErrorResponder and context-aware error rendering helpers

These files integrate with existing Phase 4 outputs:
- `gen/errors/errors.go` - Error type and constructors
- `gen/errors/db_mapping.go` - MapDBError function
- `gen/actions/<resource>.go` - Per-resource Actions interfaces and DefaultActions implementations
- `gen/actions/types.go` - Shared DB and Registry types

## Integration Points

**Recovery middleware depends on:**
- `gen/errors` package for Error type
- stdlib slog for structured logging
- stdlib runtime/debug for stack traces

**ErrorResponder depends on:**
- `gen/errors` package for Error type
- stdlib errors for As() type assertions
- HTTP Accept header and URL path for context detection

**Generate() orchestrator:**
- Calls 11 generators in dependency order
- Models → Atlas → Factories → Validation → Queries → Pagination → Transaction → SQLCConfig → Errors → Actions → Middleware
- Each generator follows standard pattern: ensureDir, renderTemplate, writeGoFile

## Next Steps

Phase 4 is now complete. All error handling infrastructure is in place:
- Error types with constructors (ERR-01)
- Database error mapping (ERR-05)
- Action layer interfaces and implementations (ACT-01, ACT-02)
- Panic recovery middleware (ERR-04)
- Context-aware error rendering (ERR-02, ERR-03)

Phase 5 will integrate Huma for OpenAPI and API handlers, leveraging the forge.Error type for RFC 9457 problem details. Phase 6 will add templ templates for HTML rendering, replacing the inline HTML in writeHTMLError.

## Performance Notes

- Duration: 2.1 minutes (faster than phase average of 2.5m)
- 2 tasks, 4 files created, 1 file modified
- 2 atomic commits with clear scope separation
- All tests pass with zero regressions

## Self-Check: PASSED

**Created files:**
```
FOUND: internal/generator/templates/middleware_recovery.go.tmpl
FOUND: internal/generator/templates/middleware_errors.go.tmpl
FOUND: internal/generator/middleware.go
FOUND: internal/generator/middleware_test.go
```

**Modified files:**
```
FOUND: internal/generator/generator.go
```

**Commits:**
```
FOUND: e4e121a
FOUND: 4a26dbb
```

All files exist and all commits are in git history.
