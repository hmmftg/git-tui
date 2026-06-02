package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"gitflow-tui/internal/config"
	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// saveWorkflowMsg is sent when workflow should be saved
type saveWorkflowMsg models.Workflow

// cancelEditMsg cancels the workflow editor
type cancelEditMsg struct{}

// WorkflowEditorModel handles editing a workflow
type WorkflowEditorModel struct {
	gitSvc    git.GitService
	configMgr *config.Manager
	styles    Styles
	workflow  models.Workflow
	isNew     bool

	// Editing mode
	mode       editorMode // 0=name, 1=description, 2=steps, 3=adding step
	stepCursor int        // for steps list

	// Inputs
	nameInput  textinput.Model
	descInput  textinput.Model
	paramInput textinput.Model

	// Step type selection for adding
	stepTypeCursor int
	availableTypes []models.StepType
}

type editorMode int

const (
	modeName editorMode = iota
	modeDescription
	modeSteps
	modeAddStep
)

// NewWorkflowEditor creates a workflow editor for a new or existing workflow
func NewWorkflowEditor(gitSvc git.GitService, configMgr *config.Manager, styles Styles, workflow *models.Workflow, isTemplate bool) *WorkflowEditorModel {
	// Initialize text inputs
	nameTi := textinput.New()
	nameTi.Placeholder = "Workflow name"
	nameTi.Focus()

	descTi := textinput.New()
	descTi.Placeholder = "Description"

	paramTi := textinput.New()
	paramTi.Placeholder = "Parameter value"

	// Available step types
	availableTypes := []models.StepType{
		models.StepStatus,
		models.StepCommit,
		models.StepPush,
		models.StepPull,
		models.StepCheckout,
		models.StepMerge,
		models.StepRebase,
	}

	var wf models.Workflow
	isNew := false

	if workflow != nil {
		wf = *workflow
		nameTi.SetValue(wf.Name)
		descTi.SetValue(wf.Description)
		isNew = isTemplate
	} else {
		// Create new blank workflow
		wf = models.Workflow{
			ID:        uuid.New().String(),
			Steps:     []models.WorkflowStep{},
			CreatedAt: time.Now(),
		}
		isNew = true
	}

	return &WorkflowEditorModel{
		gitSvc:         gitSvc,
		configMgr:      configMgr,
		styles:         styles,
		workflow:       wf,
		isNew:          isNew,
		mode:           modeName,
		stepCursor:     0,
		nameInput:      nameTi,
		descInput:      descTi,
		paramInput:     paramTi,
		availableTypes: availableTypes,
		stepTypeCursor: 0,
	}
}

// Init initializes the editor
func (m *WorkflowEditorModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m *WorkflowEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case modeName:
			return m.handleNameInput(msg)
		case modeDescription:
			return m.handleDescInput(msg)
		case modeSteps:
			return m.handleStepsInput(msg)
		case modeAddStep:
			return m.handleAddStepInput(msg)
		}
	}

	return m, cmd
}

func (m *WorkflowEditorModel) handleNameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "enter":
		m.workflow.Name = m.nameInput.Value()
		m.mode = modeDescription
		m.descInput.Focus()
		return m, nil
	case "esc":
		return m, func() tea.Msg { return cancelEditMsg{} }
	case "ctrl+s":
		m.workflow.Name = m.nameInput.Value()
		return m, m.saveWorkflow()
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m *WorkflowEditorModel) handleDescInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "enter":
		m.workflow.Description = m.descInput.Value()
		m.mode = modeSteps
		return m, nil
	case "esc":
		return m, func() tea.Msg { return cancelEditMsg{} }
	case "ctrl+s":
		m.workflow.Description = m.descInput.Value()
		return m, m.saveWorkflow()
	case "shift+tab":
		m.workflow.Description = m.descInput.Value()
		m.mode = modeName
		m.nameInput.Focus()
		return m, nil
	}

	var cmd tea.Cmd
	m.descInput, cmd = m.descInput.Update(msg)
	return m, cmd
}

func (m *WorkflowEditorModel) handleStepsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "a":
		// Add step
		m.mode = modeAddStep
		m.stepTypeCursor = 0
		return m, nil
	case "x":
		// Delete step at cursor
		if len(m.workflow.Steps) > 0 && m.stepCursor < len(m.workflow.Steps) {
			m.workflow.Steps = append(m.workflow.Steps[:m.stepCursor], m.workflow.Steps[m.stepCursor+1:]...)
			if m.stepCursor >= len(m.workflow.Steps) && m.stepCursor > 0 {
				m.stepCursor--
			}
		}
		return m, nil
	case "up", "k":
		if m.stepCursor > 0 {
			m.stepCursor--
		} else if len(m.workflow.Steps) == 0 {
			m.mode = modeDescription
			m.descInput.Focus()
		}
		return m, nil
	case "down", "j":
		if m.stepCursor < len(m.workflow.Steps)-1 {
			m.stepCursor++
		}
		return m, nil
	case "shift+up":
		// Move step up
		if m.stepCursor > 0 {
			m.workflow.Steps[m.stepCursor], m.workflow.Steps[m.stepCursor-1] = m.workflow.Steps[m.stepCursor-1], m.workflow.Steps[m.stepCursor]
			m.stepCursor--
		}
		return m, nil
	case "shift+down":
		// Move step down
		if m.stepCursor < len(m.workflow.Steps)-1 {
			m.workflow.Steps[m.stepCursor], m.workflow.Steps[m.stepCursor+1] = m.workflow.Steps[m.stepCursor+1], m.workflow.Steps[m.stepCursor]
			m.stepCursor++
		}
		return m, nil
	case "enter":
		// Edit step parameters (simplified - just show current)
		return m, nil
	case "tab":
		m.mode = modeName
		m.nameInput.Focus()
		return m, nil
	case "shift+tab":
		m.mode = modeDescription
		m.descInput.Focus()
		return m, nil
	case "esc":
		return m, func() tea.Msg { return cancelEditMsg{} }
	case "ctrl+s":
		return m, m.saveWorkflow()
	}

	return m, nil
}

func (m *WorkflowEditorModel) handleAddStepInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.stepTypeCursor > 0 {
			m.stepTypeCursor--
		}
		return m, nil
	case "down", "j":
		if m.stepTypeCursor < len(m.availableTypes)-1 {
			m.stepTypeCursor++
		}
		return m, nil
	case "enter":
		// Add the selected step type
		stepType := m.availableTypes[m.stepTypeCursor]
		newStep := models.WorkflowStep{
			Type:        stepType,
			Description: fmt.Sprintf("%s step", stepType.String()),
			Parameters:  map[string]string{},
		}

		// Add default parameters for specific step types
		switch stepType {
		case models.StepCommit:
			newStep.Parameters["message"] = "Auto commit"
			newStep.Parameters["autoAdd"] = "true"
		case models.StepPush:
			newStep.Parameters["remote"] = "origin"
		case models.StepMerge:
			newStep.Parameters["target"] = "main"
		case models.StepCheckout:
			newStep.Parameters["branch"] = "develop"
		}

		m.workflow.Steps = append(m.workflow.Steps, newStep)
		m.stepCursor = len(m.workflow.Steps) - 1
		m.mode = modeSteps
		return m, nil
	case "esc":
		m.mode = modeSteps
		return m, nil
	}

	return m, nil
}

func (m *WorkflowEditorModel) saveWorkflow() tea.Cmd {
	// Update workflow from inputs
	m.workflow.Name = m.nameInput.Value()
	m.workflow.Description = m.descInput.Value()
	m.workflow.UpdatedAt = time.Now()

	// Save to config
	if m.configMgr != nil {
		if err := m.configMgr.SaveWorkflow(m.workflow); err != nil {
			// Return error message
			return func() tea.Msg {
				return errorMsg(fmt.Sprintf("Failed to save: %v", err))
			}
		}
	}

	return func() tea.Msg {
		return saveWorkflowMsg(m.workflow)
	}
}

// View renders the editor
func (m *WorkflowEditorModel) View() string {
	var lines []string

	// Title
	title := " Edit Workflow "
	if m.isNew {
		title = " New Workflow "
	}
	lines = append(lines, m.styles.Title.Render(title))
	lines = append(lines, "")

	// Name field
	nameStyle := m.styles.Info
	if m.mode == modeName {
		nameStyle = m.styles.Key
	}
	lines = append(lines, nameStyle.Render("Name:"))
	lines = append(lines, m.nameInput.View())
	lines = append(lines, "")

	// Description field
	descStyle := m.styles.Info
	if m.mode == modeDescription {
		descStyle = m.styles.Key
	}
	lines = append(lines, descStyle.Render("Description:"))
	lines = append(lines, m.descInput.View())
	lines = append(lines, "")

	// Steps section
	stepsStyle := m.styles.Info
	if m.mode == modeSteps {
		stepsStyle = m.styles.Key
	}
	lines = append(lines, stepsStyle.Render(fmt.Sprintf("Steps (%d):", len(m.workflow.Steps))))

	if m.mode == modeAddStep {
		// Show step type selector
		lines = append(lines, m.styles.Info.Render("  Select step type to add:"))
		for i, st := range m.availableTypes {
			cursor := "  "
			if i == m.stepTypeCursor {
				cursor = m.styles.Key.Render("▸ ")
			}
			lines = append(lines, fmt.Sprintf("%s%s", cursor, st.String()))
		}
		lines = append(lines, "")
		lines = append(lines, m.styles.Help.Render("enter:add | esc:cancel"))
	} else {
		// Show steps list
		if len(m.workflow.Steps) == 0 {
			lines = append(lines, m.styles.Dimmed.Render("  (No steps yet. Press 'a' to add)"))
		} else {
			for i, step := range m.workflow.Steps {
				cursor := "  "
				if i == m.stepCursor && m.mode == modeSteps {
					cursor = m.styles.Key.Render("▸ ")
				}

				stepInfo := step.Type.String()
				if step.Description != "" {
					stepInfo = fmt.Sprintf("%s - %s", step.Type.String(), step.Description)
				}

				line := fmt.Sprintf("%s%d. %s", cursor, i+1, stepInfo)
				if i == m.stepCursor && m.mode == modeSteps {
					lines = append(lines, m.styles.Warning.Render(line))
				} else {
					lines = append(lines, line)
				}
			}
		}
		lines = append(lines, "")
	}

	// Help
	help := "tab:next | shift+tab:prev | ↑/↓:nav | a:add | x:del | shift+↑/↓:reorder | ctrl+s:save | esc:cancel"
	if m.mode == modeAddStep {
		help = "↑/↓:select | enter:add | esc:cancel"
	}
	lines = append(lines, m.styles.Help.Render(help))

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
