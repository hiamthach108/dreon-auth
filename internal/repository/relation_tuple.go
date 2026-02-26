package repository

import (
	"context"
	"time"

	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type IRelationTupleRepository interface {
	IRepository[model.RelationTuple]
	
	// Permission-specific queries
	FindByTuple(ctx context.Context, namespace, objectID, relation, subjectNamespace, subjectObjectID, subjectRelation string) (*model.RelationTuple, error)
	CheckPermission(ctx context.Context, namespace, objectID, relation, subjectNamespace, subjectObjectID string) (bool, error)
	ListByObject(ctx context.Context, namespace, objectID string, limit, offset int) ([]model.RelationTuple, int64, error)
	ListBySubject(ctx context.Context, subjectNamespace, subjectObjectID string, limit, offset int) ([]model.RelationTuple, int64, error)
	ListByRelation(ctx context.Context, namespace, relation string, limit, offset int) ([]model.RelationTuple, int64, error)
	ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.RelationTuple, int64, error)
	ExpandSubjects(ctx context.Context, namespace, objectID, relation string) ([]model.RelationTuple, error)
	DeleteByTuple(ctx context.Context, namespace, objectID, relation, subjectNamespace, subjectObjectID, subjectRelation string) error
	CleanupExpired(ctx context.Context) (int64, error)
}

type relationTupleRepository struct {
	Repository[model.RelationTuple]
}

func NewRelationTupleRepository(dbClient *gorm.DB) IRelationTupleRepository {
	return &relationTupleRepository{Repository: Repository[model.RelationTuple]{dbClient: dbClient}}
}

// FindByTuple finds a specific relation tuple
func (r *relationTupleRepository) FindByTuple(ctx context.Context, namespace, objectID, relation, subjectNamespace, subjectObjectID, subjectRelation string) (*model.RelationTuple, error) {
	var tuple model.RelationTuple
	query := r.dbClient.WithContext(ctx).Where(
		"namespace = ? AND object_id = ? AND relation = ? AND subject_namespace = ? AND subject_object_id = ?",
		namespace, objectID, relation, subjectNamespace, subjectObjectID,
	)
	
	if subjectRelation != "" {
		query = query.Where("subject_relation = ?", subjectRelation)
	} else {
		query = query.Where("subject_relation IS NULL OR subject_relation = ''")
	}
	
	if err := query.First(&tuple).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tuple, nil
}

// CheckPermission checks if a permission exists and is valid
func (r *relationTupleRepository) CheckPermission(ctx context.Context, namespace, objectID, relation, subjectNamespace, subjectObjectID string) (bool, error) {
	var count int64
	err := r.dbClient.WithContext(ctx).Model(&model.RelationTuple{}).Where(
		"namespace = ? AND object_id = ? AND relation = ? AND subject_namespace = ? AND subject_object_id = ? AND is_active = ?",
		namespace, objectID, relation, subjectNamespace, subjectObjectID, true,
	).Where("expires_at IS NULL OR expires_at > ?", time.Now()).Count(&count).Error
	
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByObject lists all permissions for a specific object
func (r *relationTupleRepository) ListByObject(ctx context.Context, namespace, objectID string, limit, offset int) ([]model.RelationTuple, int64, error) {
	var tuples []model.RelationTuple
	var total int64
	
	query := r.dbClient.WithContext(ctx).Model(&model.RelationTuple{}).Where("namespace = ? AND object_id = ?", namespace, objectID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Limit(limit).Offset(offset).Find(&tuples).Error; err != nil {
		return nil, 0, err
	}
	
	return tuples, total, nil
}

// ListBySubject lists all permissions for a specific subject
func (r *relationTupleRepository) ListBySubject(ctx context.Context, subjectNamespace, subjectObjectID string, limit, offset int) ([]model.RelationTuple, int64, error) {
	var tuples []model.RelationTuple
	var total int64
	
	query := r.dbClient.WithContext(ctx).Model(&model.RelationTuple{}).Where("subject_namespace = ? AND subject_object_id = ?", subjectNamespace, subjectObjectID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Limit(limit).Offset(offset).Find(&tuples).Error; err != nil {
		return nil, 0, err
	}
	
	return tuples, total, nil
}

// ListByRelation lists all permissions for a specific relation
func (r *relationTupleRepository) ListByRelation(ctx context.Context, namespace, relation string, limit, offset int) ([]model.RelationTuple, int64, error) {
	var tuples []model.RelationTuple
	var total int64
	
	query := r.dbClient.WithContext(ctx).Model(&model.RelationTuple{}).Where("namespace = ? AND relation = ?", namespace, relation)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Limit(limit).Offset(offset).Find(&tuples).Error; err != nil {
		return nil, 0, err
	}
	
	return tuples, total, nil
}

// ListWithFilters lists permissions with dynamic filters
func (r *relationTupleRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.RelationTuple, int64, error) {
	var tuples []model.RelationTuple
	var total int64
	
	query := r.dbClient.WithContext(ctx).Model(&model.RelationTuple{})
	
	for key, value := range filters {
		if value != "" && value != nil {
			query = query.Where(key+" = ?", value)
		}
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Limit(limit).Offset(offset).Find(&tuples).Error; err != nil {
		return nil, 0, err
	}
	
	return tuples, total, nil
}

// ExpandSubjects gets all subjects with a specific permission on an object
func (r *relationTupleRepository) ExpandSubjects(ctx context.Context, namespace, objectID, relation string) ([]model.RelationTuple, error) {
	var tuples []model.RelationTuple
	err := r.dbClient.WithContext(ctx).Where(
		"namespace = ? AND object_id = ? AND relation = ? AND is_active = ?",
		namespace, objectID, relation, true,
	).Where("expires_at IS NULL OR expires_at > ?", time.Now()).Find(&tuples).Error
	
	if err != nil {
		return nil, err
	}
	return tuples, nil
}

// DeleteByTuple deletes a specific relation tuple
func (r *relationTupleRepository) DeleteByTuple(ctx context.Context, namespace, objectID, relation, subjectNamespace, subjectObjectID, subjectRelation string) error {
	query := r.dbClient.WithContext(ctx).Where(
		"namespace = ? AND object_id = ? AND relation = ? AND subject_namespace = ? AND subject_object_id = ?",
		namespace, objectID, relation, subjectNamespace, subjectObjectID,
	)
	
	if subjectRelation != "" {
		query = query.Where("subject_relation = ?", subjectRelation)
	} else {
		query = query.Where("subject_relation IS NULL OR subject_relation = ''")
	}
	
	return query.Delete(&model.RelationTuple{}).Error
}

// CleanupExpired removes expired relation tuples
func (r *relationTupleRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result := r.dbClient.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).Delete(&model.RelationTuple{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
