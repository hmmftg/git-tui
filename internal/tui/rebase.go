package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// RebaseModel handles the rebase screen
type RebaseModel struct {
	BaseModel
	*ConflictHandler
	branches     []string
	cursor       int
	success      bool
	targetBranch string

	// Configuration options
	interactive bool
	autoStash   bool
}

// NewRebaseModel creates a new rebase model
func NewRebaseModel(gitSvc git.GitService, styles Styles) *RebaseModel {
	return &RebaseModel{
		BaseModel:       NewBaseModel(gitSvc, styles),
		ConflictHandler: NewConflictHandler("rebase"),
		interactive:     false,
		autoStash:       false,
	}
}

// Init initializes the model
func (m *RebaseModel) Init() tea.Cmd {
	return m.loadBranches()
}

// Update handles messages
func (m *RebaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Mode == ModeConfigure {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.branches)-1 {
					m.cursor++
				}
			case "i":
				// Toggle interactive option
				m.interactive = !m.interactive
			case "s":
				// Toggle auto-stash option
				m.autoStash = !m.autoStash
			case "enter":
				// Save configuration
				config := models.CommandConfig{
					StepType:   models.StepRebase,
					Parameters: m.GetParameters(),
				}
				return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
			case "esc":
				return m, func() tea.Msg { return commandCancelledMsg{} }
			}
		} else {
			// Execute mode
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.branches)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.branches) > 0 && m.cursor < len(m.branches) {
					return m, m.rebase(m.branches[m.cursor])
				}
			case "R":
				return m, m.loadBranches()
			case "c":
				// Continue rebase after resolving conflicts
				if m.HasConflict() {
					return m, m.continueRebase()
				}
			case "a":
				// Abort rebase
				if m.HasConflict() {
					return m, m.abortRebase()
				}
			case "r":
				// Open conflict resolver
				if m.HasConflict() {
					return m, func() tea.Msg { return OpenConflictResolverMsg{} }
				}
			}
		}

	case branchesLoadedMsg:
		m.branches = []string(msg)
		m.Err = nil
		if m.targetBranch != "" {
			for i, branch := range m.branches {
				if branch == m.targetBranch {
					m.cursor = i
					break
				}
			}
		}
		return m, nil

	case RebaseSuccessMsg:
		// Check for conflicts after rebase
		if m.GitSvc.HasConflicts() {
			m.SetConflict(string(msg))
			m.Message = m.GetConflictMessage(string(msg))
			return m, nil
		}
		m.success = true
		m.SetSuccess("Rebased onto %s successfully!", msg)
		m.ClearConflict()
		return m, nil

	case RebaseContinueMsg:
		// Check if conflicts are resolved
		if m.GitSvc.HasConflicts() {
			m.Message = m.GetStillConflictMessage()
			return m, nil
		}
		m.success = true
		m.SetSuccess("Rebase completed successfully!")
		m.ClearConflict()
		return m, nil

	case RebaseAbortedMsg:
		m.Reset()
		m.SetSuccess("Rebase aborted")
		return m, nil

	case error:
		if m.GitSvc.HasConflicts() {
			m.SetConflict(m.targetBranch)
			m.Message = m.GetConflictErrorMessage(msg)
		} else {
			m.SetError(msg)
		}
		return m, nil
	}

	return m, nil
}

// View renders the rebase screen
func (m *RebaseModel) View() string {
	var lines []string

	// Title
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Title.Render(" Configure Rebase Step "))
	} else {
		lines = append(lines, m.Styles.Title.Render(" Rebase Branch "))
	}
	lines = append(lines, "")

	if m.Mode == ModeConfigure {
		// Configuration options
		lines = append(lines, m.Styles.Help.Render("Configure rebase step options:"))
		lines = append(lines, "")

		// Branch selection
		lines = append(lines, m.Styles.Help.Render("Select base branch to rebase ONTO:"))
		lines = append(lines, "")

		// Branch list
		if len(m.branches) == 0 {
			lines = append(lines, m.Styles.Info.Render("Loading branches..."))
		} else {
			current, _ := m.GitSvc.CurrentBranch()
			for i, branch := range m.branches {
				cursor := "  "
				if m.cursor == i {
					cursor = m.Styles.Key.Render("▸ ")
				}

				// Don't show current branch
				if branch == current {
					lines = append(lines, fmt.Sprintf("%s%s (current)", cursor, m.Styles.Dimmed.Render(branch)))
				} else {
					lines = append(lines, fmt.Sprintf("%s%s", cursor, branch))
				}
			}
		}
		lines = append(lines, "")

		// Interactive option
		interactiveStyle := m.Styles.Info
		if m.interactive {
			interactiveStyle = m.Styles.Success
		}
		interactiveText := "Interactive rebase: "
		if m.interactive {
			interactiveText += "✓ Yes"
		} else {
			interactiveText += "✗ No"
		}
		lines = append(lines, interactiveStyle.Render(interactiveText))

		// Auto-stash option
		autoStashStyle := m.Styles.Info
		if m.autoStash {
			autoStashStyle = m.Styles.Success
		}
		autoStashText := "Auto stash: "
		if m.autoStash {
			autoStashText += "✓ Yes"
		} else {
			autoStashText += "✗ No"
		}
		lines = append(lines, autoStashStyle.Render(autoStashText))
		lines = append(lines, "")

		// Help
		lines = append(lines, m.Styles.Help.Render("↑/↓: navigate | i: toggle interactive | s: toggle auto-stash | enter: save | esc: cancel"))
	} else {
		// Execute mode
		// Description
		current, _ := m.GitSvc.CurrentBranch()
		lines = append(lines, m.Styles.Info.Render(fmt.Sprintf("Current branch: %s", current)))
		lines = append(lines, "")
		lines = append(lines, m.Styles.Help.Render("Select base branch to rebase ONTO:"))
		lines = append(lines, "")

		// Branch list
		if len(m.branches) == 0 {
			lines = append(lines, m.Styles.Info.Render("Loading branches..."))
		} else {
			for i, branch := range m.branches {
				cursor := "  "
				if m.cursor == i {
					cursor = m.Styles.Key.Render("▸ ")
				}

				// Don't show current branch
				if branch == current {
					lines = append(lines, fmt.Sprintf("%s%s (current)", cursor, m.Styles.Dimmed.Render(branch)))
				} else {
					lines = append(lines, fmt.Sprintf("%s%s", cursor, branch))
				}
			}
		}

		lines = append(lines, "")
	}

	// Conflict warning banner and message
	if banner := m.ConflictHandler.ViewBanner(m.Styles); banner != "" {
		lines = append(lines, banner)
		lines = append(lines, "")
	}

	// Message
	if m.HasMessage() {
		if m.HasConflict() {
			lines = append(lines, m.Styles.Warning.Render(m.Message))
		} else {
			lines = append(lines, m.RenderMessage())
		}
		lines = append(lines, "")
	}

	// Help
	help := m.ConflictHandler.GetHelpText("↑/↓: navigate | enter: rebase | R: refresh | esc: back")
	lines = append(lines, m.Styles.Help.Render(help))

	return m.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadBranches loads all branches
func (m *RebaseModel) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.GitSvc.GetBranches()
		if err != nil {
			return err
		}
		return branchesLoadedMsg(branches)
	}
}

// rebase rebases onto the selected branch
func (m *RebaseModel) rebase(branch string) tea.Cmd {
	m.targetBranch = branch
	return func() tea.Msg {
		err := m.GitSvc.Rebase(branch)
		if err != nil {
			return err
		}
		return RebaseSuccessMsg(branch)
	}
}

// continueRebase continues the rebase after conflicts are resolved
func (m *RebaseModel) continueRebase() tea.Cmd {
	return func() tea.Msg {
		// Check if still has conflicts
		if m.GitSvc.HasConflicts() {
			return fmt.Errorf("still has unresolved conflicts")
		}
		// Continue the rebase
		err := m.GitSvc.ContinueRebase()
		if err != nil {
			return err
		}
		return RebaseContinueMsg{}
	}
}

// abortRebase aborts the rebase
func (m *RebaseModel) abortRebase() tea.Cmd {
	return func() tea.Msg {
		err := m.GitSvc.AbortRebase()
		if err != nil {
			return err
		}
		m.ClearConflict()
		return RebaseAbortedMsg{}
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *RebaseModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
	if mode == ModeConfigure {
		m.ConflictHandler.Reset()
	}
}

// GetMode returns the current operating mode
func (m *RebaseModel) GetMode() CommandMode {
	return m.BaseModel.GetMode()
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *RebaseModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		params["target"] = m.branches[m.cursor]
	}
	params["interactive"] = "false"
	if m.interactive {
		params["interactive"] = "true"
	}
	params["autoStash"] = "false"
	if m.autoStash {
		params["autoStash"] = "true"
	}
	return params
}

// Execute returns a command to execute the operation (only valid in execute mode)
func (m *RebaseModel) SetParameters(params map[string]string) {
	if params == nil {
		return
	}
	m.targetBranch = params["target"]
	if m.targetBranch == "" {
		m.targetBranch = params["targetBranch"]
	}
	m.interactive = params["interactive"] == "true"
	m.autoStash = params["autoStash"] == "true"
	for i, branch := range m.branches {
		if branch == m.targetBranch {
			m.cursor = i
			break
		}
	}
}

func (m *RebaseModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		return m.rebase(m.branches[m.cursor])
	}
	return nil
}

// GetStepType returns the step type this command model represents
func (m *RebaseModel) GetStepType() models.StepType {
	return models.StepRebase
}
