---
phase: 05-rest-api-generation
plan: 02
subsystem: auth
tags: [bearer-token, api-key, huma, cors, rate-limiting, go-limiter, rs-cors, uuid]

requires:
  - phase: 05-rest-api-generation/05-01
    provides: "Huma v2 dependency and API generation templates that drove need for auth/CORS/rate-limit middleware"

provides:
  - "TokenStore interface for database-backed bearer token lookup, creation, and revocation"
  - "APIKeyStore interface for database-backed API key management with forg_live_/forg_test_ prefixes"
  - "AuthMiddleware (Huma middleware) validating Bearer tokens and forg_ API keys with constant-time comparison"
  - "CORSMiddleware wrapping rs/cors with forge.toml configuration and wildcard+credentials safety guard"
  - "RateLimitMiddleware wrapping go-limiter memorystore with per-IP token bucket"
  - "APIConfig struct with RateLimitConfig/CORSConfig/TierConfig for forge.toml [api] section"

affects: [05-03-server-assembly, 06-html-generation, 07-public-access-annotations]

tech-stack:
  added:
    - "github.com/google/uuid v1.6.0 — UUID types for Token/APIKey store interfaces"
    - "github.com/danielgtaylor/huma/v2 v2.36.0 — Huma Context interface for middleware"
    - "github.com/rs/cors v1.11.1 — CORS header handling"
    - "github.com/sethvargo/go-limiter v1.1.0 — Token bucket rate limiting with httplimit/memorystore"
  patterns:
    - "Auth stores as interfaces (not concrete types) — database layer implements in later phase"
    - "Huma middleware returns updated context via huma.WithValue (not mutates in-place)"
    - "crypto/subtle.ConstantTimeCompare for all token/key comparisons (never ==)"
    - "Noop pass-through pattern for disabled middleware (Enabled=false returns identity handler)"
    - "API key prefix validation before database lookup (fail fast on malformed keys)"

key-files:
  created:
    - internal/auth/token.go
    - internal/auth/apikey.go
    - internal/api/middleware/auth.go
    - internal/api/middleware/cors.go
    - internal/api/middleware/ratelimit.go
    - internal/config/api.go

key-decisions:
  - "huma.API passed to AuthMiddleware constructor so WriteErr can produce structured 401 responses"
  - "validateBearerToken and validateAPIKey return updated huma.Context (not mutate) to thread context through middleware chain"
  - "IsAPIKey helper added to auth package as convenience wrapper for middleware use"
  - "Phase 5 rate limiting uses Default tier for all requests — tiered enforcement (auth vs API key vs anonymous) deferred to Plan 03 server assembly when context is available"
  - "CORSMiddleware logs warning and disables credentials when wildcard origin is combined with AllowCredentials (CORS spec violation guard)"

patterns-established:
  - "Auth store interfaces: define interface + types in internal/auth/, implement in database layer"
  - "Context keys: typed contextKey string (not raw string) to avoid collisions across packages"
  - "Middleware constructor pattern: NewXMiddleware(deps...) returns struct with Handle method"

duration: 3min
completed: 2026-02-17
---

# Phase 5 Plan 2: Authentication Infrastructure and Middleware Summary

**Database-backed bearer token + API key auth stores with Huma middleware, rs/cors CORS handling, and go-limiter per-IP rate limiting — all wired from forge.toml configuration**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-17T18:58:08Z
- **Completed:** 2026-02-17T19:01:xx Z
- **Tasks:** 2
- **Files modified:** 6 created, 2 modified (go.mod, go.sum)

## Accomplishments

- TokenStore and APIKeyStore interfaces with full CRUD operations and cryptographically-safe key generation helpers
- AuthMiddleware using crypto/subtle.ConstantTimeCompare for both bearer token and API key validation, setting typed context keys for downstream handlers
- CORSMiddleware with wildcard+credentials safety guard to prevent CORS spec violations
- RateLimitMiddleware with per-IP token bucket using go-limiter memorystore (extensible for tiered enforcement)
- APIConfig struct with DefaultAPIConfig() covering three rate limit tiers and full CORS options

## Task Commits

1. **Task 1: Auth store interfaces and authentication middleware** - `59a1e0d` (feat)
2. **Task 2: CORS and rate limiting middleware with forge.toml configuration** - `ca398f5` (feat)

## Files Created/Modified

- `internal/auth/token.go` - Token struct, TokenStore interface, GenerateToken() helper (32-byte hex)
- `internal/auth/apikey.go` - APIKey struct, APIKeyStore interface, GenerateAPIKey(), ValidateKeyPrefix(), IsAPIKey() helpers with forg_live_/forg_test_ prefixes
- `internal/api/middleware/auth.go` - AuthMiddleware with Handle(), validateBearerToken(), validateAPIKey(); typed contextKey constants
- `internal/api/middleware/cors.go` - CORSMiddleware wrapping rs/cors with wildcard+credentials guard
- `internal/api/middleware/ratelimit.go` - RateLimitMiddleware wrapping go-limiter with memorystore and IPKeyFunc
- `internal/config/api.go` - APIConfig, RateLimitConfig, TierConfig, CORSConfig, DefaultAPIConfig()

## Decisions Made

- **AuthMiddleware takes huma.API in constructor:** WriteErr requires an huma.API argument to produce properly-structured JSON error bodies. Storing it at construction time is cleaner than passing it to every Handle call.
- **validateBearerToken/validateAPIKey return updated huma.Context:** huma.WithValue returns a new context (functional style); the updated context must be passed to `next()` in Handle, so the helpers return it.
- **IsAPIKey helper added to auth package:** The middleware needs to distinguish API keys from bearer tokens in a switch. A named helper in the auth package is cleaner than duplicating the prefix-check logic.
- **Phase 5 rate limiting uses Default tier for all requests:** Tiered enforcement (checking auth context for authenticated vs API key vs anonymous) requires knowing the auth result, which means rate limiting must run after auth in Plan 03's server assembly. The config stores all three tiers; the middleware uses Default for now.
- **Wildcard + credentials guard in CORSMiddleware:** Combining AllowedOrigins: ["*"] with AllowCredentials: true is a CORS spec violation that browsers silently reject. The guard detects this, logs a warning, and disables credentials rather than panicking, keeping the API functional.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added IsAPIKey helper to auth package**
- **Found during:** Task 1 (auth middleware implementation)
- **Issue:** The middleware needed to check if the Authorization header value is an API key (not a bearer token) in a switch statement. Importing auth.ValidateKeyPrefix and discarding the prefix added noise. A named boolean helper is more readable.
- **Fix:** Added `IsAPIKey(key string) bool` as a one-liner convenience wrapper around ValidateKeyPrefix in internal/auth/apikey.go
- **Files modified:** internal/auth/apikey.go
- **Verification:** go build ./internal/auth/ passes; grep confirms both IsAPIKey and ValidateKeyPrefix are exported
- **Committed in:** 59a1e0d (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 2 — missing helper for correctness)
**Impact on plan:** Minimal scope addition; the helper is a three-line function that prevents duplicating prefix-check logic in the middleware.

## Issues Encountered

None — plan executed cleanly. Dependencies installed correctly on first attempt.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Auth middleware ready to wire into Huma API via `api.UseMiddleware(authMiddleware.Handle)` in Plan 03 server assembly
- CORS and rate limit middleware return `func(http.Handler) http.Handler` — compatible with standard Go HTTP middleware chains
- APIConfig struct ready for integration into the root Config struct in internal/config/config.go when the server is assembled
- TokenStore and APIKeyStore interfaces ready for database implementation in Phase 6 or 7

---
*Phase: 05-rest-api-generation*
*Completed: 2026-02-17*
