package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// PushModel handles the push screen
type PushModel struct {
	gitSvc  git.GitService
	styles  Styles
	spinner spinner.Model
	running bool
	done    bool
	err     error
	output  []string
}

// NewPushModel creates a new push model
func NewPushModel(gitSvc git.GitService, styles Styles) *PushModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Success.GetForeground())

	return &PushModel{
		gitSvc:  gitSvc,
		styles:  styles,
		spinner: s,
		output:  []string{},
	}
}

// Init initializes the model
func (m *PushModel) Init() tea.Cmd {
	m.running = true
	return tea.Batch(
		m.spinner.Tick,
		m.executePush(),
	)
}

// Update handles messages
func (m *PushModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only allow quit if done or error
		if msg.String() == "esc" && (m.done || m.err != nil) {
			return m, nil
		}

	case spinner.TickMsg:
		if m.running {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case pushOutputMsg:
		m.output = append(m.output, string(msg))
		if len(m.output) > 20 {
			m.output = m.output[len(m.output)-20:]
		}
		return m, nil

	case pushSuccessMsg:
		m.running = false
		m.done = true
		m.output = append(m.output, "✔ Push completed successfully!")
		return m, nil

	case pushErrorMsg:
		m.running = false
		m.err = msg
		m.output = append(m.output, fmt.Sprintf("✘ Error: %v", msg))
		return m, nil
	}

	return m, cmd
}

// View renders the push screen
func (m *PushModel) View() string {
	var lines []string

	// Title
	lines = append(lines, m.styles.Title.Render(" Push to Remote "))
	lines = append(lines, "")

	// Status
	if m.running {
		lines = append(lines, fmt.Sprintf("%s Pushing changes...", m.spinner.View()))
	} else if m.done {
		lines = append(lines, m.styles.Success.Render("✔ Push complete"))
	} else if m.err != nil {
		lines = append(lines, m.styles.Error.Render("✘ Push failed"))
	}

	lines = append(lines, "")

	// Output
	if len(m.output) > 0 {
		lines = append(lines, m.styles.Info.Render("Output:"))
		for _, line := range m.output {
			lines = append(lines, "  "+line)
		}
	}

	// Help
	if m.done || m.err != nil {
		lines = append(lines, "")
		lines = append(lines, m.styles.Help.Render("Press esc to go back"))
	}

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// executePush runs the push command
func (m *PushModel) executePush() tea.Cmd {
	return func() tea.Msg {
		// Get current branch
		branch, err := m.gitSvc.CurrentBranch()
		if err != nil {
			return pushErrorMsg(err)
		}

		err = m.gitSvc.Push("origin", branch)
		if err != nil {
			return pushErrorMsg(err)
		}

		return pushSuccessMsg{}
	}
}

type pushOutputMsg string
type pushSuccessMsg struct{}
type pushErrorMsg error
