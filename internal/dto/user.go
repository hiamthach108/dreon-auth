package dto

import (
	"time"

	"github.com/hiamthach108/dreon-auth/internal/model"
)

// CreateUserReq is the request body for creating a user.
type CreateUserReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateUserReq is the request body for updating a user (partial update).
type UpdateUserReq struct {
	Username *string `json:"username"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"omitempty,min=8"`
}

// UserDto is the response DTO for user (password omitted).
type UserDto struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FromModel maps a model.User to UserDto (excludes password).
func (d *UserDto) FromModel(m *model.User) {
	if m == nil {
		return
	}
	d.ID = m.ID
	d.Username = m.Username
	d.Email = m.Email
	d.CreatedAt = m.CreatedAt
	d.UpdatedAt = m.UpdatedAt
}

// ToModel maps CreateUserReq to model.User (caller sets hashed password).
func (r *CreateUserReq) ToModel(hashedPassword string) *model.User {
	return &model.User{
		Username: r.Username,
		Email:    r.Email,
		Password: hashedPassword,
	}
}

// ToModelAndFields returns the model and list of fields to update for UpdateUserReq.
func (r *UpdateUserReq) ToModelAndFields() (u *model.User, fields []string) {
	u = &model.User{}
	if r.Username != nil {
		u.Username = *r.Username
		fields = append(fields, "username")
	}
	if r.Email != nil {
		u.Email = *r.Email
		fields = append(fields, "email")
	}
	if r.Password != nil {
		u.Password = *r.Password
		fields = append(fields, "password")
	}
	return u, fields
}
