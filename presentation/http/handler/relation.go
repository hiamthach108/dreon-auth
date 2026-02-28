package handler

import (
	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/service"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"github.com/hiamthach108/dreon-auth/presentation/http/middleware"
	"github.com/labstack/echo/v4"
)

type RelationHandler struct {
	relationSvc service.IRelationSvc
	logger      logger.ILogger
	verifyJWT   middleware.VerifyJWTMiddleware
}

func NewRelationHandler(
	relationSvc service.IRelationSvc,
	logger logger.ILogger,
	verifyJWT middleware.VerifyJWTMiddleware,
) *RelationHandler {
	return &RelationHandler{
		relationSvc: relationSvc,
		logger:      logger,
		verifyJWT:   verifyJWT,
	}
}

func (h *RelationHandler) RegisterRoutes(g *echo.Group) {
	g.Use(echo.MiddlewareFunc(h.verifyJWT))

	g.POST("/grant", h.HandleGrantRelation)
	g.POST("/revoke", h.HandleRevokeRelation)
	g.POST("/bulk-grant", h.HandleBulkGrantRelations)
	g.POST("/bulk-revoke", h.HandleBulkRevokeRelations)
	g.POST("/check", h.HandleCheckRelation)
	g.GET("/list", h.HandleListRelations)
	g.POST("/expand", h.HandleExpandRelation)
	g.DELETE("/cleanup", h.HandleCleanupExpired)
}

// HandleGrantRelation grants a relation to a subject
func (h *RelationHandler) HandleGrantRelation(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := BindAndValidate[dto.GrantRelationReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.relationSvc.GrantRelation(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleRevokeRelation revokes a relation from a subject
func (h *RelationHandler) HandleRevokeRelation(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := BindAndValidate[dto.RevokeRelationReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	if err := h.relationSvc.RevokeRelation(ctx, req); err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, map[string]string{"message": "Relation revoked successfully"})
}

// HandleBulkGrantRelations grants multiple relations
func (h *RelationHandler) HandleBulkGrantRelations(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := BindAndValidate[dto.BulkGrantRelationReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.relationSvc.BulkGrantRelations(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleBulkRevokeRelations revokes multiple relations
func (h *RelationHandler) HandleBulkRevokeRelations(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := BindAndValidate[dto.BulkRevokeRelationReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	if err := h.relationSvc.BulkRevokeRelations(ctx, req); err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, map[string]string{"message": "Relations revoked successfully"})
}

// HandleCheckRelation checks if a subject has a specific relation
func (h *RelationHandler) HandleCheckRelation(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := BindAndValidate[dto.CheckRelationReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.relationSvc.CheckRelation(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleListRelations lists relations with optional filters
func (h *RelationHandler) HandleListRelations(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := BindAndValidate[dto.ListRelationsReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.relationSvc.ListRelations(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleExpandRelation expands a relation to get all subjects
func (h *RelationHandler) HandleExpandRelation(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := BindAndValidate[dto.ExpandRelationReq](c)
	if err != nil {
		return HandleError(c, errorx.Wrap(errorx.ErrBadRequest, err))
	}

	result, err := h.relationSvc.ExpandRelation(ctx, req)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, result)
}

// HandleCleanupExpired removes expired relations
func (h *RelationHandler) HandleCleanupExpired(c echo.Context) error {
	ctx := c.Request().Context()

	count, err := h.relationSvc.CleanupExpiredRelations(ctx)
	if err != nil {
		return HandleError(c, err)
	}

	return HandleSuccess(c, echo.Map{
		"message": "Expired relations cleaned up successfully",
		"count":   count,
	})
}
