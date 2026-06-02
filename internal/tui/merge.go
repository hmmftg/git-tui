package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// MergeModel handles the merge screen
type MergeModel struct {
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
	noCommit bool
	squash   bool
}

// NewMergeModel creates a new merge model
func NewMergeModel(gitSvc git.GitService, styles Styles) *MergeModel {
	return &MergeModel{
		gitSvc:   gitSvc,
		styles:   styles,
		mode:     ModeExecute, // Default to execute mode for backward compatibility
		noCommit: false,
		squash:   false,
	}
}

// Init initializes the model
func (m *MergeModel) Init() tea.Cmd {
	return m.loadBranches()
}

// Update handles messages
func (m *MergeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			case "n":
				// Toggle no-commit option
				m.noCommit = !m.noCommit
			case "s":
				// Toggle squash option
				m.squash = !m.squash
			case "enter":
				// Save configuration
				config := models.CommandConfig{
					StepType:   models.StepMerge,
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
					return m, m.merge(m.branches[m.cursor])
				}
			case "R":
				return m, m.loadBranches()
			case "c":
				// Continue merge after resolving conflicts
				if m.conflict {
					return m, m.continueMerge()
				}
			case "a":
				// Abort merge
				if m.conflict {
					return m, m.abortMerge()
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

	case mergeSuccessMsg:
		// Check for conflicts after merge
		if m.gitSvc.HasConflicts() {
			m.conflict = true
			m.message = fmt.Sprintf("⚠ Merge of %s has conflicts! Resolve them and press 'c' to continue.", msg)
			return m, nil
		}
		m.success = true
		m.message = fmt.Sprintf("✔ Merged %s successfully!", msg)
		m.err = nil
		m.conflict = false
		return m, nil

	case mergeContinueMsg:
		// Check if conflicts are resolved
		if m.gitSvc.HasConflicts() {
			m.message = "⚠ Still have conflicts! Resolve them before continuing."
			return m, nil
		}
		m.success = true
		m.message = "✔ Merge completed successfully!"
		m.conflict = false
		return m, nil

	case mergeAbortedMsg:
		m.conflict = false
		m.message = "✔ Merge aborted"
		m.err = nil
		return m, nil

	case error:
		m.err = msg
		m.message = fmt.Sprintf("✘ Error: %v", msg)
		// Check if it's a conflict error
		if m.gitSvc.HasConflicts() {
			m.conflict = true
			m.message = fmt.Sprintf("⚠ Merge conflict! Resolve files and press 'c' to continue.\nError: %v", msg)
		}
		return m, nil
	}

	return m, nil
}

// View renders the merge screen
func (m *MergeModel) View() string {
	var lines []string

	// Title
	if m.mode == ModeConfigure {
		lines = append(lines, m.styles.Title.Render(" Configure Merge Step "))
	} else {
		lines = append(lines, m.styles.Title.Render(" Merge Branch "))
	}
	lines = append(lines, "")

	if m.mode == ModeConfigure {
		// Configuration options
		lines = append(lines, m.styles.Help.Render("Configure merge step options:"))
		lines = append(lines, "")

		// Branch selection
		lines = append(lines, m.styles.Help.Render("Select target branch to merge:"))
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

		// No-commit option
		noCommitStyle := m.styles.Info
		if m.noCommit {
			noCommitStyle = m.styles.Success
		}
		noCommitText := "No commit: "
		if m.noCommit {
			noCommitText += "✓ Yes"
		} else {
			noCommitText += "✗ No"
		}
		lines = append(lines, noCommitStyle.Render(noCommitText))

		// Squash option
		squashStyle := m.styles.Info
		if m.squash {
			squashStyle = m.styles.Success
		}
		squashText := "Squash merge: "
		if m.squash {
			squashText += "✓ Yes"
		} else {
			squashText += "✗ No"
		}
		lines = append(lines, squashStyle.Render(squashText))
		lines = append(lines, "")

		// Help
		lines = append(lines, m.styles.Help.Render("↑/↓: navigate | n: toggle no-commit | s: toggle squash | enter: save | esc: cancel"))
	} else {
		// Execute mode
		// Description
		current, _ := m.gitSvc.CurrentBranch()
		lines = append(lines, m.styles.Info.Render(fmt.Sprintf("Current branch: %s", current)))
		lines = append(lines, "")

		// Common targets hint
		lines = append(lines, m.styles.Help.Render("Select branch to merge INTO current branch:"))
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
		banner := m.styles.Warning.Render(" ⚠ MERGE CONFLICTS DETECTED ")
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
	help := "↑/↓: navigate | enter: merge | R: refresh | esc: back"
	if m.conflict {
		help = "r: resolve | c: continue | a: abort | esc: back"
	}
	lines = append(lines, m.styles.Help.Render(help))

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadBranches loads all branches
func (m *MergeModel) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.gitSvc.GetBranches()
		if err != nil {
			return err
		}
		return branchesLoadedMsg(branches)
	}
}

// merge merges the selected branch
func (m *MergeModel) merge(branch string) tea.Cmd {
	m.targetBranch = branch
	return func() tea.Msg {
		err := m.gitSvc.Merge(branch)
		if err != nil {
			return err
		}
		return mergeSuccessMsg(branch)
	}
}

// continueMerge continues the merge after conflicts are resolved
func (m *MergeModel) continueMerge() tea.Cmd {
	return func() tea.Msg {
		// First check if still has conflicts
		if m.gitSvc.HasConflicts() {
			return fmt.Errorf("still has unresolved conflicts")
		}
		// Try to commit the merge resolution
		err := m.gitSvc.Commit("Merge conflict resolution")
		if err != nil {
			// If no changes to commit, that's OK - merge was already done
			return mergeContinueMsg{}
		}
		return mergeContinueMsg{}
	}
}

// abortMerge aborts the merge
func (m *MergeModel) abortMerge() tea.Cmd {
	return func() tea.Msg {
		err := m.gitSvc.AbortMerge()
		if err != nil {
			return err
		}
		m.conflict = false
		return mergeAbortedMsg{}
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *MergeModel) SetMode(mode CommandMode) {
	m.mode = mode
	if mode == ModeConfigure {
		m.message = ""
		m.err = nil
		m.conflict = false
	}
}

// GetMode returns the current operating mode
func (m *MergeModel) GetMode() CommandMode {
	return m.mode
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *MergeModel) GetParameters() map[string]string {
	if m.mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		params["targetBranch"] = m.branches[m.cursor]
	}
	params["noCommit"] = "false"
	if m.noCommit {
		params["noCommit"] = "true"
	}
	params["squash"] = "false"
	if m.squash {
		params["squash"] = "true"
	}
	return params
}

// Execute returns a command to execute the operation (only valid in execute mode)
func (m *MergeModel) Execute() tea.Cmd {
	if m.mode != ModeExecute {
		return nil
	}
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		return m.merge(m.branches[m.cursor])
	}
	return nil
}

// GetStepType returns the step type this command model represents
func (m *MergeModel) GetStepType() models.StepType {
	return models.StepMerge
}

type mergeSuccessMsg string
type mergeContinueMsg struct{}
type mergeAbortedMsg struct{}
