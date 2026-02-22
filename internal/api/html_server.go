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
	// on a chi.Router group. When RequireAuth is true, the group enforces
	// session authentication. Pass nil if you only need the session middleware
	// wired without any routes.
	RegisterRoutes func(chi.Router)

	// RequireAuth controls whether HTML routes are protected by RequireSession
	// middleware. When false (default), routes are publicly accessible — suitable
	// for freshly generated projects before auth is configured. Set to true once
	// OAuth or password login routes are wired up.
	RequireAuth bool

	// RegisterPublicRoutes is an optional function that registers routes in the
	// public (unauthenticated) group. Auth routes (login, logout, OAuth callbacks)
	// are registered here to avoid RequireSession redirect loops.
	RegisterPublicRoutes func(chi.Router)
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
	// Wrap everything in a Group so that router.Use() applies to a fresh
	// sub-router instead of the top-level mux. Chi panics if Use() is called
	// after routes have already been registered (e.g., by SetupAPI).
	router.Group(func(r chi.Router) {
		// Session LoadAndSave must wrap ALL routes including auth routes so the
		// OAuth callback can commit sessions (Pitfall 9).
		r.Use(cfg.SessionManager.LoadAndSave)

		// Public group — no session authentication required.
		// Register OAuth callbacks, login/logout handlers here to prevent the
		// RequireSession middleware from redirecting these routes to /auth/login.
		if cfg.RegisterPublicRoutes != nil {
			r.Group(func(pub chi.Router) {
				cfg.RegisterPublicRoutes(pub)
			})
		}

		// Route group for resource HTML routes.
		// When RequireAuth is true, RequireSession redirects unauthenticated
		// users to /auth/login. When false, routes are publicly accessible.
		r.Group(func(rr chi.Router) {
			if cfg.RequireAuth {
				rr.Use(auth.RequireSession(cfg.SessionManager))
			}
			if cfg.RegisterRoutes != nil {
				cfg.RegisterRoutes(rr)
			}
		})
	})

	return nil
}
