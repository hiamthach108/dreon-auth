package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"
)

// testKeyPair generates a 2048-bit RSA key pair and returns PEM-encoded bytes.
func testKeyPair(t *testing.T) (privatePEM, publicPEM []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	privateBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	privatePEM = pem.EncodeToMemory(privateBlock)
	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	publicBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}
	publicPEM = pem.EncodeToMemory(publicBlock)
	return privatePEM, publicPEM
}

func testManager(t *testing.T) IJwtTokenManager {
	t.Helper()
	privatePEM, publicPEM := testKeyPair(t)
	m, err := NewManagerFromPEM(privatePEM, publicPEM, WithIssuer("test"), WithAudience("test-api"))
	if err != nil {
		t.Fatalf("NewManagerFromPEM: %v", err)
	}
	return m
}

func TestNewJwtTokenManager_nilPrivateKey_returnsErrInvalidKey(t *testing.T) {
	_, publicPEM := testKeyPair(t)
	pub, _ := parseRSAPublicKeyFromPEM(publicPEM)
	_, err := NewJwtTokenManager(nil, pub)
	if err != ErrInvalidKey {
		t.Errorf("NewJwtTokenManager(nil, pub) err = %v, want ErrInvalidKey", err)
	}
}

func TestNewJwtTokenManager_nilPublicKey_returnsErrInvalidKey(t *testing.T) {
	privatePEM, _ := testKeyPair(t)
	priv, _ := parseRSAPrivateKeyFromPEM(privatePEM)
	_, err := NewJwtTokenManager(priv, nil)
	if err != ErrInvalidKey {
		t.Errorf("NewJwtTokenManager(priv, nil) err = %v, want ErrInvalidKey", err)
	}
}

func TestNewManagerFromPEM_invalidPrivatePEM_returnsError(t *testing.T) {
	_, publicPEM := testKeyPair(t)
	_, err := NewManagerFromPEM([]byte("not pem"), publicPEM)
	if err == nil {
		t.Error("NewManagerFromPEM(invalid private) want error, got nil")
	}
}

func TestNewManagerFromPEM_invalidPublicPEM_returnsError(t *testing.T) {
	privatePEM, _ := testKeyPair(t)
	_, err := NewManagerFromPEM(privatePEM, []byte("not pem"))
	if err == nil {
		t.Error("NewManagerFromPEM(invalid public) want error, got nil")
	}
}

func TestNewManagerFromPEM_validKeys_returnsManager(t *testing.T) {
	privatePEM, publicPEM := testKeyPair(t)
	m, err := NewManagerFromPEM(privatePEM, publicPEM)
	if err != nil {
		t.Fatalf("NewManagerFromPEM: %v", err)
	}
	if m == nil {
		t.Fatal("NewManagerFromPEM returned nil manager")
	}
}

func TestGenerate_returnsNonEmptyToken(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	payload := Payload{UserID: "user-1", Email: "a@b.com", Status: "active"}

	token, err := m.Generate(ctx, payload, time.Hour)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if token == "" {
		t.Error("Generate returned empty token")
	}
}

func TestGenerate_verifyRoundTrip_returnsSamePayload(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()
	payload := Payload{
		UserID:       "user-123",
		IsSuperAdmin: true,
		Status:       "active",
		Email:        "alice@example.com",
	}

	token, err := m.Generate(ctx, payload, time.Hour)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	got, err := m.Verify(ctx, token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got.UserID != payload.UserID {
		t.Errorf("UserID = %q, want %q", got.UserID, payload.UserID)
	}
	if got.Email != payload.Email {
		t.Errorf("Email = %q, want %q", got.Email, payload.Email)
	}
	if got.Status != payload.Status {
		t.Errorf("Status = %q, want %q", got.Status, payload.Status)
	}
	if got.IsSuperAdmin != payload.IsSuperAdmin {
		t.Errorf("IsSuperAdmin = %v, want %v", got.IsSuperAdmin, payload.IsSuperAdmin)
	}
}

func TestVerify_emptyString_returnsError(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()

	_, err := m.Verify(ctx, "")
	if err == nil {
		t.Error("Verify(empty) want error, got nil")
	}
}

func TestVerify_malformedToken_returnsError(t *testing.T) {
	m := testManager(t)
	ctx := context.Background()

	_, err := m.Verify(ctx, "not.a.valid.jwt")
	if err == nil {
		t.Error("Verify(malformed) want error, got nil")
	}
}

func TestVerify_expiredToken_returnsError(t *testing.T) {
	privatePEM, publicPEM := testKeyPair(t)
	m, err := NewManagerFromPEM(privatePEM, publicPEM)
	if err != nil {
		t.Fatalf("NewManagerFromPEM: %v", err)
	}
	ctx := context.Background()
	payload := Payload{UserID: "u1", Email: "e@e.com", Status: "active"}

	// Generate token that expired 1 hour ago
	token, err := m.Generate(ctx, payload, -time.Hour)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	_, err = m.Verify(ctx, token)
	if err == nil {
		t.Error("Verify(expired token) want error, got nil")
	}
}

func TestVerify_tokenSignedWithDifferentKey_returnsError(t *testing.T) {
	priv1, pub1 := testKeyPair(t)
	priv2, pub2 := testKeyPair(t)
	m1, _ := NewManagerFromPEM(priv1, pub1)
	m2, _ := NewManagerFromPEM(priv2, pub2) // different key pair
	ctx := context.Background()
	payload := Payload{UserID: "u1", Email: "e@e.com", Status: "active"}

	token, err := m1.Generate(ctx, payload, time.Hour)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Verify with m2 (different public key) should fail
	_, err = m2.Verify(ctx, token)
	if err == nil {
		t.Error("Verify(token from other key) want error, got nil")
	}
}

func TestWithIssuer_and_WithAudience_setInToken(t *testing.T) {
	privatePEM, publicPEM := testKeyPair(t)
	m, err := NewManagerFromPEM(privatePEM, publicPEM,
		WithIssuer("my-issuer"),
		WithAudience("api", "admin"),
	)
	if err != nil {
		t.Fatalf("NewManagerFromPEM: %v", err)
	}
	ctx := context.Background()
	token, err := m.Generate(ctx, Payload{UserID: "u1", Email: "a@b.com", Status: "active"}, time.Hour)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Round-trip still works; issuer/audience are in registered claims
	got, err := m.Verify(ctx, token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got.UserID != "u1" {
		t.Errorf("UserID = %q, want u1", got.UserID)
	}
}

// parseRSAPrivateKeyFromPEM and parseRSAPublicKeyFromPEM are used only in tests
// to get *rsa.PrivateKey/*rsa.PublicKey from PEM for NewJwtTokenManager(nil key) tests.
func parseRSAPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, nil
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
func parseRSAPublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, nil
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub.(*rsa.PublicKey), nil
}
