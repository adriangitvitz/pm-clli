package domain

import (
	"context"
	"time"
)

type TaskFilter struct {
	Status    []TaskStatus
	Priority  []Priority
	ProjectID string
	Workspace string
	Tags      []string
	DueBefore *time.Time
	DueAfter  *time.Time
	Search    string
	Limit     int
	Offset    int
}

type ProjectFilter struct {
	Status []ProjectStatus
	Search string
	Limit  int
	Offset int
}

type TimeEntryFilter struct {
	TaskID     string
	ProjectID  string
	StartAfter *time.Time
	EndBefore  *time.Time
	Active     *bool
	Limit      int
	Offset     int
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id string) (*Task, error)
	List(ctx context.Context, filter TaskFilter) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
	GetByProject(ctx context.Context, projectID string) ([]*Task, error)
	GetSubtasks(ctx context.Context, parentID string) ([]*Task, error)
}

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context, filter ProjectFilter) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
	GetByName(ctx context.Context, name string) (*Project, error)
}

type TimeEntryRepository interface {
	Create(ctx context.Context, entry *TimeEntry) error
	GetByID(ctx context.Context, id string) (*TimeEntry, error)
	List(ctx context.Context, filter TimeEntryFilter) ([]*TimeEntry, error)
	Update(ctx context.Context, entry *TimeEntry) error
	Delete(ctx context.Context, id string) error
	GetActive(ctx context.Context) (*TimeEntry, error)
	GetByTask(ctx context.Context, taskID string) ([]*TimeEntry, error)
	GetByProject(ctx context.Context, projectID string) ([]*TimeEntry, error)
}

type GitRepository interface {
	GetCurrentBranch() (string, error)
	GetCurrentCommit() (string, error)
	IsInRepository() bool
	GetRepositoryRoot() (string, error)
	CreateCommitHook(taskID string) error
	RemoveCommitHook() error
}

type ConfigRepository interface {
	Load() (*Config, error)
	Save(config *Config) error
	GetConfigPath() string
}

type Config struct {
	DatabasePath   string            `yaml:"database_path"`
	DefaultProject string            `yaml:"default_project"`
	GitIntegration bool              `yaml:"git_integration"`
	TimeFormat     string            `yaml:"time_format"`
	DateFormat     string            `yaml:"date_format"`
	Theme          Theme             `yaml:"theme"`
	Aliases        map[string]string `yaml:"aliases"`
}

type Theme struct {
	Primary   string `yaml:"primary"`
	Secondary string `yaml:"secondary"`
	Success   string `yaml:"success"`
	Warning   string `yaml:"warning"`
	Error     string `yaml:"error"`
	Muted     string `yaml:"muted"`
}
