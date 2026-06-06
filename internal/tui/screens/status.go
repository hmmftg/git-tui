package screens

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"gitflow-tui/internal/tui"
	"gitflow-tui/internal/tui/forms"
)

// StatusScreen displays repository status.
type StatusScreen struct {
	ctx *tui.AppContext
}

// NewStatusScreen creates a new status screen.
func NewStatusScreen(ctx *tui.AppContext) *StatusScreen {
	return &StatusScreen{ctx: ctx}
}

// Init initializes the screen.
func (s *StatusScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s *StatusScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return NewHomeScreen(s.ctx), nil
		case "r":
			return s, nil // Refresh could trigger a status fetch
		case "c":
			return forms.NewCommitForm(s.ctx, false), nil
		case "w":
			return NewWorkflowListScreen(s.ctx), nil
		}
	}
	return s, nil
}

// View renders the status dashboard.
func (s *StatusScreen) View() tea.View {
	status, err := s.ctx.GitService.Status()
	if err != nil {
		return tea.NewView(s.ctx.Styles.Box.Render(s.ctx.Styles.Error.Render(fmt.Sprintf("Error: %v", err))))
	}

	var lines []string
	lines = append(lines, s.ctx.Styles.Title.Render(" Repository Status "))
	lines = append(lines, "")

	// Branch info
	branchInfo := fmt.Sprintf("Branch: %s", s.ctx.Styles.Key.Render(status.Branch))
	if status.Ahead > 0 || status.Behind > 0 {
		branchInfo += fmt.Sprintf(" | Ahead: %d | Behind: %d", status.Ahead, status.Behind)
	}
	lines = append(lines, branchInfo)
	lines = append(lines, "")

	// File categories
	categories := []struct {
		name  string
		files []string
		color lipgloss.Style
	}{
		{"Modified", status.Modified, s.ctx.Styles.Warning},
		{"Added", status.Added, s.ctx.Styles.Success},
		{"Deleted", status.Deleted, s.ctx.Styles.Error},
		{"Untracked", status.Untracked, s.ctx.Styles.Info},
		{"Conflicted", status.Conflicted, s.ctx.Styles.Error},
	}

	for _, cat := range categories {
		if len(cat.files) > 0 {
			lines = append(lines, cat.color.Render(fmt.Sprintf("%s (%d):", cat.name, len(cat.files))))
			for _, f := range cat.files {
				lines = append(lines, fmt.Sprintf("  %s", f))
			}
			lines = append(lines, "")
		}
	}

	if status.IsClean {
		lines = append(lines, s.ctx.Styles.Success.Render("✔ Working tree clean"))
	}

	lines = append(lines, "")
	lines = append(lines, s.ctx.Styles.Help.Render("r: refresh | c: commit | w: workflows | esc: back"))

	return tea.NewView(s.ctx.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
}
