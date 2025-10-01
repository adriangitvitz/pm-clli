package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	primaryColor   = lipgloss.Color("#3b82f6")
	secondaryColor = lipgloss.Color("#64748b")
	successColor   = lipgloss.Color("#10b981")
	warningColor   = lipgloss.Color("#f59e0b")
	errorColor     = lipgloss.Color("#ef4444")
	mutedColor     = lipgloss.Color("#6b7280")

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	SubHeaderStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Padding(0, 1)

	// Task status styles
	TodoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94a3b8")).
			Padding(0, 1)

	DoingStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true).
			Padding(0, 1)

	DoneStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Strikethrough(true).
			Padding(0, 1)

	BlockedStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Padding(0, 1)

	// Priority styles
	LowPriorityStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	NormalPriorityStyle = lipgloss.NewStyle().
				Foreground(secondaryColor)

	HighPriorityStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)

	CriticalPriorityStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	// Tag style
	TagStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Faint(true)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(mutedColor)

	TableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(lipgloss.Color("#ffffff")).
				Padding(0, 1)

	// Form styles
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	ButtonStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2).
			Margin(0, 1)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Faint(true)

	// Error styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1)

	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1)
)

// GetStatusStyle returns the appropriate style for a task status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "todo", "backlog":
		return TodoStyle
	case "doing":
		return DoingStyle
	case "done":
		return DoneStyle
	case "blocked":
		return BlockedStyle
	default:
		return TodoStyle
	}
}

// GetPriorityStyle returns the appropriate style for a priority level
func GetPriorityStyle(priority int) lipgloss.Style {
	switch priority {
	case 0: // Low
		return LowPriorityStyle
	case 1: // Normal
		return NormalPriorityStyle
	case 2: // High
		return HighPriorityStyle
	case 3: // Critical
		return CriticalPriorityStyle
	default:
		return NormalPriorityStyle
	}
}

// FormatStatusIcon returns a colored icon for the task status
func FormatStatusIcon(status string) string {
	switch status {
	case "todo":
		return TodoStyle.Render("[ ]")
	case "doing":
		return DoingStyle.Render("[~]")
	case "done":
		return DoneStyle.Render("[x]")
	case "blocked":
		return BlockedStyle.Render("[!]")
	case "backlog":
		return TodoStyle.Render("[-]")
	default:
		return TodoStyle.Render("[ ]")
	}
}

// FormatPriorityIcon returns a colored icon for the priority level
func FormatPriorityIcon(priority int) string {
	switch priority {
	case 0: // Low
		return LowPriorityStyle.Render("[LOW]")
	case 1: // Normal
		return NormalPriorityStyle.Render("[NORM]")
	case 2: // High
		return HighPriorityStyle.Render("[HIGH]")
	case 3: // Critical
		return CriticalPriorityStyle.Render("[CRIT]")
	default:
		return NormalPriorityStyle.Render("[NORM]")
	}
}