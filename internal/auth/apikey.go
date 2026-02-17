package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// PrefixLive is the prefix for live-mode API keys.
	PrefixLive = "forg_live_"
	// PrefixTest is the prefix for test-mode API keys.
	PrefixTest = "forg_test_"
)

// validPrefixes holds all recognized API key prefixes.
var validPrefixes = []string{PrefixLive, PrefixTest}

// APIKey represents an API key stored in the database.
type APIKey struct {
	ID        uuid.UUID
	Name      string
	Key       string
	Prefix    string     // "forg_live_" or "forg_test_"
	UserID    uuid.UUID
	Scopes    []string
	ExpiresAt *time.Time // nullable
	RevokedAt *time.Time // nullable
	CreatedAt time.Time
}

// APIKeyStore defines the interface for API key persistence.
// Implementations are provided by the database layer.
type APIKeyStore interface {
	// GetByKey looks up an API key by its full value.
	GetByKey(ctx context.Context, key string) (*APIKey, error)
	// Create creates a new API key for the given user.
	Create(ctx context.Context, userID uuid.UUID, name string, prefix string, scopes []string, expiresAt *time.Time) (*APIKey, error)
	// Revoke marks a key as revoked by ID.
	Revoke(ctx context.Context, keyID uuid.UUID) error
}

// GenerateAPIKey generates a new API key with the given prefix followed by
// 24 random hex characters, e.g. "forg_live_a1b2c3d4e5f6a1b2c3d4e5f6".
func GenerateAPIKey(prefix string) (string, error) {
	b := make([]byte, 12) // 12 bytes = 24 hex chars
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}

// ValidateKeyPrefix checks whether key starts with a recognised forge API key
// prefix. It returns the matched prefix and true on success, or "" and false
// if the key does not match any known prefix.
func ValidateKeyPrefix(key string) (prefix string, ok bool) {
	for _, p := range validPrefixes {
		if strings.HasPrefix(key, p) {
			return p, true
		}
	}
	return "", false
}

// IsAPIKey reports whether key starts with a recognised forge API key prefix.
// It is a convenience wrapper around ValidateKeyPrefix for boolean checks.
func IsAPIKey(key string) bool {
	_, ok := ValidateKeyPrefix(key)
	return ok
}
