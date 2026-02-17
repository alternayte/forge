---
phase: 06-hypermedia-ui-generation
plan: "02"
subsystem: auth
tags: [scs, pgxstore, bcrypt, session, postgres, middleware]

requires:
  - phase: 05-rest-api-generation
    provides: "internal/auth package with TokenStore/APIKeyStore interfaces and contextKey pattern"

provides:
  - "SCS session manager backed by PostgreSQL via pgxstore (NewSessionManager)"
  - "bcrypt password hashing/checking utilities (HashPassword, CheckPassword, BcryptCost=12)"
  - "HTML-specific auth middleware that redirects to /auth/login (RequireSession)"
  - "Session helpers: LoginUser (with RenewToken fixation prevention), LogoutUser, GetSessionUserID, GetSessionUserEmail"
  - "SessionConfig struct added to config.Config for runtime session configuration"

affects:
  - 06-03-html-routes
  - 06-08-auth-handlers
  - any plan that wires HTML routes or instantiates the server

tech-stack:
  added:
    - github.com/alexedwards/scs/v2 v2.9.0 (session management)
    - github.com/alexedwards/scs/pgxstore (PostgreSQL session store)
    - golang.org/x/crypto/bcrypt (password hashing)
    - github.com/jackc/pgx/v5 v5.7.1 (transitive, pgxpool.Pool)
  patterns:
    - "HTML middleware redirects (302) vs API middleware returns 401 JSON - intentional divergence"
    - "Session fixation prevention via sm.RenewToken() before writing credentials on login"
    - "Bcrypt 72-byte length guard (bcrypt silently truncates, causing hash collisions)"
    - "isDev bool controls Cookie.Secure - false for local HTTP, true for production HTTPS"

key-files:
  created:
    - internal/auth/session.go
    - internal/auth/password.go
    - internal/auth/html_middleware.go
  modified:
    - internal/config/config.go
    - go.mod
    - go.sum

key-decisions:
  - "SCS session middleware returns sm.LoadAndSave directly (Chi-compatible, no wrapper needed)"
  - "BcryptCost=12 (not default 10) for 2026 hardware resistance to offline brute-force"
  - "72-byte guard on HashPassword prevents silent bcrypt truncation causing hash collisions"
  - "RequireSession redirects to /auth/login with 302 Found (not 401) — HTML UX pattern"
  - "LoginUser calls sm.RenewToken before Put() to prevent session fixation"
  - "SessionConfig.Secure separate from isDev bool on NewSessionManager — config drives prod, isDev drives construction"

patterns-established:
  - "HTML auth middleware pattern: session check + redirect vs API auth: token check + 401 JSON"
  - "Session key constants (SessionKeyUserID, SessionKeyUserEmail) live in session.go for single source of truth"

requirements-completed: []

duration: 2min
completed: 2026-02-17
---

# Phase 6 Plan 02: Session Auth Infrastructure Summary

**PostgreSQL-backed SCS session manager with bcrypt password utilities and HTML redirect middleware separating browser UX from API auth**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-17T21:03:32Z
- **Completed:** 2026-02-17T21:05:32Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- SCS session manager with pgxstore PostgreSQL backend, httpOnly cookie, SameSite=Lax, 24h lifetime
- bcrypt password hashing with cost 12 and 72-byte length guard against silent truncation
- HTML auth middleware (RequireSession) that redirects to /auth/login on unauthenticated requests
- LoginUser with session fixation prevention via sm.RenewToken before credential storage
- SessionConfig struct added to config.Config for runtime secret/secure/lifetime configuration

## Task Commits

Each task was committed atomically:

1. **Task 1: Create SCS session manager and bcrypt password utilities** - `cbae0ee` (feat)
2. **Task 2: Create HTML auth middleware and add SessionConfig** - `10c8c72` (feat)

## Files Created/Modified

- `internal/auth/session.go` - NewSessionManager (pgxstore), SessionMiddleware, SessionKey constants
- `internal/auth/password.go` - HashPassword/CheckPassword with bcrypt cost 12 and 72-byte guard
- `internal/auth/html_middleware.go` - RequireSession, LoginUser, LogoutUser, session getter helpers
- `internal/config/config.go` - Added SessionConfig struct (Secret, Secure, Lifetime fields)
- `go.mod` / `go.sum` - Added scs/v2, scs/pgxstore, pgx/v5, golang.org/x/crypto

## Decisions Made

- Used SCS v2.9.0 with pgxstore to satisfy AUTH-03 (no Redis dependency, PostgreSQL-backed sessions)
- BcryptCost=12 over default 10 — appropriate for 2026 hardware to slow offline brute-force
- 72-byte password length guard: bcrypt silently truncates inputs, two different passwords >72 bytes could hash identically
- RequireSession redirects (302) rather than returning 401 — HTML pages need browser redirect, not JSON error
- LoginUser calls sm.RenewToken() before storing user data — prevents session fixation attacks where attacker captures pre-login session token

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all dependencies installed cleanly, all builds passed on first attempt.

## User Setup Required

None - no external service configuration required (pgxstore uses the existing pgxpool.Pool from the server's database connection).

## Next Phase Readiness

- Session infrastructure is ready for Plan 08 (HTML auth handlers: login/logout routes)
- RequireSession middleware ready to apply to protected HTML route groups in Plan 03
- Config SessionConfig ready for forge.toml template update in later plan

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*
