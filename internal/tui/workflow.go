package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"gitflow-tui/internal/config"
	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// runWorkflowMsg is sent when user wants to run a workflow
type runWorkflowMsg models.Workflow

// createWorkflowMsg is sent when user wants to create a new workflow
type createWorkflowMsg struct {
	fromTemplate *models.Workflow // nil for blank workflow
}

// editWorkflowMsg is sent when user wants to edit a workflow
type editWorkflowMsg models.Workflow

// deleteWorkflowMsg is sent when user wants to delete a workflow
type deleteWorkflowMsg int // index of workflow to delete

// duplicateWorkflowMsg is sent when user wants to duplicate a workflow
type duplicateWorkflowMsg int // index of workflow to duplicate

// workflowsLoadedMsg is sent when workflows are loaded from config
type workflowsLoadedMsg struct {
	workflows    []models.Workflow
	saveToConfig bool
}

// WorkflowModel handles the workflow builder screen
type WorkflowModel struct {
	gitSvc    git.GitService
	configMgr *config.Manager
	styles    Styles
	workflows []models.Workflow
	cursor    int
	selected  *models.Workflow
	editing   bool
	running   bool
	message   string
}

// NewWorkflowModel creates a new workflow model
func NewWorkflowModel(gitSvc git.GitService, configMgr *config.Manager, styles Styles) *WorkflowModel {
	return &WorkflowModel{
		gitSvc:    gitSvc,
		configMgr: configMgr,
		styles:    styles,
		workflows: []models.Workflow{}, // Will be loaded in Init()
	}
}

// Init initializes the model and loads workflows from config
func (m *WorkflowModel) Init() tea.Cmd {
	return m.loadWorkflows()
}

// loadWorkflows loads workflows from config
func (m *WorkflowModel) loadWorkflows() tea.Cmd {
	return func() tea.Msg {
		// Load config if not already loaded
		if m.configMgr == nil {
			m.configMgr = config.NewManager()
			if err := m.configMgr.Load(); err != nil {
				// Config load error - use defaults
				return workflowsLoadedMsg{workflows: m.createDefaultWorkflows(), saveToConfig: true}
			}
		}

		// Try to load workflows from config
		workflows := m.configMgr.GetWorkflows()
		if len(workflows) == 0 {
			// No workflows in config, create defaults
			return workflowsLoadedMsg{workflows: m.createDefaultWorkflows(), saveToConfig: true}
		}
		return workflowsLoadedMsg{workflows: workflows, saveToConfig: false}
	}
}

// createDefaultWorkflows creates default workflows with IDs
func (m *WorkflowModel) createDefaultWorkflows() []models.Workflow {
	templates := config.GetTemplates()
	now := time.Now()
	for i := range templates {
		templates[i].ID = uuid.New().String()
		templates[i].CreatedAt = now
	}
	return templates
}

// Update handles messages
func (m *WorkflowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case workflowsLoadedMsg:
		m.workflows = msg.workflows
		if msg.saveToConfig {
			// Save defaults to config
			for _, wf := range m.workflows {
				m.configMgr.SaveWorkflow(wf)
			}
		}
		return m, nil
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
				return m, func() tea.Msg {
					return runWorkflowMsg(m.workflows[m.cursor])
				}
			}
		case "n":
			// Create new blank workflow
			return m, func() tea.Msg {
				return createWorkflowMsg{fromTemplate: nil}
			}
		case "N":
			// Create from template
			return m, func() tea.Msg {
				return createWorkflowMsg{fromTemplate: &config.GetTemplates()[0]} // Will show template selector
			}
		case "e":
			// Edit selected workflow
			if len(m.workflows) > 0 && m.cursor < len(m.workflows) {
				return m, func() tea.Msg {
					return editWorkflowMsg(m.workflows[m.cursor])
				}
			}
		case "d":
			// Delete selected workflow
			if len(m.workflows) > 0 && m.cursor < len(m.workflows) {
				return m.deleteWorkflow(m.cursor)
			}
		case "c", "ctrl+d":
			// Duplicate selected workflow
			if len(m.workflows) > 0 && m.cursor < len(m.workflows) {
				return m.duplicateWorkflow(m.cursor)
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
	lines = append(lines, m.styles.Help.Render("↑/↓:nav | enter:run | n:new | N:template | e:edit | d:del | c:copy | esc:back"))

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// deleteWorkflow removes a workflow from the list and config
func (m *WorkflowModel) deleteWorkflow(index int) (tea.Model, tea.Cmd) {
	if index < 0 || index >= len(m.workflows) {
		return m, nil
	}

	workflow := m.workflows[index]

	// Remove from config
	if m.configMgr != nil {
		if err := m.configMgr.DeleteWorkflow(workflow.ID); err != nil {
			m.message = fmt.Sprintf("Error deleting: %v", err)
			return m, nil
		}
	}

	// Remove from local slice
	m.workflows = append(m.workflows[:index], m.workflows[index+1:]...)

	// Adjust cursor
	if m.cursor >= len(m.workflows) && m.cursor > 0 {
		m.cursor--
	}

	m.message = fmt.Sprintf("Deleted: %s", workflow.Name)
	return m, nil
}

// duplicateWorkflow creates a copy of a workflow
func (m *WorkflowModel) duplicateWorkflow(index int) (tea.Model, tea.Cmd) {
	if index < 0 || index >= len(m.workflows) {
		return m, nil
	}

	original := m.workflows[index]

	// Create copy
	copy := models.Workflow{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("%s Copy", original.Name),
		Description: original.Description,
		Steps:       make([]models.WorkflowStep, len(original.Steps)),
		CreatedAt:   time.Now(),
	}

	// Copy steps
	for i, step := range original.Steps {
		copy.Steps[i] = step
		// Deep copy parameters map
		if step.Parameters != nil {
			copy.Steps[i].Parameters = make(map[string]string)
			for k, v := range step.Parameters {
				copy.Steps[i].Parameters[k] = v
			}
		}
	}

	// Save to config
	if m.configMgr != nil {
		if err := m.configMgr.SaveWorkflow(copy); err != nil {
			m.message = fmt.Sprintf("Error saving copy: %v", err)
			return m, nil
		}
	}

	// Add to list
	m.workflows = append(m.workflows, copy)
	m.cursor = len(m.workflows) - 1

	m.message = fmt.Sprintf("Duplicated: %s", copy.Name)
	return m, nil
}
