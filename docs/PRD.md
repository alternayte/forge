# Forge — Product Requirements Document (v2)

**A schema-driven, hypermedia-native Go framework for building production SaaS applications.**

**Version:** 0.2.0 (Revised)
**Date:** February 15, 2026
**Author:** Nathan
**Status:** RFC — Revised after architectural review

---

## 1. Executive Summary

Forge is a Go web framework for building server-rendered SaaS applications with real-time capabilities. A single resource schema definition generates migration SQL, type-safe queries, validation rules, Templ form/list components, OpenAPI documentation, and route handlers — eliminating the manual synchronization that plagues the Go ecosystem.

The core philosophy: **Convention where it helps. Escape hatches where it matters. Schema drives everything.**

This revision addresses critical architectural concerns from the v1 review: migration diffing is delegated to Atlas, the override model is made explicit, Huma integration is unified through a shared action layer, SSE resource management is bounded, soft delete edge cases are handled, and the roadmap is restructured into shippable milestones.

---

## 2. Problem Statement

### 2.1 The Gap in the Go Ecosystem

Go has excellent building blocks (net/http, pgx, Templ, Datastar, SQLC, Huma, River) but no cohesive framework that integrates them into a productive SaaS development experience. Developers are forced to write migrations, then SQLC queries, then validation, then Templ forms that manually map to all of the above — repeating this for every resource. Pagination, filtering, sorting, multi-tenancy, form handling, and OpenAPI documentation are all left as exercises for the reader.

### 2.2 What Andurel Gets Right (and Wrong)

Andurel demonstrated the right stack selection (Go + Templ + Datastar + SQLC + PostgreSQL + River) but has architectural limitations:

**Right decisions:** Convention-over-configuration, CLI code generation, Datastar over HTMX, PostgreSQL commitment, River for background jobs, type safety via SQLC + Templ.

**Key gaps:** Echo dependency instead of Go 1.22+ stdlib, SQLC-only data access (can't do dynamic queries), no schema-as-source-of-truth, no form handling or validation, no pagination/filtering primitives, no multi-tenancy, no OpenAPI generation, npm-dependent asset pipeline, deep Rails-style directory nesting.

### 2.3 Why Huma

Huma is the best OpenAPI-from-code framework in Go. It derives OpenAPI 3.1 specs from Go struct definitions with automatic validation, content negotiation, and RFC 7807 error handling — the same pattern .NET uses with minimal API + Swashbuckle. Forge should not reinvent OpenAPI generation. It should generate Huma-compatible structs from the schema, and use Huma as the API layer.

The challenge is integrating Huma cleanly alongside the hypermedia (Datastar) layer without creating two incompatible HTTP stacks. This PRD addresses that with a shared action layer (Section 6).

---

## 3. Design Principles

1. **Schema is the single source of truth.** All derived artifacts (SQL, Go types, templates, OpenAPI) are generated from the resource schema.

2. **Go stdlib is the foundation.** HTTP handling uses Go 1.22+ `net/http`. Middleware is `func(http.Handler) http.Handler`. No router framework dependency.

3. **Dual interface: Hypermedia + API.** Every resource can expose both a Datastar-powered HTML interface and a Huma-powered JSON API. Both share a common action layer; they differ only in transport.

4. **Progressive complexity.** Simple CRUD works with zero custom code. Complex business logic is added by implementing specific interfaces, not by abandoning the framework.

5. **Zero npm.** Go binaries + Tailwind standalone CLI. No Node.js.

6. **PostgreSQL is the platform.** Data, job queues (River), sessions, caching — one database.

7. **Escape hatches are first-class.** Raw SQL via SQLC, custom handlers, custom Templ components. The framework makes the common case trivial and the uncommon case possible.

8. **Don't build what exists.** Use Atlas for migrations, Bob for query building, Huma for OpenAPI, River for jobs. Forge's value is the schema-to-everything pipeline and the integration glue — not reimplementing solved problems.

---

## 4. Architecture

### 4.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│                     Forge CLI                       │
│    (init, generate, migrate, dev, build, deploy)    │
└──────────────────────┬──────────────────────────────┘
                       │ codegen (go/ast parsing)
                       ▼
┌─────────────────────────────────────────────────────┐
│              Schema Definitions                     │
│       (Go structs in schema.go — zero gen/ imports) │
└──────┬──────────┬──────────┬────────────────────────┘
       │          │          │
  ┌────▼─────┐ ┌──▼────┐ ┌──▼───────┐
  │ Atlas    │ │ Bob   │ │ Huma     │
  │ desired  │ │ query │ │ OpenAPI  │
  │ state SQL│ │builder│ │ structs  │
  └──────────┘ └───────┘ └──────────┘
       │          │          │
  ┌────▼─────┐ ┌──▼────┐ ┌──▼───────┐
  │Validation│ │ Templ │ │ Route    │
  │  Rules   │ │ Views │ │ Registry │
  └──────────┘ └───────┘ └──────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│                 Action Layer                        │
│  (framework-agnostic business logic per resource)   │
│  Shared by both Huma API handlers and HTML handlers │
└────────┬──────────────────────────┬─────────────────┘
         │                          │
   ┌─────▼──────┐            ┌──────▼──────┐
   │  Datastar  │            │   Huma v2   │
   │  HTML/SSE  │            │   JSON/API  │
   │ (net/http) │            │  (OpenAPI)  │
   └────────────┘            └─────────────┘
         │                          │
   ┌─────▼──────┐ ┌──────────┐ ┌───▼─────────┐
   │   Templ    │ │  River   │ │ PostgreSQL  │
   │  (views)   │ │  (jobs)  │ │   (pgx)     │
   └────────────┘ └──────────┘ └─────────────┘
```

### 4.2 Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | Go 1.23+ | Single binary, fast compilation, goroutines for SSE |
| HTTP | `net/http` (stdlib) + thin helpers | No framework lock-in; Go 1.22 mux is sufficient |
| API layer | Huma v2 (via `humachi` adapter on stdlib) | Best OpenAPI 3.1 generation in Go; router-agnostic |
| Hypermedia | Datastar v1 | SSE-native; signals + DOM morphing; replaces HTMX + Alpine.js in ~15KB |
| Templates | Templ | Compile-time type-safe HTML; Go function = component |
| Database | PostgreSQL (pgx v5) | Transactional jobs, JSONB, full-text search, row-level security |
| Query builder | Bob (stephenafamo/bob) | Schema-generated type-safe dynamic queries |
| Raw queries | SQLC | Escape hatch for complex/custom SQL |
| Migrations | Atlas (ariga/atlas) | Declarative schema diffing; battle-tested; handles edge cases we never will |
| Background jobs | River | Transactional enqueueing; PostgreSQL-backed; Go generics |
| Sessions | PostgreSQL-backed (scs) | No Redis dependency; single-DB philosophy |
| CSS | Tailwind CSS (standalone CLI) | No npm; single binary downloaded by Forge CLI |
| Telemetry | OpenTelemetry | Traces, metrics, structured logging via slog |
| Testing | Go stdlib + generated factories | Test factories from schema |

### 4.3 The Bootstrapping Problem (and Solution)

Schema definitions are Go code that must be parseable *before* `gen/` exists. The constraint:

**Schema packages (`resources/*/schema.go`) must have zero imports from `gen/`.** They import only `github.com/forgego/forge/schema`.

`forge generate` does NOT compile-and-execute schema files. It parses them using `go/ast` to extract the `schema.Define(...)` calls, their field definitions, and their options. This is the same approach Ent uses with `entc` — it works because the schema DSL is deliberately constrained to be statically analyzable.

If a schema definition uses dynamic values (function calls, variables) that `go/ast` can't resolve, the generator errors with a clear message pointing to the offending line.

---

## 5. Schema System

### 5.1 Schema Definition

```go
package product

import "github.com/forgego/forge/schema"

var Resource = schema.Define("Product", schema.Options{
    Table:        "products",
    SoftDelete:   true,
    Auditable:    true,
    Searchable:   []string{"Title", "Description"},
    TenantScoped: true,
},
    schema.UUID("ID").PrimaryKey(),
    schema.String("Title").Required().MaxLen(200).MinLen(3).
        Sortable().Filterable().Label("Product Title"),
    schema.Text("Description").Optional(),
    schema.Enum("Status", "draft", "active", "archived").
        Default("draft").Filterable(),
    schema.Decimal("Price").Required().Min(0).Precision(10, 2).
        Sortable().Filterable(),
    schema.Int("StockQuantity").Required().Min(0).Default(0),
    schema.String("SKU").Required().Unique().MaxLen(50).Filterable(),
    schema.Bool("Featured").Default(false).Filterable(),
    schema.JSON("Metadata").Optional(),
    schema.BelongsTo("Category", "categories").Optional().
        OnDelete(schema.SetNull),
    schema.HasMany("Variants", "product_variants"),
    schema.Timestamps(),
)
```

### 5.2 Field Types

| Schema Type | PostgreSQL Type | Go Type | OpenAPI Type |
|-------------|----------------|---------|--------------|
| `UUID` | `uuid` | `uuid.UUID` | `string (format: uuid)` |
| `String` | `varchar(n)` | `string` | `string (maxLength: n)` |
| `Text` | `text` | `string` | `string` |
| `Int` | `integer` | `int32` | `integer` |
| `BigInt` | `bigint` | `int64` | `integer (format: int64)` |
| `Decimal` | `numeric(p,s)` | `decimal.Decimal` | `string (format: decimal)` |
| `Bool` | `boolean` | `bool` | `boolean` |
| `DateTime` | `timestamptz` | `time.Time` | `string (format: date-time)` |
| `Date` | `date` | `time.Time` | `string (format: date)` |
| `Enum` | `text + CHECK` | `string` (typed const) | `string (enum: [...])` |
| `JSON` | `jsonb` | `json.RawMessage` or typed struct | `object` |
| `Slug` | `varchar + UNIQUE` | `string` | `string` |
| `Email` | `varchar(320)` | `string` | `string (format: email)` |
| `URL` | `text` | `string` | `string (format: uri)` |

### 5.3 Field Modifiers

```go
schema.String("Title").
    Required().              // NOT NULL + validation
    MaxLen(200).             // VARCHAR(200) + validation
    MinLen(3).               // validation only
    Sortable().              // generates ORDER BY support
    Filterable().            // generates WHERE support (eq, contains, etc.)
    Searchable().            // included in full-text search index
    Unique().                // UNIQUE constraint (partial index if SoftDelete)
    Index().                 // B-tree index
    Default("untitled").     // DEFAULT value
    Immutable().             // cannot be updated after creation
    Label("Product Title").  // display label for forms
    Placeholder("Enter..."). // form placeholder
    Help("The public-facing product name.") // form help text
    Visibility(schema.Roles("admin", "editor")). // field-level read access
    Mutability(schema.Roles("admin")).            // field-level write access
```

### 5.4 Relationships

```go
schema.BelongsTo("Category", "categories").
    Optional().
    OnDelete(schema.SetNull).
    Eager()                                // always JOIN-loaded

schema.HasMany("Variants", "product_variants").
    OrderBy("position", schema.Asc)

schema.HasOne("SEO", "product_seo")

schema.ManyToMany("Tags", "product_tags", "tags")
```

### 5.5 What the Schema Generates

From a single `schema.Define(...)` call, `forge generate` produces two categories of output:

**Always-regenerated (in `gen/`, never edit):**

| Artifact | File | Description |
|----------|------|-------------|
| Atlas desired state | `gen/atlas/product.hcl` | Declarative schema for Atlas diffing |
| Go model types | `gen/models/product.go` | `Product`, `ProductCreate`, `ProductUpdate`, `ProductFilter`, `ProductSort` |
| Type-safe query builder | `gen/queries/product.go` | Bob-generated query mods for filtering, sorting, pagination |
| Validation | `gen/validation/product.go` | Validate functions returning typed field errors |
| Action interface | `gen/actions/product.go` | `ProductActions` interface + default implementation |
| Huma API structs | `gen/api/product_types.go` | `ListProductsInput`, `ProductResponse`, etc. for Huma |
| Huma route registration | `gen/api/product_routes.go` | Huma operation registration |
| Test factories | `gen/factories/product.go` | Factory functions for test data |

**Scaffolded once (in `resources/`, user owns and edits):**

| Artifact | File | Description |
|----------|------|-------------|
| Templ form component | `resources/product/form.templ` | Datastar-native form — generated once, then user-owned |
| Templ list component | `resources/product/list.templ` | Table with sort headers, filter controls, pagination |
| Templ detail component | `resources/product/detail.templ` | Read-only detail view |
| HTML handlers | `resources/product/handlers.go` | HTTP handlers that call the action layer |
| Custom hooks | `resources/product/hooks.go` | Validation hooks, lifecycle callbacks |

The distinction is critical: `gen/` is a build artifact. `resources/` is source code. Views and handlers are scaffolded once with `forge generate resource product`, then never overwritten unless the developer runs `forge generate resource product --force`.

If a developer modifies the schema and runs `forge generate`, only the `gen/` artifacts update. Their customized views and handlers are untouched.

### 5.6 Re-scaffolding Views After Schema Changes

When a schema changes (new field added, field removed), the developer's customized views won't automatically include the new field. Forge provides tooling for this:

```bash
forge generate resource product --diff
```

This outputs a diff showing what the *freshly scaffolded* views would look like vs. the current views, so the developer can manually merge the changes. This is analogous to `rails app:update` — explicit, not magical.

---

## 6. The Action Layer (Unified Business Logic)

### 6.1 The Problem

Huma API handlers use `func(ctx context.Context, input *T) (*Output, error)`. Hypermedia handlers use `func(http.ResponseWriter, *http.Request)`. Without a shared layer, business logic gets duplicated or one side becomes a second-class citizen.

### 6.2 The Solution: Actions

Every resource gets a generated `Actions` interface and default implementation. Both HTTP handlers (hypermedia) and Huma handlers (API) are thin adapters that call into the same action.

```go
// gen/actions/product.go (generated — always regenerated)

type ProductActions interface {
    List(ctx context.Context, params ProductListParams) (*ProductListResult, error)
    Get(ctx context.Context, id uuid.UUID) (*Product, error)
    Create(ctx context.Context, input ProductCreate) (*Product, *FieldErrors, error)
    Update(ctx context.Context, id uuid.UUID, input ProductUpdate) (*Product, *FieldErrors, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

type ProductListParams struct {
    Page     int
    PerPage  int
    Sort     string
    Filters  ProductFilter
    Search   string
    Cursor   *string  // for cursor-based pagination (API)
}

type ProductListResult struct {
    Products   []Product
    Pagination Pagination
}

// DefaultProductActions — generated default implementation
type DefaultProductActions struct {
    DB        *pgxpool.Pool
    River     *river.Client[pgx.Tx]
    Validator *ProductValidator
}

func (a *DefaultProductActions) Create(ctx context.Context, input ProductCreate) (*Product, *FieldErrors, error) {
    // 1. Run generated validation
    if errs := a.Validator.ValidateCreate(input); errs.HasErrors() {
        return nil, errs, nil
    }

    // 2. Insert via generated query builder (tenant-scoped automatically)
    product, err := a.insertProduct(ctx, input)
    if err != nil {
        return nil, nil, a.mapDBError(err) // unique violation → 409, etc.
    }

    // 3. Enqueue schema-defined AfterCreate jobs (transactional)
    // (generated from schema.Hooks)

    return product, nil, nil
}
```

### 6.3 Hypermedia Handler (thin adapter)

```go
// resources/product/handlers.go (scaffolded once, user-owned)

type ProductHTMLHandler struct {
    Actions gen.ProductActions
}

func (h *ProductHTMLHandler) Create(w http.ResponseWriter, r *http.Request) {
    var signals ProductFormSignals
    if err := datastar.ReadSignals(r, &signals); err != nil {
        forge.RenderError(w, r, err)
        return
    }

    input := models.ProductCreate{
        Title:  signals.Title,
        Price:  signals.Price,
        Status: signals.Status,
    }

    product, fieldErrs, err := h.Actions.Create(r.Context(), input)
    if err != nil {
        forge.RenderError(w, r, err)
        return
    }
    if fieldErrs != nil {
        sse := datastar.NewSSE(w, r)
        sse.MergeFragment(ProductForm(forms.WithErrors(input, fieldErrs)))
        return
    }

    sse := datastar.NewSSE(w, r)
    sse.Redirect("/products/" + product.ID.String())
}
```

### 6.4 Huma API Handler (thin adapter)

```go
// gen/api/product_routes.go (generated — always regenerated)

func RegisterProductAPI(api huma.API, actions gen.ProductActions) {
    huma.Register(api, huma.Operation{
        OperationID: "create-product",
        Method:      http.MethodPost,
        Path:        "/api/v1/products",
        Summary:     "Create a product",
        Tags:        []string{"Products"},
    }, func(ctx context.Context, input *CreateProductInput) (*CreateProductOutput, error) {
        product, fieldErrs, err := actions.Create(ctx, input.Body.ToProductCreate())
        if err != nil {
            return nil, err // Huma maps to proper HTTP status via forge.Error
        }
        if fieldErrs != nil {
            return nil, huma.Error422UnprocessableEntity("validation failed",
                fieldErrs.ToHumaErrors()...)
        }
        return &CreateProductOutput{Body: ProductResponseFrom(product)}, nil
    })
}
```

### 6.5 Overriding Actions

To add custom business logic, the developer replaces the default action implementation:

```go
// resources/product/actions.go (user-written)

type CustomProductActions struct {
    gen.DefaultProductActions // embed default for methods you don't override
    Mailer *mail.Client
}

func (a *CustomProductActions) Create(ctx context.Context, input models.ProductCreate) (*models.Product, *validation.FieldErrors, error) {
    // Custom pre-processing
    if input.Price.GreaterThan(decimal.NewFromFloat(10000)) {
        input.Status = "pending_review"
    }

    // Call default for DB insert + validation
    product, errs, err := a.DefaultProductActions.Create(ctx, input)
    if err != nil || errs != nil {
        return product, errs, err
    }

    // Custom post-processing
    if product.Status == "pending_review" {
        a.Mailer.Send(ctx, emails.HighValueProductReview(product))
    }

    return product, nil, nil
}
```

Registration happens explicitly at startup (no `init()` magic):

```go
// main.go or config/resources.go
func setupResources(app *forge.App) {
    app.Resource("products", &product.CustomProductActions{
        DefaultProductActions: gen.NewDefaultProductActions(app.DB, app.River),
        Mailer: app.Mailer,
    })

    // Resources without custom actions use the default
    app.Resource("categories", gen.NewDefaultCategoryActions(app.DB, app.River))
}
```

### 6.6 Custom Validation (No init() Hooks)

Custom validation is added by implementing an interface on the action, not via global `init()` registration:

```go
// resources/product/hooks.go (scaffolded once, user-owned)

func (a *CustomProductActions) ValidateCreate(input models.ProductCreate, errs *validation.FieldErrors) {
    // This is called by DefaultProductActions.Create AFTER generated validation
    if input.Price.GreaterThan(decimal.NewFromFloat(10000)) && input.Status == "active" {
        errs.Add("Price", "products over $10,000 require manual approval")
    }
}
```

The generated `DefaultProductActions.Create` checks if the concrete type implements `CreateValidator` and calls it:

```go
// gen/actions/product.go
type CreateValidator interface {
    ValidateCreate(input models.ProductCreate, errs *validation.FieldErrors)
}

func (a *DefaultProductActions) validate(input models.ProductCreate) *FieldErrors {
    errs := validation.ValidateProductCreate(input) // generated rules
    if v, ok := a.self.(CreateValidator); ok {
        v.ValidateCreate(input, &errs) // user rules
    }
    return errs
}
```

---

## 7. Huma Integration (OpenAPI from Code)

### 7.1 Why Huma (Not Custom OpenAPI Generation)

The goal is OpenAPI docs from code, the way .NET does it. In Go, Huma is the closest equivalent. It derives OpenAPI 3.1 specs from Go struct definitions, handles content negotiation, and produces RFC 7807 errors — all from annotated Go types.

We considered alternatives and rejected them:

| Alternative | Why not |
|---|---|
| Custom OpenAPI generation from schema | Reinventing Huma. We'd spend months on spec compliance, security schemes, response negotiation. |
| swaggo/swag (comment annotations) | Comments are stringly-typed, fragile, and detached from the actual handler logic. |
| ogen (spec-first codegen) | Spec-first is the opposite of our schema-first philosophy. We want code → spec, not spec → code. |
| kin-openapi (manual spec building) | Low-level library, not a framework. We'd still be writing all the routing/validation/negotiation glue. |

Huma is the right tool. The problem isn't Huma — it's the integration pattern.

### 7.2 Integration Pattern

Huma has a router-agnostic adapter model. We use `humachi` (or a stdlib adapter) to mount Huma on the same `http.ServeMux` as our hypermedia routes:

```go
// internal/app/app.go

func (app *App) setupRoutes() {
    mux := http.NewServeMux()

    // --- Hypermedia routes (Datastar + Templ) ---
    mux.Handle("GET /products", app.middleware(productHandler.List))
    mux.Handle("POST /products", app.middleware(productHandler.Create))
    mux.Handle("GET /products/{id}", app.middleware(productHandler.Show))
    // ...

    // --- API routes (Huma) ---
    humaConfig := huma.DefaultConfig("Forge App API", "1.0.0")
    humaConfig.Components.SecuritySchemes = app.securitySchemes()
    api := humastdlib.New(mux, humaConfig)  // mounts on same mux under /api/v1

    gen.RegisterProductAPI(api, app.productActions)
    gen.RegisterCategoryAPI(api, app.categoryActions)
    // ...

    app.server = &http.Server{Handler: mux}
}
```

Both the Huma API handlers and the HTML handlers call the same `ProductActions` interface. Huma owns validation for API transport (struct tags → OpenAPI), but the action layer also runs schema-derived validation — belt and suspenders, because the HTML side doesn't have Huma's transport validation.

### 7.3 Generated Huma Structs

The schema generates Huma-compatible input/output structs with proper struct tags:

```go
// gen/api/product_types.go (generated)

type ListProductsInput struct {
    TenantID  uuid.UUID `header:"X-Tenant-ID" required:"true" doc:"Tenant identifier"`
    Page      int       `query:"page" minimum:"1" default:"1" doc:"Page number (offset pagination)"`
    PerPage   int       `query:"per_page" minimum:"1" maximum:"100" default:"25"`
    Cursor    *string   `query:"cursor" doc:"Cursor for cursor-based pagination (mutually exclusive with page)"`
    Sort      string    `query:"sort" enum:"title,-title,price,-price,created_at,-created_at"`
    Status    *string   `query:"status" enum:"draft,active,archived"`
    Search    *string   `query:"search"`
    MinPrice  *float64  `query:"min_price" minimum:"0"`
    MaxPrice  *float64  `query:"max_price" minimum:"0"`
}

type ListProductsOutput struct {
    Body struct {
        Data       []ProductResponse `json:"data"`
        Pagination PaginationMeta    `json:"pagination"`
    }
    Link string `header:"Link"` // RFC 8288
}

type ProductResponse struct {
    ID        uuid.UUID     `json:"id"`
    Title     string        `json:"title"`
    Status    string        `json:"status" enum:"draft,active,archived"`
    Price     string        `json:"price" doc:"Decimal string"`
    SKU       string        `json:"sku"`
    Category  *CategoryRef  `json:"category,omitempty"`
    CreatedAt time.Time     `json:"created_at"`
    UpdatedAt time.Time     `json:"updated_at"`
}
```

### 7.4 OpenAPI Output

Served directly by Huma:

- `/api/openapi.json` — OpenAPI 3.1 JSON
- `/api/openapi.yaml` — OpenAPI 3.1 YAML
- `/api/docs` — Interactive docs (Scalar UI by default)

Exportable via CLI:

```bash
forge openapi export --format json > openapi.json
forge openapi export --format yaml > openapi.yaml
```

### 7.5 SDK Generation

Because we produce a Huma-powered OpenAPI 3.1 spec, standard tooling works:

```bash
# TypeScript
npx openapi-typescript-codegen --input http://localhost:8080/api/openapi.json --output ./sdk

# Go client
oapi-codegen -package client http://localhost:8080/api/openapi.json > client.go

# Any language
openapi-generator generate -i http://localhost:8080/api/openapi.json -g python -o ./python-sdk
```

### 7.6 Custom API Operations

Users extend the API by writing standard Huma operations that leverage generated types:

```go
// resources/product/api.go (user-written)

func RegisterCustomProductAPI(api huma.API, actions *CustomProductActions) {
    huma.Register(api, huma.Operation{
        OperationID: "bulk-update-prices",
        Method:      http.MethodPost,
        Path:        "/api/v1/products/bulk-price-update",
        Summary:     "Bulk update product prices",
        Tags:        []string{"Products"},
    }, func(ctx context.Context, input *BulkPriceInput) (*BulkPriceOutput, error) {
        // Uses actions layer
        return actions.BulkUpdatePrices(ctx, input.Body)
    })
}
```

---

## 8. Data Access Layer

### 8.1 Generated Query Builder (Bob)

Bob handles dynamic queries — filtering, sorting, pagination — that SQLC can't express without combinatorial SQL files.

```go
products, err := models.Products.Query(
    models.ProductWhere.Status.EQ("active"),
    models.ProductWhere.Price.GTE(decimal.NewFromFloat(10.00)),
    models.ProductWhere.Title.Contains("widget"),
    models.ProductOrderBy.Price.Desc(),
    models.ProductPaginate(page, perPage),
    models.ProductPreload.Category(),
).All(ctx, db)
```

### 8.2 SQLC Escape Hatch

For complex queries (CTEs, window functions, aggregates) that don't fit the generated builder:

```sql
-- queries/custom/product_reports.sql

-- name: GetTopSellingProducts :many
SELECT p.*, COUNT(oi.id) as order_count,
       SUM(oi.quantity * oi.unit_price) as revenue
FROM products p
JOIN order_items oi ON oi.product_id = p.id
WHERE p.tenant_id = @tenant_id
  AND p.deleted_at IS NULL
  AND oi.created_at >= @since
GROUP BY p.id
ORDER BY revenue DESC
LIMIT @limit;
```

### 8.3 Tenant Scoping

When `TenantScoped: true`, all generated queries include tenant filtering automatically. The tenant ID comes from context, set by middleware:

```go
// Automatic — user never writes this
func (q *ProductQuery) applyScopes(ctx context.Context) {
    tenantID := forge.TenantFromContext(ctx)
    q.Where(models.ProductWhere.TenantID.EQ(tenantID))

    // Soft delete scope (if SoftDelete: true)
    q.Where(models.ProductWhere.DeletedAt.IsNull())
}
```

Row-level security policies are generated in Atlas desired state:

```sql
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON products
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

### 8.4 Soft Delete (First-Class Treatment)

Soft delete is not a boolean flag — it's a pervasive concern that affects queries, uniqueness, cascading, and restoration.

**Implicit scope:** All generated queries exclude soft-deleted records by default. Explicit query mods override this:

```go
// Default: WHERE deleted_at IS NULL (implicit)
products, _ := models.Products.Query(...).All(ctx, db)

// Include soft-deleted
products, _ := models.Products.Query(models.WithTrashed()).All(ctx, db)

// Only soft-deleted
products, _ := models.Products.Query(models.OnlyTrashed()).All(ctx, db)
```

**Partial unique indexes:** When a field is `Unique()` on a `SoftDelete: true` resource, Atlas generates a partial index:

```sql
CREATE UNIQUE INDEX products_sku_unique ON products (sku, tenant_id)
    WHERE deleted_at IS NULL;
```

This allows re-creating a record with the same SKU after the original is soft-deleted.

**Cascading soft deletes:** Not automatic. If a developer wants deleting a Category to soft-delete its Products, they implement it in the action layer:

```go
func (a *CustomCategoryActions) Delete(ctx context.Context, id uuid.UUID) error {
    return forge.Transaction(ctx, a.DB, func(tx pgx.Tx) error {
        // Soft-delete all products in this category
        models.Products.Query(
            models.ProductWhere.CategoryID.EQ(id),
        ).SoftDeleteAll(ctx, tx)

        // Soft-delete the category
        return a.DefaultCategoryActions.Delete(ctx, id)
    })
}
```

This is intentionally explicit. Implicit cascading soft deletes are a source of data loss bugs.

**Restore operation:** Generated actions include `Restore`:

```go
type ProductActions interface {
    // ...
    Delete(ctx context.Context, id uuid.UUID) error
    Restore(ctx context.Context, id uuid.UUID) (*Product, error) // only if SoftDelete: true
}
```

**Eager-loaded relations:** `ProductPreload.Category()` automatically excludes soft-deleted categories. If a product's category was soft-deleted, the category field is nil. The generated detail view handles this gracefully (shows "Category removed" or similar).

### 8.5 Transaction Support

```go
err := forge.Transaction(ctx, db, func(tx pgx.Tx) error {
    product, err := models.Products.Insert(ctx, tx, models.ProductCreate{
        Title: "New Widget",
        Price: decimal.NewFromFloat(29.99),
    })
    if err != nil {
        return err
    }

    // River job enqueued in same transaction — never lost
    _, err = riverClient.InsertTx(ctx, tx, NotifyNewProduct{ProductID: product.ID}, nil)
    return err
})
```

---

## 9. Migration System (Atlas)

### 9.1 Why Atlas, Not Custom

Building a schema-to-SQL diffing engine from scratch is a multi-year effort. Edge cases include: column rename vs. drop-and-add ambiguity, enum value removal (Postgres forbids it), partial unique indexes, concurrent index creation, foreign key ordering, and data-preserving type changes. Atlas handles all of this. We don't.

### 9.2 How It Works

1. `forge generate` produces Atlas-compatible desired state files (`gen/atlas/*.hcl`) from the schema definitions.
2. `forge migrate diff` runs `atlas schema diff` against the live database and produces a migration file.
3. `forge migrate up` applies pending migrations via `atlas migrate apply`.

```bash
# Developer workflow
vim resources/product/schema.go   # add a new field
forge generate                     # regenerates gen/atlas/product.hcl
forge migrate diff                 # Atlas diffs against live DB → migrations/YYYYMMDD_add_foo.sql
forge migrate up                   # applies
```

### 9.3 Migration Review

Generated migrations are always written to `migrations/` as SQL files that the developer reviews and commits. `forge migrate diff` never applies changes automatically. The developer reads the SQL, verifies it's correct, and then runs `forge migrate up`.

For destructive changes (column drops, type changes that lose data), Atlas adds a warning comment in the generated SQL and `forge migrate diff` prints a prominent warning:

```
⚠ DESTRUCTIVE: migrations/20260215_003_drop_legacy_field.sql
   - DROP COLUMN "legacy_notes" on table "products" (data will be lost)
   Review carefully before applying.
```

### 9.4 Manual Migrations

Developers can add hand-written migration files to `migrations/`. Atlas's migration directory model supports this — it tracks applied migrations by filename, so hand-written and generated files coexist.

---

## 10. Error Handling

### 10.1 Error Philosophy

Forge defines a consistent error type that carries HTTP semantics, user-facing messages, and machine-readable codes. Both the HTML and API layers use the same error type; they differ only in rendering.

### 10.2 forge.Error

```go
package forge

type Error struct {
    Status  int    // HTTP status code
    Code    string // machine-readable code: "product.duplicate_sku"
    Message string // user-facing message
    Detail  string // developer-facing detail (omitted in production)
    Err     error  // wrapped underlying error (never exposed to user)
}

func (e *Error) Error() string { return e.Message }

// Constructors
func NotFound(resource, id string) *Error
func Conflict(message string) *Error
func Forbidden(message string) *Error
func BadRequest(message string) *Error
func InternalError(err error) *Error
```

### 10.3 Database Error Mapping

The action layer automatically maps database errors to `forge.Error`:

```go
// gen/actions/errors.go (generated)

func mapDBError(err error) *Error {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // unique_violation
            field := extractFieldFromConstraint(pgErr.ConstraintName)
            return Conflict(fmt.Sprintf("%s already exists", field))
        case "23503": // foreign_key_violation
            return BadRequest("referenced record does not exist")
        case "23514": // check_violation
            return BadRequest(pgErr.Message)
        }
    }
    if errors.Is(err, pgx.ErrNoRows) {
        return NotFound("record", "")
    }
    return InternalError(err) // logs full error; returns generic message to user
}
```

### 10.4 HTML Error Rendering

```go
// forge/errors.go

func RenderError(w http.ResponseWriter, r *http.Request, err error) {
    var fe *Error
    if !errors.As(err, &fe) {
        fe = InternalError(err)
    }

    // Log with full context
    slog.ErrorContext(r.Context(), "request error",
        "status", fe.Status, "code", fe.Code, "err", fe.Err)

    if datastar.IsSSE(r) {
        sse := datastar.NewSSE(w, r)
        sse.MergeFragment(ErrorToast(fe.Message))
    } else {
        w.WriteHeader(fe.Status)
        ErrorPage(fe).Render(r.Context(), w)
    }
}
```

### 10.5 API Error Rendering

Huma automatically handles `forge.Error` because it implements the `error` interface and Huma maps errors to RFC 7807 responses. We register a custom error transformer:

```go
humaConfig.Errors = func(ctx huma.Context, err error) {
    var fe *forge.Error
    if errors.As(err, &fe) {
        huma.WriteErr(ctx, fe.Status, fe.Message, &huma.ErrorDetail{
            Location: fe.Code,
            Message:  fe.Detail,
        })
        return
    }
    huma.DefaultErrorHandler(ctx, err)
}
```

### 10.6 Panic Recovery

Global middleware catches panics and converts them to 500 errors with stack traces logged (never exposed to users):

```go
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                slog.ErrorContext(r.Context(), "panic recovered",
                    "panic", rec, "stack", string(debug.Stack()))
                RenderError(w, r, InternalError(fmt.Errorf("panic: %v", rec)))
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

---

## 11. Hypermedia Layer (Datastar + Templ)

### 11.1 Scaffolded Views

Views are scaffolded once into `resources/<name>/` and owned by the developer:

```go
// resources/product/form.templ (scaffolded, user-owned)

templ ProductForm(form forms.Form[models.ProductCreate]) {
    <form id="product-form" data-on:submit="@post('/products')">
        @FormField(form, "Title", forms.TextInput{
            Label:       "Product Title",
            Placeholder: "Enter product title",
            MaxLen:      200,
            Required:    true,
        })
        @FormField(form, "Price", forms.DecimalInput{
            Label:    "Price",
            Min:      0,
            Step:     "0.01",
            Required: true,
        })
        @FormField(form, "Status", forms.SelectInput{
            Label:   "Status",
            Options: []string{"draft", "active", "archived"},
        })
        @FormField(form, "CategoryID", forms.RelationSelect{
            Label:    "Category",
            FetchURL: "/api/v1/categories?fields=id,name",
            Optional: true,
        })
        <button type="submit">Save Product</button>
    </form>
}
```

Because this is user-owned, the developer can freely:
- Add conditional field visibility based on Datastar signals
- Combine multiple resources into one form
- Build multi-step wizard flows
- Add inline editing to list views
- Replace the entire form with a custom component

### 11.2 Shared Form Primitives

Forge provides form field components as a library (not generated per-resource):

```go
// forge/forms — library, not generated

templ FormField[T any](form forms.Form[T], field string, input InputComponent) {
    <div class={ "form-group", templ.KV("has-error", form.HasError(field)) }>
        <label for={ field }>{ input.Label() }</label>
        @input.Render(form.Value(field))
        if form.HasError(field) {
            for _, msg := range form.Errors(field) {
                <span class="error-message">{ msg }</span>
            }
        }
        if input.Help() != "" {
            <span class="help-text">{ input.Help() }</span>
        }
    </div>
}
```

### 11.3 SSE Resource Management

SSE connections (for Datastar real-time updates and general handler responses) require careful resource management:

**Connection limits:** A global SSE connection limiter caps total concurrent SSE connections per process (default: 5000, configurable). New connections beyond the limit receive a 503 with `Retry-After` header.

**Per-user limits:** Each authenticated user is limited to N concurrent SSE connections (default: 10). This prevents a single bad actor from exhausting the pool.

**Graceful shutdown:** On server shutdown, all SSE connections receive a close event and are drained with a configurable timeout.

```go
// forge/sse/limiter.go

type SSELimiter struct {
    maxTotal   int
    maxPerUser int
    active     atomic.Int64
    perUser    sync.Map // user_id → *atomic.Int64
}

func (l *SSELimiter) Acquire(userID string) (release func(), err error) {
    if l.active.Load() >= int64(l.maxTotal) {
        return nil, ErrTooManyConnections
    }
    // ... per-user check, increment counters, return release func
}
```

### 11.4 Real-Time Updates (LISTEN/NOTIFY)

**Single shared listener:** The application maintains ONE persistent PostgreSQL connection for LISTEN, not one per SSE client. Events fan out to subscribers via Go channels:

```go
// forge/notify/hub.go

type Hub struct {
    conn     *pgx.Conn          // single dedicated LISTEN connection
    subs     map[string][]chan Event // channel -> "products:tenant_123"
    mu       sync.RWMutex
}

// One connection, many subscribers
func (h *Hub) Subscribe(channel string, tenantID uuid.UUID) *Subscription {
    key := channel + ":" + tenantID.String()
    ch := make(chan Event, 32) // buffered — if full, events are dropped (backpressure)
    h.mu.Lock()
    h.subs[key] = append(h.subs[key], ch)
    h.mu.Unlock()
    return &Subscription{ch: ch, key: key, hub: h}
}
```

**Backpressure:** If a subscriber's channel buffer fills up (client is slow), events are dropped rather than blocking the fan-out goroutine. The client receives a "refresh" signal to re-fetch current state.

**Multi-instance:** LISTEN/NOTIFY works within a single PostgreSQL instance, so multiple app replicas sharing the same database all receive notifications. This covers the common case. For external pub/sub (Redis, NATS), Forge provides a `NotifyHub` interface that can be swapped:

```go
type NotifyHub interface {
    Subscribe(channel string, tenantID uuid.UUID) *Subscription
    Publish(channel string, tenantID uuid.UUID, payload []byte) error
}

// Default: PostgreSQL LISTEN/NOTIFY (works for most deployments)
// Optional: Redis, NATS (for deployments with connection poolers like PgBouncer
// that don't support LISTEN)
```

---

## 12. Pagination

### 12.1 Dual Strategy

Pagination is a per-route concern, not a global setting:

**Offset pagination** (default for hypermedia): Simple, works with "page 3 of 12" UI. Good for browsing. Breaks when records are inserted/deleted between pages (acceptable for HTML lists where the user sees it live).

**Cursor pagination** (default for API): Stable across insertions/deletions. Required for infinite scroll, webhook deliveries, and any consumer that iterates through all records. The cursor is an opaque base64-encoded string containing the sort column value + ID.

```go
// gen/models/pagination.go

type Pagination struct {
    // Offset-based (HTML)
    Page       int  `json:"page,omitempty"`
    PerPage    int  `json:"per_page,omitempty"`
    TotalPages int  `json:"total_pages,omitempty"`
    Total      int  `json:"total,omitempty"` // excludes soft-deleted

    // Cursor-based (API)
    NextCursor *string `json:"next_cursor,omitempty"`
    PrevCursor *string `json:"prev_cursor,omitempty"`
    HasMore    bool    `json:"has_more"`
}
```

The generated query builder supports both:

```go
// Offset
models.ProductPaginate(page, perPage)

// Cursor
models.ProductCursorPaginate(cursor, perPage, models.ProductOrderBy.CreatedAt.Desc())
```

---

## 13. Authentication and Authorization

### 13.1 Authentication

Pluggable, not monolithic:

```go
func setupAuth(app *forge.App) {
    app.UseAuth(forge.AuthConfig{
        SessionStore:    forge.PostgresSessionStore(app.DB),
        SessionLifetime: 24 * time.Hour,

        Strategies: []forge.AuthStrategy{
            forge.EmailPassword(forge.EmailPasswordConfig{
                MinPasswordLength: 12,
            }),
            forge.OAuth2(forge.OAuth2Config{
                Providers: []forge.OAuth2Provider{
                    forge.Google(clientID, clientSecret),
                    forge.GitHub(clientID, clientSecret),
                },
            }),
        },

        APIAuth: []forge.APIAuthStrategy{
            forge.BearerToken(),
            forge.APIKey("X-API-Key"),
        },
    })
}
```

### 13.2 Authorization: CRUD-Level + Field-Level

CRUD-level permissions (who can do what):

```go
var Resource = schema.Define("Product", schema.Options{
    Permissions: schema.Permissions{
        List:   schema.Allow("admin", "editor", "viewer"),
        Read:   schema.Allow("admin", "editor", "viewer"),
        Create: schema.Allow("admin", "editor"),
        Update: schema.Allow("admin", "editor"),
        Delete: schema.Allow("admin"),
    },
})
```

Field-level permissions (who can see/edit what):

```go
schema.Decimal("CostPrice").Required().
    Visibility(schema.Roles("admin", "finance")).  // viewers/editors can't see this
    Mutability(schema.Roles("admin"))              // only admin can change it

schema.Decimal("Price").Required().
    Visibility(schema.AllAuthenticated()).  // everyone can see
    Mutability(schema.Roles("admin", "editor")) // viewer can't edit
```

The generated actions check field-level permissions before returning data:

```go
func (a *DefaultProductActions) Get(ctx context.Context, id uuid.UUID) (*Product, error) {
    product, err := models.Products.Find(ctx, a.DB, id)
    if err != nil {
        return nil, a.mapDBError(err)
    }
    return a.filterFields(ctx, product), nil // strips invisible fields based on role
}
```

The generated API response structs use `omitempty` and the action layer nils out fields the current user can't see. The generated form components conditionally render fields based on the user's role.

### 13.3 Huma Security Schemes

Generated into the OpenAPI spec:

```go
config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
    "bearerAuth": {
        Type:   "http",
        Scheme: "bearer",
    },
    "apiKey": {
        Type: "apiKey",
        In:   "header",
        Name: "X-API-Key",
    },
}
```

---

## 14. Background Jobs (River)

### 14.1 Integration

River is embedded as a first-class citizen with transactional enqueueing:

```go
// resources/product/jobs.go (user-written)

type NotifyNewProduct struct {
    ProductID uuid.UUID `json:"product_id"`
    TenantID  uuid.UUID `json:"tenant_id"` // always include for scoping
}

func (NotifyNewProduct) Kind() string { return "notify_new_product" }

type NotifyNewProductWorker struct {
    river.WorkerDefaults[NotifyNewProduct]
    db     *pgxpool.Pool
    mailer *mail.Client
}

func (w *NotifyNewProductWorker) Work(ctx context.Context, job *river.Job[NotifyNewProduct]) error {
    // Restore tenant context for scoped queries
    ctx = forge.WithTenant(ctx, job.Args.TenantID)

    product, err := models.Products.Find(ctx, w.db, job.Args.ProductID)
    if err != nil {
        return err
    }
    return w.mailer.Send(ctx, emails.NewProductNotification(product))
}
```

### 14.2 Schema-Triggered Jobs

```go
var Resource = schema.Define("Product", schema.Options{
    Hooks: schema.Hooks{
        AfterCreate: []schema.JobRef{
            {Kind: "notify_new_product", Queue: "notifications"},
        },
        AfterUpdate: []schema.JobRef{
            {Kind: "reindex_product_search", Queue: "indexing"},
        },
    },
})
```

The generated action layer enqueues these jobs transactionally — they're inserted in the same database transaction as the record mutation, so they're never lost and never fire for a rolled-back operation.

---

## 15. Audit Logging

### 15.1 Beyond created_by/updated_by

`Auditable: true` on a resource does two things:

**Tracking columns:** Adds `created_by` and `updated_by` columns to the resource table, automatically populated from the authenticated user in context.

**Audit log table:** Generates entries in a shared `audit_log` table that records the full change:

```sql
CREATE TABLE audit_log (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL,
    resource    text NOT NULL,        -- "products"
    resource_id uuid NOT NULL,
    action      text NOT NULL,        -- "create", "update", "delete", "restore"
    actor_id    uuid NOT NULL,
    changes     jsonb,                -- {"price": {"old": "29.99", "new": "39.99"}}
    metadata    jsonb,                -- request ID, IP, user agent
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_resource ON audit_log (resource, resource_id);
CREATE INDEX idx_audit_log_tenant ON audit_log (tenant_id, created_at);
```

The `changes` JSONB contains a diff of old and new values, computed in the action layer before the update:

```go
// gen/actions/audit.go (generated)

func (a *DefaultProductActions) auditUpdate(ctx context.Context, old, new *Product) {
    changes := diffProducts(old, new) // generated field-by-field diff
    if len(changes) == 0 {
        return // no-op update, don't log
    }
    a.auditLog.Record(ctx, AuditEntry{
        Resource:   "products",
        ResourceID: old.ID,
        Action:     "update",
        ActorID:    forge.UserFromContext(ctx).ID,
        Changes:    changes,
    })
}
```

---

## 16. Multi-Tenancy

### 16.1 Tenant Resolution

```go
func setupTenancy(app *forge.App) {
    app.UseTenancy(forge.TenancyConfig{
        Strategy:  forge.HeaderTenant("X-Tenant-ID"),  // or SubdomainTenant, PathTenant
        Fallback:  forge.TenantFromSession(),
        Required:  true,
        OnMissing: forge.Redirect("/select-tenant"),
    })
}
```

### 16.2 Automatic Scoping

When `TenantScoped: true`: all generated queries include `WHERE tenant_id = $tenant`, RLS policies are generated in Atlas desired state, API requests require tenant identification, and test factories scope data to a test tenant.

### 16.3 Tenant-Aware Jobs

Jobs always carry `TenantID` explicitly (not relying on context that won't exist in the worker):

```go
_, err = riverClient.InsertTx(ctx, tx, NotifyNewProduct{
    ProductID: product.ID,
    TenantID:  forge.TenantFromContext(ctx),
}, nil)
```

---

## 17. Testing

### 17.1 Generated Factories

```go
func TestProductCreation(t *testing.T) {
    db := forgetest.NewTestDB(t) // isolated test schema, auto-cleaned

    category := factories.Category.Create(t, db)

    product := factories.Product.Create(t, db,
        factories.ProductWith.Title("Test Widget"),
        factories.ProductWith.Category(category),
        factories.ProductWith.Price(decimal.NewFromFloat(29.99)),
    )

    assert.Equal(t, "Test Widget", product.Title)
    assert.Equal(t, category.ID, *product.CategoryID)
}
```

### 17.2 Action Testing (Unit)

Because business logic lives in the action layer, it's testable without HTTP:

```go
func TestCreateProduct_DuplicateSKU(t *testing.T) {
    db := forgetest.NewTestDB(t)
    actions := gen.NewDefaultProductActions(db, nil)

    factories.Product.Create(t, db, factories.ProductWith.SKU("WIDGET-1"))

    _, _, err := actions.Create(forgetest.TenantContext(t), models.ProductCreate{
        Title: "Another Widget",
        SKU:   "WIDGET-1",
        Price: decimal.NewFromFloat(10),
    })

    var fe *forge.Error
    require.ErrorAs(t, err, &fe)
    assert.Equal(t, 409, fe.Status)
}
```

### 17.3 HTTP Testing

```go
func TestListProductsAPI(t *testing.T) {
    app := forgetest.NewApp(t)

    factories.Product.CreateN(t, app.DB, 5,
        factories.ProductWith.Status("active"),
    )

    resp := app.Get("/api/v1/products?status=active")
    assert.Equal(t, 200, resp.StatusCode)

    var body ListProductsOutput
    json.Decode(resp.Body, &body)
    assert.Equal(t, 5, len(body.Data))
}
```

### 17.4 Datastar/Hypermedia Testing

```go
func TestProductFormValidation(t *testing.T) {
    app := forgetest.NewApp(t)

    resp := app.PostDatastar("/products", map[string]any{
        "title": "", // required
        "price": "29.99",
    })

    fragment := forgetest.ParseSSEFragment(resp)
    assert.Contains(t, fragment.HTML, "Title is required")
}
```

---

## 18. CLI

### 18.1 Commands

```bash
forge init <project-name>           # Scaffold new project
forge dev                            # Start dev server with hot reload
forge build                          # Compile production binary

forge generate                       # Regenerate all gen/ artifacts
forge generate resource <name>       # Scaffold new resource (views, handlers — once only)
forge generate resource <name> --diff # Show diff between current and fresh scaffold
forge generate resource <name> --force # Re-scaffold (overwrites user files!)
forge generate api                   # Regenerate Huma API registrations only
forge generate views                 # Regenerate view primitives library only

forge migrate diff                   # Atlas diff → new migration file (review before applying!)
forge migrate up                     # Apply pending migrations
forge migrate down                   # Roll back last migration
forge migrate status                 # Show migration state

forge db create                      # Create database
forge db drop                        # Drop database
forge db seed                        # Run seed data
forge db console                     # Open psql shell
forge db reset                       # Drop + create + migrate + seed

forge query compile                  # Run SQLC for custom queries
forge view compile                   # Compile Templ files

forge tool sync                      # Download/update tool binaries (templ, sqlc, tailwind, atlas)
forge routes                         # List all registered routes (HTML + API)
forge openapi export                 # Export OpenAPI spec to file

forge deploy                         # Build + package for deployment
```

### 18.2 Development Server

`forge dev` orchestrates:

1. File watcher for `.go`, `.templ`, `.sql`, `.css` files (via `fsnotify`)
2. Templ compilation on `.templ` change
3. SQLC compilation on `.sql` change
4. Go rebuild and restart
5. Tailwind CSS rebuild
6. Browser hot reload via Datastar SSE

---

## 19. Project Structure

```
myapp/
├── forge.toml                    # Project config
├── main.go                       # Entry point (wires forge.App, registers resources)
│
├── resources/                    # Resource definitions (YOUR CODE)
│   ├── product/
│   │   ├── schema.go             # Schema definition (source of truth)
│   │   ├── actions.go            # Custom action overrides (optional)
│   │   ├── handlers.go           # HTML handler overrides (scaffolded once)
│   │   ├── hooks.go              # Validation hooks, lifecycle (scaffolded once)
│   │   ├── form.templ            # Form view (scaffolded once, you own this)
│   │   ├── list.templ            # List view (scaffolded once, you own this)
│   │   ├── detail.templ          # Detail view (scaffolded once, you own this)
│   │   ├── api.go                # Custom Huma API operations (optional)
│   │   └── jobs.go               # Resource-specific jobs (optional)
│   ├── category/
│   │   └── schema.go
│   └── user/
│       └── schema.go
│
├── gen/                          # GENERATED CODE (never edit — always regenerated)
│   ├── atlas/                    # Atlas desired state HCL
│   ├── models/                   # Go types for all resources
│   ├── queries/                  # Bob query builders
│   ├── validation/               # Validation functions
│   ├── actions/                  # Action interfaces + default implementations
│   ├── api/                      # Huma types + route registration
│   └── factories/                # Test factories
│
├── queries/                      # Custom SQLC queries (escape hatch)
│   └── custom/
│       └── reports.sql
│
├── views/                        # Shared view components (your code)
│   ├── layouts/
│   │   ├── app.templ
│   │   └── auth.templ
│   └── components/
│       ├── nav.templ
│       └── footer.templ
│
├── middleware/                   # Custom middleware
├── config/                       # App configuration
│
├── migrations/                   # SQL migrations (generated by Atlas + manual)
│   ├── 20260215_001_initial.sql
│   └── atlas.sum                 # Atlas integrity file
│
├── css/
│   └── app.css                   # Tailwind source
│
├── static/                       # Static assets (embedded in binary)
│
└── tests/
```

Key design decisions:

- **`resources/`** is your code — schema, views, handlers, hooks, jobs all live together per resource. Scaffolded once, then you own it.
- **`gen/`** is a build artifact — regenerated on every `forge generate`. Never edit.
- **Clear boundary:** models, queries, validation, action interfaces, API types → `gen/`. Views, handlers, custom actions, hooks → `resources/`.

---

## 20. Configuration

```toml
# forge.toml

[app]
name = "myapp"
module = "github.com/myorg/myapp"

[server]
port = 8080
read_timeout = "30s"
write_timeout = "30s"

[database]
url = "${DATABASE_URL}"
max_connections = 25

[auth]
session_lifetime = "24h"

[tenancy]
enabled = true
strategy = "header"
header = "X-Tenant-ID"

[api]
enabled = true
prefix = "/api/v1"
docs_path = "/api/docs"

[jobs]
enabled = true
queues.default = 100
queues.notifications = 25
queues.indexing = 10

[sse]
max_total_connections = 5000
max_per_user = 10
buffer_size = 32

[assets]
tailwind = true
tailwind_version = "4.0"
embed_static = true

[telemetry]
tracing = "otlp"
metrics = "prometheus"
log_level = "info"
log_format = "json"

[dev]
hot_reload = true
port = 3000
```

All config overridable via environment variables: `FORGE_SERVER_PORT=9090`, `DATABASE_URL=...`, etc. 12-factor compliance.

---

## 21. Deployment

### 21.1 Single Binary

```bash
forge build
# Output: ./bin/myapp (single binary, ~20-30MB)
# Includes: compiled templates, static assets, embedded migrations
```

### 21.2 Docker

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN forge build

FROM alpine:3.19
COPY --from=builder /app/bin/myapp /usr/local/bin/
EXPOSE 8080
CMD ["myapp", "serve"]
```

Image: ~30MB. Startup: <100ms. Memory: ~20-40MB baseline.

---

## 22. Development Roadmap

### Phase 1: Schema Pipeline (6 weeks)

**Goal:** Prove the core thesis — schema → generated code that compiles and runs.

Deliverables:
- Schema DSL: field types, modifiers, relationships (no Hooks, no Permissions, no Auditable yet)
- `go/ast` parser that extracts schema definitions
- Code generator producing: Go model types, Bob query builder mods, validation functions, Atlas desired state HCL
- Atlas integration: `forge migrate diff` and `forge migrate up`
- CLI: `forge init`, `forge generate`, `forge migrate diff/up/status`
- A single example app (product + category) that compiles and runs queries

**Not included:** No views, no handlers, no API, no jobs, no auth, no tenancy. Just schema → types + queries + migrations.

**Exit criteria:** Change a field in schema.go, run `forge generate && forge migrate diff`, and get a working migration + updated Go types + updated query builder. Zero manual sync.

### Phase 2: Action Layer + CRUD Handlers (6 weeks)

**Goal:** Full CRUD loop — browser to database and back.

Deliverables:
- Action layer: generated interfaces + default implementations
- Error handling: `forge.Error`, DB error mapping, panic recovery
- Scaffolded Templ views: form, list, detail (generated once into resources/)
- Scaffolded HTML handlers (calling action layer)
- Form primitives library (FormField, inputs, error display)
- Datastar SSE helpers (MergeFragment, Redirect)
- `forge dev` file watcher with hot reload
- Test factories
- Soft delete: query scoping, partial unique indexes, WithTrashed/OnlyTrashed, Restore

**Exit criteria:** `forge init && forge generate resource product && forge dev` → working list/create/edit/delete with validation in browser. < 5 minutes.

### Phase 3: Huma API Layer (4 weeks)

**Goal:** JSON API with full OpenAPI documentation from the same schema.

Deliverables:
- Generated Huma input/output structs from schema
- Generated Huma route registration calling action layer
- OpenAPI 3.1 served at `/api/docs`
- Cursor-based pagination for API, offset for HTML
- API auth: bearer tokens, API keys
- CORS configuration
- Rate limiting middleware
- CLI: `forge openapi export`

**Exit criteria:** Same resource has both a working Datastar HTML interface AND a documented JSON API. Both call the same action layer. OpenAPI spec passes spectral linting and produces working TypeScript SDK.

### Phase 4: Multi-Tenancy + Auth (4 weeks)

**Goal:** Production-ready auth and tenant isolation.

Deliverables:
- Tenant resolution: header, subdomain, path strategies
- Automatic query scoping + RLS policy generation
- Session-based email/password auth
- OAuth2 providers (Google, GitHub)
- CRUD-level permissions from schema
- Field-level visibility/mutability
- Tenant-aware background jobs

**Exit criteria:** Two tenants on the same database, each seeing only their data, with role-based access controlling which fields are visible.

### Phase 5: Production Hardening (4 weeks)

**Goal:** Ready for real deployments.

Deliverables:
- Audit logging (change tracking with JSONB diffs)
- SSE connection management (limits, backpressure, graceful shutdown)
- LISTEN/NOTIFY hub with single-connection fan-out
- River integration: schema-triggered jobs, transactional enqueueing
- OpenTelemetry: traces, metrics, structured logging
- File uploads with S3/local storage abstraction
- Email sending with Templ templates
- CI/CD templates
- `forge deploy`

### Phase 6: Ecosystem (ongoing)

- Component library (DatastarUI-style, shadcn/ui for Templ)
- Admin panel generator
- Webhook system (outbound)
- Full-text search (PostgreSQL tsvector, generated from Searchable fields)
- Real-time collaboration primitives

---

## 23. Success Criteria

1. **Time to first CRUD resource:** < 5 minutes from `forge init` to working list/create/edit/delete with validation in browser AND documented JSON API.

2. **Schema change propagation:** Change a field in `schema.go`, run `forge generate`, and models + queries + validation + Atlas state + API types all update. Migration generated by `forge migrate diff`. Zero manual sync.

3. **Override without abandoning:** Developer can customize any view, handler, or action by editing files in `resources/` without touching `gen/`. Adding a custom validation rule takes < 5 lines of code with no `init()` functions.

4. **Production binary:** Single file < 30MB, starts in < 100ms, handles 10k req/s.

5. **Developer experience:** No JavaScript toolchain. No npm. `go install` + `forge init` = ready.

6. **API quality:** Generated OpenAPI spec passes spectral linting, produces working SDKs in TypeScript/Python/Go.

7. **SSE stability:** 1000 concurrent SSE connections with bounded memory growth, graceful degradation under load, no goroutine leaks.

---

## Appendix A: Why Huma (and Not Just Writing Our Own OpenAPI)

The goal is .NET-style "OpenAPI from code." In Go, this means deriving an OpenAPI 3.1 spec from Go struct definitions with proper types, validation rules, enum constraints, and security schemes — then serving interactive docs and enabling SDK generation.

Huma does exactly this. It's router-agnostic (works with stdlib), handles content negotiation, produces RFC 7807 errors, and has an active maintainer (Daniel Taylor). Building this ourselves would take months and produce an inferior result.

The integration challenge (two HTTP stacks) is solved by the action layer (Section 6). Huma owns the API transport. Stdlib handlers own the HTML transport. Both call the same actions. The developer never writes business logic twice.

## Appendix B: Why Atlas (and Not Custom Migration Diffing)

Building a schema-to-SQL diffing engine requires handling: column rename ambiguity, enum value lifecycle, partial unique indexes, concurrent index creation, foreign key ordering, data-preserving type changes, and dozens of PostgreSQL-specific edge cases. Atlas (by the Ent team at Ariga) has a full-time team working on this and it still has edge cases.

Forge generates Atlas-compatible desired state from the schema DSL. Atlas diffs it against the live database. We get declarative migrations without building a migration engine.

Trade-off: Atlas is a dependency (~15MB binary). We accept this because the alternative (building our own) is a multi-year detour from the actual product.

## Appendix C: Why Bob Over SQLC-Only

SQLC generates perfect type-safe code from static SQL. But a product list with filters (status, price range, category), search, sorting (title asc/desc, price asc/desc), and pagination requires either:

- One SQLC query per combination (combinatorial explosion)
- One query with COALESCE/CASE logic (query planner can't optimize)

Bob generates composable query mods from the schema. Filters, sorts, and pagination are Go functions that produce optimized SQL. SQLC remains available for complex reporting queries, CTEs, and window functions.

## Appendix D: Why Datastar Over HTMX

- **SSE-native:** Datastar uses Server-Sent Events, mapping naturally to Go's goroutine-per-connection model. HTMX uses Ajax, requiring polling for real-time.
- **Signals:** Built-in reactive signal system (like Alpine.js) — no second library needed.
- **Size:** ~15KB total vs. HTMX + Alpine.js at ~45KB combined.
- **DOM morphing:** idiomorph built in for smooth updates.
- **Go SDK:** Official, well-maintained, integrates cleanly with Templ.