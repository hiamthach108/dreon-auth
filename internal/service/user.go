package service

import (
	"context"

	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/repository"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// IUserSvc defines the contract for user operations.
type IUserSvc interface {
	Create(ctx context.Context, req dto.CreateUserReq) (*dto.UserDto, error)
	GetByID(ctx context.Context, id string) (*dto.UserDto, error)
	List(ctx context.Context, page, pageSize int) (*dto.PaginationResp[dto.UserDto], error)
	Update(ctx context.Context, id string, req dto.UpdateUserReq) (*dto.UserDto, error)
	Delete(ctx context.Context, id string) error
}

// UserSvc implements IUserSvc.
type UserSvc struct {
	logger logger.ILogger
	repo   repository.IUserRepository
}

// NewUserSvc creates a new user service.
func NewUserSvc(logger logger.ILogger, repo repository.IUserRepository) IUserSvc {
	return &UserSvc{
		logger: logger,
		repo:   repo,
	}
}

// Create creates a new user with hashed password.
func (s *UserSvc) Create(ctx context.Context, req dto.CreateUserReq) (*dto.UserDto, error) {
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("[UserSvc] failed to check email", "email", req.Email, "error", err)
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if existing != nil {
		return nil, errorx.New(errorx.ErrUserConflict, "email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("[UserSvc] failed to hash password", "error", err)
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	model := req.ToModel(string(hashed))
	created, err := s.repo.Create(ctx, model)
	if err != nil {
		s.logger.Error("[UserSvc] failed to create user", "email", req.Email, "error", err)
		return nil, errorx.Wrap(errorx.ErrCreateUser, err)
	}

	var resp dto.UserDto
	resp.FromModel(created)
	return &resp, nil
}

// GetByID returns a user by ID.
func (s *UserSvc) GetByID(ctx context.Context, id string) (*dto.UserDto, error) {
	u := s.repo.FindOneById(ctx, id)
	if u == nil {
		return nil, errorx.Wrap(errorx.ErrUserNotFound, nil)
	}
	var resp dto.UserDto
	resp.FromModel(u)
	return &resp, nil
}

// List returns a paginated list of users.
func (s *UserSvc) List(ctx context.Context, page, pageSize int) (*dto.PaginationResp[dto.UserDto], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	users, total, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		s.logger.Error("[UserSvc] failed to list users", "error", err)
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	items := make([]dto.UserDto, 0, len(users))
	for i := range users {
		var d dto.UserDto
		d.FromModel(&users[i])
		items = append(items, d)
	}

	hasNext := int64(offset+len(users)) < total
	return &dto.PaginationResp[dto.UserDto]{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  hasNext,
		Items:    items,
	}, nil
}

// Update updates a user by ID (partial update).
func (s *UserSvc) Update(ctx context.Context, id string, req dto.UpdateUserReq) (*dto.UserDto, error) {
	u := s.repo.FindOneById(ctx, id)
	if u == nil {
		return nil, errorx.Wrap(errorx.ErrUserNotFound, nil)
	}

	updated, fields := req.ToModelAndFields()
	if len(fields) == 0 {
		var resp dto.UserDto
		resp.FromModel(u)
		return &resp, nil
	}

	// Hash password if it's being updated
	for _, f := range fields {
		if f == "password" {
			hashed, err := bcrypt.GenerateFromPassword([]byte(updated.Password), bcrypt.DefaultCost)
			if err != nil {
				s.logger.Error("[UserSvc] failed to hash password", "error", err)
				return nil, errorx.Wrap(errorx.ErrInternal, err)
			}
			updated.Password = string(hashed)
			break
		}
	}

	if err := s.repo.Update(ctx, id, *updated, fields...); err != nil {
		s.logger.Error("[UserSvc] failed to update user", "id", id, "error", err)
		return nil, errorx.Wrap(errorx.ErrUpdateUser, err)
	}

	updatedUser := s.repo.FindOneById(ctx, id)
	if updatedUser == nil {
		var resp dto.UserDto
		resp.FromModel(u)
		return &resp, nil
	}
	var resp dto.UserDto
	resp.FromModel(updatedUser)
	return &resp, nil
}

// Delete deletes a user by ID.
func (s *UserSvc) Delete(ctx context.Context, id string) error {
	u := s.repo.FindOneById(ctx, id)
	if u == nil {
		return errorx.Wrap(errorx.ErrUserNotFound, nil)
	}
	if err := s.repo.DeleteById(ctx, id); err != nil {
		s.logger.Error("[UserSvc] failed to delete user", "id", id, "error", err)
		return errorx.Wrap(errorx.ErrInternal, err)
	}
	return nil
}
