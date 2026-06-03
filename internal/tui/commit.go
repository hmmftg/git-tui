package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// CommitModel handles the commit screen
type CommitModel struct {
	BaseModel
	subjectInput textinput.Model
	bodyInput    textarea.Model
	focusIndex   int // 0=subject, 1=body
	status       string
	committed    bool
	autoAdd      bool // For configure mode
	width        int
}

// NewCommitModel creates a new commit model
func NewCommitModel(gitSvc git.GitService, styles Styles) *CommitModel {
	subject := textinput.New()
	subject.Placeholder = "Enter commit subject..."
	subject.Focus()
	subject.Width = 50

	body := textarea.New()
	body.Placeholder = "Optional commit body..."
	body.SetWidth(50)
	body.SetHeight(5)
	body.Blur()

	return &CommitModel{
		BaseModel:    NewBaseModel(gitSvc, styles),
		subjectInput: subject,
		bodyInput:    body,
		focusIndex:   0,
		autoAdd:      true,
		width:        70,
	}
}

// Init initializes the model
func (m *CommitModel) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize updates input widths for the available content area.
func (m *CommitModel) SetSize(width, _ int) {
	if width <= 0 {
		return
	}
	m.width = width
	inputWidth := width - 12
	if inputWidth < 20 {
		inputWidth = 20
	}
	m.subjectInput.Width = inputWidth
	m.bodyInput.SetWidth(inputWidth)
}

// IsInputFocused reports whether one of the commit form inputs is focused.
func (m *CommitModel) IsInputFocused() bool {
	return m.subjectInput.Focused() || m.bodyInput.Focused()
}

func (m *CommitModel) setFocus(index int) tea.Cmd {
	m.focusIndex = index
	if index == 0 {
		m.bodyInput.Blur()
		return m.subjectInput.Focus()
	}
	m.subjectInput.Blur()
	m.bodyInput.Focus()
	return textarea.Blink
}

func (m *CommitModel) commitMessage() string {
	subject := strings.TrimSpace(m.subjectInput.Value())
	body := strings.TrimSpace(m.bodyInput.Value())
	if body == "" {
		return subject
	}
	if subject == "" {
		return body
	}
	return subject + "\n\n" + body
}

// Update handles messages
func (m *CommitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.Mode == ModeConfigure {
				return m, func() tea.Msg { return commandCancelledMsg{} }
			}
			return m, func() tea.Msg { return viewChangeMsg(models.ViewHome) }
		case "tab":
			if m.focusIndex == 0 {
				return m, m.setFocus(1)
			}
			return m, m.setFocus(0)
		case "shift+tab":
			if m.focusIndex == 1 {
				return m, m.setFocus(0)
			}
			return m, m.setFocus(1)
		case "enter":
			if m.focusIndex == 0 {
				return m, m.setFocus(1)
			}
		case "ctrl+s", "ctrl+enter":
			message := m.commitMessage()
			if m.Mode == ModeConfigure {
				if message != "" {
					config := models.CommandConfig{StepType: models.StepCommit, Parameters: m.GetParameters()}
					return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
				}
			} else if message != "" && !m.committed {
				return m, m.commit(message)
			}
		case "ctrl+a":
			if m.Mode == ModeExecute {
				return m, m.addAll()
			}
		case "ctrl+t":
			if m.Mode == ModeConfigure {
				m.autoAdd = !m.autoAdd
				return m, nil
			}
		case "ctrl+p":
			if m.Mode == ModeExecute {
				return m, NavigateCmd(models.ViewPush)
			}
		}

	case commitSuccessMsg:
		if m.Mode == ModeExecute {
			m.committed = true
			m.status = "✔ Changes committed successfully! Press Ctrl+P to push."
			m.subjectInput.SetValue("")
			m.bodyInput.SetValue("")
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

	if m.focusIndex == 0 {
		m.subjectInput, cmd = m.subjectInput.Update(msg)
	} else {
		m.bodyInput, cmd = m.bodyInput.Update(msg)
	}
	return m, cmd
}

// View renders the commit screen
func (m *CommitModel) View() string {
	var lines []string

	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Title.Render(" Configure Commit Step "))
	} else {
		lines = append(lines, m.Styles.Title.Render(" Create Commit "))
	}
	lines = append(lines, "")

	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Help.Render("Configure commit step parameters:"))
		lines = append(lines, " • Enter commit subject and optional body")
		lines = append(lines, " • Press ctrl+t to toggle auto-add option")
		lines = append(lines, " • Press ctrl+s or ctrl+enter to save configuration")
		lines = append(lines, "")
	} else if !m.committed {
		lines = append(lines, m.Styles.Help.Render("Instructions:"))
		lines = append(lines, " • Type your commit subject and optional body")
		lines = append(lines, " • Press ctrl+a to stage all changes")
		lines = append(lines, " • Press ctrl+s or ctrl+enter to commit")
		lines = append(lines, "")
	}

	lines = append(lines, m.Styles.Info.Render("Commit subject:"))
	lines = append(lines, m.Styles.Input.Render(m.subjectInput.View()))
	if subjectLen := len(m.subjectInput.Value()); subjectLen > 72 {
		lines = append(lines, m.Styles.Warning.Render(fmt.Sprintf("Subject is %d characters; 72 or fewer is recommended", subjectLen)))
	} else if subjectLen > 50 {
		lines = append(lines, m.Styles.Help.Render(fmt.Sprintf("Subject is %d characters; 50 or fewer is ideal", subjectLen)))
	}
	lines = append(lines, "")

	lines = append(lines, m.Styles.Info.Render("Commit body:"))
	lines = append(lines, m.Styles.Input.Render(m.bodyInput.View()))
	lines = append(lines, "")

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

	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Help.Render("tab: switch field | ctrl+t: toggle auto-add | ctrl+s/ctrl+enter: save | esc: cancel"))
	} else {
		lines = append(lines, m.Styles.Help.Render("tab: switch field | ctrl+a: add all | ctrl+s/ctrl+enter: commit | ctrl+p: push | esc: back"))
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
func (m *CommitModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
	if mode == ModeConfigure {
		m.subjectInput.Placeholder = "Enter commit subject..."
		m.subjectInput.SetValue("")
		m.bodyInput.SetValue("")
		m.committed = false
		m.status = ""
		m.setFocus(0)
	}
}

func (m *CommitModel) GetMode() CommandMode { return m.BaseModel.GetMode() }

func (m *CommitModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}
	params := make(map[string]string)
	params["message"] = m.commitMessage()
	params["autoAdd"] = "false"
	if m.autoAdd {
		params["autoAdd"] = "true"
	}
	return params
}

func (m *CommitModel) SetParameters(params map[string]string) {
	if params == nil {
		return
	}
	message := params["message"]
	if parts := strings.SplitN(message, "\n\n", 2); len(parts) == 2 {
		m.subjectInput.SetValue(parts[0])
		m.bodyInput.SetValue(parts[1])
	} else {
		m.subjectInput.SetValue(message)
		m.bodyInput.SetValue("")
	}
	m.autoAdd = params["autoAdd"] == "true"
}

func (m *CommitModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	return m.commit(m.commitMessage())
}

func (m *CommitModel) GetStepType() models.StepType { return models.StepCommit }

type commitSuccessMsg struct{}
type commitErrorMsg error
type addSuccessMsg struct{}
