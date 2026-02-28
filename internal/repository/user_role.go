package repository

import (
	"context"

	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type IUserRoleRepository interface {
	IRepository[model.UserRole]
	
	FindByUserID(ctx context.Context, userID string) ([]model.UserRole, error)
	FindByUserIDAndProjectID(ctx context.Context, userID string, projectID *string) ([]model.UserRole, error)
	FindByUserIDAndRoleID(ctx context.Context, userID, roleID string, projectID *string) (*model.UserRole, error)
	DeleteByUserIDAndRoleID(ctx context.Context, userID, roleID string, projectID *string) error
	FindWithRole(ctx context.Context, userID string, projectID *string) ([]model.UserRole, error)
}

type userRoleRepository struct {
	Repository[model.UserRole]
}

func NewUserRoleRepository(dbClient *gorm.DB) IUserRoleRepository {
	return &userRoleRepository{Repository: Repository[model.UserRole]{dbClient: dbClient}}
}

// FindByUserID finds all role assignments for a user
func (r *userRoleRepository) FindByUserID(ctx context.Context, userID string) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	if err := r.dbClient.WithContext(ctx).Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}
	return userRoles, nil
}

// FindByUserIDAndProjectID finds role assignments for a user in a specific project
func (r *userRoleRepository) FindByUserIDAndProjectID(ctx context.Context, userID string, projectID *string) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	
	query := r.dbClient.WithContext(ctx).Where("user_id = ?", userID)
	
	if projectID == nil {
		query = query.Where("project_id IS NULL")
	} else if *projectID == "system" {
		query = query.Where("project_id = ?", "system")
	} else {
		query = query.Where("project_id = ?", *projectID)
	}
	
	if err := query.Find(&userRoles).Error; err != nil {
		return nil, err
	}
	
	return userRoles, nil
}

// FindByUserIDAndRoleID finds a specific user role assignment
func (r *userRoleRepository) FindByUserIDAndRoleID(ctx context.Context, userID, roleID string, projectID *string) (*model.UserRole, error) {
	var userRole model.UserRole
	
	query := r.dbClient.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID)
	
	if projectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		query = query.Where("project_id = ?", *projectID)
	}
	
	if err := query.First(&userRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &userRole, nil
}

// DeleteByUserIDAndRoleID deletes a specific user role assignment
func (r *userRoleRepository) DeleteByUserIDAndRoleID(ctx context.Context, userID, roleID string, projectID *string) error {
	query := r.dbClient.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID)
	
	if projectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		query = query.Where("project_id = ?", *projectID)
	}
	
	return query.Delete(&model.UserRole{}).Error
}

// FindWithRole finds user roles with preloaded role information
func (r *userRoleRepository) FindWithRole(ctx context.Context, userID string, projectID *string) ([]model.UserRole, error) {
	var userRoles []model.UserRole
	
	query := r.dbClient.WithContext(ctx).Preload("Role").Where("user_id = ?", userID)
	
	if projectID != nil {
		if *projectID == "system" {
			query = query.Where("project_id = ?", "system")
		} else {
			query = query.Where("project_id = ?", *projectID)
		}
	}
	
	if err := query.Find(&userRoles).Error; err != nil {
		return nil, err
	}
	
	return userRoles, nil
}
