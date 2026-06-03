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
	BaseModel
	*ConflictHandler
	branches     []string
	cursor       int
	success      bool
	targetBranch string
	running      bool

	// Configuration options
	noCommit bool
	squash   bool
}

// NewMergeModel creates a new merge model
func NewMergeModel(gitSvc git.GitService, styles Styles) *MergeModel {
	return &MergeModel{
		BaseModel:       NewBaseModel(gitSvc, styles),
		ConflictHandler: NewConflictHandler("merge"),
		noCommit:        false,
		squash:          false,
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
				if m.HasConflict() {
					return m, m.continueMerge()
				}
			case "a":
				// Abort merge
				if m.HasConflict() {
					return m, m.abortMerge()
				}
			case "r":
				// Open conflict resolver
				if m.HasConflict() {
					return m, func() tea.Msg { return OpenConflictResolverMsg{} }
				}
			case "p":
				if !m.running && !m.HasConflict() {
					return m, NavigateCmd(models.ViewPush)
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

	case MergeSuccessMsg:
		m.running = false
		// Check for conflicts after merge
		if m.GitSvc.HasConflicts() {
			m.SetConflict(string(msg))
			m.Message = m.GetConflictMessage(string(msg))
			return m, nil
		}
		m.success = true
		m.SetSuccess("Merged %s successfully!", msg)
		m.ClearConflict()
		return m, nil

	case MergeContinueMsg:
		m.running = false
		// Check if conflicts are resolved
		if m.GitSvc.HasConflicts() {
			m.Message = m.GetStillConflictMessage()
			return m, nil
		}
		m.success = true
		m.SetSuccess("Merge completed successfully!")
		m.ClearConflict()
		return m, nil

	case MergeAbortedMsg:
		m.running = false
		m.Reset()
		m.SetSuccess("Merge aborted")
		return m, nil

	case error:
		m.running = false
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

// View renders the merge screen
func (m *MergeModel) View() string {
	var lines []string

	// Title
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Title.Render(" Configure Merge Step "))
	} else {
		lines = append(lines, m.Styles.Title.Render(" Merge Branch "))
	}
	lines = append(lines, "")

	if m.Mode == ModeConfigure {
		// Configuration options
		lines = append(lines, m.Styles.Help.Render("Configure merge step options:"))
		lines = append(lines, "")

		// Branch selection
		lines = append(lines, m.Styles.Help.Render("Select target branch to merge:"))
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

		// No-commit option
		noCommitStyle := m.Styles.Info
		if m.noCommit {
			noCommitStyle = m.Styles.Success
		}
		noCommitText := "No commit: "
		if m.noCommit {
			noCommitText += "✓ Yes"
		} else {
			noCommitText += "✗ No"
		}
		lines = append(lines, noCommitStyle.Render(noCommitText))

		// Squash option
		squashStyle := m.Styles.Info
		if m.squash {
			squashStyle = m.Styles.Success
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
		lines = append(lines, m.Styles.Help.Render("↑/↓: navigate | n: toggle no-commit | s: toggle squash | enter: save | esc: cancel"))
	} else {
		// Execute mode
		// Description
		current, _ := m.GitSvc.CurrentBranch()
		lines = append(lines, m.Styles.Info.Render(fmt.Sprintf("Current branch: %s", current)))
		lines = append(lines, "")

		// Common targets hint
		lines = append(lines, m.Styles.Help.Render("Select branch to merge INTO current branch:"))
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
	var help string
	if m.success && !m.HasConflict() {
		help = "esc: back | p: push"
	} else {
		help = "↑/↓: navigate | enter: merge | R: refresh | esc: back"
	}
	help = m.ConflictHandler.GetHelpText(help)
	lines = append(lines, m.Styles.Help.Render(help))

	return m.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadBranches loads all branches
func (m *MergeModel) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.GitSvc.GetBranches()
		if err != nil {
			return err
		}
		return branchesLoadedMsg(branches)
	}
}

// merge merges the selected branch
func (m *MergeModel) merge(branch string) tea.Cmd {
	m.targetBranch = branch
	m.running = true
	return func() tea.Msg {
		err := m.GitSvc.Merge(branch)
		if err != nil {
			return err
		}
		return MergeSuccessMsg(branch)
	}
}

// continueMerge continues the merge after conflicts are resolved
func (m *MergeModel) continueMerge() tea.Cmd {
	return func() tea.Msg {
		// First check if still has conflicts
		if m.GitSvc.HasConflicts() {
			return fmt.Errorf("still has unresolved conflicts")
		}
		// Try to commit the merge resolution
		err := m.GitSvc.Commit("Merge conflict resolution")
		if err != nil {
			// If no changes to commit, that's OK - merge was already done
			return MergeContinueMsg{}
		}
		return MergeContinueMsg{}
	}
}

// abortMerge aborts the merge
func (m *MergeModel) abortMerge() tea.Cmd {
	return func() tea.Msg {
		err := m.GitSvc.AbortMerge()
		if err != nil {
			return err
		}
		m.Reset()
		return MergeAbortedMsg{}
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *MergeModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
	if mode == ModeConfigure {
		m.ConflictHandler.Reset()
		m.running = false
	}
}

// GetMode returns the current operating mode
func (m *MergeModel) GetMode() CommandMode {
	return m.BaseModel.GetMode()
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *MergeModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		params["target"] = m.branches[m.cursor]
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
func (m *MergeModel) SetParameters(params map[string]string) {
	if params == nil {
		return
	}
	m.targetBranch = params["target"]
	if m.targetBranch == "" {
		m.targetBranch = params["targetBranch"]
	}
	m.noCommit = params["noCommit"] == "true"
	m.squash = params["squash"] == "true"
	for i, branch := range m.branches {
		if branch == m.targetBranch {
			m.cursor = i
			break
		}
	}
}

func (m *MergeModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
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
