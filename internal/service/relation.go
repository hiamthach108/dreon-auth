package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/model"
	"github.com/hiamthach108/dreon-auth/internal/repository"
	"github.com/hiamthach108/dreon-auth/internal/shared/constant"
	"github.com/hiamthach108/dreon-auth/pkg/cache"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
)

type IRelationSvc interface {
	// Grant and revoke relations
	GrantRelation(ctx context.Context, req dto.GrantRelationReq) (*dto.RelationTupleResp, error)
	RevokeRelation(ctx context.Context, req dto.RevokeRelationReq) error
	BulkGrantRelations(ctx context.Context, req dto.BulkGrantRelationReq) ([]dto.RelationTupleResp, error)
	BulkRevokeRelations(ctx context.Context, req dto.BulkRevokeRelationReq) error

	// Check relations
	CheckRelation(ctx context.Context, req dto.CheckRelationReq) (*dto.CheckRelationResp, error)

	// List and expand relations
	ListRelations(ctx context.Context, req dto.ListRelationsReq) (*dto.PaginationResp[dto.RelationTupleResp], error)
	ExpandRelation(ctx context.Context, req dto.ExpandRelationReq) (*dto.ExpandRelationResp, error)

	// Maintenance
	CleanupExpiredRelations(ctx context.Context) (int64, error)
}

type RelationSvc struct {
	logger    logger.ILogger
	tupleRepo repository.IRelationTupleRepository
	cache     cache.ICache
}

func NewRelationSvc(
	logger logger.ILogger,
	tupleRepo repository.IRelationTupleRepository,
	cache cache.ICache,
) IRelationSvc {
	return &RelationSvc{
		logger:    logger,
		tupleRepo: tupleRepo,
		cache:     cache,
	}
}

// GrantRelation grants a relation by creating a relation tuple
func (s *RelationSvc) GrantRelation(ctx context.Context, req dto.GrantRelationReq) (*dto.RelationTupleResp, error) {
	if err := s.validateRelationRequest(req); err != nil {
		return nil, errorx.Wrap(errorx.ErrInvalidPermission, err)
	}

	existing, err := s.tupleRepo.FindByTuple(
		ctx,
		req.Namespace,
		req.ObjectID,
		req.Relation,
		req.SubjectNamespace,
		req.SubjectObjectID,
		req.SubjectRelation,
	)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	if existing != nil && existing.IsValid() {
		return nil, errorx.New(errorx.ErrPermissionConflict, "Relation already exists and is active")
	}

	tuple := &model.RelationTuple{
		Namespace:        req.Namespace,
		ObjectID:         req.ObjectID,
		Relation:         req.Relation,
		SubjectNamespace: req.SubjectNamespace,
		SubjectObjectID:  req.SubjectObjectID,
		SubjectRelation:  req.SubjectRelation,
		IsActive:         true,
		ExpiresAt:        req.ExpiresAt,
	}

	created, err := s.tupleRepo.Create(ctx, tuple)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrGrantPermission, err)
	}

	s.logger.Info(fmt.Sprintf("Relation granted: %s", created.String()))

	return s.toRelationTupleResp(created), nil
}

// RevokeRelation revokes a relation by deleting the relation tuple
func (s *RelationSvc) RevokeRelation(ctx context.Context, req dto.RevokeRelationReq) error {
	existing, err := s.tupleRepo.FindByTuple(
		ctx,
		req.Namespace,
		req.ObjectID,
		req.Relation,
		req.SubjectNamespace,
		req.SubjectObjectID,
		req.SubjectRelation,
	)
	if err != nil {
		return errorx.Wrap(errorx.ErrInternal, err)
	}

	if existing == nil {
		return errorx.New(errorx.ErrPermissionNotFound, "Relation not found")
	}

	err = s.tupleRepo.DeleteByTuple(
		ctx,
		req.Namespace,
		req.ObjectID,
		req.Relation,
		req.SubjectNamespace,
		req.SubjectObjectID,
		req.SubjectRelation,
	)
	if err != nil {
		return errorx.Wrap(errorx.ErrRevokePermission, err)
	}

	s.logger.Info(fmt.Sprintf("Relation revoked: %s", existing.String()))

	return nil
}

// BulkGrantRelations grants multiple relations in a single transaction
func (s *RelationSvc) BulkGrantRelations(ctx context.Context, req dto.BulkGrantRelationReq) ([]dto.RelationTupleResp, error) {
	results := make([]dto.RelationTupleResp, 0, len(req.Relations))
	tuples := make([]model.RelationTuple, 0, len(req.Relations))

	for _, relReq := range req.Relations {
		if err := s.validateRelationRequest(relReq); err != nil {
			return nil, errorx.Wrap(errorx.ErrInvalidPermission, err)
		}

		tuples = append(tuples, model.RelationTuple{
			Namespace:        relReq.Namespace,
			ObjectID:         relReq.ObjectID,
			Relation:         relReq.Relation,
			SubjectNamespace: relReq.SubjectNamespace,
			SubjectObjectID:  relReq.SubjectObjectID,
			SubjectRelation:  relReq.SubjectRelation,
			IsActive:         true,
			ExpiresAt:        relReq.ExpiresAt,
		})
	}

	if err := s.tupleRepo.BulkCreate(ctx, tuples); err != nil {
		return nil, errorx.Wrap(errorx.ErrGrantPermission, err)
	}

	for i := range tuples {
		results = append(results, *s.toRelationTupleResp(&tuples[i]))
	}

	s.logger.Info(fmt.Sprintf("Bulk granted %d relations", len(tuples)))

	return results, nil
}

// BulkRevokeRelations revokes multiple relations
func (s *RelationSvc) BulkRevokeRelations(ctx context.Context, req dto.BulkRevokeRelationReq) error {
	for _, relReq := range req.Relations {
		if err := s.RevokeRelation(ctx, relReq); err != nil {
			if errorx.GetCode(err) != errorx.ErrPermissionNotFound {
				return err
			}
		}
	}

	s.logger.Info(fmt.Sprintf("Bulk revoked %d relations", len(req.Relations)))

	return nil
}

// CheckRelation checks if a subject has a specific relation on an object
func (s *RelationSvc) CheckRelation(ctx context.Context, req dto.CheckRelationReq) (*dto.CheckRelationResp, error) {

	var allowed bool

	err := s.cache.Get(s.buildCacheKey(&model.RelationTuple{
		Namespace:        req.Namespace,
		ObjectID:         req.ObjectID,
		Relation:         req.Relation,
		SubjectNamespace: req.SubjectNamespace,
		SubjectObjectID:  req.SubjectObjectID,
	}), &allowed)
	if err == nil {
		return &dto.CheckRelationResp{Allowed: allowed}, nil
	} else if err != cache.ErrCacheNil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	allowed, err = s.tupleRepo.CheckPermission(
		ctx,
		req.Namespace,
		req.ObjectID,
		req.Relation,
		req.SubjectNamespace,
		req.SubjectObjectID,
	)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	resp := &dto.CheckRelationResp{
		Allowed: allowed,
	}

	if !allowed {
		resp.Reason = "Relation not found or expired"
	}

	return resp, nil
}

// ListRelations lists relations with optional filters
func (s *RelationSvc) ListRelations(ctx context.Context, req dto.ListRelationsReq) (*dto.PaginationResp[dto.RelationTupleResp], error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	filters := make(map[string]interface{})
	if req.Namespace != "" {
		filters["namespace"] = req.Namespace
	}
	if req.ObjectID != "" {
		filters["object_id"] = req.ObjectID
	}
	if req.Relation != "" {
		filters["relation"] = req.Relation
	}
	if req.SubjectNamespace != "" {
		filters["subject_namespace"] = req.SubjectNamespace
	}
	if req.SubjectObjectID != "" {
		filters["subject_object_id"] = req.SubjectObjectID
	}

	tuples, total, err := s.tupleRepo.ListWithFilters(ctx, filters, pageSize, offset)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	items := make([]dto.RelationTupleResp, 0, len(tuples))
	for i := range tuples {
		items = append(items, *s.toRelationTupleResp(&tuples[i]))
	}

	hasNext := int64(offset+pageSize) < total

	return &dto.PaginationResp[dto.RelationTupleResp]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  hasNext,
	}, nil
}

// ExpandRelation expands a relation to get all subjects with that relation
func (s *RelationSvc) ExpandRelation(ctx context.Context, req dto.ExpandRelationReq) (*dto.ExpandRelationResp, error) {
	tuples, err := s.tupleRepo.ExpandSubjects(ctx, req.Namespace, req.ObjectID, req.Relation)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	subjects := make([]dto.RelationSubjectResp, 0, len(tuples))
	for _, tuple := range tuples {
		subjects = append(subjects, dto.RelationSubjectResp{
			Namespace: tuple.SubjectNamespace,
			ObjectID:  tuple.SubjectObjectID,
			Relation:  tuple.SubjectRelation,
		})
	}

	return &dto.ExpandRelationResp{
		Subjects: subjects,
		Count:    len(subjects),
	}, nil
}

// CleanupExpiredRelations removes expired relation tuples
func (s *RelationSvc) CleanupExpiredRelations(ctx context.Context) (int64, error) {
	count, err := s.tupleRepo.CleanupExpired(ctx)
	if err != nil {
		return 0, errorx.Wrap(errorx.ErrInternal, err)
	}

	if count > 0 {
		s.logger.Info(fmt.Sprintf("Cleaned up %d expired relations", count))
	}

	return count, nil
}

// validateRelationRequest validates the relation request
func (s *RelationSvc) validateRelationRequest(req dto.GrantRelationReq) error {
	if req.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if req.ObjectID == "" {
		return fmt.Errorf("objectId is required")
	}
	if req.Relation == "" {
		return fmt.Errorf("relation is required")
	}
	if req.SubjectNamespace == "" {
		return fmt.Errorf("subjectNamespace is required")
	}
	if req.SubjectObjectID == "" {
		return fmt.Errorf("subjectObjectId is required")
	}
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expiresAt must be in the future")
	}
	return nil
}

// toRelationTupleResp converts a relation tuple to a response
func (s *RelationSvc) toRelationTupleResp(tuple *model.RelationTuple) *dto.RelationTupleResp {
	return &dto.RelationTupleResp{
		ID:               tuple.ID,
		Namespace:        tuple.Namespace,
		ObjectID:         tuple.ObjectID,
		Relation:         tuple.Relation,
		SubjectNamespace: tuple.SubjectNamespace,
		SubjectObjectID:  tuple.SubjectObjectID,
		SubjectRelation:  tuple.SubjectRelation,
		IsActive:         tuple.IsActive,
		ExpiresAt:        tuple.ExpiresAt,
		CreatedAt:        tuple.CreatedAt,
		UpdatedAt:        tuple.UpdatedAt,
	}
}

// buildCacheKey builds a cache key for a relation tuple
func (s *RelationSvc) buildCacheKey(tuple *model.RelationTuple) string {
	return constant.CacheKeyPrefixRelationTuple + tuple.String()
}
