# Phase 4: Action Layer & Error Handling - Research

**Researched:** 2026-02-16
**Domain:** Go service layer patterns, error handling architecture, database error mapping, panic recovery
**Confidence:** HIGH

## Summary

Phase 4 requires implementing a shared action/service layer that prevents business logic duplication between HTML (Datastar/Templ) and API (Huma) handlers. Research confirms this follows established Go patterns: the service layer isolates business logic from transport concerns, enabling multiple client implementations (HTTP handlers, CLI commands, background jobs) to call the same validated business operations.

The action layer pattern is well-established in Go, with clear architectural benefits: handlers focus on request/response serialization while actions handle validation, database operations, and error mapping. The embedding pattern for method overrides is idiomatic Go and widely used in frameworks. Error handling should follow RFC 9457 (Problem Details) for APIs and context-aware rendering (SSE fragments for Datastar, error pages for traditional requests) for HTML.

**Primary recommendation:** Generate interface-based actions (List, Get, Create, Update, Delete) with concrete default implementations that handle validation, DB operations, and error mapping. Use Go's struct embedding for developer overrides. Implement a single forge.Error type carrying HTTP status, error codes, user messages, developer details, and wrapped errors. Leverage Huma's built-in RFC 9457 support for API errors and build SSE fragment responses for Datastar toast notifications.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| jackc/pgx/v5 | v5.x | PostgreSQL driver | Official recommendation, pgconn.PgError provides structured error codes (23505 unique, 23503 FK) |
| danielgtaylor/huma/v2 | v2.x | OpenAPI API framework | Already chosen for Phase 5, native RFC 9457 error support via ErrorDetailer interface |
| Bob (stephenafamo/bob) | - | Query builder | Already chosen in prior phases for type-safe queries |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| go-errors/errors | latest | Stack trace capture | Optional - only if stack traces in error logs are required |
| slog | stdlib | Structured logging | Standard library, use for logging errors with context |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom error struct | Third-party error packages | go-errors/errors adds stack traces but increases dependency footprint; not needed if panic recovery logs stacks |
| Manual panic recovery | Existing middleware (chi/middleware.Recoverer) | Chi's Recoverer works but using custom gives full control over logging format |

**Installation:**
```bash
# Core dependencies already in project from prior phases
go get github.com/jackc/pgx/v5
go get github.com/danielgtaylor/huma/v2

# Optional: if stack traces in wrapped errors are needed
go get github.com/go-errors/errors
```

## Architecture Patterns

### Recommended Project Structure
```
gen/
├── actions/           # Generated action interfaces and default implementations
│   ├── user.go       # type UserActions interface + DefaultUserActions struct
│   └── post.go       # type PostActions interface + DefaultPostActions struct
├── errors/           # Generated error definitions
│   └── errors.go     # forge.Error type, error codes, mapping functions
└── models/           # From Phase 2 - model types
    ├── user.go
    └── post.go

resources/
├── users/
│   ├── actions.go    # User embeds gen/actions.DefaultUserActions, overrides Create
│   └── handlers.go   # Calls actions, not DB directly
└── posts/
    └── actions.go    # Developer customization area
```

### Pattern 1: Generated Actions Interface
**What:** Each resource gets an interface defining CRUD operations with business-appropriate signatures
**When to use:** Always generated - provides contract for both default implementation and developer overrides
**Example:**
```go
// Source: Research synthesis from Three Dots Labs repository pattern
// gen/actions/user.go (GENERATED - do not edit)

package actions

import (
    "context"
    "myapp/gen/models"
    "myapp/gen/queries"
)

type UserActions interface {
    List(ctx context.Context, mods ...queries.UserMod) ([]models.User, int64, error)
    Get(ctx context.Context, id int64) (*models.User, error)
    Create(ctx context.Context, input models.UserCreate) (*models.User, error)
    Update(ctx context.Context, id int64, input models.UserUpdate) (*models.User, error)
    Delete(ctx context.Context, id int64) error
}
```

### Pattern 2: Default Actions Implementation
**What:** Concrete implementation handling validation, DB operations, error mapping
**When to use:** Generated for every resource - developer can use as-is or embed and override
**Example:**
```go
// Source: Service layer pattern from Alex Edwards "fat service" pattern
// gen/actions/user.go (GENERATED - do not edit)

type DefaultUserActions struct {
    DB *sql.DB
}

func (a *DefaultUserActions) Create(ctx context.Context, input models.UserCreate) (*models.User, error) {
    // 1. Validate input
    if err := validation.ValidateUserCreate(&input); err != nil {
        return nil, errors.NewValidationError(err)
    }

    // 2. Execute DB operation
    user, err := queries.InsertUser(ctx, a.DB, input)
    if err != nil {
        return nil, errors.MapDBError(err) // Maps 23505 → 409, etc.
    }

    // 3. Return result
    return user, nil
}
```

### Pattern 3: Developer Override via Embedding
**What:** Developer embeds DefaultActions and replaces specific methods
**When to use:** When custom business logic needed (e.g., send welcome email on user creation)
**Example:**
```go
// Source: Go struct embedding patterns from Go 101
// resources/users/actions.go (scaffolded once, developer edits)

package users

import (
    "context"
    "myapp/gen/actions"
    "myapp/gen/models"
)

type Actions struct {
    *actions.DefaultUserActions // Embed default implementation
    EmailService *email.Service
}

// Override Create to add welcome email
func (a *Actions) Create(ctx context.Context, input models.UserCreate) (*models.User, error) {
    // Call embedded default implementation
    user, err := a.DefaultUserActions.Create(ctx, input)
    if err != nil {
        return nil, err
    }

    // Add custom business logic
    a.EmailService.SendWelcome(user.Email)

    return user, nil
}

// List, Get, Update, Delete automatically promoted from DefaultUserActions
```

### Pattern 4: forge.Error - Unified Error Type
**What:** Single error type carrying HTTP status, error code, user message, developer detail, wrapped error
**When to use:** All action layer errors, mapped to appropriate format by handler layer
**Example:**
```go
// Source: RFC 9457 Problem Details + Go error handling best practices
// gen/errors/errors.go (GENERATED - do not edit)

package errors

import (
    "errors"
    "fmt"
)

type Error struct {
    Status  int    // HTTP status code (404, 409, 500, etc.)
    Code    string // Machine-readable code ("resource_not_found", "unique_violation")
    Message string // User-facing message
    Detail  string // Developer-facing detail (safe to expose in dev, hide in prod)
    Err     error  // Wrapped original error for errors.Is/As
}

func (e *Error) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *Error) Unwrap() error {
    return e.Err
}

func (e *Error) GetStatus() int {
    return e.Status
}

// Constructor functions
func NotFound(resource string, id any) *Error {
    return &Error{
        Status:  404,
        Code:    "resource_not_found",
        Message: fmt.Sprintf("%s not found", resource),
        Detail:  fmt.Sprintf("%s with id %v does not exist", resource, id),
    }
}

func UniqueViolation(field string) *Error {
    return &Error{
        Status:  409,
        Code:    "unique_violation",
        Message: fmt.Sprintf("%s already exists", field),
        Detail:  fmt.Sprintf("A record with this %s already exists", field),
    }
}
```

### Pattern 5: Database Error Mapping
**What:** Inspect pgx error codes and map to forge.Error with appropriate HTTP status
**When to use:** In MapDBError function called by all default actions after DB operations
**Example:**
```go
// Source: PostgreSQL error codes documentation + pgx error handling
// gen/errors/db_mapping.go (GENERATED - do not edit)

package errors

import (
    "errors"
    "github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes (from official docs)
const (
    ErrCodeUniqueViolation   = "23505"
    ErrCodeFKViolation       = "23503"
    ErrCodeNotNullViolation  = "23502"
    ErrCodeCheckViolation    = "23514"
)

func MapDBError(err error) error {
    if err == nil {
        return nil
    }

    // Check for pgx-specific error
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case ErrCodeUniqueViolation:
            return &Error{
                Status:  409,
                Code:    "unique_violation",
                Message: "A record with this value already exists",
                Detail:  pgErr.ConstraintName,
                Err:     err,
            }
        case ErrCodeFKViolation:
            return &Error{
                Status:  400,
                Code:    "foreign_key_violation",
                Message: "Referenced record does not exist",
                Detail:  pgErr.ConstraintName,
                Err:     err,
            }
        case ErrCodeNotNullViolation:
            return &Error{
                Status:  400,
                Code:    "not_null_violation",
                Message: "Required field is missing",
                Detail:  pgErr.ColumnName,
                Err:     err,
            }
        }
    }

    // Default to internal server error for unknown DB errors
    return &Error{
        Status:  500,
        Code:    "internal_error",
        Message: "An unexpected error occurred",
        Detail:  "Database error",
        Err:     err,
    }
}
```

### Pattern 6: Huma Error Transformer
**What:** Leverage Huma's ErrorDetailer interface to transform forge.Error into RFC 9457 format
**When to use:** API responses - Huma automatically calls this
**Example:**
```go
// Source: Huma documentation on response errors and ErrorDetailer
// internal/api/errors.go (written once, not generated)

package api

import (
    "myapp/gen/errors"
    "github.com/danielgtaylor/huma/v2"
)

// Implement ErrorDetailer so forge.Error works with Huma
func (e *errors.Error) ErrorDetail() *huma.ErrorDetail {
    return &huma.ErrorDetail{
        Message:  e.Message,
        Location: "",
        Value:    e.Detail,
    }
}

// Override huma.NewError to recognize forge.Error
func init() {
    originalNewError := huma.NewError
    huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
        // If first error is forge.Error, use its status and details
        if len(errs) > 0 {
            var forgeErr *errors.Error
            if errors.As(errs[0], &forgeErr) {
                model := &huma.ErrorModel{
                    Status: forgeErr.Status,
                    Title:  http.StatusText(forgeErr.Status),
                    Detail: forgeErr.Message,
                    Errors: []huma.ErrorDetail{*forgeErr.ErrorDetail()},
                }
                return model
            }
        }
        return originalNewError(status, message, errs...)
    }
}
```

### Pattern 7: Panic Recovery Middleware
**What:** Deferred recover() in same goroutine, log stack trace, return 500 without exposing internals
**When to use:** Wrap all HTTP handlers (both API and HTML)
**Example:**
```go
// Source: Panic recovery middleware patterns from chi and unrolled/recovery
// internal/middleware/recovery.go (written once, not generated)

package middleware

import (
    "log/slog"
    "net/http"
    "runtime/debug"
    "myapp/gen/errors"
)

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rvr := recover(); rvr != nil {
                // Log panic with stack trace
                slog.Error("panic recovered",
                    "error", rvr,
                    "stack", string(debug.Stack()),
                    "path", r.URL.Path,
                )

                // Return 500 without exposing panic details
                err := &errors.Error{
                    Status:  500,
                    Code:    "internal_error",
                    Message: "An unexpected error occurred",
                    Detail:  "", // Don't expose panic message
                }

                // Check if SSE context for Datastar fragment response
                if isSSEContext(r) {
                    writeSSEError(w, err)
                } else {
                    writeHTMLError(w, err)
                }
            }
        }()

        next.ServeHTTP(w, r)
    })
}

func isSSEContext(r *http.Request) bool {
    return r.Header.Get("Accept") == "text/event-stream"
}
```

### Pattern 8: Context-Aware Error Rendering
**What:** Render errors as SSE fragments for Datastar, error pages for traditional HTML, RFC 9457 for API
**When to use:** All error responses
**Example:**
```go
// Source: Datastar SSE patterns + Huma error responses
// internal/middleware/errors.go (written once, not generated)

package middleware

import (
    "fmt"
    "net/http"
    "myapp/gen/errors"
)

func writeSSEError(w http.ResponseWriter, err *errors.Error) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.WriteHeader(200) // SSE always returns 200, error is in data

    // Send fragment with toast notification
    fmt.Fprintf(w, "event: datastar-fragment\n")
    fmt.Fprintf(w, "data: <div id=\"toast\" class=\"error\">%s</div>\n\n", err.Message)
}

func writeHTMLError(w http.ResponseWriter, err *errors.Error) {
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(err.Status)

    // Render error page template
    errorPage := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <body>
            <h1>%d - %s</h1>
            <p>%s</p>
        </body>
        </html>
    `, err.Status, http.StatusText(err.Status), err.Message)

    w.Write([]byte(errorPage))
}

// API errors handled automatically by Huma via ErrorDetailer interface
```

### Anti-Patterns to Avoid
- **Business logic in handlers:** Handlers should only call actions and render responses, never execute validation or DB queries directly
- **Different validation for HTML vs API:** Both should call the same action methods with identical validation
- **Exposing stack traces in production API responses:** Panics should be logged server-side but return generic 500 messages to clients
- **Generic error types without status codes:** Always carry HTTP status in error type to avoid duplicated status logic in handlers
- **Using init() for registration:** Explicit registration at app startup is more testable and prevents import side effects

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Error stack traces | Custom stack capture logic | Runtime debug.Stack() or go-errors/errors | Stack trace capture is subtle (goroutines, defer ordering, performance), stdlib debug.Stack() works reliably |
| RFC 9457 Problem Details | Custom JSON error format | Huma's built-in ErrorModel | Huma natively supports RFC 9457, handles content negotiation, already in project |
| Database error code parsing | String matching on error messages | pgconn.PgError type assertion | Error messages are not stable, pgconn.PgError provides structured fields (Code, ConstraintName, ColumnName) |
| Panic recovery | Manual recover() in every handler | Middleware pattern with defer | Easy to forget in new handlers, middleware applies automatically to all routes |
| Context value storage | Global variables or request headers only | context.WithValue with typed keys | Context is designed for request-scoped data, prevents race conditions, works across middleware chain |

**Key insight:** Error handling infrastructure has many edge cases (goroutine boundaries, deferred execution order, production vs development modes, content negotiation). Leveraging stdlib patterns (context, errors.Is/As, debug.Stack) and Huma's built-in error support prevents reinventing solved problems.

## Common Pitfalls

### Pitfall 1: Embedding Method Not Calling Embedded Implementation
**What goes wrong:** Developer overrides a method but forgets to call the embedded default implementation, duplicating validation/DB logic
**Why it happens:** Embedding doesn't enforce calling "super" like inheritance in other languages
**How to avoid:**
- Document pattern clearly in scaffolded actions.go template with example
- Default pattern should call `a.DefaultActions.MethodName()` first, then add custom logic after
**Warning signs:** Validation errors not appearing, DB operations duplicated in overridden methods

### Pitfall 2: Panic Recovery Missing from Goroutines
**What goes wrong:** Panic in background goroutine spawned by handler crashes entire server
**Why it happens:** Middleware recover() only works in same goroutine as handler
**How to avoid:**
- Document in middleware comments: "Does NOT protect goroutines spawned by handlers"
- If handlers spawn goroutines, those must have their own defer/recover
**Warning signs:** Server crashes despite recovery middleware

### Pitfall 3: Error Context Lost in Wrapping
**What goes wrong:** Original database error details (constraint name, column) lost when wrapping
**Why it happens:** Wrapping without preserving original error via `Err` field
**How to avoid:** Always populate `Err` field in forge.Error with original error
**Warning signs:** errors.Is/errors.As returning false when checking for specific error types

### Pitfall 4: Different Error Handling for HTML vs API
**What goes wrong:** API returns RFC 9457 but HTML handlers use different error format, inconsistent behavior
**Why it happens:** Separate error handling logic in API vs HTML handlers
**How to avoid:**
- Single forge.Error type used everywhere
- Middleware/response layer decides rendering format based on Accept header
- Actions return forge.Error, handlers never construct HTTP responses from errors directly
**Warning signs:** Same error shows different messages in API vs HTML interface

### Pitfall 5: Exposing Database Constraint Names to Users
**What goes wrong:** Error message shows "users_email_key" constraint name to end user
**Why it happens:** Using pgErr.ConstraintName in Message field instead of Detail field
**How to avoid:**
- Message field: generic user-friendly message ("Email already exists")
- Detail field: technical details including constraint name (for developers)
- Production API responses should omit Detail field
**Warning signs:** User-facing errors contain SQL table/column names

### Pitfall 6: Not Handling errors.Is in Middleware
**What goes wrong:** Custom errors wrapped in middleware checks fail because handler uses errors.Is instead of type assertion
**Why it happens:** Middleware code uses `err == specificErr` instead of `errors.Is(err, specificErr)`
**How to avoid:** Always use errors.Is for sentinel errors, errors.As for type checking
**Warning signs:** Error handling inconsistent depending on how many times error was wrapped

### Pitfall 7: SSE Error Responses with Non-200 Status
**What goes wrong:** SSE connection closes immediately when error has non-200 status code
**Why it happens:** SSE protocol requires 200 status, error information goes in event data
**How to avoid:**
- Check if request expects SSE (Accept: text/event-stream header)
- For SSE: always return 200, embed error in event data
- For non-SSE: return appropriate HTTP status code
**Warning signs:** Datastar toast notifications not appearing, SSE connections dropping on errors

## Code Examples

Verified patterns from official sources:

### Example 1: Using errors.As for pgx Error Checking
```go
// Source: pgx documentation and PostgreSQL error codes appendix
import (
    "errors"
    "github.com/jackc/pgx/v5/pgconn"
)

func checkDBError(err error) {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        // Access structured fields
        code := pgErr.Code                 // "23505"
        constraint := pgErr.ConstraintName // "users_email_key"
        column := pgErr.ColumnName         // "email"
        detail := pgErr.Detail             // Human readable detail
    }
}
```

### Example 2: Middleware Chain with Recovery
```go
// Source: Go middleware patterns and justinas/alice
import (
    "net/http"
    "github.com/justinas/alice"
)

func setupMiddleware() http.Handler {
    chain := alice.New(
        middleware.Recovery,        // Catch panics first
        middleware.Logging,         // Then log requests
        middleware.Authentication,  // Then authenticate
    )

    return chain.Then(mux)
}
```

### Example 3: Context Request-Scoped Data
```go
// Source: Go context patterns for HTTP handlers
import "context"

type contextKey string

const userKey contextKey = "user"

// Middleware stores user in context
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := authenticateUser(r)
        ctx := context.WithValue(r.Context(), userKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Action retrieves user from context
func (a *DefaultUserActions) List(ctx context.Context, mods ...queries.UserMod) ([]models.User, int64, error) {
    user, ok := ctx.Value(userKey).(*User)
    if !ok {
        return nil, 0, errors.Unauthorized("authentication required")
    }
    // Use user for authorization checks...
}
```

### Example 4: Custom Validator Interface
```go
// Source: Go validator patterns and interface design
package actions

type CreateValidator interface {
    ValidateCreate(ctx context.Context, input any) error
}

type UpdateValidator interface {
    ValidateUpdate(ctx context.Context, id int64, input any) error
}

// Generated default action checks if it implements validators
func (a *DefaultUserActions) Create(ctx context.Context, input models.UserCreate) (*models.User, error) {
    // Check if custom validator is implemented
    if validator, ok := any(a).(CreateValidator); ok {
        if err := validator.ValidateCreate(ctx, input); err != nil {
            return nil, err
        }
    } else {
        // Fall back to generated validation
        if err := validation.ValidateUserCreate(&input); err != nil {
            return nil, errors.NewValidationError(err)
        }
    }

    // Proceed with DB operation...
}

// Developer can add custom validation
type Actions struct {
    *actions.DefaultUserActions
}

func (a *Actions) ValidateCreate(ctx context.Context, input any) error {
    userInput := input.(models.UserCreate)

    // Custom business rule: admins only created on weekdays
    if userInput.Role == "admin" && time.Now().Weekday() == time.Saturday {
        return errors.NewValidationError("Admin accounts can only be created on weekdays")
    }

    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| pkg/errors for wrapping | stdlib errors.Is/As + fmt.Errorf %w | Go 1.13 (2019) | Reduced dependency on third-party packages, pkg/errors still useful for stack traces |
| RFC 7807 Problem Details | RFC 9457 Problem Details | 2023 | Minor update, mostly compatible, Huma v2 uses 9457 |
| Global error handlers | Context-aware error rendering | 2020+ | Enables different error formats (HTML/JSON/SSE) based on request context |
| Repository pattern with DB handle | Actions pattern with interface | 2024+ | Actions are broader than repositories (can include external API calls, job queueing) |
| init() for registration | Explicit startup registration | 2022+ | Testability and explicitness valued over convenience |

**Deprecated/outdated:**
- **pkg/errors.Wrap/Cause:** Use stdlib fmt.Errorf with %w and errors.As instead (Go 1.13+)
- **SQLBoiler:** In maintenance mode as of 2026, sqlc is preferred (though Forge uses Bob)
- **Global http.DefaultServeMux:** Modern Go uses explicit router instances (chi, Huma's router)
- **Panic for application errors:** Panics should only be for programmer errors, not validation/business logic failures

## Open Questions

1. **Stack Trace Capture Strategy**
   - What we know: Can use debug.Stack() or go-errors/errors
   - What's unclear: Whether stack traces should be captured on every error (performance cost) or only on panics
   - Recommendation: Start with stack traces only on panics (via recovery middleware), add to forge.Error optionally if needed

2. **Validation Error Details Format**
   - What we know: Need to return field-level validation errors (e.g., "email: invalid format")
   - What's unclear: Best format for multiple field errors - array of ErrorDetail vs single message
   - Recommendation: Use Huma's ErrorDetail array pattern for consistency with RFC 9457

3. **Transaction Handling in Actions**
   - What we know: Phase 3 built forge.Transaction wrapper
   - What's unclear: Should action methods accept *sql.Tx or always use DB and create tx internally?
   - Recommendation: Actions take DB connection (interface), can optionally accept ongoing transaction from context for multi-action operations

4. **Developer Override Discovery**
   - What we know: Embedding enables override, type assertions can check for interfaces
   - What's unclear: How to make it obvious which methods are "safe to override" vs "don't override this"
   - Recommendation: Document in generated comments which methods are override-safe, provide examples in scaffolded actions.go

5. **Error Code Standardization**
   - What we know: Need machine-readable error codes ("unique_violation", "not_found")
   - What's unclear: Should codes be generated per-resource or global constants?
   - Recommendation: Global constants in gen/errors/codes.go for consistency, developer can add custom codes in resources/

## Sources

### Primary (HIGH confidence)
- [PostgreSQL Error Codes Appendix](https://www.postgresql.org/docs/current/errcodes-appendix.html) - Class 23 integrity constraint violations
- [Huma Response Errors Documentation](https://huma.rocks/features/response-errors/) - RFC 9457 implementation, ErrorDetailer interface
- [Huma Error.go Source](https://github.com/danielgtaylor/huma/blob/main/error.go) - ErrorModel structure and NewError pattern

### Secondary (MEDIUM confidence)
- [The Fat Service Pattern - Alex Edwards](https://www.alexedwards.net/blog/the-fat-service-pattern) - Service layer architecture in Go
- [Repository Pattern in Go - Three Dots Labs](https://threedots.tech/post/repository-pattern-in-go/) - Separating business logic from data access
- [Type Embedding - Go 101](https://go101.org/article/type-embedding.html) - Official embedding behavior and method promotion
- [Panic Recovery Middleware - Abu Ashraf Masnun](https://medium.com/@masnun/panic-recovery-middleware-for-go-http-handlers-51147c1941f9) - Recovery middleware pattern
- [Making and Using HTTP Middleware - Alex Edwards](https://www.alexedwards.net/blog/making-and-using-middleware) - Middleware chaining patterns
- [Middleware Patterns in Go - Dr. Stearns](https://drstearns.github.io/tutorials/gomiddleware/) - Context-aware middleware
- [Error Handling in Go HTTP Apps - Joe Shaw](https://www.joeshaw.org/error-handling-in-go-http-applications/) - Error handling patterns
- [Crafting Custom Errors - Leapcell](https://leapcell.io/blog/crafting-custom-errors-and-http-status-codes-in-go-apis) - Error mapping to HTTP status codes
- [PGX Error Handling - Till It's Done](https://tillitsdone.com/blogs/pgx-error-handling-in-go-apps/) - pgconn.PgError usage
- [How to Create Custom Error Types with Stack Traces - OneUpTime (2026)](https://oneuptime.com/blog/post/2026-01-30-how-to-create-custom-error-types-with-stack-traces-in-go/view) - Modern error patterns
- [How to Implement Middleware Chains - OneUpTime (2026)](https://oneuptime.com/blog/post/2026-01-30-go-middleware-chains-http/view) - Current middleware best practices
- [How to Implement Context Propagation - OneUpTime (2026)](https://oneuptime.com/blog/post/2026-02-01-go-context-propagation-microservices/view) - Context patterns

### Tertiary (LOW confidence)
- [Datastar Documentation](https://data-star.dev/guide/getting_started) - SSE event structure (specific error handling patterns not documented, needs further investigation)
- Various GitHub issues and discussions on error handling patterns (cross-referenced for patterns)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - pgx, Huma, Bob all confirmed from prior phases
- Architecture: HIGH - Service layer, error wrapping, middleware chaining well-documented
- Database error mapping: HIGH - PostgreSQL error codes official, pgx patterns verified
- Panic recovery: HIGH - Multiple sources confirm defer/recover pattern, goroutine limitation documented
- Pitfalls: MEDIUM - Synthesized from multiple sources, some based on general Go experience
- SSE error handling: LOW - Datastar docs don't detail error fragment patterns, needs validation

**Research date:** 2026-02-16
**Valid until:** ~30 days (Go patterns stable, but verify Huma v2 hasn't changed error handling in updates)
