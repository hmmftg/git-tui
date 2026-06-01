package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// WorkflowModel handles the workflow builder screen
type WorkflowModel struct {
	gitSvc      git.GitService
	styles      Styles
	workflows   []models.Workflow
	cursor      int
	selected    *models.Workflow
	editing     bool
	running     bool
	message     string
}

// NewWorkflowModel creates a new workflow model
func NewWorkflowModel(gitSvc git.GitService, styles Styles) *WorkflowModel {
	return &WorkflowModel{
		gitSvc: gitSvc,
		styles: styles,
		workflows: []models.Workflow{
			{
				ID:          uuid.New().String(),
				Name:        "Quick Commit",
				Description: "Add, commit and push in one go",
				Steps: []models.WorkflowStep{
					{Type: models.StepStatus, Description: "Check status"},
					{Type: models.StepCommit, Parameters: map[string]string{"autoAdd": "true"}, Description: "Commit all changes"},
					{Type: models.StepPush, Description: "Push to remote"},
				},
				CreatedAt: time.Now(),
			},
			{
				ID:          uuid.New().String(),
				Name:        "Release",
				Description: "Prepare a release",
				Steps: []models.WorkflowStep{
					{Type: models.StepStatus, Description: "Check status"},
					{Type: models.StepCommit, Description: "Commit changes"},
					{Type: models.StepPush, Description: "Push to remote"},
					{Type: models.StepMerge, Parameters: map[string]string{"target": "main"}, Description: "Merge to main"},
				},
				CreatedAt: time.Now(),
			},
			{
				ID:          uuid.New().String(),
				Name:        "Sync",
				Description: "Pull latest changes",
				Steps: []models.WorkflowStep{
					{Type: models.StepPull, Description: "Pull from remote"},
					{Type: models.StepStatus, Description: "Check status"},
				},
				CreatedAt: time.Now(),
			},
		},
	}
}

// Init initializes the model
func (m *WorkflowModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *WorkflowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.workflows)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.workflows) > 0 && m.cursor < len(m.workflows) {
				m.selected = &m.workflows[m.cursor]
				m.message = fmt.Sprintf("Selected: %s", m.selected.Name)
			}
		}
	}

	return m, nil
}

// View renders the workflow screen
func (m *WorkflowModel) View() string {
	var lines []string

	// Title
	lines = append(lines, m.styles.Title.Render(" Workflow Manager "))
	lines = append(lines, "")

	// Workflow list
	if len(m.workflows) == 0 {
		lines = append(lines, m.styles.Info.Render("No workflows available"))
	} else {
		lines = append(lines, m.styles.Info.Render("Available Workflows:"))
		lines = append(lines, "")

		for i, workflow := range m.workflows {
			cursor := "  "
			if m.cursor == i {
				cursor = m.styles.Key.Render("▸ ")
			}

			name := m.styles.Value.Render(workflow.Name)
			desc := m.styles.Help.Render(workflow.Description)
			
			lines = append(lines, fmt.Sprintf("%s%s", cursor, name))
			lines = append(lines, fmt.Sprintf("    %s", desc))
			
			// Show steps
			if len(workflow.Steps) > 0 {
				stepNames := []string{}
				for _, step := range workflow.Steps {
					stepNames = append(stepNames, step.Type.String())
				}
				lines = append(lines, fmt.Sprintf("    Steps: %s", m.styles.Info.Render(fmt.Sprintf("%v", stepNames))))
			}
			
			lines = append(lines, "")
		}
	}

	// Message
	if m.message != "" {
		lines = append(lines, m.styles.Success.Render(m.message))
		lines = append(lines, "")
	}

	// Help
	lines = append(lines, m.styles.Help.Render("↑/↓: navigate | enter: select | esc: back"))

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
