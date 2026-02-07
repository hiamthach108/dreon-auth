package jwt

import gojwt "github.com/golang-jwt/jwt/v5"

// Payload holds application-specific claims (no expiry/audience — use Claims for full JWT).
type Payload struct {
	UserID       string `json:"user_id"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	Status       string `json:"status"`
	Email        string `json:"email"`
}

// Claims embeds standard registered claims (exp, iat, nbf, iss, sub, jti) and Payload for JWT signing/verification.
type Claims struct {
	gojwt.RegisteredClaims
	Payload
}
