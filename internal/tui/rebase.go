package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// RebaseModel handles the rebase screen
type RebaseModel struct {
	gitSvc   git.GitService
	styles   Styles
	branches []string
	cursor   int
	message  string
	err      error
	success  bool
}

// NewRebaseModel creates a new rebase model
func NewRebaseModel(gitSvc git.GitService, styles Styles) *RebaseModel {
	return &RebaseModel{
		gitSvc: gitSvc,
		styles: styles,
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
		case "r":
			return m, m.loadBranches()
		}

	case branchesLoadedMsg:
		m.branches = []string(msg)
		m.err = nil
		return m, nil

	case rebaseSuccessMsg:
		m.success = true
		m.message = fmt.Sprintf("✔ Rebased onto %s successfully!", msg)
		m.err = nil
		return m, nil

	case error:
		m.err = msg
		m.message = fmt.Sprintf("✘ Error: %v", msg)
		return m, nil
	}

	return m, nil
}

// View renders the rebase screen
func (m *RebaseModel) View() string {
	var lines []string

	// Title
	lines = append(lines, m.styles.Title.Render(" Rebase Branch "))
	lines = append(lines, "")

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
	lines = append(lines, m.styles.Help.Render("↑/↓: navigate | enter: rebase | r: refresh | esc: back"))

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
	return func() tea.Msg {
		err := m.gitSvc.Rebase(branch)
		if err != nil {
			return err
		}
		return rebaseSuccessMsg(branch)
	}
}

type rebaseSuccessMsg string
