package handler

import (
	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/service"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"github.com/hiamthach108/dreon-auth/presentation/http/middleware"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authSvc   service.IAuthSvc
	logger    logger.ILogger
	verifyJWT middleware.VerifyJWTMiddleware
}

func NewAuthHandler(authSvc service.IAuthSvc, logger logger.ILogger, verifyJWT middleware.VerifyJWTMiddleware) *AuthHandler {
	return &AuthHandler{
		authSvc:   authSvc,
		logger:    logger,
		verifyJWT: verifyJWT,
	}
}

func (h *AuthHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/login", h.HandleLogin)
	g.POST("/register", h.HandleRegister)
	g.POST("/refresh-token", h.HandleRefreshToken)
	g.POST("/logout", h.HandleLogout)
	g.Use(echo.MiddlewareFunc(h.verifyJWT))
	g.GET("/session", h.HandleGetSession)
}

func (h *AuthHandler) HandleLogin(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.LoginReq
	if err := c.Bind(&req); err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.authSvc.Login(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}
	return HandleSuccess(c, result)
}

func (h *AuthHandler) HandleRegister(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.RegisterReq
	if err := c.Bind(&req); err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.authSvc.Register(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}
	return HandleSuccess(c, result)
}

func (h *AuthHandler) HandleRefreshToken(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.RefreshTokenReq
	if err := c.Bind(&req); err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.authSvc.RefreshToken(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}
	return HandleSuccess(c, result)
}

func (h *AuthHandler) HandleLogout(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.LogoutReq
	if err := c.Bind(&req); err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	err := h.authSvc.Logout(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}
	return HandleSuccess(c, nil)
}

func (h *AuthHandler) HandleGetSession(c echo.Context) error {
	ctx := c.Request().Context()
	payload := middleware.GetJWTPayload(ctx)
	if payload == nil {
		return HandleError(c, errorx.New(errorx.ErrUnauthorized, "missing payload"))
	}
	return HandleSuccess(c, payload)
}
