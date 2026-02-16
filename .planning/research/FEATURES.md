# Feature Research: Schema-Driven Go Web Framework

**Domain:** Schema-driven web framework for Go SaaS applications
**Researched:** 2026-02-16
**Confidence:** MEDIUM-HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Database migrations** | Every scaffold generator creates migrations (Rails, Django, Laravel, Ent) | MEDIUM | Must support up/down, dependency ordering, rollback safety |
| **CRUD operations** | Core promise of code generation frameworks | MEDIUM | Create, Read, Update, Delete for every resource |
| **Model/struct generation** | Type-safe Go structs from schema | LOW | Ent and GORM both auto-generate structs |
| **Validation rules** | Rails/Laravel/Django all generate validations from schema | MEDIUM | Must include field-level (type, length, format) and custom business rules |
| **Route generation** | Rails `resources :posts` pattern expected | LOW | RESTful routes for CRUD operations |
| **Form components** | Rails scaffolds create form views automatically | MEDIUM | For Forge: Templ components for create/edit forms |
| **List/index views** | Every scaffold includes list view with table | MEDIUM | Display all records, clickable to detail/edit |
| **Relationships** | Foreign keys, has_many, belongs_to patterns | HIGH | Must handle 1:1, 1:N, N:M with proper cascade rules |
| **Basic filtering** | List views need basic search/filter | MEDIUM | At minimum: text search on key fields |
| **Pagination** | Required for list views (standard practice 2026) | LOW | Server-side pagination is non-negotiable |
| **Type-safe queries** | Go developers expect compile-time safety | HIGH | Ent/sqlc pattern: queries checked at compile time |
| **Timestamps** | created_at, updated_at auto-managed | LOW | Every ORM provides this automatically |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Dual interface (hypermedia + API)** | One schema generates both HTML and JSON responses | HIGH | Forge's core differentiator: HTMX + REST API from same action layer |
| **Action layer abstraction** | Business logic shared between web and API | MEDIUM | Prevents duplication between controllers and API handlers |
| **OpenAPI auto-generation** | API documentation from schema (FastAPI/Laravel pattern) | MEDIUM | Swagger UI at /docs endpoint, always in sync |
| **Real-time capabilities** | WebSocket support for live updates | HIGH | SaaS apps in 2026 expect real-time (chat, dashboards, notifications) |
| **Zero npm philosophy** | No JavaScript build step required | LOW | Templ + HTMX means no frontend toolchain |
| **PostgreSQL-native** | Leverage PG-specific features (JSONB, arrays, full-text search) | MEDIUM | Not trying to be database-agnostic = better features |
| **Relationship cascade visualization** | Show what will be deleted/updated when cascading | MEDIUM | Prevents accidental data loss, unique safety feature |
| **Bulk operations** | Batch insert/update/delete without N+1 queries | MEDIUM | EF Core 7+ pattern: ExecuteDelete/ExecuteUpdate for performance |
| **Soft delete support** | Schema flag generates soft-delete behavior | LOW | `deleted_at` field auto-managed, queries auto-filtered |
| **Audit trail generation** | Who/when/what tracking from schema annotation | MEDIUM | Rails/Laravel pattern: auto-track changes for compliance |
| **Seeding from schema** | Generate realistic test data based on field types | MEDIUM | Faker integration: email field → fake emails, etc. |
| **Advanced filtering** | Query builder for complex filters (age > 18 AND status = active) | HIGH | Go beyond basic search: range, multi-select, date ranges |
| **Sorting on any column** | Click column headers to sort | LOW | Standard practice, but must work with pagination |
| **Field-level permissions** | Schema annotations define who can read/write fields | HIGH | Enterprise feature: PII fields hidden from non-admins |
| **Optimistic locking** | Prevent concurrent update conflicts | MEDIUM | Version field auto-managed, stale update detection |
| **Custom actions beyond CRUD** | `approve`, `archive`, `publish` actions on resources | MEDIUM | Rails-style member/collection actions |
| **Nested resource routing** | `/posts/:post_id/comments` patterns | MEDIUM | Rails nested resources, common for hierarchical data |
| **Relationship preloading hints** | Avoid N+1 with schema-driven eager loading | MEDIUM | Ent supports this: auto-detect needed joins |
| **Generated tests** | Basic CRUD tests auto-created | MEDIUM | Buffalo pattern: test templates for generated code |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Multi-database support** | "What if users want MySQL/SQLite?" | Dilutes PG advantages, increases complexity 3x, testing burden | **PostgreSQL-only**: Document this trade-off clearly. Modern SaaS = PG (Heroku, Render, Supabase all default to PG) |
| **NoSQL/document store** | "JSON flexibility sounds good" | Schema-driven framework contradicts schemaless DB | **JSONB fields**: Use PG's JSONB for flexible fields within relational model |
| **Client-side validation only** | Faster UX, no server round-trip | Security risk, can be bypassed | **Both client and server**: Templ components validate client-side, Go validates server-side |
| **Magic relationships** | Auto-detect relationships without declaration | Fragile, breaks on refactoring, no type safety | **Explicit schema definitions**: Developer declares relationships clearly |
| **Infinite scroll** | "More modern than pagination" | Difficult back button, poor accessibility, pagination still needed for API | **Smart pagination**: Offset-based with optional cursor for APIs |
| **Real-time everything** | "Make all views live-update" | WebSocket overhead for rarely-changing data, connection management complexity | **Selective real-time**: Annotate schema fields/resources that need live updates |
| **Auto-increment IDs** | Familiar pattern from MySQL | Enumeration attacks, not distributed-friendly | **UUIDs or BigInt**: PG supports both, UUIDs hide record count |
| **Cascading deletes by default** | Convenience, fewer orphans | Accidental data loss when relationships misunderstood | **Explicit cascade rules**: Require developer to choose RESTRICT/CASCADE/SET NULL per relationship |
| **Admin panel generation** | Rails Admin / Django Admin pattern | One-size-fits-all rarely fits, customization becomes harder than building from scratch | **Provide building blocks**: Generate list/form components developers compose into custom admin |
| **Multi-tenancy magic** | Auto-scope queries by tenant_id | Wrong abstraction level (schema vs app logic), breaks for complex tenancy models | **Middleware pattern**: Provide tenant-scoping middleware developers opt into |
| **GraphQL generation** | "More flexible than REST" | Complexity explosion, N+1 query problems, frontend tooling required | **Hypermedia + REST**: HTMX for web, clean REST for API consumers |
| **Complex authorization DSL** | Pundit/CanCanCan style policy language | Another language to learn, debugging difficulty | **Go functions**: `CanUserEditPost(user, post) bool` is clearer than DSL |

## Feature Dependencies

```
Database Migrations
    └──requires──> Model/Struct Generation
                       └──requires──> CRUD Operations
                                          └──requires──> Route Generation
                                                             └──requires──> Form Components
                                                                                └──requires──> List Views

Relationships
    └──requires──> Foreign Key Migrations
    └──enhances──> Type-Safe Queries (join support)
    └──enables──> Nested Resource Routing

Validation Rules
    └──requires──> Model/Struct Generation
    └──enhances──> Form Components (client-side validation)
    └──required-by──> OpenAPI Generation (schema constraints)

Soft Delete
    └──requires──> Timestamps (deleted_at field)
    └──conflicts──> Hard Delete (can't have both for same resource)
    └──enhances──> Audit Trail (deletion events tracked)

Real-time Capabilities
    └──requires──> Action Layer Abstraction (broadcast from actions)
    └──enhances──> List Views (auto-refresh on changes)
    └──requires──> WebSocket infrastructure

Bulk Operations
    └──requires──> Type-Safe Queries
    └──enhances──> List Views (select multiple → delete selected)

OpenAPI Generation
    └──requires──> Route Generation (endpoints to document)
    └──requires──> Validation Rules (schema constraints)
    └──requires──> Model/Struct Generation (request/response types)

Pagination
    └──required-by──> Advanced Filtering (paginate filtered results)
    └──required-by──> Sorting (sort paginated results)

Field-Level Permissions
    └──requires──> Authentication/Authorization (user context)
    └──enhances──> Form Components (hide unauthorized fields)
    └──enhances──> List Views (hide unauthorized columns)
```

## MVP Definition

### Launch With (v1.0)

Minimum viable product — what's needed to validate the concept.

- [x] **Database migrations** — Can't build SaaS without schema evolution
- [x] **Model/struct generation** — Core value: Go types from schema
- [x] **Type-safe queries** — The Go advantage over Ruby/Python
- [x] **CRUD operations** — Basic create/read/update/delete
- [x] **Route generation** — RESTful routing automatically configured
- [x] **Form components** — Templ-based create/edit forms
- [x] **List/index views** — Table view of all records
- [x] **Validation rules** — Field-level validation (required, length, format)
- [x] **Relationships (basic)** — 1:1 and 1:N foreign keys with cascade options
- [x] **Timestamps** — Auto-managed created_at/updated_at
- [x] **Pagination** — Server-side pagination for list views
- [x] **Basic filtering** — Text search on key fields
- [x] **Dual interface** — Same schema → HTML + JSON (core differentiator)

**Why these features:**
- Matches Rails `scaffold` and Laravel `make:model --all` core functionality
- Demonstrates schema-drives-everything philosophy
- Proves dual interface (hypermedia + API) works
- Enough to build a simple CRUD SaaS app end-to-end

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] **OpenAPI auto-generation** — Trigger: First API consumer requests docs
- [ ] **Soft delete support** — Trigger: User requests "don't lose data on delete"
- [ ] **Sorting on columns** — Trigger: User wants to reorder list views
- [ ] **Advanced filtering** — Trigger: Basic search insufficient for use case
- [ ] **Bulk operations** — Trigger: "Delete selected" feature requested
- [ ] **Seeding from schema** — Trigger: Developer needs test data
- [ ] **Relationships (advanced)** — Trigger: N:M (many-to-many) needed
- [ ] **Nested resource routing** — Trigger: `/posts/:id/comments` pattern needed
- [ ] **Custom actions** — Trigger: Business logic beyond CRUD (approve/publish)
- [ ] **Audit trail generation** — Trigger: Compliance requirement or data debugging
- [ ] **Generated tests** — Trigger: Team wants test coverage baseline

### Future Consideration (v2.0+)

Features to defer until product-market fit is established.

- [ ] **Real-time capabilities** — Complex, defer until validated demand
- [ ] **Field-level permissions** — Enterprise feature, not needed for early adopters
- [ ] **Optimistic locking** — Edge case, address when concurrency issues surface
- [ ] **Relationship preloading hints** — Performance optimization, premature for v1
- [ ] **Relationship cascade visualization** — Safety feature, add when users report accidental deletes

**Why defer:**
- Real-time adds WebSocket complexity (separate connection management)
- Field permissions need auth system maturity
- Performance optimizations should come after profiling real usage
- Safety visualizations valuable but not blocking

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority | Phase |
|---------|------------|---------------------|----------|-------|
| Database migrations | HIGH | MEDIUM | P1 | v1.0 |
| Model/struct generation | HIGH | MEDIUM | P1 | v1.0 |
| Type-safe queries | HIGH | HIGH | P1 | v1.0 |
| CRUD operations | HIGH | MEDIUM | P1 | v1.0 |
| Route generation | HIGH | LOW | P1 | v1.0 |
| Form components | HIGH | MEDIUM | P1 | v1.0 |
| List views | HIGH | MEDIUM | P1 | v1.0 |
| Validation rules | HIGH | MEDIUM | P1 | v1.0 |
| Basic relationships | HIGH | HIGH | P1 | v1.0 |
| Timestamps | HIGH | LOW | P1 | v1.0 |
| Pagination | HIGH | LOW | P1 | v1.0 |
| Basic filtering | HIGH | MEDIUM | P1 | v1.0 |
| Dual interface | HIGH | HIGH | P1 | v1.0 |
| OpenAPI generation | MEDIUM | MEDIUM | P2 | v1.x |
| Soft delete | MEDIUM | LOW | P2 | v1.x |
| Sorting | MEDIUM | LOW | P2 | v1.x |
| Advanced filtering | MEDIUM | HIGH | P2 | v1.x |
| Bulk operations | MEDIUM | MEDIUM | P2 | v1.x |
| Seeding | MEDIUM | MEDIUM | P2 | v1.x |
| N:M relationships | MEDIUM | HIGH | P2 | v1.x |
| Nested routing | MEDIUM | MEDIUM | P2 | v1.x |
| Custom actions | MEDIUM | MEDIUM | P2 | v1.x |
| Audit trail | MEDIUM | MEDIUM | P2 | v1.x |
| Generated tests | LOW | MEDIUM | P2 | v1.x |
| Real-time | MEDIUM | HIGH | P3 | v2.0+ |
| Field permissions | LOW | HIGH | P3 | v2.0+ |
| Optimistic locking | LOW | MEDIUM | P3 | v2.0+ |
| Preloading hints | LOW | MEDIUM | P3 | v2.0+ |
| Cascade visualization | MEDIUM | MEDIUM | P3 | v2.0+ |

**Priority key:**
- P1: Must have for launch (competitive with Rails scaffold, Laravel make:model)
- P2: Should have, add when validated (competitive with Django Admin, Ent features)
- P3: Nice to have, future consideration (competitive with enterprise frameworks)

## Competitor Feature Analysis

| Feature | Rails Scaffold | Django Admin | Laravel Artisan | Ent (Go) | Buffalo (Go) | Forge Approach |
|---------|---------------|--------------|-----------------|----------|--------------|----------------|
| **Migrations** | ✅ Full up/down | ✅ Auto-generated | ✅ Full up/down | ✅ Auto-generated | ✅ Via Pop | **✅ PG-native migrations** |
| **Model generation** | ✅ ActiveRecord | ✅ Django ORM | ✅ Eloquent | ✅ Code-gen structs | ✅ Pop models | **✅ Type-safe Go structs** |
| **CRUD** | ✅ Full scaffold | ✅ Admin interface | ✅ Resource controllers | ⚠️ Manual controllers | ✅ Resource generator | **✅ Action layer (shared)** |
| **Forms** | ✅ ERB templates | ✅ ModelForm auto-gen | ✅ Blade templates | ❌ Manual | ✅ Plush templates | **✅ Templ components** |
| **Validation** | ✅ Model validations | ✅ Form + model | ✅ Request validation | ⚠️ Manual | ⚠️ Manual | **✅ Schema-driven** |
| **API + Web** | ⚠️ Separate controllers | ⚠️ DRF separate | ⚠️ Separate | ❌ API only | ⚠️ Separate | **✅ Unified dual interface** |
| **Type safety** | ❌ Runtime only | ❌ Runtime only | ❌ Runtime only | ✅ Compile-time | ⚠️ Partial | **✅ Full compile-time** |
| **OpenAPI** | ⚠️ Via gems | ⚠️ Via DRF | ⚠️ Via packages | ⚠️ Manual | ❌ None | **✅ Auto-generated** |
| **Real-time** | ⚠️ Action Cable | ⚠️ Channels | ⚠️ Broadcasting | ❌ Manual | ⚠️ Manual | **🔮 v2.0 feature** |
| **Relationships** | ✅ Full ORM support | ✅ Full ORM support | ✅ Full ORM support | ✅ Graph traversal | ✅ Pop associations | **✅ Schema-defined edges** |
| **Soft delete** | ⚠️ Via gems | ⚠️ Via packages | ⚠️ Via traits | ❌ Manual | ❌ Manual | **✅ Schema flag** |
| **Audit trail** | ⚠️ Paper Trail gem | ⚠️ django-auditlog | ⚠️ Via packages | ❌ Manual | ❌ Manual | **✅ Schema annotation** |
| **Seeding** | ✅ seeds.rb | ✅ Fixtures/factories | ✅ Seeders | ❌ Manual | ✅ Pop seeding | **✅ Schema-based faker** |
| **Tests** | ✅ Auto-generated | ✅ Auto-generated | ⚠️ Partial | ❌ Manual | ✅ Test templates | **✅ CRUD test scaffold** |

**Key observations:**
- **Rails/Laravel** have most complete feature sets, but no type safety
- **Ent** has great type safety + code gen, but no web UI generation
- **Buffalo** closest to Forge vision but lacks schema-driven approach
- **Django Admin** excellent for admin panels, but one-size-fits-all
- **Forge differentiator**: Type-safe Go + dual interface + schema-drives-everything

## Developer Expectations for `forge generate resource`

Based on Rails, Laravel, and Buffalo patterns, developers expect:

### Command Syntax
```bash
forge generate resource Post title:string body:text published:bool author:references
```

### What Gets Created
1. **Schema file**: `schema/post.go` (or YAML/JSON if using declarative)
2. **Migration**: `migrations/20260216_create_posts.sql`
3. **Model**: `models/post.go` with type-safe struct
4. **Queries**: `queries/post.go` with CRUD + custom queries
5. **Actions**: `actions/posts.go` with Create/Read/Update/Delete logic
6. **Handlers**:
   - `handlers/web/posts.go` (HTMX/Templ responses)
   - `handlers/api/posts.go` (JSON responses)
7. **Templates**:
   - `templates/posts/index.templ` (list view)
   - `templates/posts/form.templ` (create/edit form)
   - `templates/posts/show.templ` (detail view)
8. **Routes**: Updated `routes.go` with:
   ```go
   // Web routes (hypermedia)
   r.Get("/posts", handlers.web.PostsIndex)
   r.Get("/posts/{id}", handlers.web.PostsShow)
   r.Get("/posts/new", handlers.web.PostsNew)
   r.Post("/posts", handlers.web.PostsCreate)
   r.Get("/posts/{id}/edit", handlers.web.PostsEdit)
   r.Put("/posts/{id}", handlers.web.PostsUpdate)
   r.Delete("/posts/{id}", handlers.web.PostsDelete)

   // API routes (JSON)
   r.Get("/api/posts", handlers.api.PostsIndex)
   r.Get("/api/posts/{id}", handlers.api.PostsShow)
   r.Post("/api/posts", handlers.api.PostsCreate)
   r.Put("/api/posts/{id}", handlers.api.PostsUpdate)
   r.Delete("/api/posts/{id}", handlers.api.PostsDelete)
   ```
9. **Tests**: `tests/posts_test.go` with basic CRUD test cases
10. **OpenAPI**: Updated `openapi.yaml` with new endpoints

### Expected Flags
```bash
forge generate resource Post [fields...] [options]

Options:
  --no-web          Skip web handlers and templates (API only)
  --no-api          Skip API handlers (web only)
  --no-tests        Skip test generation
  --soft-delete     Add deleted_at field with soft-delete behavior
  --audit           Add audit trail (created_by, updated_by, etc.)
  --timestamps      Add created_at, updated_at (default: true)
  --no-migration    Generate code but skip migration
  --parent Resource Parent resource for nested routing
```

### Expected Output
```
      create  schema/post.go
      create  migrations/20260216120000_create_posts.sql
      create  models/post.go
      create  queries/post.go
      create  actions/posts.go
      create  handlers/web/posts.go
      create  handlers/api/posts.go
      create  templates/posts/index.templ
      create  templates/posts/form.templ
      create  templates/posts/show.templ
      update  routes.go
      create  tests/posts_test.go
      update  openapi.yaml

✅ Post resource created successfully!

Next steps:
  1. Review the generated migration: migrations/20260216120000_create_posts.sql
  2. Run migration: forge db migrate
  3. Start server: forge dev
  4. Visit: http://localhost:3000/posts
```

## What Users Compare Against

### Coming from Rails
- Expect: `rails generate scaffold` level of completeness
- Want: Type safety Rails lacks
- Fear: Losing Rails' "magic" productivity

### Coming from Laravel
- Expect: `php artisan make:model --all` feature parity
- Want: Better performance than PHP
- Fear: Losing Eloquent's relationship elegance

### Coming from Django
- Expect: Django Admin level of polish for admin interfaces
- Want: Faster execution than Python
- Fear: More boilerplate than Django

### Coming from Node/TypeScript
- Expect: Prisma-level schema DX + type generation
- Want: Simpler deployment (single binary vs npm)
- Fear: Less mature ecosystem than npm

### Coming from Go ecosystem (Ent/GORM/Buffalo)
- Expect: Current Go patterns (context.Context, errors, interfaces)
- Want: Less boilerplate, faster CRUD development
- Fear: Too much magic, not idiomatic Go

## Confidence Assessment

| Area | Level | Source | Notes |
|------|-------|--------|-------|
| Table stakes features | HIGH | Rails, Django, Laravel official docs + multiple web sources | Well-established patterns across 3+ major frameworks |
| Go framework features | MEDIUM | Ent/Buffalo GitHub + community articles | Less mature ecosystem, fewer authoritative sources |
| Differentiator viability | MEDIUM | HTMX 2026 trends, FastAPI OpenAPI patterns | Based on 2026 web search trends showing hypermedia resurgence |
| Developer expectations | MEDIUM-HIGH | Rails/Laravel official docs + generator comparison articles | Clear patterns from established frameworks |
| Anti-features | MEDIUM | Architecture articles, post-mortems | Based on common pitfalls, less authoritative sources |
| Dual interface approach | MEDIUM | HTMX articles, architecture discussions | Novel combination, limited prior art |

## Sources

### Framework Documentation
- [Rails Scaffolding Guide](https://www.railscarma.com/blog/scaffolding-in-ruby-on-rails-complete-guide/)
- [Rails Generators Official Docs](https://guides.rubyonrails.org/generators.html)
- [Laravel Eloquent Documentation](https://laravel.com/docs/12.x/eloquent)
- [Laravel Make:Model Guide](https://laraveldaily.com/lesson/laravel-eloquent-expert/make-model-options)
- [Django CRUD Generators](https://crudgen.pro/django/)
- [Ent Framework GitHub](https://github.com/ent/ent)
- [Buffalo Framework](https://gobuffalo.io/en/)
- [GORM Migrations](https://gorm.io/docs/migration.html)

### Schema-Driven Development
- [Schema-Driven Development Guide](https://blog.noclocks.dev/schema-driven-development-and-single-source-of-truth-essential-practices-for-modern-developers)
- [Godspeed Systems: Schema-Driven Development](https://godspeed.systems/blog/schema-driven-development-and-single-source-of-truth)
- [webrpc: Schema-Driven Backend Services](https://github.com/webrpc/webrpc)

### HTMX and Hypermedia (2026)
- [HTMX in 2026: Why Hypermedia is Dominating](https://vibe.forem.com/del_rosario/htmx-in-2026-why-hypermedia-is-dominating-the-modern-web-41id)
- [HTMX Complete Guide 2026](https://devtoolbox.dedyn.io/blog/htmx-complete-guide)
- [HTMX vs React for Django Apps 2026](https://medium.com/@yogeshkrishnanseeniraj/htmx-in-2026-why-hypermedia-is-beating-react-for-faster-django-apps-48fd4adb43b2)

### Validation and OpenAPI
- [Form Validation Best Practices](https://ivyforms.com/blog/form-validation-best-practices/)
- [OWASP Input Validation](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html)
- [FastAPI OpenAPI Generation 2026](https://oneuptime.com/blog/post/2026-02-02-fastapi-openapi-documentation/view)
- [OpenAPI Auto-Generation from Code](https://codehooks.io/blog/auto-generate-openapi-docs-from-code)

### Real-Time Features
- [Real-Time Features in SaaS: WebSockets vs Pub/Sub](https://medium.com/@beta_49625/real-time-features-in-saas-websockets-pub-sub-and-when-to-use-them-83e8a447e36f)
- [WebSockets Complete Guide 2026](https://devtoolbox.dedyn.io/blog/websocket-complete-guide)
- [Building Real-Time APIs 2026](https://dasroot.net/posts/2026/01/building-real-time-apis-webscokets-sse-webrtc/)

### Type-Safe Query Builders (Go)
- [Jet: Type-Safe SQL Builder](https://github.com/go-jet/jet)
- [sqlc: Type-Safe Database Access](https://oneuptime.com/blog/post/2026-01-07-go-sqlc-type-safe-database/view)
- [Golang ORM Comparison 2025](https://www.bytebase.com/blog/golang-orm-query-builder/)

### Database Operations
- [REST API Pagination, Filtering, Sorting Best Practices](https://www.moesif.com/blog/technical/api-design/REST-API-Design-Filtering-Sorting-and-Pagination/)
- [EF Core Bulk Extensions](https://github.com/borisdj/EFCore.BulkExtensions)
- [Soft Deletes and Audit Trails](https://mykeels.medium.com/audit-trail-and-soft-deletes-in-ef-core-841f57a9096b)
- [Database Seeding Best Practices 2026](https://oneuptime.com/blog/post/2026-02-03-laravel-database-seeding/view)

### SaaS Development Trends
- [SaaS Development in 2026: Features & Stack](https://www.tvlitsolutions.com/saas-development-in-2026-features-stack-architecture/)
- [SaaS Development Frameworks 2026](https://www.thefrontendcompany.com/posts/saas-development-framework)

---
*Feature research for: Forge - Schema-Driven Go Web Framework*
*Researched: 2026-02-16*
