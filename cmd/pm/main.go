package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriannajera/project-manager-cli/internal/domain"
	"github.com/adriannajera/project-manager-cli/internal/repository/sqlite"
	"github.com/adriannajera/project-manager-cli/internal/service/task"
	"github.com/adriannajera/project-manager-cli/internal/service/project"
	"github.com/adriannajera/project-manager-cli/internal/service/git"
	"github.com/adriannajera/project-manager-cli/internal/ui/models"
	"github.com/adriannajera/project-manager-cli/pkg/config"
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
		return runCLI(taskRepo, projectRepo, timeEntryRepo, gitRepo)
	}

	// Run TUI application
	return runTUI(taskRepo, projectRepo, timeEntryRepo, gitRepo)
}

func runCLI(taskRepo *sqlite.TaskRepository, projectRepo *sqlite.ProjectRepository, timeEntryRepo *sqlite.TimeEntryRepository, gitRepo *git.GitRepository) error {
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
		return handleTimeCommand(timeEntryRepo, os.Args[2:])
	case "version":
		fmt.Println("Project Manager CLI v1.0.0")
		return nil
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
	default:
		return fmt.Errorf("unknown project subcommand: %s", subcommand)
	}
}

func handleTimeCommand(timeEntryRepo *sqlite.TimeEntryRepository, args []string) error {
	fmt.Println("Time tracking commands - TODO: implement")
	return nil
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
		fmt.Printf("  %s\n", p.Name)
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
