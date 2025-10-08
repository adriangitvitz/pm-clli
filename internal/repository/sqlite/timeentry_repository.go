package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adriannajera/project-manager-cli/internal/domain"
)

type TimeEntryRepository struct {
	db *DB
}

func NewTimeEntryRepository(db *DB) *TimeEntryRepository {
	return &TimeEntryRepository{db: db}
}

func (r *TimeEntryRepository) Create(ctx context.Context, entry *domain.TimeEntry) error {
	metadataJSON, _ := json.Marshal(entry.Metadata)

	query := `
		INSERT INTO time_entries (
			id, task_id, project_id, description, start_time, end_time,
			duration, created_at, updated_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var durationNanos *int64
	if entry.Duration > 0 {
		nanos := entry.Duration.Nanoseconds()
		durationNanos = &nanos
	}

	// Convert empty project_id to NULL to avoid foreign key constraint violations
	var projectID interface{}
	if entry.ProjectID != "" {
		projectID = entry.ProjectID
	} else {
		projectID = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.TaskID, projectID, entry.Description,
		entry.StartTime, entry.EndTime, durationNanos, entry.CreatedAt,
		entry.UpdatedAt, string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to create time entry: %w", err)
	}

	return nil
}

func (r *TimeEntryRepository) GetByID(ctx context.Context, id string) (*domain.TimeEntry, error) {
	query := `
		SELECT id, task_id, project_id, description, start_time, end_time,
		       duration, created_at, updated_at, metadata
		FROM time_entries WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	entry, err := r.scanTimeEntry(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrTimeEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get time entry by ID: %w", err)
	}

	return entry, nil
}

func (r *TimeEntryRepository) List(ctx context.Context, filter domain.TimeEntryFilter) ([]*domain.TimeEntry, error) {
	query := "SELECT id, task_id, project_id, description, start_time, end_time, duration, created_at, updated_at, metadata FROM time_entries WHERE 1=1"
	args := []interface{}{}

	if filter.TaskID != "" {
		query += " AND task_id = ?"
		args = append(args, filter.TaskID)
	}

	if filter.ProjectID != "" {
		query += " AND project_id = ?"
		args = append(args, filter.ProjectID)
	}

	if filter.StartAfter != nil {
		query += " AND start_time >= ?"
		args = append(args, filter.StartAfter)
	}

	if filter.EndBefore != nil {
		query += " AND (end_time IS NULL OR end_time <= ?)"
		args = append(args, filter.EndBefore)
	}

	if filter.Active != nil {
		if *filter.Active {
			query += " AND end_time IS NULL"
		} else {
			query += " AND end_time IS NOT NULL"
		}
	}

	query += " ORDER BY start_time DESC"

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
		return nil, fmt.Errorf("failed to list time entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.TimeEntry
	for rows.Next() {
		entry, err := r.scanTimeEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan time entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *TimeEntryRepository) Update(ctx context.Context, entry *domain.TimeEntry) error {
	metadataJSON, _ := json.Marshal(entry.Metadata)

	query := `
		UPDATE time_entries SET
			task_id = ?, project_id = ?, description = ?, start_time = ?,
			end_time = ?, duration = ?, updated_at = ?, metadata = ?
		WHERE id = ?
	`

	var durationNanos *int64
	if entry.Duration > 0 {
		nanos := entry.Duration.Nanoseconds()
		durationNanos = &nanos
	}

	// Convert empty project_id to NULL to avoid foreign key constraint violations
	var projectID interface{}
	if entry.ProjectID != "" {
		projectID = entry.ProjectID
	} else {
		projectID = nil
	}

	result, err := r.db.ExecContext(ctx, query,
		entry.TaskID, projectID, entry.Description, entry.StartTime,
		entry.EndTime, durationNanos, entry.UpdatedAt, string(metadataJSON),
		entry.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update time entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrTimeEntryNotFound
	}

	return nil
}

func (r *TimeEntryRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM time_entries WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete time entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrTimeEntryNotFound
	}

	return nil
}

func (r *TimeEntryRepository) GetActive(ctx context.Context) (*domain.TimeEntry, error) {
	query := `
		SELECT id, task_id, project_id, description, start_time, end_time,
		       duration, created_at, updated_at, metadata
		FROM time_entries WHERE end_time IS NULL
		ORDER BY start_time DESC LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query)
	entry, err := r.scanTimeEntry(row)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNoActiveTimeEntry
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active time entry: %w", err)
	}

	return entry, nil
}

func (r *TimeEntryRepository) GetByTask(ctx context.Context, taskID string) ([]*domain.TimeEntry, error) {
	filter := domain.TimeEntryFilter{TaskID: taskID}
	return r.List(ctx, filter)
}

func (r *TimeEntryRepository) GetByProject(ctx context.Context, projectID string) ([]*domain.TimeEntry, error) {
	filter := domain.TimeEntryFilter{ProjectID: projectID}
	return r.List(ctx, filter)
}

func (r *TimeEntryRepository) scanTimeEntry(row RowScanner) (*domain.TimeEntry, error) {
	var entry domain.TimeEntry
	var metadataJSON string
	var projectID sql.NullString
	var endTime sql.NullTime
	var durationNanos sql.NullInt64

	err := row.Scan(
		&entry.ID, &entry.TaskID, &projectID, &entry.Description,
		&entry.StartTime, &endTime, &durationNanos, &entry.CreatedAt,
		&entry.UpdatedAt, &metadataJSON,
	)

	if err != nil {
		return nil, err
	}

	if projectID.Valid {
		entry.ProjectID = projectID.String
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	if durationNanos.Valid {
		entry.Duration = time.Duration(durationNanos.Int64)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &entry.Metadata); err != nil {
		entry.Metadata = make(map[string]interface{})
	}

	return &entry, nil
}