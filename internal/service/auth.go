package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hiamthach108/dreon-auth/config"
	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/model"
	"github.com/hiamthach108/dreon-auth/internal/repository"
	"github.com/hiamthach108/dreon-auth/internal/shared/constant"
	"github.com/hiamthach108/dreon-auth/internal/shared/helper"
	"github.com/hiamthach108/dreon-auth/pkg/jwt"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"gorm.io/datatypes"
)

type IAuthSvc interface {
	Login(ctx context.Context, req dto.LoginReq) (*dto.TokenResp, error)
	Register(ctx context.Context, req dto.RegisterReq) (*dto.TokenResp, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenReq) (*dto.TokenResp, error)
	Logout(ctx context.Context, req dto.LogoutReq) error
	ValidateToken(ctx context.Context, token string) (*jwt.Payload, error)
}

type AuthSvc struct {
	logger          logger.ILogger
	jwtTokenManager jwt.IJwtTokenManager
	cfg             config.AppConfig
	userRepo        repository.IUserRepository
	sessionRepo     repository.ISessionRepository
	projectRepo     repository.IProjectRepository
	superAdminRepo  repository.ISuperAdminRepository
}

func NewAuthSvc(
	logger logger.ILogger,
	jwtTokenManager jwt.IJwtTokenManager,
	cfg *config.AppConfig,
	userRepo repository.IUserRepository,
	sessionRepo repository.ISessionRepository,
	projectRepo repository.IProjectRepository,
	superAdminRepo repository.ISuperAdminRepository,
) IAuthSvc {
	return &AuthSvc{
		logger:          logger,
		jwtTokenManager: jwtTokenManager,
		cfg:             *cfg,
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		projectRepo:     projectRepo,
		superAdminRepo:  superAdminRepo,
	}
}

func (s *AuthSvc) Login(ctx context.Context, req dto.LoginReq) (*dto.TokenResp, error) {
	switch req.AuthType {
	case constant.UserAuthTypeEmail:
		return s.loginWithEmail(ctx, req)
	case constant.UserAuthTypeSuperAdmin:
		return s.loginWithSuperAdmin(ctx, req)
	case constant.UserAuthTypeGoogle:
		return s.loginWithGoogle(ctx, req)
	case constant.UserAuthTypeFacebook:
		return s.loginWithFacebook(ctx, req)
	case constant.UserAuthTypeApple:
		return s.loginWithApple(ctx, req)
	default:
		return nil, errorx.Wrap(errorx.ErrInvalidAuthType, fmt.Errorf("invalid auth type: %s", req.AuthType))
	}
}

func (s *AuthSvc) Register(ctx context.Context, req dto.RegisterReq) (*dto.TokenResp, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if existing != nil {
		return nil, errorx.New(errorx.ErrUserConflict, errorx.GetErrorMessage(int(errorx.ErrUserConflict)))
	}
	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	user, err := s.userRepo.Create(ctx, &model.User{
		Username: req.Email,
		Email:    req.Email,
		Password: hashed,
		Status:   constant.UserStatusActive,
	})

	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	return s.generateTokens(ctx, jwt.Payload{
		UserID:       user.ID,
		IsSuperAdmin: false,
		Email:        user.Email,
	})
}

func (s *AuthSvc) RefreshToken(ctx context.Context, req dto.RefreshTokenReq) (*dto.TokenResp, error) {
	session := s.sessionRepo.FindByRefreshToken(ctx, req.RefreshToken)
	if session == nil {
		return nil, errorx.New(errorx.ErrInvalidRefreshToken, errorx.GetErrorMessage(int(errorx.ErrInvalidRefreshToken)))
	}
	if session.ExpiresAt.Before(time.Now()) || !session.IsActive {
		return nil, errorx.New(errorx.ErrRefreshTokenExpired, errorx.GetErrorMessage(int(errorx.ErrRefreshTokenExpired)))
	}
	return s.generateTokens(ctx, jwt.Payload{
		UserID:       session.UserID,
		IsSuperAdmin: session.IsSuperAdmin,
		Email:        session.User.Email,
	})
}

func (s *AuthSvc) Logout(ctx context.Context, req dto.LogoutReq) error {
	// remove refresh token from session table
	session := s.sessionRepo.FindByRefreshToken(ctx, req.RefreshToken)
	if session == nil {
		return errorx.New(errorx.ErrInvalidRefreshToken, errorx.GetErrorMessage(int(errorx.ErrInvalidRefreshToken)))
	}
	session.IsActive = false
	return s.sessionRepo.Update(ctx, session.ID, *session, "is_active")
}

func (s *AuthSvc) ValidateToken(ctx context.Context, token string) (*jwt.Payload, error) {
	payload, err := s.jwtTokenManager.Verify(ctx, token)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrUnauthorized, err)
	}
	return payload, nil
}

func (s *AuthSvc) generateTokens(ctx context.Context, payload jwt.Payload) (*dto.TokenResp, error) {
	refreshToken, err := helper.GenerateRefreshToken()
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	accessToken, err := s.jwtTokenManager.Generate(ctx, payload, time.Duration(s.cfg.Jwt.AccessTokenExpiresIn)*time.Second)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	metaJSON, _ := json.Marshal(metadataFromContext(ctx))
	accessExp := time.Duration(s.cfg.Jwt.AccessTokenExpiresIn) * time.Second
	refreshExp := time.Duration(s.cfg.Jwt.RefreshTokenExpiresIn) * time.Second
	session, err := s.sessionRepo.Create(ctx, &model.Session{
		UserID:       payload.UserID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(refreshExp),
		IsSuperAdmin: payload.IsSuperAdmin,
		IsActive:     true,
		BaseModel: model.BaseModel{
			CreatedBy: payload.UserID,
			UpdatedBy: payload.UserID,
			Metadata:  datatypes.JSON(metaJSON),
		},
	})
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	return &dto.TokenResp{
		UserID:                payload.UserID,
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  time.Now().Add(accessExp),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: time.Now().Add(refreshExp),
	}, nil
}

func metadataFromContext(ctx context.Context) map[string]any {
	str := func(k constant.ContextKey) string { v := ctx.Value(k); s, _ := v.(string); return s }
	return map[string]any{"ip": str(constant.ContextKeyClientIP), "user_agent": str(constant.ContextKeyUserAgent), "referer": str(constant.ContextKeyReferer)}
}

func (s *AuthSvc) loginWithSuperAdmin(ctx context.Context, req dto.LoginReq) (*dto.TokenResp, error) {
	user, err := s.superAdminRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if user == nil {
		return nil, errorx.New(errorx.ErrUserNotFound, errorx.GetErrorMessage(int(errorx.ErrUserNotFound)))
	}
	if err := helper.ComparePassword(user.Password, req.Password); err != nil {
		return nil, errorx.New(errorx.ErrInvalidPassword, errorx.GetErrorMessage(int(errorx.ErrInvalidPassword)))
	}

	tokenResp, err := s.generateTokens(ctx, jwt.Payload{
		UserID:       user.ID,
		IsSuperAdmin: true,
		Email:        user.Email,
	})

	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	err = s.updateLastLoginAt(ctx, user.ID)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	return tokenResp, nil
}

func (s *AuthSvc) loginWithEmail(ctx context.Context, req dto.LoginReq) (*dto.TokenResp, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if user == nil {
		return nil, errorx.New(errorx.ErrUserNotFound, errorx.GetErrorMessage(int(errorx.ErrUserNotFound)))
	}
	if err := helper.ComparePassword(user.Password, req.Password); err != nil {
		return nil, errorx.New(errorx.ErrInvalidPassword, errorx.GetErrorMessage(int(errorx.ErrInvalidPassword)))
	}

	tokenResp, err := s.generateTokens(ctx, jwt.Payload{
		UserID:       user.ID,
		IsSuperAdmin: false,
		Email:        user.Email,
	})
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	err = s.updateLastLoginAt(ctx, user.ID)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	return tokenResp, nil
}

func (s *AuthSvc) loginWithGoogle(ctx context.Context, req dto.LoginReq) (*dto.TokenResp, error) {
	panic("not implemented")
}

func (s *AuthSvc) loginWithFacebook(ctx context.Context, req dto.LoginReq) (*dto.TokenResp, error) {
	panic("not implemented")
}

func (s *AuthSvc) loginWithApple(ctx context.Context, req dto.LoginReq) (*dto.TokenResp, error) {
	panic("not implemented")
}

func (s *AuthSvc) updateLastLoginAt(ctx context.Context, userID string) error {
	return s.userRepo.Update(ctx, userID, model.User{
		LastLoginAt: time.Now(),
	}, "last_login_at")
}
