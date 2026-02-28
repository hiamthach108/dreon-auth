package handler

import (
	"strconv"

	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/service"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	echomw "github.com/hiamthach108/dreon-auth/presentation/http/middleware"
	"github.com/labstack/echo/v4"
)

// ProjectHandler handles HTTP requests for project CRUD.
type ProjectHandler struct {
	projectSvc       service.IProjectSvc
	logger           logger.ILogger
	verifyJWT        echomw.VerifyJWTMiddleware
	verifySuperAdmin echomw.VerifySuperAdminMiddleware
}

// NewProjectHandler creates a new project handler.
func NewProjectHandler(projectSvc service.IProjectSvc, logger logger.ILogger, verifyJWT echomw.VerifyJWTMiddleware, verifySuperAdmin echomw.VerifySuperAdminMiddleware) *ProjectHandler {
	return &ProjectHandler{
		projectSvc:       projectSvc,
		logger:           logger,
		verifyJWT:        verifyJWT,
		verifySuperAdmin: verifySuperAdmin,
	}
}

// RegisterRoutes registers project routes on the given group and applies JWT verification middleware.
func (h *ProjectHandler) RegisterRoutes(g *echo.Group) {
	g.Use(echo.MiddlewareFunc(h.verifyJWT))
	g.Use(echo.MiddlewareFunc(h.verifySuperAdmin))
	g.GET("", h.HandleListProjects)
	g.GET("/:id", h.HandleGetProjectByID)
	g.POST("", h.HandleCreateProject)
	g.PUT("/:id", h.HandleUpdateProject)
	g.DELETE("/:id", h.HandleDeleteProject)
}

// List returns a paginated list of projects.
// Query: page (default 1), pageSize (default 10, max 100).
func (h *ProjectHandler) HandleListProjects(c echo.Context) error {
	ctx := c.Request().Context()
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	result, err := h.projectSvc.List(ctx, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list projects", "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, result)
}

// GetByID returns a project by ID.
func (h *ProjectHandler) HandleGetProjectByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, nil))
	}

	project, err := h.projectSvc.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get project", "id", id, "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, project)
}

// Create creates a new project.
func (h *ProjectHandler) HandleCreateProject(c echo.Context) error {
	ctx := c.Request().Context()

	req, err := BindAndValidate[dto.CreateProjectReq](c)
	if err != nil {
		h.logger.Error("Failed to bind create project request", "error", err)
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	project, err := h.projectSvc.Create(ctx, req)
	if err != nil {
		h.logger.Error("Failed to create project", "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, project)
}

// Update updates a project by ID.
func (h *ProjectHandler) HandleUpdateProject(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, nil))
	}

	req, err := BindAndValidate[dto.UpdateProjectReq](c)
	if err != nil {
		h.logger.Error("Failed to bind update project request", "error", err)
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	project, err := h.projectSvc.Update(ctx, id, req)
	if err != nil {
		h.logger.Error("Failed to update project", "id", id, "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, project)
}

// Delete deletes a project by ID.
func (h *ProjectHandler) HandleDeleteProject(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, nil))
	}

	if err := h.projectSvc.Delete(ctx, id); err != nil {
		h.logger.Error("Failed to delete project", "id", id, "error", err)
		return HandleError(c, err)
	}
	return HandleSuccess(c, nil)
}
