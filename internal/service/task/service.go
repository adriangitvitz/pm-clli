package task

import (
	"context"
	"fmt"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
	"github.com/adriannajera/project-manager-cli/internal/domain"
)

// Service provides task management functionality
type Service struct {
	taskRepo domain.TaskRepository
	gitRepo  domain.GitRepository
	parser   *when.Parser
}

// NewService creates a new task service
func NewService(taskRepo domain.TaskRepository, gitRepo domain.GitRepository) *Service {
	// Initialize natural language date parser
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	return &Service{
		taskRepo: taskRepo,
		gitRepo:  gitRepo,
		parser:   w,
	}
}

// CreateTaskInput represents input for creating a task
type CreateTaskInput struct {
	Title       string
	Description string
	Priority    domain.Priority
	ProjectID   string
	ParentID    *string
	Tags        []string
	Changelist  string
	DueDate     string // Natural language date
}

// UpdateTaskInput represents input for updating a task
type UpdateTaskInput struct {
	ID          string
	Title       *string
	Description *string
	Status      *domain.TaskStatus
	Priority    *domain.Priority
	ProjectID   *string
	ParentID    *string
	Tags        []string
	Changelist  *string
	DueDate     *string // Natural language date
}

// ListOptions represents options for listing tasks
type ListOptions struct {
	Status    []domain.TaskStatus
	Priority  []domain.Priority
	ProjectID string
	Tags      []string
	Search    string
	DueBefore *time.Time
	DueAfter  *time.Time
	Limit     int
	Offset    int
}

// CreateTask creates a new task
func (s *Service) CreateTask(ctx context.Context, input CreateTaskInput) (*domain.Task, error) {
	if input.Title == "" {
		return nil, domain.ErrEmptyTitle
	}

	task := domain.NewTask(input.Title, input.Description)
	task.Priority = input.Priority
	task.ProjectID = input.ProjectID
	task.ParentID = input.ParentID
	task.Tags = input.Tags
	task.Changelist = input.Changelist

	// Parse due date if provided
	if input.DueDate != "" {
		if dueDate, err := s.parseDueDate(input.DueDate); err == nil {
			task.DueDate = dueDate
		} else {
			return nil, fmt.Errorf("invalid due date format: %w", err)
		}
	}

	// Add Git integration if available
	if s.gitRepo != nil && s.gitRepo.IsInRepository() {
		if branch, err := s.gitRepo.GetCurrentBranch(); err == nil && branch != "" {
			task.AddTag("branch:" + branch)
			task.Metadata["git_branch"] = branch
		}
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (s *Service) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		return nil, domain.ErrInvalidTaskID
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// UpdateTask updates an existing task
func (s *Service) UpdateTask(ctx context.Context, input UpdateTaskInput) (*domain.Task, error) {
	if input.ID == "" {
		return nil, domain.ErrInvalidTaskID
	}

	task, err := s.taskRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task for update: %w", err)
	}

	// Update fields if provided
	if input.Title != nil {
		if *input.Title == "" {
			return nil, domain.ErrEmptyTitle
		}
		task.Title = *input.Title
	}

	if input.Description != nil {
		task.Description = *input.Description
	}

	if input.Status != nil {
		task.Status = *input.Status
		if *input.Status == domain.StatusDone {
			now := time.Now()
			task.CompletedAt = &now
		} else {
			task.CompletedAt = nil
		}
	}

	if input.Priority != nil {
		task.Priority = *input.Priority
	}

	if input.ProjectID != nil {
		task.ProjectID = *input.ProjectID
	}

	if input.ParentID != nil {
		task.ParentID = input.ParentID
	}

	if input.Tags != nil {
		task.Tags = input.Tags
	}

	if input.Changelist != nil {
		task.Changelist = *input.Changelist
	}

	if input.DueDate != nil {
		if *input.DueDate == "" {
			task.DueDate = nil
		} else if dueDate, err := s.parseDueDate(*input.DueDate); err == nil {
			task.DueDate = dueDate
		} else {
			return nil, fmt.Errorf("invalid due date format: %w", err)
		}
	}

	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

// DeleteTask deletes a task
func (s *Service) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidTaskID
	}

	if err := s.taskRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// ListTasks retrieves tasks based on the provided options
func (s *Service) ListTasks(ctx context.Context, options ListOptions) ([]*domain.Task, error) {
	filter := domain.TaskFilter{
		Status:    options.Status,
		Priority:  options.Priority,
		ProjectID: options.ProjectID,
		Tags:      options.Tags,
		Search:    options.Search,
		DueBefore: options.DueBefore,
		DueAfter:  options.DueAfter,
		Limit:     options.Limit,
		Offset:    options.Offset,
	}

	tasks, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil
}

// CompleteTask marks a task as complete
func (s *Service) CompleteTask(ctx context.Context, id string) error {
	status := domain.StatusDone
	input := UpdateTaskInput{
		ID:     id,
		Status: &status,
	}

	_, err := s.UpdateTask(ctx, input)
	return err
}

// StartTask marks a task as in progress
func (s *Service) StartTask(ctx context.Context, id string) error {
	status := domain.StatusDoing
	input := UpdateTaskInput{
		ID:     id,
		Status: &status,
	}

	_, err := s.UpdateTask(ctx, input)
	return err
}

// BlockTask marks a task as blocked
func (s *Service) BlockTask(ctx context.Context, id string) error {
	status := domain.StatusBlocked
	input := UpdateTaskInput{
		ID:     id,
		Status: &status,
	}

	_, err := s.UpdateTask(ctx, input)
	return err
}

// GetSubtasks retrieves all subtasks for a given parent task
func (s *Service) GetSubtasks(ctx context.Context, parentID string) ([]*domain.Task, error) {
	if parentID == "" {
		return nil, domain.ErrInvalidTaskID
	}

	subtasks, err := s.taskRepo.GetSubtasks(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtasks: %w", err)
	}

	return subtasks, nil
}

// GetTasksByProject retrieves all tasks for a given project
func (s *Service) GetTasksByProject(ctx context.Context, projectID string) ([]*domain.Task, error) {
	if projectID == "" {
		return nil, domain.ErrInvalidProjectID
	}

	tasks, err := s.taskRepo.GetByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by project: %w", err)
	}

	return tasks, nil
}

// AddTag adds a tag to a task
func (s *Service) AddTag(ctx context.Context, taskID, tag string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.AddTag(tag)

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// RemoveTag removes a tag from a task
func (s *Service) RemoveTag(ctx context.Context, taskID, tag string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.RemoveTag(tag)

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// GetOverdueTasks retrieves all overdue tasks
func (s *Service) GetOverdueTasks(ctx context.Context) ([]*domain.Task, error) {
	now := time.Now()
	options := ListOptions{
		Status:    []domain.TaskStatus{domain.StatusTodo, domain.StatusDoing, domain.StatusBlocked},
		DueBefore: &now,
	}

	return s.ListTasks(ctx, options)
}

// GetTasksDueToday retrieves all tasks due today
func (s *Service) GetTasksDueToday(ctx context.Context) ([]*domain.Task, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	options := ListOptions{
		Status:    []domain.TaskStatus{domain.StatusTodo, domain.StatusDoing, domain.StatusBlocked},
		DueAfter:  &startOfDay,
		DueBefore: &endOfDay,
	}

	return s.ListTasks(ctx, options)
}

// parseDueDate parses a natural language due date
func (s *Service) parseDueDate(dateStr string) (*time.Time, error) {
	result, err := s.parser.Parse(dateStr, time.Now())
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("could not parse date: %s", dateStr)
	}

	return &result.Time, nil
}