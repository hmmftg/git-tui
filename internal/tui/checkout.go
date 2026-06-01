package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// CheckoutModel handles the checkout screen
type CheckoutModel struct {
	gitSvc     git.GitService
	styles     Styles
	branches   []string
	cursor     int
	selected   string
	err        error
	message    string
}

// NewCheckoutModel creates a new checkout model
func NewCheckoutModel(gitSvc git.GitService, styles Styles) *CheckoutModel {
	return &CheckoutModel{
		gitSvc: gitSvc,
		styles: styles,
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
				return m, m.checkout(m.branches[m.cursor])
			}
		case "r":
			return m, m.loadBranches()
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
	lines = append(lines, m.styles.Title.Render(" Checkout Branch "))
	lines = append(lines, "")

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
	lines = append(lines, m.styles.Help.Render("↑/↓: navigate | enter: checkout | r: refresh | esc: back"))

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

type branchesLoadedMsg []string
type checkoutSuccessMsg string
