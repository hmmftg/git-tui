package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// StatusModel handles the status screen
type StatusModel struct {
	gitSvc git.GitService
	styles Styles
	status StatusResult
	err    error
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
		gitSvc: gitSvc,
		styles: styles,
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
		switch msg.String() {
		case "r":
			return m, m.loadStatus()
		}

	case statusLoadedMsg:
		m.status = StatusResult(msg)
		m.err = nil
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View renders the status screen
func (m *StatusModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var lines []string

	// Title
	lines = append(lines, m.styles.Title.Render(" Repository Status "))
	lines = append(lines, "")

	// Branch info
	branchInfo := fmt.Sprintf("🔀 Branch: %s", m.styles.Value.Render(m.status.Branch))
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
		lines = append(lines, m.styles.Success.Render("✔ Working tree clean"))
	} else {
		lines = append(lines, m.styles.Warning.Render("⚡ Changes detected"))
	}
	lines = append(lines, "")

	// Modified files
	if len(m.status.Modified) > 0 {
		lines = append(lines, m.styles.Warning.Render(fmt.Sprintf("📝 Modified (%d):", len(m.status.Modified))))
		for _, f := range m.status.Modified {
			lines = append(lines, fmt.Sprintf("  • %s", f))
		}
		lines = append(lines, "")
	}

	// Added files
	if len(m.status.Added) > 0 {
		lines = append(lines, m.styles.Success.Render(fmt.Sprintf("✚ Staged (%d):", len(m.status.Added))))
		for _, f := range m.status.Added {
			lines = append(lines, fmt.Sprintf("  • %s", f))
		}
		lines = append(lines, "")
	}

	// Deleted files
	if len(m.status.Deleted) > 0 {
		lines = append(lines, m.styles.Error.Render(fmt.Sprintf("🗑 Deleted (%d):", len(m.status.Deleted))))
		for _, f := range m.status.Deleted {
			lines = append(lines, fmt.Sprintf("  • %s", f))
		}
		lines = append(lines, "")
	}

	// Untracked files
	if len(m.status.Untracked) > 0 {
		lines = append(lines, m.styles.Info.Render(fmt.Sprintf("❔ Untracked (%d):", len(m.status.Untracked))))
		for _, f := range m.status.Untracked {
			lines = append(lines, fmt.Sprintf("  • %s", f))
		}
		lines = append(lines, "")
	}

	// Conflicts
	if len(m.status.Conflicted) > 0 {
		lines = append(lines, m.styles.Error.Render(fmt.Sprintf("⚠ Conflicts (%d):", len(m.status.Conflicted))))
		for _, f := range m.status.Conflicted {
			lines = append(lines, fmt.Sprintf("  • %s", f))
		}
		lines = append(lines, "")
	}

	// Summary
	if !m.status.IsClean {
		lines = append(lines, m.styles.Help.Render("Press 'r' to refresh | esc to go back"))
	} else {
		lines = append(lines, m.styles.Help.Render("Press 'r' to refresh | esc to go back"))
	}

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadStatus loads the git status
func (m *StatusModel) loadStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.gitSvc.Status()
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

type statusLoadedMsg StatusResult
