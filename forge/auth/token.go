package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// Token represents a bearer token stored in the database.
type Token struct {
	ID        uuid.UUID
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
}

// TokenStore defines the interface for bearer token persistence.
// Implementations are provided by the database layer.
type TokenStore interface {
	// GetByToken looks up a token by its value.
	GetByToken(ctx context.Context, token string) (*Token, error)
	// Create creates a new bearer token for the given user expiring at expiresAt.
	Create(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*Token, error)
	// Delete revokes a token by ID.
	Delete(ctx context.Context, tokenID uuid.UUID) error
}

// GenerateToken generates a cryptographically random 64-character hex string
// (32 random bytes hex-encoded) suitable for use as a bearer token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
