package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adriannajera/project-manager-cli/internal/domain"
)

type ProjectRepository struct {
	db *DB
}

func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	metadataJSON, _ := json.Marshal(project.Metadata)

	query := `
		INSERT INTO projects (
			id, name, description, status, color, created_at, updated_at, archived_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Description, string(project.Status),
		project.Color, project.CreatedAt, project.UpdatedAt, project.ArchivedAt,
		string(metadataJSON),
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrDuplicateProject
		}
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	query := `
		SELECT id, name, description, status, color, created_at, updated_at, archived_at, metadata
		FROM projects WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	project, err := r.scanProject(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project by ID: %w", err)
	}

	return project, nil
}

func (r *ProjectRepository) GetByName(ctx context.Context, name string) (*domain.Project, error) {
	query := `
		SELECT id, name, description, status, color, created_at, updated_at, archived_at, metadata
		FROM projects WHERE name = ?
	`

	row := r.db.QueryRowContext(ctx, query, name)
	project, err := r.scanProject(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project by name: %w", err)
	}

	return project, nil
}

func (r *ProjectRepository) List(ctx context.Context, filter domain.ProjectFilter) ([]*domain.Project, error) {
	query := "SELECT id, name, description, status, color, created_at, updated_at, archived_at, metadata FROM projects WHERE 1=1"
	args := []interface{}{}

	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			placeholders[i] = "?"
			args = append(args, string(status))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	if filter.Search != "" {
		query += " AND (name LIKE ? OR description LIKE ?)"
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
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		project, err := r.scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	metadataJSON, _ := json.Marshal(project.Metadata)

	query := `
		UPDATE projects SET
			name = ?, description = ?, status = ?, color = ?,
			updated_at = ?, archived_at = ?, metadata = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		project.Name, project.Description, string(project.Status), project.Color,
		project.UpdatedAt, project.ArchivedAt, string(metadataJSON), project.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrDuplicateProject
		}
		return fmt.Errorf("failed to update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrProjectNotFound
	}

	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM projects WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrProjectNotFound
	}

	return nil
}

func (r *ProjectRepository) scanProject(row RowScanner) (*domain.Project, error) {
	var project domain.Project
	var metadataJSON string
	var archivedAt sql.NullTime

	err := row.Scan(
		&project.ID, &project.Name, &project.Description, &project.Status,
		&project.Color, &project.CreatedAt, &project.UpdatedAt, &archivedAt,
		&metadataJSON,
	)

	if err != nil {
		return nil, err
	}

	if archivedAt.Valid {
		project.ArchivedAt = &archivedAt.Time
	}

	if err := json.Unmarshal([]byte(metadataJSON), &project.Metadata); err != nil {
		project.Metadata = make(map[string]interface{})
	}

	return &project, nil
}