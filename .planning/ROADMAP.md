# Roadmap: Forge

## Overview

Forge is built in 8 phases that progress from foundational architecture through code generation capabilities to user-facing features. The roadmap prioritizes solving the bootstrapping constraint (schema files cannot import generated code) before building the dual-interface code generator (HTML + API from single schema). Phases follow strict dependencies: foundation before features, data layer before business logic, API before HTML, and simple relationships before complex ones. Each phase delivers a coherent, verifiable capability that enables subsequent work.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation & Schema DSL** - Bootstrapping architecture and schema definition API
- [ ] **Phase 2: Code Generation Engine** - Template system and basic model/migration generation
- [ ] **Phase 3: Query & Data Access** - Type-safe queries, CRUD operations, and database integration
- [ ] **Phase 4: Action Layer & Error Handling** - Shared business logic layer for dual interface
- [ ] **Phase 5: REST API Generation** - Huma integration, OpenAPI 3.1, and route generation
- [ ] **Phase 6: Hypermedia UI Generation** - Templ/Datastar integration and HTML handlers
- [ ] **Phase 7: Advanced Data Features** - Relationships, soft delete, multi-tenancy, audit logging
- [ ] **Phase 8: Background Jobs & Production Readiness** - River integration, observability, CLI, deployment

## Phase Details

### Phase 1: Foundation & Schema DSL
**Goal**: Developer can define resource schemas that are parseable without gen/ existing (bootstrapping constraint solved)
**Depends on**: Nothing (first phase)
**Requirements**: SCHEMA-01, SCHEMA-02, SCHEMA-03, SCHEMA-04, SCHEMA-05, SCHEMA-06, PARSE-01, PARSE-02, PARSE-03, CLI-01, CLI-06, DEPLOY-04
**Success Criteria** (what must be TRUE):
  1. Developer can define a resource schema with field types, modifiers, and relationships without importing gen/ packages
  2. CLI can parse schema.go files using go/ast and extract definitions into an intermediate representation
  3. Parser produces clear error messages pointing to exact line numbers when schemas use dynamic values
  4. Project can be initialized with forge init creating proper directory structure and forge.toml
  5. Required tool binaries (templ, sqlc, tailwind, atlas) can be downloaded via forge tool sync
**Plans:** 5 plans

Plans:
- [ ] 01-01-PLAN.md — Schema DSL package (field types, modifiers, relationships, options, Define)
- [ ] 01-02-PLAN.md — IR types, rich error diagnostics, and terminal UI styles
- [ ] 01-03-PLAN.md — go/ast parser (TDD: parse schemas into IR with error collection)
- [ ] 01-04-PLAN.md — Cobra CLI skeleton and forge init scaffolding
- [ ] 01-05-PLAN.md — forge tool sync (binary download and management)

### Phase 2: Code Generation Engine
**Goal**: Developer runs forge generate and gets compilable Go models with migrations and proper code quality
**Depends on**: Phase 1
**Requirements**: GEN-01, GEN-04, GEN-08, GEN-11, GEN-13, MIGRATE-01, MIGRATE-02, MIGRATE-03, MIGRATE-04, MIGRATE-05, MIGRATE-06, CLI-02
**Success Criteria** (what must be TRUE):
  1. Running forge generate produces Go model types (Resource, ResourceCreate, ResourceUpdate) in gen/models/
  2. Running forge generate produces Atlas HCL migration files in gen/atlas/
  3. Generated code passes go/format and compiles without errors
  4. forge migrate diff creates SQL migration files based on schema changes
  5. forge migrate up/down applies and rolls back migrations successfully
  6. forge dev starts a development server with file watching and hot reload
**Plans:** 5 plans

Plans:
- [ ] 02-01-PLAN.md — Generator infrastructure and Go model type generation
- [ ] 02-02-PLAN.md — Atlas HCL schema generation and test factory generation
- [ ] 02-03-PLAN.md — forge generate CLI command (parser-generator pipeline)
- [ ] 02-04-PLAN.md — Atlas migration commands (diff, up, down, status) with destructive detection
- [ ] 02-05-PLAN.md — forge dev command with file watching and hot reload

### Phase 3: Query & Data Access
**Goal**: Developer can execute type-safe CRUD queries with dynamic filtering, sorting, and pagination
**Depends on**: Phase 2
**Requirements**: GEN-02, GEN-03, DATA-01, DATA-02, DATA-03, DATA-04, DATA-10, DATA-11, CLI-04
**Success Criteria** (what must be TRUE):
  1. Generated query builders support dynamic filtering with type-safe WHERE clauses (eq, contains, gte, lte)
  2. Generated query builders support dynamic sorting with type-safe ORDER BY
  3. Both offset-based and cursor-based pagination work correctly
  4. Developer can write raw SQLC queries in queries/custom/ as an escape hatch
  5. forge.Transaction wraps operations in database transactions
  6. forge db commands (create, drop, seed, console, reset) successfully manage development database
**Plans:** 3 plans

Plans:
- [ ] 03-01-PLAN.md — Validation function generation and query builder mod generation
- [ ] 03-02-PLAN.md — Pagination utilities, SQLC config, and transaction wrapper generation
- [ ] 03-03-PLAN.md — forge db CLI commands (create, drop, seed, console, reset)

### Phase 4: Action Layer & Error Handling
**Goal**: Both HTML and API handlers call the same action layer (no business logic duplication)
**Depends on**: Phase 3
**Requirements**: GEN-05, ACTION-01, ACTION-02, ACTION-03, ACTION-04, ACTION-05, ACTION-06, ACTION-07, ERR-01, ERR-02, ERR-03, ERR-04
**Success Criteria** (what must be TRUE):
  1. Each resource gets a generated Actions interface with List, Get, Create, Update, Delete methods
  2. Each resource gets a DefaultActions implementation handling validation, DB operations, and error mapping
  3. Developer can override actions by embedding DefaultActions and replacing specific methods
  4. Action layer automatically maps database errors to forge.Error with proper HTTP status codes
  5. Panic recovery middleware catches panics and returns 500 errors without exposing internals
**Plans:** 3 plans

Plans:
- [ ] 04-01-PLAN.md — forge.Error type, error constructors, and database error mapping generation
- [ ] 04-02-PLAN.md — Action interfaces and DefaultActions implementation generation
- [ ] 04-03-PLAN.md — Panic recovery middleware, error rendering, and orchestrator wiring

### Phase 5: REST API Generation
**Goal**: Developer gets a production-ready REST API with OpenAPI 3.1 documentation automatically generated from schema
**Depends on**: Phase 4
**Requirements**: GEN-06, GEN-07, API-01, API-02, API-03, API-04, API-05, API-06, API-07, API-08, API-09, API-10, AUTH-04, CLI-05
**Success Criteria** (what must be TRUE):
  1. Generated Huma handlers expose CRUD endpoints under /api/v1/<resource> with proper validation
  2. OpenAPI 3.1 spec is served at /api/openapi.json and /api/openapi.yaml
  3. Interactive API docs (Scalar UI) are served at /api/docs
  4. API supports bearer token and API key authentication
  5. CORS configuration and rate limiting middleware protect API endpoints
  6. forge routes command lists all registered API routes
  7. forge openapi export successfully exports the OpenAPI spec to a file
**Plans:** 4 plans

Plans:
- [ ] 05-01-PLAN.md — Huma API struct templates (inputs, outputs, route registration)
- [ ] 05-02-PLAN.md — Auth infrastructure and middleware (bearer tokens, API keys, CORS, rate limiting)
- [ ] 05-03-PLAN.md — API server wiring, Scalar UI docs, and generator orchestrator integration
- [ ] 05-04-PLAN.md — CLI commands (forge routes, forge openapi export)

### Phase 6: Hypermedia UI Generation
**Goal**: Developer gets scaffolded HTML forms and views that use the same action layer as the API
**Depends on**: Phase 5
**Requirements**: GEN-09, GEN-10, GEN-12, HTML-01, HTML-02, HTML-03, HTML-04, HTML-05, HTML-06, HTML-07, AUTH-01, AUTH-02, AUTH-03, DEPLOY-05, TEST-01, TEST-02, TEST-03, TEST-04
**Success Criteria** (what must be TRUE):
  1. forge generate resource <name> scaffolds Templ form, list, and detail components into resources/
  2. forge generate resource <name> scaffolds HTML handlers that call the action layer
  3. Scaffolded forms render with Datastar-native interactivity and display field-level validation errors
  4. Scaffolded list views render tables with sort headers, filter controls, and pagination
  5. forge generate resource <name> --diff shows differences between current and freshly scaffolded views
  6. Session-based email/password and OAuth2 (Google, GitHub) authentication work end-to-end
  7. Generated test factories create valid resource instances with builder pattern
  8. forgetest.NewTestDB and forgetest.NewApp enable HTTP testing
  9. Tailwind CSS compiles via standalone CLI binary (zero npm dependency verified)
**Plans:** 5/9 plans executed

Plans:
- [ ] 06-01-PLAN.md — Form primitives library (FormField, TextInput, etc.) and Datastar SSE helpers
- [ ] 06-02-PLAN.md — Session auth (SCS + pgxstore, bcrypt, HTML auth middleware)
- [ ] 06-03-PLAN.md — Templ view scaffold templates (form, list, detail)
- [ ] 06-04-PLAN.md — HTML handler scaffold templates and route registration
- [ ] 06-05-PLAN.md — OAuth2 providers (Goth) and sessions table
- [ ] 06-06-PLAN.md — Scaffold generator (ScaffoldResource, DiffResource) and tests
- [ ] 06-07-PLAN.md — Testing infrastructure (forgetest: NewTestDB, NewApp, PostDatastar)
- [ ] 06-08-PLAN.md — HTML generation orchestrator and htmlRoutes() CLI
- [ ] 06-09-PLAN.md — forge generate resource CLI, HTML server wiring, Tailwind compilation

### Phase 7: Advanced Data Features
**Goal**: Developer can use relationships, soft delete, multi-tenancy, permissions, and audit logging from schema annotations
**Depends on**: Phase 6
**Requirements**: SCHEMA-07, SCHEMA-08, DATA-05, DATA-06, DATA-07, DATA-08, DATA-09, DATA-12, AUTH-05, AUTH-06, TENANT-01, TENANT-02, TENANT-03, TENANT-04, AUDIT-01, AUDIT-02, AUDIT-03
**Success Criteria** (what must be TRUE):
  1. Developer can define field-level Visibility and Mutability, and generated actions strip invisible fields based on user role
  2. Developer can define CRUD-level Permissions, and generated actions check permissions before operations
  3. When SoftDelete is enabled, queries exclude soft-deleted records by default and WithTrashed/OnlyTrashed work correctly
  4. When SoftDelete is enabled with Unique fields, Atlas generates partial unique indexes (WHERE deleted_at IS NULL)
  5. When TenantScoped is enabled, all queries automatically include tenant_id filtering from context
  6. Atlas generates row-level security (RLS) policies for tenant-scoped resources
  7. When Auditable is enabled, created_by/updated_by columns are auto-populated and JSONB change diffs are recorded
  8. Relationship preloading (Eager) automatically excludes soft-deleted related records
**Plans**: TBD

Plans:
- TBD

### Phase 8: Background Jobs & Production Readiness
**Goal**: Production binary is deployable with background jobs, observability, and full CLI tooling
**Depends on**: Phase 7
**Requirements**: SCHEMA-09, JOBS-01, JOBS-02, JOBS-03, JOBS-04, SSE-01, SSE-02, SSE-03, SSE-04, SSE-05, SSE-06, CLI-03, CLI-07, DEPLOY-01, DEPLOY-02, DEPLOY-03, OTEL-01, OTEL-02, OTEL-03
**Success Criteria** (what must be TRUE):
  1. Schema-defined AfterCreate/AfterUpdate hooks automatically enqueue River jobs in the same DB transaction
  2. Job queues are configurable via forge.toml with queue names and concurrency limits
  3. SSE connection limits (global and per-user) prevent resource exhaustion
  4. PostgreSQL LISTEN/NOTIFY hub fans out events to subscribers via Go channels
  5. forge build compiles a single production binary < 30MB with embedded templates, static assets, and migrations
  6. Production binary starts in < 100ms cold start
  7. forge deploy packages for deployment successfully
  8. OpenTelemetry traces, Prometheus metrics, and structured logging (slog) work in production
  9. All configuration via forge.toml is overridable by environment variables (12-factor)
**Plans**: TBD

Plans:
- TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Schema DSL | 0/5 | Planned | - |
| 2. Code Generation Engine | 0/5 | Planned | - |
| 3. Query & Data Access | 0/3 | Planned | - |
| 4. Action Layer & Error Handling | 0/3 | Planned | - |
| 5. REST API Generation | 0/4 | Planned | - |
| 6. Hypermedia UI Generation | 5/9 | In Progress|  |
| 7. Advanced Data Features | 0/TBD | Not started | - |
| 8. Background Jobs & Production Readiness | 0/TBD | Not started | - |
