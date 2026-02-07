package repository

import (
	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type ISessionRepository interface {
	IRepository[model.Session]
}

type sessionRepository struct {
	Repository[model.Session]
}

func NewSessionRepository(dbClient *gorm.DB) ISessionRepository {
	return &sessionRepository{Repository: Repository[model.Session]{dbClient: dbClient}}
}
