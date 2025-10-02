# Project Manager CLI Usage Guide

## Overview

Project Manager CLI (pm) is a powerful command-line tool for managing projects and tasks locally. It features both command-line interface (CLI) and terminal user interface (TUI) modes with full support for time tracking, exports, configuration management, and Git integration.

## Installation

### From Source
```bash
git clone https://github.com/adriannajera/project-manager-cli
cd project-manager-cli
make build
sudo make install-system
```

### Using Go Install
```bash
go install github.com/adriannajera/project-manager-cli/cmd/pm@latest
```

### Pre-built Binary
Download the latest release for your platform from the releases page and add it to your PATH.

## Quick Start

### Interactive Mode (TUI)
Simply run the command without arguments to start the interactive dashboard:
```bash
pm
```

### Command-Line Mode
Use specific commands for quick operations:
```bash
# Get help
pm help

# Create a task
pm task add "Implement user authentication"

# List tasks
pm task list

# Complete a task
pm task complete <task-id>

# Create a project
pm project add "Web Application"

# Track time
pm time start --task <task-id>
pm time stop

# Export data
pm export tasks --format json --output tasks.json

# View configuration
pm config show
```

## Task Management

### Creating Tasks
```bash
# Basic task creation
pm task add "Fix login bug"

# With priority
pm task add "Deploy to production" --priority high

# With project and tags
pm task add "Add user profiles" --project "web-app" --tags feature,backend

# With changelist/CL tracking
pm task add "Fix authentication" --cl 123456

# With workspace tracking
pm task add "Refactor authentication" --workspace "workspace-1"
pm task add "Fix bug" --ws "workspace-2"  # --ws is short for --workspace

# With description
pm task add "Review pull requests" --description "Weekly PR review"

# Combine multiple flags
pm task add "New feature" --project "web-app" --workspace "main-workspace" --cl 789 --priority high
```

### Listing Tasks
```bash
# List all tasks
pm task list
pm task ls  # alias

# Filter by project
pm task list --project "web-app"

# Filter by workspace
pm task list --workspace "workspace-1"

# Filter by status
pm task list --status todo
pm task list --status doing
pm task list --status done
pm task list --status blocked

# Combine filters
pm task list --project "web-app" --workspace "workspace-1" --status doing
```

**Task List Output Format:**
```
Tasks:
  [ ] [LOW] Task title (task-id)
     * Project: project-name
     * cl: changelist-number
     * workspace: workspace-name
     * completed: 2025-10-02 15:04:05
```

**Status Icons:**
- `[ ]` - Todo
- `[~]` - Doing (in progress)
- `[x]` - Done (completed)
- `[!]` - Blocked

**Note:** Advanced search and other filtering options are currently available in the TUI mode.

### Updating Tasks
```bash
# Update task status
pm task update <task-id> --status doing
pm task update <task-id> --status done
pm task update <task-id> --status blocked

# Update task priority
pm task update <task-id> --priority high
pm task update <task-id> --priority critical

# Update task title
pm task update <task-id> --title "New task title"

# Update changelist
pm task update <task-id> --cl 789012

# Update workspace
pm task update <task-id> --workspace "workspace-2"
pm task update <task-id> --ws "new-workspace"  # --ws is short for --workspace

# Update multiple fields at once
pm task update <task-id> --workspace "workspace-1" --status doing --priority high

# Complete a task (shortcut)
pm task complete <task-id>

# Delete a task
pm task delete <task-id>
```

## Workspace Management

Workspaces allow you to organize tasks by development environment, feature branch, or any other context. This is particularly useful when working on multiple tasks in different workspaces simultaneously.

### Using Workspaces

```bash
# Create a task with a workspace
pm task add "Implement feature X" --workspace "feature-branch-x" --project "web-app"

# List tasks in a specific workspace
pm task list --workspace "feature-branch-x"

# Update task workspace
pm task update <task-id> --workspace "main-workspace"

# Switch task to different workspace
pm task update <task-id> --ws "hotfix-workspace"
```

### Workspace Use Cases

1. **Feature Development**: Organize tasks by feature branch
   ```bash
   pm task add "Add login form" --workspace "feature/auth" --project "web-app"
   pm task add "Add JWT validation" --workspace "feature/auth" --project "web-app"
   ```

2. **Environment Separation**: Separate tasks by development environment
   ```bash
   pm task add "Test deployment" --workspace "staging"
   pm task add "Fix production bug" --workspace "production"
   ```

3. **Multi-tasking**: Keep track of different work contexts
   ```bash
   pm task list --workspace "client-a"
   pm task list --workspace "client-b"
   ```

4. **Team Coordination**: Share workspace names with team members
   ```bash
   pm task add "Review PR #123" --workspace "team-sprint-5"
   ```

**Note:** Workspace names are free-form text - use any naming convention that works for your workflow.

## Project Management

### Creating Projects
```bash
# Create a new project
pm project add "Mobile App"
```

### Listing Projects
```bash
# List all projects (shows name and ID)
pm project list
```

Example output:
```
Projects:
  Web Application (ID: b699950c-ab65-4d4d-8174-6f0dcf2fb6d4)
  Mobile App (ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890)
```

### Deleting Projects
```bash
# Delete a project
pm project delete <project-id>
```

**Important:** When you delete a project, all associated tasks and time entries are automatically deleted (cascade delete).

**Note:** Project status management (archive, complete, activate) is currently available in the TUI mode.

## Time Tracking

Time tracking is fully integrated and allows you to track time spent on tasks with detailed reporting.

### Starting Time Tracking
```bash
# Start tracking time for a task
pm time start --task <task-id>

# With description
pm time start --task <task-id> --description "Working on authentication"
```

**Note:** Starting time tracking automatically updates the task status to "doing" if it's not already.

### Stopping Time Tracking
```bash
# Stop current time tracking
pm time stop
```

This will display the duration and save the time entry to the database.

### Listing Time Entries
```bash
# List all time entries
pm time list
```

### Time Reports
```bash
# Today's time report
pm time report --today

# Yesterday's report
pm time report --yesterday

# This week's report
pm time report --week

# This month's report
pm time report --month
```

Reports show:
- Total duration
- Breakdown by task
- Task titles and IDs

## Git Integration

The CLI integrates with Git repositories to enhance your workflow.

### Commit Hooks
```bash
# Create commit hook for a task
pm git hook --task <task-id>

# Remove commit hook
pm git hook --remove
```

When you create a commit hook, the CLI will automatically prepend `[Task #<task-id>]` to your commit messages, helping you track which commits are related to which tasks.

**Example:**
```bash
# Create hook
pm git hook --task abc123

# Your commit
git commit -m "Fix authentication bug"

# Actual commit message will be:
# [Task #abc123] Fix authentication bug
```

## Export

The export functionality allows you to export your data to various formats for backup, reporting, or integration with other tools.

### Exporting Tasks
```bash
# Export tasks to JSON (default)
pm export tasks --format json

# Export to file
pm export tasks --format json --output tasks.json

# Export tasks to CSV
pm export tasks --format csv --output tasks.csv

# Export tasks with due dates to iCal format
pm export tasks --format ical --output tasks.ics
```

**Supported formats:** `json`, `csv`, `ical`

### Exporting Time Entries
```bash
# Export time entries to JSON
pm export time --format json

# Export to CSV file
pm export time --format csv --output timesheet.csv
```

**Supported formats:** `json`, `csv`

### Export Use Cases
- **Backup:** Export all data to JSON for safekeeping
- **Reporting:** Export time entries to CSV for billing or reports
- **Calendar Integration:** Export tasks with due dates to iCal for calendar apps
- **Data Analysis:** Export to CSV for analysis in Excel/Google Sheets

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
# Display current configuration
pm config show
```

This will show all configuration settings including database path, git integration status, time/date formats, theme colors, and aliases.

### Updating Configuration
```bash
# Enable/disable git integration
pm config set git_integration true
pm config set git_integration false

# Set default project
pm config set default_project "my-project"

# Customize time format
pm config set time_format "15:04"

# Customize date format
pm config set date_format "2006-01-02"
```

**Available configuration keys:**
- `git_integration` - Enable/disable git features (true/false)
- `default_project` - Default project for new tasks
- `time_format` - Time display format
- `date_format` - Date display format

**Note:** Theme and alias customization requires manual editing of `~/.pm/config.yaml`

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
6. Use workspaces to organize tasks by development environment or context
7. Track changelists (CLs) for code review workflows

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
pm --help
pm -h
```

## Advanced Features

### Scripting and Automation
The CLI is designed to be scriptable:

```bash
#!/bin/bash
# Daily standup report

echo "=== Daily Standup Report ==="
echo ""

echo "Time tracked yesterday:"
pm time report --yesterday

echo ""
echo "All current tasks:"
pm task list

echo ""
echo "Export backup:"
pm export tasks --format json --output backup-$(date +%Y%m%d).json
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

## Command Reference

### Complete Command List

```bash
# General
pm                              # Launch interactive TUI
pm help                         # Show help
pm version                      # Show version

# Task Commands
pm task add <title> [flags]     # Create task
pm task list [flags]            # List tasks
pm task update <id> [flags]     # Update task
pm task complete <id>           # Complete task
pm task delete <id>             # Delete task

# Task Flags
--priority <low|medium|high|critical>
--project <name>
--tags <tag1,tag2>
--cl <changelist>
--workspace <workspace-name>    # or --ws
--description <text>
--title <text>
--status <todo|doing|done|blocked>

# Project Commands
pm project add <name>           # Create project
pm project list                 # List projects
pm project delete <id>          # Delete project (cascades to tasks)

# Time Commands
pm time start --task <id>       # Start time tracking
pm time stop                    # Stop time tracking
pm time list                    # List time entries
pm time report [flags]          # Generate report

# Time Report Flags
--today                         # Today's report
--yesterday                     # Yesterday's report
--week                          # This week's report
--month                         # This month's report

# Export Commands
pm export tasks [flags]         # Export tasks
pm export time [flags]          # Export time entries

# Export Flags
--format <json|csv|ical>        # Output format
--output <file>                 # Output file path

# Config Commands
pm config show                  # Show configuration
pm config set <key> <value>     # Update configuration

# Git Commands
pm git hook --task <id>         # Create commit hook
pm git hook --remove            # Remove commit hook
```

## Support and Contribution

- Report issues on GitHub
- Contribute features via pull requests
- Join discussions for feature requests
- Share usage tips and workflows

For more information, visit the project repository or run `pm help` for built-in documentation.