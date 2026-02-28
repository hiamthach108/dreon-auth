package handler

import (
	"github.com/hiamthach108/dreon-auth/internal/shared/permission"
	"github.com/hiamthach108/dreon-auth/presentation/http/middleware"
	"github.com/labstack/echo/v4"
)

type PermissionHandler struct {
	registry  *permission.Registry
	verifyJWT middleware.VerifyJWTMiddleware
}

func NewPermissionHandler(registry *permission.Registry, verifyJWT middleware.VerifyJWTMiddleware) *PermissionHandler {
	return &PermissionHandler{registry: registry, verifyJWT: verifyJWT}
}

func (h *PermissionHandler) RegisterRoutes(g *echo.Group) {
	g.Use(echo.MiddlewareFunc(h.verifyJWT))
	g.GET("", h.HandleListPermissions)
}

func (h *PermissionHandler) HandleListPermissions(c echo.Context) error {
	if h.registry == nil {
		return HandleSuccess(c, []struct{}{})
	}
	list := h.registry.List()
	return HandleSuccess(c, list)
}
