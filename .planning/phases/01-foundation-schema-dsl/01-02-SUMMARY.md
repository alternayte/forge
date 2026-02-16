---
phase: 01-foundation-schema-dsl
plan: 02
subsystem: foundation
tags:
  - ir
  - error-handling
  - cli-ui
  - parser-infrastructure
dependency_graph:
  requires:
    - none
  provides:
    - intermediate-representation-types
    - rust-style-error-diagnostics
    - terminal-ui-styles
  affects:
    - parser (plan 03 - will use IR types and error system)
    - cli (plan 04 - will use UI styles)
tech_stack:
  added:
    - github.com/charmbracelet/lipgloss: terminal styling with graceful degradation
  patterns:
    - Intermediate representation (IR) as domain model (not AST mirror)
    - Fluent builder API for diagnostic construction
    - Rust-style error formatting (file:line:col, source snippet, underline, hint)
key_files:
  created:
    - internal/parser/ir.go: IR type definitions for schema parsing
    - internal/errors/codes.go: error code registry with help links
    - internal/errors/diagnostic.go: diagnostic type with fluent builder
    - internal/errors/formatter.go: Rust-style error rendering
    - internal/errors/errors_test.go: comprehensive error system tests
    - internal/ui/styles.go: Lipgloss-based CLI styling
  modified:
    - go.mod: added Lipgloss dependency
    - go.sum: dependency checksums
decisions:
  - decision: IR uses strings (not enums) for field types and modifiers
    rationale: Decouples IR from schema package, making it consumable by code generation without circular imports
    impact: Code generator will map string types to appropriate Go types
  - decision: Error codes use E0xx/E1xx ranges (parser/validation)
    rationale: Clear categorization, room for expansion within each category
    impact: Help system can group errors by concern
  - decision: Help links use CLI format (forge help E001) not web URLs
    rationale: Offline-first, no web dependency for getting help
    impact: CLI must implement help command for error codes
  - decision: DiagnosticSet collects all errors for single-pass reporting
    rationale: Developer fixes everything at once, no cascading noise (per user requirement)
    impact: Parser must continue after first error to collect all diagnostics
metrics:
  duration: 3 minutes
  tasks_completed: 3
  files_created: 8
  tests_added: 7
  commits: 3
  completed_at: 2026-02-16T15:34:12Z
---

# Phase 1 Plan 2: IR Types, Error Diagnostics, and UI Styles Summary

**One-liner:** Domain-model IR types for parsed schemas, Rust-style error diagnostics with source context and hints, and Lipgloss-based terminal UI styling for consistent CLI output.

## What Was Built

This plan delivered the foundational data structures and formatting infrastructure that the parser (Plan 03) and CLI (Plan 04) depend on:

1. **Intermediate Representation (IR) Types** - Domain model structs representing parsed schema definitions:
   - `ResourceIR` - Top-level resource with fields, relationships, options, and source position
   - `FieldIR` - Parsed field with type, modifiers, enum values
   - `ModifierIR` - Field modifier with type and value
   - `RelationshipIR` - Parsed relationship with type and cascade rules
   - `ResourceOptionsIR` - Resource-level flags (soft delete, auditable, tenant-scoped, searchable)
   - `ParseResult` - Collects multiple resources and all errors for single-pass reporting

2. **Rich Error Diagnostic System** - Rust-style error reporting with source context:
   - Error code registry with E0xx (parser errors) and E1xx (validation errors) ranges
   - `Diagnostic` type with file position, source line, underline range, and hint
   - `DiagnosticBuilder` with fluent API for constructing diagnostics
   - `DiagnosticSet` for collecting multiple errors (satisfies "fix everything at once" requirement)
   - Formatter producing Rust-style output: error[CODE], file:line:col, source snippet, caret underline, hint, help command
   - Lipgloss-based colored output (red errors, blue file positions, cyan hints, dim help)

3. **Terminal UI Styles** - Consistent Lipgloss-based styling for CLI output:
   - Icon constants: SuccessIcon (✓), ErrorIcon (✗), WarnIcon (!), InfoIcon (i)
   - Style objects: HeaderStyle, SuccessStyle, ErrorStyle, WarnStyle, DimStyle, BoldStyle, FilePathStyle, CommandStyle
   - Helper functions: Success(), Error(), Warn(), Info() for formatted messages with icons
   - Grouped() function for Cargo-style grouped output (header + indented items)

## Task Summary

| Task | Name                                  | Commit  | Status   | Files                                      |
| ---- | ------------------------------------- | ------- | -------- | ------------------------------------------ |
| 1    | Define IR types                       | 056dfb1 | Complete | internal/parser/ir.go                      |
| 2    | Build error diagnostic system         | 54a98e7 | Complete | internal/errors/{codes,diagnostic,formatter,errors_test}.go |
| 3    | Create terminal UI styles             | 4331ed6 | Complete | internal/ui/styles.go                      |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed unused variable in NewDiagnostic**
- **Found during:** Task 2
- **Issue:** Variable `info` declared in `NewDiagnostic()` but never used, causing compilation error
- **Fix:** Removed unused variable declaration
- **Files modified:** internal/errors/diagnostic.go
- **Commit:** Included in 54a98e7

No other deviations - plan executed exactly as written.

## Success Criteria

All success criteria met:

- ✓ IR types can represent any schema that Plan 01's DSL can express
- ✓ Error diagnostic system formats errors in Rust-style with file positions, source context, underlines, hints, and error codes
- ✓ UI styles provide a consistent Cargo-like visual language for CLI output
- ✓ All packages compile without errors
- ✓ Error system tests pass (7 tests covering formatting, collection, lookup, fluent API)
- ✓ Diagnostic Format() output matches Rust-style pattern
- ✓ Error codes registered with descriptions and help links
- ✓ UI styles use Lipgloss with icons and color categories

## Testing

**Tests Added:**
- `TestFormat_ProducesExpectedStructure` - Verifies Rust-style output structure
- `TestFormat_HandlesMinimalDiagnostic` - Tests minimal diagnostic rendering
- `TestDiagnosticSet_CollectsMultipleErrors` - Validates multi-error collection
- `TestErrorCodeLookup` - Tests error code registry lookup
- `TestDiagnosticBuilder_FluentAPI` - Validates fluent builder API
- `TestDiagnosticSet_ErrorInterface` - Tests error interface implementation
- `TestDiagnostic_ErrorInterface` - Tests single diagnostic error interface

All tests pass with 100% success rate.

## Technical Notes

**IR Design Decisions:**
- IR is a domain model, not an AST mirror - easier for code generation to consume
- Uses strings for types/modifiers rather than enums - decouples from schema package, prevents circular imports
- Every IR struct includes SourceLine for precise error reporting
- ParseResult.Errors is []error (not []Diagnostic) for flexibility - parser can collect any error type

**Error System Design:**
- Error codes grouped by concern: E0xx for parser, E1xx for validation
- Help links use CLI format (`forge help E001`) for offline-first design
- Underline calculations account for gutter width (line number + pipe character)
- Multi-line hints properly indented for readability
- DiagnosticSet implements error interface so it can be returned directly

**UI Styles Design:**
- All styles defined at package level for consistency
- Helper functions provide standard formatting patterns
- Grouped() enables Cargo-style output: header followed by indented list
- Lipgloss handles terminal capability detection and graceful degradation

## Next Steps

**Immediate (Plan 03 - go/ast parser):**
- Use IR types as target output of parser
- Use DiagnosticSet to collect all parse errors in single pass
- Use error codes (ErrDynamicValue, etc.) when creating diagnostics
- Parse schema.go files into ResourceIR structs

**Downstream Dependencies:**
- Plan 04 (CLI skeleton): Use UI styles for all command output
- Phase 2 (code generation): Consume IR types as input
- All CLI commands: Use error formatter for user-facing errors

## Artifacts

**Key Files:**
- `/Users/nathananderson-tennant/Development/forge-go/internal/parser/ir.go` - IR type definitions (54 lines)
- `/Users/nathananderson-tennant/Development/forge-go/internal/errors/codes.go` - Error code registry (78 lines)
- `/Users/nathananderson-tennant/Development/forge-go/internal/errors/diagnostic.go` - Diagnostic types and builder (109 lines)
- `/Users/nathananderson-tennant/Development/forge-go/internal/errors/formatter.go` - Rust-style formatter (122 lines)
- `/Users/nathananderson-tennant/Development/forge-go/internal/errors/errors_test.go` - Comprehensive tests (150 lines)
- `/Users/nathananderson-tennant/Development/forge-go/internal/ui/styles.go` - Terminal UI styles (124 lines)

**Commits:**
- 056dfb1: feat(01-02): define intermediate representation types
- 54a98e7: feat(01-02): build rich error diagnostic system with Rust-style formatting
- 4331ed6: feat(01-02): create terminal UI styles for CLI output

## Self-Check: PASSED

**Files verified:**
```
FOUND: internal/parser/ir.go
FOUND: internal/errors/codes.go
FOUND: internal/errors/diagnostic.go
FOUND: internal/errors/formatter.go
FOUND: internal/errors/errors_test.go
FOUND: internal/ui/styles.go
```

**Commits verified:**
```
FOUND: 056dfb1
FOUND: 54a98e7
FOUND: 4331ed6
```

All artifacts confirmed present.
