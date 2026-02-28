package dto

import (
	"time"

	"github.com/hiamthach108/dreon-auth/internal/model"
)

// CreateRoleReq represents a request to create a role
type CreateRoleReq struct {
	Code        string   `json:"code" validate:"required,min=2,max=255"`
	Name        string   `json:"name" validate:"required,min=2,max=255"`
	Description string   `json:"description"`
	ProjectID   *string  `json:"projectId"` // null for system roles
	Permissions []string `json:"permissions"`
}

// UpdateRoleReq represents a request to update a role
type UpdateRoleReq struct {
	Name        string   `json:"name" validate:"required,min=2,max=255"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsActive    *bool    `json:"isActive"`
}

// RoleResp represents a role response
type RoleResp struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"isActive"`
	ProjectID   *string   `json:"projectId"`
	IsSystem    bool      `json:"isSystem"` // true if ProjectID is "system"
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (r *RoleResp) FromModel(m *model.Role) {
	if m == nil {
		return
	}
	r.ID = m.ID
	r.Code = m.Code
	r.Name = m.Name
	r.Description = m.Description
	r.IsActive = m.IsActive
	r.ProjectID = m.ProjectID
	r.CreatedAt = m.CreatedAt
	r.UpdatedAt = m.UpdatedAt
	r.Permissions = model.PermissionsFromJSON(m.Permissions)
	r.IsSystem = m.ProjectID != nil && *m.ProjectID == "system"
}

// RoleRespFromModel returns a RoleResp from a model.Role.
func RoleRespFromModel(m *model.Role) *RoleResp {
	if m == nil {
		return nil
	}
	r := &RoleResp{}
	r.FromModel(m)
	return r
}

func (r *RoleResp) ToModel() *model.Role {
	if r == nil {
		return nil
	}
	return &model.Role{
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		IsActive:    r.IsActive,
		ProjectID:   r.ProjectID,
		Permissions: model.PermissionsToJSON(r.Permissions),
		BaseModel: model.BaseModel{
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			ID:        r.ID,
		},
	}
}

// ToModel returns a model.Role for create (no ID; IsActive true). BaseModel is filled by repo/BeforeCreate.
func (r *CreateRoleReq) ToModel() *model.Role {
	if r == nil {
		return nil
	}
	return &model.Role{
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		ProjectID:   r.ProjectID,
		Permissions: model.PermissionsToJSON(r.Permissions),
		IsActive:    true,
	}
}

// ApplyTo updates the role model with request fields (name, description, permissions, is_active if set).
func (r *UpdateRoleReq) ApplyTo(m *model.Role) {
	if r == nil || m == nil {
		return
	}
	m.Name = r.Name
	m.Description = r.Description
	m.Permissions = model.PermissionsToJSON(r.Permissions)
	if r.IsActive != nil {
		m.IsActive = *r.IsActive
	}
}

// ListRolesReq represents a request to list roles
type ListRolesReq struct {
	ProjectID *string `form:"projectId" json:"projectId"` // filter by project, "system" for system roles
	IsActive  *bool   `form:"isActive" json:"isActive"`   // filter by active status
	Search    string  `form:"search" json:"search"`       // search by code or name
	PaginationReq
}

// AssignRoleToUserReq represents a request to assign a role to a user
type AssignRoleToUserReq struct {
	UserID    string  `json:"userId" validate:"required"`
	RoleID    string  `json:"roleId" validate:"required"`
	ProjectID *string `json:"projectId"` // null for system role assignment
}

// RemoveRoleFromUserReq represents a request to remove a role from a user
type RemoveRoleFromUserReq struct {
	UserID    string  `json:"userId" validate:"required"`
	RoleID    string  `json:"roleId" validate:"required"`
	ProjectID *string `json:"projectId"`
}

// UserRoleResp represents a user role assignment response
type UserRoleResp struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	RoleID    string    `json:"roleId"`
	ProjectID *string   `json:"projectId"`
	Role      *RoleResp `json:"role,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetUserRolesReq represents a request to get user roles
type GetUserRolesReq struct {
	UserID    string  `form:"userId" json:"userId" validate:"required"`
	ProjectID *string `form:"projectId" json:"projectId"` // filter by project
}

// UserRoleRespFromModel returns a UserRoleResp from model UserRole and optional Role.
func UserRoleRespFromModel(userRole *model.UserRole, role *model.Role) *UserRoleResp {
	if userRole == nil {
		return nil
	}
	r := &UserRoleResp{
		ID:        userRole.ID,
		UserID:    userRole.UserID,
		RoleID:    userRole.RoleID,
		ProjectID: userRole.ProjectID,
		CreatedAt: userRole.CreatedAt,
	}
	if role != nil {
		r.Role = RoleRespFromModel(role)
	}
	return r
}

// UserPermission
type UserPermissions map[string]bool
