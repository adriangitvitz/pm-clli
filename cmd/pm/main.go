package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriannajera/project-manager-cli/internal/domain"
	"github.com/adriannajera/project-manager-cli/internal/repository/sqlite"
	"github.com/adriannajera/project-manager-cli/internal/service/task"
	"github.com/adriannajera/project-manager-cli/internal/service/project"
	"github.com/adriannajera/project-manager-cli/internal/service/git"
	timeService "github.com/adriannajera/project-manager-cli/internal/service/time"
	"github.com/adriannajera/project-manager-cli/internal/service/export"
	"github.com/adriannajera/project-manager-cli/internal/ui/models"
	"github.com/adriannajera/project-manager-cli/pkg/config"
	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	db, err := sqlite.NewDB(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	taskRepo := sqlite.NewTaskRepository(db)
	projectRepo := sqlite.NewProjectRepository(db)
	timeEntryRepo := sqlite.NewTimeEntryRepository(db)

	// Initialize Git repository
	gitRepo := git.NewGitRepository()

	// Check if we should run CLI commands or TUI
	if len(os.Args) > 1 {
		return runCLI(taskRepo, projectRepo, timeEntryRepo, gitRepo, cfg)
	}

	// Run TUI application
	return runTUI(taskRepo, projectRepo, timeEntryRepo, gitRepo)
}

func runCLI(taskRepo *sqlite.TaskRepository, projectRepo *sqlite.ProjectRepository, timeEntryRepo *sqlite.TimeEntryRepository, gitRepo *git.GitRepository, cfg *domain.Config) error {
	// Initialize services
	taskService := task.NewService(taskRepo, gitRepo)
	projectService := project.NewService(projectRepo)

	// Simple command routing
	command := os.Args[1]

	switch command {
	case "task":
		return handleTaskCommand(taskService, projectRepo, os.Args[2:])
	case "project":
		return handleProjectCommand(projectService, os.Args[2:])
	case "time":
		timeSvc := timeService.NewService(timeEntryRepo, taskRepo)
		return handleTimeCommand(timeSvc, os.Args[2:])
	case "export":
		exportService := export.NewService(taskRepo, projectRepo, timeEntryRepo)
		return handleExportCommand(exportService, os.Args[2:])
	case "config":
		return handleConfigCommand(cfg, os.Args[2:])
	case "git":
		return handleGitCommand(gitRepo, os.Args[2:])
	case "version":
		fmt.Println("Project Manager CLI v1.0.0")
		return nil
	case "help", "--help", "-h":
		return showHelp()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func runTUI(taskRepo *sqlite.TaskRepository, projectRepo *sqlite.ProjectRepository, timeEntryRepo *sqlite.TimeEntryRepository, gitRepo *git.GitRepository) error {
	// Create the Bubble Tea application
	app := models.NewAppModel(taskRepo, projectRepo, timeEntryRepo, gitRepo)

	// Start the TUI
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func handleTaskCommand(taskService *task.Service, projectRepo *sqlite.ProjectRepository, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task command requires a subcommand")
	}

	ctx := context.Background()
	subcommand := args[0]

	switch subcommand {
	case "list", "ls":
		return listTasks(ctx, taskService, projectRepo, args[1:])
	case "add", "create":
		if len(args) < 2 {
			return fmt.Errorf("task add requires a title")
		}
		return addTask(ctx, taskService, projectRepo, args[1:])
	case "update":
		if len(args) < 2 {
			return fmt.Errorf("task update requires a task ID")
		}
		return updateTask(ctx, taskService, projectRepo, args[1:])
	case "complete":
		if len(args) < 2 {
			return fmt.Errorf("task complete requires a task ID")
		}
		return completeTask(ctx, taskService, args[1])
	case "delete", "rm":
		if len(args) < 2 {
			return fmt.Errorf("task delete requires a task ID")
		}
		return deleteTask(ctx, taskService, args[1])
	default:
		return fmt.Errorf("unknown task subcommand: %s", subcommand)
	}
}

func handleProjectCommand(projectService *project.Service, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("project command requires a subcommand")
	}

	ctx := context.Background()
	subcommand := args[0]

	switch subcommand {
	case "list", "ls":
		return listProjects(ctx, projectService)
	case "add", "create":
		if len(args) < 2 {
			return fmt.Errorf("project add requires a name")
		}
		return addProject(ctx, projectService, args[1])
	case "delete", "rm":
		if len(args) < 2 {
			return fmt.Errorf("project delete requires a project ID")
		}
		return deleteProject(ctx, projectService, args[1])
	default:
		return fmt.Errorf("unknown project subcommand: %s", subcommand)
	}
}

func handleTimeCommand(timeSvc *timeService.Service, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("time command requires a subcommand")
	}

	ctx := context.Background()
	subcommand := args[0]

	switch subcommand {
	case "start":
		return startTimeTracking(ctx, timeSvc, args[1:])
	case "stop":
		return stopTimeTracking(ctx, timeSvc)
	case "report":
		return generateTimeReport(ctx, timeSvc, args[1:])
	case "list":
		return listTimeEntries(ctx, timeSvc, args[1:])
	default:
		return fmt.Errorf("unknown time subcommand: %s", subcommand)
	}
}

func listTasks(ctx context.Context, taskService *task.Service, projectRepo *sqlite.ProjectRepository, args []string) error {
	// Parse flags
	options := task.ListOptions{}
	for i := 0; i < len(args); i++ {
		if args[i] == "--project" && i+1 < len(args) {
			projectNameOrID := args[i+1]
			// Try to look up project by name first
			if proj, err := projectRepo.GetByName(ctx, projectNameOrID); err == nil {
				options.ProjectID = proj.ID
			} else {
				// If not found by name, assume it's an ID
				options.ProjectID = projectNameOrID
			}
			i++ // Skip the next argument as it's the value
		}
	}

	tasks, err := taskService.ListTasks(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	fmt.Println("Tasks:")
	for _, t := range tasks {
		status := "[ ]"
		switch t.Status {
		case "doing":
			status = "[~]"
		case "done":
			status = "[x]"
		case "blocked":
			status = "[!]"
		}

		priority := ""
		switch t.Priority {
		case 0: // Low
			priority = "LOW"
		case 1: // Normal
			priority = "NORM"
		case 2: // High
			priority = "HIGH"
		case 3: // Critical
			priority = "CRIT"
		}

		changelistStr := ""
		if t.Changelist != "" {
			changelistStr = fmt.Sprintf(" (%s)", t.Changelist)
		}

		fmt.Printf("  %s [%s] %s%s (ID: %s)\n", status, priority, t.Title, changelistStr, t.ID)
	}

	return nil
}

func addTask(ctx context.Context, taskService *task.Service, projectRepo *sqlite.ProjectRepository, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task add requires a title")
	}

	input := task.CreateTaskInput{
		Title: args[0],
	}

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--priority":
			if i+1 < len(args) {
				i++
				switch args[i] {
				case "low":
					input.Priority = 1
				case "medium":
					input.Priority = 2
				case "high":
					input.Priority = 3
				}
			}
		case "--tags":
			if i+1 < len(args) {
				i++
				input.Tags = strings.Split(args[i], ",")
			}
		case "--changelist", "--cl":
			if i+1 < len(args) {
				i++
				input.Changelist = args[i]
			}
		case "--due":
			if i+1 < len(args) {
				i++
				input.DueDate = args[i]
			}
		case "--project":
			if i+1 < len(args) {
				i++
				projectNameOrID := args[i]
				// Try to look up project by name first
				if proj, err := projectRepo.GetByName(ctx, projectNameOrID); err == nil {
					input.ProjectID = proj.ID
				} else {
					// If not found by name, assume it's an ID
					input.ProjectID = projectNameOrID
				}
			}
		case "--description":
			if i+1 < len(args) {
				i++
				input.Description = args[i]
			}
		}
	}

	createdTask, err := taskService.CreateTask(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("Created task: %s (ID: %s)\n", createdTask.Title, createdTask.ID)
	return nil
}

func updateTask(ctx context.Context, taskService *task.Service, projectRepo *sqlite.ProjectRepository, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task update requires a task ID")
	}

	taskID := args[0]
	input := task.UpdateTaskInput{
		ID: taskID,
	}

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				i++
				title := args[i]
				input.Title = &title
			}
		case "--description":
			if i+1 < len(args) {
				i++
				desc := args[i]
				input.Description = &desc
			}
		case "--status":
			if i+1 < len(args) {
				i++
				var status domain.TaskStatus
				switch args[i] {
				case "todo":
					status = domain.StatusTodo
				case "doing":
					status = domain.StatusDoing
				case "done":
					status = domain.StatusDone
				case "blocked":
					status = domain.StatusBlocked
				default:
					return fmt.Errorf("invalid status: %s (must be todo, doing, done, or blocked)", args[i])
				}
				input.Status = &status
			}
		case "--priority":
			if i+1 < len(args) {
				i++
				var priority domain.Priority
				switch args[i] {
				case "low":
					priority = domain.PriorityLow
				case "normal":
					priority = domain.PriorityNormal
				case "high":
					priority = domain.PriorityHigh
				case "critical":
					priority = domain.PriorityCritical
				default:
					return fmt.Errorf("invalid priority: %s (must be low, normal, high, or critical)", args[i])
				}
				input.Priority = &priority
			}
		case "--changelist", "--cl":
			if i+1 < len(args) {
				i++
				changelist := args[i]
				input.Changelist = &changelist
			}
		case "--due":
			if i+1 < len(args) {
				i++
				dueDate := args[i]
				input.DueDate = &dueDate
			}
		case "--project":
			if i+1 < len(args) {
				i++
				projectNameOrID := args[i]
				// Try to look up project by name first
				if proj, err := projectRepo.GetByName(ctx, projectNameOrID); err == nil {
					input.ProjectID = &proj.ID
				} else {
					// If not found by name, assume it's an ID
					input.ProjectID = &projectNameOrID
				}
			}
		}
	}

	updatedTask, err := taskService.UpdateTask(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Updated task: %s (ID: %s)\n", updatedTask.Title, updatedTask.ID)
	return nil
}

func completeTask(ctx context.Context, taskService *task.Service, taskID string) error {
	err := taskService.CompleteTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	fmt.Printf("Task %s completed\n", taskID)
	return nil
}

func deleteTask(ctx context.Context, taskService *task.Service, taskID string) error {
	err := taskService.DeleteTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("Task %s deleted\n", taskID)
	return nil
}

func listProjects(ctx context.Context, projectService *project.Service) error {
	projects, err := projectService.ListProjects(ctx, project.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	fmt.Println("Projects:")
	for _, p := range projects {
		fmt.Printf("  %s (ID: %s)\n", p.Name, p.ID)
	}

	return nil
}

func addProject(ctx context.Context, projectService *project.Service, name string) error {
	input := project.CreateProjectInput{
		Name: name,
	}

	createdProject, err := projectService.CreateProject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("Created project: %s (ID: %s)\n", createdProject.Name, createdProject.ID)
	return nil
}

func deleteProject(ctx context.Context, projectService *project.Service, projectID string) error {
	err := projectService.DeleteProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	fmt.Printf("Project %s deleted (associated tasks and time entries also deleted)\n", projectID)
	return nil
}

// Time tracking handlers
func startTimeTracking(ctx context.Context, timeSvc *timeService.Service, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("time start requires --task flag")
	}

	var taskID string
	var description string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--task":
			if i+1 < len(args) {
				i++
				taskID = args[i]
			}
		case "--description":
			if i+1 < len(args) {
				i++
				description = args[i]
			}
		}
	}

	if taskID == "" {
		return fmt.Errorf("--task flag is required")
	}

	entry, err := timeSvc.StartTimeTracking(ctx, timeService.StartTimeEntryInput{
		TaskID:      taskID,
		Description: description,
	})
	if err != nil {
		return fmt.Errorf("failed to start time tracking: %w", err)
	}

	fmt.Printf("Started tracking time for task %s (Entry ID: %s)\n", taskID, entry.ID)
	return nil
}

func stopTimeTracking(ctx context.Context, timeSvc *timeService.Service) error {
	entry, err := timeSvc.StopTimeTracking(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop time tracking: %w", err)
	}

	duration := timeSvc.FormatDuration(entry.GetDuration())
	fmt.Printf("Stopped tracking time. Duration: %s (Entry ID: %s)\n", duration, entry.ID)
	return nil
}

func generateTimeReport(ctx context.Context, timeSvc *timeService.Service, args []string) error {
	var report *timeService.TimeReport
	var err error

	// Parse flags
	reportType := "today"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--today":
			reportType = "today"
		case "--week":
			reportType = "week"
		case "--month":
			reportType = "month"
		case "--yesterday":
			reportType = "yesterday"
		}
	}

	switch reportType {
	case "today":
		report, err = timeSvc.GetTodayReport(ctx)
	case "yesterday":
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)
		startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		report, err = timeSvc.GenerateReport(ctx, startOfDay, endOfDay)
	case "week":
		report, err = timeSvc.GetWeekReport(ctx)
	case "month":
		report, err = timeSvc.GetMonthReport(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	fmt.Printf("Time Tracking Report\n")
	fmt.Printf("===================\n")
	fmt.Printf("Total Duration: %s\n\n", timeSvc.FormatDuration(report.TotalDuration))

	if len(report.ByTask) > 0 {
		fmt.Println("By Task:")
		for _, taskReport := range report.ByTask {
			fmt.Printf("  %s: %s\n", taskReport.TaskTitle, timeSvc.FormatDuration(taskReport.TotalDuration))
		}
	}

	return nil
}

func listTimeEntries(ctx context.Context, timeSvc *timeService.Service, args []string) error {
	entries, err := timeSvc.ListTimeEntries(ctx, timeService.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list time entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found")
		return nil
	}

	fmt.Println("Time Entries:")
	for _, entry := range entries {
		status := "Active"
		duration := "In progress"
		if entry.EndTime != nil {
			status = "Completed"
			duration = timeSvc.FormatDuration(entry.GetDuration())
		}
		fmt.Printf("  [%s] Task: %s | Duration: %s | Started: %s\n",
			status, entry.TaskID, duration, entry.StartTime.Format("2006-01-02 15:04"))
	}

	return nil
}

// Export handlers
func handleExportCommand(exportSvc *export.Service, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("export command requires a subcommand (tasks, time)")
	}

	ctx := context.Background()
	subcommand := args[0]

	switch subcommand {
	case "tasks":
		return exportTasks(ctx, exportSvc, args[1:])
	case "time":
		return exportTime(ctx, exportSvc, args[1:])
	default:
		return fmt.Errorf("unknown export subcommand: %s", subcommand)
	}
}

func exportTasks(ctx context.Context, exportSvc *export.Service, args []string) error {
	format := "json"
	output := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				i++
				format = args[i]
			}
		case "--output":
			if i+1 < len(args) {
				i++
				output = args[i]
			}
		}
	}

	var data []byte
	var err error

	filter := domain.TaskFilter{}

	switch format {
	case "json":
		data, err = exportSvc.ExportTasksToJSON(ctx, filter)
	case "csv":
		data, err = exportSvc.ExportTasksToCSV(ctx, filter)
	case "ical":
		data, err = exportSvc.ExportTasksToICAL(ctx, filter)
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, csv, ical)", format)
	}

	if err != nil {
		return fmt.Errorf("failed to export tasks: %w", err)
	}

	if output != "" {
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Exported tasks to %s\n", output)
	} else {
		fmt.Println(string(data))
	}

	return nil
}

func exportTime(ctx context.Context, exportSvc *export.Service, args []string) error {
	format := "json"
	output := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				i++
				format = args[i]
			}
		case "--output":
			if i+1 < len(args) {
				i++
				output = args[i]
			}
		}
	}

	var data []byte
	var err error

	filter := domain.TimeEntryFilter{}

	switch format {
	case "json":
		data, err = exportSvc.ExportTimeEntriesToJSON(ctx, filter)
	case "csv":
		data, err = exportSvc.ExportTimeEntriesToCSV(ctx, filter)
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, csv)", format)
	}

	if err != nil {
		return fmt.Errorf("failed to export time entries: %w", err)
	}

	if output != "" {
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Exported time entries to %s\n", output)
	} else {
		fmt.Println(string(data))
	}

	return nil
}

// Config handlers
func handleConfigCommand(cfg *domain.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config command requires a subcommand (show, set)")
	}

	subcommand := args[0]

	switch subcommand {
	case "show":
		return showConfig(cfg)
	case "set":
		return setConfig(cfg, args[1:])
	default:
		return fmt.Errorf("unknown config subcommand: %s", subcommand)
	}
}

func showConfig(cfg *domain.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println("Current Configuration:")
	fmt.Println(string(data))
	return nil
}

func setConfig(cfg *domain.Config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("config set requires key and value")
	}

	key := args[0]
	value := args[1]

	switch key {
	case "git_integration":
		cfg.GitIntegration = value == "true"
	case "default_project":
		cfg.DefaultProject = value
	case "time_format":
		cfg.TimeFormat = value
	case "date_format":
		cfg.DateFormat = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Updated %s to %s\n", key, value)
	return nil
}

// Git handlers
func handleGitCommand(gitRepo *git.GitRepository, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("git command requires a subcommand (hook)")
	}

	subcommand := args[0]

	switch subcommand {
	case "hook":
		return handleGitHook(gitRepo, args[1:])
	default:
		return fmt.Errorf("unknown git subcommand: %s", subcommand)
	}
}

func handleGitHook(gitRepo *git.GitRepository, args []string) error {
	var taskID string
	remove := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--task":
			if i+1 < len(args) {
				i++
				taskID = args[i]
			}
		case "--remove":
			remove = true
		}
	}

	if remove {
		if err := gitRepo.RemoveCommitHook(); err != nil {
			return fmt.Errorf("failed to remove commit hook: %w", err)
		}
		fmt.Println("Removed commit hook")
		return nil
	}

	if taskID == "" {
		return fmt.Errorf("--task flag is required")
	}

	if err := gitRepo.CreateCommitHook(taskID); err != nil {
		return fmt.Errorf("failed to create commit hook: %w", err)
	}

	fmt.Printf("Created commit hook for task %s\n", taskID)
	return nil
}

// Help system
func showHelp() error {
	helpText := `Project Manager CLI - Task and Time Management

USAGE:
  pm [command] [subcommand] [flags]

COMMANDS:
  task        Manage tasks
  project     Manage projects
  time        Track time
  export      Export data
  config      Manage configuration
  git         Git integration
  version     Show version
  help        Show this help

TASK COMMANDS:
  pm task add <title> [--priority high|medium|low] [--project <name>] [--tags tag1,tag2] [--cl <changelist>]
  pm task list [--status todo|doing|done] [--project <name>]
  pm task update <id> [--title <title>] [--status todo|doing|done|blocked] [--priority low|normal|high|critical]
  pm task complete <id>
  pm task delete <id>

PROJECT COMMANDS:
  pm project add <name>
  pm project list
  pm project delete <id>

TIME COMMANDS:
  pm time start --task <task-id> [--description <desc>]
  pm time stop
  pm time report [--today|--week|--month|--yesterday]
  pm time list

EXPORT COMMANDS:
  pm export tasks --format <json|csv|ical> [--output <file>]
  pm export time --format <json|csv> [--output <file>]

CONFIG COMMANDS:
  pm config show
  pm config set <key> <value>

GIT COMMANDS:
  pm git hook --task <task-id>
  pm git hook --remove

INTERACTIVE MODE:
  pm          Launch interactive TUI

For more information, visit: https://github.com/adriannajera/project-manager-cli
`
	fmt.Println(helpText)
	return nil
}
