# Non-Interactive Usage Guide

This guide covers using the Project Manager CLI in non-interactive mode for automation, scripting, and integration with other tools.

## Table of Contents

- [Overview](#overview)
- [Task Management](#task-management)
- [Project Management](#project-management)
- [Time Tracking](#time-tracking)
- [Scripting & Automation](#scripting--automation)
- [CI/CD Integration](#cicd-integration)
- [Data Export & Processing](#data-export--processing)

## Overview

The CLI can be used entirely without the interactive TUI, making it perfect for:
- Shell scripts and automation
- CI/CD pipelines
- Git hooks
- Cron jobs
- Integration with other tools

All commands output structured data suitable for piping and processing.

## Task Management

### Creating Tasks

```bash
# Basic task creation
pm task add "Implement user authentication"

# With priority (low, medium, high, critical)
pm task add "Fix security vulnerability" --priority critical

# With tags
pm task add "Update documentation" --tags docs,maintenance

# With due date
pm task add "Deploy v2.0" --due "2024-03-15"
pm task add "Weekly meeting" --due "next friday"

# With project association
pm task add "Setup database schema" --project backend

# Combining options
pm task add "Review pull requests" \
  --priority high \
  --tags review,urgent \
  --due today \
  --project web-app
```

### Listing & Filtering Tasks

```bash
# List all tasks
pm task list

# Filter by status
pm task list --status todo
pm task list --status doing
pm task list --status done
pm task list --status blocked

# Filter by priority
pm task list --priority high
pm task list --priority critical

# Filter by project
pm task list --project backend

# Filter by tags
pm task list --tags urgent
pm task list --tags deployment,production

# Filter by due date
pm task list --due today
pm task list --due "this week"
pm task list --overdue

# Search by text
pm task list --search "authentication"

# Combine filters
pm task list --status todo --priority high --project web-app
```

### Updating Tasks

```bash
# Change task status
pm task start <task-id>
pm task complete <task-id>
pm task block <task-id>
pm task unblock <task-id>

# Update priority
pm task update <task-id> --priority critical

# Update due date
pm task update <task-id> --due "next monday"

# Add/update tags
pm task update <task-id> --tags bug,urgent

# Move to different project
pm task update <task-id> --project mobile-app
```

### Deleting Tasks

```bash
# Delete single task
pm task delete <task-id>

# Bulk delete (with confirmation)
pm task list --status done | xargs -n1 pm task delete
```

## Project Management

### Creating Projects

```bash
# Basic project
pm project add "Web Application"

# With description
pm project add "Mobile App" --description "iOS and Android application"

# With status
pm project add "Legacy System" --status archived
```

### Listing Projects

```bash
# All projects
pm project list

# Active projects only
pm project list --status active

# Archived projects
pm project list --status archived
```

### Managing Projects

```bash
# Archive project
pm project archive <project-id>

# Complete project
pm project complete <project-id>

# Delete project
pm project delete <project-id>
```

## Time Tracking

### Starting & Stopping Time

```bash
# Start tracking time on a task
pm time start --task <task-id>

# Start with description
pm time start --task <task-id> --description "Debugging authentication flow"

# Stop current timer
pm time stop
```

### Time Reports

```bash
# Today's time
pm time report --today

# Yesterday
pm time report --yesterday

# Current week
pm time report --week

# Current month
pm time report --month

# Custom date range
pm time report --start "2024-01-01" --end "2024-01-31"

# Filter by project
pm time report --week --project web-app

# Filter by task
pm time report --month --task <task-id>
```

## Scripting & Automation

### Daily Standup Script

```bash
#!/bin/bash
# daily-standup.sh

echo "=== Daily Standup Report ==="
echo ""

echo "✅ Completed Yesterday:"
pm task list --status done --completed-after yesterday
echo ""

echo "🎯 Planned for Today:"
pm task list --status todo --due today
echo ""

echo "🔄 In Progress:"
pm task list --status doing
echo ""

echo "⏱️  Time Tracked Yesterday:"
pm time report --yesterday
```

### Weekly Review Script

```bash
#!/bin/bash
# weekly-review.sh

echo "=== Weekly Review ==="
echo ""

echo "Tasks Completed This Week:"
pm task list --status done --completed-after "7 days ago"
echo ""

echo "Total Time Tracked:"
pm time report --week
echo ""

echo "Overdue Tasks:"
pm task list --overdue
echo ""

echo "High Priority Tasks:"
pm task list --status todo --priority high
```

### Git Commit Hook

```bash
#!/bin/bash
# .git/hooks/prepare-commit-msg

# Get current branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)

# List tasks for current branch
TASK_ID=$(pm task list --branch $BRANCH --status doing | head -n1 | awk '{print $1}')

if [ ! -z "$TASK_ID" ]; then
  # Prepend task ID to commit message
  echo "[$TASK_ID] $(cat $1)" > $1
fi
```

### Pomodoro Timer Script

```bash
#!/bin/bash
# pomodoro.sh

TASK_ID=$1
DURATION=${2:-25}  # Default 25 minutes

if [ -z "$TASK_ID" ]; then
  echo "Usage: pomodoro.sh <task-id> [duration-in-minutes]"
  exit 1
fi

# Start time tracking
pm time start --task $TASK_ID --description "Pomodoro session"

# Work for specified duration
echo "🍅 Working for $DURATION minutes..."
sleep ${DURATION}m

# Stop time tracking
pm time stop

# Notify completion
echo "✅ Pomodoro complete! Take a break."
```

### Task Creation from Template

```bash
#!/bin/bash
# create-sprint-tasks.sh

PROJECT=$1
SPRINT=$2

if [ -z "$PROJECT" ] || [ -z "$SPRINT" ]; then
  echo "Usage: create-sprint-tasks.sh <project> <sprint-number>"
  exit 1
fi

# Create standard sprint tasks
pm task add "Sprint $SPRINT planning" --project $PROJECT --priority high
pm task add "Sprint $SPRINT review" --project $PROJECT --due "2 weeks"
pm task add "Sprint $SPRINT retrospective" --project $PROJECT --due "2 weeks"
pm task add "Update documentation" --project $PROJECT --tags docs
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Daily Task Report

on:
  schedule:
    - cron: '0 9 * * 1-5'  # 9 AM weekdays

jobs:
  report:
    runs-on: ubuntu-latest
    steps:
      - name: Install PM CLI
        run: |
          wget https://github.com/adriannajera/project-manager-cli/releases/latest/download/pm-linux-amd64.tar.gz
          tar -xzf pm-linux-amd64.tar.gz
          sudo mv pm /usr/local/bin/

      - name: Generate Report
        run: |
          pm task list --status todo --due today > daily-tasks.txt
          pm time report --yesterday > time-report.txt

      - name: Send to Slack
        run: |
          curl -X POST ${{ secrets.SLACK_WEBHOOK }} \
            -H 'Content-Type: application/json' \
            -d "{\"text\":\"$(cat daily-tasks.txt)\"}"
```

### Jenkins Pipeline Example

```groovy
pipeline {
    agent any

    stages {
        stage('Track Deployment Task') {
            steps {
                sh 'pm task add "Deploy build ${BUILD_NUMBER}" --project production --priority high'
            }
        }

        stage('Deploy') {
            steps {
                sh './deploy.sh'
            }
        }

        stage('Mark Complete') {
            steps {
                sh 'TASK_ID=$(pm task list --search "Deploy build ${BUILD_NUMBER}" | head -n1 | awk "{print \\$1}")'
                sh 'pm task complete $TASK_ID'
            }
        }
    }
}
```

## Data Export & Processing

### Export Tasks to JSON

```bash
# Export all tasks
pm export tasks --format json --output tasks.json

# Export filtered tasks
pm export tasks --status done --format json --output completed.json

# Process with jq
pm export tasks --format json | jq '.[] | select(.priority == "high")'

# Extract specific fields
pm export tasks --format json | jq -r '.[] | "\(.id): \(.title)"'
```

### Export to CSV for Spreadsheets

```bash
# Export tasks
pm export tasks --format csv --output tasks.csv

# Export time tracking
pm export time --format csv --output timesheet.csv

# Filter by date range
pm export time --start "2024-01-01" --end "2024-01-31" --format csv
```

### Export to Calendar (iCal)

```bash
# Export tasks with due dates
pm export tasks --format ical --output tasks.ics

# Import into calendar applications (Google Calendar, Apple Calendar, etc.)
```

### Batch Processing Example

```bash
#!/bin/bash
# bulk-update-tags.sh

# Add "migration" tag to all tasks in old-project
pm export tasks --project old-project --format json | \
  jq -r '.[].id' | \
  while read task_id; do
    pm task update $task_id --tags migration
  done
```

### Generate HTML Report

```bash
#!/bin/bash
# generate-report.sh

cat > report.html <<EOF
<!DOCTYPE html>
<html>
<head>
  <title>Task Report</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 20px; }
    table { border-collapse: collapse; width: 100%; }
    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
    th { background-color: #4CAF50; color: white; }
  </style>
</head>
<body>
  <h1>Project Task Report</h1>
  <h2>Tasks by Status</h2>
  <pre>$(pm task list)</pre>

  <h2>Time Tracking Summary</h2>
  <pre>$(pm time report --week)</pre>
</body>
</html>
EOF

echo "Report generated: report.html"
```

### Integration with Other Tools

```bash
# Create tasks from GitHub issues
gh issue list --json title,body,labels | \
  jq -r '.[] | "pm task add \"\(.title)\" --tags \(.labels[].name | join(","))"' | \
  bash

# Create tasks from Jira
curl -u user:token "https://your-domain.atlassian.net/rest/api/3/search?jql=project=PROJ" | \
  jq -r '.issues[] | "pm task add \"\(.fields.summary)\" --priority high"' | \
  bash

# Sync completed tasks to external system
pm export tasks --status done --format json | \
  curl -X POST https://api.example.com/tasks \
    -H "Content-Type: application/json" \
    -d @-
```

## Tips for Non-Interactive Usage

1. **Use JSON output for parsing**: Most commands support `--format json` for machine-readable output
2. **Capture task IDs**: Store task IDs in variables for subsequent operations
3. **Use filters**: Narrow down results with status, priority, project, and tag filters
4. **Error handling**: Check exit codes in scripts (`$?`) to handle failures
5. **Configuration**: Set up aliases in `~/.pm/config.yaml` for shorter commands
6. **Batch operations**: Use `xargs` or loops for bulk updates
7. **Date parsing**: Leverage natural language dates ("today", "next week", "2 days ago")

## Configuration for Scripting

```yaml
# ~/.pm/config.yaml
database_path: ~/.pm/tasks.db
git_integration: true

# Useful aliases for scripting
aliases:
  ls: "list"
  new: "add"
  rm: "delete"
  done: "complete"
  start: "start"
  stop: "stop"
```

## Environment Variables

```bash
# Override database location
export PM_DATABASE_PATH=/path/to/tasks.db

# Override config file
export PM_CONFIG_PATH=/path/to/config.yaml

# Use in scripts
PM_DATABASE_PATH=/tmp/test.db pm task list
```

---

**For more examples and use cases, see the main [README](../README.md) and [contributing guide](CONTRIBUTING.md).**
