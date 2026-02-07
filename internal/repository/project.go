package repository

import (
	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type IProjectRepository interface {
	IRepository[model.Project]
}

type projectRepository struct {
	Repository[model.Project]
}

func NewProjectRepository(dbClient *gorm.DB) IProjectRepository {
	return &projectRepository{Repository: Repository[model.Project]{dbClient: dbClient}}
}
