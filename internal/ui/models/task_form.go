package models

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriannajera/project-manager-cli/internal/domain"
	"github.com/adriannajera/project-manager-cli/internal/ui"
)

// TaskFormModel represents the task form view model
type TaskFormModel struct {
	// Form fields
	titleInput       textinput.Model
	descriptionInput textinput.Model
	tagsInput        textinput.Model
	changelistInput  textinput.Model
	dueDateInput     textinput.Model

	// Form state
	focusedField int
	task         *domain.Task
	isEditing    bool
	keys         TaskFormKeyMap
}

// TaskFormKeyMap defines key bindings for the task form
type TaskFormKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
	Next   key.Binding
	Prev   key.Binding
}

// NewTaskFormModel creates a new task form model
func NewTaskFormModel() TaskFormModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter task title..."
	titleInput.Focus()
	titleInput.CharLimit = 200
	titleInput.Width = 50

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "Enter task description..."
	descriptionInput.CharLimit = 500
	descriptionInput.Width = 50

	tagsInput := textinput.New()
	tagsInput.Placeholder = "Enter tags (comma-separated)..."
	tagsInput.CharLimit = 200
	tagsInput.Width = 50

	changelistInput := textinput.New()
	changelistInput.Placeholder = "Enter changelist (e.g., c/1234, CL/456)..."
	changelistInput.CharLimit = 100
	changelistInput.Width = 50

	dueDateInput := textinput.New()
	dueDateInput.Placeholder = "Enter due date (YYYY-MM-DD)..."
	dueDateInput.CharLimit = 10
	dueDateInput.Width = 50

	return TaskFormModel{
		titleInput:       titleInput,
		descriptionInput: descriptionInput,
		tagsInput:        tagsInput,
		changelistInput:  changelistInput,
		dueDateInput:     dueDateInput,
		focusedField:     0,
		keys: TaskFormKeyMap{
			Submit: key.NewBinding(
				key.WithKeys("ctrl+s"),
				key.WithHelp("ctrl+s", "save"),
			),
			Cancel: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "cancel"),
			),
			Next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next field"),
			),
			Prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev field"),
			),
		},
	}
}

// LoadTask loads an existing task into the form for editing
func (m *TaskFormModel) LoadTask(task *domain.Task) {
	m.task = task
	m.isEditing = true

	m.titleInput.SetValue(task.Title)
	m.descriptionInput.SetValue(task.Description)

	if len(task.Tags) > 0 {
		m.tagsInput.SetValue(strings.Join(task.Tags, ", "))
	}

	if task.Changelist != "" {
		m.changelistInput.SetValue(task.Changelist)
	}

	if task.DueDate != nil {
		m.dueDateInput.SetValue(task.DueDate.Format("2006-01-02"))
	}
}

// Update handles task form updates
func (m TaskFormModel) Update(msg tea.Msg) (TaskFormModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Submit):
			return m.handleSubmit()

		case key.Matches(msg, m.keys.Cancel):
			return m, func() tea.Msg {
				return TaskFormCancelMsg{}
			}

		case key.Matches(msg, m.keys.Next):
			m.focusedField = (m.focusedField + 1) % 5
			m.updateFocus()

		case key.Matches(msg, m.keys.Prev):
			m.focusedField = (m.focusedField - 1 + 5) % 5
			m.updateFocus()
		}
	}

	// Update the focused input
	switch m.focusedField {
	case 0:
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		m.descriptionInput, cmd = m.descriptionInput.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		m.tagsInput, cmd = m.tagsInput.Update(msg)
		cmds = append(cmds, cmd)
	case 3:
		m.changelistInput, cmd = m.changelistInput.Update(msg)
		cmds = append(cmds, cmd)
	case 4:
		m.dueDateInput, cmd = m.dueDateInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the task form
func (m TaskFormModel) View() string {
	var b strings.Builder

	// Header
	if m.isEditing {
		b.WriteString(ui.HeaderStyle.Render("Edit Task"))
	} else {
		b.WriteString(ui.HeaderStyle.Render("New Task"))
	}
	b.WriteString("\n\n")

	// Title field
	b.WriteString("Title:")
	b.WriteString("\n")
	if m.focusedField == 0 {
		b.WriteString(ui.FocusedInputStyle.Render(m.titleInput.View()))
	} else {
		b.WriteString(ui.InputStyle.Render(m.titleInput.View()))
	}
	b.WriteString("\n\n")

	// Description field
	b.WriteString("Description:")
	b.WriteString("\n")
	if m.focusedField == 1 {
		b.WriteString(ui.FocusedInputStyle.Render(m.descriptionInput.View()))
	} else {
		b.WriteString(ui.InputStyle.Render(m.descriptionInput.View()))
	}
	b.WriteString("\n\n")

	// Tags field
	b.WriteString("Tags (comma-separated):")
	b.WriteString("\n")
	if m.focusedField == 2 {
		b.WriteString(ui.FocusedInputStyle.Render(m.tagsInput.View()))
	} else {
		b.WriteString(ui.InputStyle.Render(m.tagsInput.View()))
	}
	b.WriteString("\n\n")

	// Changelist field
	b.WriteString("Changelist (e.g., c/1234):")
	b.WriteString("\n")
	if m.focusedField == 3 {
		b.WriteString(ui.FocusedInputStyle.Render(m.changelistInput.View()))
	} else {
		b.WriteString(ui.InputStyle.Render(m.changelistInput.View()))
	}
	b.WriteString("\n\n")

	// Due date field
	b.WriteString("Due Date (YYYY-MM-DD):")
	b.WriteString("\n")
	if m.focusedField == 4 {
		b.WriteString(ui.FocusedInputStyle.Render(m.dueDateInput.View()))
	} else {
		b.WriteString(ui.InputStyle.Render(m.dueDateInput.View()))
	}
	b.WriteString("\n\n")

	// Help footer
	helpText := "tab: next field • shift+tab: prev field • ctrl+s: save • esc: cancel"
	b.WriteString(ui.HelpStyle.Render(helpText))

	return ui.BaseStyle.Render(b.String())
}

// updateFocus updates which field has focus
func (m *TaskFormModel) updateFocus() {
	m.titleInput.Blur()
	m.descriptionInput.Blur()
	m.tagsInput.Blur()
	m.changelistInput.Blur()
	m.dueDateInput.Blur()

	switch m.focusedField {
	case 0:
		m.titleInput.Focus()
	case 1:
		m.descriptionInput.Focus()
	case 2:
		m.tagsInput.Focus()
	case 3:
		m.changelistInput.Focus()
	case 4:
		m.dueDateInput.Focus()
	}
}

// handleSubmit processes form submission
func (m TaskFormModel) handleSubmit() (TaskFormModel, tea.Cmd) {
	// Validate required fields
	if strings.TrimSpace(m.titleInput.Value()) == "" {
		return m, func() tea.Msg {
			return ErrorMsg("Title is required")
		}
	}

	var task *domain.Task
	if m.isEditing && m.task != nil {
		task = m.task
	} else {
		task = domain.NewTask(
			strings.TrimSpace(m.titleInput.Value()),
			strings.TrimSpace(m.descriptionInput.Value()),
		)
	}

	// Update task fields
	task.Title = strings.TrimSpace(m.titleInput.Value())
	task.Description = strings.TrimSpace(m.descriptionInput.Value())

	// Parse tags
	if tagText := strings.TrimSpace(m.tagsInput.Value()); tagText != "" {
		tags := strings.Split(tagText, ",")
		task.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				task.Tags = append(task.Tags, trimmed)
			}
		}
	}

	// Parse changelist
	if changelistText := strings.TrimSpace(m.changelistInput.Value()); changelistText != "" {
		task.Changelist = changelistText
	}

	// Parse due date
	if dueDateText := strings.TrimSpace(m.dueDateInput.Value()); dueDateText != "" {
		if dueDate, err := time.Parse("2006-01-02", dueDateText); err == nil {
			task.DueDate = &dueDate
		} else {
			return m, func() tea.Msg {
				return ErrorMsg("Invalid due date format. Use YYYY-MM-DD")
			}
		}
	}

	task.UpdatedAt = time.Now()

	return m, func() tea.Msg {
		return TaskFormSubmitMsg{Task: task}
	}
}

// Message types
type TaskFormCancelMsg struct{}

// ProjectFormModel placeholder
type ProjectFormModel struct{}

func NewProjectFormModel() ProjectFormModel {
	return ProjectFormModel{}
}

func (m ProjectFormModel) Update(msg tea.Msg) (ProjectFormModel, tea.Cmd) {
	return m, nil
}

func (m ProjectFormModel) View() string {
	return "Project Form - TODO: implement"
}