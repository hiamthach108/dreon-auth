package handler

import (
	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/service"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"github.com/hiamthach108/dreon-auth/presentation/http/middleware"
	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	roleSvc          service.IRoleSvc
	logger           logger.ILogger
	verifyJWT        middleware.VerifyJWTMiddleware
	verifySuperAdmin middleware.VerifySuperAdminMiddleware
}

func NewRoleHandler(
	roleSvc service.IRoleSvc,
	logger logger.ILogger,
	verifyJWT middleware.VerifyJWTMiddleware,
	verifySuperAdmin middleware.VerifySuperAdminMiddleware,
) *RoleHandler {
	return &RoleHandler{
		roleSvc:          roleSvc,
		logger:           logger,
		verifyJWT:        verifyJWT,
		verifySuperAdmin: verifySuperAdmin,
	}
}

func (h *RoleHandler) RegisterRoutes(g *echo.Group) {
	// All routes require JWT authentication
	g.Use(echo.MiddlewareFunc(h.verifyJWT))

	// Role CRUD - Create, Update, Delete require super admin for system roles
	g.POST("", h.HandleCreateRole)
	g.GET("/:id", h.HandleGetRole)
	g.PUT("/:id", h.HandleUpdateRole)
	g.DELETE("/:id", h.HandleDeleteRole)
	g.GET("", h.HandleListRoles)

	// User role assignments - require super admin for system roles
	g.POST("/assign", h.HandleAssignRoleToUser)
	g.POST("/remove", h.HandleRemoveRoleFromUser)
	g.GET("/user/:userId", h.HandleGetUserRoles)
}

// HandleCreateRole creates a new role
func (h *RoleHandler) HandleCreateRole(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := HandleValidateBind[dto.CreateRoleReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	// Get JWT payload to check if user is super admin
	payload := middleware.GetJWTPayload(ctx)
	isSuperAdmin := payload != nil && payload.IsSuperAdmin

	result, err := h.roleSvc.CreateRole(ctx, req, isSuperAdmin)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleGetRole retrieves a role by ID
func (h *RoleHandler) HandleGetRole(c echo.Context) error {
	ctx := c.Request().Context()
	roleID := c.Param("id")

	result, err := h.roleSvc.GetRole(ctx, roleID)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleUpdateRole updates an existing role
func (h *RoleHandler) HandleUpdateRole(c echo.Context) error {
	ctx := c.Request().Context()
	roleID := c.Param("id")
	req, err := HandleValidateBind[dto.UpdateRoleReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	// Get JWT payload to check if user is super admin
	payload := middleware.GetJWTPayload(ctx)
	isSuperAdmin := payload != nil && payload.IsSuperAdmin

	result, err := h.roleSvc.UpdateRole(ctx, roleID, req, isSuperAdmin)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleDeleteRole deletes a role
func (h *RoleHandler) HandleDeleteRole(c echo.Context) error {
	ctx := c.Request().Context()
	roleID := c.Param("id")

	// Get JWT payload to check if user is super admin
	payload := middleware.GetJWTPayload(ctx)
	isSuperAdmin := payload != nil && payload.IsSuperAdmin

	if err := h.roleSvc.DeleteRole(ctx, roleID, isSuperAdmin); err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, map[string]string{"message": "Role deleted successfully"})
}

// HandleListRoles lists roles with optional filters
func (h *RoleHandler) HandleListRoles(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := HandleValidateBind[dto.ListRolesReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.roleSvc.ListRoles(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleAssignRoleToUser assigns a role to a user
func (h *RoleHandler) HandleAssignRoleToUser(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := HandleValidateBind[dto.AssignRoleToUserReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	// Get JWT payload to check if user is super admin
	payload := middleware.GetJWTPayload(ctx)
	isSuperAdmin := payload != nil && payload.IsSuperAdmin

	result, err := h.roleSvc.AssignRoleToUser(ctx, req, isSuperAdmin)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleRemoveRoleFromUser removes a role from a user
func (h *RoleHandler) HandleRemoveRoleFromUser(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := HandleValidateBind[dto.RemoveRoleFromUserReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	// Get JWT payload to check if user is super admin
	payload := middleware.GetJWTPayload(ctx)
	isSuperAdmin := payload != nil && payload.IsSuperAdmin

	if err := h.roleSvc.RemoveRoleFromUser(ctx, req, isSuperAdmin); err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, map[string]string{"message": "Role removed from user successfully"})
}

// HandleGetUserRoles retrieves all roles assigned to a user
func (h *RoleHandler) HandleGetUserRoles(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("userId")

	req, err := HandleValidateBind[dto.GetUserRolesReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}
	req.UserID = userID

	result, err := h.roleSvc.GetUserRoles(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}
