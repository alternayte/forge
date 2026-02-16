---
phase: 02-code-generation-engine
plan: 05
subsystem: development-workflow
status: complete
completed: 2026-02-16T18:40:35Z
duration_minutes: 3.1
tags: [cli, file-watching, hot-reload, fsnotify]

dependency_graph:
  requires: [02-03]
  provides: [forge-dev-command, file-watching, auto-regeneration]
  affects: [cli, watcher, generator, parser]

tech_stack:
  added: [github.com/fsnotify/fsnotify@v1.9.0]
  patterns: [fsnotify-wrapper, debouncing, signal-handling, recursive-directory-watching]

key_files:
  created:
    - internal/watcher/watcher.go
    - internal/watcher/dev.go
    - internal/watcher/watcher_test.go
    - internal/cli/dev.go
  modified:
    - internal/cli/root.go
    - go.mod
    - go.sum

decisions:
  - title: 300ms debounce for file changes
    rationale: Prevents rapid-fire regeneration during multi-file saves; research recommended 300ms
    impact: Smoother developer experience with grouped file saves
  - title: Skip Chmod events unconditionally
    rationale: Chmod events are noisy from Spotlight, antivirus, editors; never useful for rebuild triggers
    impact: Eliminates infinite rebuild loops and spurious regeneration
  - title: Watch parent directories not individual files
    rationale: Editors use atomic writes (temp + rename) which breaks individual file watches
    impact: Reliable file watching that survives editor save operations
  - title: Exclude gen/ paths from triggering regeneration
    rationale: Generated files should not trigger regeneration (infinite loop prevention)
    impact: Safe regeneration cycle without recursive triggering

metrics:
  tasks_completed: 2
  files_created: 4
  files_modified: 3
  tests_added: 8
  commits: 2
---

# Phase 02 Plan 05: Development Server with File Watching Summary

**One-liner:** Implemented forge dev command with fsnotify-based file watching, 300ms debouncing, and automatic code regeneration on schema changes.

## Completed Tasks

| Task | Description | Commit | Key Changes |
|------|-------------|--------|-------------|
| 1 | Create watcher package with fsnotify, debouncing, and dev server orchestration | a77130d | Created watcher.go, dev.go, watcher_test.go; added fsnotify dependency |
| 2 | Wire forge dev CLI command | c04c36d | Created dev.go CLI command, registered in root.go |

## Implementation Details

### Watcher Package (internal/watcher/)

**watcher.go - fsnotify wrapper:**
- Wraps fsnotify.Watcher with debouncing logic (300ms default)
- Unconditionally skips fsnotify.Chmod events (research pitfall #6)
- Filters events by file extension: .go, .templ, .sql, .css
- Excludes gen/ paths to prevent infinite regeneration loops
- Graceful shutdown via done channel
- Handles fsnotify errors with logging (non-fatal)

**Key function: isRelevantFile()**
- Returns true for .go, .templ, .sql, .css files
- Returns false for temp files (.swp, .tmp, ~)
- Returns false for any path containing "gen/"
- Simple extension-based filtering for clarity

**dev.go - DevServer orchestration:**
- Runs initial `forge generate` on startup
- Creates watcher with onChange callback
- Recursively watches resources/ and internal/ directories
- Skips hidden directories (starting with ".")
- Skips gen/, node_modules/
- onChange callback: timestamp + parse + generate cycle
- Parse errors logged but don't crash watcher
- Context-based graceful shutdown
- Styled terminal output with ui package

**Development workflow:**
1. Run initial generation to ensure gen/ up to date
2. Start watcher on resources/ and internal/
3. Print startup message with watched directories
4. Block on ctx.Done() (Ctrl+C triggers cancellation)
5. On file change: debounce → parse → generate → print status
6. On shutdown: close watcher, print cleanup message

**Watch strategy:**
- Watch parent directories (not individual files) per fsnotify FAQ
- Handles atomic writes correctly (write temp → rename)
- Recursive directory walking with filepath.WalkDir
- Each subdirectory added separately to watcher

### forge dev CLI Command (internal/cli/dev.go)

**Command structure:**
- Use: "dev"
- Short: "Start development server with file watching"
- Long: Multi-line description mentioning file watching and Ctrl+C
- RunE: runDev implementation

**runDev() implementation:**
1. Find project root (walk up to forge.toml)
2. Load config from forge.toml
3. Create context with signal.NotifyContext for SIGINT/SIGTERM
4. Create DevServer with project root and config
5. Call server.Start(ctx) - blocks until Ctrl+C
6. Graceful shutdown on context cancellation

**Signal handling pattern:**
```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()
```

This ensures Ctrl+C triggers context cancellation, allowing DevServer to clean up before exit.

### Tests (internal/watcher/watcher_test.go)

Comprehensive coverage of file filtering logic:
- TestIsRelevantFile_GoFiles: .go files are relevant
- TestIsRelevantFile_TemplFiles: .templ files are relevant
- TestIsRelevantFile_SQLFiles: .sql files are relevant
- TestIsRelevantFile_CSSFiles: .css files are relevant
- TestIsRelevantFile_IgnoredExtensions: temp files ignored
- TestIsRelevantFile_IgnoreGenDir: gen/ paths ignored
- TestNewWatcher_CreatesSuccessfully: Constructor works
- TestWatcher_Close: Cleanup works

**Note:** Debounce timing not tested (timing-dependent, hard to test reliably in unit tests per plan guidance).

## Verification

All verification criteria met:

1. `go build .` compiles forge binary with dev command ✓
2. `go test ./internal/watcher/ -v` all tests pass ✓
3. `forge dev --help` shows file watching description ✓
4. forge dev watches resources/ and regenerates on schema changes ✓ (implementation verified)
5. Chmod events are ignored ✓ (event.Has(fsnotify.Chmod) check)
6. Ctrl+C gracefully stops the watcher ✓ (signal.NotifyContext handling)
7. gen/ directory paths excluded from triggering regeneration ✓ (isRelevantFile check)

**Build verification:**
```bash
go build .                        # Success
go test ./internal/watcher/ -v   # All 8 tests pass
go vet ./internal/watcher/        # No issues
go vet ./internal/cli/            # No issues
forge dev --help                  # Shows correct description
```

## Deviations from Plan

None - plan executed exactly as written.

## Integration with Existing Systems

**Parser integration:**
- DevServer calls parser.ParseDir on resources/
- Handles ParseResult.Errors collection
- Logs parse errors without crashing watcher

**Generator integration:**
- DevServer calls generator.Generate with parsed ResourceIR
- Passes GenerateConfig with outputDir and ProjectModule
- Handles generation errors gracefully

**UI integration:**
- Uses ui.Info(), ui.Success(), ui.Error() for status messages
- Uses ui.Header() for section headers
- Uses ui.FilePathStyle and ui.DimStyle for terminal output
- Follows Cargo-style grouped output patterns

**Config integration:**
- Loads config.Config from forge.toml
- Uses Config.Project.Module for generator configuration

## Technical Decisions

### Why 300ms debounce?
Research document recommends 300ms as sweet spot:
- Short enough for responsive feedback
- Long enough to batch multi-file saves
- Prevents rapid-fire regeneration when editor saves multiple files

### Why skip ALL Chmod events?
Research pitfall #6: Chmod events are always noise, never signal.
- macOS Spotlight triggers Chmod constantly
- Antivirus software modifies permissions
- Backup software touches files
- No legitimate use case for Chmod triggering rebuild

### Why watch directories not files?
Research pattern #4 and fsnotify FAQ:
- Editors use atomic writes (write temp → rename old → rename temp)
- Watching individual files breaks on atomic writes (old inode gone)
- Watching parent directory catches all file events reliably

### Why exclude gen/ paths?
Generated files should not trigger regeneration:
- gen/models/*.go are outputs, not inputs
- Regenerating on gen/ changes would create infinite loop
- Simple path check prevents this issue entirely

## Potential Issues

None identified. Implementation follows research recommendations exactly, avoiding all documented pitfalls.

## Next Steps

Plan complete. forge dev is ready for use. Future phases may enhance with:
- Hot reload of running Go server (recompile + restart binary)
- Live browser reload for HTML changes
- Build error notifications
- Migration auto-apply on schema changes

For now, forge dev focuses on schema watching + code regeneration, which is the core Phase 2 workflow.

## Self-Check: PASSED

**Created files exist:**
- internal/watcher/watcher.go ✓
- internal/watcher/dev.go ✓
- internal/watcher/watcher_test.go ✓
- internal/cli/dev.go ✓

**Modified files exist:**
- internal/cli/root.go ✓
- go.mod ✓
- go.sum ✓

**Commits exist:**
- a77130d (watcher package) ✓
- c04c36d (CLI command) ✓

**Build verification:**
```bash
$ go build .
# Success

$ go test ./internal/watcher/ -v
PASS
ok      github.com/forge-framework/forge/internal/watcher    0.368s

$ ./forge dev --help
Starts a development server that watches for file changes in resources/
and internal/ directories. When schema files change, code is automatically
regenerated. Watches .go, .templ, .sql, and .css files.
```
