package models

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriannajera/project-manager-cli/internal/domain"
)

// ViewState represents the current view state of the application
type ViewState int

const (
	TaskListView ViewState = iota
	TaskDetailView
	TaskFormView
	ProjectListView
	ProjectFormView
	TimeTrackingView
	DashboardView
	HelpView
)

// AppModel represents the main application model
type AppModel struct {
	// Current state
	currentView ViewState
	width       int
	height      int

	// Repositories
	taskRepo      domain.TaskRepository
	projectRepo   domain.ProjectRepository
	timeEntryRepo domain.TimeEntryRepository
	gitRepo       domain.GitRepository

	// Sub-models
	taskList    TaskListModel
	taskDetail  TaskDetailModel
	taskForm    TaskFormModel
	projectList ProjectListModel
	projectForm ProjectFormModel
	dashboard   DashboardModel

	// Global state
	selectedTask    *domain.Task
	selectedProject *domain.Project
	activeTimeEntry *domain.TimeEntry
	error           string
	success         string

	// Key bindings
	keys AppKeyMap
}

// AppKeyMap defines key bindings for the application
type AppKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Enter   key.Binding
	Back    key.Binding
	Quit    key.Binding
	Help    key.Binding
	Refresh key.Binding
	New     key.Binding
	Edit    key.Binding
	Delete  key.Binding
	Start   key.Binding
	Stop    key.Binding
	Toggle  key.Binding
}

// DefaultKeyMap returns the default key mappings
func DefaultKeyMap() AppKeyMap {
	return AppKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start"),
		),
		Stop: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "stop"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle"),
		),
	}
}

// NewAppModel creates a new application model
func NewAppModel(
	taskRepo domain.TaskRepository,
	projectRepo domain.ProjectRepository,
	timeEntryRepo domain.TimeEntryRepository,
	gitRepo domain.GitRepository,
) AppModel {
	return AppModel{
		currentView:   DashboardView,
		taskRepo:      taskRepo,
		projectRepo:   projectRepo,
		timeEntryRepo: timeEntryRepo,
		gitRepo:       gitRepo,
		keys:          DefaultKeyMap(),
		taskList:      NewTaskListModel(),
		taskDetail:    NewTaskDetailModel(),
		taskForm:      NewTaskFormModel(),
		projectList:   NewProjectListModel(),
		projectForm:   NewProjectFormModel(),
		dashboard:     NewDashboardModel(),
	}
}

// Init implements tea.Model
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadActiveTimeEntry(),
		m.loadInitialData(),
	)
}

// Update implements tea.Model
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.currentView = HelpView
			return m, nil

		case key.Matches(msg, m.keys.Back):
			if m.currentView != DashboardView {
				m.currentView = DashboardView
				m.error = ""
				m.success = ""
				return m, m.loadInitialData()
			}

		case key.Matches(msg, m.keys.Refresh):
			return m, m.loadInitialData()
		}

	case ErrorMsg:
		m.error = string(msg)
		return m, nil

	case SuccessMsg:
		m.success = string(msg)
		return m, nil

	case ActiveTimeEntryMsg:
		m.activeTimeEntry = msg.Entry
		return m, nil
	}

	// Delegate to current view
	switch m.currentView {
	case DashboardView:
		m.dashboard, cmd = m.dashboard.Update(msg)
		cmds = append(cmds, cmd)

		// Handle dashboard navigation
		if dashMsg, ok := msg.(DashboardActionMsg); ok {
			switch dashMsg.Action {
			case "tasks":
				m.currentView = TaskListView
				cmds = append(cmds, m.loadTasks())
			case "projects":
				m.currentView = ProjectListView
				cmds = append(cmds, m.loadProjects())
			case "new_task":
				m.currentView = TaskFormView
				m.taskForm = NewTaskFormModel()
			}
		}

	case TaskListView:
		m.taskList, cmd = m.taskList.Update(msg)
		cmds = append(cmds, cmd)

		// Handle task list actions
		if taskMsg, ok := msg.(TaskActionMsg); ok {
			switch taskMsg.Action {
			case "select":
				m.selectedTask = taskMsg.Task
				m.taskDetail.SetTask(taskMsg.Task)
				m.currentView = TaskDetailView
			case "new":
				m.currentView = TaskFormView
				m.taskForm = NewTaskFormModel()
			case "edit":
				m.selectedTask = taskMsg.Task
				m.currentView = TaskFormView
				m.taskForm = NewTaskFormModel()
				m.taskForm.LoadTask(taskMsg.Task)
			case "delete":
				cmds = append(cmds, m.deleteTask(taskMsg.Task))
				cmds = append(cmds, m.loadTasks())
			case "update":
				cmds = append(cmds, m.saveTask(taskMsg.Task))
				cmds = append(cmds, m.loadTasks())
			}
		}

	case TaskDetailView:
		m.taskDetail, cmd = m.taskDetail.Update(msg)
		cmds = append(cmds, cmd)

		// Handle task detail actions
		if taskMsg, ok := msg.(TaskActionMsg); ok {
			switch taskMsg.Action {
			case "edit":
				m.selectedTask = taskMsg.Task
				m.currentView = TaskFormView
				m.taskForm = NewTaskFormModel()
				m.taskForm.LoadTask(taskMsg.Task)
			case "delete":
				m.currentView = TaskListView
				cmds = append(cmds, m.deleteTask(taskMsg.Task))
				cmds = append(cmds, m.loadTasks())
			case "update":
				cmds = append(cmds, m.saveTask(taskMsg.Task))
				m.taskDetail.SetTask(taskMsg.Task)
			}
		}

	case TaskFormView:
		m.taskForm, cmd = m.taskForm.Update(msg)
		cmds = append(cmds, cmd)

		// Handle form submission
		if formMsg, ok := msg.(TaskFormSubmitMsg); ok {
			m.currentView = TaskListView
			cmds = append(cmds, m.saveTask(formMsg.Task), m.loadTasks())
		}

	case ProjectListView:
		m.projectList, cmd = m.projectList.Update(msg)
		cmds = append(cmds, cmd)

		// Handle project list messages
		if projMsg, ok := msg.(ProjectListLoadedMsg); ok {
			m.projectList.LoadProjects(projMsg.Projects)
		}

		// Handle project actions
		if projMsg, ok := msg.(ProjectActionMsg); ok {
			switch projMsg.Action {
			case "select":
				m.selectedProject = projMsg.Project
				m.currentView = TaskListView
				cmds = append(cmds, m.loadTasksForProject(projMsg.Project.ID))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m AppModel) View() string {
	switch m.currentView {
	case DashboardView:
		return m.dashboard.View()
	case TaskListView:
		return m.taskList.View()
	case TaskDetailView:
		return m.taskDetail.View()
	case TaskFormView:
		return m.taskForm.View()
	case ProjectListView:
		return m.projectList.View()
	case HelpView:
		return m.renderHelp()
	default:
		return "Unknown view"
	}
}

// Message types
type ErrorMsg string
type SuccessMsg string

type ActiveTimeEntryMsg struct {
	Entry *domain.TimeEntry
}

type DashboardActionMsg struct {
	Action string
}

type TaskActionMsg struct {
	Action string
	Task   *domain.Task
}

type TaskFormSubmitMsg struct {
	Task *domain.Task
}

// Commands
func (m AppModel) loadActiveTimeEntry() tea.Cmd {
	return func() tea.Msg {
		// This would call the repository in a real implementation
		// For now, return nil to avoid compilation errors
		return ActiveTimeEntryMsg{Entry: nil}
	}
}

func (m AppModel) loadInitialData() tea.Cmd {
	return tea.Batch(
		m.loadTasks(),
		m.loadProjects(),
	)
}

func (m AppModel) loadTasks() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		tasks, err := m.taskRepo.List(ctx, domain.TaskFilter{})
		if err != nil {
			return ErrorMsg("Failed to load tasks: " + err.Error())
		}
		return TaskListLoadedMsg{Tasks: tasks}
	}
}

func (m AppModel) loadTasksForProject(projectID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		tasks, err := m.taskRepo.List(ctx, domain.TaskFilter{ProjectID: projectID})
		if err != nil {
			return ErrorMsg("Failed to load tasks: " + err.Error())
		}
		return TaskListLoadedMsg{Tasks: tasks}
	}
}

func (m AppModel) loadProjects() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		projects, err := m.projectRepo.List(ctx, domain.ProjectFilter{})
		if err != nil {
			return ErrorMsg("Failed to load projects: " + err.Error())
		}
		return ProjectListLoadedMsg{Projects: projects}
	}
}

func (m AppModel) saveTask(task *domain.Task) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.taskRepo.Update(ctx, task); err != nil {
			return ErrorMsg("Failed to save task: " + err.Error())
		}
		return SuccessMsg("Task saved successfully")
	}
}

func (m AppModel) deleteTask(task *domain.Task) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.taskRepo.Delete(ctx, task.ID); err != nil {
			return ErrorMsg("Failed to delete task: " + err.Error())
		}
		return SuccessMsg("Task deleted successfully")
	}
}

func (m AppModel) renderHelp() string {
	return "Help view - TODO: implement help content"
}