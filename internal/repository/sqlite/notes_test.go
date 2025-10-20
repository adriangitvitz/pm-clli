package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adriannajera/project-manager-cli/internal/domain"
	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func createTestTask(t *testing.T, repo *TaskRepository) *domain.Task {
	task := &domain.Task{
		ID:          uuid.New().String(),
		Title:       "Test Task for Notes",
		Description: "Testing note linking functionality",
		Status:      domain.StatusTodo,
		Priority:    domain.PriorityNormal,
		ProjectID:   "",
		Tags:        []string{"test", "notes"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()
	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	return task
}

func TestTaskNoteLink(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)
	ctx := context.Background()

	// Create a test task
	task := createTestTask(t, repo)

	// Verify task has no note initially
	if task.HasNote {
		t.Error("New task should not have a note")
	}
	if task.NoteID != nil {
		t.Error("New task should have nil NoteID")
	}

	// Link a note to the task
	noteID := uuid.New().String()
	notePath := "/home/user/.debug-notes/test-note.md"
	noteCreatedAt := time.Now().Add(-1 * time.Hour)
	noteUpdatedAt := time.Now()

	task.NoteID = &noteID
	task.NotePath = &notePath
	task.HasNote = true
	task.NoteCreatedAt = &noteCreatedAt
	task.NoteUpdatedAt = &noteUpdatedAt
	task.UpdatedAt = time.Now()

	err := repo.Update(ctx, task)
	if err != nil {
		t.Fatalf("Failed to link note to task: %v", err)
	}

	// Retrieve task and verify note link
	retrievedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve task: %v", err)
	}

	if !retrievedTask.HasNote {
		t.Error("Task should have HasNote = true")
	}

	if retrievedTask.NoteID == nil {
		t.Fatal("Task NoteID should not be nil")
	}
	if *retrievedTask.NoteID != noteID {
		t.Errorf("Expected NoteID %s, got %s", noteID, *retrievedTask.NoteID)
	}

	if retrievedTask.NotePath == nil {
		t.Fatal("Task NotePath should not be nil")
	}
	if *retrievedTask.NotePath != notePath {
		t.Errorf("Expected NotePath %s, got %s", notePath, *retrievedTask.NotePath)
	}

	if retrievedTask.NoteCreatedAt == nil {
		t.Fatal("Task NoteCreatedAt should not be nil")
	}
	if retrievedTask.NoteUpdatedAt == nil {
		t.Fatal("Task NoteUpdatedAt should not be nil")
	}

	// Verify timestamps are preserved (within 1 second tolerance)
	if retrievedTask.NoteCreatedAt.Unix() != noteCreatedAt.Unix() {
		t.Errorf("NoteCreatedAt timestamp mismatch")
	}
	if retrievedTask.NoteUpdatedAt.Unix() != noteUpdatedAt.Unix() {
		t.Errorf("NoteUpdatedAt timestamp mismatch")
	}
}

func TestTaskNoteUnlink(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)
	ctx := context.Background()

	// Create a test task with a linked note
	task := createTestTask(t, repo)

	noteID := uuid.New().String()
	notePath := "/home/user/.debug-notes/test-note.md"
	noteCreatedAt := time.Now().Add(-1 * time.Hour)
	noteUpdatedAt := time.Now()

	task.NoteID = &noteID
	task.NotePath = &notePath
	task.HasNote = true
	task.NoteCreatedAt = &noteCreatedAt
	task.NoteUpdatedAt = &noteUpdatedAt
	task.UpdatedAt = time.Now()

	err := repo.Update(ctx, task)
	if err != nil {
		t.Fatalf("Failed to link note to task: %v", err)
	}

	// Verify note is linked
	retrievedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve task: %v", err)
	}
	if !retrievedTask.HasNote {
		t.Fatal("Task should have a linked note")
	}

	// Unlink the note
	retrievedTask.NoteID = nil
	retrievedTask.NotePath = nil
	retrievedTask.HasNote = false
	retrievedTask.NoteCreatedAt = nil
	retrievedTask.NoteUpdatedAt = nil
	retrievedTask.UpdatedAt = time.Now()

	err = repo.Update(ctx, retrievedTask)
	if err != nil {
		t.Fatalf("Failed to unlink note from task: %v", err)
	}

	// Verify note is unlinked
	unlinkedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve unlinked task: %v", err)
	}

	if unlinkedTask.HasNote {
		t.Error("Task should not have a linked note")
	}
	if unlinkedTask.NoteID != nil {
		t.Error("Task NoteID should be nil")
	}
	if unlinkedTask.NotePath != nil {
		t.Error("Task NotePath should be nil")
	}
	if unlinkedTask.NoteCreatedAt != nil {
		t.Error("Task NoteCreatedAt should be nil")
	}
	if unlinkedTask.NoteUpdatedAt != nil {
		t.Error("Task NoteUpdatedAt should be nil")
	}
}

func TestTaskWithNoteQuery(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)
	ctx := context.Background()

	// Create multiple tasks, some with notes, some without
	task1 := createTestTask(t, repo)
	_ = createTestTask(t, repo) // task2 - intentionally without note
	task3 := createTestTask(t, repo)

	// Link notes to task1 and task3
	noteID1 := uuid.New().String()
	notePath1 := "/home/user/.debug-notes/note1.md"
	task1.NoteID = &noteID1
	task1.NotePath = &notePath1
	task1.HasNote = true
	task1.UpdatedAt = time.Now()
	repo.Update(ctx, task1)

	noteID3 := uuid.New().String()
	notePath3 := "/home/user/.debug-notes/note3.md"
	task3.NoteID = &noteID3
	task3.NotePath = &notePath3
	task3.HasNote = true
	task3.UpdatedAt = time.Now()
	repo.Update(ctx, task3)

	// Query all tasks
	allTasks, err := repo.List(ctx, domain.TaskFilter{})
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(allTasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(allTasks))
	}

	// Verify task1 and task3 have notes, task2 does not
	tasksWithNotes := 0
	tasksWithoutNotes := 0

	for _, task := range allTasks {
		if task.HasNote {
			tasksWithNotes++
			if task.NoteID == nil || task.NotePath == nil {
				t.Error("Task with HasNote=true should have NoteID and NotePath")
			}
		} else {
			tasksWithoutNotes++
			if task.NoteID != nil || task.NotePath != nil {
				t.Error("Task with HasNote=false should not have NoteID or NotePath")
			}
		}
	}

	if tasksWithNotes != 2 {
		t.Errorf("Expected 2 tasks with notes, got %d", tasksWithNotes)
	}
	if tasksWithoutNotes != 1 {
		t.Errorf("Expected 1 task without notes, got %d", tasksWithoutNotes)
	}
}

func TestTaskNoteJSONExport(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)
	ctx := context.Background()

	// Create a task with a linked note
	task := createTestTask(t, repo)

	noteID := uuid.New().String()
	notePath := "/home/user/.debug-notes/test-note.md"
	noteCreatedAt := time.Now().Add(-1 * time.Hour)
	noteUpdatedAt := time.Now()

	task.NoteID = &noteID
	task.NotePath = &notePath
	task.HasNote = true
	task.NoteCreatedAt = &noteCreatedAt
	task.NoteUpdatedAt = &noteUpdatedAt
	task.UpdatedAt = time.Now()

	err := repo.Update(ctx, task)
	if err != nil {
		t.Fatalf("Failed to link note to task: %v", err)
	}

	// Retrieve and verify task can be queried with note fields
	retrievedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve task: %v", err)
	}

	// Verify all note fields are populated for JSON export
	if retrievedTask.NoteID == nil {
		t.Error("NoteID should not be nil for JSON export")
	}
	if retrievedTask.NotePath == nil {
		t.Error("NotePath should not be nil for JSON export")
	}
	if retrievedTask.NoteCreatedAt == nil {
		t.Error("NoteCreatedAt should not be nil for JSON export")
	}
	if retrievedTask.NoteUpdatedAt == nil {
		t.Error("NoteUpdatedAt should not be nil for JSON export")
	}
	if !retrievedTask.HasNote {
		t.Error("HasNote should be true for JSON export")
	}
}

func TestTaskNoteFieldsNullHandling(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)
	ctx := context.Background()

	// Create a task without a note
	task := createTestTask(t, repo)

	// Verify null fields are handled correctly
	retrievedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve task: %v", err)
	}

	if retrievedTask.NoteID != nil {
		t.Error("NoteID should be nil for task without note")
	}
	if retrievedTask.NotePath != nil {
		t.Error("NotePath should be nil for task without note")
	}
	if retrievedTask.NoteCreatedAt != nil {
		t.Error("NoteCreatedAt should be nil for task without note")
	}
	if retrievedTask.NoteUpdatedAt != nil {
		t.Error("NoteUpdatedAt should be nil for task without note")
	}
	if retrievedTask.HasNote {
		t.Error("HasNote should be false for task without note")
	}
}

func TestTaskNotePartialUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)
	ctx := context.Background()

	// Create a task with a note
	task := createTestTask(t, repo)

	noteID := uuid.New().String()
	notePath := "/home/user/.debug-notes/test-note.md"
	noteCreatedAt := time.Now().Add(-1 * time.Hour)
	noteUpdatedAt := time.Now()

	task.NoteID = &noteID
	task.NotePath = &notePath
	task.HasNote = true
	task.NoteCreatedAt = &noteCreatedAt
	task.NoteUpdatedAt = &noteUpdatedAt

	err := repo.Update(ctx, task)
	if err != nil {
		t.Fatalf("Failed to link note to task: %v", err)
	}

	// Update task without changing note fields
	retrievedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve task: %v", err)
	}

	retrievedTask.Title = "Updated Title"
	retrievedTask.UpdatedAt = time.Now()

	err = repo.Update(ctx, retrievedTask)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Verify note fields are preserved
	updatedTask, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated task: %v", err)
	}

	if updatedTask.Title != "Updated Title" {
		t.Error("Task title should be updated")
	}

	if !updatedTask.HasNote {
		t.Error("HasNote should still be true")
	}
	if updatedTask.NoteID == nil || *updatedTask.NoteID != noteID {
		t.Error("NoteID should be preserved")
	}
	if updatedTask.NotePath == nil || *updatedTask.NotePath != notePath {
		t.Error("NotePath should be preserved")
	}
}
