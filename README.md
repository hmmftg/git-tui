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
│       └── main.go              # Application entry point
├── internal/
│   ├── git/
│   │   ├── git.go               # Git service interface
│   │   ├── mock.go              # Mock implementation for testing
│   │   └── git_test.go          # Unit tests
│   ├── models/
│   │   └── models.go            # Data models (views, step types, workflow)
│   ├── tui/
│   │   ├── app.go               # Main TUI application coordinator
│   │   ├── styles.go            # Styling system (lipgloss)
│   │   ├── base_model.go        # Embeddable base model for command screens
│   │   ├── messages.go          # Centralized message types
│   │   ├── conflict_handler.go  # Shared conflict resolution UI
│   │   ├── async_executor.go    # Async operation helper
│   │   ├── command.go           # CommandModel interface
│   │   ├── registry.go          # Command registry
│   │   ├── status.go            # Status screen
│   │   ├── commit.go            # Commit screen
│   │   ├── checkout.go          # Checkout screen
│   │   ├── push.go              # Push screen
│   │   ├── pull.go              # Pull screen
│   │   ├── merge.go             # Merge screen (uses ConflictHandler)
│   │   ├── rebase.go            # Rebase screen (uses ConflictHandler)
│   │   ├── workflow.go          # Workflow manager screen
│   │   ├── workflow_editor.go   # Workflow editor screen
│   │   └── components/
│   │       └── list.go          # bubbles/list wrapper
│   ├── workflow/
│   │   └── engine.go            # Workflow execution engine
│   └── config/
│       └── config.go            # Configuration management
├── go.mod
├── go.sum
└── README.md
```

## Architecture

### Elm Architecture Pattern

GitFlow TUI follows the Elm Architecture pattern:

- **Model** - Application state (struct with fields)
- **Update** - Message handler that updates model and returns commands
- **View** - Renders the UI based on current model state

### BaseModel Pattern

All command screens embed `BaseModel` for common functionality:

```go
type CommandModel struct {
    BaseModel           // GitSvc, Styles, Mode, Message, Err
    // screen-specific fields
}
```

**BaseModel provides:**
- `GitSvc git.GitService` - Git operations
- `Styles Styles` - UI styling
- `Mode CommandMode` - Execute or Configure mode
- `Message string` - Status message
- `Err error` - Error state

**Helper methods:**
- `SetMode(mode)` / `GetMode()` - Mode management
- `SetError(err)` / `SetSuccess(format, ...)` - Message helpers
- `Reset()` - Clear state
- `RenderMessage()` - Render styled message
- `IsExecuteMode()` / `IsConfigureMode()` - Mode checks

### ConflictHandler Pattern

Merge and Rebase screens embed `ConflictHandler` for conflict resolution:

```go
type MergeModel struct {
    BaseModel
    *ConflictHandler    // Shared conflict UI
    // merge-specific fields
}
```

**ConflictHandler provides:**
- `SetConflict(branch)` / `ClearConflict()` - Conflict state
- `HasConflict()` - Check conflict status
- `ViewBanner(styles)` - Render conflict warning banner
- `GetHelpText(baseHelp)` - Get contextual help text
- `HandleKey(msg, continueCmd, abortCmd)` - Key handling

### Command Modes

Each command screen operates in two modes:

- **ModeExecute** - Run the git operation immediately
- **ModeConfigure** - Set parameters for workflow step configuration

### Message Types

Centralized message types in `internal/tui/messages.go`:

- `ViewChangeMsg(view)` - Navigate to different view
- `CommandConfiguredMsg{Config}` - Workflow step configured
- `CommandCancelledMsg{} - Workflow step cancelled
- Operation-specific: `MergeSuccessMsg`, `RebaseContinueMsg`, etc.

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o gitflow ./cmd/gitflow
```

### Debug Configuration

VS Code launch configuration (`.vscode/launch.json`):

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "git",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/gitflow",
            "cwd": "${workspaceFolder}"
        }
    ]
}
```

### Adding a New Command Screen

1. Create file in `internal/tui/<command>.go`
2. Embed `BaseModel` in your struct
3. Implement `CommandModel` interface:
   - `Init() tea.Cmd`
   - `Update(msg) (tea.Model, tea.Cmd)`
   - `View() string`
   - `SetMode(mode)` / `GetMode() CommandMode`
   - `GetParameters() map[string]string`
   - `Execute() tea.Cmd`
   - `GetStepType() models.StepType`
4. Register in `internal/tui/registry.go`
5. Add message types to `messages.go` if needed

**Example skeleton:**

```go
type MyCommandModel struct {
    BaseModel
    // additional fields
}

func NewMyCommandModel(gitSvc git.GitService, styles Styles) *MyCommandModel {
    return &MyCommandModel{
        BaseModel: NewBaseModel(gitSvc, styles),
    }
}

func (m *MyCommandModel) Init() tea.Cmd {
    return nil
}

func (m *MyCommandModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // handle messages
    return m, nil
}

func (m *MyCommandModel) View() string {
    return m.Styles.Title.Render("My Command")
}

func (m *MyCommandModel) SetMode(mode CommandMode) {
    m.BaseModel.SetMode(mode)
}

func (m *MyCommandModel) GetMode() CommandMode {
    return m.BaseModel.GetMode()
}

func (m *MyCommandModel) GetParameters() map[string]string {
    if m.Mode != ModeConfigure { return nil }
    return map[string]string{}
}

func (m *MyCommandModel) Execute() tea.Cmd {
    if m.Mode != ModeExecute { return nil }
    return func() tea.Msg { return MySuccessMsg{} }
}

func (m *MyCommandModel) GetStepType() models.StepType {
    return models.StepMyCommand
}
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) v1.3.10 - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) v1.1.1 - Terminal styling
- [bubbles](https://github.com/charmbracelet/bubbles) v1.0.0 - UI components

## License

MIT License
