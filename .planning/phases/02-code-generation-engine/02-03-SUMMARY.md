---
phase: 02-code-generation-engine
plan: 03
subsystem: code-generation-cli
status: complete
completed: 2026-02-16T17:32:33Z
duration_minutes: 2.9
tags: [cli, integration, pipeline, diagnostics]

dependency_graph:
  requires: [02-01]
  provides: [forge-generate-command]
  affects: [cli, parser, generator]

tech_stack:
  added: []
  patterns: [cobra-command, error-formatting, directory-cleanup]

key_files:
  created:
    - internal/cli/generate.go
  modified:
    - internal/cli/root.go

decisions:
  - title: Only display generated directories that exist
    rationale: Factories and Atlas generators not yet implemented, showing "0 files" is misleading
    impact: Clean output that only shows what was actually generated
  - title: Clean gen/ directory before generation
    rationale: Ensures stale files from removed resources don't persist
    impact: Idempotent generation - gen/ always reflects current resources/

metrics:
  tasks_completed: 1
  files_created: 1
  files_modified: 1
  tests_added: 0
  commits: 1
---

# Phase 02 Plan 03: CLI Generate Command Summary

**One-liner:** Implemented forge generate CLI command that orchestrates parser-generator pipeline with rich error diagnostics and styled output.

## Completed Tasks

| Task | Description | Commit | Key Changes |
|------|-------------|--------|-------------|
| 1 | Create forge generate command with parser-generator pipeline | 9a9dcd3 | Created generate.go, registered in root CLI |

## Implementation Details

### forge generate Command

Created the main user-facing command that ties together the entire code generation pipeline:

**Pipeline stages:**
1. **Project discovery** - Finds forge.toml by walking up directories
2. **Config loading** - Loads forge.toml and validates module path
3. **Schema parsing** - Calls parser.ParseDir on resources/ directory
4. **Error handling** - Displays rich diagnostics for parse errors (file:line:col with context)
5. **Directory cleanup** - Removes gen/ before generation to clear stale files
6. **Code generation** - Calls generator.Generate with parsed ResourceIR
7. **Output summary** - Shows success with file counts and timing

**Error handling patterns:**
- Diagnostic errors formatted with errors.Format() for rich output
- Non-diagnostic errors shown with ui.Error()
- All errors collected and displayed together (not fail-fast)
- Empty resources/ shows info message, not error

**Output features:**
- Styled terminal output using ui package (lipgloss)
- Only shows directories that exist (no "0 files" noise)
- Timing information in milliseconds
- Resource count in summary

**GEN-11 Protection:**
The command only reads from resources/ (via parser.ParseDir) and writes to gen/. It never modifies anything in resources/. This is enforced by:
- Parser is read-only
- Generator outputDir is always gen/
- Clean operation only touches gen/

### Integration with Existing Systems

**Parser integration:**
- Uses parser.ParseDir to discover and parse all resources/*/schema.go files
- Handles ParseResult.Errors collection for batch error reporting
- Checks for empty Resources array

**Generator integration:**
- Passes parsed ResourceIR to generator.Generate
- Provides GenerateConfig with outputDir and ProjectModule
- Handles generation errors

**UI integration:**
- Uses ui.Success(), ui.Info(), ui.Error() for consistent formatting
- Uses ui.Header() for section headers
- Uses ui.DimStyle for secondary information

## Verification

All verification criteria met:

1. `go build .` compiles forge binary with generate command ✓
2. `forge generate --help` shows correct usage text ✓
3. `forge generate` in initialized project produces gen/models/ ✓
4. `forge generate` with schema errors shows all errors with file:line positions ✓
5. `forge generate` does not touch resources/ directory ✓
6. Output styled with icons and timing information ✓

**Integration testing:**
- Created test project with forge init
- Ran forge generate successfully
- Verified gen/models/ created with correct files
- Verified resources/ untouched
- Tested error case with dynamic value - rich diagnostics displayed
- Tested empty resources/ - info message shown

## Deviations from Plan

None - plan executed exactly as written.

## Potential Issues

None identified. Command works as specified with proper error handling and output formatting.

## Next Steps

Plan complete. This command is ready for Phase 2 Plan 4 (Atlas schema generation) and Plan 5 (Factory generation), which will add more output sections.

## Self-Check: PASSED

**Created files:**
- internal/cli/generate.go ✓

**Modified files:**
- internal/cli/root.go ✓

**Commits:**
- 9a9dcd3 ✓
