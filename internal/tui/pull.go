package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// PullModel handles the pull screen
type PullModel struct {
	gitSvc  git.GitService
	styles  Styles
	spinner spinner.Model
	running bool
	done    bool
	err     error
	output  []string
}

// NewPullModel creates a new pull model
func NewPullModel(gitSvc git.GitService, styles Styles) *PullModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Info.GetForeground())

	return &PullModel{
		gitSvc:  gitSvc,
		styles:  styles,
		spinner: s,
		output:  []string{},
	}
}

// Init initializes the model
func (m *PullModel) Init() tea.Cmd {
	m.running = true
	return tea.Batch(
		m.spinner.Tick,
		m.executePull(),
	)
}

// Update handles messages
func (m *PullModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case pullOutputMsg:
		m.output = append(m.output, string(msg))
		if len(m.output) > 20 {
			m.output = m.output[len(m.output)-20:]
		}
		return m, nil

	case pullSuccessMsg:
		m.running = false
		m.done = true
		m.output = append(m.output, "✔ Pull completed successfully!")
		return m, nil

	case pullErrorMsg:
		m.running = false
		m.err = msg
		m.output = append(m.output, "✘ Error: "+msg.Error())
		return m, nil
	}

	return m, cmd
}

// View renders the pull screen
func (m *PullModel) View() string {
	var lines []string

	// Title
	lines = append(lines, m.styles.Title.Render(" Pull from Remote "))
	lines = append(lines, "")

	// Status
	if m.running {
		lines = append(lines, m.spinner.View()+" Pulling changes...")
	} else if m.done {
		lines = append(lines, m.styles.Success.Render("✔ Pull complete"))
	} else if m.err != nil {
		lines = append(lines, m.styles.Error.Render("✘ Pull failed"))
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

// executePull runs the pull command
func (m *PullModel) executePull() tea.Cmd {
	return func() tea.Msg {
		err := m.gitSvc.Pull()
		if err != nil {
			return pullErrorMsg(err)
		}
		return pullSuccessMsg{}
	}
}

type pullOutputMsg string
type pullSuccessMsg struct{}
type pullErrorMsg error
