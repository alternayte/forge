# Project Research Summary

**Project:** Forge
**Domain:** Schema-driven Go web framework for SaaS applications
**Researched:** 2026-02-16
**Confidence:** HIGH

## Executive Summary

Forge is a schema-driven code generation framework for building full-stack Go web applications with PostgreSQL backends. Research shows this domain follows established patterns from Rails scaffolding, Django Admin, and Laravel Artisan, but with Go's type safety advantages. The recommended approach combines modern Go stdlib features (1.24+ routing), specialized libraries (Huma for OpenAPI, pgx v5 for PostgreSQL, Bob for queries, River for background jobs), and hypermedia-driven UI (Templ + Datastar) to deliver both HTML and JSON from a single schema definition.

The key technical challenge is the bootstrapping constraint: schema definitions must never import generated code to avoid circular dependencies. This requires careful architectural planning with clear separation between schema API, parsing layer, code generation, and runtime library. The primary competitive advantage is the dual interface approach—generating both hypermedia web UI and REST API from the same action layer, eliminating duplication between API and HTML handlers that plagues traditional frameworks.

Critical risks center on Go-specific code generation pitfalls: circular imports, comment preservation failures, template debugging complexity, and multi-library integration issues. These can be mitigated through proven patterns: AST-based parsing (not compile-and-execute), dave/dst for comment preservation, small focused templates, and an adapter layer isolating the framework from library-specific types. The research provides high confidence for core functionality (CRUD, migrations, relationships) but medium confidence for advanced features (real-time, field permissions) which should be deferred to post-MVP.

## Key Findings

### Recommended Stack

The technology stack prioritizes Go 1.24+ stdlib capabilities combined with specialized libraries for OpenAPI generation, PostgreSQL operations, and hypermedia rendering. No JavaScript tooling required.

**Core technologies:**
- **Go 1.24+ with net/http.ServeMux**: Stdlib routing with wildcards and method matching eliminates need for third-party routers. Clean middleware chains without dependencies.
- **Huma v2 (v2.35.0+)**: Router-agnostic REST framework with automatic OpenAPI 3.1 generation. Production-ready with active maintenance.
- **PostgreSQL 13+ with pgx v5**: Native Go driver with full PostgreSQL feature support. Industry standard for production SaaS.
- **Bob v0.42.0 + SQLC v1.30.0**: Hybrid query approach—Bob for dynamic queries, SQLC for type-safe complex queries. Balances flexibility and safety.
- **Atlas v0.25+**: Declarative schema-as-code migrations with linting and automatic planning. Superior to imperative migrations.
- **River v0.30.2**: PostgreSQL-native job queue with atomic transaction safety. Uses generics for type-safe worker arguments.
- **Templ v0.3.977 + Datastar v4.0.0**: Type-safe HTML templates compiled to Go + lightweight (11KB) SSE-native hypermedia framework. Zero npm dependencies.
- **Tailwind CSS standalone CLI**: Self-contained executable, no Node.js required.

**Critical version requirements:**
- Go 1.24+ required by Huma v2
- PostgreSQL 13+ required by River and pgx v5
- Use stdlib errors, log/slog, validator v10 for validation

### Expected Features

Users comparing against Rails scaffold, Laravel Artisan, Django Admin, and Go frameworks (Ent, Buffalo) have clear feature expectations.

**Must have (table stakes):**
- Database migrations with up/down support and rollback safety
- CRUD operations for all resources
- Type-safe Go struct generation from schema
- Validation rules (field-level and custom business logic)
- RESTful route generation
- Form components for create/edit
- List/index views with tables
- Relationships (1:1, 1:N with cascade rules)
- Basic filtering and pagination
- Timestamps (created_at, updated_at auto-managed)

**Should have (competitive advantage):**
- Dual interface (hypermedia + REST API) from single schema—core differentiator
- Action layer abstraction to share business logic between HTML and API handlers
- OpenAPI auto-generation with Swagger UI
- Soft delete support with schema flag
- Sorting on columns
- Advanced filtering (range, multi-select, date ranges)
- Bulk operations (batch insert/update/delete)
- Custom actions beyond CRUD (approve, archive, publish)
- Audit trail generation from schema annotations
- Seeding with realistic test data based on field types

**Defer (v2+):**
- Real-time capabilities (WebSocket/SSE for live updates)—complex, defer until demand validated
- Field-level permissions (enterprise feature)
- Optimistic locking (edge case, address when needed)
- N:M relationships (many-to-many)—add after 1:1 and 1:N proven

**Anti-features to avoid:**
- Multi-database support (dilutes PostgreSQL advantages)
- NoSQL/document store (contradicts schema-driven approach—use JSONB instead)
- Magic relationship detection (fragile—require explicit declaration)
- Cascading deletes by default (data loss risk—require explicit choice)
- Admin panel generation (provide building blocks instead)
- GraphQL generation (complexity explosion—stick to hypermedia + REST)

### Architecture Approach

Schema-driven frameworks follow a consistent pattern with five layers: schema definition API, CLI parser, code generator, generated code, and runtime library. The bootstrapping constraint is central—schema files cannot import generated code.

**Major components:**
1. **Schema Package** (`schema/`) — User-facing API for defining entities, fields, edges, validations. Only imports framework's schema definition package, never generated code.
2. **CLI Parser** (`internal/parser/`) — Uses go/ast to parse schema files without compilation. Builds intermediate representation (IR) from AST. Validates semantics.
3. **Code Generator** (`internal/generator/`) — Transforms IR to Go code, migrations, API types, actions. Uses templates with go/format post-processing.
4. **Generated Code** (`gen/`) — Type-safe models, query builders, validators, API types, action interfaces. Imports runtime library. Always regenerated.
5. **Runtime Library** (`runtime/`) — Stable foundation with core abstractions, database connectivity, middleware. Separate module with independent versioning.
6. **Scaffolded Code** (`resources/`) — User-editable handlers, views, hooks. Generated once, then owned by user. Not regenerated.

**Key architectural patterns:**
- **Bootstrapping constraint via import isolation**: Schema → Parser (AST-based, no compilation) → Generator → Generated Code (imports runtime). No cycles.
- **Two-phase generation**: Always-regenerated code in `gen/`, edit-once scaffolded code in `resources/`. Clear separation prevents overwrites.
- **Intermediate representation decoupling**: Parser produces language-agnostic IR, multiple generators consume same IR (Go, TypeScript, docs).
- **Runtime library as stable foundation**: Generated code disposable, runtime API stable. Users import both.
- **CLI as orchestrator**: Use `//go:generate` directives, not build integration. Explicit generation step.

**Data flow:**
```
Schema Definition (schema/user.go)
  ↓ go generate ./schema
Parser (go/ast) → IR
  ↓
Generator → Templates → go/format
  ↓
gen/user.go, gen/user_query.go, gen/atlas.hcl
  ↓
Application imports gen/ + runtime/
```

### Critical Pitfalls

Research identified 10 major pitfalls based on analysis of Ent, SQLC, Buffalo, and Go best practices.

1. **Circular Import Bootstrapping** — Schema files importing generated code creates circular dependency. Go disallows this. **Solution:** Zero gen/ imports in schema packages. Use string references, resolve at codegen time.

2. **Comment Preservation Failures** — go/ast loses comments during AST manipulation. Generated code lacks documentation. **Solution:** Use dave/dst (Decorated Syntax Tree) for comment-preserving transformations. Parse with `parser.ParseComments`.

3. **Action Layer Duplication** — Business logic duplicated between API and HTML handlers. Update one, forget the other. **Solution:** Shared action layer below handlers. Generate action interfaces, handlers only map responses.

4. **Template Debugging Hell** — 500+ line templates with nested conditionals are unmaintainable. **Solution:** One template per output type, keep logic in Go not templates, run go/format on generated code, write template tests.

5. **Multi-Library Integration Complexity** — 6+ libraries (Huma, Bob, Atlas, River, Datastar, Templ) create integration hell. **Solution:** Adapter layer between framework and libraries. Pin exact versions. Integration tests for all library pairs.

6. **go/types vs go/ast Confusion** — Using go/ast for type resolution instead of go/types causes incorrect type inference. **Solution:** Use go/ast for structure, go/types for type checking. Set `parser.SkipObjectResolution`. Use `golang.org/x/tools/go/packages`.

7. **Generated Code Version Control** — Committing generated code causes merge conflicts. Not committing causes "works on my machine". **Solution:** Commit with clear headers. CI check: `go generate && git diff --exit-code`. Git attributes mark generated files.

8. **Forgetting go generate** — Schema updated, generation forgotten, stale code deployed. **Solution:** Pre-commit hook, CI verification, file watcher in dev mode, make generate part of build.

9. **Runtime Reflection vs Compile-Time** — Using reflection for DI causes runtime panics, cryptic errors. **Solution:** Prefer compile-time code generation (Wire) over runtime reflection (Dig). Generate explicit constructors.

10. **go/format Doesn't Validate** — Formatted code may have semantic errors (wrong types, undefined vars). **Solution:** Run go/types checker after generation. Compile generated code in tests. Use astutil for import management.

## Implications for Roadmap

Based on research findings, the roadmap should follow dependency order with clear separation between foundational architecture, code generation capabilities, and user-facing features. Critical pitfalls must be addressed in Phase 1 before building features on top.

### Phase 1: Foundation & Schema DSL
**Rationale:** Must establish bootstrapping architecture and schema definition API before any code generation. This is the one-way door—getting it wrong requires complete rewrite. Addresses circular import pitfall and establishes parser strategy.

**Delivers:**
- Schema definition API (`schema/` package) with field types, edges, validation builders
- go/ast-based parser that reads schema files without compilation
- Intermediate representation (IR) with semantic validation
- Runtime library structure (separate module, stable API)
- CI/CD verification (go generate check, commit workflow)

**Addresses features:** None directly—foundational work
**Avoids pitfalls:** #1 (circular imports), #6 (go/types setup), #8 (go generate automation)
**Research depth:** Standard patterns—no additional research needed

### Phase 2: Code Generation Engine
**Rationale:** Build template system and code emitters before generating domain logic. Must solve comment preservation and template architecture early. Establishes quality gates (go/format, go/types checking).

**Delivers:**
- Template architecture (one template per output type, under 200 lines each)
- Code emission with go/format and go/types verification
- Comment preservation via dave/dst
- Basic model generation (Go structs from schema)
- Migration generation (Atlas HCL output)
- Import management with astutil

**Addresses features:** Model/struct generation (table stakes)
**Avoids pitfalls:** #2 (comment preservation), #4 (template debugging), #9 (go/format validation)
**Uses stack:** Atlas for migrations, dave/dst for AST manipulation
**Research depth:** Standard patterns—no additional research needed

### Phase 3: Query & Data Access
**Rationale:** Type-safe queries are Go's advantage over Ruby/Python frameworks. Hybrid Bob+SQLC approach provides flexibility and safety. Must integrate with pgx v5.

**Delivers:**
- Bob query builder integration for dynamic queries
- SQLC integration for complex queries
- Type-safe query generation from schema
- Basic CRUD query methods
- Connection pooling and pgx integration
- Transaction support

**Addresses features:** Type-safe queries, CRUD operations (table stakes)
**Avoids pitfalls:** #5 (multi-library integration with Bob+pgx+SQLC)
**Uses stack:** Bob v0.42.0, SQLC v1.30.0, pgx v5
**Research depth:** Bob/SQLC integration needs phase research for query generation patterns

### Phase 4: Action Layer Abstraction
**Rationale:** This is Forge's core differentiator—shared business logic for dual interface. Must be solved before generating handlers. Prevents duplication between API and HTML.

**Delivers:**
- Action interface generation (Create, Read, Update, Delete)
- Default action implementation
- Hook points (BeforeCreate, AfterUpdate, etc.)
- Validation in action layer (defense in depth)
- Error handling patterns
- User extension mechanism

**Addresses features:** Action layer abstraction (differentiator)
**Avoids pitfalls:** #3 (action layer duplication—the critical one)
**Research depth:** Novel architecture—needs phase research for hook design and extension patterns

### Phase 5: REST API Generation (JSON)
**Rationale:** Build API first because it's simpler than HTML (no Templ complexity). Validates action layer works. Establishes Huma integration patterns.

**Delivers:**
- Huma handler generation for CRUD endpoints
- OpenAPI 3.1 schema generation
- Request/response type generation
- Validation integration (validator v10)
- Route registration (net/http.ServeMux)
- Content negotiation

**Addresses features:** Dual interface (API half), OpenAPI generation, route generation, validation rules (table stakes/differentiator)
**Avoids pitfalls:** #5 (Huma+Bob integration)
**Uses stack:** Huma v2, validator v10, net/http.ServeMux
**Research depth:** Standard REST patterns—minimal research needed

### Phase 6: Hypermedia UI Generation (HTML)
**Rationale:** HTML handlers consume same action layer as API. Proves dual interface works. Templ+Datastar integration is novel.

**Delivers:**
- Templ component generation (list, form, show)
- HTML handler generation (uses action layer)
- Datastar integration for interactivity
- Form rendering with client-side validation
- Pagination UI components
- Basic filtering UI

**Addresses features:** Form components, list views, basic filtering, pagination, dual interface (HTML half) (table stakes/differentiator)
**Avoids pitfalls:** #3 (uses action layer, no duplication)
**Uses stack:** Templ v0.3.977, Datastar v4.0.0, Tailwind CSS standalone
**Research depth:** Templ+Datastar integration needs phase research for SSE patterns and component architecture

### Phase 7: Relationships & Edges
**Rationale:** Relationships build on existing CRUD foundation. 1:1 and 1:N first (table stakes), defer N:M to later. Foreign key migrations and cascade rules.

**Delivers:**
- Relationship definition in schema (BelongsTo, HasMany)
- Foreign key migration generation
- Cascade rule options (RESTRICT, CASCADE, SET NULL)
- Join query generation
- Relationship preloading (avoid N+1)
- UI for related entities

**Addresses features:** Relationships (table stakes—1:1 and 1:N only)
**Avoids pitfalls:** #5 (Bob join generation)
**Uses stack:** Bob for joins, Atlas for foreign keys
**Research depth:** Standard patterns—minimal research needed

### Phase 8: Polish & Developer Experience
**Rationale:** After core features work, improve ergonomics. Seeding, sorting, soft delete enhance but aren't blocking.

**Delivers:**
- Soft delete support with deleted_at field
- Sorting on list views
- Test factory generation
- Seeding with faker integration
- CLI scaffolding commands
- Error message improvements

**Addresses features:** Soft delete, sorting, seeding, advanced filtering (partial) (competitive features)
**Research depth:** Standard patterns—minimal research needed

**Deferred to v2.0+:**
- Real-time capabilities (Phase 9+) — Complex, needs WebSocket/SSE architecture research
- Field-level permissions (Phase 10+) — Enterprise, needs auth system maturity
- N:M relationships (Phase 9+) — After 1:1 and 1:N proven
- Audit trail generation (Phase 9+) — After core features stable

### Phase Ordering Rationale

1. **Foundation before features**: Phases 1-2 establish architecture. Getting this wrong requires rewrite.
2. **Data layer before business logic**: Phase 3 (queries) before Phase 4 (actions) because actions use queries.
3. **API before HTML**: Phase 5 before Phase 6 because API is simpler, validates action layer works.
4. **Simple relationships before complex**: Phase 7 only 1:1 and 1:N. Defer N:M until core proven.
5. **Core before polish**: Phases 1-7 are MVP. Phase 8+ are enhancements.

**Dependencies addressed:**
- Circular import prevention (Phase 1) enables all later phases
- Action layer (Phase 4) shared by API (Phase 5) and HTML (Phase 6)
- Query generation (Phase 3) required by actions (Phase 4)
- Schema/parser (Phases 1-2) required by all generation phases

**Pitfalls mapped to prevention phases:**
- Phase 1: #1, #6, #8 (architecture and automation)
- Phase 2: #2, #4, #9 (code generation quality)
- Phase 3-8: #3, #5, #7 (feature implementation)

### Research Flags

**Phases needing phase-specific research:**
- **Phase 3 (Query & Data Access):** Bob+SQLC query generation patterns—how to generate efficient queries with Bob mods, when to use SQLC vs Bob
- **Phase 4 (Action Layer):** Hook extension patterns—how to design hook system, validation placement, user override mechanisms
- **Phase 6 (Hypermedia UI):** Templ+Datastar SSE integration—component architecture, SSE event patterns, progressive enhancement

**Phases with standard patterns (skip phase research):**
- **Phase 1:** Bootstrapping architecture is well-documented (Ent, SQLC, protobuf)
- **Phase 2:** Template systems and AST manipulation have established patterns
- **Phase 5:** REST API generation follows Huma documentation
- **Phase 7:** Relationship patterns standard across ORMs
- **Phase 8:** Polish features use existing infrastructure

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All libraries verified with official docs, GitHub releases, version compatibility confirmed. Go 1.24+ requirement clear. |
| Features | HIGH | Table stakes validated against Rails/Laravel/Django. Competitive features based on 2026 SaaS trends. Clear MVP definition. |
| Architecture | HIGH | Bootstrapping pattern proven by Ent/SQLC/protobuf. Component separation validated by multiple frameworks. |
| Pitfalls | MEDIUM-HIGH | 10 critical pitfalls identified from official sources, framework issues, community discussions. Some based on inference from similar frameworks. |

**Overall confidence:** HIGH

Research is comprehensive with strong primary sources (official documentation, Context7, framework repositories). The core technical approach—schema-driven code generation with bootstrapping constraint—is proven by Ent and SQLC. The differentiator—dual interface from action layer—is a novel combination but built on established patterns (Rails actions, HTMX hypermedia, content negotiation).

### Gaps to Address

**Gap 1: Datastar + Go integration patterns**
- **Issue:** Datastar is newer (v4.0.0), fewer Go-specific examples than HTMX
- **Impact:** Phase 6 (Hypermedia UI) implementation details
- **Mitigation:** Phase research when reaching Phase 6. Datastar docs are good, may need experimentation with SSE patterns

**Gap 2: Bob query builder code generation**
- **Issue:** Bob is used for dynamic queries, but generating Bob mods from schema needs validation
- **Impact:** Phase 3 (Query & Data Access) implementation approach
- **Mitigation:** Phase research for query generation patterns. Bob is well-documented, SQLC provides comparison

**Gap 3: Action layer hook design**
- **Issue:** Hook system design (before/after, sync/async, error handling) needs deeper exploration
- **Impact:** Phase 4 (Action Layer) extensibility
- **Mitigation:** Phase research examining Rails callbacks, Ent hooks, Django signals for design patterns

**Gap 4: Real-time architecture (deferred)**
- **Issue:** Real-time features (Phase 9+) architecture unclear—WebSocket vs SSE, connection pooling, River integration
- **Impact:** Deferred to v2.0+, not blocking MVP
- **Mitigation:** Address when validated demand emerges. Datastar SSE patterns may inform this.

**Gap 5: Multi-tenancy patterns (not researched)**
- **Issue:** Not covered in initial research, may be needed for SaaS applications
- **Impact:** Unknown—could affect schema design if needed early
- **Mitigation:** Monitor user requests. If needed, can be added as middleware pattern (anti-feature to build into framework, good as opt-in pattern)

All gaps have clear mitigation strategies. None are blocking for Phase 1-2 work. Phases 3-4 gaps addressed by phase-specific research. Phases 9+ gaps deferred appropriately.

## Sources

### Primary (HIGH confidence)
- **STACK.md** — Technology choices validated with official documentation, GitHub releases, version compatibility
- **FEATURES.md** — Feature expectations validated against Rails, Laravel, Django official docs and generator comparisons
- **ARCHITECTURE.md** — Architecture patterns validated with Ent, SQLC, protobuf/gRPC official sources and Go standard library docs
- **PITFALLS.md** — Pitfalls validated with framework source code analysis, official Go documentation, community discussions

### Research file sources (aggregated)
- [Huma GitHub Repository](https://github.com/danielgtaylor/huma) — v2.35.0 release, Go 1.24 requirement
- [pgx GitHub Repository](https://github.com/jackc/pgx) — v5 support, PostgreSQL 13+ compatibility
- [River GitHub Repository](https://github.com/riverqueue/river) — v0.30.2 release (2026-01-27)
- [Ent Code Generation Documentation](https://entgo.io/docs/code-gen/) — Bootstrapping patterns, architecture
- [Go AST Package Documentation](https://pkg.go.dev/go/ast) — AST parsing, go/types usage
- [Go Generate Blog Post](https://go.dev/blog/generate) — Code generation conventions
- [dave/dst - Decorated Syntax Tree](https://github.com/dave/dst) — Comment preservation
- [Rails Scaffolding Guide](https://www.railscarma.com/blog/scaffolding-in-ruby-on-rails-complete-guide/) — Feature expectations
- [Laravel Make:Model Guide](https://laraveldaily.com/lesson/laravel-eloquent-expert/make-model-options) — Generator patterns
- [HTMX in 2026 Trends](https://vibe.forem.com/del_rosario/htmx-in-2026-why-hypermedia-is-dominating-the-modern-web-41id) — Hypermedia resurgence
- [Bob GitHub Repository](https://github.com/stephenafamo/bob) — v0.42.0, PostgreSQL query builder
- [Atlas Official Site](https://atlasgo.io/) — Declarative migrations
- [Templ GitHub Releases](https://github.com/a-h/templ/releases) — v0.3.977 (2024-12-31)
- [Datastar Official Site](https://data-star.dev/) — v4.0.0, SSE features

---
*Research completed: 2026-02-16*
*Ready for roadmap: yes*
