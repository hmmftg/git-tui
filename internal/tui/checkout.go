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

// CheckoutModel handles the checkout screen
type CheckoutModel struct {
	BaseModel
	branches []string
	cursor   int
	selected string

	creating     bool
	createNew    bool
	createBase   string
	createBranch string
	branchInput  textinput.Model
	width        int
}

// NewCheckoutModel creates a new checkout model
func NewCheckoutModel(gitSvc git.GitService, styles Styles) *CheckoutModel {
	branchInput := textinput.New()
	branchInput.Placeholder = "new-branch-name"
	branchInput.Width = 40

	return &CheckoutModel{
		BaseModel:   NewBaseModel(gitSvc, styles),
		branchInput: branchInput,
		width:       70,
	}
}

// Init initializes the model
func (m *CheckoutModel) Init() tea.Cmd {
	return tea.Batch(m.loadBranches(), textinput.Blink)
}

// SetSize updates input widths for the available content area.
func (m *CheckoutModel) SetSize(width, _ int) {
	if width <= 0 {
		return
	}
	m.width = width
	inputWidth := width - 12
	if inputWidth < 20 {
		inputWidth = 20
	}
	m.branchInput.Width = inputWidth
}

// IsInputFocused reports whether the create-branch input is focused.
func (m *CheckoutModel) IsInputFocused() bool {
	return m.branchInput.Focused()
}

// Update handles messages
func (m *CheckoutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.creating {
			return m.handleCreateBranchInput(msg)
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.branches)-1 {
				m.cursor++
			}
		case "n":
			if len(m.branches) > 0 && m.cursor < len(m.branches) {
				m.creating = true
				m.createNew = true
				m.createBase = m.branches[m.cursor]
				m.createBranch = ""
				m.branchInput.SetValue("")
				return m, m.branchInput.Focus()
			}
		case "enter":
			if len(m.branches) > 0 && m.cursor < len(m.branches) {
				if m.Mode == ModeConfigure {
					config := models.CommandConfig{
						StepType:   models.StepCheckout,
						Parameters: m.GetParameters(),
					}
					return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
				}
				return m, m.checkout(m.branches[m.cursor])
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
		if m.createBase != "" {
			for i, b := range m.branches {
				if b == m.createBase {
					m.cursor = i
					return m, nil
				}
			}
		}
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

	case branchCreatedMsg:
		m.selected = msg.branch
		m.createBranch = msg.branch
		m.createBase = msg.base
		m.Message = fmt.Sprintf("✔ Created %s from %s", msg.branch, msg.base)
		return m, m.loadBranches()

	case error:
		m.Err = msg
		m.Message = fmt.Sprintf("✘ Error: %v", msg)
		return m, nil
	}

	return m, nil
}

func (m *CheckoutModel) handleCreateBranchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "ctrl+s", "ctrl+enter":
		branch := strings.TrimSpace(m.branchInput.Value())
		if branch == "" {
			return m, nil
		}
		m.createBranch = branch
		m.createNew = true
		m.creating = false
		m.branchInput.Blur()
		if m.Mode == ModeConfigure {
			config := models.CommandConfig{
				StepType:   models.StepCheckout,
				Parameters: m.GetParameters(),
			}
			return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
		}
		return m, m.createBranchFromBase(branch, m.createBase)
	case "esc":
		m.creating = false
		m.createNew = false
		m.createBranch = ""
		m.branchInput.SetValue("")
		m.branchInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.branchInput, cmd = m.branchInput.Update(msg)
	return m, cmd
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
		lines = append(lines, m.Styles.Help.Render("Select a branch to checkout, or create a new branch from a base:"))
		lines = append(lines, " • Use ↑/↓ to choose the checkout/base branch")
		lines = append(lines, " • Press n to create a new branch from the selected base")
		lines = append(lines, " • Press enter to save checkout configuration")
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
			if m.creating && branch == m.createBase {
				branchDisplay = m.Styles.Warning.Render(fmt.Sprintf("%s (base)", branch))
			}

			lines = append(lines, fmt.Sprintf("%s%s", cursor, branchDisplay))
		}
	}

	if m.creating {
		lines = append(lines, "")
		lines = append(lines, m.Styles.Info.Render(fmt.Sprintf("New branch based on: %s", m.createBase)))
		lines = append(lines, m.Styles.Input.Render(m.branchInput.View()))
		lines = append(lines, m.Styles.Help.Render("enter/ctrl+s: create | esc: cancel"))
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
		lines = append(lines, m.Styles.Help.Render("↑/↓: navigate | n: new from selected | enter: save checkout | esc: cancel"))
	} else {
		lines = append(lines, m.Styles.Help.Render("↑/↓: navigate | enter: checkout | n: new from selected | r: refresh | esc: back"))
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

func (m *CheckoutModel) createBranchFromBase(branch, base string) tea.Cmd {
	return func() tea.Msg {
		err := m.GitSvc.CreateBranch(branch, base)
		if err != nil {
			return err
		}
		return branchCreatedMsg{branch: branch, base: base}
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
	if m.createNew {
		params["create"] = "true"
		params["branch"] = m.createBranch
		params["base"] = m.createBase
		return params
	}
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		params["branch"] = m.branches[m.cursor]
	}
	return params
}

// SetParameters loads an existing checkout command configuration.
func (m *CheckoutModel) SetParameters(params map[string]string) {
	if params == nil {
		return
	}
	m.createNew = params["create"] == "true"
	m.selected = params["branch"]
	m.createBranch = params["branch"]
	m.createBase = params["base"]
	if m.createNew {
		m.branchInput.SetValue(m.createBranch)
	}
	for i, branch := range m.branches {
		if branch == m.selected || branch == m.createBase {
			m.cursor = i
			break
		}
	}
}

func (m *CheckoutModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	if m.createNew {
		return m.createBranchFromBase(m.createBranch, m.createBase)
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
type branchCreatedMsg struct {
	branch string
	base   string
}
