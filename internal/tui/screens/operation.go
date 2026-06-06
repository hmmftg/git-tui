package screens

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"gitflow-tui/internal/tui"
	"gitflow-tui/internal/tui/components"
)

// OperationScreen is a generic async operation execution screen.
type OperationScreen struct {
	ctx      *tui.AppContext
	title    string
	command  string
	run      func() error
	executor *components.AsyncExecutor
}

// NewOperationScreen creates a new operation screen.
func NewOperationScreen(ctx *tui.AppContext, title string, command string, run func() error) *OperationScreen {
	executor := components.NewAsyncExecutor(ctx.Styles)
	executor.SetCommand(command)
	return &OperationScreen{
		ctx:      ctx,
		title:    title,
		command:  command,
		run:      run,
		executor: executor,
	}
}

// Init starts the operation.
func (s *OperationScreen) Init() tea.Cmd {
	return tea.Batch(
		s.executor.Start(),
		func() tea.Msg {
			err := s.run()
			return tui.OperationFinishedMsg{Err: err}
		},
	)
}

// Update handles messages.
func (s *OperationScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.executor.IsDone() && msg.String() == "esc" {
			return NewHomeScreen(s.ctx), nil
		}
	case tui.OperationFinishedMsg:
		if msg.Err != nil {
			s.executor.Fail(msg.Err)
		} else {
			s.executor.Stop()
		}
		return s, nil
	}

	cmd := s.executor.Update(msg)
	return s, cmd
}

// View renders the screen.
func (s *OperationScreen) View() tea.View {
	var lines []string
	lines = append(lines, s.ctx.Styles.Title.Render(" "+s.title+" "))
	lines = append(lines, "")
	lines = append(lines, s.executor.View(s.ctx.Styles, s.title))
	if s.executor.IsDone() {
		lines = append(lines, "")
		lines = append(lines, s.ctx.Styles.Help.Render("esc: go back"))
	}
	return tea.NewView(s.ctx.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
}
