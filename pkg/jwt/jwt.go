package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/hiamthach108/dreon-auth/config"
)

// Signing algorithm: asymmetric RS256.
const SigningMethodAlg = "RS256"

var (
	ErrInvalidToken = errors.New("jwt: invalid token")
	ErrInvalidKey   = errors.New("jwt: invalid key")
)

// IJwtTokenManager defines the contract for generating and verifying JWTs (asymmetric).
type IJwtTokenManager interface {
	Generate(ctx context.Context, payload Payload, expiry time.Duration) (string, error)
	Verify(ctx context.Context, tokenString string) (*Payload, error)
}

// Manager implements IJwtTokenManager using RS256 (RSA private key to sign, public key to verify).
type JwtTokenManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	audience   []string
}

// Option configures a Manager.
type Option func(*JwtTokenManager)

// WithIssuer sets the issuer (iss) claim.
func WithIssuer(issuer string) Option {
	return func(m *JwtTokenManager) { m.issuer = issuer }
}

// WithAudience sets the audience (aud) claim.
func WithAudience(audience ...string) Option {
	return func(m *JwtTokenManager) { m.audience = audience }
}

// NewJwtTokenManagerFromConfig creates a JWT manager from config.
// JWT_PRIVATE_KEY and JWT_PUBLIC_KEY may be either:
// - PEM string (e.g. "-----BEGIN RSA PRIVATE KEY-----\n...")
// - Raw base64-encoded DER (PKCS#1 or PKCS#8 for private, PKIX for public)
func NewJwtTokenManagerFromConfig(cfg *config.AppConfig) (IJwtTokenManager, error) {
	privateKey, err := parseRSAPrivateKeyFromString(cfg.Jwt.PrivateKey)
	if err != nil {
		return nil, err
	}
	publicKey, err := parseRSAPublicKeyFromString(cfg.Jwt.PublicKey)
	if err != nil {
		return nil, err
	}
	return NewJwtTokenManager(privateKey, publicKey, WithIssuer(cfg.App.Name))
}

// parseRSAPrivateKeyFromString parses an RSA private key from a string.
// Accepts PEM (with -----BEGIN ...-----) or raw base64-encoded DER (PKCS#1 or PKCS#8).
func parseRSAPrivateKeyFromString(s string) (*rsa.PrivateKey, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrInvalidKey
	}
	// PEM format
	if strings.Contains(s, "-----BEGIN") {
		key, err := gojwt.ParseRSAPrivateKeyFromPEM([]byte(s))
		if err != nil {
			return nil, err
		}
		return key, nil
	}
	// Raw base64 DER
	der, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	// Try PKCS#1 first, then PKCS#8
	key, err := x509.ParsePKCS1PrivateKey(der)
	if err == nil {
		return key, nil
	}
	generic, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, err
	}
	k, ok := generic.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidKey
	}
	return k, nil
}

// parseRSAPublicKeyFromString parses an RSA public key from a string.
// Accepts PEM (with -----BEGIN ...-----) or raw base64-encoded DER (PKIX or PKCS#1).
func parseRSAPublicKeyFromString(s string) (*rsa.PublicKey, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrInvalidKey
	}
	// PEM format
	if strings.Contains(s, "-----BEGIN") {
		key, err := gojwt.ParseRSAPublicKeyFromPEM([]byte(s))
		if err != nil {
			return nil, err
		}
		return key, nil
	}
	// Raw base64 DER
	der, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	// Try PKIX first, then PKCS#1
	pub, err := x509.ParsePKIXPublicKey(der)
	if err == nil {
		k, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, ErrInvalidKey
		}
		return k, nil
	}
	key, err := x509.ParsePKCS1PublicKey(der)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// NewJwtTokenManager creates a JWT manager that signs with the private key and verifies with the public key.
// Keys must be PEM-encoded RSA; use ParseRSAPrivateKeyFromPEM / ParseRSAPublicKeyFromPEM to obtain them.
func NewJwtTokenManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, opts ...Option) (IJwtTokenManager, error) {
	if privateKey == nil {
		return nil, ErrInvalidKey
	}
	if publicKey == nil {
		return nil, ErrInvalidKey
	}
	m := &JwtTokenManager{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m, nil
}

// NewManagerFromPEM creates a Manager from PEM-encoded private and public key bytes.
// Private key PEM can be PKCS#1 or PKCS#8; public key PEM can be PKCS#1 or PKCS#8.
func NewManagerFromPEM(privateKeyPEM, publicKeyPEM []byte, opts ...Option) (IJwtTokenManager, error) {
	privateKey, err := gojwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	publicKey, err := gojwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, err
	}
	return NewJwtTokenManager(privateKey, publicKey, opts...)
}

// Generate signs a new JWT with the given payload and expiry using RS256.
func (m *JwtTokenManager) Generate(ctx context.Context, payload Payload, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  m.audience,
			Subject:   payload.UserID,
			IssuedAt:  gojwt.NewNumericDate(now),
			NotBefore: gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(expiry)),
			ID:        "",
		},
		Payload: payload,
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, &claims)
	tokenString, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Verify parses and verifies the token with the public key and returns the payload.
func (m *JwtTokenManager) Verify(ctx context.Context, tokenString string) (*Payload, error) {
	token, err := gojwt.ParseWithClaims(tokenString, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return &claims.Payload, nil
}
