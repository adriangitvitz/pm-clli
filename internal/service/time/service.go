package time

import (
	"context"
	"fmt"
	"time"

	"github.com/adriannajera/project-manager-cli/internal/domain"
)

// Service provides time tracking functionality
type Service struct {
	timeEntryRepo domain.TimeEntryRepository
	taskRepo      domain.TaskRepository
}

// NewService creates a new time tracking service
func NewService(timeEntryRepo domain.TimeEntryRepository, taskRepo domain.TaskRepository) *Service {
	return &Service{
		timeEntryRepo: timeEntryRepo,
		taskRepo:      taskRepo,
	}
}

// StartTimeEntryInput represents input for starting time tracking
type StartTimeEntryInput struct {
	TaskID      string
	Description string
}

// ListOptions represents options for listing time entries
type ListOptions struct {
	TaskID     string
	ProjectID  string
	StartAfter *time.Time
	EndBefore  *time.Time
	Active     *bool
	Limit      int
	Offset     int
}

// TimeReport represents a time tracking report
type TimeReport struct {
	TotalDuration time.Duration
	Entries       []*domain.TimeEntry
	ByTask        map[string]TaskTimeReport
	ByProject     map[string]ProjectTimeReport
	ByDay         map[string]DayTimeReport
}

// TaskTimeReport represents time tracking for a specific task
type TaskTimeReport struct {
	TaskID        string
	TaskTitle     string
	TotalDuration time.Duration
	Entries       []*domain.TimeEntry
}

// ProjectTimeReport represents time tracking for a specific project
type ProjectTimeReport struct {
	ProjectID     string
	ProjectName   string
	TotalDuration time.Duration
	Entries       []*domain.TimeEntry
}

// DayTimeReport represents time tracking for a specific day
type DayTimeReport struct {
	Date          time.Time
	TotalDuration time.Duration
	Entries       []*domain.TimeEntry
}

// StartTimeTracking starts time tracking for a task
func (s *Service) StartTimeTracking(ctx context.Context, input StartTimeEntryInput) (*domain.TimeEntry, error) {
	if input.TaskID == "" {
		return nil, domain.ErrInvalidTaskID
	}

	// Check if there's already an active time entry
	if activeEntry, err := s.timeEntryRepo.GetActive(ctx); err == nil && activeEntry != nil {
		return nil, domain.ErrActiveTimeEntry
	}

	// Verify the task exists
	task, err := s.taskRepo.GetByID(ctx, input.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Create new time entry
	entry := domain.NewTimeEntry(input.TaskID, task.ProjectID, input.Description)

	if err := s.timeEntryRepo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create time entry: %w", err)
	}

	// Update task status to "doing" if it's not already
	if task.Status != domain.StatusDoing {
		task.Start()
		if err := s.taskRepo.Update(ctx, task); err != nil {
			// Log error but don't fail the time tracking start
			fmt.Printf("Warning: failed to update task status: %v\n", err)
		}
	}

	return entry, nil
}

// StopTimeTracking stops the currently active time tracking
func (s *Service) StopTimeTracking(ctx context.Context) (*domain.TimeEntry, error) {
	// Get the active time entry
	entry, err := s.timeEntryRepo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active time entry: %w", err)
	}

	if entry == nil {
		return nil, domain.ErrNoActiveTimeEntry
	}

	// Stop the time entry
	entry.Stop()

	if err := s.timeEntryRepo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update time entry: %w", err)
	}

	return entry, nil
}

// GetActiveTimeEntry returns the currently active time entry, if any
func (s *Service) GetActiveTimeEntry(ctx context.Context) (*domain.TimeEntry, error) {
	entry, err := s.timeEntryRepo.GetActive(ctx)
	if err == domain.ErrNoActiveTimeEntry {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active time entry: %w", err)
	}

	return entry, nil
}

// ListTimeEntries retrieves time entries based on the provided options
func (s *Service) ListTimeEntries(ctx context.Context, options ListOptions) ([]*domain.TimeEntry, error) {
	filter := domain.TimeEntryFilter{
		TaskID:     options.TaskID,
		ProjectID:  options.ProjectID,
		StartAfter: options.StartAfter,
		EndBefore:  options.EndBefore,
		Active:     options.Active,
		Limit:      options.Limit,
		Offset:     options.Offset,
	}

	entries, err := s.timeEntryRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list time entries: %w", err)
	}

	return entries, nil
}

// GetTimeEntriesByTask retrieves all time entries for a specific task
func (s *Service) GetTimeEntriesByTask(ctx context.Context, taskID string) ([]*domain.TimeEntry, error) {
	if taskID == "" {
		return nil, domain.ErrInvalidTaskID
	}

	entries, err := s.timeEntryRepo.GetByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries by task: %w", err)
	}

	return entries, nil
}

// GetTimeEntriesByProject retrieves all time entries for a specific project
func (s *Service) GetTimeEntriesByProject(ctx context.Context, projectID string) ([]*domain.TimeEntry, error) {
	if projectID == "" {
		return nil, domain.ErrInvalidProjectID
	}

	entries, err := s.timeEntryRepo.GetByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries by project: %w", err)
	}

	return entries, nil
}

// DeleteTimeEntry deletes a time entry
func (s *Service) DeleteTimeEntry(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("time entry ID cannot be empty")
	}

	if err := s.timeEntryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete time entry: %w", err)
	}

	return nil
}

// UpdateTimeEntry updates an existing time entry
func (s *Service) UpdateTimeEntry(ctx context.Context, entry *domain.TimeEntry) error {
	if entry == nil {
		return fmt.Errorf("time entry cannot be nil")
	}

	entry.UpdatedAt = time.Now()

	if err := s.timeEntryRepo.Update(ctx, entry); err != nil {
		return fmt.Errorf("failed to update time entry: %w", err)
	}

	return nil
}

// GenerateReport generates a time tracking report for the given time range
func (s *Service) GenerateReport(ctx context.Context, startDate, endDate time.Time) (*TimeReport, error) {
	entries, err := s.ListTimeEntries(ctx, ListOptions{
		StartAfter: &startDate,
		EndBefore:  &endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries for report: %w", err)
	}

	report := &TimeReport{
		Entries:   entries,
		ByTask:    make(map[string]TaskTimeReport),
		ByProject: make(map[string]ProjectTimeReport),
		ByDay:     make(map[string]DayTimeReport),
	}

	// Calculate totals and group by task, project, and day
	for _, entry := range entries {
		duration := entry.GetDuration()
		report.TotalDuration += duration

		// Group by task
		taskKey := entry.TaskID
		if taskReport, exists := report.ByTask[taskKey]; exists {
			taskReport.TotalDuration += duration
			taskReport.Entries = append(taskReport.Entries, entry)
			report.ByTask[taskKey] = taskReport
		} else {
			// Get task title
			task, err := s.taskRepo.GetByID(ctx, entry.TaskID)
			taskTitle := entry.TaskID // fallback
			if err == nil {
				taskTitle = task.Title
			}

			report.ByTask[taskKey] = TaskTimeReport{
				TaskID:        entry.TaskID,
				TaskTitle:     taskTitle,
				TotalDuration: duration,
				Entries:       []*domain.TimeEntry{entry},
			}
		}

		// Group by project
		projectKey := entry.ProjectID
		if projectReport, exists := report.ByProject[projectKey]; exists {
			projectReport.TotalDuration += duration
			projectReport.Entries = append(projectReport.Entries, entry)
			report.ByProject[projectKey] = projectReport
		} else {
			report.ByProject[projectKey] = ProjectTimeReport{
				ProjectID:     entry.ProjectID,
				ProjectName:   entry.ProjectID, // TODO: Get actual project name
				TotalDuration: duration,
				Entries:       []*domain.TimeEntry{entry},
			}
		}

		// Group by day
		dayKey := entry.StartTime.Format("2006-01-02")
		if dayReport, exists := report.ByDay[dayKey]; exists {
			dayReport.TotalDuration += duration
			dayReport.Entries = append(dayReport.Entries, entry)
			report.ByDay[dayKey] = dayReport
		} else {
			report.ByDay[dayKey] = DayTimeReport{
				Date:          entry.StartTime.Truncate(24 * time.Hour),
				TotalDuration: duration,
				Entries:       []*domain.TimeEntry{entry},
			}
		}
	}

	return report, nil
}

// GetTodayReport generates a report for today's time tracking
func (s *Service) GetTodayReport(ctx context.Context) (*TimeReport, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.GenerateReport(ctx, startOfDay, endOfDay)
}

// GetWeekReport generates a report for this week's time tracking
func (s *Service) GetWeekReport(ctx context.Context) (*TimeReport, error) {
	now := time.Now()

	// Calculate start of week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, 1-weekday).Truncate(24 * time.Hour)
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return s.GenerateReport(ctx, startOfWeek, endOfWeek)
}

// GetMonthReport generates a report for this month's time tracking
func (s *Service) GetMonthReport(ctx context.Context) (*TimeReport, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return s.GenerateReport(ctx, startOfMonth, endOfMonth)
}

// FormatDuration formats a duration in a human-readable way
func (s *Service) FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}