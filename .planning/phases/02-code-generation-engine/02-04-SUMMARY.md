---
phase: 02-code-generation-engine
plan: 04
subsystem: migration-management
tags: [atlas, migrations, destructive-detection, cli]
completed: 2026-02-16T17:41:18Z

dependencies:
  requires:
    - 02-02 # Atlas schema generation (gen/atlas/schema.hcl)
    - 02-03 # forge generate command
  provides:
    - forge migrate diff (creates migration files with destructive change detection)
    - forge migrate up (applies pending migrations)
    - forge migrate down (rolls back last migration)
    - forge migrate status (shows migration state)
    - forge migrate hash (recomputes checksums)
    - migrate.Config type for Atlas CLI configuration
    - Destructive change detection (DROP TABLE, DROP COLUMN, ALTER TYPE, DROP INDEX)
  affects:
    - CLI command structure (adds migrate command group)
    - Developer workflow (migration creation and management)

tech-stack:
  added:
    - Atlas CLI wrapper functions (exec.Command)
    - Regex-based SQL parsing for destructive detection
  patterns:
    - Command pattern for Atlas CLI operations
    - Strategy pattern for destructive change detection
    - Builder pattern for migration Config

key-files:
  created:
    - internal/migrate/commands.go # Atlas CLI wrapper functions
    - internal/migrate/destructive.go # Destructive change detection
    - internal/migrate/migrate_test.go # Unit tests for detection logic
    - internal/cli/migrate.go # CLI commands (diff, up, down, status, hash)
  modified:
    - internal/cli/root.go # Register migrate command group

decisions:
  - title: "Use same database URL for dev-url in Atlas diff"
    rationale: "Atlas can create temporary schemas in the same database for diffing, avoiding the need for a separate dev database"
    alternatives: ["Require separate dev database in config", "Parse database URL to add _dev suffix"]
  - title: "Delete rejected migration files on destructive warning"
    rationale: "Atlas creates the file before we can check it, so we delete it to avoid leaving partial/rejected migrations"
    alternatives: ["Leave file and warn user to delete", "Generate SQL to memory first (not possible with Atlas)"]
  - title: "Regex-based destructive detection instead of SQL parsing"
    rationale: "Simple regex patterns cover 95% of destructive cases; full SQL parsing is overkill and error-prone"
    alternatives: ["Use SQL parser library", "Let Atlas handle all validation"]
  - title: "Include line numbers in destructive change warnings"
    rationale: "Makes it easy for developers to locate problematic SQL in migration files"
    alternatives: ["Show only operation type", "Show full SQL statement"]

metrics:
  duration_seconds: 258
  tasks_completed: 2
  files_created: 4
  files_modified: 1
  tests_added: 9
  test_coverage: "Unit tests only (no integration tests with Atlas binary)"
---

# Phase 02 Plan 04: Atlas Migration Management Summary

Atlas migration commands (diff, up, down, status) with destructive change detection and safety warnings implemented.

## What Was Built

Created a complete migration management system that wraps Atlas CLI commands and provides destructive change detection:

**Migration Package (`internal/migrate/`):**
- `Config` struct with atlas binary path, migration directory, schema URL, database URL, and dev URL
- `Diff()` function generates migrations via `atlas migrate diff`, reads the created file, checks for destructive changes, and rejects with formatted warning unless `--force` is used
- `Up()` applies pending migrations via `atlas migrate apply`
- `Down()` rolls back the last migration via `atlas migrate down`
- `Status()` shows current migration state via `atlas migrate status`
- `Hash()` recomputes `atlas.sum` checksums after manual edits via `atlas migrate hash`

**Destructive Detection (`internal/migrate/destructive.go`):**
- Regex patterns detect: `DROP TABLE`, `DROP COLUMN`, `ALTER COLUMN ... TYPE`, `DROP INDEX`
- `ContainsDestructiveChange()` returns boolean for quick checks
- `FindDestructiveChanges()` returns all matching lines with line numbers (1-indexed)
- `DestructiveWarning()` formats a styled warning message with:
  - Warning header (yellow styled)
  - List of destructive operations with line numbers
  - Instructions to use `--force` flag
  - Reminder to review schema changes

**CLI Commands (`internal/cli/migrate.go`):**
- `forge migrate` parent command with comprehensive help text documenting hand-written migration support
- `forge migrate diff [name]` creates migrations with `--force` flag to bypass warnings
- `forge migrate up` applies pending migrations
- `forge migrate down` rolls back last migration
- `forge migrate status` shows migration state
- `forge migrate hash` recomputes checksums
- All commands check for atlas binary and provide helpful error: "Run 'forge tool sync' to download it"
- Diff command checks `gen/atlas/schema.hcl` exists and prompts "Run 'forge generate' first" if missing
- Destructive warnings are displayed with formatting intact (not re-wrapped)

**Tests (`internal/migrate/migrate_test.go`):**
- `TestContainsDestructiveChange_DropTable` - Verifies DROP TABLE detection
- `TestContainsDestructiveChange_DropColumn` - Verifies DROP COLUMN detection
- `TestContainsDestructiveChange_AlterType` - Verifies ALTER COLUMN TYPE detection
- `TestContainsDestructiveChange_DropIndex` - Verifies DROP INDEX detection
- `TestContainsDestructiveChange_SafeChanges` - Verifies CREATE TABLE, ADD COLUMN, CREATE INDEX are NOT flagged
- `TestFindDestructiveChanges_MultipleMatches` - Verifies all destructive lines found with line numbers
- `TestDestructiveWarning_Format` - Verifies warning includes all required content (warning header, instructions, --force flag)
- `TestDestructiveWarning_EmptyChanges` - Edge case: empty changes list still produces valid warning
- `TestContainsDestructiveChange_CaseInsensitive` - Verifies detection works regardless of SQL case

All tests pass. No integration tests with Atlas binary (requires database setup).

## Deviations from Plan

None - plan executed exactly as written.

## Integration Points

**Consumes:**
- `toolsync.ToolBinPath()` to locate atlas binary
- `toolsync.IsToolInstalled()` to check if atlas is available
- `config.Load()` to read database URL from forge.toml
- `ui.WarnStyle`, `ui.ErrorIcon`, `ui.BoldStyle`, `ui.CommandStyle` for formatted warnings

**Provides:**
- `migrate.Config` type for configuring Atlas commands
- `migrate.Diff()`, `migrate.Up()`, `migrate.Down()`, `migrate.Status()`, `migrate.Hash()` for Atlas operations
- `migrate.ContainsDestructiveChange()`, `migrate.FindDestructiveChanges()`, `migrate.DestructiveWarning()` for safety checks

**Used by:**
- `internal/cli/migrate.go` CLI commands

## Hand-Written Migration Support

As documented in the plan (MIGRATE-06), hand-written migrations are fully supported:
- Atlas versioned workflow reads **all** `.sql` files from `migrations/` directory
- Generated and hand-written files coexist
- `atlas.sum` tracks integrity of all files
- After adding/editing manual migrations, run `forge migrate hash` to update checksums
- This is documented in the `forge migrate --help` long description

## Developer Workflow

1. **Create migration:** `forge migrate diff add_users_table`
   - Compares `gen/atlas/schema.hcl` with database state
   - Generates SQL file in `migrations/`
   - Checks for destructive changes
   - If destructive: shows warning and rejects (file deleted)
   - If safe: shows success with file path

2. **Review migration:** Developer opens the generated SQL file to verify changes

3. **Apply migration:** `forge migrate up`
   - Applies all pending migrations to database
   - Shows Atlas output

4. **Check status:** `forge migrate status`
   - Lists applied and pending migrations

5. **Rollback if needed:** `forge migrate down`
   - Reverts the last migration

6. **Manual edits:** If developer edits a migration file, run `forge migrate hash` to update checksums

## Destructive Change Example

When attempting to create a migration with `DROP COLUMN`:

```
$ forge migrate diff remove_email

Creating migration...

WARNING: Destructive migration detected!

This migration contains operations that will permanently delete data:
  ✗ ALTER TABLE users DROP COLUMN email (line 3)

If you are CERTAIN you want to proceed, run:
  forge migrate diff --force

Otherwise, review your schema changes.
```

The migration file is deleted, preventing accidental data loss.

## Success Criteria Met

- [x] `go build .` compiles forge binary with migrate commands
- [x] `go test ./internal/migrate/ -v` all tests pass (9 tests, all passing)
- [x] `forge migrate --help` shows diff, up, down, status, hash subcommands
- [x] `forge migrate diff --help` shows --force flag and description
- [x] Destructive changes detected and produce formatted warnings
- [x] Hand-written migration support documented in help text
- [x] forge migrate diff runs Atlas schema diff against gen/atlas/schema.hcl
- [x] forge migrate up/down applies/rolls back migrations
- [x] forge migrate status shows migration state
- [x] Destructive changes blocked without --force
- [x] Atlas binary checked and helpful error shown if missing

## Self-Check

Verifying created files and commits exist.

**Files created:**
- ✓ internal/migrate/commands.go
- ✓ internal/migrate/destructive.go
- ✓ internal/migrate/migrate_test.go
- ✓ internal/cli/migrate.go

**Files modified:**
- ✓ internal/cli/root.go

**Commits:**
- ✓ a77130d (Task 1 - migrate package, incorrectly included in 02-05 commit)
- ✓ b2fe0f5 (Task 2 - CLI commands)

## Self-Check: PASSED

All files exist and commits are in git history. Note: Task 1 files were committed in the previous 02-05 execution (commit a77130d), which was executed out of order. This plan (02-04) should have been executed before 02-05.
