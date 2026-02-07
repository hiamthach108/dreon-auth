package repository

import (
	"context"

	"github.com/hiamthach108/dreon-auth/internal/model"
	"gorm.io/gorm"
)

// IUserRepository defines the contract for user persistence.
type IUserRepository interface {
	IRepository[model.User]
	// List returns users with pagination. total is the total count before pagination.
	List(ctx context.Context, offset, limit int) ([]model.User, int64, error)
	// FindByEmail returns a user by email, or nil if not found.
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type userRepository struct {
	Repository[model.User]
}

// NewUserRepository creates a new user repository.
func NewUserRepository(dbClient *gorm.DB) IUserRepository {
	return &userRepository{
		Repository: Repository[model.User]{dbClient: dbClient},
	}
}

// List returns a paginated list of users and total count.
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var total int64
	if err := r.dbClient.WithContext(ctx).Model(new(model.User)).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var results []model.User
	q := r.dbClient.WithContext(ctx).Offset(offset).Limit(limit)
	if err := q.Find(&results).Error; err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// FindByEmail returns one user by email.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var result model.User
	if err := r.dbClient.WithContext(ctx).Where("email = ?", email).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}
