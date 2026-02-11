package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost (10). Higher values are more secure but slower.
	DefaultCost = bcrypt.DefaultCost
)

// HashPassword hashes a plaintext password using bcrypt.
// Returns the hashed password as a string, or an error if hashing fails.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// ComparePassword compares a plaintext password with a bcrypt hash.
// Returns nil if they match; returns bcrypt.ErrMismatchedHashAndPassword otherwise.
func ComparePassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32) // 256-bit token
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashRefreshToken returns a SHA256 hex digest of the token for storage and lookup.
// Use this when storing the refresh token in the session table and when looking up by token.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
