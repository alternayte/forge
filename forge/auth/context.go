package auth

import (
	"context"

	"github.com/google/uuid"
)

// userContextKey is the private type for the user ID context key.
type userContextKey struct{}

// roleContextKey is the private type for the role context key.
type roleContextKey struct{}

// WithUserRole stores both a user ID and role in the context.
// These values can be retrieved with UserFromContext and RoleFromContext.
func WithUserRole(ctx context.Context, userID uuid.UUID, role string) context.Context {
	ctx = context.WithValue(ctx, userContextKey{}, userID)
	ctx = context.WithValue(ctx, roleContextKey{}, role)
	return ctx
}

// UserFromContext retrieves the authenticated user's UUID from the context.
// Returns uuid.Nil if no user ID has been stored.
func UserFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(userContextKey{}).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// RoleFromContext retrieves the authenticated user's role from the context.
// Returns an empty string if no role has been stored.
func RoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value(roleContextKey{}).(string); ok {
		return role
	}
	return ""
}
