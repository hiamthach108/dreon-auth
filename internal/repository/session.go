package repository

import (
	"context"

	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

type ISessionRepository interface {
	IRepository[model.Session]
	FindByRefreshToken(ctx context.Context, refreshToken string) *model.Session
}

type sessionRepository struct {
	Repository[model.Session]
}

func NewSessionRepository(dbClient *gorm.DB) ISessionRepository {
	return &sessionRepository{Repository: Repository[model.Session]{dbClient: dbClient}}
}

func (r *sessionRepository) FindByRefreshToken(ctx context.Context, refreshToken string) *model.Session {
	var result model.Session
	err := r.dbClient.WithContext(ctx).Where(&model.Session{
		RefreshToken: refreshToken,
	}).First(&result).Error
	if err != nil {
		return nil
	}
	return &result
}
