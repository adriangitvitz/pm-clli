package project

import (
	"context"
	"fmt"

	"github.com/adriannajera/project-manager-cli/internal/domain"
)

// Service provides project management functionality
type Service struct {
	projectRepo domain.ProjectRepository
}

// NewService creates a new project service
func NewService(projectRepo domain.ProjectRepository) *Service {
	return &Service{
		projectRepo: projectRepo,
	}
}

// CreateProjectInput represents input for creating a project
type CreateProjectInput struct {
	Name        string
	Description string
	Color       string
}

// UpdateProjectInput represents input for updating a project
type UpdateProjectInput struct {
	ID          string
	Name        *string
	Description *string
	Status      *domain.ProjectStatus
	Color       *string
}

// ListOptions represents options for listing projects
type ListOptions struct {
	Status []domain.ProjectStatus
	Search string
	Limit  int
	Offset int
}

// CreateProject creates a new project
func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (*domain.Project, error) {
	if input.Name == "" {
		return nil, domain.ErrEmptyName
	}

	// Check if project with same name already exists
	if _, err := s.projectRepo.GetByName(ctx, input.Name); err == nil {
		return nil, domain.ErrDuplicateProject
	}

	project := domain.NewProject(input.Name, input.Description)
	if input.Color != "" {
		project.Color = input.Color
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// GetProject retrieves a project by ID
func (s *Service) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	if id == "" {
		return nil, domain.ErrInvalidProjectID
	}

	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// GetProjectByName retrieves a project by name
func (s *Service) GetProjectByName(ctx context.Context, name string) (*domain.Project, error) {
	if name == "" {
		return nil, domain.ErrEmptyName
	}

	project, err := s.projectRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get project by name: %w", err)
	}

	return project, nil
}

// UpdateProject updates an existing project
func (s *Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (*domain.Project, error) {
	if input.ID == "" {
		return nil, domain.ErrInvalidProjectID
	}

	project, err := s.projectRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project for update: %w", err)
	}

	// Update fields if provided
	if input.Name != nil {
		if *input.Name == "" {
			return nil, domain.ErrEmptyName
		}

		// Check if another project with same name exists
		if existing, err := s.projectRepo.GetByName(ctx, *input.Name); err == nil && existing.ID != project.ID {
			return nil, domain.ErrDuplicateProject
		}

		project.Name = *input.Name
	}

	if input.Description != nil {
		project.Description = *input.Description
	}

	if input.Status != nil {
		project.Status = *input.Status
		switch *input.Status {
		case domain.ProjectStatusArchived:
			project.Archive()
		case domain.ProjectStatusCompleted:
			project.Complete()
		case domain.ProjectStatusActive:
			project.Activate()
		case domain.ProjectStatusOnHold:
			project.PutOnHold()
		}
	}

	if input.Color != nil {
		project.Color = *input.Color
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// DeleteProject deletes a project
func (s *Service) DeleteProject(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidProjectID
	}

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ListProjects retrieves projects based on the provided options
func (s *Service) ListProjects(ctx context.Context, options ListOptions) ([]*domain.Project, error) {
	filter := domain.ProjectFilter{
		Status: options.Status,
		Search: options.Search,
		Limit:  options.Limit,
		Offset: options.Offset,
	}

	projects, err := s.projectRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}

// ArchiveProject archives a project
func (s *Service) ArchiveProject(ctx context.Context, id string) error {
	status := domain.ProjectStatusArchived
	input := UpdateProjectInput{
		ID:     id,
		Status: &status,
	}

	_, err := s.UpdateProject(ctx, input)
	return err
}

// CompleteProject marks a project as completed
func (s *Service) CompleteProject(ctx context.Context, id string) error {
	status := domain.ProjectStatusCompleted
	input := UpdateProjectInput{
		ID:     id,
		Status: &status,
	}

	_, err := s.UpdateProject(ctx, input)
	return err
}

// ActivateProject activates an archived or on-hold project
func (s *Service) ActivateProject(ctx context.Context, id string) error {
	status := domain.ProjectStatusActive
	input := UpdateProjectInput{
		ID:     id,
		Status: &status,
	}

	_, err := s.UpdateProject(ctx, input)
	return err
}

// PutProjectOnHold puts a project on hold
func (s *Service) PutProjectOnHold(ctx context.Context, id string) error {
	status := domain.ProjectStatusOnHold
	input := UpdateProjectInput{
		ID:     id,
		Status: &status,
	}

	_, err := s.UpdateProject(ctx, input)
	return err
}

// GetActiveProjects retrieves all active projects
func (s *Service) GetActiveProjects(ctx context.Context) ([]*domain.Project, error) {
	options := ListOptions{
		Status: []domain.ProjectStatus{domain.ProjectStatusActive},
	}

	return s.ListProjects(ctx, options)
}