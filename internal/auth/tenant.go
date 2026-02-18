package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// tenantContextKey is a private context key type for storing tenant IDs.
// Using a struct{} type prevents key collisions with other packages.
type tenantContextKey struct{}

// WithTenant stores a tenant ID in the context. Use TenantFromContext to retrieve it.
func WithTenant(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, id)
}

// TenantFromContext retrieves the tenant ID from the context.
// Returns the tenant ID and true if found, or uuid.Nil and false if not present.
func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(tenantContextKey{}).(uuid.UUID)
	return id, ok
}

// TenantResolver extracts a tenant ID from an HTTP request.
// Implement this interface to define your tenant identification strategy.
type TenantResolver interface {
	Resolve(r *http.Request) (uuid.UUID, error)
}

// HeaderTenantResolver reads the tenant ID from a request header.
// The Header field defaults to X-Tenant-ID if not set.
type HeaderTenantResolver struct {
	// Header is the HTTP header name to read the tenant UUID from.
	// Defaults to "X-Tenant-ID" if empty.
	Header string
}

// Resolve extracts the tenant ID from the configured header.
func (h HeaderTenantResolver) Resolve(r *http.Request) (uuid.UUID, error) {
	header := h.Header
	if header == "" {
		header = "X-Tenant-ID"
	}
	raw := r.Header.Get(header)
	if raw == "" {
		return uuid.Nil, fmt.Errorf("missing tenant header %s", header)
	}
	return uuid.Parse(raw)
}

// SubdomainTenantResolver extracts the tenant ID from the first subdomain segment
// of the Host header. For example, host "abc123.example.com" extracts "abc123".
// The subdomain must be a valid UUID.
type SubdomainTenantResolver struct{}

// Resolve extracts the tenant ID from the first subdomain segment of the Host header.
func (s SubdomainTenantResolver) Resolve(r *http.Request) (uuid.UUID, error) {
	host := r.Host
	// Strip port if present.
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	parts := strings.SplitN(host, ".", 2)
	if len(parts) < 2 || parts[0] == "" {
		return uuid.Nil, fmt.Errorf("no subdomain found in host %q", r.Host)
	}
	id, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, fmt.Errorf("subdomain %q is not a valid tenant UUID: %w", parts[0], err)
	}
	return id, nil
}

// PathTenantResolver extracts the tenant ID from a URL path prefix.
// It expects paths of the form /tenants/{uuid}/... and strips the prefix
// before forwarding the request.
type PathTenantResolver struct{}

// Resolve extracts the tenant UUID from the second path segment (e.g. /tenants/{uuid}/...).
func (p PathTenantResolver) Resolve(r *http.Request) (uuid.UUID, error) {
	// Expect /tenants/{uuid}/... or /{uuid}/...
	path := r.URL.Path
	// Trim leading slash and split.
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 3)

	// Try /tenants/{uuid}/... first.
	if len(parts) >= 2 && parts[0] == "tenants" {
		id, err := uuid.Parse(parts[1])
		if err != nil {
			return uuid.Nil, fmt.Errorf("path segment %q is not a valid tenant UUID: %w", parts[1], err)
		}
		return id, nil
	}

	// Fall back to /{uuid}/...
	if len(parts) >= 1 && parts[0] != "" {
		id, err := uuid.Parse(parts[0])
		if err != nil {
			return uuid.Nil, fmt.Errorf("path segment %q is not a valid tenant UUID: %w", parts[0], err)
		}
		return id, nil
	}

	return uuid.Nil, fmt.Errorf("no tenant UUID found in path %q", r.URL.Path)
}

// TenantMiddleware returns an HTTP middleware that resolves the tenant from each
// request and stores it in the context via WithTenant. If the resolver returns an
// error, the middleware responds with 401 Unauthorized and does not call next.
func TenantMiddleware(resolver TenantResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID, err := resolver.Resolve(r)
			if err != nil {
				http.Error(w, "tenant not found", http.StatusUnauthorized)
				return
			}
			ctx := WithTenant(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
