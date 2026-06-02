package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// CommitModel handles the commit screen
type CommitModel struct {
	BaseModel
	input     textinput.Model
	status    string
	committed bool
	autoAdd   bool // For configure mode
}

// NewCommitModel creates a new commit model
func NewCommitModel(gitSvc git.GitService, styles Styles) *CommitModel {
	ti := textinput.New()
	ti.Placeholder = "Enter commit message..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return &CommitModel{
		BaseModel: NewBaseModel(gitSvc, styles),
		input:     ti,
		autoAdd:   true,
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
			if m.Mode == ModeConfigure {
				// In configure mode, save and return configuration
				if m.input.Value() != "" {
					config := models.CommandConfig{
						StepType:   models.StepCommit,
						Parameters: m.GetParameters(),
					}
					return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
				}
			} else {
				// In execute mode, commit the changes
				if m.input.Value() != "" && !m.committed {
					return m, m.commit(m.input.Value())
				}
			}
		case "ctrl+a":
			if m.Mode == ModeExecute {
				return m, m.addAll()
			}
		case "a":
			if m.Mode == ModeConfigure {
				// Toggle auto-add option
				m.autoAdd = !m.autoAdd
			}
		case "esc":
			if m.Mode == ModeConfigure {
				return m, func() tea.Msg { return commandCancelledMsg{} }
			}
		}

	case commitSuccessMsg:
		if m.Mode == ModeExecute {
			m.committed = true
			m.status = "✔ Changes committed successfully!"
			m.input.SetValue("")
		}

	case commitErrorMsg:
		if m.Mode == ModeExecute {
			m.Err = msg
			m.status = fmt.Sprintf("✘ Error: %v", msg)
		}

	case addSuccessMsg:
		if m.Mode == ModeExecute {
			m.status = "✔ All changes staged"
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the commit screen
func (m *CommitModel) View() string {
	var lines []string

	// Title
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Title.Render(" Configure Commit Step "))
	} else {
		lines = append(lines, m.Styles.Title.Render(" Create Commit "))
	}
	lines = append(lines, "")

	// Instructions
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Help.Render("Configure commit step parameters:"))
		lines = append(lines, " • Enter commit message")
		lines = append(lines, " • Press 'a' to toggle auto-add option")
		lines = append(lines, " • Press enter to save configuration")
		lines = append(lines, "")
	} else if !m.committed {
		lines = append(lines, m.Styles.Help.Render("Instructions:"))
		lines = append(lines, " • Type your commit message")
		lines = append(lines, " • Press ctrl+a to stage all changes")
		lines = append(lines, " • Press enter to commit")
		lines = append(lines, "")
	}

	// Input field
	lines = append(lines, m.Styles.Info.Render("Commit message:"))
	lines = append(lines, m.Styles.Input.Render(m.input.View()))
	lines = append(lines, "")

	// Auto-add option (only in configure mode)
	if m.Mode == ModeConfigure {
		autoAddStyle := m.Styles.Info
		if m.autoAdd {
			autoAddStyle = m.Styles.Success
		}
		autoAddText := "Auto-add all changes: "
		if m.autoAdd {
			autoAddText += "✓ Yes"
		} else {
			autoAddText += "✗ No"
		}
		lines = append(lines, autoAddStyle.Render(autoAddText))
		lines = append(lines, "")
	}

	// Status message
	if m.status != "" {
		if m.Err != nil {
			lines = append(lines, m.Styles.Error.Render(m.status))
		} else if strings.Contains(m.status, "✔") {
			lines = append(lines, m.Styles.Success.Render(m.status))
		} else {
			lines = append(lines, m.Styles.Info.Render(m.status))
		}
		lines = append(lines, "")
	}

	// Help
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Help.Render("a: toggle auto-add | enter: save | esc: cancel"))
	} else {
		lines = append(lines, m.Styles.Help.Render("ctrl+a: add all | enter: commit | esc: back"))
	}

	return m.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// commit executes the commit
func (m *CommitModel) commit(message string) tea.Cmd {
	return func() tea.Msg {
		err := m.GitSvc.Commit(message)
		if err != nil {
			return commitErrorMsg(err)
		}
		return commitSuccessMsg{}
	}
}

// addAll stages all changes
func (m *CommitModel) addAll() tea.Cmd {
	return func() tea.Msg {
		err := m.GitSvc.AddAll()
		if err != nil {
			return commitErrorMsg(err)
		}
		return addSuccessMsg{}
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *CommitModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
	if mode == ModeConfigure {
		m.input.Placeholder = "Enter commit message..."
		m.input.SetValue("")
		m.committed = false
		m.status = ""
		m.input.Focus()
	}
}

// GetMode returns the current operating mode
func (m *CommitModel) GetMode() CommandMode {
	return m.BaseModel.GetMode()
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *CommitModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	params["message"] = m.input.Value()
	params["autoAdd"] = "false"
	if m.autoAdd {
		params["autoAdd"] = "true"
	}
	return params
}

// Execute returns a command to execute the operation (only valid in execute mode)
func (m *CommitModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	return m.commit(m.input.Value())
}

// GetStepType returns the step type this command model represents
func (m *CommitModel) GetStepType() models.StepType {
	return models.StepCommit
}

type commitSuccessMsg struct{}
type commitErrorMsg error
type addSuccessMsg struct{}
