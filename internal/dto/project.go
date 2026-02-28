package dto

import (
	"time"

	"github.com/hiamthach108/dreon-auth/internal/model"
)

// CreateProjectReq is the request body for creating a project.
type CreateProjectReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// UpdateProjectReq is the request body for updating a project (partial update).
type UpdateProjectReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// ProjectDto is the response DTO for project.
type ProjectDto struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// FromModel maps a model.Project to ProjectDto.
func (d *ProjectDto) FromModel(m *model.Project) {
	if m == nil {
		return
	}
	d.ID = m.ID
	d.Code = m.Code
	d.Name = m.Name
	d.Description = m.Description
	d.CreatedAt = m.CreatedAt
	d.UpdatedAt = m.UpdatedAt
}

// ToModel maps CreateProjectReq to model.Project.
func (r *CreateProjectReq) ToModel() *model.Project {
	return &model.Project{
		Name:        r.Name,
		Description: r.Description,
	}
}

// ToModelAndFields returns the model and list of fields to update for UpdateProjectReq.
func (r *UpdateProjectReq) ToModelAndFields() (p *model.Project, fields []string) {
	p = &model.Project{}
	if r.Name != nil {
		p.Name = *r.Name
		fields = append(fields, "name")
	}
	if r.Description != nil {
		p.Description = *r.Description
		fields = append(fields, "description")
	}
	return p, fields
}
