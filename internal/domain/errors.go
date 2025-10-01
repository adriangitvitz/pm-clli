package domain

import "errors"

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrProjectNotFound    = errors.New("project not found")
	ErrTimeEntryNotFound  = errors.New("time entry not found")
	ErrInvalidDueDate     = errors.New("due date cannot be in the past")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidPriority    = errors.New("invalid priority")
	ErrEmptyTitle         = errors.New("title cannot be empty")
	ErrEmptyName          = errors.New("name cannot be empty")
	ErrDuplicateProject   = errors.New("project with this name already exists")
	ErrActiveTimeEntry    = errors.New("there is already an active time entry")
	ErrNoActiveTimeEntry  = errors.New("no active time entry found")
	ErrCircularDependency = errors.New("circular dependency detected")
	ErrInvalidTaskID      = errors.New("invalid task ID")
	ErrInvalidProjectID   = errors.New("invalid project ID")
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrMigrationFailed    = errors.New("database migration failed")
	ErrConfigNotFound     = errors.New("configuration file not found")
	ErrInvalidConfig      = errors.New("invalid configuration")
)
