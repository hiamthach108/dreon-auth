package dto

import (
	"time"

	"github.com/hiamthach108/dreon-auth/internal/shared/constant"
)

type LoginReq struct {
	IsSuperAdmin bool                  `json:"isSuperAdmin"`
	AuthType     constant.UserAuthType `json:"auth_type" binding:"required,oneof=email google facebook apple"`
	Email        string                `json:"email"`
	Password     string                `json:"password"`
}

type TokenResp struct {
	UserID                string    `json:"userId"`
	SessionID             string    `json:"sessionId"`
	AccessToken           string    `json:"accessToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
