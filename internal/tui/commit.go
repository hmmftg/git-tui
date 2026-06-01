package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// CommitModel handles the commit screen
type CommitModel struct {
	gitSvc    git.GitService
	styles    Styles
	input     textinput.Model
	status    string
	err       error
	committed bool
}

// NewCommitModel creates a new commit model
func NewCommitModel(gitSvc git.GitService, styles Styles) *CommitModel {
	ti := textinput.New()
	ti.Placeholder = "Enter commit message..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return &CommitModel{
		gitSvc: gitSvc,
		styles: styles,
		input:  ti,
	}
}

// Init initializes the model
func (m *CommitModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m *CommitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.input.Value() != "" && !m.committed {
				return m, m.commit(m.input.Value())
			}
		case "ctrl+a":
			return m, m.addAll()
		}

	case commitSuccessMsg:
		m.committed = true
		m.status = "✔ Changes committed successfully!"
		m.input.SetValue("")

	case commitErrorMsg:
		m.err = msg
		m.status = fmt.Sprintf("✘ Error: %v", msg)

	case addSuccessMsg:
		m.status = "✔ All changes staged"
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the commit screen
func (m *CommitModel) View() string {
	var lines []string

	// Title
	lines = append(lines, m.styles.Title.Render(" Create Commit "))
	lines = append(lines, "")

	// Instructions
	if !m.committed {
		lines = append(lines, m.styles.Help.Render("Instructions:"))
		lines = append(lines, " • Type your commit message")
		lines = append(lines, " • Press ctrl+a to stage all changes")
		lines = append(lines, " • Press enter to commit")
		lines = append(lines, "")
	}

	// Input field
	lines = append(lines, m.styles.Info.Render("Commit message:"))
	lines = append(lines, m.styles.Input.Render(m.input.View()))
	lines = append(lines, "")

	// Status message
	if m.status != "" {
		if m.err != nil {
			lines = append(lines, m.styles.Error.Render(m.status))
		} else if strings.Contains(m.status, "✔") {
			lines = append(lines, m.styles.Success.Render(m.status))
		} else {
			lines = append(lines, m.styles.Info.Render(m.status))
		}
		lines = append(lines, "")
	}

	// Help
	lines = append(lines, m.styles.Help.Render("ctrl+a: add all | enter: commit | esc: back"))

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// commit executes the commit
func (m *CommitModel) commit(message string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitSvc.Commit(message)
		if err != nil {
			return commitErrorMsg(err)
		}
		return commitSuccessMsg{}
	}
}

// addAll stages all changes
func (m *CommitModel) addAll() tea.Cmd {
	return func() tea.Msg {
		err := m.gitSvc.AddAll()
		if err != nil {
			return commitErrorMsg(err)
		}
		return addSuccessMsg{}
	}
}

type commitSuccessMsg struct{}
type commitErrorMsg error
type addSuccessMsg struct{}
