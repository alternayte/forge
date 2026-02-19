package api

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"

	"github.com/alternayte/forge/forge/auth"
)

// HTMLServerConfig holds the dependencies required to set up HTML (browser-facing)
// routes. It separates session management from route registration so that the
// forge framework layer never needs to import generated application code.
type HTMLServerConfig struct {
	// SessionManager is the SCS session manager used for LoadAndSave middleware
	// and RequireSession auth enforcement.
	SessionManager *scs.SessionManager

	// RegisterRoutes is an optional function that registers resource HTML routes
	// on a protected (session-authenticated) chi.Router group. Pass nil if you
	// only need the session middleware wired without any routes.
	RegisterRoutes func(chi.Router)
}

// SetupHTML wires session middleware and HTML route groups onto a Chi router.
//
// Route group structure:
//
//  1. Session LoadAndSave middleware wraps ALL routes (including auth routes) so
//     that the OAuth callback handler can commit sessions after login. This is
//     critical — if LoadAndSave only wraps the protected group, the OAuth callback
//     (which lives in the public group) cannot write session data.
//     (Pitfall 9: Session LoadAndSave must wrap ALL routes including auth routes)
//
//  2. Public group (no auth required): intended for OAuth callback routes,
//     /auth/login, /auth/logout, and any other unauthenticated pages.
//     These must be registered BEFORE the protected RequireSession group to avoid
//     infinite redirect loops.
//     (Pitfall 4: OAuth callback routes must be outside RequireSession group)
//
//  3. Protected group (RequireSession enforced): resource HTML routes live here.
//     cfg.RegisterRoutes is called on this group so every resource route
//     automatically inherits session auth without each handler checking itself.
//
// Example wiring in a generated project's main.go:
//
//	err := apiserver.SetupHTML(router, api.HTMLServerConfig{
//	    SessionManager: sm,
//	    RegisterRoutes: genhtml.RegisterAllHTMLRoutes,
//	})
func SetupHTML(router chi.Router, cfg HTMLServerConfig) error {
	// Session LoadAndSave must wrap ALL routes including auth routes so the
	// OAuth callback can commit sessions (Pitfall 9).
	router.Use(cfg.SessionManager.LoadAndSave)

	// Public group — no session authentication required.
	// Register OAuth callbacks, login/logout handlers here to prevent the
	// RequireSession middleware from redirecting these routes to /auth/login.
	router.Group(func(r chi.Router) {
		// Public auth routes are registered by the application via RegisterOAuthRoutes
		// and password login/logout handlers. This group is intentionally left empty
		// by SetupHTML so the generated application can inject its own auth handlers.
	})

	// Protected group — RequireSession redirects unauthenticated users to /auth/login.
	router.Group(func(r chi.Router) {
		r.Use(auth.RequireSession(cfg.SessionManager))
		if cfg.RegisterRoutes != nil {
			cfg.RegisterRoutes(r)
		}
	})

	return nil
}
