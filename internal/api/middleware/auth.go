package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/forge-framework/forge/internal/auth"
)

// contextKey is a typed key for context values set by auth middleware.
// Using a custom type prevents key collisions with other packages.
type contextKey string

const (
	// ContextKeyUserID is set in context when a bearer token is validated.
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyAPIKeyID is set in context when an API key is validated.
	ContextKeyAPIKeyID contextKey = "api_key_id"
	// ContextKeyAPIKeyScopes is set in context with the API key's scopes.
	ContextKeyAPIKeyScopes contextKey = "api_key_scopes"
)

// AuthMiddleware validates bearer tokens and API keys on every request.
// It sets auth context values for downstream handlers on success, and returns
// 401 Unauthorized when no valid credential is presented.
type AuthMiddleware struct {
	api         huma.API
	tokenStore  auth.TokenStore
	apiKeyStore auth.APIKeyStore
}

// NewAuthMiddleware creates an AuthMiddleware that uses the given stores.
// The huma.API is required so the middleware can write structured error
// responses via huma.WriteErr.
func NewAuthMiddleware(api huma.API, tokenStore auth.TokenStore, apiKeyStore auth.APIKeyStore) *AuthMiddleware {
	return &AuthMiddleware{
		api:         api,
		tokenStore:  tokenStore,
		apiKeyStore: apiKeyStore,
	}
}

// Handle implements the Huma middleware interface. It checks the Authorization
// header for a bearer token or an API key and rejects requests without a valid
// credential with HTTP 401.
func (m *AuthMiddleware) Handle(ctx huma.Context, next func(huma.Context)) {
	authHeader := ctx.Header("Authorization")
	apiKeyHeader := ctx.Header("X-API-Key")

	var (
		updatedCtx huma.Context
		err        error
	)

	switch {
	case strings.HasPrefix(authHeader, "Bearer "):
		token := strings.TrimPrefix(authHeader, "Bearer ")
		updatedCtx, err = m.validateBearerToken(ctx, token)
		if err != nil {
			huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Unauthorized", err) //nolint:errcheck
			return
		}

	case auth.IsAPIKey(authHeader):
		updatedCtx, err = m.validateAPIKey(ctx, authHeader)
		if err != nil {
			huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Unauthorized", err) //nolint:errcheck
			return
		}

	case apiKeyHeader != "":
		updatedCtx, err = m.validateAPIKey(ctx, apiKeyHeader)
		if err != nil {
			huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Unauthorized", err) //nolint:errcheck
			return
		}

	default:
		huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Authorization header required") //nolint:errcheck
		return
	}

	next(updatedCtx)
}

// validateBearerToken validates the provided raw token value. It uses
// constant-time comparison to prevent timing attacks. On success it returns an
// updated huma.Context with ContextKeyUserID set.
func (m *AuthMiddleware) validateBearerToken(ctx huma.Context, provided string) (huma.Context, error) {
	stored, err := m.tokenStore.GetByToken(ctx.Context(), provided)
	if err != nil {
		return ctx, err
	}

	// Constant-time comparison to prevent timing attacks (NEVER use ==).
	if subtle.ConstantTimeCompare([]byte(provided), []byte(stored.Token)) != 1 {
		return ctx, &authError{"invalid bearer token"}
	}

	if stored.ExpiresAt.Before(time.Now()) {
		return ctx, &authError{"bearer token has expired"}
	}

	return huma.WithValue(ctx, ContextKeyUserID, stored.UserID), nil
}

// validateAPIKey validates the provided raw API key value. It uses
// constant-time comparison to prevent timing attacks. On success it returns an
// updated huma.Context with ContextKeyAPIKeyID and ContextKeyAPIKeyScopes set.
func (m *AuthMiddleware) validateAPIKey(ctx huma.Context, provided string) (huma.Context, error) {
	if _, ok := auth.ValidateKeyPrefix(provided); !ok {
		return ctx, &authError{"invalid API key prefix"}
	}

	stored, err := m.apiKeyStore.GetByKey(ctx.Context(), provided)
	if err != nil {
		return ctx, err
	}

	// Constant-time comparison to prevent timing attacks (NEVER use ==).
	if subtle.ConstantTimeCompare([]byte(provided), []byte(stored.Key)) != 1 {
		return ctx, &authError{"invalid API key"}
	}

	if stored.RevokedAt != nil {
		return ctx, &authError{"API key has been revoked"}
	}

	if stored.ExpiresAt != nil && stored.ExpiresAt.Before(time.Now()) {
		return ctx, &authError{"API key has expired"}
	}

	ctx = huma.WithValue(ctx, ContextKeyAPIKeyID, stored.ID)
	ctx = huma.WithValue(ctx, ContextKeyAPIKeyScopes, stored.Scopes)
	return ctx, nil
}

// authError is a simple error type for authentication failures.
type authError struct {
	msg string
}

func (e *authError) Error() string { return e.msg }
