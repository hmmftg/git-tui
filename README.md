# GitFlow TUI

A cross-platform Git Workflow Manager built with Go and the Charmbracelet ecosystem.

## Features

- **Interactive Git Operations** - Status, commit, push, pull, merge, rebase with a beautiful TUI
- **Repository Discovery** - Auto-detects git repositories and shows current branch info
- **Workflow Automation** - Create and execute multi-step git workflows
- **Saved Templates** - Built-in workflow templates (Quick Commit, Release, Sync)
- **Live Execution Monitoring** - Visual progress indicators and streaming output
- **Merge/Rebase Assistance** - Interactive branch selection and conflict detection

## Installation

### From Source

```bash
git clone https://github.com/yourusername/gitflow-tui
cd gitflow-tui
go build -o gitflow ./cmd/gitflow
```

### Prerequisites

- Go 1.21 or higher
- Git installed and available in PATH

## Usage

Navigate to any git repository and run:

```bash
gitflow
```

### Navigation

| Key | Action |
|-----|--------|
| `1` | View Status |
| `2` | Create Commit |
| `3` | Push to Remote |
| `4` | Pull from Remote |
| `5` | Checkout Branch |
| `6` | Merge Branch |
| `7` | Rebase Branch |
| `8` | Workflows |
| `q` | Quit |
| `esc` | Go Back |
| `r` | Refresh |

## Project Structure

```
gitflow-tui/
├── cmd/
│   └── gitflow/
│       └── main.go          # Application entry point
├── internal/
│   ├── git/
│   │   ├── git.go           # Git service interface
│   │   ├── mock.go          # Mock implementation for testing
│   │   └── git_test.go      # Unit tests
│   ├── models/
│   │   └── models.go        # Data models
│   ├── tui/
│   │   ├── app.go           # Main TUI application
│   │   ├── styles.go        # Styling system
│   │   ├── status.go        # Status screen
│   │   ├── commit.go        # Commit screen
│   │   ├── checkout.go      # Checkout screen
│   │   ├── push.go          # Push screen
│   │   ├── pull.go          # Pull screen
│   │   ├── merge.go         # Merge screen
│   │   ├── rebase.go        # Rebase screen
│   │   └── workflow.go      # Workflow manager screen
│   ├── workflow/
│   │   └── engine.go        # Workflow execution engine
│   └── config/
│       └── config.go        # Configuration management
├── go.mod
├── go.sum
└── README.md
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o gitflow ./cmd/gitflow
```

## License

MIT License
