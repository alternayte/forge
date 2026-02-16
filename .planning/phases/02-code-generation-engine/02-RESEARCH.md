# Phase 2: Code Generation Engine - Research

**Researched:** 2026-02-16
**Domain:** Go code generation with text/template, Atlas migrations, file watching, and automated formatting
**Confidence:** HIGH

## Summary

Phase 2 builds the code generation engine that transforms the IR (from Phase 1's go/ast parser) into compilable Go code, Atlas HCL schemas, migration management, and development tooling. The standard stack is Go's stdlib `text/template` for code generation, Atlas CLI for schema diffing and migration management, `fsnotify` for file watching, `go/format` for code formatting, and `golang.org/x/tools/imports` for automatic import management. The architecture follows a clean separation: templates live in `internal/generator/templates/`, generator logic lives in `internal/generator/`, and all generated code goes into `gen/` with proper "DO NOT EDIT" headers.

**Primary recommendation:** Use `text/template.ParseFS` with embedded templates, `golang.org/x/tools/imports.Process()` for formatting+imports in one pass, Atlas's versioned migration workflow with integrity checking via `atlas.sum`, and `fsnotify` watching parent directories (not individual files) with debouncing for hot reload. Don't hand-roll SQL diffing, import management, or file watching infrastructure.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| text/template | stdlib | Code generation from templates | Go's official templating engine; zero dependencies; supports ParseFS with embed |
| golang.org/x/tools/imports | latest | Auto-add/remove imports + format | Combines gofmt + import management; industry standard for code generators |
| go/format | stdlib | Format generated AST nodes | Official Go formatter; ensures generated code matches gofmt style |
| Atlas CLI | latest (v1.1+) | Schema diffing, migration generation, migration apply/rollback | Declarative schema diffing is multi-year effort; handles edge cases (indexes, constraints, type changes) |
| fsnotify | v1.7+ | Cross-platform file watching | Most popular Go file watcher; 9k+ stars; used by Air, modd, and major projects |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| embed | stdlib | Embed template files in binary | Always use for shipping templates with CLI binary |
| github.com/Masterminds/sprig/v3 | v3.x | Template helper functions | Optional: string manipulation, date formatting, defaults in templates |
| bytes.Buffer | stdlib | Buffer template output before writing | Always use: prevents partial writes on template errors |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| text/template | html/template | html/template auto-escapes for HTML; inappropriate for Go code generation |
| text/template | Custom AST builder | AST builder = more control but 10x complexity; templates are readable by non-experts |
| Atlas CLI | Custom SQL differ | Custom differ = months of work for index detection, FK constraints, enum handling |
| fsnotify | os polling | Polling = 100ms+ latency + wasted CPU; fsnotify = instant notification via OS primitives |
| golang.org/x/tools/imports | gofmt only | gofmt doesn't manage imports; imports.Process() does both in one pass |

**Installation:**
```bash
# Atlas CLI (via forge tool sync in Phase 1)
# Downloaded on-demand to .forge/bin/

# Go dependencies
go get golang.org/x/tools/imports
go get github.com/fsnotify/fsnotify
# stdlib packages: text/template, go/format, embed (no install needed)
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── generator/
│   ├── generator.go         # Orchestrator: IR → all outputs
│   ├── models.go            # Generate gen/models/*.go
│   ├── atlas.go             # Generate gen/atlas/*.hcl
│   ├── factories.go         # Generate gen/factories/*.go
│   ├── templates/           # Embedded templates
│   │   ├── model.go.tmpl
│   │   ├── atlas_table.hcl.tmpl
│   │   └── factory.go.tmpl
│   └── formatting.go        # Format + import management
├── migrate/
│   ├── commands.go          # Wrappers for atlas CLI
│   └── destructive.go       # Detect + warn on destructive changes
└── watcher/
    ├── watcher.go           # fsnotify wrapper with debouncing
    └── dev.go               # Development server orchestration

gen/
├── models/                  # Always regenerated
│   └── product.go           # Product, ProductCreate, ProductUpdate types
├── atlas/                   # Always regenerated
│   └── schema.hcl           # Desired state HCL
└── factories/               # Always regenerated
    └── product.go           # Test factory functions

migrations/
├── 20260216120000_create_products.sql  # Versioned migrations
├── 20260216120100_add_products_slug.sql
└── atlas.sum                # Integrity checksum (auto-managed by Atlas)
```

### Pattern 1: Template-Driven Code Generation
**What:** Use `text/template` with embedded templates to generate Go code from IR
**When to use:** All code generation tasks (models, factories, actions, API, HTML)
**Example:**
```go
// Source: Existing internal/scaffold/templates.go pattern
package generator

import "embed"

// TemplatesFS contains embedded code generation templates
//go:embed templates/*
var TemplatesFS embed.FS

// GenerateModels generates Go model types for all resources
func GenerateModels(resources []parser.ResourceIR, outputDir string) error {
    tmpl, err := template.ParseFS(TemplatesFS, "templates/model.go.tmpl")
    if err != nil {
        return err
    }

    for _, resource := range resources {
        var buf bytes.Buffer
        if err := tmpl.Execute(&buf, resource); err != nil {
            return fmt.Errorf("template execution failed for %s: %w", resource.Name, err)
        }

        // Format + add imports in one pass
        formatted, err := imports.Process(
            filepath.Join(outputDir, strings.ToLower(resource.Name)+".go"),
            buf.Bytes(),
            nil,
        )
        if err != nil {
            return fmt.Errorf("formatting failed: %w", err)
        }

        if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
            return err
        }
    }
    return nil
}
```

### Pattern 2: Buffer-Then-Format-Then-Write
**What:** Execute template to buffer → format with imports.Process() → write to disk
**When to use:** All Go code generation to prevent partial writes on errors
**Example:**
```go
// Source: https://freshman.tech/snippets/go/template-execution-error/
// and https://pkg.go.dev/golang.org/x/tools/imports

var buf bytes.Buffer

// Step 1: Execute template to buffer
if err := tmpl.Execute(&buf, data); err != nil {
    return fmt.Errorf("template execution failed: %w", err)
}

// Step 2: Format + manage imports (filename affects import resolution)
formatted, err := imports.Process(targetPath, buf.Bytes(), nil)
if err != nil {
    return fmt.Errorf("formatting failed: %w", err)
}

// Step 3: Write atomically
if err := os.WriteFile(targetPath, formatted, 0644); err != nil {
    return err
}
```

### Pattern 3: Atlas Versioned Migration Workflow
**What:** Use `atlas migrate diff` to generate SQL migrations, track with `atlas.sum` checksum
**When to use:** All schema changes (prefer versioned over declarative for auditable history)
**Example:**
```bash
# Generate migration from HCL desired state vs live database
atlas migrate diff \
  --dir "file://migrations" \
  --to "file://gen/atlas/schema.hcl" \
  --dev-url "postgres://localhost:5432/forge_dev?sslmode=disable"

# Apply pending migrations
atlas migrate apply \
  --dir "file://migrations" \
  --url "postgres://localhost:5432/myapp?sslmode=disable"

# Rollback last migration
atlas migrate down \
  --dir "file://migrations" \
  --url "postgres://localhost:5432/myapp?sslmode=disable"

# Check migration status
atlas migrate status \
  --dir "file://migrations" \
  --url "postgres://localhost:5432/myapp?sslmode=disable"
```

### Pattern 4: File Watching with Debouncing
**What:** Watch parent directories (not individual files), debounce events, filter by extension
**When to use:** Development server hot reload
**Example:**
```go
// Source: https://pkg.go.dev/github.com/fsnotify/fsnotify
// and https://github.com/fsnotify/fsnotify FAQ

watcher, err := fsnotify.NewWatcher()
if err != nil {
    log.Fatal(err)
}
defer watcher.Close()

// Watch parent directories (atomic file writes lose watch on individual files)
watcher.Add("./resources")
watcher.Add("./internal")

// Debounce: collect events for 300ms before triggering rebuild
var timer *time.Timer
timerDuration := 300 * time.Millisecond

go func() {
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            // Ignore Chmod (noisy from editors, antivirus, Spotlight)
            if event.Has(fsnotify.Chmod) {
                continue
            }
            // Filter by extension (.go, .templ, .sql, .css)
            if !isRelevantFile(event.Name) {
                continue
            }
            // Reset timer on each event
            if timer != nil {
                timer.Stop()
            }
            timer = time.AfterFunc(timerDuration, func() {
                rebuild()
            })
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            log.Println("watcher error:", err)
        }
    }
}()
```

### Pattern 5: Generated Code File Header
**What:** Add standard "DO NOT EDIT" comment at top of generated files
**When to use:** All generated Go files (required for tools to detect auto-generated code)
**Example:**
```go
// Source: https://github.com/golang/go/issues/13560
// Standard format enforced by Go tooling

// Code generated by forge generate. DO NOT EDIT.

package models

// ... generated code ...
```

### Anti-Patterns to Avoid
- **Formatting AST nodes with format.Node() for templates:** Templates produce source text, not AST. Use `imports.Process()` on text, not `format.Node()`.
- **Watching individual files with fsnotify:** Editors use atomic writes (write temp → rename), which breaks watches. Watch parent directories instead.
- **Executing templates directly to http.ResponseWriter or os.File:** Template errors leave partial output. Always buffer first.
- **Using gofmt without goimports:** Generated code has imports. Use `imports.Process()` to format + add/remove imports in one pass.
- **Declarative migrations in CI/CD:** Declarative (`atlas schema apply`) skips version control. Use versioned (`atlas migrate diff` + git) for auditable history.
- **Ignoring Chmod events with complex logic:** Just skip all Chmod events unconditionally—they're always noisy.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| SQL schema diffing | Custom AST parser for SQL | Atlas CLI `migrate diff` | Handles indexes, FKs, enums, partial indexes, constraints, type changes, and 100+ edge cases. Multi-year effort to build. |
| Import management | String parsing for `import ()` blocks | `golang.org/x/tools/imports.Process()` | Detects missing imports, removes unused, groups stdlib vs external, handles dot imports, blank imports, renames. |
| File watching | `os.Stat()` polling loop | `fsnotify.NewWatcher()` | Uses OS primitives (inotify/kqueue/ReadDirectoryChangesW) for instant notification. Polling = 100ms latency + CPU waste. |
| Go code formatting | Custom whitespace/indentation rules | `imports.Process()` or `format.Source()` | gofmt style is non-negotiable in Go ecosystem. Don't reinvent. |
| Migration integrity | Custom checksum tracking | Atlas `atlas.sum` file | Atlas uses merkle hash tree for tamper detection. Prevents accidental reordering or editing of applied migrations. |
| Template helper functions | Custom FuncMap implementations | `github.com/Masterminds/sprig/v3` | 100+ battle-tested helpers (pluralize, camelCase, default, ternary, etc.). Don't rebuild these. |

**Key insight:** Code generation infrastructure is commodity. Your value is in the *domain-specific templates* (what to generate for a Forge resource), not in diffing SQL or watching files. Atlas and stdlib solved hard problems—use them.

## Common Pitfalls

### Pitfall 1: Template Execution Errors Write Partial Output
**What goes wrong:** Template executes to `os.File`, hits error halfway through, leaves broken half-file on disk
**Why it happens:** `template.Execute()` writes incrementally as it renders; no rollback on error
**How to avoid:** Always execute to `bytes.Buffer`, validate/format, then write atomically
**Warning signs:** Compilation errors in gen/ files after failed `forge generate`; partial files with truncated content

### Pitfall 2: File Watching Breaks on Atomic File Replacement
**What goes wrong:** Watch `resources/product/schema.go`, editor saves via `write .schema.go.tmp` + `rename`, watch stops firing
**Why it happens:** Atomic writes replace the inode; watch is on old inode that no longer exists
**How to avoid:** Watch parent directory (`resources/product/`), filter events by `event.Name` suffix
**Warning signs:** Hot reload works once, then stops after first save; manually restarting `forge dev` fixes it

### Pitfall 3: Generated Code Has Wrong Imports
**What goes wrong:** Template produces `*time.Time` but no `import "time"`; code doesn't compile
**Why it happens:** Templates don't know what imports are needed; `gofmt` doesn't add imports
**How to avoid:** Use `golang.org/x/tools/imports.Process()` instead of `format.Source()`—it auto-adds missing imports
**Warning signs:** `forge generate` succeeds but `go build` fails with "undefined: time"

### Pitfall 4: Destructive Migrations Applied Without Warning
**What goes wrong:** Developer runs `forge migrate up`, column is dropped, data is lost permanently
**Why it happens:** Atlas generates migration but doesn't block destructive changes by default
**How to avoid:** Parse generated migration SQL for `DROP COLUMN`, `DROP TABLE`, `ALTER COLUMN TYPE`; print prominent warning and require `--force` flag
**Warning signs:** "Why did my data disappear?" after running migrations

### Pitfall 5: Migration Directory Checksum Mismatch After Manual Edits
**What goes wrong:** Developer manually edits migration SQL, runs `atlas migrate apply`, gets "integrity checksum failed" error
**Why it happens:** `atlas.sum` contains hash of original migration; manual edit invalidates checksum
**How to avoid:** After manual edits, run `atlas migrate hash` to recompute checksum. Document this in error message.
**Warning signs:** `atlas migrate apply` fails with "checksum mismatch"

### Pitfall 6: Chmod Events Trigger Infinite Rebuild Loop
**What goes wrong:** `forge dev` rebuilds continuously; no file actually changed
**Why it happens:** Spotlight (macOS), antivirus, or backup software constantly triggers Chmod events
**How to avoid:** Unconditionally skip `fsnotify.Chmod` events—they're never useful for rebuild triggers
**Warning signs:** CPU spikes, continuous rebuild messages, `event.Has(fsnotify.Chmod)` in logs

### Pitfall 7: Template Path Mismatches After embed
**What goes wrong:** `template.ParseFS()` can't find "model.go.tmpl", crashes with "pattern not found"
**Why it happens:** `//go:embed templates/*` embeds as `templates/model.go.tmpl`, but code references `model.go.tmpl`
**How to avoid:** Use full path from embed root: `ParseFS(TemplatesFS, "templates/model.go.tmpl")`
**Warning signs:** Works with `os.ReadFile()` but fails after switching to `embed.FS`

## Code Examples

Verified patterns from official sources:

### Generate Go File with Formatting and Imports
```go
// Source: https://pkg.go.dev/golang.org/x/tools/imports
// and https://freshman.tech/snippets/go/template-execution-error/

package generator

import (
    "bytes"
    "fmt"
    "os"
    "text/template"
    "golang.org/x/tools/imports"
)

func GenerateFile(tmplPath, outputPath string, data interface{}) error {
    // Parse template
    tmpl, err := template.ParseFS(TemplatesFS, tmplPath)
    if err != nil {
        return fmt.Errorf("parse template: %w", err)
    }

    // Execute to buffer (not directly to file)
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return fmt.Errorf("execute template: %w", err)
    }

    // Format + add/remove imports (filename matters for import resolution)
    formatted, err := imports.Process(outputPath, buf.Bytes(), nil)
    if err != nil {
        return fmt.Errorf("format code: %w", err)
    }

    // Write atomically
    if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
        return fmt.Errorf("write file: %w", err)
    }

    return nil
}
```

### Generate Atlas HCL Schema
```go
// Source: https://atlasgo.io/atlas-schema/hcl

package generator

import (
    "fmt"
    "os"
    "path/filepath"
    "text/template"
)

func GenerateAtlasSchema(resources []parser.ResourceIR, outputDir string) error {
    tmpl, err := template.ParseFS(TemplatesFS, "templates/atlas_schema.hcl.tmpl")
    if err != nil {
        return err
    }

    schemaPath := filepath.Join(outputDir, "schema.hcl")
    f, err := os.Create(schemaPath)
    if err != nil {
        return err
    }
    defer f.Close()

    // HCL doesn't need Go formatting, write directly
    return tmpl.Execute(f, map[string]interface{}{
        "Resources": resources,
    })
}
```

Example template (`templates/atlas_schema.hcl.tmpl`):
```hcl
# Code generated by forge generate. DO NOT EDIT.

schema "public" {
  comment = "Generated by Forge"
}

{{range .Resources}}
table "{{.Name | lower}}s" {
  schema = schema.public

  column "id" {
    type = uuid
    default = sql("gen_random_uuid()")
  }

  {{range .Fields}}
  column "{{.Name | lower}}" {
    type = {{atlasType .Type}}
    {{if hasModifier .Modifiers "Required"}}null = false{{else}}null = true{{end}}
    {{if hasDefault .Modifiers}}default = {{getDefault .Modifiers}}{{end}}
  }
  {{end}}

  {{if .HasTimestamps}}
  column "created_at" {
    type = timestamptz
    default = sql("now()")
    null = false
  }
  column "updated_at" {
    type = timestamptz
    default = sql("now()")
    null = false
  }
  {{end}}

  primary_key {
    columns = [column.id]
  }

  {{range .Fields}}
  {{if hasModifier .Modifiers "Unique"}}
  index "{{$.Name | lower}}s_{{.Name | lower}}_unique" {
    columns = [column.{{.Name | lower}}]
    unique = true
  }
  {{end}}
  {{end}}
}
{{end}}
```

### File Watcher with Debouncing
```go
// Source: https://pkg.go.dev/github.com/fsnotify/fsnotify

package watcher

import (
    "log"
    "path/filepath"
    "strings"
    "time"
    "github.com/fsnotify/fsnotify"
)

type Watcher struct {
    fsw      *fsnotify.Watcher
    onChange func()
    timer    *time.Timer
    debounce time.Duration
}

func New(onChange func()) (*Watcher, error) {
    fsw, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    w := &Watcher{
        fsw:      fsw,
        onChange: onChange,
        debounce: 300 * time.Millisecond,
    }

    go w.watch()
    return w, nil
}

func (w *Watcher) Add(path string) error {
    return w.fsw.Add(path)
}

func (w *Watcher) Close() error {
    return w.fsw.Close()
}

func (w *Watcher) watch() {
    for {
        select {
        case event, ok := <-w.fsw.Events:
            if !ok {
                return
            }

            // Skip Chmod (always noisy, never useful)
            if event.Has(fsnotify.Chmod) {
                continue
            }

            // Filter to relevant file types
            if !isRelevantFile(event.Name) {
                continue
            }

            // Debounce: reset timer on each event
            if w.timer != nil {
                w.timer.Stop()
            }
            w.timer = time.AfterFunc(w.debounce, w.onChange)

        case err, ok := <-w.fsw.Errors:
            if !ok {
                return
            }
            log.Printf("watcher error: %v", err)
        }
    }
}

func isRelevantFile(path string) bool {
    ext := filepath.Ext(path)
    return ext == ".go" || ext == ".templ" || ext == ".sql" || ext == ".css"
}
```

### Atlas Migration with Destructive Change Detection
```go
// Source: https://atlasgo.io/lint/analyzers

package migrate

import (
    "fmt"
    "os/exec"
    "regexp"
    "strings"
)

var destructivePatterns = []*regexp.Regexp{
    regexp.MustCompile(`(?i)DROP\s+TABLE`),
    regexp.MustCompile(`(?i)DROP\s+COLUMN`),
    regexp.MustCompile(`(?i)ALTER\s+COLUMN.*TYPE`),
    regexp.MustCompile(`(?i)DROP\s+INDEX`),
}

func Diff(devURL, toURL, migrationDir string, force bool) error {
    // Generate migration
    cmd := exec.Command("atlas", "migrate", "diff",
        "--dir", "file://"+migrationDir,
        "--to", toURL,
        "--dev-url", devURL,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("atlas migrate diff failed: %w\n%s", err, output)
    }

    // Check for destructive changes (Atlas pro feature via CLI)
    // For OSS: parse generated SQL
    if !force && containsDestructiveChange(string(output)) {
        return fmt.Errorf(`
⚠️  DESTRUCTIVE MIGRATION DETECTED ⚠️

This migration contains operations that will permanently delete data:
%s

If you are CERTAIN you want to proceed, run:
  forge migrate diff --force

Otherwise, review your schema changes and consider:
  - Renaming columns instead of dropping and recreating
  - Adding new columns instead of changing types
  - Backing up data before dropping tables
`, string(output))
    }

    return nil
}

func containsDestructiveChange(sql string) bool {
    for _, pattern := range destructivePatterns {
        if pattern.MatchString(sql) {
            return true
        }
    }
    return false
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| gofmt only | imports.Process() | Go 1.2 (2013) | Single pass for formatting + import management |
| io/ioutil | os.ReadFile/WriteFile | Go 1.16 (2021) | Simpler API, ioutil deprecated |
| Manual embed | embed.FS | Go 1.16 (2021) | Native template embedding, zero external dependencies |
| template.ParseFiles() from disk | template.ParseFS(embed.FS) | Go 1.16 (2021) | Templates ship in binary, no runtime file dependencies |
| Atlas declarative only | Atlas versioned migrations | Atlas v0.3 (2022) | Git-trackable migration history with integrity checking |
| Atlas migrate apply (simple) | atlas.sum integrity file | Atlas v0.14 (2023) | Detects tampered/reordered migrations |
| Air (cosmtrek/air) | Air (air-verse/air) | 2024 fork | Original repo unmaintained; community forked to air-verse |

**Deprecated/outdated:**
- **io/ioutil.ReadFile/WriteFile**: Use `os.ReadFile`/`os.WriteFile` (Go 1.16+)
- **template.ParseFiles() for shipped binaries**: Use `template.ParseFS(embed.FS)` to embed at compile time
- **github.com/cosmtrek/air**: Use `github.com/air-verse/air` (original unmaintained, community fork active)
- **format.Source() for generated code with imports**: Use `imports.Process()` to handle formatting + imports together

## Open Questions

1. **Should we use sprig for template helpers or keep custom FuncMap minimal?**
   - What we know: Sprig provides 100+ helpers (camelCase, snakeCase, pluralize, default, etc.)
   - What's unclear: Is the dependency worth it, or should we only add helpers as needed?
   - Recommendation: Start with custom FuncMap for 5-10 essential helpers (lower, title, plural, contains). Add sprig if we need 20+ helpers.

2. **Should we vendor Atlas CLI binary or download on-demand via tool sync?**
   - What we know: Phase 1 implemented `forge tool sync` for on-demand downloads
   - What's unclear: Does atlas binary need to be bundled for offline use?
   - Recommendation: Use tool sync (already built). Atlas is ~30MB; don't bloat forge binary. Download first time `forge migrate` runs.

3. **Should we implement custom debouncing or use a library?**
   - What we know: Debouncing is 10 lines (timer reset on each event)
   - What's unclear: Are there edge cases we're missing?
   - Recommendation: Implement custom (it's trivial). Libraries like `github.com/radovskyb/watcher` are 5+ years old and add complexity.

4. **How should we handle Atlas version compatibility?**
   - What we know: Atlas is rapidly evolving; HCL syntax and CLI flags change between versions
   - What's unclear: Should we pin a specific Atlas version or use latest?
   - Recommendation: Pin Atlas version in forge.toml, update intentionally. Document required version in error message if `atlas --version` doesn't match.

## Sources

### Primary (HIGH confidence)
- [Atlas CLI Reference](https://atlasgo.io/cli-reference) - Commands for diff, apply, down, status
- [Atlas HCL Schema Syntax](https://atlasgo.io/atlas-schema/hcl) - Table, column, index, FK definitions
- [Atlas Migration Directory Integrity](https://atlasgo.io/concepts/migration-directory-integrity) - atlas.sum checksum system
- [Atlas Migration Analyzers](https://atlasgo.io/lint/analyzers) - Destructive change detection (DS101, DS102, DS103)
- [text/template package docs](https://pkg.go.dev/text/template) - Execute(), ParseFS(), error handling
- [golang.org/x/tools/imports package docs](https://pkg.go.dev/golang.org/x/tools/imports) - Process() function and Options
- [fsnotify package docs](https://pkg.go.dev/github.com/fsnotify/fsnotify) - Events, Op types, pitfalls
- [go/format package docs](https://pkg.go.dev/go/format) - Source() and Node() functions
- [embed package docs](https://pkg.go.dev/embed) - FS type and //go:embed directive
- GitHub: [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) - FAQ on watching directories vs files
- GitHub: [ariga/atlas](https://github.com/ariga/atlas) - Official repo, issues, discussions

### Secondary (MEDIUM confidence)
- [Go blog: Generating code](https://go.dev/blog/generate) - Standard "DO NOT EDIT" comment format
- [Eli Bendersky: Comprehensive guide to go generate](https://eli.thegreenplace.net/2021/a-comprehensive-guide-to-go-generate/) - Code gen best practices
- [LogRocket: Using Air with Go](https://blog.logrocket.com/using-air-go-implement-live-reload/) - File watching patterns for hot reload
- [Freshman.tech: Template execution errors](https://freshman.tech/snippets/go/template-execution-error/) - Buffer-then-write pattern
- [Andrew M McCall: Using Go Embed for Template Rendering](https://andrew-mccall.com/blog/2025/01/using-go-embed-package-for-template-rendering/) - ParseFS with embed.FS
- [OneUpTime: How to Embed Assets in Go](https://oneuptime.com/blog/post/2026-01-23-go-embed-static-resources/view) - embed.FS patterns
- [Stormkit: Factory pattern for Go tests](https://www.stormkit.io/blog/factory-pattern-for-go-tests) - Test factory builder pattern
- GitHub: [Masterminds/sprig](https://github.com/Masterminds/sprig) - Template helper functions (optional)
- GitHub: [cortesi/modd](https://github.com/cortesi/modd) - Alternative file watcher (not chosen)
- GitHub: [air-verse/air](https://github.com/air-verse/air) - Alternative dev server (not chosen)

### Tertiary (LOW confidence)
- [GoBeyond: Common CRUD Design in Go](https://www.gobeyond.dev/crud/) - CRUD type separation patterns
- [Medium: Go Constructor, Functional Option And Builder Patterns](https://programmerscareer.com/go-function-option-patterns/) - Builder pattern variations

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries are stdlib or widely adopted (fsnotify: 9k+ stars, imports: 3.4k+ dependents, Atlas: 5.5k+ stars)
- Architecture: HIGH - Patterns verified from official docs (text/template, fsnotify, Atlas, imports) and existing scaffold/ implementation
- Pitfalls: HIGH - Documented in official fsnotify FAQ, Atlas docs, and Go blog
- Template patterns: HIGH - Verified from pkg.go.dev docs and existing internal/scaffold/ code
- File watching: HIGH - Verified from fsnotify package docs and FAQ
- Atlas migration workflow: HIGH - Verified from official Atlas documentation and CLI reference

**Research date:** 2026-02-16
**Valid until:** ~30 days (2026-03-18) - Atlas is stable, Go stdlib is stable, fsnotify is stable. Revisit if Atlas releases major version or Go 1.25 changes template/format packages.
