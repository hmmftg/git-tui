package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// NewBranchModel handles the new branch creation screen.
type NewBranchModel struct {
	BaseModel
	branches    []string
	cursor      int
	branchInput textinput.Model
	width       int
	success     bool
}

// NewNewBranchModel creates a new NewBranchModel.
func NewNewBranchModel(gitSvc git.GitService, styles Styles) *NewBranchModel {
	input := textinput.New()
	input.Placeholder = "new-branch-name"
	input.Width = 40
	input.Focus()

	return &NewBranchModel{
		BaseModel:   NewBaseModel(gitSvc, styles),
		branchInput: input,
		width:       70,
	}
}

// Init initializes the model.
func (m *NewBranchModel) Init() tea.Cmd {
	return tea.Batch(m.loadBranches(), textinput.Blink)
}

// SetSize updates input widths for the available content area.
func (m *NewBranchModel) SetSize(width, _ int) {
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

// Update handles messages and key events.
func (m *NewBranchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.branchInput.Focused() {
			switch msg.String() {
			case "esc":
				m.branchInput.Blur()
				return m, nil
			case "enter", "ctrl+s":
				m.branchInput.Blur()
				return m, nil
			case "tab":
				m.branchInput.Blur()
				return m, nil
			default:
				m.branchInput, cmd = m.branchInput.Update(msg)
				return m, cmd
			}
		}

		if m.Mode == ModeConfigure {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.branches)-1 {
					m.cursor++
				}
			case "enter", "ctrl+s":
				config := models.CommandConfig{
					StepType:   models.StepNewBranch,
					Parameters: m.GetParameters(),
				}
				return m, func() tea.Msg { return commandConfiguredMsg{Config: config} }
			case "esc":
				return m, func() tea.Msg { return commandCancelledMsg{} }
			case "i":
				m.branchInput.Focus()
				return m, textinput.Blink
			}
		} else {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.branches)-1 {
					m.cursor++
				}
			case "enter", "ctrl+s":
				return m, m.createBranch()
			case "esc":
				return m, NavigateCmd(models.ViewHome)
			case "i":
				m.branchInput.Focus()
				return m, textinput.Blink
			}
		}

	case branchesLoadedMsg:
		m.branches = []string(msg)
		m.Err = nil
		// Default base to current branch if possible
		if cur, err := m.GitSvc.CurrentBranch(); err == nil {
			for i, b := range m.branches {
				if b == cur {
					m.cursor = i
					break
				}
			}
		}
		return m, nil

	case newBranchSuccessMsg:
		m.success = true
		m.branchInput.SetValue("")
		m.SetSuccess("Created branch %s successfully!", msg)
		return m, nil

	case error:
		m.SetError(msg)
		return m, nil
	}

	return m, nil
}

// View renders the model.
func (m *NewBranchModel) View() string {
	var lines []string
	lines = append(lines, m.Styles.Title.Render("🌿 New Branch"))
	lines = append(lines, "")

	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Help.Render("Create a new branch from a base:"))
		lines = append(lines, "")
	}

	// Branch name input
	lines = append(lines, m.Styles.Info.Render("Branch Name:"))
	lines = append(lines, m.Styles.Input.Render(m.branchInput.View()))
	lines = append(lines, "")

	// Base branch list
	if len(m.branches) > 0 {
		lines = append(lines, m.Styles.Info.Render("Base Branch:"))
		for i, branch := range m.branches {
			cursor := "  "
			if m.cursor == i {
				cursor = m.Styles.MenuActive.Render("> ")
				branch = m.Styles.MenuActive.Render(branch)
			}
			lines = append(lines, cursor+branch)
		}
	}

	if m.HasMessage() {
		lines = append(lines, "")
		lines = append(lines, m.RenderMessage())
	}

	// Help
	if m.Mode == ModeConfigure {
		lines = append(lines, m.Styles.Help.Render("i: edit name | ↑/↓: choose base | enter/ctrl+s: save | esc: cancel"))
	} else {
		if m.success {
			lines = append(lines, m.Styles.Help.Render("esc: back | i: new branch"))
		} else {
			lines = append(lines, m.Styles.Help.Render("i: edit name | ↑/↓: choose base | enter: create | esc: back"))
		}
	}

	return m.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// loadBranches loads all branches.
func (m *NewBranchModel) loadBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.GitSvc.GetBranches()
		if err != nil {
			return err
		}
		return branchesLoadedMsg(branches)
	}
}

// createBranch creates the new branch from the selected base.
func (m *NewBranchModel) createBranch() tea.Cmd {
	name := m.branchInput.Value()
	if name == "" {
		return func() tea.Msg { return fmt.Errorf("branch name cannot be empty") }
	}
	base := ""
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		base = m.branches[m.cursor]
	}
	return func() tea.Msg {
		err := m.GitSvc.CreateBranch(name, base)
		if err != nil {
			return err
		}
		return newBranchSuccessMsg(name)
	}
}

// IsInputFocused reports whether the branch name input is focused.
func (m *NewBranchModel) IsInputFocused() bool {
	return m.branchInput.Focused()
}

// CommandModel interface implementation.

// SetMode sets the operating mode.
func (m *NewBranchModel) SetMode(mode CommandMode) {
	m.BaseModel.SetMode(mode)
	if mode == ModeConfigure {
		m.success = false
	}
}

// GetMode returns the current operating mode.
func (m *NewBranchModel) GetMode() CommandMode {
	return m.BaseModel.GetMode()
}

// GetParameters returns the collected parameters.
func (m *NewBranchModel) GetParameters() map[string]string {
	if m.Mode != ModeConfigure {
		return nil
	}
	params := make(map[string]string)
	params["branch"] = m.branchInput.Value()
	if m.cursor >= 0 && m.cursor < len(m.branches) {
		params["base"] = m.branches[m.cursor]
	}
	return params
}

// SetParameters loads existing parameters.
func (m *NewBranchModel) SetParameters(params map[string]string) {
	if params == nil {
		return
	}
	m.branchInput.SetValue(params["branch"])
	base := params["base"]
	for i, b := range m.branches {
		if b == base {
			m.cursor = i
			break
		}
	}
}

// Execute returns a command to create the branch.
func (m *NewBranchModel) Execute() tea.Cmd {
	if m.Mode != ModeExecute {
		return nil
	}
	return m.createBranch()
}

// GetStepType returns the step type this command model represents.
func (m *NewBranchModel) GetStepType() models.StepType {
	return models.StepNewBranch
}

type newBranchSuccessMsg string
