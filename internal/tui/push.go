package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// PushModel handles the push screen
type PushModel struct {
	BaseModel
	spinner spinner.Model
	running bool
	done    bool
	output  []string
	remote  string
	force   bool
	branch  string
}

// NewPushModel creates a new push model
func NewPushModel(gitSvc git.GitService, styles Styles) *PushModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Success.GetForeground())

	return &PushModel{
		BaseModel: NewBaseModel(gitSvc, styles),
		spinner:   s,
		output:    []string{},
		remote:    "origin",
		force:     false,
		branch:    "",
	}
}

// Init initializes the model
func (m *PushModel) Init() tea.Cmd {
	if m.Mode == ModeExecute {
		m.running = true
		return tea.Batch(
			m.spinner.Tick,
			m.executePush(),
		)
	}
	return nil
}

// Update handles messages
func (m *PushModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Mode == ModeConfigure {
			switch msg.String() {
			case "r":
				// Change remote (simplified - could be enhanced with remote selection)
				if m.remote == "origin" {
					m.remote = "upstream"
				} else {
					m.remote = "origin"
				}
			case "f":
				// Toggle force push
				m.force = !m.force
			case "enter":
				// Save configuration
				config := models.CommandConfig{
					StepType:   models.StepPush,
					Parameters: m.GetParameters(),
				}
				return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
			case "esc":
				return m, func() tea.Msg { return commandCancelledMsg{} }
			}
		} else {
			// Execute mode - only allow quit if done or error
			if msg.String() == "esc" && (m.done || m.Err != nil) {
				return m, nil
			}
		}

	case spinner.TickMsg:
		if m.IsExecuteMode() && m.running {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case pushOutputMsg:
		if m.Mode == ModeExecute {
			m.output = append(m.output, string(msg))
			if len(m.output) > 20 {
				m.output = m.output[len(m.output)-20:]
			}
		}
		return m, nil

	case pushSuccessMsg:
		if m.Mode == ModeExecute {
			m.running = false
			m.done = true
			m.output = append(m.output, "✔ Push completed successfully!")
		}
		return m, nil

	case pushErrorMsg:
		if m.Mode == ModeExecute {
			m.running = false
			m.Err = msg
			m.output = append(m.output, fmt.Sprintf("✘ Error: %v", msg))
		}
		return m, nil
	}

	return m, cmd
}

// View renders the push screen
func (m *PushModel) View() string {
	var lines []string

	// Title
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Title.Render(" Configure Push Step "))
	} else {
		lines = append(lines, m.Styles.Title.Render(" Push to Remote "))
	}
	lines = append(lines, "")

	if m.Mode == ModeConfigure {
		// Configuration options
		lines = append(lines, m.Styles.Help.Render("Configure push step options:"))
		lines = append(lines, "")

		// Remote option
		remoteStyle := m.Styles.Info
		lines = append(lines, remoteStyle.Render(fmt.Sprintf("Remote: %s", m.remote)))

		// Force option
		forceStyle := m.Styles.Info
		if m.force {
			forceStyle = m.Styles.Success
		}
		forceText := "Force push: "
		if m.force {
			forceText += "✓ Yes"
		} else {
			forceText += "✗ No"
		}
		lines = append(lines, forceStyle.Render(forceText))
		lines = append(lines, "")

		// Help
		lines = append(lines, m.Styles.Help.Render("r: toggle remote | f: toggle force | enter: save | esc: cancel"))
	} else {
		// Execute mode
		// Status
		if m.running {
			lines = append(lines, fmt.Sprintf("%s Pushing changes...", m.spinner.View()))
		} else if m.done {
			lines = append(lines, m.Styles.Success.Render("✔ Push complete"))
		} else if m.Err != nil {
			lines = append(lines, m.Styles.Error.Render("✘ Push failed"))
		}

		lines = append(lines, "")

		// Output
		if len(m.output) > 0 {
			lines = append(lines, m.Styles.Info.Render("Output:"))
			for _, line := range m.output {
				lines = append(lines, "  "+line)
			}
		}

		// Help
		if m.done || m.Err != nil {
			lines = append(lines, "")
			lines = append(lines, m.Styles.Help.Render("Press esc to go back"))
		}
	}

	return m.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// executePush runs the push command
func (m *PushModel) executePush() tea.Cmd {
	return func() tea.Msg {
		// Get current branch
		branch, err := m.GitSvc.CurrentBranch()
		if err != nil {
			return pushErrorMsg(err)
		}

		err = m.GitSvc.Push("origin", branch)
		if err != nil {
			return pushErrorMsg(err)
		}

		return pushSuccessMsg{}
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *PushModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
	if mode == ModeConfigure {
		m.running = false
		m.done = false
		m.output = []string{}
	}
}

// GetMode returns the current operating mode
func (m *PushModel) GetMode() CommandMode {
	return m.BaseModel.GetMode()
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *PushModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	params["remote"] = m.remote
	params["force"] = "false"
	if m.force {
		params["force"] = "true"
	}
	if m.branch != "" {
		params["branch"] = m.branch
	}
	return params
}

// Execute returns a command to execute the operation (only valid in execute mode)
func (m *PushModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	return m.executePush()
}

// GetStepType returns the step type this command model represents
func (m *PushModel) GetStepType() models.StepType {
	return models.StepPush
}

type pushOutputMsg string
type pushSuccessMsg struct{}
type pushErrorMsg error
