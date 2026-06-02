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
	gitSvc   git.GitService
	styles   Styles
	branches []string
	cursor   int
	selected string
	err      error
	message  string
	mode     CommandMode
}

// NewCheckoutModel creates a new checkout model
func NewCheckoutModel(gitSvc git.GitService, styles Styles) *CheckoutModel {
	return &CheckoutModel{
		gitSvc: gitSvc,
		styles: styles,
		mode:   ModeExecute, // Default to execute mode for backward compatibility
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
				if m.mode == ModeConfigure {
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
			if m.mode == ModeExecute {
				return m, m.loadBranches()
			}
		case "esc":
			if m.mode == ModeConfigure {
				return m, func() tea.Msg { return commandCancelledMsg{} }
			}
		}

	case branchesLoadedMsg:
		m.branches = []string(msg)
		m.err = nil
		// Find current position
		current, _ := m.gitSvc.CurrentBranch()
		for i, b := range m.branches {
			if b == current {
				m.cursor = i
				break
			}
		}
		return m, nil

	case checkoutSuccessMsg:
		m.selected = string(msg)
		m.message = fmt.Sprintf("✔ Switched to %s", m.selected)
		return m, m.loadBranches()

	case error:
		m.err = msg
		m.message = fmt.Sprintf("✘ Error: %v", msg)
		return m, nil
	}

	return m, nil
}

// View renders the checkout screen
func (m *CheckoutModel) View() string {
	var lines []string

	// Title
	if m.mode == ModeConfigure {
		lines = append(lines, m.styles.Title.Render(" Configure Checkout Step "))
	} else {
		lines = append(lines, m.styles.Title.Render(" Checkout Branch "))
	}
	lines = append(lines, "")

	// Instructions
	if m.mode == ModeConfigure {
		lines = append(lines, m.styles.Help.Render("Select branch for checkout step:"))
		lines = append(lines, " • Use ↑/↓ to navigate")
		lines = append(lines, " • Press enter to save configuration")
		lines = append(lines, "")
	}

	// Branch list
	if len(m.branches) == 0 {
		lines = append(lines, m.styles.Info.Render("Loading branches..."))
	} else {
		current, _ := m.gitSvc.CurrentBranch()

		for i, branch := range m.branches {
			cursor := "  "
			if m.cursor == i {
				cursor = m.styles.Key.Render("▸ ")
			}

			branchDisplay := branch
			if branch == current {
				branchDisplay = m.styles.Success.Render(fmt.Sprintf("%s (current)", branch))
			}

			lines = append(lines, fmt.Sprintf("%s%s", cursor, branchDisplay))
		}
	}

	lines = append(lines, "")

	// Message
	if m.message != "" {
		if m.err != nil {
			lines = append(lines, m.styles.Error.Render(m.message))
		} else {
			lines = append(lines, m.styles.Success.Render(m.message))
		}
		lines = append(lines, "")
	}

	// Help
	if m.mode == ModeConfigure {
		lines = append(lines, m.styles.Help.Render("↑/↓: navigate | enter: save | esc: cancel"))
	} else {
		lines = append(lines, m.styles.Help.Render("↑/↓: navigate | enter: checkout | r: refresh | esc: back"))
	}

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadBranches loads all branches
func (m *CheckoutModel) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.gitSvc.GetBranches()
		if err != nil {
			return err
		}
		return branchesLoadedMsg(branches)
	}
}

// checkout switches to the selected branch
func (m *CheckoutModel) checkout(branch string) tea.Cmd {
	return func() tea.Msg {
		err := m.gitSvc.Checkout(branch)
		if err != nil {
			return err
		}
		return checkoutSuccessMsg(branch)
	}
}

// CommandModel interface implementation

// SetMode sets the operating mode of the command model
func (m *CheckoutModel) SetMode(mode CommandMode) {
	m.mode = mode
	if mode == ModeConfigure {
		m.message = ""
		m.err = nil
	}
}

// GetMode returns the current operating mode
func (m *CheckoutModel) GetMode() CommandMode {
	return m.mode
}

// GetParameters returns the collected parameters (only valid in configure mode)
func (m *CheckoutModel) GetParameters() map[string]string {
	if m.mode != ModeConfigure {
		return nil
	}

	params := make(map[string]string)
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		params["branch"] = m.branches[m.cursor]
	}
	return params
}

// Execute returns a command to execute the operation (only valid in execute mode)
func (m *CheckoutModel) Execute() tea.Cmd {
	if m.mode != ModeExecute {
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
