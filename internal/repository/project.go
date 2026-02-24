package repository

import (
	"context"

	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type IProjectRepository interface {
	IRepository[model.Project]
	// List returns projects with pagination. total is the total count before pagination.
	List(ctx context.Context, offset, limit int) ([]model.Project, int64, error)
	// FindByCode returns a project by code, or nil if not found.
	FindByCode(ctx context.Context, code string) (*model.Project, error)
}

type projectRepository struct {
	Repository[model.Project]
}

func NewProjectRepository(dbClient *gorm.DB) IProjectRepository {
	return &projectRepository{Repository: Repository[model.Project]{dbClient: dbClient}}
}

// List returns a paginated list of projects and total count.
func (r *projectRepository) List(ctx context.Context, offset, limit int) ([]model.Project, int64, error) {
	var total int64
	if err := r.dbClient.WithContext(ctx).Model(new(model.Project)).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var results []model.Project
	q := r.dbClient.WithContext(ctx).Offset(offset).Limit(limit)
	if err := q.Find(&results).Error; err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// FindByCode returns one project by code.
func (r *projectRepository) FindByCode(ctx context.Context, code string) (*model.Project, error) {
	var result model.Project
	if err := r.dbClient.WithContext(ctx).Where("code = ?", code).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}
