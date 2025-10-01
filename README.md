# Project Manager CLI

> A powerful, production-ready CLI tool for developers to manage projects locally with beautiful terminal interfaces

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## ✨ Features

- 🎯 **Task Management** - Create, organize, and track tasks with priorities, tags, due dates, and changelist tracking
- ⏱️ **Time Tracking** - Built-in timer with detailed reporting and analytics
- 🌳 **Git Integration** - Automatic branch association and commit message hooks
- 📝 **Changelist Support** - Associate tasks with changelists/CLs for better code review workflow
- 🎨 **Beautiful TUI** - Interactive dashboard powered by Charm libraries (Bubble Tea + Lip Gloss)
- 📊 **Export & Reports** - JSON, CSV, and iCal exports with customizable filters
- 🔍 **Smart Search** - Natural language date parsing and advanced filtering
- 📱 **Cross-Platform** - Single binary for Linux, macOS, and Windows
- 🗃️ **Local-First** - SQLite database, no cloud dependencies, complete data ownership

## 🚀 Quick Start

### Installation

**Option 1: Download Pre-built Binary**
```bash
# Download latest release for your platform
# Extract and move to your PATH
wget https://github.com/adriannajera/project-manager-cli/releases/latest/download/pm-linux-amd64.tar.gz
tar -xzf pm-linux-amd64.tar.gz
sudo mv pm /usr/local/bin/
```

**Option 2: Build from Source**
```bash
git clone https://github.com/adriannajera/project-manager-cli
cd project-manager-cli
make build
sudo make install-system
```

**Option 3: Go Install**
```bash
go install github.com/adriannajera/project-manager-cli/cmd/pm@latest
```

### First Steps

```bash
# Check installation
pm version

# Create your first task
pm task add "Set up development environment"

# Launch interactive dashboard
pm

# Or use CLI commands
pm task list
pm task complete <task-id>
```

## 📋 Task Management

### Creating Tasks
```bash
# Basic task creation
pm task add "Fix authentication bug"
pm task add "Implement user profiles" --priority high

# With projects and tags
pm task add "Deploy to production" --project web-app --tags deployment,urgent

# With changelist/CL tracking
pm task add "Fix login flow" --cl 123456
pm task add "Add user validation" --changelist cl/789012
```

### Managing Tasks
```bash
# List all tasks
pm task list

# Filter by status
pm task list --status todo
pm task list --status doing
pm task list --status done

# Filter by project
pm task list --project web-app

# Search tasks
pm task list --search "authentication"

# Update task status and properties
pm task update <task-id> --status doing
pm task update <task-id> --priority high
pm task update <task-id> --cl 123456
pm task update <task-id> --title "New title"

# Quick status changes
pm task complete <task-id>

# Delete task
pm task delete <task-id>
```

## 📁 Project Management

```bash
# Create projects
pm project add "Web Application"
pm project add "Mobile App" --description "iOS and Android app"

# List projects
pm project list

# Manage project status
pm project archive <project-id>
pm project complete <project-id>
```

## ⏰ Time Tracking

```bash
# Start tracking time
pm time start --task <task-id>
pm time start --task <task-id> --description "Working on login feature"

# Stop tracking
pm time stop

# View reports
pm time report --today
pm time report --week
pm time report --month
pm time report --start "2023-01-01" --end "2023-01-31"
```

## 🎨 Interactive Mode (TUI)

Launch the beautiful terminal interface:

```bash
pm
```

**Navigation:**
- `↑/↓` or `j/k` - Navigate lists
- `Enter` - Select/confirm
- `n` - New task/project
- `e` - Edit selected item
- `d` - Delete selected item
- `t` - Toggle task status
- `s` - Start time tracking
- `S` - Stop time tracking
- `?` - Help
- `q` - Quit

## 🌳 Git Integration

When working in a Git repository, the CLI automatically:

- Associates tasks with the current branch
- Can create commit hooks to reference task IDs
- Tracks which branch a task was worked on

```bash
# Manual Git operations
pm git hook --task <task-id>     # Create commit hook
pm git hook --remove             # Remove commit hook
pm task list --branch current    # List tasks for current branch
```

## 📤 Export & Import

```bash
# Export tasks
pm export tasks --format json --output tasks.json
pm export tasks --format csv --output tasks.csv
pm export tasks --format ical --output tasks.ics

# Export with filters
pm export tasks --status done --format json
pm export tasks --project "web-app" --format csv

# Export time tracking data
pm export time --format csv --output timesheet.csv
pm export time --start "2023-01-01" --end "2023-01-31" --format csv
```

## ⚙️ Configuration

The CLI uses `~/.pm/config.yaml` for configuration:

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

```bash
# View configuration
pm config show

# Update settings
pm config set git_integration false
pm config set default_project "my-project"
```

## 🔧 Development

### Prerequisites
- Go 1.21 or later
- Make (optional, for build automation)

### Building from Source
```bash
# Clone repository
git clone https://github.com/adriannajera/project-manager-cli
cd project-manager-cli

# Install dependencies
make deps

# Run tests
make test

# Build binary
make build

# Install locally
make install-system
```

### Available Make Targets
```bash
make build         # Build binary
make test          # Run tests
make test-coverage # Run tests with coverage
make check         # Run all checks (fmt, vet, lint, test)
make clean         # Clean build artifacts
make install       # Install to GOPATH/bin
make release       # Create release packages
```

### Project Structure
```
├── cmd/pm/                    # Main application entry point
├── internal/
│   ├── domain/               # Core business logic and entities
│   ├── repository/sqlite/    # Data persistence layer
│   ├── service/             # Business logic services
│   └── ui/                  # Terminal user interface
├── pkg/config/              # Configuration management
├── docs/                    # Documentation
└── Makefile                 # Build automation
```

## 📚 Advanced Usage

### Automation & Scripting
```bash
#!/bin/bash
# Daily standup script
echo "=== Daily Standup Report ==="
echo "Completed yesterday:"
pm task list --status done --completed-after yesterday

echo "Planned for today:"
pm task list --status todo --due today

echo "Time tracked yesterday:"
pm time report --yesterday
```

### Integration Examples
```bash
# Export for external tools
pm export tasks --format json | jq '.[] | select(.priority > 1)'

# Create tasks from issues
gh issue list --json title,body | jq -r '.[] | "pm task add \"\(.title)\""' | bash

# Weekly time report
pm time report --week --format csv > weekly_timesheet.csv
```

## 🔍 Tips & Best Practices

### Task Organization
- Use descriptive titles and add context in descriptions
- Apply consistent tagging for easy filtering
- Set realistic due dates and priorities
- Break large tasks into smaller, manageable subtasks

### Time Tracking
- Start tracking when you begin work
- Add meaningful descriptions to time entries
- Review reports regularly for insights
- Use data for better project estimation

### Git Workflow
- Create tasks before starting new features
- Let the CLI create commit hooks for traceability
- Review task status when merging branches
- Use branch-based task filtering

## 🐛 Troubleshooting

### Common Issues

**Database errors:**
```bash
# Reset database (will lose all data)
rm ~/.pm/tasks.db
pm task list  # Recreates database
```

**Permission issues:**
```bash
chmod 644 ~/.pm/tasks.db
chmod 755 ~/.pm/
```

**Configuration problems:**
```bash
# Reset to defaults
rm ~/.pm/config.yaml
pm config show
```

### Getting Help
```bash
pm help                    # General help
pm task help              # Task commands help
pm time help              # Time tracking help
pm task add --help        # Specific command help
```

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Charm](https://charm.sh/) libraries (Bubble Tea, Lip Gloss)
- Inspired by tools like TaskWarrior and Todo.txt
- Natural language date parsing by [olebedev/when](https://github.com/olebedev/when)

## 📞 Support

- 📖 [Full Documentation](docs/USAGE.md)
- 🐛 [Report Issues](https://github.com/adriannajera/project-manager-cli/issues)
- 💬 [Discussions](https://github.com/adriannajera/project-manager-cli/discussions)

---

**Made with ❤️ for developers who love the terminal**