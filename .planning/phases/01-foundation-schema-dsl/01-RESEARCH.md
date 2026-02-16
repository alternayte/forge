# Phase 1: Foundation & Schema DSL - Research

**Researched:** 2026-02-16
**Domain:** Go static analysis, CLI frameworks, fluent API design, binary distribution
**Confidence:** HIGH

## Summary

Phase 1 requires building three core systems: (1) a fluent schema definition API parseable via go/ast without generated code, (2) a polished Cobra-based CLI with Rust-style error reporting, and (3) cross-platform tool binary management. The bootstrapping constraint (parsing schemas before gen/ exists) is the defining technical challenge.

The Go ecosystem provides mature solutions for all requirements. The go/ast and go/types packages enable sophisticated static analysis. Cobra is the industry-standard CLI framework (used by Kubernetes, Docker, Hugo). Charmbracelet's ecosystem (Lipgloss, Huh, Spinner) provides production-ready terminal UI components. No npm dependency is achievable using embed.FS for templates and pure Go implementations of templ, sqlc, tailwind CSS processing, and Atlas.

**Primary recommendation:** Use go/ast for AST parsing with go/types for constant resolution, Cobra for CLI structure, Charmbracelet Lipgloss for styled output, pelletier/go-toml/v2 for configuration, and embed.FS for project scaffolding templates.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Schema authoring feel
- Fluent method chaining for field definitions: `schema.String("Title").Required().MaxLen(200)`
- Package-level variables for resource definitions: `var Post = schema.Define("Post", ...)`
- `schema.Define()` takes variadic args — fields, relationships, timestamps all at same level in flat list
- Compact style: one field per line, prioritize scannability
- Relationships are inline entries alongside fields, not a separate block: `schema.BelongsTo("Category", "categories").Optional().OnDelete(schema.SetNull)`

#### CLI behavior & output
- `forge init` creates a full starter project: forge.toml, resources/ directory, example schema, main.go, go.mod, .gitignore, README — ready to `forge generate` immediately
- `forge init my-project` creates new directory; `forge init` (no arg) initializes current directory using its name
- Polished styled output: colors, icons (checkmarks/crosses), grouped by category — similar feel to `cargo build` or `next dev`
- Tool sync is on-demand: binaries (templ, sqlc, tailwind, atlas) downloaded only when a command needs them, not upfront
- `forge generate` works offline — no database connection required. Only `forge migrate` and `forge db` commands need a live DB

#### Error experience
- Rust-style rich errors: show the offending line, underline the problem, suggest a fix — file:line, code snippet, "did you mean...?"
- Collect all errors in a single parse pass — developer fixes everything at once, no cascading noise
- Dynamic value errors explain the constraint: "Forge schemas must use literal values for static analysis. Found variable 'maxLen' — use a constant or literal instead." Teach the why, not just the what
- Each error type includes a reference link or error code (e.g., "See: forge.dev/errors/E001" or `forge help schema-errors`) for deeper context

#### Project structure
- Resource-colocated layout: each resource is its own package directory under `resources/` — schema.go, handlers.go, form.templ all live together per resource
- Generated code in structured `gen/` subdirectories: gen/models/, gen/queries/, gen/atlas/, gen/handlers/ — subdirectories by concern
- forge.toml uses commented template style: all available options present but commented out with defaults shown — developer uncomments to customize

### Claude's Discretion
- Exact field type set and modifier names (String, Int, UUID, etc.)
- Internal IR (intermediate representation) structure from go/ast parsing
- How tool sync detects platform and downloads binaries
- Error code numbering scheme
- Exact forge.toml section organization

</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go/ast | stdlib | Parse Go source into AST | Only way to parse Go without execution |
| go/parser | stdlib | Parse Go files | Required entry point for go/ast |
| go/token | stdlib | Position tracking | Essential for file:line:col error reporting |
| go/types | stdlib | Type checking, constant resolution | Distinguishes constants from variables |
| github.com/spf13/cobra | v1.8+ | CLI framework | Industry standard (Kubernetes, Docker, Hugo, GitHub CLI) |
| github.com/charmbracelet/lipgloss | v0.9+ | Terminal styling | Professional colored output, degrades gracefully |
| github.com/pelletier/go-toml/v2 | v2.1+ | TOML parsing | Faster than BurntSushi, actively maintained |
| embed | stdlib | Embed templates | Bundle scaffolding templates in binary |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/charmbracelet/huh | v0.3+ | Interactive forms | Optional: for `forge init` interactive mode |
| github.com/briandowns/spinner | v1.23+ | Progress indicators | Tool download progress |
| github.com/cuu/grab | v3.0+ | HTTP downloads with checksum | Tool binary downloads with verification |
| golang.org/x/tools/go/ast/astutil | latest | AST utilities | Simplifies import path handling |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Cobra | Kong | Kong simpler but less ecosystem support, no subcommand groups |
| pelletier/go-toml | BurntSushi/toml | BurntSushi simpler API but slower (1.7-5.1x), unsupported |
| Lipgloss | fatih/color | fatih/color lower-level, no layout/styling abstractions |
| grab | Manual http.Get | Manual approach requires implementing progress, resume, checksums |

**Installation:**
```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/pelletier/go-toml/v2@latest
go get github.com/briandowns/spinner@latest
go get github.com/cuu/grab/v3@latest
go get golang.org/x/tools/go/ast/astutil@latest
```

## Architecture Patterns

### Recommended Project Structure
```
forge/
├── cmd/                      # CLI commands
│   ├── root.go              # Root command, persistent flags
│   ├── init.go              # forge init
│   ├── generate.go          # forge generate (Phase 2)
│   └── tool/                # forge tool subcommands
│       └── sync.go          # forge tool sync
├── internal/
│   ├── schema/              # Schema definition API (fluent builders)
│   │   ├── definition.go    # schema.Define()
│   │   ├── field.go         # Field types (String, Int, UUID, etc.)
│   │   ├── modifier.go      # Modifiers (Required, MaxLen, etc.)
│   │   └── relationship.go  # BelongsTo, HasMany, etc.
│   ├── parser/              # AST parsing
│   │   ├── parser.go        # Main parsing logic
│   │   ├── extractor.go     # Extract schema.Define() calls
│   │   ├── validator.go     # Validate literal-only values
│   │   └── ir.go            # Intermediate representation
│   ├── errors/              # Rich error formatting
│   │   ├── diagnostic.go    # Error with source position
│   │   ├── formatter.go     # Rust-style rendering
│   │   └── codes.go         # Error code registry
│   ├── toolsync/            # Binary management
│   │   ├── download.go      # HTTP download with progress
│   │   ├── platform.go      # GOOS/GOARCH detection
│   │   └── verify.go        # Checksum verification
│   ├── scaffold/            # Project initialization
│   │   ├── template.go      # Template rendering
│   │   └── files.go         # File creation logic
│   └── ui/                  # Terminal UI components
│       ├── styles.go        # Lipgloss styles
│       └── spinner.go       # Progress indicators
├── templates/               # Embedded templates
│   ├── forge.toml.tmpl
│   ├── main.go.tmpl
│   ├── resource_schema.go.tmpl
│   └── README.md.tmpl
└── main.go
```

### Pattern 1: AST-Based Schema Parsing (Bootstrapping Solution)

**What:** Parse Go source files containing schema.Define() calls using go/ast, extract literal values without executing code. Enables defining schemas that reference types from gen/ package that doesn't exist yet.

**When to use:** Required for bootstrapping constraint — must parse schemas before generation phase creates gen/ package.

**Example:**
```go
// Source: Composite of go/parser, go/ast, go/types stdlib examples
package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

// Parse schema.go file without requiring gen/ to exist
func ParseSchemaFile(path string) (*SchemaDefinition, error) {
	fset := token.NewFileSet()

	// Parse with comments (for future doc extraction)
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Walk AST to find schema.Define() calls
	var schemaDef *SchemaDefinition
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if call is schema.Define()
		if !isSchemaDefine(call) {
			return true
		}

		// Extract arguments
		schemaDef = extractSchemaDefinition(fset, call)
		return false
	})

	return schemaDef, nil
}

// Check if CallExpr is schema.Define()
func isSchemaDefine(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "schema" && sel.Sel.Name == "Define"
}
```

### Pattern 2: Literal Value Extraction with go/types

**What:** Use go/types for constant folding to distinguish between literal values (allowed) and variables (rejected). Provides precise error messages about why dynamic values aren't supported.

**When to use:** When validating that schema definitions use only literals/constants for static analysis.

**Example:**
```go
// Source: Adapted from go/types stdlib documentation
package parser

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

func ValidateLiteralValues(fset *token.FileSet, files []*ast.File) error {
	// Type check to resolve constants
	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	_, err := conf.Check("resources/product", fset, files, info)
	if err != nil {
		return err
	}

	// Now check if values are constants
	for expr, tv := range info.Types {
		if tv.Value == nil {
			// This is a variable, not a constant
			pos := fset.Position(expr.Pos())
			return &DiagnosticError{
				File:    pos.Filename,
				Line:    pos.Line,
				Column:  pos.Column,
				Message: "schema values must be literals or constants",
				Code:    "E001",
			}
		}
	}

	return nil
}
```

### Pattern 3: Fluent Builder with Deferred Error Handling

**What:** Fluent API that chains methods returning pointers, with validation deferred to schema.Define() which is called during AST parsing, not runtime.

**When to use:** Schema definition API where errors are caught at parse time, not method call time.

**Example:**
```go
// Source: Adapted from builder pattern best practices
package schema

// Field builder accumulates configuration
type Field struct {
	name       string
	fieldType  FieldType
	modifiers  []Modifier
}

// String creates a string field (chainable)
func String(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeString,
		modifiers: []Modifier{},
	}
}

// Required adds required modifier (chainable)
func (f *Field) Required() *Field {
	f.modifiers = append(f.modifiers, ModifierRequired)
	return f
}

// MaxLen adds max length (chainable)
func (f *Field) MaxLen(n int) *Field {
	f.modifiers = append(f.modifiers, Modifier{
		Type:  ModifierMaxLen,
		Value: n,
	})
	return f
}

// Define creates resource definition (validation happens during AST parse)
func Define(name string, items ...interface{}) *Definition {
	// This is never actually called at runtime
	// AST parser extracts literal arguments from the call site
	return &Definition{Name: name}
}
```

### Pattern 4: Cobra Command Organization

**What:** Modular command structure where each feature returns a constructor, keeping boundaries clean.

**When to use:** CLI with multiple subcommands and command groups.

**Example:**
```go
// Source: https://cobra.dev/docs/how-to-guides/working-with-commands/
package cmd

import "github.com/spf13/cobra"

// Root command setup in cmd/root.go
var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Full-stack Go framework with integrated tooling",
}

func init() {
	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")

	// Add subcommands
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newToolCmd())
}

// Each command in separate file
func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new Forge project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Implementation
			return nil
		},
	}

	// Local flags only for this command
	cmd.Flags().BoolP("interactive", "i", false, "interactive mode")

	return cmd
}

// Command groups for organization
func newToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Manage tool binaries",
	}

	cmd.AddCommand(newToolSyncCmd())
	return cmd
}
```

### Pattern 5: Rich Error Formatting

**What:** Rust-style diagnostics showing source context, file:line:col position, underlined error location, and helpful suggestions.

**When to use:** All parser errors and validation errors.

**Example:**
```go
// Source: Inspired by go/token Position and error formatting patterns
package errors

import (
	"fmt"
	"go/token"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Diagnostic struct {
	Pos     token.Position
	Message string
	Code    string
	Hint    string
}

func (d *Diagnostic) Format() string {
	var b strings.Builder

	// Header: error[E001]: message
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true).
		Render(fmt.Sprintf("error[%s]:", d.Code))

	b.WriteString(header)
	b.WriteString(" ")
	b.WriteString(d.Message)
	b.WriteString("\n")

	// Position: --> file:line:col
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Render(fmt.Sprintf("  --> %s:%d:%d", d.Pos.Filename, d.Pos.Line, d.Pos.Column)))
	b.WriteString("\n")

	// Source context with line numbers
	src, err := os.ReadFile(d.Pos.Filename)
	if err == nil {
		lines := strings.Split(string(src), "\n")
		if d.Pos.Line <= len(lines) {
			lineNum := fmt.Sprintf("%d", d.Pos.Line)

			// Line number gutter
			gutter := lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Render(lineNum)

			b.WriteString("   ")
			b.WriteString(gutter)
			b.WriteString(" | ")
			b.WriteString(lines[d.Pos.Line-1])
			b.WriteString("\n")

			// Underline caret
			spaces := strings.Repeat(" ", len(lineNum)+3+d.Pos.Column)
			caret := lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Render("^")

			b.WriteString(spaces)
			b.WriteString(caret)
			b.WriteString("\n")
		}
	}

	// Hint
	if d.Hint != "" {
		hint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Render("hint:")

		b.WriteString("   ")
		b.WriteString(hint)
		b.WriteString(" ")
		b.WriteString(d.Hint)
		b.WriteString("\n")
	}

	return b.String()
}
```

### Pattern 6: Cross-Platform Binary Download

**What:** Detect platform (GOOS/GOARCH), download appropriate binary, verify checksum, make executable.

**When to use:** `forge tool sync` for downloading templ, sqlc, atlas, tailwind binaries.

**Example:**
```go
// Source: https://www.digitalocean.com/community/tutorials/building-go-applications-for-different-operating-systems-and-architectures
package toolsync

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/cuu/grab/v3"
)

type Tool struct {
	Name     string
	Version  string
	Checksums map[string]string // platform -> checksum
}

func (t *Tool) Download(destDir string) error {
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	url := t.downloadURL(platform)

	// Download with progress
	client := grab.NewClient()
	req, _ := grab.NewRequest(destDir, url)

	// Set expected checksum
	expectedSum := t.Checksums[platform]
	req.SetChecksum(sha256.New(), []byte(expectedSum), true)

	resp := client.Do(req)

	// Monitor progress
	for !resp.IsComplete() {
		fmt.Printf("\rDownloading %s: %.2f%%",
			t.Name,
			100*resp.Progress())
	}

	if err := resp.Err(); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Make executable (Unix only)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(resp.Filename, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tool) downloadURL(platform string) string {
	// Construct URL based on tool and platform
	return fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s_%s",
		t.Name, t.Version, t.Name, platform)
}
```

### Pattern 7: Embedded Project Templates

**What:** Embed scaffolding templates in binary using embed.FS, render with text/template, write to disk.

**When to use:** `forge init` to create starter project structure.

**Example:**
```go
// Source: https://blog.jetbrains.com/go/2021/06/09/how-to-use-go-embed-in-go-1-16/
package scaffold

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templatesFS embed.FS

type ProjectData struct {
	Name    string
	Module  string
	GoVersion string
}

func CreateProject(projectPath string, data ProjectData) error {
	// Create directory structure
	dirs := []string{
		"resources/product",
		"gen/models",
		"gen/queries",
		"gen/atlas",
		"cmd",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return err
		}
		// Fix permissions (umask may restrict)
		os.Chmod(fullPath, 0755)
	}

	// Render templates
	templates := []struct {
		src  string
		dest string
	}{
		{"templates/forge.toml.tmpl", "forge.toml"},
		{"templates/main.go.tmpl", "main.go"},
		{"templates/resource_schema.go.tmpl", "resources/product/schema.go"},
	}

	for _, t := range templates {
		if err := renderTemplate(t.src, filepath.Join(projectPath, t.dest), data); err != nil {
			return err
		}
	}

	return nil
}

func renderTemplate(src, dest string, data ProjectData) error {
	// Read template from embedded FS
	content, err := templatesFS.ReadFile(src)
	if err != nil {
		return err
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(src)).Parse(string(content))
	if err != nil {
		return err
	}

	// Write to destination
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
```

### Anti-Patterns to Avoid

- **Executing schema code at parse time:** Never import or run resource schema files. Use go/ast static analysis only.
- **Using SkipObjectResolution mode:** Don't skip object resolution if you need Ident.Obj. However, for simple literal extraction, it's recommended.
- **Accessing flags in init():** Cobra flags aren't available until Execute() runs. Use PreRunE or RunE.
- **Required persistent flags:** Breaks built-in help/completion commands. Use optional persistent flags only.
- **Method chaining with error returns:** Go's error handling model conflicts with fluent APIs. Defer validation to final step.
- **Forgetting os.Chmod after MkdirAll:** umask restricts permissions. Always chmod after directory creation.
- **Assuming CallExpr.Fun is always Ident:** Can be SelectorExpr, FuncLit, etc. Use type assertion with checks.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Terminal colors/styling | Manual ANSI codes | Lipgloss | Color degradation, layout primitives, terminal detection |
| CLI argument parsing | Manual os.Args slicing | Cobra | Flag binding, help generation, subcommands, completions |
| HTTP download with progress | Manual http.Get + io.Copy | grab package | Progress tracking, resume, checksum verification, cancellation |
| TOML parsing | Custom parser | pelletier/go-toml/v2 | TOML 1.0 compliance, performance, encoding support |
| AST traversal | Manual node walking | ast.Inspect | Depth-first traversal, early termination, handles all node types |
| Position tracking | Manual line counting | go/token.FileSet | Handles position-altering comments (//line directives) |
| Constant folding | String literal checks | go/types constant evaluation | Handles const declarations, arithmetic expressions, type conversions |

**Key insight:** Go's stdlib excels at language tooling (go/ast, go/types, go/parser). Don't reimplement what the compiler already provides.

## Common Pitfalls

### Pitfall 1: Type Conversion vs Function Call Ambiguity

**What goes wrong:** AST parsers treat `int(x)` as either a type conversion or function call. Simple patterns like `schema.String("Name")` may parse incorrectly.

**Why it happens:** Go's grammar allows both forms with identical syntax. Without type information, AST can't disambiguate.

**How to avoid:** When extracting function calls, verify Fun is a SelectorExpr (package.Function) or Ident, not a type identifier. Use go/types if disambiguation is critical.

**Warning signs:** Unexpected parsing of field definitions, missing schema.Define() calls in extraction.

### Pitfall 2: BasicLit Values Include Quotes

**What goes wrong:** `ast.BasicLit.Value` for strings is `"hello"` (with quotes), not `hello`.

**Why it happens:** AST preserves source representation, including delimiters.

**How to avoid:** Use `strconv.Unquote()` for STRING and CHAR literals. For raw strings (backticks), handle `\r` removal.

**Warning signs:** Extracted field names have quotes, max length values are off by 2.

**Example:**
```go
// Wrong
name := lit.Value // "Title" (includes quotes)

// Right
name, err := strconv.Unquote(lit.Value) // Title
```

### Pitfall 3: Flag Values Not Available in init()

**What goes wrong:** Accessing Cobra flag values in init() returns defaults, not user-provided values.

**Why it happens:** Cobra parses flags during Execute(), which runs after all init() functions.

**How to avoid:** Access flags in PreRunE, RunE, or PostRunE. Bind flags with Viper in init(), access via Viper getters.

**Warning signs:** Flags always have default values, config file overrides don't work.

**Example:**
```go
// Wrong
func init() {
	rootCmd.Flags().StringP("output", "o", "text", "output format")
	format := viper.GetString("output") // Always "text"
}

// Right
var rootCmd = &cobra.Command{
	PreRunE: func(cmd *cobra.Command, args []string) error {
		format := viper.GetString("output") // Correct value
		return nil
	},
}
```

### Pitfall 4: Persistent Flag Access via Wrong Method

**What goes wrong:** `cmd.PersistentFlags().GetString("flag")` returns empty for flags defined on parent commands.

**Why it happens:** PersistentFlags() only returns flags defined on that specific command, not inherited ones.

**How to avoid:** Always use `cmd.Flags()` (not `cmd.PersistentFlags()`) to access flag values. It returns all applicable flags.

**Warning signs:** Flags work on root command but not subcommands, empty values for parent flags.

### Pitfall 5: Umask Restricts Directory Permissions

**What goes wrong:** `os.MkdirAll(path, 0755)` creates directories with 0700 or 0750 instead of 0755.

**Why it happens:** Unix umask (typically 0022) subtracts permissions from requested mode.

**How to avoid:** Call `os.Chmod()` after `os.MkdirAll()` to set exact permissions.

**Warning signs:** Generated directories not readable by group/other, permission denied errors.

**Example:**
```go
// Wrong
os.MkdirAll(path, 0755) // May create 0733 due to umask

// Right
os.MkdirAll(path, 0755)
os.Chmod(path, 0755) // Force exact permissions
```

### Pitfall 6: Confusing Constant Declaration with Constant Value

**What goes wrong:** Detecting `const MaxLen = 200` as constant, but `schema.MaxLen(MaxLen)` still flagged as dynamic.

**Why it happens:** AST shows an Ident node for `MaxLen`, not a BasicLit. Requires go/types to resolve.

**How to avoid:** Use go/types Info.Types map. Check if TypeAndValue.Value is non-nil (indicates compile-time constant).

**Warning signs:** All non-literal values rejected, even const declarations.

**Example:**
```go
// AST alone: MaxLen is an Ident, can't tell if constant
const MaxLen = 200
schema.String("Title").MaxLen(MaxLen) // Ident, not BasicLit

// With go/types: TypeAndValue.Value = 200 (constant)
tv := info.Types[identNode]
if tv.Value != nil {
	// It's a constant, extract tv.Value
}
```

### Pitfall 7: CallExpr.Ellipsis for Variadic Detection

**What goes wrong:** Assuming CallExpr.Args length indicates variadic call.

**Why it happens:** Variadic calls like `f(args...)` have `...` token stored in Ellipsis field, not reflected in Args length.

**How to avoid:** Check `CallExpr.Ellipsis.IsValid()` to detect variadic calls.

**Warning signs:** Misinterpreting variadic schema.Define() calls, incorrect argument extraction.

### Pitfall 8: Parser Mode Affects Object Resolution

**What goes wrong:** File.Scope, File.Unresolved, Ident.Obj are nil unexpectedly.

**Why it happens:** parser.SkipObjectResolution mode (recommended) disables these deprecated fields.

**How to avoid:** Don't rely on File.Scope or Ident.Obj. Use go/types for object resolution instead.

**Warning signs:** Nil pointer panics accessing Ident.Obj, empty File.Unresolved.

## Code Examples

Verified patterns from official sources:

### Parsing Go File and Walking AST
```go
// Source: https://pkg.go.dev/go/parser
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()

	// Parse Go file
	file, err := parser.ParseFile(fset, "schema.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Walk AST to find all function calls
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true // Continue traversal
		}

		// Extract function name
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			fmt.Printf("Found call: %s.%s\n", sel.X, sel.Sel.Name)
		}

		return true
	})
}
```

### Extracting Literal String Arguments
```go
// Source: https://github.com/golang/go/issues/39590
package main

import (
	"go/ast"
	"strconv"
)

func extractStringLiteral(arg ast.Expr) (string, bool) {
	lit, ok := arg.(*ast.BasicLit)
	if !ok {
		return "", false
	}

	// Unquote to remove surrounding quotes
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}

	return value, true
}
```

### Cobra Subcommand with Flags
```go
// Source: https://cobra.dev/docs/how-to-guides/working-with-commands/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new Forge project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "."
			if len(args) > 0 {
				name = args[0]
			}

			fmt.Printf("Initializing project: %s\n", name)
			fmt.Printf("Interactive mode: %v\n", interactive)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode")

	return cmd
}
```

### Lipgloss Styled Output
```go
// Source: https://github.com/charmbracelet/lipgloss
package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Define reusable styles
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)

	// Render styled output
	fmt.Println(successStyle.Render("✓") + " Project initialized")
	fmt.Println(errorStyle.Render("✗") + " Failed to parse schema")
}
```

### TOML Marshaling with Custom Indent
```go
// Source: https://pkg.go.dev/github.com/pelletier/go-toml/v2
package main

import (
	"bytes"
	"github.com/pelletier/go-toml/v2"
	"os"
)

type Config struct {
	Project struct {
		Name    string
		Version string
	}
	Database struct {
		URL string
	}
}

func main() {
	config := Config{}
	config.Project.Name = "myapp"
	config.Project.Version = "1.0.0"
	config.Database.URL = "postgres://localhost/db"

	// Create encoder with custom indentation
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentSymbol("  ") // 2 spaces
	enc.SetIndentTables(true)

	if err := enc.Encode(config); err != nil {
		panic(err)
	}

	// Write to file
	os.WriteFile("forge.toml", buf.Bytes(), 0644)
}
```

### Binary Download with Progress and Checksum
```go
// Source: https://pkg.go.dev/github.com/cuu/grab
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cuu/grab/v3"
)

func downloadBinary(url, dest, expectedChecksum string) error {
	client := grab.NewClient()
	req, _ := grab.NewRequest(dest, url)

	// Set expected checksum for automatic verification
	checksumBytes, _ := hex.DecodeString(expectedChecksum)
	req.SetChecksum(sha256.New(), checksumBytes, true)

	resp := client.Do(req)

	// Display progress
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			fmt.Printf("\r%.2f%% complete", 100*resp.Progress())
		case <-resp.Done:
			if err := resp.Err(); err != nil {
				return fmt.Errorf("download failed: %w", err)
			}
			fmt.Printf("\nDownload saved to %s\n", resp.Filename)
			return nil
		}
	}
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| BurntSushi/toml | pelletier/go-toml/v2 | 2021+ | BurntSushi unmaintained, v2 is 1.7-5.1x faster, TOML 1.0 spec |
| manual ANSI codes | Lipgloss | 2021+ | Declarative styling, automatic degradation, layout primitives |
| fatih/color | Lipgloss + Bubbletea | 2020-2021 | Color alone insufficient for modern TUIs, need layout + state |
| go/parser alone | go/parser + go/types | Always recommended | go/types required for constant folding, type resolution |
| File.Scope, Ident.Obj | go/types for objects | Deprecated 2013+ | Old fields unreliable, go/types is canonical |
| kingpin CLI | Kong or Cobra | 2019+ | Kingpin archived, Kong is successor, Cobra more popular |

**Deprecated/outdated:**
- **File.Scope, File.Unresolved, Ident.Obj:** Deprecated since Go 1.5. Use go/types for object resolution.
- **BurntSushi/toml:** Maintainer stepped away, use pelletier/go-toml/v2 for active maintenance.
- **parser.ParseDir for project-wide parsing:** Doesn't handle modules well. Use packages.Load from golang.org/x/tools/go/packages.

## Open Questions

1. **Error code numbering scheme**
   - What we know: Need codes like E001, E002 for error categories (parse error, validation error, etc.)
   - What's unclear: Exact numbering scheme (E001-E099 for parser, E100-E199 for validation?)
   - Recommendation: Start simple with sequential numbering, group by concern (E0xx parser, E1xx validation)

2. **Intermediate representation (IR) structure**
   - What we know: Need to represent schema.Define() calls with fields, relationships, modifiers
   - What's unclear: Exact struct hierarchy, whether to mirror AST or create domain model
   - Recommendation: Domain model (not AST mirror) — easier for code generation phase to consume

3. **Tool binary versioning strategy**
   - What we know: Need to download specific versions of templ, sqlc, atlas, tailwind
   - What's unclear: Pin versions in forge.toml? Allow version ranges? Auto-update?
   - Recommendation: Pin exact versions in forge.toml (reproducibility), manual upgrade via `forge tool sync --upgrade`

4. **Interactive mode for forge init**
   - What we know: User requested optional interactive mode (CONTEXT.md locked decisions don't mandate it)
   - What's unclear: Complexity vs value, questions to ask, fallback behavior
   - Recommendation: Defer to Phase 1.5 or Phase 2. Non-interactive mode sufficient for MVP.

## Sources

### Primary (HIGH confidence)
- [go/ast package](https://pkg.go.dev/go/ast) - AST node types and traversal
- [go/parser package](https://pkg.go.dev/go/parser) - Parsing Go source files
- [go/token package](https://pkg.go.dev/go/token) - Position tracking and file sets
- [go/types package](https://pkg.go.dev/go/types) - Type checking and constant folding
- [Cobra documentation](https://cobra.dev/) - CLI framework official docs
- [Cobra user guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md) - Detailed usage guide
- [Lipgloss GitHub](https://github.com/charmbracelet/lipgloss) - Terminal styling library
- [pelletier/go-toml/v2](https://pkg.go.dev/github.com/pelletier/go-toml/v2) - TOML library documentation
- [embed package](https://pkg.go.dev/embed) - Embedded files documentation
- [go/ast BasicLit documentation issue](https://github.com/golang/go/issues/39590) - strconv.Unquote requirement

### Secondary (MEDIUM confidence)
- [Understanding Go's AST | Leapcell](https://leapcell.io/blog/understanding-go-s-abstract-syntax-tree-ast) - AST overview
- [Rewriting Go with AST | Eli Bendersky](https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/) - AST transformation patterns
- [How to Use go:embed | GoLand Blog](https://blog.jetbrains.com/go/2021/06/09/how-to-use-go-embed-in-go-1-16/) - embed.FS tutorial
- [Cobra and Kong comparison](https://gist.github.com/andreykaipov/3006701e3ee57df397db827b18716b45) - CLI framework comparison
- [TOML library comparison issue](https://github.com/golang/dep/issues/789) - Performance benchmarks
- [Building Go for different platforms | DigitalOcean](https://www.digitalocean.com/community/tutorials/building-go-applications-for-different-operating-systems-and-architectures) - Cross-compilation guide
- [grab package documentation](https://pkg.go.dev/github.com/cuu/grab) - HTTP download with progress
- [briandowns/spinner](https://github.com/briandowns/spinner) - Terminal spinner library
- [charmbracelet/huh](https://github.com/charmbracelet/huh) - Interactive forms (optional)

### Tertiary (LOW confidence - validation needed)
- Various blog posts and Medium articles on Go AST, fluent APIs, and builder patterns - provided directional guidance but require verification against official sources for production use

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries widely used (Cobra in Kubernetes/Docker, Lipgloss in charm ecosystem, go/ast in stdlib)
- Architecture: HIGH - Patterns verified from official documentation and stdlib examples
- Pitfalls: HIGH - Documented in Go issues, official docs, and community consensus
- Binary download strategy: MEDIUM - Multiple approaches exist (aqua, asdf, custom), chose custom for zero-npm constraint
- Error formatting: MEDIUM - No direct Rust miette/codespan equivalent in Go, composed from lipgloss + manual formatting

**Research date:** 2026-02-16
**Valid until:** ~30 days (stdlib and Cobra stable, Charm libraries update frequently)
