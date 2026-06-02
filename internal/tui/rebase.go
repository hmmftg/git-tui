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
	gitSvc       git.GitService
	styles       Styles
	branches     []string
	cursor       int
	message      string
	err          error
	success      bool
	conflict     bool
	targetBranch string
	mode         CommandMode

	// Configuration options
	interactive bool
	autoStash   bool
}

// NewRebaseModel creates a new rebase model
func NewRebaseModel(gitSvc git.GitService, styles Styles) *RebaseModel {
	return &RebaseModel{
		gitSvc:      gitSvc,
		styles:      styles,
		mode:        ModeExecute, // Default to execute mode for backward compatibility
		interactive: false,
		autoStash:   false,
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
		if m.mode == ModeConfigure {
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
				if m.conflict {
					return m, m.continueRebase()
				}
			case "a":
				// Abort rebase
				if m.conflict {
					return m, m.abortRebase()
				}
			case "r":
				// Open conflict resolver
				if m.conflict {
					return m, func() tea.Msg { return openConflictResolverMsg{} }
				}
			}
		}

	case branchesLoadedMsg:
		m.branches = []string(msg)
		m.err = nil
		return m, nil

	case rebaseSuccessMsg:
		// Check for conflicts after rebase
		if m.gitSvc.HasConflicts() {
			m.conflict = true
			m.message = fmt.Sprintf("⚠ Rebase onto %s has conflicts! Resolve them and press 'c' to continue.", msg)
			return m, nil
		}
		m.success = true
		m.message = fmt.Sprintf("✔ Rebased onto %s successfully!", msg)
		m.err = nil
		m.conflict = false
		return m, nil

	case rebaseContinueMsg:
		// Check if conflicts are resolved
		if m.gitSvc.HasConflicts() {
			m.message = "⚠ Still have conflicts! Resolve them before continuing."
			return m, nil
		}
		m.success = true
		m.message = "✔ Rebase completed successfully!"
		m.conflict = false
		return m, nil

	case rebaseAbortedMsg:
		m.conflict = false
		m.message = "✔ Rebase aborted"
		m.err = nil
		return m, nil

	case error:
		m.err = msg
		m.message = fmt.Sprintf("✘ Error: %v", msg)
		// Check if it's a conflict error
		if m.gitSvc.HasConflicts() {
			m.conflict = true
			m.message = fmt.Sprintf("⚠ Rebase conflict! Resolve files and press 'c' to continue.\nError: %v", msg)
		}
		return m, nil
	}

	return m, nil
}

// View renders the rebase screen
func (m *RebaseModel) View() string {
	var lines []string

	// Title
	if m.mode == ModeConfigure {
		lines = append(lines, m.styles.Title.Render(" Configure Rebase Step "))
	} else {
		lines = append(lines, m.styles.Title.Render(" Rebase Branch "))
	}
	lines = append(lines, "")

	if m.mode == ModeConfigure {
		// Configuration options
		lines = append(lines, m.styles.Help.Render("Configure rebase step options:"))
		lines = append(lines, "")

		// Branch selection
		lines = append(lines, m.styles.Help.Render("Select base branch to rebase ONTO:"))
		lines = append(lines, "")

		// Branch list
		if len(m.branches) == 0 {
			lines = append(lines, m.styles.Info.Render("Loading branches..."))
		} else {
			current, _ := m.gitSvc.CurrentBranch()
			for i, branch := range m.branches {
				cursor := "  "
				if m.cursor == i {
					cursor = m.styles.Key.Render("▸ ")
				}

				// Don't show current branch
				if branch == current {
					lines = append(lines, fmt.Sprintf("%s%s (current)", cursor, m.styles.Dimmed.Render(branch)))
				} else {
					lines = append(lines, fmt.Sprintf("%s%s", cursor, branch))
				}
			}
		}
		lines = append(lines, "")

		// Interactive option
		interactiveStyle := m.styles.Info
		if m.interactive {
			interactiveStyle = m.styles.Success
		}
		interactiveText := "Interactive rebase: "
		if m.interactive {
			interactiveText += "✓ Yes"
		} else {
			interactiveText += "✗ No"
		}
		lines = append(lines, interactiveStyle.Render(interactiveText))

		// Auto-stash option
		autoStashStyle := m.styles.Info
		if m.autoStash {
			autoStashStyle = m.styles.Success
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
		lines = append(lines, m.styles.Help.Render("↑/↓: navigate | i: toggle interactive | s: toggle auto-stash | enter: save | esc: cancel"))
	} else {
		// Execute mode
		// Description
		current, _ := m.gitSvc.CurrentBranch()
		lines = append(lines, m.styles.Info.Render(fmt.Sprintf("Current branch: %s", current)))
		lines = append(lines, "")
		lines = append(lines, m.styles.Help.Render("Select base branch to rebase ONTO:"))
		lines = append(lines, "")

		// Branch list
		if len(m.branches) == 0 {
			lines = append(lines, m.styles.Info.Render("Loading branches..."))
		} else {
			for i, branch := range m.branches {
				cursor := "  "
				if m.cursor == i {
					cursor = m.styles.Key.Render("▸ ")
				}

				// Don't show current branch
				if branch == current {
					lines = append(lines, fmt.Sprintf("%s%s (current)", cursor, m.styles.Dimmed.Render(branch)))
				} else {
					lines = append(lines, fmt.Sprintf("%s%s", cursor, branch))
				}
			}
		}

		lines = append(lines, "")
	}

	// Conflict warning banner
	if m.conflict {
		banner := m.styles.Warning.Render(" ⚠ REBASE CONFLICTS DETECTED ")
		lines = append(lines, banner)
		lines = append(lines, m.styles.Warning.Render("Please resolve conflicts in your editor, then:"))
		lines = append(lines, "")
	}

	// Message
	if m.message != "" {
		if m.conflict {
			lines = append(lines, m.styles.Warning.Render(m.message))
		} else if m.err != nil {
			lines = append(lines, m.styles.Error.Render(m.message))
		} else {
			lines = append(lines, m.styles.Success.Render(m.message))
		}
		lines = append(lines, "")
	}

	// Help
	help := "↑/↓: navigate | enter: rebase | R: refresh | esc: back"
	if m.conflict {
		help = "r: resolve | c: continue | a: abort | esc: back"
	}
	lines = append(lines, m.styles.Help.Render(help))

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadBranches loads all branches
func (m *RebaseModel) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.gitSvc.GetBranches()
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
		err := m.gitSvc.Rebase(branch)
		if err != nil {
			return err
		}
		return rebaseSuccessMsg(branch)
	}
}

// continueRebase continues the rebase after conflicts are resolved
func (m *RebaseModel) continueRebase() tea.Cmd {
	return func() tea.Msg {
		// Check if still has conflicts
		if m.gitSvc.HasConflicts() {
			return fmt.Errorf("still has unresolved conflicts")
		}
		// Continue the rebase
		err := m.gitSvc.ContinueRebase()
		if err != nil {
			return err
		}
		return rebaseContinueMsg{}
	}
}

// abortRebase aborts the rebase
func (m *RebaseModel) abortRebase() tea.Cmd {
	return func() tea.Msg {
		err := m.gitSvc.AbortRebase()
		if err != nil {
			return err
		}
		m.conflict = false
		return rebaseAbortedMsg{}
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *RebaseModel) SetMode(mode CommandMode) {
	m.mode = mode
	if mode == ModeConfigure {
		m.message = ""
		m.err = nil
		m.conflict = false
	}
}

// GetMode returns the current operating mode
func (m *RebaseModel) GetMode() CommandMode {
	return m.mode
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *RebaseModel) GetParameters() map[string]string {
	if m.mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		params["targetBranch"] = m.branches[m.cursor]
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
func (m *RebaseModel) Execute() tea.Cmd {
	if m.mode != ModeExecute {
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

type rebaseSuccessMsg string
type rebaseContinueMsg struct{}
type rebaseAbortedMsg struct{}
