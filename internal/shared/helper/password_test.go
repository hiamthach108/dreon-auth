package helper

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	plain := "mySecurePassword123"
	hashed, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword(%q) err = %v", plain, err)
	}
	if hashed == "" {
		t.Error("HashPassword returned empty string")
	}
	if hashed == plain {
		t.Error("HashPassword returned plaintext")
	}
	if !strings.HasPrefix(hashed, "$2") {
		t.Errorf("HashPassword should produce bcrypt hash (prefix $2a or $2b), got %q", hashed[:3])
	}
}

func TestComparePassword_match(t *testing.T) {
	plain := "matchMe"
	hashed, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := ComparePassword(hashed, plain); err != nil {
		t.Errorf("ComparePassword(match) err = %v, want nil", err)
	}
}

func TestComparePassword_mismatch(t *testing.T) {
	plain := "original"
	hashed, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	err = ComparePassword(hashed, "wrongPassword")
	if err == nil {
		t.Error("ComparePassword(mismatch) err = nil, want non-nil")
	}
}

func TestComparePassword_invalidHash(t *testing.T) {
	err := ComparePassword("not-a-valid-bcrypt-hash", "password")
	if err == nil {
		t.Error("ComparePassword(invalid hash) err = nil, want non-nil")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() err = %v", err)
	}
	if token == "" {
		t.Error("GenerateRefreshToken returned empty string")
	}
	// Base64 RawURLEncoding of 32 bytes => 43 chars
	if len(token) != 43 {
		t.Errorf("GenerateRefreshToken() len = %d, want 43", len(token))
	}
	// Should be different each time
	token2, _ := GenerateRefreshToken()
	if token == token2 {
		t.Error("GenerateRefreshToken returned same value twice")
	}
}

func TestHashRefreshToken(t *testing.T) {
	input := "my-refresh-token"
	got := HashRefreshToken(input)
	// SHA256 hex = 64 chars
	if len(got) != 64 {
		t.Errorf("HashRefreshToken len = %d, want 64", len(got))
	}
	// Deterministic
	got2 := HashRefreshToken(input)
	if got != got2 {
		t.Error("HashRefreshToken should be deterministic")
	}
	// Different input => different hash
	other := HashRefreshToken("other-token")
	if got == other {
		t.Error("HashRefreshToken different input should produce different hash")
	}
}
