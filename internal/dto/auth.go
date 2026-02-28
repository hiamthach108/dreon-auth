package dto

import (
	"time"

	"github.com/hiamthach108/dreon-auth/internal/shared/constant"
)

type LoginReq struct {
	IsSuperAdmin bool                  `json:"isSuperAdmin"`
	AuthType     constant.UserAuthType `json:"authType" validate:"required,oneof=EMAIL SUPER_ADMIN GOOGLE FACEBOOK APPLE"`
	Email        string                `json:"email"`
	Password     string                `json:"password"`
	RedirectURL  string                `json:"redirectUrl"`
}

type TokenResp struct {
	UserID                string    `json:"userId"`
	SessionID             string    `json:"sessionId"`
	AccessToken           string    `json:"accessToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

type LoginResp struct {
	TokenResp
	RedirectURL  string `json:"redirectUrl,omitempty"`
	RefreshState string `json:"refreshState,omitempty"`
}

// GoogleUserData is the shape returned by Google userinfo / used in store request.
type GoogleUserData struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    string `json:"id"`
}

// OAuthUserData is provider-agnostic user data stored in cache (Google, Facebook, Apple).
type OAuthUserData struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	ProviderID string `json:"providerId"`
}

// CachedOAuthState is the value stored in cache under refresh_state:{state}.
type CachedOAuthState struct {
	AuthType constant.UserAuthType `json:"authType"`
	UserData OAuthUserData         `json:"userData"`
}

// SessionFromStateReq is the request to exchange a valid refreshState for a session.
type SessionFromStateReq struct {
	RefreshState string `json:"refreshState" validate:"required"`
}

type RegisterReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type LogoutReq struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}
