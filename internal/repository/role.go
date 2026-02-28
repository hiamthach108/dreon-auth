package repository

import (
	"context"

	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type IRoleRepository interface {
	IRepository[model.Role]
	
	FindByCode(ctx context.Context, code string) (*model.Role, error)
	FindByProjectID(ctx context.Context, projectID *string, limit, offset int) ([]model.Role, int64, error)
	FindSystemRoles(ctx context.Context, limit, offset int) ([]model.Role, int64, error)
	SearchRoles(ctx context.Context, search string, projectID *string, isActive *bool, limit, offset int) ([]model.Role, int64, error)
	IsSystemRole(ctx context.Context, roleID string) (bool, error)
}

type roleRepository struct {
	Repository[model.Role]
}

func NewRoleRepository(dbClient *gorm.DB) IRoleRepository {
	return &roleRepository{Repository: Repository[model.Role]{dbClient: dbClient}}
}

// FindByCode finds a role by its code
func (r *roleRepository) FindByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	if err := r.dbClient.WithContext(ctx).Where("code = ?", code).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// FindByProjectID finds all roles for a specific project
func (r *roleRepository) FindByProjectID(ctx context.Context, projectID *string, limit, offset int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	
	query := r.dbClient.WithContext(ctx).Model(&model.Role{})
	
	if projectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		query = query.Where("project_id = ?", *projectID)
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	
	return roles, total, nil
}

// FindSystemRoles finds all system roles (ProjectID = "system")
func (r *roleRepository) FindSystemRoles(ctx context.Context, limit, offset int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	
	query := r.dbClient.WithContext(ctx).Model(&model.Role{}).Where("project_id = ?", "system")
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	
	return roles, total, nil
}

// SearchRoles searches roles with filters
func (r *roleRepository) SearchRoles(ctx context.Context, search string, projectID *string, isActive *bool, limit, offset int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	
	query := r.dbClient.WithContext(ctx).Model(&model.Role{})
	
	if search != "" {
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	if projectID != nil {
		if *projectID == "system" {
			query = query.Where("project_id = ?", "system")
		} else {
			query = query.Where("project_id = ?", *projectID)
		}
	}
	
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	
	return roles, total, nil
}

// IsSystemRole checks if a role is a system role
func (r *roleRepository) IsSystemRole(ctx context.Context, roleID string) (bool, error) {
	var count int64
	err := r.dbClient.WithContext(ctx).Model(&model.Role{}).
		Where("id = ? AND project_id = ?", roleID, "system").
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}
