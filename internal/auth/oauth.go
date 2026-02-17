package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

// OAuthConfig holds OAuth2 provider credentials and settings.
//
// OAuth temp state (the 30-60 second CSRF token used during the OAuth handshake)
// is stored in a signed cookie via gothic.Store. Completed authentication
// sessions are stored in PostgreSQL via SCS + pgxstore (AUTH-03 compliant).
type OAuthConfig struct {
	Google struct {
		ClientID     string
		ClientSecret string
	}
	GitHub struct {
		ClientID     string
		ClientSecret string
	}
	// CallbackBaseURL is the base URL for OAuth callback routes,
	// e.g. "http://localhost:8080".
	CallbackBaseURL string
	// SessionSecret is used to sign the gothic cookie store. Must be a
	// sufficiently random secret (32+ bytes recommended).
	SessionSecret string
}

// UserFinder maps an OAuth user returned by Goth to an internal user ID.
// The implementation is provided by the generated application and is
// responsible for creating the user record if one does not already exist.
type UserFinder func(ctx context.Context, gothUser goth.User) (userID string, err error)

// PasswordAuthenticator verifies an email/password credential pair and returns
// the internal user ID on success. The implementation is provided by the
// generated application.
type PasswordAuthenticator func(ctx context.Context, email, password string) (userID string, err error)

// SetupOAuth registers Google and GitHub OAuth2 providers with Goth and
// configures gothic.Store with a signed cookie store for OAuth temp state.
//
// OAuth flow:
//   - gothic.Store holds the short-lived CSRF/state token (30-60 s) in a
//     signed, HttpOnly cookie. This is only required during the OAuth
//     handshake; it is discarded after the callback completes.
//   - Completed auth sessions are stored in PostgreSQL via SCS + pgxstore
//     (AUTH-03 compliant). The SCS session is loaded by SessionMiddleware.
func SetupOAuth(cfg OAuthConfig) {
	goth.UseProviders(
		google.New(
			cfg.Google.ClientID,
			cfg.Google.ClientSecret,
			cfg.CallbackBaseURL+"/auth/google/callback",
		),
		github.New(
			cfg.GitHub.ClientID,
			cfg.GitHub.ClientSecret,
			cfg.CallbackBaseURL+"/auth/github/callback",
		),
	)

	gothic.Store = sessions.NewCookieStore([]byte(cfg.SessionSecret))
}

// HandleOAuthCallback returns an http.HandlerFunc that completes the OAuth2
// flow. It:
//  1. Exchanges the OAuth code for a user via gothic.CompleteUserAuth
//  2. Calls findOrCreateUser to obtain the internal user ID
//  3. Stores the user in the SCS session via LoginUser
//  4. Redirects the browser to "/"
func HandleOAuthCallback(sm *scs.SessionManager, findOrCreateUser UserFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gothUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("OAuth callback error: %v", err), http.StatusInternalServerError)
			return
		}

		userID, err := findOrCreateUser(r.Context(), gothUser)
		if err != nil {
			http.Error(w, fmt.Sprintf("find or create user: %v", err), http.StatusInternalServerError)
			return
		}

		if err := LoginUser(sm, r, userID, gothUser.Email); err != nil {
			http.Error(w, fmt.Sprintf("login failed: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// RegisterOAuthRoutes mounts the OAuth2 and email/password auth routes onto
// the given router. The routes are intentionally registered outside any
// RequireSession middleware group — placing auth routes inside RequireSession
// would cause an infinite redirect loop for unauthenticated users.
//
// Routes:
//   - GET  /auth/login          -> HandleLogin (renders login page)
//   - POST /auth/login          -> HandleLoginSubmit (processes email/password)
//   - GET  /auth/logout         -> HandleLogout
//   - GET  /auth/{provider}     -> gothic.BeginAuthHandler (starts OAuth flow)
//   - GET  /auth/{provider}/callback -> HandleOAuthCallback
func RegisterOAuthRoutes(
	router chi.Router,
	sm *scs.SessionManager,
	findOrCreateUser UserFinder,
	authenticateUser PasswordAuthenticator,
) {
	router.Group(func(r chi.Router) {
		r.Get("/auth/login", HandleLogin(sm))
		r.Post("/auth/login", HandleLoginSubmit(sm, authenticateUser))
		r.Get("/auth/logout", HandleLogout(sm))
		r.Get("/auth/{provider}", gothic.BeginAuthHandler)
		r.Get("/auth/{provider}/callback", HandleOAuthCallback(sm, findOrCreateUser))
	})
}

// HandleLogin returns an http.HandlerFunc that renders the login page.
// The page includes an email/password form and OAuth provider buttons for
// Google and GitHub. A templ template replaces this inline HTML in Phase 6.
func HandleLogin(sm *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, loginPageHTML(""))
	}
}

// HandleLoginSubmit returns an http.HandlerFunc that processes an
// email/password form POST. On success it stores the session and redirects to
// "/". On failure it re-renders the login page with an error message.
func HandleLoginSubmit(sm *scs.SessionManager, authenticateUser PasswordAuthenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		userID, err := authenticateUser(r.Context(), email, password)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, loginPageHTML("Invalid email or password."))
			return
		}

		if err := LoginUser(sm, r, userID, email); err != nil {
			http.Error(w, fmt.Sprintf("login failed: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// HandleLogout returns an http.HandlerFunc that destroys the session and
// redirects the user to the login page.
func HandleLogout(sm *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := LogoutUser(sm, r); err != nil {
			http.Error(w, fmt.Sprintf("logout failed: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/auth/login", http.StatusFound)
	}
}

// loginPageHTML returns a minimal HTML login page. The errMsg is displayed
// when non-empty (e.g. "Invalid email or password."). This inline template
// will be replaced by a proper templ component in Phase 6.
func loginPageHTML(errMsg string) string {
	errSection := ""
	if errMsg != "" {
		errSection = fmt.Sprintf(`<p style="color:red">%s</p>`, errMsg)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Login</title></head>
<body>
<h1>Sign in</h1>
%s
<form method="POST" action="/auth/login">
  <label>Email <input type="email" name="email" required></label><br>
  <label>Password <input type="password" name="password" required></label><br>
  <button type="submit">Sign in</button>
</form>
<hr>
<a href="/auth/google">Sign in with Google</a><br>
<a href="/auth/github">Sign in with GitHub</a>
</body>
</html>`, errSection)
}
