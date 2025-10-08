package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adriannajera/project-manager-cli/internal/domain"
)

type TaskRepository struct {
	db *DB
}

func NewTaskRepository(db *DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	tagsJSON, _ := json.Marshal(task.Tags)
	metadataJSON, _ := json.Marshal(task.Metadata)

	query := `
		INSERT INTO tasks (
			id, title, description, status, priority, project_id, parent_id,
			tags, changelist, workspace, due_date, created_at, updated_at, completed_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert empty project_id to NULL to avoid foreign key constraint violations
	var projectID interface{}
	if task.ProjectID != "" {
		projectID = task.ProjectID
	} else {
		projectID = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, string(task.Status),
		int(task.Priority), projectID, task.ParentID, string(tagsJSON),
		task.Changelist, task.Workspace, task.DueDate, task.CreatedAt, task.UpdatedAt, task.CompletedAt,
		string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, project_id, parent_id,
		       tags, changelist, workspace, due_date, created_at, updated_at, completed_at, metadata
		FROM tasks WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	task, err := r.scanTask(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task by ID: %w", err)
	}

	return task, nil
}

func (r *TaskRepository) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, error) {
	query := "SELECT id, title, description, status, priority, project_id, parent_id, tags, changelist, workspace, due_date, created_at, updated_at, completed_at, metadata FROM tasks WHERE 1=1"
	args := []interface{}{}

	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			placeholders[i] = "?"
			args = append(args, string(status))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	if len(filter.Priority) > 0 {
		placeholders := make([]string, len(filter.Priority))
		for i, priority := range filter.Priority {
			placeholders[i] = "?"
			args = append(args, int(priority))
		}
		query += fmt.Sprintf(" AND priority IN (%s)", strings.Join(placeholders, ","))
	}

	if filter.ProjectID != "" {
		query += " AND project_id = ?"
		args = append(args, filter.ProjectID)
	}

	if filter.Workspace != "" {
		query += " AND workspace = ?"
		args = append(args, filter.Workspace)
	}

	if filter.DueBefore != nil {
		query += " AND due_date <= ?"
		args = append(args, filter.DueBefore)
	}

	if filter.DueAfter != nil {
		query += " AND due_date >= ?"
		args = append(args, filter.DueAfter)
	}

	if filter.Search != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	tagsJSON, _ := json.Marshal(task.Tags)
	metadataJSON, _ := json.Marshal(task.Metadata)

	query := `
		UPDATE tasks SET
			title = ?, description = ?, status = ?, priority = ?,
			project_id = ?, parent_id = ?, tags = ?, changelist = ?, workspace = ?, due_date = ?,
			updated_at = ?, completed_at = ?, metadata = ?
		WHERE id = ?
	`

	// Convert empty project_id to NULL to avoid foreign key constraint violations
	var projectID interface{}
	if task.ProjectID != "" {
		projectID = task.ProjectID
	} else {
		projectID = nil
	}

	result, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, string(task.Status), int(task.Priority),
		projectID, task.ParentID, string(tagsJSON), task.Changelist, task.Workspace, task.DueDate,
		task.UpdatedAt, task.CompletedAt, string(metadataJSON), task.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM tasks WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) GetByProject(ctx context.Context, projectID string) ([]*domain.Task, error) {
	filter := domain.TaskFilter{ProjectID: projectID}
	return r.List(ctx, filter)
}

func (r *TaskRepository) GetSubtasks(ctx context.Context, parentID string) ([]*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, project_id, parent_id,
		       tags, changelist, workspace, due_date, created_at, updated_at, completed_at, metadata
		FROM tasks WHERE parent_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subtask: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

type RowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *TaskRepository) scanTask(row RowScanner) (*domain.Task, error) {
	var task domain.Task
	var tagsJSON, metadataJSON string
	var projectID, parentID, changelist, workspace sql.NullString
	var dueDate, completedAt sql.NullTime

	err := row.Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.Priority, &projectID, &parentID, &tagsJSON, &changelist, &workspace,
		&dueDate, &task.CreatedAt, &task.UpdatedAt, &completedAt,
		&metadataJSON,
	)

	if err != nil {
		return nil, err
	}

	if projectID.Valid {
		task.ProjectID = projectID.String
	}

	if parentID.Valid {
		task.ParentID = &parentID.String
	}

	if changelist.Valid {
		task.Changelist = changelist.String
	}

	if workspace.Valid {
		task.Workspace = workspace.String
	}

	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}

	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	if err := json.Unmarshal([]byte(tagsJSON), &task.Tags); err != nil {
		task.Tags = make([]string, 0)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &task.Metadata); err != nil {
		task.Metadata = make(map[string]interface{})
	}

	return &task, nil
}