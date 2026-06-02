package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
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
}

// NewMergeModel creates a new merge model
func NewMergeModel(gitSvc git.GitService, styles Styles) *MergeModel {
	return &MergeModel{
		gitSvc: gitSvc,
		styles: styles,
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
	lines = append(lines, m.styles.Title.Render(" Merge Branch "))
	lines = append(lines, "")

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

type mergeSuccessMsg string
type mergeContinueMsg struct{}
type mergeAbortedMsg struct{}
