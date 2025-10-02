package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	migrationVersion = 3
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY
	);`,

	`CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		color TEXT NOT NULL DEFAULT '#3498db',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		archived_at DATETIME,
		metadata TEXT
	);`,

	`CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'todo',
		priority INTEGER NOT NULL DEFAULT 1,
		project_id TEXT,
		parent_id TEXT,
		tags TEXT,
		due_date DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		completed_at DATETIME,
		metadata TEXT,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`,

	`CREATE TABLE IF NOT EXISTS time_entries (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		project_id TEXT NOT NULL,
		description TEXT,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		duration INTEGER,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		metadata TEXT,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);`,

	`CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_task_id ON time_entries(task_id);`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_project_id ON time_entries(project_id);`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_start_time ON time_entries(start_time);`,

	// Migration v2: Add changelist column
	`ALTER TABLE tasks ADD COLUMN changelist TEXT DEFAULT '';`,

	// Migration v3: Add workspace column
	`ALTER TABLE tasks ADD COLUMN workspace TEXT DEFAULT '';`,
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
	currentVersion, err := getCurrentVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	if currentVersion >= migrationVersion {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Only run migrations that haven't been executed yet
	// First 10 migrations are base schema, v2 is migration 11 (index 11), v3 is migration 12 (index 12)
	startIndex := 0
	if currentVersion > 0 {
		// Base schema (migrations 0-9) + version migrations
		// If we're at version 2, we've run migrations 0-11, so start at 12
		startIndex = 10 + currentVersion
	}

	for i := startIndex; i < len(migrations); i++ {
		if _, err := tx.ExecContext(ctx, migrations[i]); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", i, err)
		}
	}

	if err := setVersion(ctx, tx, migrationVersion); err != nil {
		return fmt.Errorf("failed to update schema version: %w", err)
	}

	return tx.Commit()
}

func getCurrentVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	err := db.QueryRowContext(ctx, "SELECT version FROM schema_version ORDER BY version DESC LIMIT 1").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil && err.Error() == "SQL logic error: no such table: schema_version (1)" {
		return 0, nil
	}
	return version, err
}

func setVersion(ctx context.Context, tx *sql.Tx, version int) error {
	_, err := tx.ExecContext(ctx, "INSERT OR REPLACE INTO schema_version (version) VALUES (?)", version)
	return err
}