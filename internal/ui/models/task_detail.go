package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriannajera/project-manager-cli/internal/domain"
	"github.com/adriannajera/project-manager-cli/internal/ui"
)

// TaskDetailModel represents the task detail view model
type TaskDetailModel struct {
	task *domain.Task
	keys TaskDetailKeyMap
}

// TaskDetailKeyMap defines key bindings for the task detail view
type TaskDetailKeyMap struct {
	Back   key.Binding
	Edit   key.Binding
	Delete key.Binding
	Toggle key.Binding
}

// NewTaskDetailModel creates a new task detail model
func NewTaskDetailModel() TaskDetailModel {
	return TaskDetailModel{
		keys: TaskDetailKeyMap{
			Back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
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
		},
	}
}

// SetTask sets the task to display
func (m *TaskDetailModel) SetTask(task *domain.Task) {
	m.task = task
}

// Update handles task detail updates
func (m TaskDetailModel) Update(msg tea.Msg) (TaskDetailModel, tea.Cmd) {
	if m.task == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Edit):
			return m, func() tea.Msg {
				return TaskActionMsg{
					Action: "edit",
					Task:   m.task,
				}
			}

		case key.Matches(msg, m.keys.Delete):
			return m, func() tea.Msg {
				return TaskActionMsg{
					Action: "delete",
					Task:   m.task,
				}
			}

		case key.Matches(msg, m.keys.Toggle):
			if m.task.Status == domain.StatusDone {
				m.task.Status = domain.StatusTodo
			} else {
				m.task.Complete()
			}
			return m, func() tea.Msg {
				return TaskActionMsg{
					Action: "update",
					Task:   m.task,
				}
			}
		}
	}

	return m, nil
}

// View renders the task detail view
func (m TaskDetailModel) View() string {
	if m.task == nil {
		return ui.BaseStyle.Render("No task selected")
	}

	var b strings.Builder

	// Header
	b.WriteString(ui.HeaderStyle.Render("Task Details"))
	b.WriteString("\n\n")

	// Title
	b.WriteString(ui.SubHeaderStyle.Render("Title:"))
	b.WriteString("\n")
	b.WriteString(m.task.Title)
	b.WriteString("\n\n")

	// Status
	b.WriteString(ui.SubHeaderStyle.Render("Status:"))
	b.WriteString("\n")
	b.WriteString(ui.FormatStatusIcon(string(m.task.Status)))
	b.WriteString(" ")
	b.WriteString(string(m.task.Status))
	b.WriteString("\n\n")

	// Priority
	b.WriteString(ui.SubHeaderStyle.Render("Priority:"))
	b.WriteString("\n")
	b.WriteString(ui.FormatPriorityIcon(int(m.task.Priority)))
	b.WriteString("\n\n")

	// Description
	if m.task.Description != "" {
		b.WriteString(ui.SubHeaderStyle.Render("Description:"))
		b.WriteString("\n")
		b.WriteString(m.task.Description)
		b.WriteString("\n\n")
	}

	// Changelist
	if m.task.Changelist != "" {
		b.WriteString(ui.SubHeaderStyle.Render("Changelist:"))
		b.WriteString("\n")
		b.WriteString(ui.TagStyle.Render(m.task.Changelist))
		b.WriteString("\n\n")
	}

	// Tags
	if len(m.task.Tags) > 0 {
		b.WriteString(ui.SubHeaderStyle.Render("Tags:"))
		b.WriteString("\n")
		b.WriteString(ui.TagStyle.Render(strings.Join(m.task.Tags, ", ")))
		b.WriteString("\n\n")
	}

	// Due Date
	if m.task.DueDate != nil {
		b.WriteString(ui.SubHeaderStyle.Render("Due Date:"))
		b.WriteString("\n")
		dueStyle := ui.HelpStyle
		if m.task.IsOverdue() {
			dueStyle = ui.ErrorStyle
		}
		b.WriteString(dueStyle.Render(m.task.DueDate.Format("2006-01-02")))
		b.WriteString("\n\n")
	}

	// Created/Updated
	b.WriteString(ui.HelpStyle.Render(fmt.Sprintf("Created: %s", m.task.CreatedAt.Format("2006-01-02 15:04"))))
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render(fmt.Sprintf("Updated: %s", m.task.UpdatedAt.Format("2006-01-02 15:04"))))
	b.WriteString("\n\n")

	// Help footer
	helpText := "e: edit • d: delete • t: toggle status • esc: back"
	b.WriteString(ui.HelpStyle.Render(helpText))

	return ui.BaseStyle.Render(b.String())
}
