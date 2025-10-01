package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusBacklog TaskStatus = "backlog"
	StatusTodo    TaskStatus = "todo"
	StatusDoing   TaskStatus = "doing"
	StatusDone    TaskStatus = "done"
	StatusBlocked TaskStatus = "blocked"
)

type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "normal"
	}
}

type Task struct {
	ID          string                 `json:"id" db:"id"`
	Title       string                 `json:"title" db:"title"`
	Description string                 `json:"description" db:"description"`
	Status      TaskStatus             `json:"status" db:"status"`
	Priority    Priority               `json:"priority" db:"priority"`
	ProjectID   string                 `json:"project_id" db:"project_id"`
	ParentID    *string                `json:"parent_id" db:"parent_id"`
	Tags        []string               `json:"tags" db:"tags"`
	Changelist  string                 `json:"changelist" db:"changelist"`
	DueDate     *time.Time             `json:"due_date" db:"due_date"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at" db:"completed_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

func NewTask(title, description string) *Task {
	now := time.Now()
	return &Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		Priority:    PriorityNormal,
		Tags:        make([]string, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]interface{}),
	}
}

func (t *Task) Complete() {
	now := time.Now()
	t.Status = StatusDone
	t.CompletedAt = &now
	t.UpdatedAt = now
}

func (t *Task) Start() {
	t.Status = StatusDoing
	t.UpdatedAt = time.Now()
}

func (t *Task) Block() {
	t.Status = StatusBlocked
	t.UpdatedAt = time.Now()
}

func (t *Task) AddTag(tag string) {
	for _, existingTag := range t.Tags {
		if existingTag == tag {
			return
		}
	}
	t.Tags = append(t.Tags, tag)
	t.UpdatedAt = time.Now()
}

func (t *Task) RemoveTag(tag string) {
	for i, existingTag := range t.Tags {
		if existingTag == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			t.UpdatedAt = time.Now()
			return
		}
	}
}

func (t *Task) HasTag(tag string) bool {
	for _, existingTag := range t.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == StatusDone {
		return false
	}
	return time.Now().After(*t.DueDate)
}

func (t *Task) DaysUntilDue() int {
	if t.DueDate == nil {
		return 0
	}
	duration := time.Until(*t.DueDate)
	return int(duration.Hours() / 24)
}
