# Architecture Research: Schema-Driven Go Code Generation Frameworks

**Domain:** Schema-driven code generation frameworks for Go
**Researched:** 2026-02-16
**Confidence:** HIGH

## Standard Architecture Pattern

Schema-driven code generation frameworks in Go follow a consistent architectural pattern with clear separation of concerns:

```
┌────────────────────────────────────────────────────────────────────┐
│                    User Schema Definition                          │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │  ent/schema/user.go  (imports only schema package)       │     │
│  │  schema.Define(...) calls define data model              │     │
│  └──────────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────────┘
                              ↓
┌────────────────────────────────────────────────────────────────────┐
│                         CLI Tool                                   │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │  Parser (go/ast, ANTLR, custom)                          │     │
│  │    - Loads schema files                                  │     │
│  │    - Builds intermediate representation                  │     │
│  │    - Validates schema semantics                          │     │
│  └────────────┬─────────────────────────────────────────────┘     │
│               │                                                    │
│  ┌────────────▼─────────────────────────────────────────────┐     │
│  │  Generator (templates, code emission)                    │     │
│  │    - Transforms IR to Go code                            │     │
│  │    - Produces type-safe APIs                             │     │
│  │    - Generates supporting code (tests, migrations)       │     │
│  └──────────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────────┘
                              ↓
┌────────────────────────────────────────────────────────────────────┐
│                      Generated Code                                │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │  gen/user.go, gen/user_query.go, gen/client.go          │     │
│  │    - Imports runtime library                             │     │
│  │    - Type-safe model types                               │     │
│  │    - Builder APIs                                        │     │
│  │    - Supporting utilities                                │     │
│  └──────────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────────┘
                              ↓
┌────────────────────────────────────────────────────────────────────┐
│                      Runtime Library                               │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │  Client, Tx, hooks, predicates, database drivers         │     │
│  │    - Core abstractions generated code depends on         │     │
│  │    - Database connectivity                               │     │
│  │    - Query execution                                     │     │
│  └──────────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────────┘
                              ↓
┌────────────────────────────────────────────────────────────────────┐
│                   Application Code                                 │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │  Imports generated code + runtime library                │     │
│  │  Uses type-safe APIs for business logic                  │     │
│  └──────────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **Schema Package** | Provides schema definition API that user code imports | Minimal API surface, no dependencies on generated code |
| **Parser** | Reads schema definitions, builds intermediate representation (IR) | go/ast for Go schemas, ANTLR for SQL, custom parsers for config formats |
| **Generator** | Transforms IR into Go source code | Go templates, AST builders like `dave/jennifer`, or direct string emission |
| **Generated Code** | Type-safe APIs for working with schema entities | Imports runtime library, never imported by schema definitions |
| **Runtime Library** | Core abstractions and utilities used by generated code | Database drivers, query builders, hooks, middleware |
| **CLI Tool** | Orchestrates parsing and generation | Cobra/flag-based CLI that invokes parser → generator pipeline |

## Recommended Project Structure

Based on analysis of Ent, SQLC, protobuf, and OpenAPI generators:

```
forge/
├── cmd/
│   └── forge/              # CLI entry point
│       └── main.go         # Cobra commands: generate, scaffold, migrate
├── schema/                 # User-facing schema definition API
│   ├── schema.go           # Define() and field builder APIs
│   ├── field.go            # Field type definitions
│   ├── edge.go             # Relationship definitions
│   └── validation.go       # Validation rule builders
├── internal/
│   ├── parser/             # Schema parsing (go/ast based)
│   │   ├── loader.go       # Load Go files
│   │   ├── extract.go      # Extract schema.Define() calls
│   │   └── ir.go           # Intermediate representation
│   ├── generator/          # Code generation
│   │   ├── model.go        # Generate model types
│   │   ├── query.go        # Generate query builders
│   │   ├── validation.go   # Generate validators
│   │   ├── atlas.go        # Generate Atlas HCL
│   │   ├── api.go          # Generate Huma API types
│   │   ├── action.go       # Generate action interfaces
│   │   ├── factory.go      # Generate test factories
│   │   └── template/       # Go templates for code emission
│   ├── scaffolder/         # One-time scaffolding
│   │   ├── view.go         # Scaffold Templ views
│   │   ├── handler.go      # Scaffold HTML handlers
│   │   └── hook.go         # Scaffold hook implementations
│   └── validator/          # Schema validation
│       └── validate.go     # Semantic validation of schema IR
├── runtime/                # Runtime library (separate module)
│   ├── app.go              # forge.App
│   ├── middleware.go       # Standard middleware
│   ├── context.go          # Tenant/auth context helpers
│   ├── errors.go           # Error handling
│   ├── sse.go              # Server-sent events
│   └── forms.go            # Form primitives
└── examples/               # Example applications
    └── blog/
        ├── schema/
        │   └── schema.go   # User schema definitions
        ├── gen/            # Generated code (gitignored during dev)
        └── resources/      # Scaffolded code (committed)
```

### Structure Rationale

- **`schema/` as separate package**: Enables the bootstrapping constraint where schema files import only `github.com/forgego/forge/schema`, never generated code, preventing circular dependencies.

- **`internal/parser/` and `internal/generator/` separation**: Parser builds intermediate representation (IR) independent of output format; generator transforms IR to code. This allows multiple generators (Go, TypeScript, docs) from same IR.

- **`runtime/` as separate module**: Runtime library has independent versioning and can be imported by both generated code and user code without pulling in CLI/parser/generator dependencies.

- **Generated code in `gen/`**: Standard convention (Ent, SQLC) places all generated code in a dedicated directory, making it easy to `.gitignore` or identify auto-generated files.

- **Scaffolded code in `resources/`**: User-editable code generated once lives separately from regenerated code in `gen/`, preventing accidental overwrites.

## Architectural Patterns

### Pattern 1: Bootstrapping Constraint via Import Isolation

**What:** Schema definitions must never import generated code, only the schema definition API package.

**When to use:** All schema-driven frameworks that use Go source as schema definition language.

**Trade-offs:**
- **Pro:** Prevents circular dependency (schema → codegen → generated code → schema)
- **Pro:** Schema files can be parsed without compiling generated code first
- **Con:** Schema can't reference generated types directly (must use strings/reflection)

**Example:**
```go
// schema/user.go - CORRECT
package schema

import "github.com/forgego/forge/schema"

var User = schema.Define("User", func(s *schema.Schema) {
    s.Field("email").String().Unique()
    s.BelongsTo("Organization") // String reference, not type
})

// schema/user.go - WRONG (circular dependency)
package schema

import (
    "github.com/forgego/forge/schema"
    "myapp/gen" // NEVER do this
)

var User = schema.Define("User", func(s *schema.Schema) {
    s.Field("org").Type(gen.Organization{}) // Circular!
})
```

### Pattern 2: Two-Phase Generation (Generate vs Scaffold)

**What:** Separate always-regenerated code from edit-once scaffolded code.

**When to use:** When framework needs to provide customizable implementations alongside type-safe generated APIs.

**Trade-offs:**
- **Pro:** Generated code stays in sync with schema; scaffolded code is user-owned
- **Pro:** Clear separation prevents accidental overwrites
- **Con:** Two directories to maintain (`gen/` and `resources/`)

**Example:**
```go
// gen/user_action.go - ALWAYS REGENERATED
package gen

type UserAction interface {
    Create(ctx context.Context, params UserCreateParams) (*User, error)
    Update(ctx context.Context, id int, params UserUpdateParams) (*User, error)
}

// gen/user_action_default.go - ALWAYS REGENERATED
package gen

type DefaultUserAction struct {
    db *DB
}

func (a *DefaultUserAction) Create(ctx context.Context, params UserCreateParams) (*User, error) {
    return a.db.User.Create().SetEmail(params.Email).Save(ctx)
}

// resources/users/hooks.go - SCAFFOLDED ONCE, USER EDITS
package users

func BeforeUserCreate(ctx context.Context, params gen.UserCreateParams) error {
    // Custom validation logic - user adds this
    return nil
}
```

### Pattern 3: Intermediate Representation (IR) Decoupling

**What:** Parser produces language-agnostic IR; generator transforms IR to target language.

**When to use:** When supporting multiple output formats or allowing extensibility.

**Trade-offs:**
- **Pro:** Single parser supports multiple generators (Go, TypeScript, docs)
- **Pro:** IR can be validated independently of output format
- **Pro:** Extensible via custom generators consuming IR
- **Con:** Additional abstraction layer adds complexity

**Example:**
```go
// internal/parser/ir.go
type Schema struct {
    Name   string
    Fields []Field
    Edges  []Edge
}

type Field struct {
    Name       string
    Type       FieldType
    Unique     bool
    Validators []Validator
}

// Parser produces IR
func Parse(files []*ast.File) ([]Schema, error) {
    // AST → IR transformation
}

// Multiple generators consume same IR
func GenerateGoModels(schemas []Schema) ([]byte, error)
func GenerateTypeScript(schemas []Schema) ([]byte, error)
func GenerateDocs(schemas []Schema) ([]byte, error)
```

### Pattern 4: Runtime Library as Stable Foundation

**What:** Generated code depends on runtime library; runtime has stable API.

**When to use:** Always. Generated code is disposable; runtime must be stable.

**Trade-offs:**
- **Pro:** Generated code can change freely without breaking user imports
- **Pro:** Runtime version can be pinned independently
- **Con:** Must maintain backward compatibility in runtime API

**Example:**
```go
// runtime/app.go - STABLE API
package runtime

type App struct {
    db     *sql.DB
    config Config
}

func New(config Config) (*App, error) {
    // Stable constructor
}

func (a *App) Handler() http.Handler {
    // Stable API
}

// gen/client.go - GENERATED, UNSTABLE
package gen

import "github.com/forgego/forge/runtime"

type Client struct {
    app *runtime.App // Depends on stable runtime
    User UserClient
    Post PostClient
}

// User code imports generated + runtime
import (
    "myapp/gen"
    "github.com/forgego/forge/runtime"
)
```

### Pattern 5: CLI as Orchestrator, Not Builder

**What:** CLI invokes code generation but doesn't integrate with `go build`. Use `//go:generate` for build integration.

**When to use:** Following Go conventions (Ent, SQLC, protoc-gen-go all use this pattern).

**Trade-offs:**
- **Pro:** Explicit generation step makes regeneration visible
- **Pro:** Works with any build system
- **Pro:** Generated code can be committed and reviewed
- **Con:** Must remember to run generation after schema changes

**Example:**
```go
// schema/generate.go
package schema

//go:generate forge generate

// Developer workflow:
// 1. Edit schema/user.go
// 2. Run: go generate ./schema
// 3. Generated code appears in gen/
// 4. Run: go build
```

## Data Flow

### Schema Definition → Generated Code Flow

```
1. Developer writes schema:
   schema/user.go contains schema.Define("User", ...)

2. Developer runs generation:
   go generate ./schema  (invokes //go:generate forge generate)

3. CLI loads schemas:
   forge generate
     → parser.Load("./schema")
     → parser.ParseFile() uses go/ast
     → Extract schema.Define() call expressions

4. Parser builds IR:
   AST → Intermediate Representation
     → Validate semantics (unique constraints, relationships)
     → Resolve cross-schema references

5. Generator produces code:
   IR → Generate Go models (gen/user.go)
      → Generate query builders (gen/user_query.go)
      → Generate validators (gen/user_validate.go)
      → Generate API types (gen/user_api.go)
      → Generate action interfaces (gen/user_action.go)
      → Generate test factories (gen/user_factory.go)
      → Generate Atlas HCL (gen/atlas.hcl)

6. Application imports generated code:
   app/main.go imports myapp/gen
     → gen/ imports github.com/forgego/forge/runtime
     → Application never imports schema package
```

### Runtime Request Flow

```
HTTP Request
    ↓
forge.App middleware stack
    ↓
HTML Handler (resources/users/handler.go)
    ↓
Action Layer (gen.DefaultUserAction)
    ↓
Generated Query Builder (gen.UserQuery)
    ↓
Bob Query Builder (underlying SQL)
    ↓
Database Driver
    ↓
Response
```

### Build Order Dependencies

```
1. schema package (no dependencies on generated code)
2. runtime library (independent, can be pre-built)
3. go generate ./schema (parser reads AST, emits gen/)
4. gen/ package (depends on runtime, imports it)
5. resources/ package (depends on gen/, imports action interfaces)
6. application code (depends on gen/ and resources/)
```

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **Bob Query Builder** | Generated code uses Bob mods | Import `github.com/stephenafamo/bob`, generate type-safe query mods |
| **Atlas Migrations** | Generate Atlas HCL from schema | Parser produces `.hcl` alongside Go code, `atlas migrate diff` uses it |
| **Huma API Framework** | Generate Huma-compatible types | Struct tags, validation rules generated from schema |
| **Templ Views** | Scaffold once, reference generated types | Views import `gen/` for type safety |
| **Database Drivers** | Runtime library wraps `database/sql` | Bob handles driver abstraction |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| schema → CLI | File I/O (go/ast parsing) | Schema never calls CLI; CLI reads schema files |
| CLI → gen/ | File writes (code emission) | One-way: CLI produces gen/, gen/ never calls CLI |
| gen/ → runtime | Go imports | Generated code imports stable runtime library |
| resources/ → gen/ | Go imports | Scaffolded code imports action interfaces |
| app → gen/ + runtime | Go imports | Application imports both generated code and runtime |

## Anti-Patterns to Avoid

### Anti-Pattern 1: Schema Imports Generated Code

**What people do:** Try to reference generated types in schema definitions.

**Why it's wrong:** Creates circular dependency that prevents parsing schema before code generation.

**Do this instead:** Use string-based references in schema; generated code provides the types.

```go
// WRONG
import "myapp/gen"
s.BelongsTo(gen.Organization{}) // Can't work - gen doesn't exist yet!

// CORRECT
s.BelongsTo("Organization") // String reference, resolved at codegen time
```

### Anti-Pattern 2: Editing Generated Code

**What people do:** Modify files in `gen/` directory to add custom logic.

**Why it's wrong:** Next generation run overwrites changes without warning.

**Do this instead:** Use extension points (hooks, custom actions) in `resources/` that import and wrap generated interfaces.

```go
// WRONG: Edit gen/user_action_default.go
func (a *DefaultUserAction) Create(...) {
    // Custom logic here - LOST on next generation!
}

// CORRECT: Extend in resources/users/hooks.go
func BeforeUserCreate(ctx context.Context, params gen.UserCreateParams) error {
    // Custom logic here - preserved across generations
}
```

### Anti-Pattern 3: Compile-and-Execute Schema Parsing

**What people do:** Try to compile and run schema files to extract definitions.

**Why it's wrong:** Requires generated code to exist before parsing, creates bootstrapping problem.

**Do this instead:** Use go/ast to parse schema files statically without compiling.

```go
// WRONG
// 1. Compile schema package
// 2. Run it and call schema.GetAll()
// Problem: Can't compile if schema imports gen/, and gen/ doesn't exist yet!

// CORRECT
// 1. Parse schema/*.go with go/ast
// 2. Extract schema.Define() calls from AST
// 3. No compilation needed - works even if gen/ doesn't exist
```

### Anti-Pattern 4: Tight Coupling Between CLI and Runtime

**What people do:** Put CLI, generator, and runtime in single module.

**Why it's wrong:** User applications must depend on heavy CLI/parser dependencies just to use runtime.

**Do this instead:** Separate runtime into independent module with minimal dependencies.

```go
// WRONG go.mod
module github.com/forgego/forge
require (
    go/ast           // CLI needs this
    text/template    // Generator needs this
    database/sql     // Runtime needs this
)
// User imports runtime, gets all CLI dependencies too!

// CORRECT: Two modules
// github.com/forgego/forge (CLI + generator)
// github.com/forgego/forge/runtime (minimal dependencies)
```

### Anti-Pattern 5: Magic Globals in Schema Package

**What people do:** Use `init()` functions or global registration in schema package.

**Why it's wrong:** Forces schema package to execute at runtime; prevents static analysis.

**Do this instead:** Pure declarations that CLI parses via AST.

```go
// WRONG
func init() {
    schema.Register(User) // Requires compiling and running schema package
}

// CORRECT
var User = schema.Define("User", ...) // Pure declaration, parseable via AST
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| **1-5 schemas** | Simple flat structure in `schema/` is fine. Single `schema.go` file works. |
| **5-20 schemas** | Split into multiple files (`schema/user.go`, `schema/post.go`). Parser loads all `*.go` in directory. |
| **20-50 schemas** | Organize into subdirectories by domain (`schema/auth/`, `schema/content/`). Parser walks tree. Consider splitting runtime into domain-specific packages. |
| **50+ schemas** | May hit codegen performance limits. Consider: (1) Incremental generation (only changed schemas), (2) Splitting into multiple modules, (3) Plugin architecture for custom generators. |

### Scaling Priorities

1. **First bottleneck: Code generation time**
   - **Symptom:** `go generate` takes >10 seconds
   - **Fix:** Implement incremental generation (hash schema files, only regenerate changed)
   - **Fix:** Parallelize generation (generate User and Post in separate goroutines)

2. **Second bottleneck: Generated code size**
   - **Symptom:** `gen/` directory is >50MB, slow to compile
   - **Fix:** Split generated code into per-schema files (already planned: `gen/user.go`, `gen/post.go`)
   - **Fix:** Use build tags to conditionally compile subsets

3. **Third bottleneck: Circular imports in generated code**
   - **Symptom:** User → Post → User relationships create import cycles
   - **Fix:** Generate interfaces in separate package (`gen/types/`) that `gen/user` and `gen/post` both import

## Forge-Specific Architecture Decisions

Based on the project context, Forge should implement:

### 1. Package Structure
```
github.com/forgego/forge
  ├── schema/           # User-facing API (stable)
  ├── internal/parser/  # AST-based parser
  ├── internal/gen/     # Code generators
  └── cmd/forge/        # CLI

github.com/forgego/forge/runtime  # Separate module
  └── runtime/          # Runtime library
```

### 2. Schema Parsing Strategy
- **Use go/ast** (not compile-and-execute) to parse `schema.Define()` calls
- Extract call expressions: `schema.Define("User", func(s *schema.Schema) { ... })`
- Build IR from AST without executing code

### 3. Code Generation Outputs
Generated in `gen/`:
- Models (Go structs)
- Bob query builder mods
- Validation functions
- Atlas HCL schema
- Huma API structs
- Action interfaces
- Test factories

Scaffolded once in `resources/`:
- Templ views
- HTML handlers
- Hook implementations

### 4. Action Layer Pattern
```go
// gen/user_action.go - GENERATED
type UserAction interface {
    Create(ctx context.Context, params UserCreateParams) (*User, error)
}

type DefaultUserAction struct { /* generated implementation */ }

// gen/user_handler.go - GENERATED
type UserHandler struct {
    action UserAction // Injectable
}

// resources/users/custom_action.go - USER EXTENDS
type CustomUserAction struct {
    gen.DefaultUserAction
}

func (a *CustomUserAction) Create(ctx context.Context, params gen.UserCreateParams) (*User, error) {
    // Custom logic before/after default
}

// app/main.go - USER WIRES UP
app.RegisterAction(&users.CustomUserAction{})
```

## Sources

### High Confidence (Official Documentation)
- [Ent Code Generation Documentation](https://entgo.io/docs/code-gen/)
- [Go AST Package Documentation](https://pkg.go.dev/go/ast)
- [Go Generate Blog Post](https://go.dev/blog/generate)
- [SQLC GitHub Repository](https://github.com/sqlc-dev/sqlc)
- [Protocol Buffers Go Generated Code Guide](https://protobuf.dev/reference/go/go-generated/)
- [gRPC Generated Code Reference](https://grpc.io/docs/languages/go/generated-code/)

### Medium Confidence (Ecosystem Analysis)
- [Ent ORM: Separating Schema Definitions from Generated Code](https://medium.com/@duckdevv/ent-orm-separating-schema-definitions-from-generated-code-3ed4d2d54cc0)
- [Code Generation From the AST - Gopher Academy](https://blog.gopheracademy.com/code-generation-from-the-ast/)
- [A Comprehensive Guide to go generate - Eli Bendersky](https://eli.thegreenplace.net/2021/a-comprehensive-guide-to-go-generate/)
- [Go Project Structure: Clean Architecture Patterns](https://dasroot.net/posts/2026/01/go-project-structure-clean-architecture/)
- [SQLC Type-Safe Database Access Tutorial](https://oneuptime.com/blog/post/2026-01-07-go-sqlc-type-safe-database/view)
- [Managing Tool Dependencies in Go 1.24+](https://medium.com/@yuseferi/managing-tool-dependencies-in-go-1-24-a-deep-dive-feb2c9e07fe9)
- [Atlas Schema Migration Documentation](https://atlasgo.io/)
- [oapi-codegen OpenAPI Generator](https://github.com/oapi-codegen/oapi-codegen)

### Build Order and Dependencies
- [Go Clean Architecture Patterns](https://threedots.tech/post/introducing-clean-architecture/)
- [Structuring Go CLI Applications](https://www.bytesizego.com/blog/structure-go-cli-app)
- [Using go generate to Reduce Boilerplate](https://blog.logrocket.com/using-go-generate-reduce-boilerplate-code/)

---

*Architecture research for: Schema-driven Go web framework code generation*
*Researched: 2026-02-16*
*Primary sources: Ent, SQLC, protobuf/gRPC, OpenAPI generators, Go standard library documentation*
