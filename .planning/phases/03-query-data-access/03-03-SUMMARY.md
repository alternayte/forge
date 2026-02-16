---
phase: 03-query-data-access
plan: 03
subsystem: CLI-04
tags: [database, cli, development-tools]
dependency_graph:
  requires: [Atlas migrations (02-04), Config loading (01-02)]
  provides: [forge db commands]
  affects: [Developer workflow, Database management]
tech_stack:
  added: [PostgreSQL client tools (createdb, dropdb, psql)]
  patterns: [CLI command composition, URL parsing, Process spawning]
key_files:
  created:
    - internal/cli/db.go
  modified:
    - internal/cli/root.go
decisions:
  - "Reuse migrate.Up() directly instead of shelling out to forge migrate up (avoids recursive binary invocation)"
  - "Parse host/port/dbname from database URL for PostgreSQL client tool flags"
  - "Treat 'already exists' as no-op for idempotent create command"
  - "Display info message when seed file missing (not an error)"
  - "Pass DATABASE_URL env var to seed file for database connection"
metrics:
  duration: 152s
  completed: 2026-02-16T20:23:35Z
  tasks_completed: 2
  files_created: 1
  files_modified: 1
  commits: 1
---

# Phase 03 Plan 03: Database Management CLI Summary

**One-liner:** PostgreSQL database lifecycle management commands wrapping createdb/dropdb/psql with Atlas migration integration.

## Overview

Implemented `forge db` command group with 5 subcommands (create, drop, console, seed, reset) for development database management. Commands parse connection parameters from forge.toml database URL, wrap PostgreSQL client tools, integrate with Atlas migrations, and support user-defined seed files.

## What Was Built

### forge db Command Group

**Parent Command:**
- `forge db`: Command group for database management with 5 subcommands

**Subcommands:**

1. **forge db create**
   - Creates PostgreSQL database using `createdb`
   - Parses host, port, database name from forge.toml URL
   - Idempotent: displays info message if database already exists
   - Provides helpful error if PostgreSQL client tools not installed

2. **forge db drop**
   - Drops PostgreSQL database using `dropdb --if-exists`
   - Parses connection parameters from forge.toml URL
   - Idempotent: no-op if database doesn't exist

3. **forge db console**
   - Opens interactive `psql` session
   - Connects using full database URL from forge.toml
   - Hands off terminal control (stdin/stdout/stderr) to psql

4. **forge db seed**
   - Runs user-defined `db/seed.go` file
   - Executes as `go run ./db/seed.go`
   - Passes `DATABASE_URL` environment variable
   - Displays info message if seed file missing (not an error)
   - Pipes stdout/stderr for visibility

5. **forge db reset**
   - Full database reset sequence:
     1. Drop database (dropdb)
     2. Create database (createdb)
     3. Apply migrations (migrate.Up)
     4. Run seed (if exists)
   - Progress messages for each step
   - Stops on first error
   - Checks Atlas binary availability before starting
   - Skips seed step gracefully if db/seed.go doesn't exist

### Helper Functions

**parseDatabaseURL(rawURL string) (host, port, dbName string, err error)**
- Parses postgres:// and postgresql:// URLs
- Extracts hostname (default "localhost"), port (default "5432"), database name
- Returns structured connection parameters for PostgreSQL client tools

**createDatabase(host, port, dbName string) error**
- Executes `createdb -h host -p port dbname`
- Handles "already exists" case as no-op
- Provides helpful error if createdb not installed

**dropDatabase(host, port, dbName string) error**
- Executes `dropdb --if-exists -h host -p port dbname`
- Always succeeds (--if-exists handles missing database)

**runSeed(projectRoot string, cfg *config.Config) error**
- Checks if db/seed.go exists
- Executes `go run ./db/seed.go` from project root
- Sets DATABASE_URL environment variable
- Displays info message if seed file missing (returns nil, not error)

## Implementation Notes

### URL Parsing Strategy

Database URLs in forge.toml follow the format:
```
postgres://localhost:5432/my-app?sslmode=disable
```

The `parseDatabaseURL` helper extracts:
- **Host:** From URL hostname (default "localhost")
- **Port:** From URL port (default "5432")
- **Database name:** From URL path, stripping leading slash

This allows passing individual parameters to createdb/dropdb while using the full URL for psql.

### Reset Command Architecture

The reset command orchestrates 4 steps with clear separation:

1. **Drop/Create:** Use shared helper functions that wrap PostgreSQL tools
2. **Migrate:** Call `migrate.Up()` directly (no shell-out to avoid recursive binary call)
3. **Seed:** Use shared runSeed helper (skips if missing)

Each step displays progress messages and stops on error. This provides transparency and debuggability.

### Seed File Contract

The seed file at `db/seed.go` is user-written. Forge provides:
- **DATABASE_URL env var:** Connection string from forge.toml
- **Execution context:** Run from project root with `go run`
- **Output visibility:** stdout/stderr piped to terminal

The user's seed file is responsible for:
- Parsing DATABASE_URL and connecting to database
- Using generated factories (from gen/factories/) or raw SQL
- Creating test data appropriate for development

Forge doesn't generate or scaffold the seed file - it's fully user-controlled.

### Error Handling: PostgreSQL Client Tools

All commands that use createdb, dropdb, or psql check for `exec.ErrNotFound` and provide a helpful message:

```
PostgreSQL client tools not found.
Install postgresql-client or postgresql to get createdb, dropdb, and psql.
```

This guides developers to install the required system dependencies.

## Deviations from Plan

None - plan executed exactly as written. Both tasks were implemented together in a single commit since they were part of the same file and the implementation was more efficient as a cohesive unit.

## Testing Notes

**Manual Verification:**
- ✅ `go build .` - compiles successfully
- ✅ `./forge db --help` - lists all 5 subcommands
- ✅ `./forge db create --help` - correct help text
- ✅ `./forge db reset --help` - correct help text
- ✅ `go vet ./internal/cli/` - no issues

**URL Parsing Coverage:**
- Handles `postgres://` and `postgresql://` schemes
- Extracts host, port, database name correctly
- Defaults to localhost:5432 when not specified
- Strips query parameters from database name

## Integration Points

### Depends On
- **Config loading (01-02):** Reads forge.toml for database URL
- **Atlas migrations (02-04):** Calls migrate.Up() for reset command
- **UI styling (01-05):** Uses ui.Success/Error/Info/Header for consistent output

### Provides To
- **Developer workflow:** Complete database lifecycle management
- **CI/CD scripts:** Scriptable database setup (forge db reset)
- **Future query layer (03-01, 03-02):** Database creation/seeding for testing

### Affects
- **forge dev (02-05):** May integrate db setup in future
- **Testing workflows:** Enables clean database state for tests
- **Documentation:** Developer guide will reference these commands

## Known Limitations

1. **PostgreSQL only:** Commands assume PostgreSQL client tools. Other databases not supported.
2. **No connection pooling:** Each command spawns new process. Fine for CLI, not suitable for programmatic use.
3. **No auth prompt handling:** If database requires password, user must configure .pgpass or environment variables.
4. **Seed file not scaffolded:** Users must create db/seed.go manually. Future enhancement could add `forge db seed:init`.

## Success Criteria

✅ forge db create/drop use createdb/dropdb with parsed URL
✅ forge db console opens interactive psql session
✅ forge db seed runs db/seed.go with DATABASE_URL
✅ forge db reset orchestrates full drop/create/migrate/seed sequence
✅ All commands follow existing CLI patterns (ui styling, project root detection, config loading)

## Self-Check: PASSED

**Commits exist:**
```bash
$ git log --oneline --all | grep d070703
d070703 feat(03-03): implement forge db create, drop, and console commands
```

**Files created:**
```bash
$ [ -f "internal/cli/db.go" ] && echo "FOUND: internal/cli/db.go" || echo "MISSING: internal/cli/db.go"
FOUND: internal/cli/db.go
```

**Files modified:**
```bash
$ git diff --name-only d070703^..d070703
internal/cli/db.go
internal/cli/root.go
```

All files verified. Implementation complete and committed.
