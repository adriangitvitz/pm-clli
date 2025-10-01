package domain

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	title := "Test Task"
	description := "Test Description"

	task := NewTask(title, description)

	if task.ID == "" {
		t.Error("Expected task ID to be generated")
	}

	if task.Title != title {
		t.Errorf("Expected title %s, got %s", title, task.Title)
	}

	if task.Description != description {
		t.Errorf("Expected description %s, got %s", description, task.Description)
	}

	if task.Status != StatusTodo {
		t.Errorf("Expected status %s, got %s", StatusTodo, task.Status)
	}

	if task.Priority != PriorityNormal {
		t.Errorf("Expected priority %d, got %d", PriorityNormal, task.Priority)
	}

	if len(task.Tags) != 0 {
		t.Errorf("Expected empty tags slice, got %v", task.Tags)
	}

	if task.Metadata == nil {
		t.Error("Expected metadata map to be initialized")
	}
}

func TestTaskComplete(t *testing.T) {
	task := NewTask("Test Task", "Description")
	originalUpdatedAt := task.UpdatedAt

	// Wait a moment to ensure timestamp difference
	time.Sleep(time.Millisecond)

	task.Complete()

	if task.Status != StatusDone {
		t.Errorf("Expected status %s, got %s", StatusDone, task.Status)
	}

	if task.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}

	if !task.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestTaskStart(t *testing.T) {
	task := NewTask("Test Task", "Description")
	originalUpdatedAt := task.UpdatedAt

	time.Sleep(time.Millisecond)

	task.Start()

	if task.Status != StatusDoing {
		t.Errorf("Expected status %s, got %s", StatusDoing, task.Status)
	}

	if !task.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestTaskAddTag(t *testing.T) {
	task := NewTask("Test Task", "Description")
	tag := "urgent"

	task.AddTag(tag)

	if len(task.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(task.Tags))
	}

	if task.Tags[0] != tag {
		t.Errorf("Expected tag %s, got %s", tag, task.Tags[0])
	}

	// Test adding duplicate tag
	task.AddTag(tag)
	if len(task.Tags) != 1 {
		t.Errorf("Expected 1 tag after adding duplicate, got %d", len(task.Tags))
	}
}

func TestTaskRemoveTag(t *testing.T) {
	task := NewTask("Test Task", "Description")
	tag1 := "urgent"
	tag2 := "bug"

	task.AddTag(tag1)
	task.AddTag(tag2)

	task.RemoveTag(tag1)

	if len(task.Tags) != 1 {
		t.Errorf("Expected 1 tag after removal, got %d", len(task.Tags))
	}

	if task.Tags[0] != tag2 {
		t.Errorf("Expected remaining tag %s, got %s", tag2, task.Tags[0])
	}

	if task.HasTag(tag1) {
		t.Errorf("Expected tag %s to be removed", tag1)
	}
}

func TestTaskIsOverdue(t *testing.T) {
	task := NewTask("Test Task", "Description")

	// Task without due date should not be overdue
	if task.IsOverdue() {
		t.Error("Expected task without due date to not be overdue")
	}

	// Task with future due date should not be overdue
	futureDate := time.Now().Add(24 * time.Hour)
	task.DueDate = &futureDate
	if task.IsOverdue() {
		t.Error("Expected task with future due date to not be overdue")
	}

	// Task with past due date should be overdue
	pastDate := time.Now().Add(-24 * time.Hour)
	task.DueDate = &pastDate
	if !task.IsOverdue() {
		t.Error("Expected task with past due date to be overdue")
	}

	// Completed task should not be overdue even with past due date
	task.Complete()
	if task.IsOverdue() {
		t.Error("Expected completed task to not be overdue")
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityLow, "low"},
		{PriorityNormal, "normal"},
		{PriorityHigh, "high"},
		{PriorityCritical, "critical"},
		{Priority(999), "normal"}, // Invalid priority should default to normal
	}

	for _, test := range tests {
		result := test.priority.String()
		if result != test.expected {
			t.Errorf("Expected %s for priority %d, got %s", test.expected, test.priority, result)
		}
	}
}