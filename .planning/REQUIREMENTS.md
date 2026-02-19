# Requirements: Forge

**Defined:** 2026-02-16
**Core Value:** Schema is the single source of truth — define a resource once and everything is generated automatically with zero manual sync.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Schema

- [ ] **SCHEMA-01**: Developer can define a resource with UUID, String, Text, Int, BigInt, Decimal, Bool, DateTime, Date, Enum, JSON, Slug, Email, URL field types
- [ ] **SCHEMA-02**: Developer can apply field modifiers: Required, MaxLen, MinLen, Sortable, Filterable, Searchable, Unique, Index, Default, Immutable, Label, Placeholder, Help
- [ ] **SCHEMA-03**: Developer can define relationships: BelongsTo, HasMany, HasOne, ManyToMany with OnDelete cascade options
- [ ] **SCHEMA-04**: Developer can enable SoftDelete, Auditable, TenantScoped, and Searchable options per resource
- [ ] **SCHEMA-05**: Developer can define Timestamps() to auto-add created_at and updated_at fields
- [ ] **SCHEMA-06**: Developer can define Enum fields with specific allowed values and a Default
- [x] **SCHEMA-07**: Developer can define field-level Visibility and Mutability based on roles
- [x] **SCHEMA-08**: Developer can define CRUD-level Permissions (List, Read, Create, Update, Delete) per resource
- [ ] **SCHEMA-09**: Developer can define schema Hooks (AfterCreate, AfterUpdate) referencing River job kinds

### Parser

- [ ] **PARSE-01**: CLI can parse schema.go files using go/ast without requiring gen/ to exist (bootstrapping constraint)
- [ ] **PARSE-02**: Parser extracts schema.Define() calls, field definitions, modifiers, relationships, and options into an intermediate representation
- [ ] **PARSE-03**: Parser errors with a clear message pointing to the offending line when a schema uses dynamic values go/ast can't resolve

### Code Generation

- [ ] **GEN-01**: `forge generate` produces Go model types (Resource, ResourceCreate, ResourceUpdate, ResourceFilter, ResourceSort) in gen/models/
- [ ] **GEN-02**: `forge generate` produces Bob query builder mods for filtering, sorting, and pagination in gen/queries/
- [ ] **GEN-03**: `forge generate` produces validation functions returning typed field errors in gen/validation/
- [ ] **GEN-04**: `forge generate` produces Atlas desired state HCL files in gen/atlas/
- [ ] **GEN-05**: `forge generate` produces action interfaces and default implementations in gen/actions/
- [ ] **GEN-06**: `forge generate` produces Huma API input/output structs with proper struct tags in gen/api/
- [ ] **GEN-07**: `forge generate` produces Huma route registration functions in gen/api/
- [ ] **GEN-08**: `forge generate` produces test factory functions in gen/factories/
- [ ] **GEN-09**: `forge generate resource <name>` scaffolds Templ form, list, and detail components once into resources/
- [ ] **GEN-10**: `forge generate resource <name>` scaffolds HTML handlers and hooks once into resources/
- [ ] **GEN-11**: Running `forge generate` after schema changes updates gen/ artifacts without overwriting user-edited resources/ files
- [ ] **GEN-12**: `forge generate resource <name> --diff` shows diff between current and freshly scaffolded views
- [ ] **GEN-13**: Generated code passes go/format and compiles without errors

### Migrations

- [ ] **MIGRATE-01**: `forge migrate diff` runs Atlas schema diff against the live database and produces a SQL migration file
- [ ] **MIGRATE-02**: `forge migrate up` applies pending migrations via Atlas
- [ ] **MIGRATE-03**: `forge migrate down` rolls back the last migration
- [ ] **MIGRATE-04**: `forge migrate status` shows current migration state
- [ ] **MIGRATE-05**: Destructive migrations (column drops, type changes) print a prominent warning
- [ ] **MIGRATE-06**: Developer can add hand-written SQL migration files alongside generated ones

### Data Access

- [ ] **DATA-01**: Generated query builder supports dynamic filtering with type-safe WHERE clauses (eq, contains, gte, lte)
- [ ] **DATA-02**: Generated query builder supports dynamic sorting with type-safe ORDER BY
- [ ] **DATA-03**: Generated query builder supports offset-based pagination
- [ ] **DATA-04**: Generated query builder supports cursor-based pagination with opaque base64 cursors
- [x] **DATA-05**: When TenantScoped is true, all generated queries automatically include tenant_id filtering from context
- [x] **DATA-06**: When SoftDelete is true, all queries exclude soft-deleted records by default
- [x] **DATA-07**: Developer can use WithTrashed() and OnlyTrashed() to include/only-show soft-deleted records
- [x] **DATA-08**: When a field is Unique on a SoftDelete resource, Atlas generates a partial unique index (WHERE deleted_at IS NULL)
- [x] **DATA-09**: Generated actions include a Restore method for SoftDelete resources
- [ ] **DATA-10**: Developer can write raw SQLC queries in queries/custom/ as an escape hatch
- [ ] **DATA-11**: forge.Transaction wraps operations in a database transaction with River job enqueueing in the same transaction
- [ ] **DATA-12**: Relationship preloading (Eager) automatically excludes soft-deleted related records

### Action Layer

- [ ] **ACTION-01**: Each resource gets a generated Actions interface with List, Get, Create, Update, Delete methods
- [ ] **ACTION-02**: Each resource gets a generated DefaultActions implementation that handles validation, DB operations, and error mapping
- [ ] **ACTION-03**: Developer can override actions by embedding DefaultActions and replacing specific methods
- [ ] **ACTION-04**: Developer can add custom validation by implementing the CreateValidator/UpdateValidator interface
- [ ] **ACTION-05**: Both HTML handlers and Huma API handlers call the same Actions interface (no business logic duplication)
- [ ] **ACTION-06**: Resource registration happens explicitly at app startup (no init() magic)
- [ ] **ACTION-07**: Action layer automatically maps database errors to forge.Error (unique violation → 409, FK violation → 400, not found → 404)

### Error Handling

- [ ] **ERR-01**: forge.Error carries HTTP status, machine-readable code, user-facing message, developer detail, and wrapped error
- [ ] **ERR-02**: HTML errors render via SSE fragment (toast) when in SSE context, or error page otherwise
- [ ] **ERR-03**: API errors render as RFC 7807 responses via Huma's error transformer
- [ ] **ERR-04**: Panic recovery middleware catches panics, logs stack traces, and returns 500 without exposing internals

### Hypermedia Layer

- [ ] **HTML-01**: Scaffolded Templ form component renders Datastar-native forms with validation error display
- [ ] **HTML-02**: Scaffolded Templ list component renders tables with sort headers, filter controls, and pagination
- [ ] **HTML-03**: Scaffolded Templ detail component renders read-only resource view
- [ ] **HTML-04**: Form primitives library provides FormField, TextInput, DecimalInput, SelectInput, RelationSelect components
- [ ] **HTML-05**: Datastar SSE helpers provide MergeFragment and Redirect operations
- [ ] **HTML-06**: Forms display field-level validation errors returned from the action layer
- [ ] **HTML-07**: Generated form components conditionally render fields based on user's role (Visibility/Mutability)

### API Layer

- [ ] **API-01**: Generated Huma structs include proper struct tags for query params, headers, and body with validation constraints
- [ ] **API-02**: Generated Huma route registration exposes CRUD endpoints under /api/v1/<resource>
- [ ] **API-03**: OpenAPI 3.1 spec served at /api/openapi.json and /api/openapi.yaml
- [ ] **API-04**: Interactive API docs served at /api/docs (Scalar UI)
- [ ] **API-05**: `forge openapi export` exports the OpenAPI spec to a file
- [ ] **API-06**: API uses cursor-based pagination by default with Link headers (RFC 8288)
- [ ] **API-07**: API supports bearer token and API key authentication
- [ ] **API-08**: CORS configuration is available via forge.toml
- [ ] **API-09**: Rate limiting middleware protects API endpoints
- [ ] **API-10**: Generated OpenAPI spec passes spectral linting and produces working SDKs

### Authentication

- [ ] **AUTH-01**: Developer can configure session-based email/password authentication
- [ ] **AUTH-02**: Developer can configure OAuth2 providers (Google, GitHub)
- [ ] **AUTH-03**: Sessions are stored in PostgreSQL (no Redis dependency)
- [ ] **AUTH-04**: API authentication supports bearer tokens and API keys (X-API-Key header)
- [x] **AUTH-05**: Generated actions check CRUD-level permissions before executing operations
- [x] **AUTH-06**: Generated actions strip invisible fields based on current user's role before returning data

### Multi-Tenancy

- [x] **TENANT-01**: Developer can configure tenant resolution via header, subdomain, or path strategy
- [x] **TENANT-02**: When TenantScoped, all generated queries include tenant_id WHERE clause automatically
- [x] **TENANT-03**: Atlas generates row-level security (RLS) policies for tenant-scoped resources
- [x] **TENANT-04**: Test factories scope data to a test tenant

### Background Jobs

- [ ] **JOBS-01**: River client is integrated as first-class citizen with transactional enqueueing
- [ ] **JOBS-02**: Schema-defined AfterCreate/AfterUpdate hooks automatically enqueue River jobs in the same DB transaction
- [ ] **JOBS-03**: Jobs carry TenantID explicitly for scoped queries in workers
- [ ] **JOBS-04**: Job queues are configurable via forge.toml (queue names, concurrency limits)

### Audit Logging

- [x] **AUDIT-01**: When Auditable is true, created_by and updated_by columns are auto-populated from authenticated user
- [x] **AUDIT-02**: When Auditable is true, changes are recorded in audit_log table with JSONB diffs of old/new values
- [x] **AUDIT-03**: No-op updates (no actual changes) do not generate audit log entries

### SSE & Real-Time

- [x] **SSE-01**: Global SSE connection limiter caps concurrent connections per process (default: 5000)
- [x] **SSE-02**: Per-user SSE connection limit prevents single-user exhaustion (default: 10)
- [x] **SSE-03**: On server shutdown, SSE connections receive close event and drain gracefully
- [x] **SSE-04**: Single shared PostgreSQL LISTEN connection fans out events to subscribers via Go channels
- [x] **SSE-05**: Backpressure: full subscriber channel buffers drop events and send refresh signal
- [x] **SSE-06**: NotifyHub interface allows swapping to Redis/NATS for connection pooler deployments

### Testing

- [ ] **TEST-01**: Generated test factories create valid resource instances with builder pattern (ProductWith.Title("..."))
- [ ] **TEST-02**: forgetest.NewTestDB provides isolated test schema with auto-cleanup
- [ ] **TEST-03**: Action layer is testable without HTTP (direct function calls)
- [ ] **TEST-04**: forgetest.NewApp provides full HTTP testing with PostDatastar helper for Datastar form submissions

### CLI

- [ ] **CLI-01**: `forge init <name>` scaffolds a new project with forge.toml, main.go, directory structure
- [ ] **CLI-02**: `forge dev` starts development server with file watching and hot reload for .go, .templ, .sql, .css files
- [ ] **CLI-03**: `forge build` compiles a single production binary with embedded templates, static assets, and migrations
- [ ] **CLI-04**: `forge db create/drop/seed/console/reset` manages the development database
- [ ] **CLI-05**: `forge routes` lists all registered routes (HTML + API)
- [ ] **CLI-06**: `forge tool sync` downloads/updates tool binaries (templ, sqlc, tailwind, atlas)
- [ ] **CLI-07**: `forge deploy` builds and packages for deployment

### Configuration & Deployment

- [ ] **DEPLOY-01**: All configuration via forge.toml, overridable by environment variables (12-factor)
- [ ] **DEPLOY-02**: Production binary is a single file < 30MB
- [ ] **DEPLOY-03**: Server starts in < 100ms cold start
- [ ] **DEPLOY-04**: Zero npm — no Node.js dependency anywhere in the toolchain
- [ ] **DEPLOY-05**: Tailwind CSS compiled via standalone CLI binary

### Observability

- [ ] **OTEL-01**: OpenTelemetry traces on HTTP requests and database queries
- [ ] **OTEL-02**: Prometheus-compatible metrics endpoint
- [ ] **OTEL-03**: Structured logging via slog with JSON format in production

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Ecosystem

- **ECO-01**: Component library (DatastarUI-style, shadcn/ui for Templ)
- **ECO-02**: Admin panel generator from schema
- **ECO-03**: Outbound webhook system
- **ECO-04**: Full-text search generation from Searchable fields using PostgreSQL tsvector
- **ECO-05**: Real-time collaboration primitives

### Advanced Features

- **ADV-01**: N:M (many-to-many) relationship support with join table generation
- **ADV-02**: Optimistic locking with version field auto-management
- **ADV-03**: File uploads with S3/local storage abstraction
- **ADV-04**: Email sending with Templ templates
- **ADV-05**: CI/CD template generation

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Multi-database support (MySQL, SQLite) | Dilutes PostgreSQL advantages; increases complexity 3x; modern SaaS defaults to PG |
| NoSQL/document store | Contradicts schema-driven philosophy; use JSONB fields within relational model instead |
| GraphQL generation | Complexity explosion, N+1 query problems; hypermedia + REST covers all use cases |
| Magic relationship detection | Fragile, breaks on refactoring; explicit schema declarations are safer |
| Cascading soft deletes by default | Data loss risk; require explicit implementation in action layer |
| Admin panel generation | One-size-fits-all rarely fits; provide composable building blocks instead |
| Client-side only validation | Security risk; always validate server-side, client-side is enhancement |
| Mobile app support | Web-first framework; single binary serves HTML + API |
| Non-Go schema definition (YAML/JSON) | Go structs provide type checking, IDE support, and are the established pattern (Ent) |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| SCHEMA-01 | Phase 1 | Pending |
| SCHEMA-02 | Phase 1 | Pending |
| SCHEMA-03 | Phase 1 | Pending |
| SCHEMA-04 | Phase 1 | Pending |
| SCHEMA-05 | Phase 1 | Pending |
| SCHEMA-06 | Phase 1 | Pending |
| SCHEMA-07 | Phase 7 | Complete |
| SCHEMA-08 | Phase 7 | Complete |
| SCHEMA-09 | Phase 8 | Pending |
| PARSE-01 | Phase 1 | Pending |
| PARSE-02 | Phase 1 | Pending |
| PARSE-03 | Phase 1 | Pending |
| GEN-01 | Phase 2 | Pending |
| GEN-02 | Phase 3 | Pending |
| GEN-03 | Phase 3 | Pending |
| GEN-04 | Phase 2 | Pending |
| GEN-05 | Phase 4 | Pending |
| GEN-06 | Phase 5 | Pending |
| GEN-07 | Phase 5 | Pending |
| GEN-08 | Phase 2 | Pending |
| GEN-09 | Phase 6 | Pending |
| GEN-10 | Phase 6 | Pending |
| GEN-11 | Phase 2 | Pending |
| GEN-12 | Phase 6 | Pending |
| GEN-13 | Phase 2 | Pending |
| MIGRATE-01 | Phase 2 | Pending |
| MIGRATE-02 | Phase 2 | Pending |
| MIGRATE-03 | Phase 2 | Pending |
| MIGRATE-04 | Phase 2 | Pending |
| MIGRATE-05 | Phase 2 | Pending |
| MIGRATE-06 | Phase 2 | Pending |
| DATA-01 | Phase 3 | Pending |
| DATA-02 | Phase 3 | Pending |
| DATA-03 | Phase 3 | Pending |
| DATA-04 | Phase 3 | Pending |
| DATA-05 | Phase 7 | Complete |
| DATA-06 | Phase 7 | Complete |
| DATA-07 | Phase 7 | Complete |
| DATA-08 | Phase 7 | Complete |
| DATA-09 | Phase 7 | Complete |
| DATA-10 | Phase 3 | Pending |
| DATA-11 | Phase 3 | Pending |
| DATA-12 | Phase 7 | Pending |
| ACTION-01 | Phase 4 | Pending |
| ACTION-02 | Phase 4 | Pending |
| ACTION-03 | Phase 4 | Pending |
| ACTION-04 | Phase 4 | Pending |
| ACTION-05 | Phase 4 | Pending |
| ACTION-06 | Phase 4 | Pending |
| ACTION-07 | Phase 4 | Pending |
| ERR-01 | Phase 4 | Pending |
| ERR-02 | Phase 4 | Pending |
| ERR-03 | Phase 4 | Pending |
| ERR-04 | Phase 4 | Pending |
| HTML-01 | Phase 6 | Pending |
| HTML-02 | Phase 6 | Pending |
| HTML-03 | Phase 6 | Pending |
| HTML-04 | Phase 6 | Pending |
| HTML-05 | Phase 6 | Pending |
| HTML-06 | Phase 6 | Pending |
| HTML-07 | Phase 6 | Pending |
| API-01 | Phase 5 | Pending |
| API-02 | Phase 5 | Pending |
| API-03 | Phase 5 | Pending |
| API-04 | Phase 5 | Pending |
| API-05 | Phase 5 | Pending |
| API-06 | Phase 5 | Pending |
| API-07 | Phase 5 | Pending |
| API-08 | Phase 5 | Pending |
| API-09 | Phase 5 | Pending |
| API-10 | Phase 5 | Pending |
| AUTH-01 | Phase 6 | Pending |
| AUTH-02 | Phase 6 | Pending |
| AUTH-03 | Phase 6 | Pending |
| AUTH-04 | Phase 5 | Pending |
| AUTH-05 | Phase 7 | Complete |
| AUTH-06 | Phase 7 | Complete |
| TENANT-01 | Phase 7 | Complete |
| TENANT-02 | Phase 7 | Complete |
| TENANT-03 | Phase 7 | Complete |
| TENANT-04 | Phase 7 | Complete |
| JOBS-01 | Phase 8 | Pending |
| JOBS-02 | Phase 8 | Pending |
| JOBS-03 | Phase 8 | Pending |
| JOBS-04 | Phase 8 | Pending |
| AUDIT-01 | Phase 7 | Complete |
| AUDIT-02 | Phase 7 | Complete |
| AUDIT-03 | Phase 7 | Complete |
| SSE-01 | Phase 8 | Complete |
| SSE-02 | Phase 8 | Complete |
| SSE-03 | Phase 8 | Complete |
| SSE-04 | Phase 8 | Complete |
| SSE-05 | Phase 8 | Complete |
| SSE-06 | Phase 8 | Complete |
| TEST-01 | Phase 6 | Pending |
| TEST-02 | Phase 6 | Pending |
| TEST-03 | Phase 6 | Pending |
| TEST-04 | Phase 6 | Pending |
| CLI-01 | Phase 1 | Pending |
| CLI-02 | Phase 2 | Pending |
| CLI-03 | Phase 8 | Pending |
| CLI-04 | Phase 3 | Pending |
| CLI-05 | Phase 5 | Pending |
| CLI-06 | Phase 1 | Pending |
| CLI-07 | Phase 8 | Pending |
| DEPLOY-01 | Phase 8 | Pending |
| DEPLOY-02 | Phase 8 | Pending |
| DEPLOY-03 | Phase 8 | Pending |
| DEPLOY-04 | Phase 1 | Pending |
| DEPLOY-05 | Phase 6 | Pending |
| OTEL-01 | Phase 8 | Pending |
| OTEL-02 | Phase 8 | Pending |
| OTEL-03 | Phase 8 | Pending |

**Coverage:**
- v1 requirements: 89 total
- Mapped to phases: 89
- Unmapped: 0

---
*Requirements defined: 2026-02-16*
*Last updated: 2026-02-16 after roadmap creation*
