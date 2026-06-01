package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// MergeModel handles the merge screen
type MergeModel struct {
	gitSvc   git.GitService
	styles   Styles
	branches []string
	cursor   int
	message  string
	err      error
	success  bool
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
		case "r":
			return m, m.loadBranches()
		}

	case branchesLoadedMsg:
		m.branches = []string(msg)
		m.err = nil
		return m, nil

	case mergeSuccessMsg:
		m.success = true
		m.message = fmt.Sprintf("✔ Merged %s successfully!", msg)
		m.err = nil
		return m, nil

	case error:
		m.err = msg
		m.message = fmt.Sprintf("✘ Error: %v", msg)
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

	// Message
	if m.message != "" {
		if m.err != nil {
			lines = append(lines, m.styles.Error.Render(m.message))
		} else {
			lines = append(lines, m.styles.Success.Render(m.message))
		}
		lines = append(lines, "")
	}

	// Help
	lines = append(lines, m.styles.Help.Render("↑/↓: navigate | enter: merge | r: refresh | esc: back"))

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
	return func() tea.Msg {
		err := m.gitSvc.Merge(branch)
		if err != nil {
			return err
		}
		return mergeSuccessMsg(branch)
	}
}

type mergeSuccessMsg string
