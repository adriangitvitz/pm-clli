package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusArchived  ProjectStatus = "archived"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
)

type Project struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Status      ProjectStatus          `json:"status" db:"status"`
	Color       string                 `json:"color" db:"color"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	ArchivedAt  *time.Time             `json:"archived_at" db:"archived_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

func NewProject(name, description string) *Project {
	now := time.Now()
	return &Project{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Status:      ProjectStatusActive,
		Color:       "#3498db",
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]interface{}),
	}
}

func (p *Project) Archive() {
	now := time.Now()
	p.Status = ProjectStatusArchived
	p.ArchivedAt = &now
	p.UpdatedAt = now
}

func (p *Project) Complete() {
	p.Status = ProjectStatusCompleted
	p.UpdatedAt = time.Now()
}

func (p *Project) Activate() {
	p.Status = ProjectStatusActive
	p.ArchivedAt = nil
	p.UpdatedAt = time.Now()
}

func (p *Project) PutOnHold() {
	p.Status = ProjectStatusOnHold
	p.UpdatedAt = time.Now()
}
