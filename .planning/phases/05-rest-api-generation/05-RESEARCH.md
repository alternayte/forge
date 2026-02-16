# Phase 5: REST API Generation - Research

**Researched:** 2026-02-17
**Domain:** REST API generation with Huma framework, OpenAPI 3.1, authentication, and API documentation
**Confidence:** HIGH

## Summary

Phase 5 generates production-ready REST APIs from resource schemas using the Huma v2 framework. Huma provides automatic OpenAPI 3.1 spec generation, built-in validation via struct tags, and a "bring your own router" philosophy that integrates cleanly with existing middleware chains. The research confirms Huma v2 is the appropriate choice for code generation (generates Input/Output structs with validation tags, route registration functions call action layer from Phase 4).

Key architectural decisions are locked in CONTEXT.md: flat URL design with path versioning, opaque bearer tokens and API keys in separate database tables, response envelopes with pagination metadata, RFC 9457 error format, and Scalar UI embedded in the binary. Research validates these decisions align with current best practices and have proven Go implementations.

**Primary recommendation:** Use Huma v2.35+ with Chi router adapter, generate Input/Output structs with comprehensive validation tags from schema definitions, wire middleware in order (logging → panic recovery → CORS → rate limiting → authentication), and embed Scalar UI using nyxstack/scalarui package for CDN-free documentation.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**API URL design:**
- Path-based versioning: `/api/v1/<resource>`
- Flat URLs only — no nested resource routes (use query params like `?post_id=123` for filtering)
- Plural kebab-case resource names in URLs (`/api/v1/blog-posts`)
- Standard 5 CRUD endpoints per resource (List, Get, Create, Update, Delete) — no bulk operations

**Authentication & API keys:**
- Opaque bearer tokens stored in a database table (not JWTs) — revocable, queryable
- API keys are a separate table from bearer tokens — different use cases, own name/prefix/scopes/expiry
- Prefixed key format: `forg_live_abc123` / `forg_test_abc123` (Stripe/GitHub style)
- All API endpoints require authentication by default — public access is opt-in via schema annotation

**Response shape & errors:**
- All responses wrapped in envelope: `{"data": ...}` for single resources, `{"data": [...], "pagination": {...}}` for lists
- Pagination metadata inside the envelope: `{"next_cursor": "...", "has_more": true}`
- Error responses follow RFC 9457 (Problem Details for HTTP APIs)
- Validation errors include per-field detail: `"errors": [{"field": "email", "message": "invalid format"}, ...]`

**Docs & developer UX:**
- Scalar UI embedded directly in the app binary at `/api/docs` — no CDN dependency, matches single-binary philosophy
- OpenAPI spec designed for SDK generation: strict operationIds (listPosts, getPost), consistent naming, passes spectral linting
- `forge routes` output grouped by resource (not flat table)
- `forge openapi export` supports both JSON and YAML formats via `--format` flag

### Claude's Discretion

Research and recommend:
- Rate limiting strategy and configuration shape in forge.toml
- CORS default configuration
- Exact Huma middleware wiring order
- OpenAPI spec metadata (title, description, contact, license)
- How Scalar UI assets are embedded (go:embed vs inline)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/danielgtaylor/huma/v2 | v2.35.0+ | REST API framework with OpenAPI 3.1 | De facto Go framework for OpenAPI-first APIs; automatic validation, type-safe handlers, router-agnostic design |
| github.com/go-chi/chi/v5 | v5.x | HTTP router | Lightweight, stdlib-compatible, excellent middleware ecosystem; recommended Huma adapter |
| github.com/nyxstack/scalarui | latest | Embedded API documentation UI | Self-contained Scalar UI (no CDN), fluent config API, supports embedded specs |
| github.com/rs/cors | v1.11+ | CORS middleware | Industry standard, 9,500+ dependents, comprehensive options, well-tested |
| github.com/sethvargo/go-limiter | latest | Rate limiting | Token bucket algorithm, memory and Redis stores, RFC-compliant headers, zero external deps |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/neocotic/go-problem | latest | RFC 9457 Problem Details | Creating standardized error responses with proper HTTP problem detail format |
| github.com/mtiller/rfc8288 | latest | RFC 8288 Link headers | Generating RFC-compliant Link headers for cursor pagination |
| crypto/subtle | stdlib | Constant-time comparison | Token authentication (CRITICAL: prevents timing attacks on bearer tokens/API keys) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Huma v2 | go-swagger/go-openapi | More boilerplate, spec-first (not code-first), less ergonomic validation |
| Huma v2 | Manual OpenAPI generation | Full control but massive maintenance burden, error-prone, no validation integration |
| nyxstack/scalarui | Stoplight Elements | CDN dependency (violates single-binary constraint) |
| go-limiter | github.com/ulule/limiter | More features but heavier dependencies |
| Chi router | stdlib http.NewServeMux | Go 1.22+ stdlib viable but Chi has richer middleware ecosystem |

**Installation:**
```bash
go get -u github.com/danielgtaylor/huma/v2
go get -u github.com/danielgtaylor/huma/v2/adapters/humachi
go get -u github.com/go-chi/chi/v5
go get -u github.com/nyxstack/scalarui
go get -u github.com/rs/cors
go get -u github.com/sethvargo/go-limiter/httplimit
go get -u github.com/neocotic/go-problem
go get -u github.com/mtiller/rfc8288
```

## Architecture Patterns

### Recommended Project Structure
```
gen/
├── api/
│   ├── inputs.go         # Generated Huma Input structs (query, path, header params)
│   ├── outputs.go        # Generated Huma Output structs (response envelopes)
│   ├── register.go       # Generated route registration functions
│   └── transforms.go     # Schema to Input/Output conversion helpers
internal/
├── api/
│   ├── middleware/
│   │   ├── auth.go       # Bearer token + API key authentication
│   │   ├── ratelimit.go  # Rate limiting configuration
│   │   └── cors.go       # CORS configuration
│   ├── handlers/
│   │   └── resources.go  # Thin handlers that call action layer
│   └── server.go         # API server setup and middleware wiring
│   └── docs.go           # Scalar UI handler
cmd/
└── forge/
    └── routes.go          # CLI command to list routes grouped by resource
```

### Pattern 1: Generated Huma Input Structs
**What:** Code generator produces Input structs from schema definitions with proper validation tags
**When to use:** Every CRUD endpoint needs input validation for path params, query params, headers, and body
**Example:**
```go
// Source: Huma v2 documentation - https://huma.rocks/features/request-validation/
// Generated from schema definition
type ListBlogPostsInput struct {
    // Query parameters for pagination
    Cursor string `query:"cursor" doc:"Pagination cursor for next page"`
    Limit  int    `query:"limit" default:"10" minimum:"1" maximum:"100" doc:"Number of items per page"`

    // Query parameters for filtering
    PostID *int64 `query:"post_id" doc:"Filter by post ID"`

    // Authentication (populated by middleware)
    Authorization string `header:"Authorization" doc:"Bearer token or API key"`
}

type CreateBlogPostInput struct {
    // Request body
    Body struct {
        Title   string `json:"title" minLength:"1" maxLength:"200" doc:"Post title"`
        Content string `json:"content" minLength:"1" doc:"Post content"`
        Tags    []string `json:"tags,omitempty" maxItems:"10" doc:"Post tags"`
    }
}

type GetBlogPostInput struct {
    // Path parameters (always required)
    ID int64 `path:"id" doc:"Blog post ID"`
}

type UpdateBlogPostInput struct {
    ID int64 `path:"id" doc:"Blog post ID"`
    Body struct {
        Title   *string `json:"title,omitempty" minLength:"1" maxLength:"200"`
        Content *string `json:"content,omitempty" minLength:"1"`
        Tags    *[]string `json:"tags,omitempty" maxItems:"10"`
    }
}

type DeleteBlogPostInput struct {
    ID int64 `path:"id" doc:"Blog post ID"`
}
```

**Validation tag reference:**
- `minimum`, `maximum` — numeric bounds
- `minLength`, `maxLength` — string length
- `minItems`, `maxItems` — array length
- `pattern` — regex validation
- `enum:"foo,bar,baz"` — allowed values
- `default:"value"` — default if omitted
- `doc:"description"` — OpenAPI description
- `example:"value"` — OpenAPI example
- `required:"true"` — explicitly require (path params always required)

**Required field logic:**
1. Path parameters: ALWAYS required
2. Cookie, header, query: Optional by default unless `required:"true"`
3. Body fields: Required unless `omitempty`, `omitzero`, or `required:"false"`

### Pattern 2: Generated Huma Output Structs with Response Envelope
**What:** Code generator wraps responses in standardized envelope matching user decision
**When to use:** All API responses (single resources and lists)
**Example:**
```go
// Source: User decision in CONTEXT.md + Huma patterns
// Single resource response
type GetBlogPostOutput struct {
    Body struct {
        Data BlogPost `json:"data" doc:"Blog post resource"`
    }
}

// List response with pagination
type ListBlogPostsOutput struct {
    Body struct {
        Data       []BlogPost        `json:"data" doc:"List of blog posts"`
        Pagination PaginationMeta    `json:"pagination" doc:"Pagination metadata"`
    }
}

type PaginationMeta struct {
    NextCursor string `json:"next_cursor,omitempty" doc:"Cursor for next page"`
    HasMore    bool   `json:"has_more" doc:"Whether more results exist"`
}

// Error response (RFC 9457)
type ErrorOutput struct {
    Body struct {
        Type     string            `json:"type" doc:"URI identifying the problem type"`
        Title    string            `json:"title" doc:"Short human-readable summary"`
        Status   int               `json:"status" doc:"HTTP status code"`
        Detail   string            `json:"detail,omitempty" doc:"Explanation specific to this occurrence"`
        Instance string            `json:"instance,omitempty" doc:"URI identifying this specific occurrence"`
        Errors   []ValidationError `json:"errors,omitempty" doc:"Per-field validation errors"`
    }
}

type ValidationError struct {
    Field   string `json:"field" doc:"Field name that failed validation"`
    Message string `json:"message" doc:"Human-readable error message"`
}
```

### Pattern 3: Route Registration Functions
**What:** Generator produces functions that register all CRUD endpoints for a resource
**When to use:** Called during server initialization to wire handlers to Huma API
**Example:**
```go
// Source: Huma v2 registration patterns
// Generated registration function
func RegisterBlogPostRoutes(api huma.API, actions BlogPostActions) {
    // List
    huma.Register(api, huma.Operation{
        OperationID: "listBlogPosts",  // SDK generation: client.listBlogPosts()
        Method:      http.MethodGet,
        Path:        "/api/v1/blog-posts",
        Summary:     "List blog posts",
        Tags:        []string{"blog-posts"},
    }, func(ctx context.Context, input *ListBlogPostsInput) (*ListBlogPostsOutput, error) {
        // Call action layer (no business logic in handler)
        posts, nextCursor, hasMore, err := actions.List(ctx, input.Cursor, input.Limit, input.PostID)
        if err != nil {
            return nil, toProblemDetail(err)
        }

        return &ListBlogPostsOutput{
            Body: struct {
                Data       []BlogPost     `json:"data"`
                Pagination PaginationMeta `json:"pagination"`
            }{
                Data: posts,
                Pagination: PaginationMeta{
                    NextCursor: nextCursor,
                    HasMore:    hasMore,
                },
            },
        }, nil
    })

    // Get
    huma.Register(api, huma.Operation{
        OperationID: "getBlogPost",
        Method:      http.MethodGet,
        Path:        "/api/v1/blog-posts/{id}",
        Summary:     "Get a blog post",
        Tags:        []string{"blog-posts"},
    }, func(ctx context.Context, input *GetBlogPostInput) (*GetBlogPostOutput, error) {
        post, err := actions.Get(ctx, input.ID)
        if err != nil {
            return nil, toProblemDetail(err)
        }

        return &GetBlogPostOutput{
            Body: struct {
                Data BlogPost `json:"data"`
            }{Data: post},
        }, nil
    })

    // Create
    huma.Register(api, huma.Operation{
        OperationID: "createBlogPost",
        Method:      http.MethodPost,
        Path:        "/api/v1/blog-posts",
        Summary:     "Create a blog post",
        Tags:        []string{"blog-posts"},
    }, func(ctx context.Context, input *CreateBlogPostInput) (*GetBlogPostOutput, error) {
        post, err := actions.Create(ctx, input.Body.Title, input.Body.Content, input.Body.Tags)
        if err != nil {
            return nil, toProblemDetail(err)
        }

        return &GetBlogPostOutput{
            Body: struct {
                Data BlogPost `json:"data"`
            }{Data: post},
        }, nil
    })

    // Update
    huma.Register(api, huma.Operation{
        OperationID: "updateBlogPost",
        Method:      http.MethodPut,
        Path:        "/api/v1/blog-posts/{id}",
        Summary:     "Update a blog post",
        Tags:        []string{"blog-posts"},
    }, func(ctx context.Context, input *UpdateBlogPostInput) (*GetBlogPostOutput, error) {
        post, err := actions.Update(ctx, input.ID, input.Body.Title, input.Body.Content, input.Body.Tags)
        if err != nil {
            return nil, toProblemDetail(err)
        }

        return &GetBlogPostOutput{
            Body: struct {
                Data BlogPost `json:"data"`
            }{Data: post},
        }, nil
    })

    // Delete
    huma.Register(api, huma.Operation{
        OperationID: "deleteBlogPost",
        Method:      http.MethodDelete,
        Path:        "/api/v1/blog-posts/{id}",
        Summary:     "Delete a blog post",
        Tags:        []string{"blog-posts"},
    }, func(ctx context.Context, input *DeleteBlogPostInput) (*struct{}, error) {
        err := actions.Delete(ctx, input.ID)
        if err != nil {
            return nil, toProblemDetail(err)
        }

        return &struct{}{}, nil
    })
}
```

**OperationID naming conventions (for SDK generation):**
- Verb + noun format: `listPosts`, `getPost`, `createPost`, `updatePost`, `deletePost`
- Standard verbs: `list` (collections), `get` (single), `create` (POST), `update` (PUT/PATCH), `delete` (DELETE)
- Singular resource name in camelCase
- Max 30 characters
- Alphanumeric only (no special chars except camelCase boundaries)
- Unique across entire spec (case-insensitive uniqueness)

### Pattern 4: Middleware Wiring Order
**What:** Middleware execution order matters; wire in specific sequence to avoid issues
**When to use:** Server initialization before route registration
**Example:**
```go
// Source: Huma middleware docs + industry best practices
func SetupAPI(router chi.Router, config *Config) huma.API {
    // 1. Router-level middleware (Chi-specific, runs before Huma)
    router.Use(middleware.RealIP)           // Extract real IP (for rate limiting)
    router.Use(loggingMiddleware)           // Logging FIRST (capture all requests)
    router.Use(panicRecoveryMiddleware)     // Panic recovery SECOND (catch all panics)

    // 2. Create Huma API instance
    api := humachi.New(router, huma.Config{
        OpenAPIPath: "/api/openapi",  // Serves .json and .yaml
        DocsPath:    "/api/docs",     // Scalar UI
        Info: &huma.Info{
            Title:   "Forge API",
            Version: "1.0.0",
            Description: "Production-ready REST API generated by Forge",
            Contact: &huma.Contact{
                Name:  "API Support",
                Email: "api@example.com",
            },
            License: &huma.License{
                Name: "MIT",
                URL:  "https://opensource.org/licenses/MIT",
            },
        },
    })

    // 3. Router-agnostic Huma middleware (runs within Huma processing chain)
    // CORS THIRD (before authentication, needs to handle preflight)
    api.UseMiddleware(corsMiddleware(config.CORS))

    // Rate limiting FOURTH (before auth, prevents auth brute force)
    api.UseMiddleware(rateLimitMiddleware(config.RateLimit))

    // Authentication LAST (after rate limit, before handlers)
    api.UseMiddleware(authenticationMiddleware(config.Auth))

    return api
}
```

**Middleware ordering principle:**
1. Logging — capture ALL requests (first to run)
2. Panic recovery — catch ALL panics including from later middleware
3. CORS — handle preflight OPTIONS requests early
4. Rate limiting — prevent resource exhaustion before expensive operations
5. Authentication — validate credentials after rate limiting
6. Business logic — handlers run last

### Pattern 5: Scalar UI Embedding
**What:** Embed Scalar documentation UI directly in binary without CDN dependencies
**When to use:** Serving interactive API docs at `/api/docs`
**Example:**
```go
// Source: nyxstack/scalarui package documentation
import "github.com/nyxstack/scalarui"

func RegisterDocsHandler(router chi.Router, api huma.API) {
    config := scalarui.NewConfig().
        WithTitle("Forge API Documentation").
        WithDescription("Production-ready REST API").
        WithURL("http://localhost:8080/api/openapi.json").  // Point to Huma's OpenAPI endpoint
        WithTheme("purple").
        WithDarkMode(true).
        WithSidebar(true).
        WithInteractive(true).  // Enable "Try It" feature
        WithExpandAllResponses(true).
        HideDownload()  // Optional: hide download button

    ui := scalarui.New(config)

    // Huma already serves /api/docs, but for custom control:
    router.Get("/api/docs", func(w http.ResponseWriter, r *http.Request) {
        html, err := ui.Render()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Write([]byte(html))
    })
}
```

**Note:** Huma v2 has built-in docs support at `DocsPath`, but for single-binary embedding without CDN, use scalarui package which renders complete HTML with embedded assets.

### Pattern 6: Authentication Middleware (Bearer + API Key)
**What:** Validate opaque bearer tokens and API keys from database, prevent timing attacks
**When to use:** All authenticated endpoints (default for all API routes)
**Example:**
```go
// Source: Security best practices + crypto/subtle for timing attack prevention
import (
    "crypto/subtle"
    "strings"
)

type AuthMiddleware struct {
    tokenStore TokenStore
    apiKeyStore APIKeyStore
}

func (m *AuthMiddleware) Handle(ctx huma.Context, next func(huma.Context)) {
    authHeader := ctx.Header("Authorization")
    if authHeader == "" {
        huma.WriteErr(ctx, 401, "Missing Authorization header")
        return
    }

    // Determine auth type by prefix
    if strings.HasPrefix(authHeader, "Bearer ") {
        token := strings.TrimPrefix(authHeader, "Bearer ")
        m.validateBearerToken(ctx, token, next)
    } else if strings.HasPrefix(authHeader, "forg_live_") || strings.HasPrefix(authHeader, "forg_test_") {
        m.validateAPIKey(ctx, authHeader, next)
    } else {
        huma.WriteErr(ctx, 401, "Invalid authorization format")
        return
    }
}

func (m *AuthMiddleware) validateBearerToken(ctx huma.Context, token string, next func(huma.Context)) {
    // Lookup token in database
    storedToken, err := m.tokenStore.GetByToken(ctx.Context(), token)
    if err != nil {
        huma.WriteErr(ctx, 401, "Invalid or expired token")
        return
    }

    // CRITICAL: Use constant-time comparison to prevent timing attacks
    // DO NOT use: if token == storedToken.Token (vulnerable!)
    if subtle.ConstantTimeCompare([]byte(token), []byte(storedToken.Token)) != 1 {
        huma.WriteErr(ctx, 401, "Invalid token")
        return
    }

    // Check expiration
    if storedToken.ExpiresAt.Before(time.Now()) {
        huma.WriteErr(ctx, 401, "Token expired")
        return
    }

    // Attach user to context
    ctx = huma.WithValue(ctx, "user_id", storedToken.UserID)
    next(ctx)
}

func (m *AuthMiddleware) validateAPIKey(ctx huma.Context, key string, next func(huma.Context)) {
    // Lookup API key in database
    storedKey, err := m.apiKeyStore.GetByKey(ctx.Context(), key)
    if err != nil {
        huma.WriteErr(ctx, 401, "Invalid API key")
        return
    }

    // CRITICAL: Constant-time comparison
    if subtle.ConstantTimeCompare([]byte(key), []byte(storedKey.Key)) != 1 {
        huma.WriteErr(ctx, 401, "Invalid API key")
        return
    }

    // Check expiration and revocation
    if storedKey.RevokedAt != nil {
        huma.WriteErr(ctx, 401, "API key revoked")
        return
    }
    if storedKey.ExpiresAt != nil && storedKey.ExpiresAt.Before(time.Now()) {
        huma.WriteErr(ctx, 401, "API key expired")
        return
    }

    // Attach API key metadata to context
    ctx = huma.WithValue(ctx, "api_key_id", storedKey.ID)
    ctx = huma.WithValue(ctx, "api_key_scopes", storedKey.Scopes)
    next(ctx)
}
```

**CRITICAL SECURITY:** ALWAYS use `crypto/subtle.ConstantTimeCompare` for token/key comparison to prevent timing attacks. Regular string comparison (`==`) leaks timing information allowing brute-force attacks.

### Pattern 7: RFC 9457 Error Response Conversion
**What:** Convert action layer errors to RFC 9457 Problem Details format
**When to use:** Error handling in all route handlers
**Example:**
```go
// Source: RFC 9457 spec + go-problem library patterns
import "github.com/neocotic/go-problem"

// Map application errors to HTTP problem details
func toProblemDetail(err error) error {
    // Check for known error types from action layer
    switch e := err.(type) {
    case *NotFoundError:
        return huma.Error404NotFound(e.Message, err)
    case *ValidationError:
        // Return validation errors with per-field detail
        return &huma.ErrorDetail{
            Status: 400,
            Detail: "Validation failed",
            Errors: convertFieldErrors(e.Fields),
        }
    case *UnauthorizedError:
        return huma.Error401Unauthorized(e.Message)
    case *ForbiddenError:
        return huma.Error403Forbidden(e.Message)
    default:
        // Internal server error for unknown errors
        return huma.Error500InternalServerError("An unexpected error occurred", err)
    }
}

func convertFieldErrors(fields map[string]string) []map[string]any {
    errors := make([]map[string]any, 0, len(fields))
    for field, message := range fields {
        errors = append(errors, map[string]any{
            "field":   field,
            "message": message,
        })
    }
    return errors
}
```

Huma's error helpers automatically generate RFC 9457 compliant responses with `application/problem+json` content type.

### Anti-Patterns to Avoid

**❌ Business logic in handlers:**
```go
// WRONG: Handler contains database queries and business rules
huma.Register(api, op, func(ctx context.Context, input *Input) (*Output, error) {
    post := db.Query("SELECT * FROM posts WHERE id = ?", input.ID)
    if post.AuthorID != currentUser.ID {
        return nil, errors.New("forbidden")
    }
    // ... more logic
})
```

**✅ Correct: Handlers delegate to action layer:**
```go
// RIGHT: Handler is thin orchestration, delegates to actions
huma.Register(api, op, func(ctx context.Context, input *Input) (*Output, error) {
    post, err := actions.Get(ctx, input.ID)
    if err != nil {
        return nil, toProblemDetail(err)
    }
    return &Output{Body: struct{Data Post `json:"data"`}{Data: post}}, nil
})
```

**❌ Non-constant time token comparison:**
```go
// WRONG: Vulnerable to timing attacks
if providedToken == storedToken {
    // authenticate
}
```

**✅ Correct: Constant-time comparison:**
```go
// RIGHT: Prevents timing attacks
if subtle.ConstantTimeCompare([]byte(providedToken), []byte(storedToken)) == 1 {
    // authenticate
}
```

**❌ Inconsistent operationId naming:**
```go
// WRONG: Inconsistent, not SDK-friendly
OperationID: "blog-posts-list"      // Mix of formats
OperationID: "GetAllBlogPosts"      // Too verbose
OperationID: "posts_get"            // Wrong separator
```

**✅ Correct: Consistent verb+noun camelCase:**
```go
// RIGHT: SDK generators produce clean method names
OperationID: "listBlogPosts"   // client.listBlogPosts()
OperationID: "getBlogPost"     // client.getBlogPost(id)
OperationID: "createBlogPost"  // client.createBlogPost(data)
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| OpenAPI spec generation | Custom reflection-based generator | Huma v2 automatic generation | Huma handles complex validation rules, recursive types, oneOf/anyOf, content negotiation; maintaining custom generator is error-prone |
| Input validation | Manual validation in handlers | Huma struct tags + JSON Schema | Huma provides 20+ validators, automatic error formatting, OpenAPI sync, locale support; manual validation duplicates effort and drifts from docs |
| CORS preflight handling | Custom OPTIONS handlers | github.com/rs/cors | CORS spec has subtle edge cases (credentials + wildcards, Vary headers, private networks); battle-tested library handles all scenarios |
| Rate limiting algorithm | Custom token bucket | github.com/sethvargo/go-limiter | Token bucket has correctness issues (race conditions, timer drift); go-limiter is benchmarked, tested, supports distributed stores |
| RFC 9457 error formatting | Manual JSON error responses | Huma error helpers + go-problem | RFC 9457 compliance requires specific content-type, optional fields, extension members; Huma integrates seamlessly |
| API documentation UI | Custom HTML + JavaScript | Scalar UI via nyxstack/scalarui | Modern API docs need Try-It client, OAuth flows, code samples in multiple languages, theme support; thousands of engineering hours in Scalar |
| Constant-time string comparison | Custom comparison loop | crypto/subtle.ConstantTimeCompare | Timing attack prevention requires assembly-level control; subtle package is compiler-optimized and audited |

**Key insight:** OpenAPI ecosystem has mature, battle-tested solutions. Custom implementations introduce security risks (timing attacks, CORS misconfig), maintenance burden (OpenAPI spec changes), and feature gaps (SDK generation compatibility). Code generation + proven libraries = production-ready faster.

## Common Pitfalls

### Pitfall 1: Timing Attacks on Token Validation
**What goes wrong:** Using `==` or `strings.EqualFold` to compare authentication tokens allows attackers to deduce token contents by measuring response times.
**Why it happens:** Standard string comparison short-circuits on first mismatch; attacker benchmarks responses to determine correct characters byte-by-byte.
**How to avoid:**
- ALWAYS use `crypto/subtle.ConstantTimeCompare` for any secret comparison (tokens, API keys, passwords)
- Function returns 1 for equal, 0 for not equal
- Comparison time is constant regardless of where strings differ
**Warning signs:** Authentication succeeds/fails in variable time; string comparison of sensitive values in code review.
**References:** [Avoiding Timing Attacks with Constant-Time Comparisons in Go](https://www.slingacademy.com/article/avoiding-timing-attacks-with-constant-time-comparisons-in-go/), [GoCD timing attack vulnerability](https://github.com/gocd/gocd/security/advisories/GHSA-999p-fp84-jcpq)

### Pitfall 2: Middleware Ordering Breaks Authentication
**What goes wrong:** Authentication middleware runs before rate limiting, allowing brute-force attacks to overwhelm server. Or CORS runs after authentication, breaking preflight requests.
**Why it happens:** Middleware order is not obvious; incorrect order silently breaks security assumptions.
**How to avoid:**
- Standard order: Logging → Panic Recovery → CORS → Rate Limiting → Authentication → Handlers
- CORS must run before authentication (handles OPTIONS preflight which has no auth)
- Rate limiting must run before authentication (prevents auth brute-force)
- Document middleware order in code comments
**Warning signs:** CORS errors in browser for authenticated endpoints; authentication rate limiting ineffective; panic recovery misses errors.

### Pitfall 3: Incorrect HTTP Status Codes Break Client Tools
**What goes wrong:** Returning 200 OK with error in body, or 500 for validation errors, breaks automated tooling and client retry logic.
**Why it happens:** Developers focus on response body, ignore HTTP semantics.
**How to avoid:**
- 400 Bad Request: client error (malformed JSON, invalid params)
- 401 Unauthorized: missing/invalid authentication
- 403 Forbidden: authenticated but lacks permission
- 404 Not Found: resource doesn't exist
- 422 Unprocessable Entity: valid request but semantic validation failed
- 500 Internal Server Error: server bug (unexpected panic, database down)
- Use Huma's error helpers: `huma.Error400BadRequest`, `huma.Error404NotFound`, etc.
**Warning signs:** Clients retry non-retryable errors; monitoring tools miss errors; HTTP caches store error responses.
**References:** [Common REST API Pitfalls](https://zuplo.com/learning-center/common-pitfalls-in-restful-api-design), [Stop These 20 REST API Status Code Errors](https://medium.com/tools-trips/stop-these-20-rest-api-status-code-errors-a-developers-survival-guide-c5dd51b5dc55)

### Pitfall 4: Missing operationId Breaks SDK Generation
**What goes wrong:** Omitting or inconsistently naming operationIds causes SDK generators to create unpredictable method names (e.g., `getApiV1BlogPosts` instead of `listBlogPosts`).
**Why it happens:** OpenAPI spec makes operationId optional; developers skip it or use path-based defaults.
**How to avoid:**
- ALWAYS set explicit operationId for every operation
- Use verb+noun camelCase: `listPosts`, `getPost`, `createPost`, `updatePost`, `deletePost`
- Keep under 30 characters
- Verify uniqueness (case-insensitive) across entire spec
- Run spectral linting to catch issues: `spectral lint openapi.yaml`
**Warning signs:** Generated SDK method names are verbose or inconsistent; SDK breaking changes when adding routes.
**References:** [OpenAPI operationId Best Practices](https://www.stainless.com/sdk-api-best-practices/what-is-openapi-operationid-and-why-it-matters-for-sdks), [Naming Operation IDs for Better SDK Method Names](https://konfigthis.com/docs/tutorials/naming-operation-ids/)

### Pitfall 5: CORS Misconfiguration Security Holes
**What goes wrong:** Setting `AllowedOrigins: ["*"]` with `AllowCredentials: true` creates security vulnerability; wildcards with credentials don't work and expose sensitive data.
**Why it happens:** Developers encounter CORS errors, try permissive config to "fix" without understanding spec.
**How to avoid:**
- NEVER combine wildcard origins with credentials
- Production: explicit allowlist of origins (e.g., `[]string{"https://app.example.com"}`)
- Development: use specific localhost origins (e.g., `http://localhost:3000`)
- Set reasonable MaxAge (300-600 seconds) to cache preflight
- Limit AllowedHeaders to required headers (don't use `*`)
**Warning signs:** Browser console shows CORS errors despite wildcard; credentials not sent to API; security audit flags CORS config.
**References:** [Solving Golang CORS Issues](https://www.stackhawk.com/blog/golang-cors-guide-what-it-is-and-how-to-enable-it/), [Fearless CORS Design Philosophy](https://jub0bs.com/posts/2023-02-08-fearless-cors/)

### Pitfall 6: Validation Errors Without Field-Level Detail
**What goes wrong:** Returning generic "validation failed" message forces clients to guess which field caused error, poor UX.
**Why it happens:** Error handling returns first validation error or aggregates without field context.
**How to avoid:**
- Huma automatically provides per-field validation errors via struct tags
- For custom validation in resolvers, return structured errors with field names
- RFC 9457 response includes `errors` array: `[{"field": "email", "message": "invalid format"}]`
- Always include field path for nested objects: `"user.address.zip_code"`
**Warning signs:** Support requests asking "what's wrong with my request?"; clients show generic error messages.

### Pitfall 7: Forgetting to Set Response Content-Type
**What goes wrong:** Browsers and tools misinterpret JSON responses as HTML, causing rendering issues or security warnings.
**Why it happens:** Huma handles this automatically, but custom handlers (like Scalar docs) must set explicitly.
**How to avoid:**
- Huma sets `Content-Type: application/json` automatically for registered operations
- For custom handlers: `w.Header().Set("Content-Type", "application/json; charset=utf-8")`
- For error responses: Huma uses `application/problem+json` (RFC 9457)
- For HTML docs: `w.Header().Set("Content-Type", "text/html; charset=utf-8")`
**Warning signs:** Browser prompts to download JSON responses; JSON rendered as plain text; security scanners flag missing charset.

### Pitfall 8: Rate Limiting Per-Endpoint Instead of Per-Client
**What goes wrong:** Naive rate limiting counts requests per endpoint, allowing clients to bypass limits by rotating endpoints.
**Why it happens:** Simple implementations track endpoint counters without client identification.
**How to avoid:**
- Use client identifier as rate limit key (IP address, user ID, API key ID)
- go-limiter's `IPKeyFunc()` extracts IP from X-Forwarded-For / X-Real-IP headers
- For authenticated APIs: use user ID or API key ID as key
- For multi-tenant: combine tenant ID + user ID
- Configure per-client limits in forge.toml, apply globally
**Warning signs:** Rate limits ineffective; single client overwhelms API by distributing requests across endpoints.

### Pitfall 9: Response Envelope Inconsistency
**What goes wrong:** Some endpoints return `{"data": ...}`, others return raw objects; clients need conditional parsing logic.
**Why it happens:** Copy-paste from different code examples; mixing conventions.
**How to avoid:**
- Generator enforces consistent envelope: ALWAYS `{"data": ...}` for single resources
- Lists: `{"data": [...], "pagination": {...}}`
- Errors: RFC 9457 format (no envelope)
- Document envelope contract in OpenAPI spec examples
**Warning signs:** Client code has `if response.data then ... else ...`; inconsistent TypeScript types.

### Pitfall 10: Exposing Internal Database IDs in Cursor Pagination
**What goes wrong:** Using sequential database IDs as cursors leaks business metrics (growth rate, total records) and enables enumeration attacks.
**Why it happens:** Cursors often derived from ORDER BY id LIMIT offset pattern.
**How to avoid:**
- Use opaque cursors: base64-encode cursor data
- Include timestamp + ID for stable ordering: `base64(timestamp:id)`
- Never use raw database IDs or predictable sequences
- Validate cursor format server-side to prevent injection
**Warning signs:** Cursors are numeric and sequential; competitors can estimate database size; enumeration attacks succeed.

## Code Examples

Verified patterns from official sources:

### Huma Handler Registration
```go
// Source: https://pkg.go.dev/github.com/danielgtaylor/huma/v2
package main

import (
    "context"
    "fmt"
    "net/http"
    "github.com/danielgtaylor/huma/v2"
    "github.com/danielgtaylor/huma/v2/adapters/humachi"
    "github.com/go-chi/chi/v5"
)

type GreetingInput struct {
    Name string `path:"name" maxLength:"30" example:"world" doc:"Name to greet"`
}

type GreetingOutput struct {
    Body struct {
        Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
    }
}

func main() {
    router := chi.NewMux()
    api := humachi.New(router, huma.DefaultConfig("My API", "1.0.0"))

    huma.Register(api, huma.Operation{
        OperationID: "getGreeting",
        Method:      http.MethodGet,
        Path:        "/greeting/{name}",
        Summary:     "Get a greeting",
        Tags:        []string{"greetings"},
    }, func(ctx context.Context, input *GreetingInput) (*GreetingOutput, error) {
        resp := &GreetingOutput{}
        resp.Body.Message = fmt.Sprintf("Hello, %s!", input.Name)
        return resp, nil
    })

    http.ListenAndServe(":8888", router)
}
```

### CORS Middleware Configuration
```go
// Source: https://pkg.go.dev/github.com/rs/cors
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/rs/cors"
)

func setupCORS() *cors.Cors {
    return cors.New(cors.Options{
        AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        ExposedHeaders:   []string{"X-Total-Count", "X-Page-Number"},
        AllowCredentials: true,
        MaxAge:           600, // 10 minutes
        Debug:            false, // Disable in production
    })
}

func main() {
    router := chi.NewRouter()

    // Apply CORS middleware at router level
    corsHandler := setupCORS()
    router.Use(corsHandler.Handler)

    // ... register routes

    http.ListenAndServe(":8080", router)
}
```

### Rate Limiting Middleware
```go
// Source: https://github.com/sethvargo/go-limiter
package main

import (
    "net/http"
    "time"
    "github.com/sethvargo/go-limiter/httplimit"
    "github.com/sethvargo/go-limiter/memorystore"
)

func setupRateLimiting() (*httplimit.Middleware, error) {
    // Create memory store (100 requests per minute per IP)
    store, err := memorystore.New(&memorystore.Config{
        Tokens:   100,
        Interval: time.Minute,
    })
    if err != nil {
        return nil, err
    }

    // Create middleware with IP-based key function
    middleware, err := httplimit.NewMiddleware(store, httplimit.IPKeyFunc())
    if err != nil {
        return nil, err
    }

    return middleware, nil
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/v1/posts", handlePosts)

    rateLimiter, _ := setupRateLimiting()

    // Wrap handler with rate limiting
    http.ListenAndServe(":8080", rateLimiter.Handle(mux))
}
```

### Constant-Time Token Comparison
```go
// Source: crypto/subtle package documentation
package main

import (
    "crypto/subtle"
    "net/http"
)

type TokenStore interface {
    GetByToken(token string) (*Token, error)
}

type Token struct {
    Token     string
    UserID    int64
    ExpiresAt time.Time
}

func validateToken(provided string, store TokenStore) (int64, error) {
    stored, err := store.GetByToken(provided)
    if err != nil {
        return 0, err
    }

    // CRITICAL: Use constant-time comparison to prevent timing attacks
    // Returns 1 if equal, 0 if not equal
    if subtle.ConstantTimeCompare([]byte(provided), []byte(stored.Token)) != 1 {
        return 0, errors.New("invalid token")
    }

    if stored.ExpiresAt.Before(time.Now()) {
        return 0, errors.New("token expired")
    }

    return stored.UserID, nil
}
```

### Scalar UI Embedding
```go
// Source: https://pkg.go.dev/github.com/nyxstack/scalarui
package main

import (
    "fmt"
    "net/http"
    "github.com/nyxstack/scalarui"
)

func setupDocs(openAPIURL string) http.HandlerFunc {
    config := scalarui.NewConfig().
        WithTitle("Forge API Documentation").
        WithDescription("Production-ready REST API").
        WithURL(openAPIURL).
        WithTheme("purple").
        WithDarkMode(true).
        WithSidebar(true).
        WithInteractive(true).
        WithExpandAllResponses(true).
        HideDownload()

    ui := scalarui.New(config)

    return func(w http.ResponseWriter, r *http.Request) {
        html, err := ui.Render()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        fmt.Fprint(w, html)
    }
}

func main() {
    http.HandleFunc("/api/docs", setupDocs("http://localhost:8080/api/openapi.json"))
    http.ListenAndServe(":8080", nil)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Swagger 2.0 (OpenAPI 2.0) | OpenAPI 3.1 | 2021 | JSON Schema 2020-12 support, webhooks, better discriminators, more accurate schema validation |
| JWT bearer tokens | Opaque tokens in database | 2020-2022 trend | Revocation control, audit trails, no crypto dependencies, queryable sessions |
| Offset pagination (`?page=5`) | Cursor pagination | 2018-2020 | Stable results during concurrent modifications, better performance on large offsets |
| Custom error JSON | RFC 9457 Problem Details | 2023 (RFC published) | Standardized error format, tooling support, consistent client handling |
| Stoplight Elements | Scalar UI | 2023-2024 | Better performance, modern design, built-in Try-It client, no React dependency |
| Code-first with manual OpenAPI | Code-first with auto-generation (Huma) | 2022-2024 | Zero drift between code and spec, validation integrated, fewer bugs |
| API versioning in headers | Path-based versioning (`/api/v1`) | Industry standard | Easier debugging, better caching, simpler routing |
| go-swagger | Huma v2 | 2023-2024 | Spec-first vs code-first, Huma has better DX, less boilerplate |

**Deprecated/outdated:**
- **Swagger 2.0**: Replaced by OpenAPI 3.0/3.1; lacks discriminators, webhooks, modern JSON Schema features
- **Offset pagination**: Replaced by cursor pagination; offset is unstable during writes and slow for large offsets
- **JWT for session tokens**: Still used but opaque tokens preferred for revocable sessions; JWTs better for stateless service-to-service auth
- **Stoplight Elements**: Scalar UI is newer, faster, better maintained
- **go-swagger**: Huma v2 is more actively maintained and has better ergonomics

## Recommendations (Claude's Discretion Areas)

### Rate Limiting Strategy
**Recommendation:** Per-client (IP or API key) token bucket with tiered limits.

**forge.toml configuration:**
```toml
[api.rate_limit]
enabled = true
strategy = "token_bucket"

# Default limits (unauthenticated IPs)
[api.rate_limit.default]
tokens = 100      # requests per interval
interval = "1m"   # 1 minute

# Authenticated user limits (higher)
[api.rate_limit.authenticated]
tokens = 1000
interval = "1m"

# API key limits (highest)
[api.rate_limit.api_key]
tokens = 5000
interval = "1m"

# Per-endpoint overrides (optional)
[api.rate_limit.endpoints]
"/api/v1/search" = { tokens = 10, interval = "1m" }  # Expensive search endpoint
```

**Rationale:** Token bucket smooths bursts, per-client prevents abuse, tiered limits reward authentication without being overly restrictive. Memory store sufficient for single-instance, Redis store for multi-instance deployments.

### CORS Default Configuration
**Recommendation:** Restrictive defaults, require explicit origin allowlist in production.

**forge.toml configuration:**
```toml
[api.cors]
enabled = true
allowed_origins = []  # Empty = require explicit configuration
allowed_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
allowed_headers = ["Accept", "Authorization", "Content-Type"]
exposed_headers = ["X-Total-Count"]
allow_credentials = true
max_age = 600  # 10 minutes

# Environment-specific overrides
[api.cors.development]
allowed_origins = ["http://localhost:3000", "http://localhost:5173"]

[api.cors.production]
allowed_origins = ["https://example.com"]  # Must be explicitly set
```

**Rationale:** Security-first defaults prevent accidental exposure. Separate dev/prod configs reduce friction during development while maintaining production safety.

### Huma Middleware Wiring Order
**Recommendation:** Router-level → Huma-level with specific sequence.

**Execution order:**
1. **Router-level (Chi):** RealIP → Logging → Panic Recovery
2. **Huma-level:** CORS → Rate Limiting → Authentication

**Code pattern:**
```go
// Router-level middleware (runs before Huma)
router.Use(middleware.RealIP)      // Extract real client IP
router.Use(loggingMiddleware)      // Log all requests
router.Use(panicRecoveryMiddleware) // Catch panics

// Create Huma API
api := humachi.New(router, config)

// Huma-level middleware (router-agnostic)
api.UseMiddleware(corsMiddleware)        // Handle preflight, runs before auth
api.UseMiddleware(rateLimitMiddleware)   // Prevent brute-force before auth
api.UseMiddleware(authMiddleware)        // Validate credentials last
```

**Rationale:** This order ensures (1) all requests logged including panics, (2) CORS handles OPTIONS before auth, (3) rate limiting prevents auth brute-force, (4) panic recovery catches errors from all middleware.

### OpenAPI Spec Metadata
**Recommendation:** Complete metadata for professional developer experience.

**Config in Huma:**
```go
api := humachi.New(router, huma.Config{
    OpenAPIPath: "/api/openapi",
    DocsPath:    "/api/docs",
    Info: &huma.Info{
        Title:       "Forge API",
        Version:     "1.0.0",
        Description: "Production-ready REST API generated by Forge framework. " +
                     "Provides CRUD operations for all defined resources with " +
                     "automatic validation, pagination, and authentication.",
        Contact: &huma.Contact{
            Name:  "API Support",
            Email: "api@example.com",
            URL:   "https://example.com/support",
        },
        License: &huma.License{
            Name: "MIT",
            URL:  "https://opensource.org/licenses/MIT",
        },
    },
    Servers: []*huma.Server{
        {
            URL:         "https://api.example.com",
            Description: "Production",
        },
        {
            URL:         "https://api.staging.example.com",
            Description: "Staging",
        },
    },
})
```

**Rationale:** Complete metadata enables SDK generation, provides context for developers, and looks professional in generated docs. Version from git tags or build process.

### Scalar UI Embedding Strategy
**Recommendation:** Use `nyxstack/scalarui` package with rendered HTML (not go:embed of CDN assets).

**Implementation:**
- The scalarui package renders complete HTML with embedded Scalar UI JavaScript inline
- No go:embed needed — package handles all asset embedding internally
- Single `.Render()` call produces complete HTML string
- Zero external dependencies at runtime (no CDN)

**Rationale:** The nyxstack/scalarui package already embeds all Scalar UI assets in the package itself. Using go:embed would be redundant. The package's `.Render()` method generates complete self-contained HTML that includes all JavaScript, CSS, and assets inline. This matches the single-binary philosophy perfectly and requires no build-time asset management.

## Open Questions

1. **API Key Prefix Validation in Schema**
   - What we know: Stripe uses `sk_live_`, `sk_test_` prefixes; format is `prefix_environment_random`
   - What's unclear: Should schema validation enforce prefix format, or just generator + middleware?
   - Recommendation: Enforce in both — generator validates at creation time, middleware validates at runtime (defense in depth)

2. **Cursor Pagination Encoding Strategy**
   - What we know: Cursors should be opaque; common pattern is base64(timestamp:id)
   - What's unclear: Should encoding happen in action layer or API layer?
   - Recommendation: Action layer returns raw cursor data (struct with timestamp+id), API layer base64-encodes for wire format; keeps actions testable without encoding concerns

3. **Rate Limit Headers Specification**
   - What we know: go-limiter sets `X-RateLimit-*` headers; RFCs exist but no single standard
   - What's unclear: Use `X-RateLimit-*` (common) or `RateLimit-*` (draft RFC 6585bis)?
   - Recommendation: Use `X-RateLimit-*` (broader ecosystem support); document in OpenAPI responses

4. **Authentication Middleware Position in Huma Chain**
   - What we know: Middleware order matters; CORS before auth, rate limiting before auth
   - What's unclear: Should auth be Huma middleware or Chi middleware?
   - Recommendation: Huma middleware — allows operation-level control (public endpoints opt-out), access to Huma context, better OpenAPI security scheme integration

## Sources

### Primary (HIGH confidence)
- [Huma v2 Go Package Documentation](https://pkg.go.dev/github.com/danielgtaylor/huma/v2) - API reference, validation tags, middleware patterns
- [Huma Features Overview](https://huma.rocks/features/) - Core capabilities, production features
- [Huma Validation Documentation](https://huma.rocks/features/request-validation/) - Struct tag validation, required field logic
- [Huma Middleware Documentation](https://huma.rocks/features/middleware/) - Middleware types, execution order, patterns
- [nyxstack/scalarui Go Package](https://pkg.go.dev/github.com/nyxstack/scalarui) - Scalar UI embedding API
- [rs/cors Go Package](https://pkg.go.dev/github.com/rs/cors) - CORS configuration options
- [go-limiter Repository](https://github.com/sethvargo/go-limiter) - Rate limiting strategies, store implementations
- [go-problem Repository](https://github.com/neocotic/go-problem) - RFC 9457 implementation patterns

### Secondary (MEDIUM confidence)
- [OpenAPI operationId Best Practices](https://www.stainless.com/sdk-api-best-practices/what-is-openapi-operationid-and-why-it-matters-for-sdks) - SDK generation naming conventions
- [Naming Operation IDs for Better SDK Method Names](https://konfigthis.com/docs/tutorials/naming-operation-ids/) - operationId conventions
- [Spectral OpenAPI Linter](https://stoplight.io/open-source/spectral) - Linting rules for SDK generation
- [Avoiding Timing Attacks in Go](https://www.slingacademy.com/article/avoiding-timing-attacks-with-constant-time-comparisons-in-go/) - crypto/subtle usage
- [Common REST API Pitfalls](https://zuplo.com/learning-center/common-pitfalls-in-restful-api-design) - HTTP status code errors
- [API Pagination Best Practices](https://www.speakeasy.com/api-design/pagination) - Cursor vs offset patterns
- [RFC 9457 Problem Details Overview](https://www.rfc-editor.org/rfc/rfc9457.html) - Error response standard
- [CORS Security Best Practices](https://www.stackhawk.com/blog/golang-cors-guide-what-it-is-and-how-to-enable-it/) - Production configuration

### Tertiary (LOW confidence, marked for validation)
- [Stripe Keys and IDs Gist](https://gist.github.com/fnky/76f533366f75cf75802c8052b577e2a5) - API key prefix patterns (community documentation, verify with official Stripe docs)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Huma v2, Chi, and supporting libraries verified from official docs and package registries
- Architecture patterns: HIGH - Patterns derived from official Huma documentation, verified with code examples
- Authentication security: HIGH - crypto/subtle requirement verified from Go stdlib docs and security advisories
- Middleware ordering: MEDIUM - Best practices from Huma docs + industry patterns, but not dogmatic (context-dependent)
- Pitfalls: MEDIUM-HIGH - Timing attacks and CORS misconfig verified from CVEs; other pitfalls from multiple sources
- Recommendations (discretion areas): MEDIUM - Based on research and best practices, but subject to project-specific constraints

**Research date:** 2026-02-17
**Valid until:** 2026-03-19 (30 days — Huma is stable but ecosystem evolves)
