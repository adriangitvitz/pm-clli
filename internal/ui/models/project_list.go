package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriannajera/project-manager-cli/internal/domain"
	"github.com/adriannajera/project-manager-cli/internal/ui"
)

// ProjectListModel represents the project list view model
type ProjectListModel struct {
	projects      []*domain.Project
	selectedIndex int
	loading       bool
	keys          ProjectListKeyMap
}

// ProjectListKeyMap defines key bindings for the project list
type ProjectListKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Back   key.Binding
}

// NewProjectListModel creates a new project list model
func NewProjectListModel() ProjectListModel {
	return ProjectListModel{
		projects:      make([]*domain.Project, 0),
		selectedIndex: 0,
		loading:       false,
		keys: ProjectListKeyMap{
			Up: key.NewBinding(
				key.WithKeys("up", "k"),
				key.WithHelp("up/k", "up"),
			),
			Down: key.NewBinding(
				key.WithKeys("down", "j"),
				key.WithHelp("down/j", "down"),
			),
			Enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "view details"),
			),
			New: key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new project"),
			),
			Edit: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit"),
			),
			Delete: key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete"),
			),
			Back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
		},
	}
}

// Update handles project list updates
func (m ProjectListModel) Update(msg tea.Msg) (ProjectListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(m.projects) == 0 {
			switch {
			case key.Matches(msg, m.keys.New):
				return m, func() tea.Msg {
					return ProjectActionMsg{Action: "new"}
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
			if m.selectedIndex < len(m.projects)-1 {
				m.selectedIndex++
			}

		case key.Matches(msg, m.keys.Enter):
			if m.selectedIndex < len(m.projects) {
				return m, func() tea.Msg {
					return ProjectActionMsg{
						Action:  "select",
						Project: m.projects[m.selectedIndex],
					}
				}
			}

		case key.Matches(msg, m.keys.New):
			return m, func() tea.Msg {
				return ProjectActionMsg{Action: "new"}
			}

		case key.Matches(msg, m.keys.Edit):
			if m.selectedIndex < len(m.projects) {
				return m, func() tea.Msg {
					return ProjectActionMsg{
						Action:  "edit",
						Project: m.projects[m.selectedIndex],
					}
				}
			}
		}

	case ProjectListLoadedMsg:
		m.projects = msg.Projects
		m.loading = false
		if m.selectedIndex >= len(m.projects) && len(m.projects) > 0 {
			m.selectedIndex = len(m.projects) - 1
		}
	}

	return m, nil
}

// View renders the project list
func (m ProjectListModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(ui.HeaderStyle.Render("Projects"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("Loading projects...")
		return ui.BaseStyle.Render(b.String())
	}

	if len(m.projects) == 0 {
		b.WriteString(ui.HelpStyle.Render("No projects found. Press 'n' to create a new project."))
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("n: new project • esc: back"))
		return ui.BaseStyle.Render(b.String())
	}

	// Project list
	for i, project := range m.projects {
		style := ui.TableRowStyle
		if i == m.selectedIndex {
			style = ui.TableSelectedStyle
		}

		projectLine := fmt.Sprintf("[P] %s", project.Name)

		if project.Description != "" && i != m.selectedIndex {
			projectLine += ui.HelpStyle.Render(fmt.Sprintf(" - %s", project.Description))
		}

		b.WriteString(style.Render(projectLine))
		b.WriteString("\n")

		// Show description for selected project
		if i == m.selectedIndex && project.Description != "" {
			b.WriteString(ui.HelpStyle.Render("  " + project.Description))
			b.WriteString("\n")
		}
	}

	// Help footer
	b.WriteString("\n")
	helpText := "up/down: navigate • enter: details • n: new • e: edit • d: delete • esc: back"
	b.WriteString(ui.HelpStyle.Render(helpText))

	return ui.BaseStyle.Render(b.String())
}

// LoadProjects sets the projects for the model
func (m *ProjectListModel) LoadProjects(projects []*domain.Project) {
	m.projects = projects
	if m.selectedIndex >= len(projects) && len(projects) > 0 {
		m.selectedIndex = len(projects) - 1
	}
}

// Message types
type ProjectListLoadedMsg struct {
	Projects []*domain.Project
}

type ProjectActionMsg struct {
	Action  string
	Project *domain.Project
}
