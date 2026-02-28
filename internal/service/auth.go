package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hiamthach108/dreon-auth/config"
	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/model"
	"github.com/hiamthach108/dreon-auth/internal/repository"
	"github.com/hiamthach108/dreon-auth/internal/shared/constant"
	"github.com/hiamthach108/dreon-auth/internal/shared/helper"
	"github.com/hiamthach108/dreon-auth/pkg/cache"
	"github.com/hiamthach108/dreon-auth/pkg/jwt"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/datatypes"
)

type IAuthSvc interface {
	Login(ctx context.Context, req dto.LoginReq) (*dto.LoginResp, error)
	Register(ctx context.Context, req dto.RegisterReq) (*dto.TokenResp, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenReq) (*dto.TokenResp, error)
	Logout(ctx context.Context, req dto.LogoutReq) error
	ValidateToken(ctx context.Context, token string) (*jwt.Payload, error)
	SessionFromState(ctx context.Context, req dto.SessionFromStateReq) (*dto.TokenResp, error)
	ExchangeGoogleCode(ctx context.Context, code, state string) (redirectURL string, err error)
}

type AuthSvc struct {
	logger             logger.ILogger
	jwtTokenManager    jwt.IJwtTokenManager
	cfg                config.AppConfig
	userRepo           repository.IUserRepository
	sessionRepo        repository.ISessionRepository
	projectRepo        repository.IProjectRepository
	superAdminRepo     repository.ISuperAdminRepository
	cache              cache.ICache
	googleOAuth2Config *oauth2.Config
}

func NewAuthSvc(
	logger logger.ILogger,
	jwtTokenManager jwt.IJwtTokenManager,
	cfg *config.AppConfig,
	cache cache.ICache,
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
		cache:           cache,
		googleOAuth2Config: &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (s *AuthSvc) Login(ctx context.Context, req dto.LoginReq) (*dto.LoginResp, error) {
	switch req.AuthType {
	case constant.UserAuthTypeEmail:
		tokenResp, err := s.loginWithEmail(ctx, req)
		if err != nil {
			return nil, err
		}
		return &dto.LoginResp{
			TokenResp: *tokenResp,
		}, nil
	case constant.UserAuthTypeSuperAdmin:
		tokenResp, err := s.loginWithSuperAdmin(ctx, req)
		if err != nil {
			return nil, err
		}
		return &dto.LoginResp{
			TokenResp: *tokenResp,
		}, nil
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
		Email:        session.Email,
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

func (s *AuthSvc) ExchangeGoogleCode(ctx context.Context, code, state string) (redirectURL string, err error) {
	if code == "" || state == "" {
		return "", errorx.New(errorx.ErrBadRequest, "code and state are required")
	}
	token, err := s.googleOAuth2Config.Exchange(ctx, code)
	if err != nil {
		return "", errorx.Wrap(errorx.ErrUnauthorized, fmt.Errorf("google token exchange: %w", err))
	}
	userInfo, err := s.fetchGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		return "", errorx.Wrap(errorx.ErrInternal, err)
	}
	cached := dto.CachedOAuthState{
		AuthType: constant.UserAuthTypeGoogle,
		UserData: dto.OAuthUserData{
			Email:      userInfo.Email,
			Name:       userInfo.Name,
			ProviderID: userInfo.ID,
		},
	}
	stateKey := s.buildRefreshStateCacheKey(ctx, state)
	ttl := constant.RefreshStateTTL
	if err := s.cache.Set(stateKey, cached, &ttl); err != nil {
		return "", errorx.Wrap(errorx.ErrInternal, err)
	}
	redirectKey := s.buildOAuthRedirectCacheKey(ctx, state)
	var redirectPayload struct {
		URL string `json:"url"`
	}
	if getErr := s.cache.Get(redirectKey, &redirectPayload); getErr == nil {
		_ = s.cache.Delete(redirectKey)
		frontendRedirect := redirectPayload.URL
		u, err := url.Parse(frontendRedirect)
		if err != nil {
			redirectURL = frontendRedirect + "?refreshState=" + url.QueryEscape(state)
		} else {
			q := u.Query()
			q.Set("refreshState", state)
			u.RawQuery = q.Encode()
			redirectURL = u.String()
		}
	}
	if redirectURL == "" {
		return "", errorx.New(errorx.ErrBadRequest, "missing redirect_uri; pass redirectUrl in login request")
	}
	return redirectURL, nil
}

func (s *AuthSvc) SessionFromState(ctx context.Context, req dto.SessionFromStateReq) (*dto.TokenResp, error) {
	key := s.buildRefreshStateCacheKey(ctx, req.RefreshState)
	var cached dto.CachedOAuthState
	if err := s.cache.Get(key, &cached); err != nil {
		if err == cache.ErrCacheNil {
			return nil, errorx.New(errorx.ErrInvalidRefreshState, errorx.GetErrorMessage(int(errorx.ErrInvalidRefreshState)))
		}
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if err := s.cache.Delete(key); err != nil {
		s.logger.Error("failed to delete refresh state after use", "key", key, "error", err)
	}
	userData := cached.UserData
	if userData.Email == "" {
		return nil, errorx.New(errorx.ErrInvalidRefreshState, errorx.GetErrorMessage(int(errorx.ErrInvalidRefreshState)))
	}
	authType := cached.AuthType
	user, err := s.userRepo.FindByEmail(ctx, userData.Email)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if user == nil {
		randomPass, err := helper.GenerateRefreshToken()
		if err != nil {
			return nil, errorx.Wrap(errorx.ErrInternal, err)
		}
		hashed, err := helper.HashPassword(randomPass)
		if err != nil {
			return nil, errorx.Wrap(errorx.ErrInternal, err)
		}
		user, err = s.userRepo.Create(ctx, &model.User{
			Username:   userData.Email,
			Email:      userData.Email,
			Password:   hashed,
			Status:     constant.UserStatusActive,
			AuthType:   authType,
			AuthTypeID: userData.ProviderID,
		})
		if err != nil {
			return nil, errorx.Wrap(errorx.ErrInternal, err)
		}
	} else {
		if err := s.updateLastLoginAt(ctx, user.ID); err != nil {
			return nil, errorx.Wrap(errorx.ErrInternal, err)
		}
	}
	return s.generateTokens(ctx, jwt.Payload{
		UserID:       user.ID,
		IsSuperAdmin: false,
		Email:        user.Email,
	})
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
		Email:        payload.Email,
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

func (s *AuthSvc) loginWithGoogle(ctx context.Context, req dto.LoginReq) (*dto.LoginResp, error) {
	refreshState, err := helper.GenerateRefreshToken()
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if req.RedirectURL != "" {
		redirectKey := s.buildOAuthRedirectCacheKey(ctx, refreshState)
		ttl := constant.RefreshStateTTL
		if err := s.cache.Set(redirectKey, struct {
			URL string `json:"url"`
		}{URL: req.RedirectURL}, &ttl); err != nil {
			return nil, errorx.Wrap(errorx.ErrInternal, err)
		}
	}
	authURL, err := s.buildGoogleAuthURL(refreshState)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResp{
		RefreshState: refreshState,
		RedirectURL:  authURL,
	}, nil
}

func (s *AuthSvc) buildGoogleAuthURL(state string) (string, error) {
	return s.googleOAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")), nil
}

func (s *AuthSvc) fetchGoogleUserInfo(ctx context.Context, accessToken string) (*dto.GoogleUserData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned %d", resp.StatusCode)
	}
	var info dto.GoogleUserData
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *AuthSvc) buildOAuthRedirectCacheKey(ctx context.Context, state string) string {
	return fmt.Sprintf("oauth_redirect:%s", state)
}

func (s *AuthSvc) loginWithFacebook(ctx context.Context, req dto.LoginReq) (*dto.LoginResp, error) {
	panic("not implemented")
}

func (s *AuthSvc) loginWithApple(ctx context.Context, req dto.LoginReq) (*dto.LoginResp, error) {
	panic("not implemented")
}

func (s *AuthSvc) updateLastLoginAt(ctx context.Context, userID string) error {
	return s.userRepo.Update(ctx, userID, model.User{
		LastLoginAt: time.Now(),
	}, "last_login_at")
}

func (s *AuthSvc) buildRefreshStateCacheKey(ctx context.Context, state string) string {
	return fmt.Sprintf("refresh_state:%s", state)
}

func metadataFromContext(ctx context.Context) map[string]any {
	str := func(k constant.ContextKey) string { v := ctx.Value(k); s, _ := v.(string); return s }
	return map[string]any{"ip": str(constant.ContextKeyClientIP), "user_agent": str(constant.ContextKeyUserAgent), "referer": str(constant.ContextKeyReferer)}
}
