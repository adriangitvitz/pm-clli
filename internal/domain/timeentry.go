package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TimeEntry struct {
	ID          string                 `json:"id" db:"id"`
	TaskID      string                 `json:"task_id" db:"task_id"`
	ProjectID   string                 `json:"project_id" db:"project_id"`
	Description string                 `json:"description" db:"description"`
	StartTime   time.Time              `json:"start_time" db:"start_time"`
	EndTime     *time.Time             `json:"end_time" db:"end_time"`
	Duration    time.Duration          `json:"duration" db:"duration"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

func NewTimeEntry(taskID, projectID, description string) *TimeEntry {
	now := time.Now()
	return &TimeEntry{
		ID:          uuid.New().String(),
		TaskID:      taskID,
		ProjectID:   projectID,
		Description: description,
		StartTime:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]interface{}),
	}
}

func (te *TimeEntry) Stop() {
	now := time.Now()
	te.EndTime = &now
	te.Duration = now.Sub(te.StartTime)
	te.UpdatedAt = now
}

func (te *TimeEntry) IsActive() bool {
	return te.EndTime == nil
}

func (te *TimeEntry) GetDuration() time.Duration {
	if te.EndTime != nil {
		return te.Duration
	}
	return time.Since(te.StartTime)
}

func (te *TimeEntry) GetFormattedDuration() string {
	duration := te.GetDuration()
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
