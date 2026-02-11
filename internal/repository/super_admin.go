package repository

import (
	"context"

	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type ISuperAdminRepository interface {
	IRepository[model.SuperAdmin]
	FindByEmail(ctx context.Context, email string) (*model.SuperAdmin, error)
}

type superAdminRepository struct {
	Repository[model.SuperAdmin]
}

func NewSuperAdminRepository(dbClient *gorm.DB) ISuperAdminRepository {
	return &superAdminRepository{Repository: Repository[model.SuperAdmin]{dbClient: dbClient}}
}

func (r *superAdminRepository) FindByEmail(ctx context.Context, email string) (*model.SuperAdmin, error) {
	var result model.SuperAdmin
	err := r.dbClient.WithContext(ctx).
		Where(&model.SuperAdmin{Email: email, IsActive: true}).
		First(&result).
		Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}
