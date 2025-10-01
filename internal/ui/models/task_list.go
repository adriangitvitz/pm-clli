package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriannajera/project-manager-cli/internal/domain"
	"github.com/adriannajera/project-manager-cli/internal/ui"
)

// TaskListModel represents the task list view model
type TaskListModel struct {
	tasks         []*domain.Task
	selectedIndex int
	filter        domain.TaskFilter
	loading       bool
	keys          TaskListKeyMap
}

// TaskListKeyMap defines key bindings for the task list
type TaskListKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Toggle key.Binding
	Filter key.Binding
	Back   key.Binding
}

// NewTaskListModel creates a new task list model
func NewTaskListModel() TaskListModel {
	return TaskListModel{
		tasks:         make([]*domain.Task, 0),
		selectedIndex: 0,
		filter:        domain.TaskFilter{},
		loading:       false,
		keys: TaskListKeyMap{
			Up: key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp("↑/k", "up"),
			),
			Down: key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp("↓/j", "down"),
			),
			Enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "view details"),
			),
			New: key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new task"),
			),
			Edit: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit"),
			),
			Delete: key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete"),
			),
			Toggle: key.NewBinding(
				key.WithKeys("t"),
				key.WithHelp("t", "toggle status"),
			),
			Filter: key.NewBinding(
				key.WithKeys("f"),
				key.WithHelp("f", "filter"),
			),
			Back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
		},
	}
}

// Update handles task list updates
func (m TaskListModel) Update(msg tea.Msg) (TaskListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(m.tasks) == 0 {
			switch {
			case key.Matches(msg, m.keys.New):
				return m, func() tea.Msg {
					return TaskActionMsg{Action: "new"}
				}
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case key.Matches(msg, m.keys.Down):
			if m.selectedIndex < len(m.tasks)-1 {
				m.selectedIndex++
			}

		case key.Matches(msg, m.keys.Enter):
			if m.selectedIndex < len(m.tasks) {
				return m, func() tea.Msg {
					return TaskActionMsg{
						Action: "select",
						Task:   m.tasks[m.selectedIndex],
					}
				}
			}

		case key.Matches(msg, m.keys.New):
			return m, func() tea.Msg {
				return TaskActionMsg{Action: "new"}
			}

		case key.Matches(msg, m.keys.Edit):
			if m.selectedIndex < len(m.tasks) {
				return m, func() tea.Msg {
					return TaskActionMsg{
						Action: "edit",
						Task:   m.tasks[m.selectedIndex],
					}
				}
			}

		case key.Matches(msg, m.keys.Toggle):
			if m.selectedIndex < len(m.tasks) {
				task := m.tasks[m.selectedIndex]
				if task.Status == domain.StatusDone {
					task.Status = domain.StatusTodo
				} else {
					task.Complete()
				}
				return m, func() tea.Msg {
					return TaskActionMsg{
						Action: "update",
						Task:   task,
					}
				}
			}

		case key.Matches(msg, m.keys.Delete):
			if m.selectedIndex < len(m.tasks) {
				return m, func() tea.Msg {
					return TaskActionMsg{
						Action: "delete",
						Task:   m.tasks[m.selectedIndex],
					}
				}
			}
		}

	case TaskListLoadedMsg:
		m.tasks = msg.Tasks
		m.loading = false
		if m.selectedIndex >= len(m.tasks) && len(m.tasks) > 0 {
			m.selectedIndex = len(m.tasks) - 1
		}
	}

	return m, nil
}

// View renders the task list
func (m TaskListModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(ui.HeaderStyle.Render("Tasks"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("Loading tasks...")
		return ui.BaseStyle.Render(b.String())
	}

	if len(m.tasks) == 0 {
		b.WriteString(ui.HelpStyle.Render("No tasks found. Press 'n' to create a new task."))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("n: new task • esc: back"))
		return ui.BaseStyle.Render(b.String())
	}

	// Task list
	for i, task := range m.tasks {
		style := ui.TableRowStyle
		if i == m.selectedIndex {
			style = ui.TableSelectedStyle
		}

		// Format task line
		statusIcon := ui.FormatStatusIcon(string(task.Status))
		priorityIcon := ui.FormatPriorityIcon(int(task.Priority))

		taskLine := fmt.Sprintf("%s %s %s",
			statusIcon,
			priorityIcon,
			task.Title,
		)

		// Add changelist if any
		if task.Changelist != "" {
			taskLine += ui.TagStyle.Render(fmt.Sprintf(" (%s)", task.Changelist))
		}

		// Add tags if any
		if len(task.Tags) > 0 {
			tags := strings.Join(task.Tags, ", ")
			taskLine += ui.TagStyle.Render(fmt.Sprintf(" [%s]", tags))
		}

		// Add due date if exists
		if task.DueDate != nil {
			dueStyle := ui.HelpStyle
			if task.IsOverdue() {
				dueStyle = ui.ErrorStyle
			}
			taskLine += dueStyle.Render(fmt.Sprintf(" (due: %s)", task.DueDate.Format("Jan 2")))
		}

		b.WriteString(style.Render(taskLine))
		b.WriteString("\n")

		// Show description for selected task
		if i == m.selectedIndex && task.Description != "" {
			b.WriteString(ui.HelpStyle.Render("  " + task.Description))
			b.WriteString("\n")
		}
	}

	// Help footer
	b.WriteString("\n")
	helpText := "↑/↓: navigate • enter: details • n: new • e: edit • t: toggle • d: delete • esc: back"
	b.WriteString(ui.HelpStyle.Render(helpText))

	return ui.BaseStyle.Render(b.String())
}

// LoadTasks sets the tasks for the model
func (m *TaskListModel) LoadTasks(tasks []*domain.Task) {
	m.tasks = tasks
	if m.selectedIndex >= len(tasks) && len(tasks) > 0 {
		m.selectedIndex = len(tasks) - 1
	}
}

// Message types
type TaskListLoadedMsg struct {
	Tasks []*domain.Task
}