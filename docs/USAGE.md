# Project Manager CLI Usage Guide

## Overview

Project Manager CLI (pm) is a powerful command-line tool for managing projects and tasks locally. It features both command-line interface (CLI) and terminal user interface (TUI) modes.

## Installation

### From Source
```bash
git clone https://github.com/adriannajera/project-manager-cli
cd project-manager-cli
go build -o pm cmd/pm/main.go
sudo mv pm /usr/local/bin/
```

### Direct Build
```bash
go install github.com/adriannajera/project-manager-cli/cmd/pm@latest
```

## Quick Start

### Interactive Mode (TUI)
Simply run the command without arguments to start the interactive dashboard:
```bash
pm
```

### Command-Line Mode
Use specific commands for quick operations:
```bash
# Create a task
pm task add "Implement user authentication"

# List tasks
pm task list

# Complete a task
pm task complete <task-id>

# Create a project
pm project add "Web Application"

# List projects
pm project list
```

## Task Management

### Creating Tasks
```bash
# Basic task creation
pm task add "Fix login bug"
pm task create "Add user profiles"

# With natural language due dates (future feature)
pm task add "Deploy to production" --due "tomorrow"
pm task add "Review pull requests" --due "friday"
```

### Listing Tasks
```bash
# List all tasks
pm task list
pm task ls

# Filter by status
pm task list --status todo
pm task list --status doing
pm task list --status done

# Filter by project
pm task list --project "web-app"

# Search tasks
pm task list --search "authentication"
```

### Updating Tasks
```bash
# Complete a task
pm task complete <task-id>
pm task done <task-id>

# Start working on a task
pm task start <task-id>

# Block a task
pm task block <task-id>

# Delete a task
pm task delete <task-id>
pm task rm <task-id>
```

## Project Management

### Creating Projects
```bash
# Create a new project
pm project add "Mobile App"
pm project create "API Backend"
```

### Listing Projects
```bash
# List all projects
pm project list
pm project ls

# Filter by status
pm project list --status active
pm project list --status archived
```

### Managing Projects
```bash
# Archive a project
pm project archive <project-id>

# Complete a project
pm project complete <project-id>

# Activate a project
pm project activate <project-id>
```

## Time Tracking

### Starting Time Tracking
```bash
# Start tracking time for a task
pm time start --task <task-id>
pm time start --task <task-id> --description "Working on authentication"
```

### Stopping Time Tracking
```bash
# Stop current time tracking
pm time stop
```

### Time Reports
```bash
# Today's time report
pm time report --today

# This week's report
pm time report --week

# This month's report
pm time report --month

# Custom date range
pm time report --start "2023-01-01" --end "2023-01-31"
```

## Git Integration

When working in a Git repository, the CLI automatically:
- Associates tasks with the current branch
- Can create commit hooks to reference task IDs in commit messages
- Tracks which branch a task was worked on

### Manual Git Operations
```bash
# Create commit hook for a task
pm git hook --task <task-id>

# Remove commit hook
pm git hook --remove

# List tasks for current branch
pm task list --branch current
```

## Export and Import

### Exporting Data
```bash
# Export tasks to JSON
pm export tasks --format json --output tasks.json

# Export tasks to CSV
pm export tasks --format csv --output tasks.csv

# Export tasks with due dates to iCal
pm export tasks --format ical --output tasks.ics

# Export time entries
pm export time --format csv --output timesheet.csv

# Export projects
pm export projects --format json --output projects.json
```

### Filtering Exports
```bash
# Export only completed tasks
pm export tasks --status done --format json

# Export tasks from specific project
pm export tasks --project "web-app" --format csv

# Export time entries for date range
pm export time --start "2023-01-01" --end "2023-01-31" --format csv
```

## Configuration

The CLI uses a configuration file located at `~/.pm/config.yaml`. You can customize:

```yaml
database_path: ~/.pm/tasks.db
default_project: ""
git_integration: true
time_format: "15:04"
date_format: "2006-01-02"
theme:
  primary: "#3b82f6"
  secondary: "#64748b"
  success: "#10b981"
  warning: "#f59e0b"
  error: "#ef4444"
  muted: "#6b7280"
aliases:
  ls: "list"
  new: "add"
  rm: "delete"
  done: "complete"
```

### Viewing Configuration
```bash
pm config show
pm config path
```

### Updating Configuration
```bash
pm config set git_integration false
pm config set default_project "my-project"
```

## Interactive Dashboard (TUI)

The interactive mode provides a rich terminal interface with:

- **Dashboard**: Overview of tasks, projects, and time tracking
- **Task List**: Browse and manage tasks with keyboard navigation
- **Task Forms**: Create and edit tasks with guided input
- **Project Management**: Organize and manage projects
- **Time Tracking**: Visual time tracking interface
- **Reports**: Interactive time and productivity reports

### Keyboard Shortcuts

#### Global
- `q` or `Ctrl+C`: Quit
- `?`: Show help
- `Esc`: Go back/cancel
- `r`: Refresh

#### Navigation
- `↑/k`: Move up
- `↓/j`: Move down
- `←/h`: Move left
- `→/l`: Move right
- `Enter`: Select/confirm

#### Task List
- `n`: New task
- `e`: Edit task
- `d`: Delete task
- `t`: Toggle task status
- `s`: Start time tracking
- `S`: Stop time tracking

#### Forms
- `Tab`: Next field
- `Shift+Tab`: Previous field
- `Ctrl+S`: Save
- `Esc`: Cancel

## Tips and Best Practices

### Task Organization
1. Use descriptive task titles
2. Add tags for categorization
3. Set realistic due dates
4. Break large tasks into subtasks
5. Use projects to group related tasks

### Time Tracking
1. Start tracking when you begin work
2. Add descriptive notes to time entries
3. Review time reports regularly
4. Use time data for project estimation

### Git Integration
1. Create tasks before starting new features
2. Use consistent branch naming
3. Let the CLI create commit hooks
4. Review task status when merging branches

### Productivity
1. Use the interactive dashboard for planning
2. Export data for external reporting
3. Set up aliases for common commands
4. Customize the configuration for your workflow

## Troubleshooting

### Common Issues

#### Database Errors
```bash
# Reset database (will lose all data)
rm ~/.pm/tasks.db
pm task list  # This will recreate the database
```

#### Permission Issues
```bash
# Fix database permissions
chmod 644 ~/.pm/tasks.db
chmod 755 ~/.pm/
```

#### Configuration Problems
```bash
# Reset configuration to defaults
rm ~/.pm/config.yaml
pm config show  # This will create default config
```

### Getting Help
```bash
# Show version
pm version

# Show help
pm help
pm task help
pm project help
pm time help

# Show command-specific help
pm task add --help
pm time start --help
```

## Advanced Features

### Scripting and Automation
The CLI is designed to be scriptable:

```bash
#!/bin/bash
# Daily standup report
echo "Tasks completed yesterday:"
pm task list --status done --since yesterday

echo "Tasks planned for today:"
pm task list --status todo --due today

echo "Time tracked yesterday:"
pm time report --yesterday
```

### Integration with Other Tools
- Export to project management tools via JSON/CSV
- Import time tracking data into billing systems
- Use with shell scripts for automation
- Integrate with Git hooks for workflow automation

### Performance Tips
- Use filters to limit large task lists
- Archive completed projects regularly
- Export old data and clean database periodically
- Use aliases for frequently used commands

## Support and Contribution

- Report issues on GitHub
- Contribute features via pull requests
- Join discussions for feature requests
- Share usage tips and workflows

For more information, visit the project repository or run `pm help` for built-in documentation.