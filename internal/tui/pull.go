package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// PullModel handles the pull screen
type PullModel struct {
	BaseModel
	spinner spinner.Model
	running bool
	done    bool
	output  []string

	// Configuration options
	remote string
	rebase bool
	branch string
}

// NewPullModel creates a new pull model
func NewPullModel(gitSvc git.GitService, styles Styles) *PullModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Info.GetForeground())

	return &PullModel{
		BaseModel: NewBaseModel(gitSvc, styles),
		spinner:   s,
		output:    []string{},
		remote:    "origin",
		rebase:    false,
		branch:    "",
	}
}

// Init initializes the model
func (m *PullModel) Init() tea.Cmd {
	if m.Mode == ModeExecute {
		m.running = true
		return tea.Batch(
			m.spinner.Tick,
			m.executePull(),
		)
	}
	// In configure mode, no initialization needed
	return nil
}

// Update handles messages
func (m *PullModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			case "b":
				// Toggle rebase option
				m.rebase = !m.rebase
			case "enter":
				// Save configuration
				config := models.CommandConfig{
					StepType:   models.StepPull,
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
		if m.Mode == ModeExecute && m.running {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case pullOutputMsg:
		if m.Mode == ModeExecute {
			m.output = append(m.output, string(msg))
			if len(m.output) > 20 {
				m.output = m.output[len(m.output)-20:]
			}
		}
		return m, nil

	case pullSuccessMsg:
		if m.Mode == ModeExecute {
			m.running = false
			m.done = true
			m.output = append(m.output, "✔ Pull completed successfully!")
		}
		return m, nil

	case pullErrorMsg:
		if m.Mode == ModeExecute {
			m.running = false
			m.Err = msg
			m.output = append(m.output, "✘ Error: "+msg.Error())
		}
		return m, nil
	}

	return m, cmd
}

// View renders the pull screen
func (m *PullModel) View() string {
	var lines []string

	// Title
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Title.Render(" Configure Pull Step "))
	} else {
		lines = append(lines, m.Styles.Title.Render(" Pull from Remote "))
	}
	lines = append(lines, "")

	if m.Mode == ModeConfigure {
		// Configuration options
		lines = append(lines, m.Styles.Help.Render("Configure pull step options:"))
		lines = append(lines, "")

		// Remote option
		remoteStyle := m.Styles.Info
		lines = append(lines, remoteStyle.Render(fmt.Sprintf("Remote: %s", m.remote)))

		// Rebase option
		rebaseStyle := m.Styles.Info
		if m.rebase {
			rebaseStyle = m.Styles.Success
		}
		rebaseText := "Use rebase: "
		if m.rebase {
			rebaseText += "✓ Yes"
		} else {
			rebaseText += "✗ No"
		}
		lines = append(lines, rebaseStyle.Render(rebaseText))
		lines = append(lines, "")

		// Help
		lines = append(lines, m.Styles.Help.Render("r: toggle remote | b: toggle rebase | enter: save | esc: cancel"))
	} else {
		// Execute mode
		// Status
		if m.running {
			lines = append(lines, m.spinner.View()+" Pulling changes...")
		} else if m.done {
			lines = append(lines, m.Styles.Success.Render("✔ Pull complete"))
		} else if m.Err != nil {
			lines = append(lines, m.Styles.Error.Render("✘ Pull failed"))
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

// executePull runs the pull command
func (m *PullModel) executePull() tea.Cmd {
	return func() tea.Msg {
		err := m.GitSvc.PullOptions(m.remote, m.rebase, m.branch)
		if err != nil {
			return pullErrorMsg(err)
		}
		return pullSuccessMsg{}
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *PullModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
	if mode == ModeConfigure {
		m.running = false
		m.done = false
		m.output = []string{}
	}
}

// GetMode returns the current operating mode
func (m *PullModel) GetMode() CommandMode {
	return m.BaseModel.GetMode()
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *PullModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	params["remote"] = m.remote
	params["rebase"] = "false"
	if m.rebase {
		params["rebase"] = "true"
	}
	if m.branch != "" {
		params["branch"] = m.branch
	}
	return params
}

// Execute returns a command to execute the operation (only valid in execute mode)
func (m *PullModel) SetParameters(params map[string]string) {
	if params == nil {
		return
	}
	if remote := params["remote"]; remote != "" {
		m.remote = remote
	}
	m.rebase = params["rebase"] == "true"
	m.branch = params["branch"]
}

func (m *PullModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	return m.executePull()
}

// GetStepType returns the step type this command model represents
func (m *PullModel) GetStepType() models.StepType {
	return models.StepPull
}

type pullOutputMsg string
type pullSuccessMsg struct{}
type pullErrorMsg error
