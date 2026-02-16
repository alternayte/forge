# Technology Stack

**Project:** Forge
**Researched:** 2026-02-16
**Confidence:** HIGH

## Recommended Stack

### Core Framework

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.24+ | Programming language | Required by Huma v2. Go 1.22+ stdlib routing with wildcards/method matching eliminates need for third-party router. Go 1.23+ adds `slices.Backward` for clean middleware chains without dependencies. |
| net/http.ServeMux | stdlib (1.24+) | HTTP routing | Go 1.22 added method matching and wildcards (`/items/{id}`). Sufficient for most use cases. Avoid third-party routers unless you need advanced features like nested route groups. |
| Huma v2 | v2.35.0+ | OpenAPI framework | Router-agnostic REST framework with automatic OpenAPI 3.1 generation. Requires Go 1.24+. Production-ready with active maintenance (released 2026-01-18). Integrates with stdlib ServeMux via `humago` adapter. |

### Database

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| PostgreSQL | 13+ | Primary database | Required by River. pgx v5 supports PostgreSQL 13+. Industry standard for production SaaS. |
| pgx | v5.7.0+ | PostgreSQL driver | Pure Go driver with native interface for performance and PostgreSQL-specific features. Supports Go 1.24+. pgx v5 removed tri-state Status system in favor of sql-like Valid boolean. |
| SQLC | v1.30.0+ | Type-safe SQL | Generates type-safe Go from SQL queries. pgx/v5 support since v1.18.0. Use for complex queries where type safety and explicit SQL control are needed. |
| Bob | v0.42.0+ | Query builder | SQL query builder and ORM for PostgreSQL with spec-compliant query building. Use for dynamic query construction in resource handlers. Latest release 2024-11-25. |
| Atlas | v0.25+ | Schema migrations | Declarative schema-as-code migrations. State-of-the-art planning, linting, and applying. Only supports last two minor versions (e.g., v0.24 and v0.25). |

### Background Jobs

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| River | v0.30.2+ | Job queue | Atomic, transaction-safe job queueing backed by PostgreSQL. Jobs never run before transaction commits, never lost. Leverages generics for strongly-typed worker args. Latest release 2026-01-27. Optional self-hosted UI (River UI). |

### Frontend/Hypermedia

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Templ | v0.3.977+ | HTML templates | Type-safe Go template language that compiles to performant Go code. Allows calling any Go code, standard control flow. Latest release 2024-12-31. |
| Datastar | v4.0.0+ | Hypermedia/SSE | Lightweight (11.08 KiB) reactive framework using `data-*` attributes. Supports SSE for real-time updates. Backend-driven state management. Go SDK available. |
| Tailwind CSS | Standalone CLI v4+ | Styling | Zero npm/Node.js required. Self-contained executable for macOS/Linux/Windows. v4 is easier to use with standalone CLI than v3. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| go-playground/validator | v10+ | Struct validation | Cross-field, cross-struct, map/slice/array diving validation via struct tags. De facto standard. Use when struct tags suffice. |
| ozzo-validation | v4+ | Code-based validation | Validation via code (not tags) for complex logic. Use when struct tags become cumbersome or you need greater flexibility. |
| log/slog | stdlib (1.21+) | Structured logging | Standard library structured logging with JSON/Text handlers. Avoid third-party loggers (Zap, Zerolog) unless you need performance beyond slog. |
| errors | stdlib | Error handling | Standard library errors package (updated 2026-02-10). Use `errors.New`, `errors.Is`, `errors.As`. Avoid `pkg/errors` (maintenance mode due to Go2 error proposals). |
| stretchr/testify | v1.10.0+ | Testing/mocking | Common assertions and mocks for testing. Use with `mockery` for auto-generated mocks. Standard for Go testing in 2026. |

### Code Generation/Tooling

| Tool | Purpose | Notes |
|------|---------|-------|
| go/ast + go/parser | Schema parsing | Stdlib packages for parsing Go source without compilation. Use `parser.ParseFile` to build AST, `ast.Inspect` to traverse. Mutable AST can be modified and formatted back via `go/format`. |
| go/format | Code formatting | Emit Go-formatted source from AST. |
| Atlas CLI | Migration generation | Declarative migration planning from schema. |
| SQLC CLI | Query codegen | Generate type-safe Go from SQL files. |

## Installation

```bash
# Core dependencies
go get github.com/danielgtaylor/huma/v2@v2.35.0
go get github.com/jackc/pgx/v5@latest
go get github.com/stephenafamo/bob@v0.42.0
go get github.com/sqlc-dev/sqlc@v1.30.0
go get github.com/riverqueue/river@v0.30.2
go get github.com/a-h/templ@v0.3.977

# Validation
go get github.com/go-playground/validator/v10@latest

# Testing
go get github.com/stretchr/testify@v1.10.0

# Atlas CLI (install standalone binary)
curl -sSf https://atlasgo.sh | sh

# Tailwind CSS standalone CLI (macOS arm64 example)
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
chmod +x tailwindcss-macos-arm64
mv tailwindcss-macos-arm64 tailwindcss
```

## Alternatives Considered

| Category | Recommended | Alternative | Why Not Alternative |
|----------|-------------|-------------|---------------------|
| HTTP Router | net/http.ServeMux (stdlib) | Chi, Gorilla Mux, Fiber | Go 1.22+ stdlib has wildcards and method matching. Third-party routers add complexity unless you need advanced features (nested groups). Huma works with any router. |
| OpenAPI Framework | Huma v2 | Gin, Echo, Fiber | Huma is router-agnostic and generates OpenAPI 3.1 automatically. Active development (2026-01-18 release). |
| Database Driver | pgx v5 | lib/pq | lib/pq is in maintenance mode. pgx is pure Go, faster, and provides native PostgreSQL interface. |
| Query Builder | Bob | Squirrel, Goqu | Bob is spec-compliant and designed for PostgreSQL/MySQL/SQLite. Active development. |
| ORM | Bob + SQLC (hybrid) | GORM, Ent | GORM is magic-heavy. Ent requires buy-in to full framework. Bob gives query building, SQLC gives type safety for complex queries. Hybrid approach balances flexibility and safety. |
| Migrations | Atlas | golang-migrate, Goose | Atlas is declarative (schema-as-code) with linting and automatic planning. golang-migrate is imperative. Atlas can import golang-migrate format for migration. |
| Job Queue | River | Asynq, Machinery | River is PostgreSQL-native (no Redis), atomic with transactions, and uses generics for type safety. Latest release 2026-01-27. |
| Templates | Templ | html/template, Gomponents | html/template is stdlib but not type-safe. Templ compiles to Go code with type safety and IDE support. Gomponents is Go-only (no template syntax). |
| Hypermedia | Datastar | HTMX | Datastar is designed for SSE from the ground up (HTMX added SSE later). Datastar is lighter (11 KiB) and more backend-driven. |
| Validation | validator v10 | ozzo-validation | Both are good. validator uses struct tags (concise). ozzo uses code (flexible). Pick validator for most cases, ozzo for complex logic. |
| Logging | log/slog (stdlib) | Zap, Zerolog, Logrus | slog is stdlib (Go 1.21+) with structured logging. Avoid third-party unless performance is critical. |
| Error Handling | errors (stdlib) | pkg/errors | pkg/errors is in maintenance mode (Go2 proposals). stdlib errors has `Is`, `As`, wrapping via `fmt.Errorf("%w", err)`. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| lib/pq | Maintenance mode. Slower than pgx. | pgx v5 |
| pkg/errors | Maintenance mode due to Go2 error proposals. | errors (stdlib) |
| golang-migrate (directly) | Imperative migrations lack linting/planning. | Atlas (can import golang-migrate format) |
| GORM | Magic-heavy, reflection-based, opaque queries. Not explicit enough for schema-driven framework. | Bob + SQLC hybrid |
| Third-party routers (Chi, Gorilla) | Unnecessary complexity since Go 1.22+ stdlib routing. | net/http.ServeMux |
| HTMX | SSE support added later. Not as SSE-native as Datastar. | Datastar |
| npm/Node.js for Tailwind | Adds Node.js dependency for CSS tooling. | Tailwind standalone CLI |
| go-bindata, embed alternatives | Use `//go:embed` (stdlib since Go 1.16). | `//go:embed` directive |

## Stack Patterns by Variant

### For Simple CRUD Endpoints
- Use Bob query builder for dynamic queries
- Use SQLC for complex joins or performance-critical queries
- Use validator v10 for struct validation

### For Real-Time Features
- Use Datastar with SSE
- Use River for async background updates
- Use Templ for streaming HTML

### For Schema-Driven Codegen
- Parse schema with go/ast + go/parser (no compilation)
- Generate migrations with Atlas (declarative)
- Generate queries with SQLC (type-safe SQL)
- Generate Templ components from schema
- Emit code with go/format

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| Huma v2.35.0 | Go 1.24+ | Requires Go 1.24 or newer. |
| pgx v5 | Go 1.24+, PostgreSQL 13+ | Supports last 2 Go releases, last 5 years PostgreSQL. |
| SQLC v1.30.0 | pgx/v5 (since v1.18.0) | Use `sql_package: "pgx/v5"` in sqlc.yaml. |
| River v0.30.2 | PostgreSQL | PostgreSQL-native job queue. |
| Bob v0.42.0 | pgx v5 | Works with pgx driver. |
| Atlas | Go projects | CLI-based, not Go version dependent. |
| Templ v0.3.977 | Go 1.18+ (generics) | Compiles to Go code. |
| validator v10 | Go 1.18+ | Requires generics. |
| testify v1.10.0 | Go 1.17+ | Standard testing library. |

## Sources

### High Confidence (Official Docs, Context7)
- [Huma GitHub Repository](https://github.com/danielgtaylor/huma) - v2.35.0 release info, Go 1.24 requirement
- [pgx GitHub Repository](https://github.com/jackc/pgx) - Version support, PostgreSQL compatibility
- [River GitHub Repository](https://github.com/riverqueue/river) - v0.30.2 release (2026-01-27)
- [Templ GitHub Releases](https://github.com/a-h/templ/releases) - v0.3.977 (2024-12-31)
- [Bob GitHub Repository](https://github.com/stephenafamo/bob) - v0.42.0 release (2024-11-25)
- [Datastar Official Site](https://data-star.dev/) - v4.0.0, SSE features
- [SQLC GitHub Repository](https://github.com/sqlc-dev/sqlc) - v1.30.0 release
- [Atlas Official Site](https://atlasgo.io/) - Version support policy
- [Go Blog: Routing Enhancements](https://go.dev/blog/routing-enhancements) - Go 1.22 routing features
- [Go Blog: Structured Logging with slog](https://go.dev/blog/slog) - log/slog introduction
- [Tailwind CSS Standalone CLI](https://tailwindcss.com/blog/standalone-cli) - Installation, usage

### Medium Confidence (WebSearch verified with official source)
- [SQLC pgx v5 Support Documentation](https://docs.sqlc.dev/en/stable/guides/using-go-and-pgx.html) - pgx/v5 since v1.18.0
- [Better Stack: PostgreSQL in Go using PGX](https://betterstack.com/community/guides/scaling-go/postgresql-pgx-golang/) - pgx usage patterns
- [Alex Edwards: Organize Go Middleware](https://www.alexedwards.net/blog/organize-your-go-middleware-without-dependencies) - Go 1.23 slices.Backward for middleware
- [Better Stack: Logging in Go with Slog](https://betterstack.com/community/guides/logging/logging-in-go/) - slog best practices
- [Better Stack: Testing with Testify](https://betterstack.com/community/guides/scaling-go/golang-testify/) - testify v1.10.0 usage
- [Eli Bendersky: Rewriting Go with AST](https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/) - go/ast code generation patterns

### Medium Confidence (Multiple sources agree)
- [GitHub: go-playground/validator](https://github.com/go-playground/validator) - De facto standard validation
- [GitHub: pkg/errors maintenance mode](https://github.com/pkg/errors) - Go2 error proposals
- [Go Packages: errors stdlib](https://pkg.go.dev/errors) - Updated 2026-02-10

---
*Stack research for: Schema-driven Go web framework*
*Researched: 2026-02-16*
