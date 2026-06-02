package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// CheckoutModel handles the checkout screen
type CheckoutModel struct {
	BaseModel
	branches []string
	cursor   int
	selected string
}

// NewCheckoutModel creates a new checkout model
func NewCheckoutModel(gitSvc git.GitService, styles Styles) *CheckoutModel {
	return &CheckoutModel{
		BaseModel: NewBaseModel(gitSvc, styles),
	}
}

// Init initializes the model
func (m *CheckoutModel) Init() tea.Cmd {
	return m.loadBranches()
}

// Update handles messages
func (m *CheckoutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.branches)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.branches) > 0 && m.cursor < len(m.branches) {
				if m.Mode == ModeConfigure {
					// In configure mode, save and return configuration
					config := models.CommandConfig{
						StepType:   models.StepCheckout,
						Parameters: m.GetParameters(),
					}
					return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
				} else {
					// In execute mode, checkout the branch
					return m, m.checkout(m.branches[m.cursor])
				}
			}
		case "r":
			if m.Mode == ModeExecute {
				return m, m.loadBranches()
			}
		case "esc":
			if m.Mode == ModeConfigure {
				return m, func() tea.Msg { return commandCancelledMsg{} }
			}
		}

	case branchesLoadedMsg:
		m.branches = []string(msg)
		m.Err = nil
		if m.selected != "" {
			for i, b := range m.branches {
				if b == m.selected {
					m.cursor = i
					return m, nil
				}
			}
		}
		// Find current position
		current, _ := m.GitSvc.CurrentBranch()
		for i, b := range m.branches {
			if b == current {
				m.cursor = i
				break
			}
		}
		return m, nil

	case checkoutSuccessMsg:
		m.selected = string(msg)
		m.Message = fmt.Sprintf("✔ Switched to %s", m.selected)
		return m, m.loadBranches()

	case error:
		m.Err = msg
		m.Message = fmt.Sprintf("✘ Error: %v", msg)
		return m, nil
	}

	return m, nil
}

// View renders the checkout screen
func (m *CheckoutModel) View() string {
	var lines []string

	// Title
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Title.Render(" Configure Checkout Step "))
	} else {
		lines = append(lines, m.Styles.Title.Render(" Checkout Branch "))
	}
	lines = append(lines, "")

	// Instructions
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Help.Render("Select branch for checkout step:"))
		lines = append(lines, " • Use ↑/↓ to navigate")
		lines = append(lines, " • Press enter to save configuration")
		lines = append(lines, "")
	}

	// Branch list
	if len(m.branches) == 0 {
		lines = append(lines, m.Styles.Info.Render("Loading branches..."))
	} else {
		current, _ := m.GitSvc.CurrentBranch()

		for i, branch := range m.branches {
			cursor := "  "
			if m.cursor == i {
				cursor = m.Styles.Key.Render("▸ ")
			}

			branchDisplay := branch
			if branch == current {
				branchDisplay = m.Styles.Success.Render(fmt.Sprintf("%s (current)", branch))
			}

			lines = append(lines, fmt.Sprintf("%s%s", cursor, branchDisplay))
		}
	}

	lines = append(lines, "")

	// Message
	if m.Message != "" {
		if m.Err != nil {
			lines = append(lines, m.Styles.Error.Render(m.Message))
		} else {
			lines = append(lines, m.Styles.Success.Render(m.Message))
		}
		lines = append(lines, "")
	}

	// Help
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Help.Render("↑/↓: navigate | enter: save | esc: cancel"))
	} else {
		lines = append(lines, m.Styles.Help.Render("↑/↓: navigate | enter: checkout | r: refresh | esc: back"))
	}

	return m.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadBranches loads all branches
func (m *CheckoutModel) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.GitSvc.GetBranches()
		if err != nil {
			return err
		}
		return branchesLoadedMsg(branches)
	}
}

// checkout switches to the selected branch
func (m *CheckoutModel) checkout(branch string) tea.Cmd {
	return func() tea.Msg {
		err := m.GitSvc.Checkout(branch)
		if err != nil {
			return err
		}
		return checkoutSuccessMsg(branch)
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *CheckoutModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
}

// GetMode returns the current operating mode
func (m *CheckoutModel) GetMode() CommandMode {
	return m.BaseModel.GetMode()
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *CheckoutModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		params["branch"] = m.branches[m.cursor]
	}
	return params
}

// Execute returns a command to execute the operation (only valid in execute mode)
func (m *CheckoutModel) SetParameters(params map[string]string) {
	if params == nil {
		return
	}
	m.selected = params["branch"]
	for i, branch := range m.branches {
		if branch == m.selected {
			m.cursor = i
			break
		}
	}
}

func (m *CheckoutModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		return m.checkout(m.branches[m.cursor])
	}
	return nil
}

// GetStepType returns the step type this command model represents
func (m *CheckoutModel) GetStepType() models.StepType {
	return models.StepCheckout
}

type branchesLoadedMsg []string
type checkoutSuccessMsg string
