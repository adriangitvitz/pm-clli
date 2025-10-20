package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	migrationVersion = 5
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
		changelist TEXT DEFAULT '',
		workspace TEXT DEFAULT '',
		note_id TEXT,
		note_path TEXT,
		has_note BOOLEAN DEFAULT 0,
		note_created_at DATETIME,
		note_updated_at DATETIME,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`,

	`CREATE TABLE IF NOT EXISTS time_entries (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		project_id TEXT,
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
	`CREATE INDEX IF NOT EXISTS idx_tasks_has_note ON tasks(has_note);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_note_id ON tasks(note_id);`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_task_id ON time_entries(task_id);`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_project_id ON time_entries(project_id);`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_start_time ON time_entries(start_time);`,

	// Migration v2: Add changelist column
	`ALTER TABLE tasks ADD COLUMN changelist TEXT DEFAULT '';`,

	// Migration v3: Add workspace column
	`ALTER TABLE tasks ADD COLUMN workspace TEXT DEFAULT '';`,

	// Migration v4: Make project_id nullable in time_entries table
	`CREATE TABLE IF NOT EXISTS time_entries_new (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		project_id TEXT,
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
	`INSERT INTO time_entries_new SELECT * FROM time_entries;`,
	`DROP TABLE time_entries;`,
	`ALTER TABLE time_entries_new RENAME TO time_entries;`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_task_id ON time_entries(task_id);`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_project_id ON time_entries(project_id);`,
	`CREATE INDEX IF NOT EXISTS idx_time_entries_start_time ON time_entries(start_time);`,

	// Migration v5: Add note linking columns
	`ALTER TABLE tasks ADD COLUMN note_id TEXT;`,
	`ALTER TABLE tasks ADD COLUMN note_path TEXT;`,
	`ALTER TABLE tasks ADD COLUMN has_note BOOLEAN DEFAULT 0;`,
	`ALTER TABLE tasks ADD COLUMN note_created_at DATETIME;`,
	`ALTER TABLE tasks ADD COLUMN note_updated_at DATETIME;`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_has_note ON tasks(has_note);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_note_id ON tasks(note_id);`,
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

	// For fresh install (version 0), run base schema with all fields
	if currentVersion == 0 {
		// Run base schema (migrations 0-12: 4 tables + 9 indices)
		// Base schema already includes changelist, workspace, and note fields
		// Skip ALTER TABLE migrations (13+) since base schema has everything
		for i := 0; i <= 12; i++ {
			if _, err := tx.ExecContext(ctx, migrations[i]); err != nil {
				return fmt.Errorf("failed to execute migration %d: %w", i, err)
			}
		}
	} else {
		// For existing databases, run incremental migrations
		startIndex := 10 + currentVersion
		for i := startIndex; i < len(migrations); i++ {
			if _, err := tx.ExecContext(ctx, migrations[i]); err != nil {
				return fmt.Errorf("failed to execute migration %d: %w", i, err)
			}
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