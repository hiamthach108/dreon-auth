package repository

import (
	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type ISuperAdminRepository interface {
	IRepository[model.SuperAdmin]
}

type superAdminRepository struct {
	Repository[model.SuperAdmin]
}

func NewSuperAdminRepository(dbClient *gorm.DB) ISuperAdminRepository {
	return &superAdminRepository{Repository: Repository[model.SuperAdmin]{dbClient: dbClient}}
}
