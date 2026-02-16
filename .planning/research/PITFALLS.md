# Domain Pitfalls: Schema-Driven Go Web Framework

**Domain:** Schema-driven Go code generation framework
**Researched:** 2026-02-16
**Confidence:** MEDIUM-HIGH

## Critical Pitfalls

### Pitfall 1: Circular Import Bootstrapping (The "Zero Gen Imports" Problem)

**What goes wrong:**
Generated code in `gen/` packages imports schema packages, and schema packages import generated types, creating circular import dependency that Go explicitly disallows. This is the most common reason schema-driven Go frameworks fail or require ugly workarounds.

**Why it happens:**
Developers naturally want schema definitions (source of truth) and generated code to reference each other. Schema structs need generated client methods, generated clients need schema types. In languages with circular import tolerance, this works. In Go, it breaks immediately.

**How to avoid:**
- **Enforce zero gen/ imports in schema packages** — Schema packages must be dependency-free beyond stdlib and third-party libraries
- **Use interface abstraction layers** — Generated code implements interfaces defined in a shared `contracts/` package that both schema and gen packages can import
- **Flatten generated output** — Instead of `gen/users`, `gen/posts` with cross-references, use single `gen/models` package (Ent's approach)
- **Code generation in compilation order** — Generate types first (importable), then clients that use those types

**Warning signs:**
- `import cycle not allowed` errors during `go generate`
- Developers adding `// +build ignore` tags to "temporarily" fix imports
- Generated code that won't compile without manual editing
- Schema files importing anything from `gen/` or `internal/generated`

**Phase to address:**
Phase 1 (Foundation) — This must be solved architecturally before any generation. Test with a minimal two-entity schema with relationship between them.

---

### Pitfall 2: go/ast Comment Preservation Failures

**What goes wrong:**
Generated code loses all comments from schema definitions, or re-arranging AST nodes breaks comment positioning, producing unreadable/undocumented generated code. Users expect schema comments to propagate to OpenAPI docs, database migrations, and generated Go code.

**Why it happens:**
The go/ast package wasn't designed for source manipulation. Comments are stored by byte offset, not attached to nodes. When you rearrange nodes or insert new ones, byte offsets become invalid and comments disappear or attach to wrong code.

**How to avoid:**
- **Use `dave/dst` (Decorated Syntax Tree)** instead of raw go/ast for code manipulation — dst preserves comment associations through transformations
- **Parse with `parser.ParseComments` mode** — Standard go/ast discards comments by default
- **Use templates for code generation, go/ast for parsing** — Templates preserve comments you write, go/ast reads source schemas
- **Document in schema, render in templates** — Extract comment text during parsing, inject into template variables

**Warning signs:**
- Generated files with zero comments despite documented schemas
- Comments appearing above wrong functions/fields
- OpenAPI specs missing descriptions that exist in schema code
- Developers manually copying comments into generated files

**Phase to address:**
Phase 1 (Foundation) — Comment preservation is table stakes for schema-driven frameworks. Test early with heavily commented schema.

---

### Pitfall 3: Runtime Reflection vs Compile-Time Generation Trade-offs

**What goes wrong:**
Using runtime reflection (like Dig DI framework) for dependency resolution causes runtime panics for missing dependencies, cryptic 5-frame-deep stack traces, and no way to detect problems until code executes in production.

**Why it happens:**
Reflection-based frameworks feel magical ("just register providers and go"), hiding complexity. Developers discover errors too late because type system can't validate reflection calls.

**How to avoid:**
- **Prefer compile-time code generation** (like Wire) over runtime reflection frameworks
- **Generate explicit constructors** rather than using `reflect.New()` or similar
- **Use go/types for type checking during generation** — Validate types at codegen time, not runtime
- **Make generated code boring and obvious** — Explicit `NewUserService(db, logger)` beats `container.Resolve(&UserService{})`

**Warning signs:**
- Dependency injection errors only appearing in integration tests or production
- Stack traces pointing into framework internals, not your code
- "Cannot find provider for X" errors at runtime
- Type assertions with `panic` instead of compile errors

**Phase to address:**
Phase 2 (Code Generation Engine) — Choose generation over reflection early. This is a one-way door decision.

---

### Pitfall 4: Template Debugging Hell

**What goes wrong:**
Code generation templates become 500+ line monsters with nested conditionals, making errors nearly impossible to debug. Template syntax errors manifest as cryptic "unexpected EOF" messages. Generated code has syntax errors but templates report "success".

**Why it happens:**
`text/template` has minimal error reporting and no IDE support. Developers incrementally add logic until templates are unmaintainable. Template execution errors don't fail loudly—partial output may be written before error occurs.

**How to avoid:**
- **One template per output file type** — Separate `model.go.tmpl`, `client.go.tmpl`, `migration.sql.tmpl`
- **Keep logic in Go, not templates** — Templates receive pre-processed data structures, not raw schema
- **Run `go/format` on all generated code** — This catches syntax errors immediately: `formatted, err := format.Source(generated)`
- **Generate small, composable functions** — Not monolithic files
- **Template helper functions** — Register custom functions for complex logic: `template.Funcs(template.FuncMap{"toCamelCase": ...})`
- **Write template tests** — Test templates with fixture data before using in real generation

**Warning signs:**
- Template files over 200 lines
- Template syntax like `{{if .Field}}{{if .Field.Nested}}{{if .Field.Nested.Deep}}`
- Generated code that won't compile but template execution "succeeds"
- No automated tests for templates
- Developers editing generated code instead of fixing templates

**Phase to address:**
Phase 2 (Code Generation Engine) — Establish template architecture early. Refactoring large templates later is painful.

---

### Pitfall 5: Multi-Library Integration Complexity Explosion

**What goes wrong:**
Integrating 6+ libraries (Huma, Bob, Atlas, River, Datastar, Templ, pgx) creates 15+ integration points where version incompatibilities, conflicting conventions, and duplicated concepts cause maintenance nightmare. One library upgrade breaks three others.

**Why it happens:**
Each library has opinions about database access, error handling, context usage, logging. These opinions conflict. Example: Huma wants `context.Context` first parameter, River wants `JobArgs` struct, Bob wants query builder pattern, Atlas wants declarative schemas.

**How to avoid:**
- **Adapter layer between framework and libraries** — Don't expose library types directly in generated code
- **Pin exact versions, not version ranges** — `go.mod` should specify exact commits: `github.com/riverqueue/river v0.12.1`
- **Integration tests for every library pair** — Test Huma+Bob, Huma+River, Bob+Atlas, etc.
- **Facade pattern for database access** — Single `db.Query()` that internally uses pgx, Bob, or raw SQL
- **Document integration constraints** — "River requires pgx v5.7+, Atlas migration tool conflicts with GORM"

**Warning signs:**
- Dependency hell during `go get -u`
- Different parts of codebase using different database libraries for same task
- Generated code exposing `*sql.DB`, `pgx.Conn`, `bob.Query` in same function signature
- Copy-pasted integration code across generators
- Version pins with comments like "DO NOT UPDATE - breaks everything"

**Phase to address:**
Phase 1 (Foundation) — Design adapter layer before implementing generators. Phase 3+ re-evaluates as libraries evolve.

---

### Pitfall 6: Action Layer Duplication (API vs HTML Handler Split)

**What goes wrong:**
Business logic duplicated across API handlers (JSON responses) and HTML handlers (template rendering). Update validation logic in API → forget to update HTML handler → inconsistent behavior. This is the "shared action layer" problem.

**Why it happens:**
API handlers return `application/json`, HTML handlers return `text/html`. Developers write separate handlers because response formats differ, then copy business logic instead of extracting it.

**How to avoid:**
- **Shared action layer below handlers** — Extract `CreateUser(ctx, input) (*User, error)` used by both API and HTML handlers
- **Handlers only map responses** — API handler: `json.Marshal(user)`, HTML handler: `templ.Render(userTemplate, user)`
- **Generate action layer, not handlers** — Code generation produces `actions/users.go`, handlers call it
- **Single validation pass** — Validation happens in action layer, not in each handler
- **Content negotiation pattern** — One handler with `Accept` header checking, branching only at response time

**Warning signs:**
- Identical validation code in `api/handlers.go` and `web/handlers.go`
- Tests for API passing but HTML version broken (or vice versa)
- "We only test the API, HTML is too hard to test" comments
- PRs that update one handler but not the other
- Six-month-old bug reports: "works in API but not in web UI"

**Phase to address:**
Phase 3 (Action Layer) — This is the core abstraction. Must be solved before generating both API and HTML handlers.

---

### Pitfall 7: go/types vs go/ast Confusion

**What goes wrong:**
Using `go/ast` for type resolution instead of `go/types`, resulting in incorrect type inference, unresolved imports, and phantom type errors in generated code. Developers waste days debugging "why doesn't ast.Object show the type?"

**Why it happens:**
`go/ast` is a syntax tree (structure). `go/types` is a type checker (semantics). They're different tools. Old examples and tutorials use `ast.Object` which was deprecated, leading developers astray.

**How to avoid:**
- **Use go/types for type checking** — If you need to know "is this field a string or *string?", use go/types
- **Use go/ast only for structure** — If you need to know "how many fields does this struct have?", use go/ast
- **Set parser.SkipObjectResolution flag** — Disable broken ast.Object resolution, use go/types instead
- **golang.org/x/tools/go/packages** — Use this instead of manual ast + types setup

**Warning signs:**
- Code checking `field.Type.(*ast.Ident).Name` for type names
- Type resolution failing for imported types
- Generated code with wrong types for embedded structs
- Comments in code: "TODO: why doesn't this work for pointers?"
- Using deprecated `ast.Object`

**Phase to address:**
Phase 2 (Code Generation Engine) — Solve during parser implementation. Changing mid-development requires rewriting parser.

---

### Pitfall 8: Generated Code Version Control Strategy

**What goes wrong:**
Team commits generated code to git → merge conflicts on every schema change → developers manually resolve conflicts in generated files → manual edits overwritten on next generation → chaos.

**Why it happens:**
No clear policy on whether generated code is source or artifact. Committing it helps CI/CD, not committing it causes "works on my machine" when someone forgets `go generate`.

**How to avoid:**
- **Commit generated code with clear ownership** — Add header comment: `// Code generated by forge. DO NOT EDIT.`
- **CI verification check** — Run `go generate && git diff --exit-code gen/` to ensure generated code is up-to-date
- **Regenerate on every build** — `make build` runs `go generate` first, treating gen/ as build artifact
- **Use //go:generate directives** — Make generation automatic and discoverable
- **Git attributes for generated files** — `.gitattributes`: `gen/* linguist-generated=true` (excludes from diffs)
- **Pre-commit hook option** — Auto-regenerate before commit

**Warning signs:**
- Merge conflicts in `gen/` directory
- Generated files with manual edits
- "Did you run go generate?" in every PR review
- Generated code out of sync between developer machines
- Production deployment with stale generated code

**Phase to address:**
Phase 1 (Foundation) — Establish policy before first generated code. Changing later causes git history chaos.

---

### Pitfall 9: go/format Doesn't Fix Everything

**What goes wrong:**
Generated code passes `go/format` but has semantic errors (undefined variables, wrong types, missing imports). Developers assume formatted = correct.

**Why it happens:**
`go/format` only fixes whitespace and formatting. It doesn't type-check. Templates can generate syntactically valid but semantically broken code.

**How to avoid:**
- **Run go/types type checker after generation** — Use `golang.org/x/tools/go/packages` to load and type-check
- **Compile generated code as test step** — Generation pipeline: template → format → compile → test
- **Unit tests for generated code** — Not just "does it compile" but "does CreateUser() work"
- **golang.org/x/tools/go/ast/astutil for import management** — Automatically add/remove imports instead of manual template import blocks

**Warning signs:**
- Generated code compiles but panics at runtime
- Import blocks with unused imports
- `variable not used` errors after generation
- Tests only check if generation succeeded, not if generated code works

**Phase to address:**
Phase 2 (Code Generation Engine) — Build verification into generation pipeline from day one.

---

### Pitfall 10: Forgetting go generate After Schema Changes

**What goes wrong:**
Developer updates schema → forgets `go generate` → commits old generated code → CI fails or worse, deploys broken code → production incident.

**Why it happens:**
`go generate` is manual step, easy to forget. Not part of muscle memory like `go build` or `go test`.

**How to avoid:**
- **Make `go generate` part of `go build`** — Use `//go:generate` directives so `go generate ./...` is obvious
- **Pre-commit git hook** — Auto-run `go generate` before allowing commit
- **CI check for stale generated code** — `go generate && git diff --exit-code || exit 1`
- **Watch mode during development** — File watcher that runs `go generate` on schema file changes
- **Make target dependencies** — `make build` depends on `make generate`

**Warning signs:**
- "Did you run go generate?" in code review comments
- Deployed code that doesn't match schema
- Generated methods missing for new schema fields
- Production bugs that don't reproduce locally (different generated code versions)

**Phase to address:**
Phase 1 (Foundation) — Automate as much as possible. This is a process problem, not a technical one.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Using string templates instead of go/ast construction | Faster to write, easier to read | Hard to validate, fragile to Go syntax changes, no comment preservation | Early prototyping only (Phases 0-1) |
| Runtime reflection for DI instead of code generation | Less code to write, feels magical | Runtime errors, performance cost, opaque failures | Never — Go's type system is your friend |
| Copying business logic between API/HTML handlers | Quick implementation | Duplicate bugs, inconsistent behavior, double maintenance | Never — extract shared action layer |
| Committing generated code without verification | Skips CI complexity | Stale generated code in production | Never — always verify gen/ matches schemas |
| Using go/ast.Object instead of go/types | Simpler API, fewer imports | Incorrect type resolution, deprecated API | Never — ast.Object is formally deprecated |
| Skipping template tests | Faster iteration | Template bugs caught late, production errors | Early experimentation (Phase 1), never in production |
| Monolithic 1000-line templates | Fewer files to manage | Unmaintainable, hard to debug, impossible to test | Never — split templates by concern |
| Direct library type exposure in generated code | Less adapter code | Tight coupling, upgrade brittleness | Early prototyping (Phase 1-2) |

---

## Integration Gotchas

Common mistakes when connecting to specific libraries.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| **Huma + pgx** | Passing `*sql.DB` to Huma handlers expecting pgx-specific features | Use adapter interface, convert in handler: `pgxPool := db.(*pgxpool.Pool)` |
| **Bob + Atlas** | Generating Atlas migrations from Bob query builder output | Atlas uses declarative schema, Bob is query runtime — keep separate, Atlas owns schema truth |
| **River + pgx** | Creating separate database connection for River workers | River requires same pgx pool for transactional enqueueing — share `*pgxpool.Pool` |
| **Datastar + Templ** | Rendering Templ components without SSE content-type header | Use `datastar.NewSSE(w, r)` to set correct headers before templ.Render() |
| **Huma + Templ** | Trying to use Huma's content negotiation with Templ | Huma is API-first, Templ is HTML-first — use separate handler registration |
| **Atlas + go generate** | Running Atlas migrations during `go generate` | Atlas runs at deploy time, not build time — keep separate from code generation |

---

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| **N+1 queries in generated code** | 100 users = 101 queries, slow page loads | Generate eager loading / join queries, batch fetches | 1000+ records per page |
| **Regenerating on every build** | Build times balloon from 2s to 30s | Cache parsed schema AST, only regenerate on schema changes | 50+ schema files |
| **Reflection in tight loops** | Generated code slow vs hand-written | Use code generation to avoid reflection, benchmark generated code | High-throughput APIs |
| **Unmarshaling full schema AST repeatedly** | `go generate` takes minutes | Parse once, cache, reuse across generators | 100+ schema files |
| **SSE connections without pooling** | Memory leak, connection exhaustion | Datastar connection pooling, cleanup in defer | 1000+ concurrent users |
| **Template re-parsing on every generation** | Slow `go generate` | `template.ParseFiles()` once, execute many times | 20+ templates |

---

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| **SQL injection in generated queries** | Generated code concatenates user input into SQL | Use parameterized queries in templates: `WHERE id = $1`, never string concatenation |
| **Exposing internal schema in OpenAPI** | Leaking database structure, internal field names | Whitelist exported fields in schema tags, generate separate API types |
| **No input validation in action layer** | Trusting generated OpenAPI validation alone | Generate validation in action layer, defense in depth |
| **Generated migration files with hardcoded credentials** | Atlas migration templates accidentally include DB URLs | Template variables for credentials, never hardcode |
| **SSRF via user-provided URLs in schema** | Generated code fetching arbitrary URLs | Validate/whitelist URLs during generation or in generated validators |

---

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Code Generation:** Generated code compiles but isn't type-checked with `go/types` — verify semantic correctness
- [ ] **go generate directives:** Written but not tested — run `go generate ./...` from clean checkout
- [ ] **Action layer:** Exists but still has duplication between API/HTML — audit for copy-pasted validation
- [ ] **Template tests:** Templates exist but no fixture tests — one bad schema will crash generation
- [ ] **Integration tests:** Libraries individually tested but not together — test Huma+Bob+River in combination
- [ ] **Comment preservation:** Comments in schema but missing in generated code — check OpenAPI docs for descriptions
- [ ] **Import management:** Generated code has import blocks but some unused — run `goimports`, not just `gofmt`
- [ ] **Circular dependency prevention:** Compiles now but will break with next entity — test two-entity bidirectional relationship
- [ ] **CI verification:** `go generate` in CI but doesn't verify output matches committed code — add `git diff --exit-code`
- [ ] **Error messages in generated code:** Code works but errors are generic "error occurred" — generate descriptive errors with context

---

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| **Circular imports in production** | HIGH | 1. Identify cycle with `go build -v`, 2. Extract shared types to `contracts/` package, 3. Regenerate all code, 4. Verify with fresh build |
| **Lost comments after refactor** | MEDIUM | 1. Switch to `dave/dst`, 2. Re-parse source with comment preservation, 3. Update templates to inject comments, 4. Regenerate |
| **Runtime DI framework causing prod panic** | HIGH | 1. Generate explicit constructors with Wire, 2. Replace reflection calls, 3. Add compile-time verification tests, 4. Gradual rollout |
| **Unmaintainable 500-line template** | MEDIUM | 1. Extract template helper functions, 2. Split into focused templates, 3. Build data prep layer, 4. Add template tests |
| **Stale generated code deployed** | LOW | 1. Add CI check `go generate && git diff`, 2. Document in CONTRIBUTING.md, 3. Pre-commit hook optional, 4. Make generate part of build |
| **Multi-library version conflict** | MEDIUM | 1. Pin exact versions in go.mod, 2. Run integration tests, 3. Document compatibility matrix, 4. Consider vendoring |
| **Action layer duplication discovered late** | HIGH | 1. Extract common logic to `actions/` package, 2. Refactor handlers to call actions, 3. Write tests for action layer, 4. Remove duplication |
| **go/ast instead of go/types throughout codebase** | HIGH | 1. Incremental migration, 2. Use `golang.org/x/tools/go/packages`, 3. Replace `ast.Object` with `types.Object`, 4. Extensive testing |

---

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Circular import bootstrapping | Phase 1: Foundation | Build two-entity schema with relationship, verify zero gen/ imports |
| Comment preservation failures | Phase 1: Foundation | Generate code from heavily commented schema, check OpenAPI output |
| Runtime reflection vs compile-time | Phase 2: Code Generation | Replace any reflection with Wire/generation, benchmark |
| Template debugging hell | Phase 2: Code Generation | Template files under 200 lines, all have fixture tests |
| Multi-library integration complexity | Phase 1: Foundation, ongoing | Integration test matrix for all library pairs |
| Action layer duplication | Phase 3: Action Layer | Single `CreateUser` test used by both API and HTML handler tests |
| go/types vs go/ast confusion | Phase 2: Code Generation | Type resolution tests for pointers, embedded structs, imports |
| Generated code version control | Phase 1: Foundation | CI check: `go generate && git diff --exit-code` |
| go/format doesn't validate | Phase 2: Code Generation | Type-check generated code with `go/types` in test suite |
| Forgetting go generate | Phase 1: Foundation | Pre-commit hook + CI check, make generate part of build |

---

## Sources

**Go AST and Code Generation:**
- [Rewriting Go source code with AST tooling - Eli Bendersky](https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/)
- [dave/dst - Decorated Syntax Tree](https://github.com/dave/dst)
- [Code Generation From the AST - Gopher Academy](https://blog.gopheracademy.com/code-generation-from-the-ast/)
- [go/ast package documentation](https://pkg.go.dev/go/ast)
- [Using code generation to survive without generics in Go - Calhoun.io](https://www.calhoun.io/using-code-generation-to-survive-without-generics-in-go/)

**Framework Lessons:**
- [Ent Framework - GitHub](https://github.com/ent/ent)
- [GORM in Go: My Experience and Trade-offs](https://medium.com/@felipe.ascari_49171/gorm-in-go-my-experience-and-trade-offs-9eb89408ee34)
- [Common Anti-Patterns in Go Web Applications - Three Dots Labs](https://threedots.tech/post/common-anti-patterns-in-go-web-applications/)
- [Buffalo Framework Issues - GitHub](https://github.com/gobuffalo/buffalo/issues)

**Circular Dependencies:**
- [Managing Circular Dependencies in Go: Best Practices](https://medium.com/@cosmicray001/managing-circular-dependencies-in-go-best-practices-and-solutions-723532f04dde)
- [Import Cycles in Golang - Jogendra](https://jogendra.dev/import-cycles-in-golang-and-how-to-deal-with-them)
- [Kiota circular references issue](https://github.com/microsoft/kiota/issues/2834)

**Dependency Injection:**
- [You probably don't need a DI framework - Redowan's Reflections](https://rednafi.com/go/di-frameworks-bleh/)
- [Compile-time Dependency Injection With Go Cloud's Wire](https://go.dev/blog/wire)
- [Dependency Injection in Go: Patterns & Best Practices](https://medium.com/@rosgluk/dependency-injection-in-go-patterns-best-practices-5e5136df5357)

**Template and Code Generation:**
- [A comprehensive guide to go generate - Eli Bendersky](https://eli.thegreenplace.net/2021/a-comprehensive-guide-to-go-generate/)
- [text/template package](https://pkg.go.dev/text/template)
- [Dealing with Go Template errors at runtime](https://medium.com/@leeprovoost/dealing-with-go-template-errors-at-runtime-1b429e8b854a)

**Multi-Library Integration:**
- [Go Database Patterns: GORM, sqlx, and pgx Compared](https://dasroot.net/posts/2025/12/go-database-patterns-gorm-sqlx-pgx-compared/)
- [Database migrations in Go with golang-migrate](https://betterstack.com/community/guides/scaling-go/golang-migrate/)
- [Picking a database migration tool for Go projects - Atlas](https://atlasgo.io/blog/2022/12/01/picking-database-migration-tool)
- [Golang: Database Migration Using Atlas and Goose](https://volomn.com/blog/database-migration-using-atlas-and-goose)

**River Background Jobs:**
- [River: Fast and reliable background jobs in Go](https://riverqueue.com/)
- [River announcement blog post](https://riverqueue.com/blog/announcing-river)
- [River: a Fast, Robust Job Queue for Go + Postgres](https://brandur.org/river)

**Datastar and Hypermedia:**
- [Datastar Getting Started Guide](https://data-star.dev/guide/getting_started)
- [Real-time Sys Stats: Go + SSE + Data-Star](https://kitemetric.com/blogs/real-time-system-stats-with-go-server-sent-events-sse-and-data-star)
- [Datastar documentation](https://data-star.dev/)

**Huma Framework:**
- [Huma REST/HTTP API Framework - GitHub](https://github.com/danielgtaylor/huma)
- [Building OpenAPI Based REST API In Go Using HUMA Framework](https://shijuvar.medium.com/building-openapi-based-rest-api-in-go-using-huma-framework-with-surrealdb-844ded6a856e)
- [Huma documentation](https://huma.rocks/)

**Templ:**
- [HTMX with Go templ - Callista](https://callistaenterprise.se/blogg/teknik/2024/01/08/htmx-with-go-templ/)
- [Type-safe HTML generation with templ](https://tasukehub.com/articles/templ-go-htmx-type-safe-guide?lang=en)
- [Datastar + Templ documentation](https://templ.guide/server-side-rendering/datastar/)

**Type Safety and Validation:**
- [genqlient: A truly type-safe Go GraphQL client](https://blog.khanacademy.org/genqlient-a-truly-type-safe-go-graphql-client/)
- [Type-Safe Database Operations in Go with sqlc](https://leapcell.io/blog/type-safe-database-operations-in-go-with-go-generate-and-sqlc)
- [Exploring SQLC: Generating Type-Safe Go Code From SQL](https://medium.com/goturkiye/exploring-sqlc-generating-type-safe-go-code-from-sql-7dd57e76b245)

---

*Research confidence: MEDIUM-HIGH — Validated through official documentation, framework source code, community discussions, and established Go best practices. Some pitfalls based on common patterns across multiple frameworks (Ent, Buffalo, GORM) rather than Forge-specific experience.*
