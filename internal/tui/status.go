package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// StatusModel handles the status screen
type StatusModel struct {
	BaseModel
	status StatusResult

	// Configuration options
	verbose          bool
	includeUntracked bool
}

// StatusResult holds the git status data
type StatusResult struct {
	Branch     string
	IsClean    bool
	Modified   []string
	Added      []string
	Deleted    []string
	Untracked  []string
	Conflicted []string
	Ahead      int
	Behind     int
}

// NewStatusModel creates a new status model
func NewStatusModel(gitSvc git.GitService, styles Styles) *StatusModel {
	return &StatusModel{
		BaseModel:        NewBaseModel(gitSvc, styles),
		verbose:          false,
		includeUntracked: true,
	}
}

// Init initializes the model
func (m *StatusModel) Init() tea.Cmd {
	return m.loadStatus()
}

// Update handles messages
func (m *StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Mode == ModeConfigure {
			switch msg.String() {
			case "v":
				// Toggle verbose option
				m.verbose = !m.verbose
			case "u":
				// Toggle include untracked option
				m.includeUntracked = !m.includeUntracked
			case "enter":
				// Save configuration
				config := models.CommandConfig{
					StepType:   models.StepStatus,
					Parameters: m.GetParameters(),
				}
				return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
			case "esc":
				return m, func() tea.Msg { return commandCancelledMsg{} }
			}
		} else {
			switch msg.String() {
			case "r":
				return m, m.loadStatus()
			case "w":
				return m, NavigateCmd(models.ViewWorkflowBuilder)
			}
		}

	case statusLoadedMsg:
		if m.Mode == ModeExecute {
			m.status = StatusResult(msg)
			m.Err = nil
		}
		return m, nil

	case error:
		if m.Mode == ModeExecute {
			m.Err = msg
		}
		return m, nil
	}

	return m, nil
}

// View renders the status screen
func (m *StatusModel) View() string {
	if m.Err != nil {
		return m.Styles.Error.Render(fmt.Sprintf("Error: %v", m.Err))
	}

	var lines []string

	// Title
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Title.Render(" Configure Status Step "))
	} else {
		lines = append(lines, m.Styles.Title.Render(" Repository Status "))
	}
	lines = append(lines, "")

	if m.Mode == ModeConfigure {
		// Configuration options
		lines = append(lines, m.Styles.Help.Render("Configure status step options:"))
		lines = append(lines, "")

		// Verbose option
		verboseStyle := m.Styles.Info
		if m.verbose {
			verboseStyle = m.Styles.Success
		}
		verboseText := "Verbose output: "
		if m.verbose {
			verboseText += "✓ Yes"
		} else {
			verboseText += "✗ No"
		}
		lines = append(lines, verboseStyle.Render(verboseText))

		// Include untracked option
		untrackedStyle := m.Styles.Info
		if m.includeUntracked {
			untrackedStyle = m.Styles.Success
		}
		untrackedText := "Include untracked files: "
		if m.includeUntracked {
			untrackedText += "✓ Yes"
		} else {
			untrackedText += "✗ No"
		}
		lines = append(lines, untrackedStyle.Render(untrackedText))
		lines = append(lines, "")

		// Help
		lines = append(lines, m.Styles.Help.Render("v: toggle verbose | u: toggle untracked | enter: save | esc: cancel"))
	} else {
		// Execute mode - show status information
		// Branch info
		branchInfo := fmt.Sprintf("🔀 Branch: %s", m.Styles.Value.Render(m.status.Branch))
		if m.status.Ahead > 0 {
			branchInfo += fmt.Sprintf(" | ⬆ Ahead: %d", m.status.Ahead)
		}
		if m.status.Behind > 0 {
			branchInfo += fmt.Sprintf(" | ⬇ Behind: %d", m.status.Behind)
		}
		lines = append(lines, branchInfo)
		lines = append(lines, "")

		// Repository state
		if m.status.IsClean {
			lines = append(lines, m.Styles.Success.Render("✔ Working tree clean"))
		} else {
			lines = append(lines, m.Styles.Warning.Render("⚡ Changes detected"))
		}
		lines = append(lines, "")

		// Modified files
		if len(m.status.Modified) > 0 {
			lines = append(lines, m.Styles.Warning.Render(fmt.Sprintf("📝 Modified (%d):", len(m.status.Modified))))
			for _, f := range m.status.Modified {
				lines = append(lines, fmt.Sprintf("  • %s", f))
			}
			lines = append(lines, "")
		}

		// Added files
		if len(m.status.Added) > 0 {
			lines = append(lines, m.Styles.Success.Render(fmt.Sprintf("✚ Staged (%d):", len(m.status.Added))))
			for _, f := range m.status.Added {
				lines = append(lines, fmt.Sprintf("  • %s", f))
			}
			lines = append(lines, "")
		}

		// Deleted files
		if len(m.status.Deleted) > 0 {
			lines = append(lines, m.Styles.Error.Render(fmt.Sprintf("🗑 Deleted (%d):", len(m.status.Deleted))))
			for _, f := range m.status.Deleted {
				lines = append(lines, fmt.Sprintf("  • %s", f))
			}
			lines = append(lines, "")
		}

		// Untracked files
		if len(m.status.Untracked) > 0 {
			lines = append(lines, m.Styles.Info.Render(fmt.Sprintf("❔ Untracked (%d):", len(m.status.Untracked))))
			for _, f := range m.status.Untracked {
				lines = append(lines, fmt.Sprintf("  • %s", f))
			}
			lines = append(lines, "")
		}

		// Conflicts
		if len(m.status.Conflicted) > 0 {
			lines = append(lines, m.Styles.Error.Render(fmt.Sprintf("⚠ Conflicts (%d):", len(m.status.Conflicted))))
			for _, f := range m.status.Conflicted {
				lines = append(lines, fmt.Sprintf("  • %s", f))
			}
			lines = append(lines, "")
		}

		// Summary
		lines = append(lines, m.Styles.Help.Render("r: refresh | w: workflows | esc: back"))
	}

	return m.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadStatus loads the git status
func (m *StatusModel) loadStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.GitSvc.Status()
		if err != nil {
			return err
		}

		return statusLoadedMsg(StatusResult{
			Branch:     status.Branch,
			IsClean:    status.IsClean,
			Modified:   status.Modified,
			Added:      status.Added,
			Deleted:    status.Deleted,
			Untracked:  status.Untracked,
			Conflicted: status.Conflicted,
			Ahead:      status.Ahead,
			Behind:     status.Behind,
		})
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *StatusModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
}

// GetMode returns the current operating mode
func (m *StatusModel) GetMode() CommandMode {
	return m.BaseModel.GetMode()
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *StatusModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	params["verbose"] = "false"
	if m.verbose {
		params["verbose"] = "true"
	}
	params["includeUntracked"] = "false"
	if m.includeUntracked {
		params["includeUntracked"] = "true"
	}
	return params
}

// Execute returns a command to execute the operation (only valid in execute mode)
func (m *StatusModel) SetParameters(params map[string]string) {
	if params == nil {
		return
	}
	m.verbose = params["verbose"] == "true"
	if includeUntracked, ok := params["includeUntracked"]; ok {
		m.includeUntracked = includeUntracked == "true"
	}
}

func (m *StatusModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	return m.loadStatus()
}

// GetStepType returns the step type this command model represents
func (m *StatusModel) GetStepType() models.StepType {
	return models.StepStatus
}

type statusLoadedMsg StatusResult
