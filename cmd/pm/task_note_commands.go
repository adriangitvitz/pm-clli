package main

import (
	"context"
	"fmt"
	"time"

	"github.com/adriannajera/project-manager-cli/internal/notes"
	"github.com/adriannajera/project-manager-cli/internal/repository/sqlite"
)

func handleTaskNoteCommand(taskRepo *sqlite.TaskRepository, args []string) error {
	if len(args) == 0 {
		return showTaskNoteHelp()
	}

	ctx := context.Background()
	subcommand := args[0]

	switch subcommand {
	case "help", "--help", "-h":
		return showTaskNoteHelp()
	case "link":
		if len(args) < 3 {
			return fmt.Errorf("task note link requires <task-id> and <note-id>")
		}
		return linkTaskNote(ctx, taskRepo, args[1], args[2])
	case "unlink":
		if len(args) < 2 {
			return fmt.Errorf("task note unlink requires <task-id>")
		}
		return unlinkTaskNote(ctx, taskRepo, args[1])
	case "show":
		if len(args) < 2 {
			return fmt.Errorf("task note show requires <task-id>")
		}
		return showTaskNote(ctx, taskRepo, args[1])
	default:
		return fmt.Errorf("unknown task note subcommand: %s", subcommand)
	}
}

func linkTaskNote(ctx context.Context, taskRepo *sqlite.TaskRepository, taskID, noteID string) error {
	// Get the task
	task, err := taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Find the note file by ID
	notePath, err := notes.FindNoteByID(noteID)
	if err != nil {
		return fmt.Errorf("failed to find note: %w", err)
	}

	// Parse note frontmatter to get metadata
	noteMeta, err := notes.ParseNoteFrontmatter(notePath)
	if err != nil {
		return fmt.Errorf("failed to parse note metadata: %w", err)
	}

	// Update task with note information
	task.NoteID = &noteMeta.ID
	task.NotePath = &noteMeta.Path
	task.HasNote = true
	task.NoteCreatedAt = &noteMeta.CreatedAt
	task.NoteUpdatedAt = &noteMeta.UpdatedAt
	task.UpdatedAt = time.Now()

	// Save the updated task
	if err := taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Linked note %s to task %s\n", noteID, taskID)
	fmt.Printf("  Note path: %s\n", notePath)
	fmt.Printf("  Note created: %s\n", noteMeta.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Note updated: %s\n", noteMeta.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func unlinkTaskNote(ctx context.Context, taskRepo *sqlite.TaskRepository, taskID string) error {
	// Get the task
	task, err := taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task has a note
	if !task.HasNote {
		return fmt.Errorf("task %s does not have a linked note", taskID)
	}

	// Clear note fields
	task.NoteID = nil
	task.NotePath = nil
	task.HasNote = false
	task.NoteCreatedAt = nil
	task.NoteUpdatedAt = nil
	task.UpdatedAt = time.Now()

	// Save the updated task
	if err := taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Unlinked note from task %s\n", taskID)

	return nil
}

func showTaskNote(ctx context.Context, taskRepo *sqlite.TaskRepository, taskID string) error {
	// Get the task
	task, err := taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Display note information
	fmt.Printf("Task: %s (%s)\n", task.Title, task.ID)
	fmt.Println()

	if !task.HasNote {
		fmt.Println("No note linked to this task")
		return nil
	}

	fmt.Println("Linked Note:")
	if task.NoteID != nil {
		fmt.Printf("  Note ID: %s\n", *task.NoteID)
	}
	if task.NotePath != nil {
		fmt.Printf("  Note Path: %s\n", *task.NotePath)
	}
	fmt.Printf("  Has Note: %t\n", task.HasNote)
	if task.NoteCreatedAt != nil {
		fmt.Printf("  Note Created: %s\n", task.NoteCreatedAt.Format("2006-01-02 15:04:05"))
	}
	if task.NoteUpdatedAt != nil {
		fmt.Printf("  Note Updated: %s\n", task.NoteUpdatedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func showTaskNoteHelp() error {
	helpText := `Task Note Management Commands

USAGE:
  pm task note [subcommand] [args]

SUBCOMMANDS:
  link <task-id> <note-id>    Link a note to a task
  unlink <task-id>            Remove note link from a task
  show <task-id>              Show linked note information

EXAMPLES:
  pm task note link abc123 def456-789a-bcde-f012-3456789abcde
  pm task note show abc123
  pm task note unlink abc123

DESCRIPTION:
  Note linking allows you to associate dn-tui debug notes with tasks.
  When you link a note, the task will store the note's ID, path, and
  creation/modification timestamps.

  The note-id should be a UUID from a dn-tui note's frontmatter.
`
	fmt.Println(helpText)
	return nil
}
