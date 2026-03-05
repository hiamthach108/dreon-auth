package service

import (
	"context"
	"strings"

	"github.com/hiamthach108/dreon-auth/internal/aggregate"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/repository"
	"github.com/hiamthach108/dreon-auth/internal/shared/helper"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
)

// IProjectSvc defines the contract for project operations.
type IProjectSvc interface {
	Create(ctx context.Context, req aggregate.CreateProjectReq) (*aggregate.ProjectDto, error)
	GetByID(ctx context.Context, id string) (*aggregate.ProjectDto, error)
	List(ctx context.Context, page, pageSize int) (*aggregate.PaginationResp[aggregate.ProjectDto], error)
	Update(ctx context.Context, id string, req aggregate.UpdateProjectReq) (*aggregate.ProjectDto, error)
	Delete(ctx context.Context, id string) error
}

// ProjectSvc implements IProjectSvc.
type ProjectSvc struct {
	logger logger.ILogger
	repo   repository.IProjectRepository
}

// NewProjectSvc creates a new project service.
func NewProjectSvc(logger logger.ILogger, repo repository.IProjectRepository) IProjectSvc {
	return &ProjectSvc{
		logger: logger,
		repo:   repo,
	}
}

// Create creates a new project.
func (s *ProjectSvc) Create(ctx context.Context, req aggregate.CreateProjectReq) (*aggregate.ProjectDto, error) {

	model := req.ToModel()
	model.Code = s.generateCode(req.Name)
	created, err := s.repo.Create(ctx, model)
	if err != nil {
		s.logger.Error("[ProjectSvc] failed to create project", "code", model.Code, "error", err)
		return nil, errorx.Wrap(errorx.ErrCreateProject, err)
	}

	var resp aggregate.ProjectDto
	resp.FromModel(created)
	return &resp, nil
}

// GetByID returns a project by ID.
func (s *ProjectSvc) GetByID(ctx context.Context, id string) (*aggregate.ProjectDto, error) {
	p := s.repo.FindOneById(ctx, id)
	if p == nil {
		return nil, errorx.Wrap(errorx.ErrProjectNotFound, nil)
	}
	var resp aggregate.ProjectDto
	resp.FromModel(p)
	return &resp, nil
}

// List returns a paginated list of projects.
func (s *ProjectSvc) List(ctx context.Context, page, pageSize int) (*aggregate.PaginationResp[aggregate.ProjectDto], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	projects, total, err := s.repo.List(ctx, offset, pageSize)
	if err != nil {
		s.logger.Error("[ProjectSvc] failed to list projects", "error", err)
		return nil, errorx.Wrap(errorx.ErrInternal, err)
	}

	items := make([]aggregate.ProjectDto, 0, len(projects))
	for i := range projects {
		var d aggregate.ProjectDto
		d.FromModel(&projects[i])
		items = append(items, d)
	}

	hasNext := int64(offset+len(projects)) < total
	return &aggregate.PaginationResp[aggregate.ProjectDto]{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  hasNext,
		Items:    items,
	}, nil
}

// Update updates a project by ID (partial update).
func (s *ProjectSvc) Update(ctx context.Context, id string, req aggregate.UpdateProjectReq) (*aggregate.ProjectDto, error) {
	p := s.repo.FindOneById(ctx, id)
	if p == nil {
		return nil, errorx.Wrap(errorx.ErrProjectNotFound, nil)
	}

	updated, fields := req.ToModelAndFields()
	if len(fields) == 0 {
		var resp aggregate.ProjectDto
		resp.FromModel(p)
		return &resp, nil
	}

	if err := s.repo.Update(ctx, id, *updated, fields...); err != nil {
		s.logger.Error("[ProjectSvc] failed to update project", "id", id, "error", err)
		return nil, errorx.Wrap(errorx.ErrUpdateProject, err)
	}

	updatedProject := s.repo.FindOneById(ctx, id)
	if updatedProject == nil {
		var resp aggregate.ProjectDto
		resp.FromModel(p)
		return &resp, nil
	}
	var resp aggregate.ProjectDto
	resp.FromModel(updatedProject)
	return &resp, nil
}

// Delete deletes a project by ID.
func (s *ProjectSvc) Delete(ctx context.Context, id string) error {
	p := s.repo.FindOneById(ctx, id)
	if p == nil {
		return errorx.Wrap(errorx.ErrProjectNotFound, nil)
	}
	if err := s.repo.DeleteById(ctx, id); err != nil {
		s.logger.Error("[ProjectSvc] failed to delete project", "id", id, "error", err)
		return errorx.Wrap(errorx.ErrInternal, err)
	}
	return nil
}

func (s *ProjectSvc) generateCode(name string) string {
	return strings.ToUpper(helper.NormalizeSlug(name) + "-" + helper.RandomString(6))
}
