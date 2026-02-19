package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is the bcrypt work factor used when hashing passwords.
// Cost 12 is recommended for 2026 hardware — significantly higher than the
// default of 10 to slow down offline brute-force attacks.
const BcryptCost = 12

// bcryptMaxLength is the maximum plaintext length that bcrypt will process.
// bcrypt silently truncates inputs longer than 72 bytes, which can lead to
// different plaintexts hashing to the same digest.
const bcryptMaxLength = 72

// HashPassword hashes plaintext using bcrypt with BcryptCost. It returns an
// error if plaintext exceeds 72 bytes (the bcrypt truncation limit) or if the
// hashing operation fails.
func HashPassword(plaintext string) (string, error) {
	if len(plaintext) > bcryptMaxLength {
		return "", errors.New("password exceeds maximum length of 72 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CheckPassword verifies that plaintext matches the stored bcrypt hash. It
// returns nil on success and bcrypt.ErrMismatchedHashAndPassword when the
// password does not match. Other bcrypt errors (e.g. malformed hash) are
// returned as-is.
func CheckPassword(plaintext, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}
