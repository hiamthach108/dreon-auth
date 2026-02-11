package handler

import (
	"net/http"
	"strconv"

	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/service"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	echomw "github.com/hiamthach108/dreon-auth/presentation/http/middleware"
	"github.com/labstack/echo/v4"
)

// UserHandler handles HTTP requests for user CRUD.
type UserHandler struct {
	userSvc   service.IUserSvc
	logger    logger.ILogger
	verifyJWT echomw.VerifyJWTMiddleware
}

// NewUserHandler creates a new user handler. verifyJWT is injected by fx for protected routes.
func NewUserHandler(userSvc service.IUserSvc, logger logger.ILogger, verifyJWT echomw.VerifyJWTMiddleware) *UserHandler {
	return &UserHandler{
		userSvc:   userSvc,
		logger:    logger,
		verifyJWT: verifyJWT,
	}
}

// RegisterRoutes registers user routes on the given group and applies JWT verification middleware.
func (h *UserHandler) RegisterRoutes(g *echo.Group) {
	g.Use(echo.MiddlewareFunc(h.verifyJWT))
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

// List returns a paginated list of users.
// Query: page (default 1), pageSize (default 10, max 100).
func (h *UserHandler) List(c echo.Context) error {
	ctx := c.Request().Context()
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	result, err := h.userSvc.List(ctx, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list users", "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, result)
}

// GetByID returns a user by ID.
func (h *UserHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, nil))
	}

	user, err := h.userSvc.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get user", "id", id, "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, user)
}

// Create creates a new user.
func (h *UserHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var req dto.CreateUserReq
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind create user request", "error", err)
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	user, err := h.userSvc.Create(ctx, req)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err)
		return HandleError(c, err)
	}
	return c.JSON(http.StatusCreated, BaseResp{
		Code:    http.StatusCreated,
		Message: "success",
		Data:    user,
	})
}

// Update updates a user by ID.
func (h *UserHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, nil))
	}

	var req dto.UpdateUserReq
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind update user request", "error", err)
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	user, err := h.userSvc.Update(ctx, id, req)
	if err != nil {
		h.logger.Error("Failed to update user", "id", id, "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, user)
}

// Delete deletes a user by ID.
func (h *UserHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, nil))
	}

	if err := h.userSvc.Delete(ctx, id); err != nil {
		h.logger.Error("Failed to delete user", "id", id, "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, nil)
}
