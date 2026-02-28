package service

import (
	"context"
	"fmt"

	"github.com/hiamthach108/dreon-auth/internal/dto"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/model"
	"github.com/hiamthach108/dreon-auth/internal/repository"
	"github.com/hiamthach108/dreon-auth/internal/shared/constant"
	"github.com/hiamthach108/dreon-auth/internal/shared/permission"
	"github.com/hiamthach108/dreon-auth/pkg/cache"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
)

type IRoleSvc interface {
	// Role CRUD
	CreateRole(ctx context.Context, req dto.CreateRoleReq, isSuperAdmin bool) (*dto.RoleResp, error)
	GetRole(ctx context.Context, roleID string) (*dto.RoleResp, error)
	UpdateRole(ctx context.Context, roleID string, req dto.UpdateRoleReq, isSuperAdmin bool) (*dto.RoleResp, error)
	DeleteRole(ctx context.Context, roleID string, isSuperAdmin bool) error
	ListRoles(ctx context.Context, req dto.ListRolesReq) (*dto.PaginationResp[dto.RoleResp], error)

	// User role assignment
	AssignRoleToUser(ctx context.Context, req dto.AssignRoleToUserReq, isSuperAdmin bool) (*dto.UserRoleResp, error)
	RemoveRoleFromUser(ctx context.Context, req dto.RemoveRoleFromUserReq, isSuperAdmin bool) error
	GetUserRoles(ctx context.Context, req dto.GetUserRolesReq) ([]dto.UserRoleResp, error)
	GetUserPermissions(ctx context.Context, userID string) (dto.UserPermissions, error)
}

type RoleSvc struct {
	logger             logger.ILogger
	roleRepo           repository.IRoleRepository
	userRoleRepo       repository.IUserRoleRepository
	userRepo           repository.IUserRepository
	permissionRegistry *permission.Registry
	cache              cache.ICache
}

func NewRoleSvc(
	logger logger.ILogger,
	roleRepo repository.IRoleRepository,
	userRoleRepo repository.IUserRoleRepository,
	userRepo repository.IUserRepository,
	permissionRegistry *permission.Registry,
	cache cache.ICache,
) IRoleSvc {
	return &RoleSvc{
		logger:             logger,
		roleRepo:           roleRepo,
		userRoleRepo:       userRoleRepo,
		userRepo:           userRepo,
		permissionRegistry: permissionRegistry,
		cache:              cache,
	}
}

// CreateRole creates a new role
func (s *RoleSvc) CreateRole(ctx context.Context, req dto.CreateRoleReq, isSuperAdmin bool) (*dto.RoleResp, error) {
	// Validate system role creation
	if req.ProjectID != nil && *req.ProjectID == constant.SystemProjectID && !isSuperAdmin {
		return nil, errorx.New(errorx.ErrSystemRoleProtected, "Only super admins can create system roles")
	}

	if s.permissionRegistry != nil {
		if err := s.permissionRegistry.ValidateCodes(req.Permissions); err != nil {
			return nil, errorx.New(errorx.ErrInvalidPermission, err.Error())
		}
	}

	// Check if role code already exists
	existing, err := s.roleRepo.FindByCode(ctx, req.Code)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if existing != nil {
		return nil, errorx.New(errorx.ErrRoleConflict, "Role with this code already exists")
	}

	role := req.ToModel()
	created, err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrCreateRole, err)
	}

	s.logger.Info(fmt.Sprintf("Role created: %s (code: %s)", created.Name, created.Code))
	return dto.RoleRespFromModel(created), nil
}

// GetRole retrieves a role by ID
func (s *RoleSvc) GetRole(ctx context.Context, roleID string) (*dto.RoleResp, error) {
	role := s.roleRepo.FindOneById(ctx, roleID)
	if role == nil {
		return nil, errorx.New(errorx.ErrRoleNotFound, "Role not found")
	}
	return dto.RoleRespFromModel(role), nil
}

// UpdateRole updates an existing role
func (s *RoleSvc) UpdateRole(ctx context.Context, roleID string, req dto.UpdateRoleReq, isSuperAdmin bool) (*dto.RoleResp, error) {
	// Check if role exists
	role := s.roleRepo.FindOneById(ctx, roleID)
	if role == nil {
		return nil, errorx.New(errorx.ErrRoleNotFound, "Role not found")
	}

	// Validate system role update
	if role.ProjectID != nil && *role.ProjectID == constant.SystemProjectID && !isSuperAdmin {
		return nil, errorx.New(errorx.ErrSystemRoleProtected, "Only super admins can update system roles")
	}

	if s.permissionRegistry != nil {
		if err := s.permissionRegistry.ValidateCodes(req.Permissions); err != nil {
			return nil, errorx.New(errorx.ErrInvalidPermission, err.Error())
		}
	}

	updateFields := []string{"name", "description", "permissions", "updated_at"}
	req.ApplyTo(role)
	if req.IsActive != nil {
		updateFields = append(updateFields, "is_active")
	}

	if err := s.roleRepo.Update(ctx, roleID, *role, updateFields...); err != nil {
		return nil, errorx.Wrap(errorx.ErrUpdateRole, err)
	}

	s.logger.Info(fmt.Sprintf("Role updated: %s (id: %s)", role.Name, roleID))
	updated := s.roleRepo.FindOneById(ctx, roleID)
	return dto.RoleRespFromModel(updated), nil
}

// DeleteRole deletes a role
func (s *RoleSvc) DeleteRole(ctx context.Context, roleID string, isSuperAdmin bool) error {
	// Check if role exists
	role := s.roleRepo.FindOneById(ctx, roleID)
	if role == nil {
		return errorx.New(errorx.ErrRoleNotFound, "Role not found")
	}

	// Validate system role deletion
	if role.ProjectID != nil && *role.ProjectID == constant.SystemProjectID && !isSuperAdmin {
		return errorx.New(errorx.ErrSystemRoleProtected, "Only super admins can delete system roles")
	}

	if err := s.roleRepo.DeleteById(ctx, roleID); err != nil {
		return errorx.Wrap(errorx.ErrDeleteRole, err)
	}

	s.logger.Info(fmt.Sprintf("Role deleted: %s (id: %s)", role.Name, roleID))

	return nil
}

// ListRoles lists roles with filters
func (s *RoleSvc) ListRoles(ctx context.Context, req dto.ListRolesReq) (*dto.PaginationResp[dto.RoleResp], error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * pageSize

	var roles []model.Role
	var total int64
	var err error

	if req.Search != "" || req.ProjectID != nil || req.IsActive != nil {
		roles, total, err = s.roleRepo.SearchRoles(ctx, req.Search, req.ProjectID, req.IsActive, pageSize, offset)
	} else {
		roles, err = s.roleRepo.FindAll(ctx)
		total = int64(len(roles))
		// Apply pagination manually
		start := offset
		end := offset + pageSize
		if start > len(roles) {
			roles = []model.Role{}
		} else {
			if end > len(roles) {
				end = len(roles)
			}
			roles = roles[start:end]
		}
	}

	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	items := make([]dto.RoleResp, 0, len(roles))
	for i := range roles {
		if r := dto.RoleRespFromModel(&roles[i]); r != nil {
			items = append(items, *r)
		}
	}

	hasNext := int64(offset+pageSize) < total

	return &dto.PaginationResp[dto.RoleResp]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  hasNext,
	}, nil
}

// AssignRoleToUser assigns a role to a user
func (s *RoleSvc) AssignRoleToUser(ctx context.Context, req dto.AssignRoleToUserReq, isSuperAdmin bool) (*dto.UserRoleResp, error) {
	// Check if user exists
	user := s.userRepo.FindOneById(ctx, req.UserID)
	if user == nil {
		return nil, errorx.New(errorx.ErrUserNotFound, "User not found")
	}

	// Check if role exists
	role := s.roleRepo.FindOneById(ctx, req.RoleID)
	if role == nil {
		return nil, errorx.New(errorx.ErrRoleNotFound, "Role not found")
	}

	// Validate system role assignment
	if role.ProjectID != nil && *role.ProjectID == constant.SystemProjectID && !isSuperAdmin {
		return nil, errorx.New(errorx.ErrSystemRoleProtected, "Only super admins can assign system roles")
	}

	// Check if assignment already exists
	existing, err := s.userRoleRepo.FindByUserIDAndRoleID(ctx, req.UserID, req.RoleID, req.ProjectID)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}
	if existing != nil {
		return nil, errorx.New(errorx.ErrConflict, "User already has this role")
	}

	// Create user role assignment
	userRole := &model.UserRole{
		UserID:    req.UserID,
		RoleID:    req.RoleID,
		ProjectID: req.ProjectID,
	}

	created, err := s.userRoleRepo.Create(ctx, userRole)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrRoleAssignment, err)
	}

	go s.clearUserPermissionsCache(req.UserID)

	s.logger.Info(fmt.Sprintf("Role assigned: user=%s, role=%s", req.UserID, req.RoleID))
	return dto.UserRoleRespFromModel(created, role), nil
}

// RemoveRoleFromUser removes a role from a user
func (s *RoleSvc) RemoveRoleFromUser(ctx context.Context, req dto.RemoveRoleFromUserReq, isSuperAdmin bool) error {
	// Check if role exists
	role := s.roleRepo.FindOneById(ctx, req.RoleID)
	if role == nil {
		return errorx.New(errorx.ErrRoleNotFound, "Role not found")
	}

	// Validate system role removal
	if role.ProjectID != nil && *role.ProjectID == constant.SystemProjectID && !isSuperAdmin {
		return errorx.New(errorx.ErrSystemRoleProtected, "Only super admins can remove system roles")
	}

	// Check if assignment exists
	existing, err := s.userRoleRepo.FindByUserIDAndRoleID(ctx, req.UserID, req.RoleID, req.ProjectID)
	if err != nil {
		return errorx.Wrap(errorx.ErrInternal, err)
	}
	if existing == nil {
		return errorx.New(errorx.ErrNotFound, "User role assignment not found")
	}

	if err := s.userRoleRepo.DeleteByUserIDAndRoleID(ctx, req.UserID, req.RoleID, req.ProjectID); err != nil {
		return errorx.Wrap(errorx.ErrRoleAssignment, err)
	}

	go s.clearUserPermissionsCache(req.UserID)

	s.logger.Info(fmt.Sprintf("Role removed: user=%s, role=%s", req.UserID, req.RoleID))

	return nil
}

// GetUserRoles retrieves all roles assigned to a user
func (s *RoleSvc) GetUserRoles(ctx context.Context, req dto.GetUserRolesReq) ([]dto.UserRoleResp, error) {
	// Check if user exists
	user := s.userRepo.FindOneById(ctx, req.UserID)
	if user == nil {
		return nil, errorx.New(errorx.ErrUserNotFound, "User not found")
	}

	userRoles, err := s.userRoleRepo.FindWithRole(ctx, req.UserID, req.ProjectID)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	results := make([]dto.UserRoleResp, 0, len(userRoles))
	for i := range userRoles {
		if ur := dto.UserRoleRespFromModel(&userRoles[i], &userRoles[i].Role); ur != nil {
			results = append(results, *ur)
		}
	}

	return results, nil
}

// GetUserPermissions retrieves all permissions assigned to a user
func (s *RoleSvc) GetUserPermissions(ctx context.Context, userID string) (dto.UserPermissions, error) {
	// cache the permissions for the user
	cacheKey := s.userPermissionsCacheKey(userID)
	var permissions dto.UserPermissions
	err := s.cache.Get(cacheKey, &permissions)
	if err == nil {
		return permissions, nil
	} else if err != cache.ErrCacheNil {
		return dto.UserPermissions{}, errorx.Wrap(errorx.ErrInternal, err)
	}

	userRoles, err := s.userRoleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	// Get all permissions from the user roles and loop through each role permissions with the project ID
	permissions = make(dto.UserPermissions)
	for _, userRole := range userRoles {
		for _, permissionCode := range model.PermissionsFromJSON(userRole.Role.Permissions) {
			permissions[s.buildPermissionKey(permissionCode, userRole.ProjectID)] = true
		}
	}

	ttl := constant.CacheDefaultTTL
	if err := s.cache.Set(cacheKey, permissions, &ttl); err != nil {
		return dto.UserPermissions{}, errorx.Wrap(errorx.ErrInternal, err)
	}

	return permissions, nil
}

func (s *RoleSvc) buildPermissionKey(permissionCode string, projectID *string) string {
	projectKey := constant.SystemProjectID
	if projectID != nil {
		projectKey = *projectID
	}
	return fmt.Sprintf("%s/%s", projectKey, permissionCode)
}

func (s *RoleSvc) userPermissionsCacheKey(userID string) string {
	return fmt.Sprintf("user_permissions:%s", userID)
}

func (s *RoleSvc) clearUserPermissionsCache(userID string) {
	cacheKey := s.userPermissionsCacheKey(userID)
	s.cache.Delete(cacheKey)
}
