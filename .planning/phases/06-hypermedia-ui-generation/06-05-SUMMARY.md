---
phase: 06-hypermedia-ui-generation
plan: 05
subsystem: auth
tags: [oauth2, goth, google, github, sessions, pgxstore, scs, gorilla-sessions, atlas, hcl]

# Dependency graph
requires:
  - phase: 06-02
    provides: session.go with LoginUser, LogoutUser, SessionKeyUserID and SCS session manager setup
provides:
  - OAuth2 provider setup via Goth (Google + GitHub) in internal/auth/oauth.go
  - SetupOAuth configures goth providers and gothic cookie store for OAuth temp state
  - HandleOAuthCallback completes OAuth handshake and stores user in SCS session
  - RegisterOAuthRoutes mounts public auth routes outside RequireSession middleware
  - HandleLogin/HandleLoginSubmit for email/password authentication
  - HandleLogout for session destruction and redirect
  - UserFinder and PasswordAuthenticator callback types for app integration
  - Sessions table (token PK, data bytea, expiry timestamptz) in Atlas HCL schema template
affects: [06-06, 06-08, 06-09, html-routes, server-assembly]

# Tech tracking
tech-stack:
  added:
    - github.com/markbates/goth v1.82.0 (OAuth2 provider framework)
    - golang.org/x/oauth2 v0.27.0 (transitive, OAuth2 token exchange)
    - github.com/gorilla/sessions v1.1.1 (transitive, gothic cookie store)
    - github.com/gorilla/securecookie v1.1.1 (transitive, signed cookie support)
  patterns:
    - OAuth temp state in signed cookie (gothic.Store), completed sessions in PostgreSQL (SCS+pgxstore)
    - Auth routes registered outside RequireSession middleware group to prevent infinite redirect loops
    - UserFinder/PasswordAuthenticator callback types let generated app provide user lookup without circular imports
    - Sessions table always included in Atlas schema (infrastructure, not per-resource)

key-files:
  created:
    - internal/auth/oauth.go
  modified:
    - internal/generator/templates/atlas_schema.hcl.tmpl
    - go.mod
    - go.sum

key-decisions:
  - "OAuth temp state (CSRF token) stored in gorilla sessions signed cookie via gothic.Store; completed auth sessions stored in PostgreSQL via SCS+pgxstore (AUTH-03 compliant)"
  - "RegisterOAuthRoutes places all auth routes in a plain group (no RequireSession middleware) — critical to prevent infinite redirect loops for unauthenticated users"
  - "UserFinder and PasswordAuthenticator are function types (not interfaces) passed as callbacks, so generated app code can provide user lookup without the forge tool importing gen/ packages"
  - "Sessions table appended as static block after resource range loop in Atlas template — infrastructure table not dependent on any user-defined resource"

patterns-established:
  - "Pattern: Goth OAuth callback -> UserFinder -> LoginUser -> Redirect / flow for OAuth2 authentication"
  - "Pattern: HandleLoginSubmit re-renders login page with error on failure (no redirect) for email/password auth"

requirements-completed: []

# Metrics
duration: 2min
completed: 2026-02-17
---

# Phase 6 Plan 05: OAuth2 Providers and pgxstore Sessions Table Summary

**Google+GitHub OAuth2 via Goth with signed-cookie temp state and PostgreSQL sessions table generated in Atlas HCL schema**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-17T21:09:44Z
- **Completed:** 2026-02-17T21:11:44Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- SetupOAuth registers Google and GitHub OAuth2 providers with Goth and sets gothic.Store to a gorilla sessions signed cookie store for OAuth temp state
- Full OAuth callback flow: gothic.CompleteUserAuth -> UserFinder -> LoginUser -> redirect to "/"
- Email/password login (HandleLogin, HandleLoginSubmit) and logout (HandleLogout) handlers with session integration
- Sessions table (token text PK, data bytea, expiry timestamptz, sessions_expiry_idx) appended as static block in Atlas HCL schema template

## Task Commits

Each task was committed atomically:

1. **Task 1: Create OAuth2 provider setup and callback handler** - `66a37e6` (feat)
2. **Task 2: Add sessions table to Atlas HCL schema generation** - `414430d` (feat)

**Plan metadata:** (docs commit to follow)

## Files Created/Modified

- `internal/auth/oauth.go` - OAuthConfig, SetupOAuth, HandleOAuthCallback, RegisterOAuthRoutes, HandleLogin, HandleLoginSubmit, HandleLogout, UserFinder, PasswordAuthenticator
- `internal/generator/templates/atlas_schema.hcl.tmpl` - Appended static sessions table block after resource range loop
- `go.mod` - Added github.com/markbates/goth v1.82.0 and transitive deps
- `go.sum` - Updated checksums for goth and gorilla transitive deps

## Decisions Made

- OAuth temp state (CSRF token) stays in gorilla sessions signed cookie via gothic.Store. Completed authentication sessions go to PostgreSQL via SCS+pgxstore (AUTH-03 compliant). Clear separation of concerns: short-lived handshake cookie vs. long-lived DB-backed session.
- All auth routes registered outside RequireSession middleware group — essential to prevent infinite redirect loops for unauthenticated users visiting /auth/login or /auth/{provider}.
- UserFinder/PasswordAuthenticator as function types (not interfaces) — generated apps pass closures, forge tool does not need to import gen/ packages.
- Sessions table appended as a static block (not conditional on resource presence) — it is infrastructure required by SCS pgxstore, not a user-defined schema entity.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed non-constant format string in fmt.Fprintf calls**
- **Found during:** Task 1 (verification via go vet)
- **Issue:** `fmt.Fprintf(w, loginPageHTML(""))` and `fmt.Fprintf(w, loginPageHTML("Invalid..."))` — go vet flagged non-constant format string
- **Fix:** Changed both to `fmt.Fprint(w, ...)` since no format string substitution is needed
- **Files modified:** internal/auth/oauth.go
- **Verification:** `go vet ./internal/auth/` reports no issues
- **Committed in:** 66a37e6 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Minor vet correctness fix. No scope changes.

## Issues Encountered

None — goth dependency installed cleanly, gorilla/sessions pulled as transitive dependency via `go mod tidy`.

## User Setup Required

None - no external service configuration required at generation time. OAuth client credentials (ClientID, ClientSecret, SessionSecret) are runtime configuration provided by the application using the generated framework.

## Next Phase Readiness

- OAuth2 + email/password auth handlers ready for registration in the HTML route layer (06-06)
- Sessions table will be present in all generated Atlas schemas, so pgxstore can persist sessions after `forge migrate`
- RegisterOAuthRoutes requires a chi.Router, SessionManager, UserFinder, and PasswordAuthenticator — server assembly phase (06-08/09) will wire these together

---
*Phase: 06-hypermedia-ui-generation*
*Completed: 2026-02-17*
