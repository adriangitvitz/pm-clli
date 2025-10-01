package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/adriannajera/project-manager-cli/internal/ui"
)

// DashboardModel represents the dashboard view model
type DashboardModel struct {
	selectedIndex int
	menuItems     []DashboardItem
	keys          DashboardKeyMap
}

// DashboardItem represents a dashboard menu item
type DashboardItem struct {
	Title       string
	Description string
	Action      string
	Icon        string
}

// DashboardKeyMap defines key bindings for the dashboard
type DashboardKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Quit   key.Binding
	Escape key.Binding
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() DashboardModel {
	return DashboardModel{
		selectedIndex: 0,
		menuItems: []DashboardItem{
			{
				Title:       "Tasks",
				Description: "View and manage your tasks",
				Action:      "tasks",
				Icon:        "[T]",
			},
			{
				Title:       "Projects",
				Description: "Manage your projects",
				Action:      "projects",
				Icon:        "[P]",
			},
			{
				Title:       "Time Tracking",
				Description: "Track time spent on tasks",
				Action:      "time",
				Icon:        "[TM]",
			},
			{
				Title:       "New Task",
				Description: "Create a new task",
				Action:      "new_task",
				Icon:        "[+]",
			},
			{
				Title:       "Reports",
				Description: "View productivity reports",
				Action:      "reports",
				Icon:        "[R]",
			},
		},
		keys: DashboardKeyMap{
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
				key.WithHelp("enter", "select"),
			),
			Quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
			Escape: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
		},
	}
}

// Update handles dashboard updates
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case key.Matches(msg, m.keys.Down):
			if m.selectedIndex < len(m.menuItems)-1 {
				m.selectedIndex++
			}

		case key.Matches(msg, m.keys.Enter):
			selectedItem := m.menuItems[m.selectedIndex]
			return m, func() tea.Msg {
				return DashboardActionMsg{Action: selectedItem.Action}
			}
		}
	}

	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(ui.HeaderStyle.Render("Project Manager"))
	b.WriteString("\n\n")

	// Subtitle
	b.WriteString(ui.SubHeaderStyle.Render("Choose an action:"))
	b.WriteString("\n\n")

	// Menu items
	for i, item := range m.menuItems {
		style := ui.TableRowStyle
		if i == m.selectedIndex {
			style = ui.TableSelectedStyle
		}

		itemView := fmt.Sprintf("%s %s", item.Icon, item.Title)
		b.WriteString(style.Render(itemView))
		b.WriteString("\n")

		if i == m.selectedIndex {
			b.WriteString(ui.HelpStyle.Render("  " + item.Description))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("↑/↓: navigate • enter: select • q: quit • ?: help"))

	return ui.BaseStyle.Render(b.String())
}