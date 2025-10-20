package export

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/adriannajera/project-manager-cli/internal/domain"
)

// Service provides export functionality
type Service struct {
	taskRepo      domain.TaskRepository
	projectRepo   domain.ProjectRepository
	timeEntryRepo domain.TimeEntryRepository
}

// NewService creates a new export service
func NewService(
	taskRepo domain.TaskRepository,
	projectRepo domain.ProjectRepository,
	timeEntryRepo domain.TimeEntryRepository,
) *Service {
	return &Service{
		taskRepo:      taskRepo,
		projectRepo:   projectRepo,
		timeEntryRepo: timeEntryRepo,
	}
}

// ExportFormat represents supported export formats
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
	FormatICAL ExportFormat = "ical"
)

// ExportTasksToJSON exports tasks to JSON format
func (s *Service) ExportTasksToJSON(ctx context.Context, filter domain.TaskFilter) ([]byte, error) {
	tasks, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	return json.MarshalIndent(tasks, "", "  ")
}

// ExportTasksToCSV exports tasks to CSV format
func (s *Service) ExportTasksToCSV(ctx context.Context, filter domain.TaskFilter) ([]byte, error) {
	tasks, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"ID", "Title", "Description", "Status", "Priority", "Project ID",
		"Tags", "Due Date", "Created At", "Updated At", "Completed At",
		"Note ID", "Note Path", "Has Note", "Note Created", "Note Updated",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write tasks
	for _, task := range tasks {
		var dueDate, completedAt string
		if task.DueDate != nil {
			dueDate = task.DueDate.Format("2006-01-02 15:04:05")
		}
		if task.CompletedAt != nil {
			completedAt = task.CompletedAt.Format("2006-01-02 15:04:05")
		}

		// Handle note fields
		var noteID, notePath, noteCreated, noteUpdated string
		var hasNote string
		if task.NoteID != nil {
			noteID = *task.NoteID
		}
		if task.NotePath != nil {
			notePath = *task.NotePath
		}
		if task.HasNote {
			hasNote = "true"
		} else {
			hasNote = "false"
		}
		if task.NoteCreatedAt != nil {
			noteCreated = task.NoteCreatedAt.Format("2006-01-02 15:04:05")
		}
		if task.NoteUpdatedAt != nil {
			noteUpdated = task.NoteUpdatedAt.Format("2006-01-02 15:04:05")
		}

		record := []string{
			task.ID,
			task.Title,
			task.Description,
			string(task.Status),
			task.Priority.String(),
			task.ProjectID,
			strings.Join(task.Tags, ";"),
			dueDate,
			task.CreatedAt.Format("2006-01-02 15:04:05"),
			task.UpdatedAt.Format("2006-01-02 15:04:05"),
			completedAt,
			noteID,
			notePath,
			hasNote,
			noteCreated,
			noteUpdated,
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return []byte(buf.String()), nil
}

// ExportTasksToICAL exports tasks with due dates to iCal format
func (s *Service) ExportTasksToICAL(ctx context.Context, filter domain.TaskFilter) ([]byte, error) {
	tasks, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var buf strings.Builder

	// iCal header
	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//Project Manager CLI//NONSGML v1.0//EN\r\n")
	buf.WriteString("CALSCALE:GREGORIAN\r\n")

	// Export tasks with due dates as events
	for _, task := range tasks {
		if task.DueDate != nil {
			buf.WriteString("BEGIN:VEVENT\r\n")
			buf.WriteString(fmt.Sprintf("UID:%s@pm-cli\r\n", task.ID))
			buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", time.Now().UTC().Format("20060102T150405Z")))
			buf.WriteString(fmt.Sprintf("DTSTART:%s\r\n", task.DueDate.UTC().Format("20060102T150405Z")))
			buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICALText(task.Title)))

			if task.Description != "" {
				buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICALText(task.Description)))
			}

			buf.WriteString(fmt.Sprintf("STATUS:%s\r\n", mapTaskStatusToICAL(task.Status)))
			buf.WriteString("END:VEVENT\r\n")
		}
	}

	buf.WriteString("END:VCALENDAR\r\n")

	return []byte(buf.String()), nil
}

// ExportProjectsToJSON exports projects to JSON format
func (s *Service) ExportProjectsToJSON(ctx context.Context, filter domain.ProjectFilter) ([]byte, error) {
	projects, err := s.projectRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return json.MarshalIndent(projects, "", "  ")
}

// ExportTimeEntriesToJSON exports time entries to JSON format
func (s *Service) ExportTimeEntriesToJSON(ctx context.Context, filter domain.TimeEntryFilter) ([]byte, error) {
	entries, err := s.timeEntryRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries: %w", err)
	}

	return json.MarshalIndent(entries, "", "  ")
}

// ExportTimeEntriesToCSV exports time entries to CSV format
func (s *Service) ExportTimeEntriesToCSV(ctx context.Context, filter domain.TimeEntryFilter) ([]byte, error) {
	entries, err := s.timeEntryRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries: %w", err)
	}

	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"ID", "Task ID", "Project ID", "Description", "Start Time",
		"End Time", "Duration (seconds)", "Created At", "Updated At",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write time entries
	for _, entry := range entries {
		var endTime string
		if entry.EndTime != nil {
			endTime = entry.EndTime.Format("2006-01-02 15:04:05")
		}

		record := []string{
			entry.ID,
			entry.TaskID,
			entry.ProjectID,
			entry.Description,
			entry.StartTime.Format("2006-01-02 15:04:05"),
			endTime,
			fmt.Sprintf("%.0f", entry.GetDuration().Seconds()),
			entry.CreatedAt.Format("2006-01-02 15:04:05"),
			entry.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return []byte(buf.String()), nil
}

// escapeICALText escapes special characters in iCal text fields
func escapeICALText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "\\r")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, ";", "\\;")
	return text
}

// mapTaskStatusToICAL maps task status to iCal status
func mapTaskStatusToICAL(status domain.TaskStatus) string {
	switch status {
	case domain.StatusDone:
		return "COMPLETED"
	case domain.StatusDoing:
		return "IN-PROCESS"
	case domain.StatusBlocked:
		return "CANCELLED"
	default:
		return "NEEDS-ACTION"
	}
}