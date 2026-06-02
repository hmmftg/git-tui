package tui

import (
	"fmt"
	"sort"
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

	// Command registry for creating command models
	commandRegistry *CommandRegistry

	// Editing mode
	mode       editorMode // 0=name, 1=description, 2=steps, 3=adding step, 4=editing step, 5=command selection
	stepCursor int        // for steps list

	// Inputs
	nameInput  textinput.Model
	descInput  textinput.Model
	paramInput textinput.Model

	// Focus management
	focusIndex int // 0=name, 1=desc, 2=param, 3=no focus

	// Step type selection for adding
	stepTypeCursor int
	availableTypes []models.StepType

	// Parameter editing
	paramCursor        int      // cursor for parameter list
	paramKeys          []string // parameter keys for current step
	editingParamKey    string   // currently editing parameter key
	editingParamValue  bool
	previousParamValue string

	// Command configuration
	currentCommandModel CommandModel // Currently active command model for configuration
}

type editorMode int

const (
	modeName editorMode = iota
	modeDescription
	modeSteps
	modeAddStep
	modeEditStep
	modeCommandSelection
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

	// Initialize command registry
	commandRegistry := NewCommandRegistry(gitSvc, styles)
	availableTypes := commandRegistry.GetAvailableStepTypes()

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

	model := &WorkflowEditorModel{
		gitSvc:          gitSvc,
		configMgr:       configMgr,
		styles:          styles,
		workflow:        wf,
		isNew:           isNew,
		commandRegistry: commandRegistry,
		mode:            modeName,
		stepCursor:      0,
		nameInput:       nameTi,
		descInput:       descTi,
		paramInput:      paramTi,
		focusIndex:      0, // Start with name input focused (index 0)
		availableTypes:  availableTypes,
		stepTypeCursor:  0,
	}

	// Initialize focus state
	model.updateFocus()

	return model
}

// Init initializes the editor
func (m *WorkflowEditorModel) Init() tea.Cmd {
	return textinput.Blink
}

// IsInputFocused returns true if any text input is currently focused
func (m *WorkflowEditorModel) IsInputFocused() bool {
	return m.nameInput.Focused() || m.descInput.Focused() || m.paramInput.Focused()
}

// SetSize updates input widths for the available content area.
func (m *WorkflowEditorModel) SetSize(width, _ int) {
	if width <= 0 {
		return
	}
	inputWidth := width - 12
	if inputWidth < 20 {
		inputWidth = 20
	}
	m.nameInput.Width = inputWidth
	m.descInput.Width = inputWidth
	m.paramInput.Width = inputWidth
}

// updateFocus updates focus state for all inputs based on current focusIndex
func (m *WorkflowEditorModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, 3)

	for i := range cmds {
		if i == m.focusIndex {
			// Set focused state
			switch i {
			case 0:
				cmds[i] = m.nameInput.Focus()
			case 1:
				cmds[i] = m.descInput.Focus()
			case 2:
				cmds[i] = m.paramInput.Focus()
			}
		} else {
			// Remove focused state
			switch i {
			case 0:
				m.nameInput.Blur()
			case 1:
				m.descInput.Blur()
			case 2:
				m.paramInput.Blur()
			}
		}
	}

	return tea.Batch(cmds...)
}

// setFocus sets the focus to the specified input index
func (m *WorkflowEditorModel) setFocus(index int) tea.Cmd {
	m.focusIndex = index
	return m.updateFocus()
}

// setFocusByName sets focus by input name (for backward compatibility)
func (m *WorkflowEditorModel) setFocusByName(input string) tea.Cmd {
	switch input {
	case "name":
		return m.setFocus(0)
	case "desc":
		return m.setFocus(1)
	case "param":
		return m.setFocus(2)
	default:
		return m.setFocus(3) // No focus
	}
}

// Update handles messages
func (m *WorkflowEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Forward ordinary messages to an active command model before editor-level key handling.
	// Configuration completion/cancellation messages are emitted by the command model
	// and must be handled by the editor itself.
	if m.currentCommandModel != nil {
		switch msg.(type) {
		case commandConfiguredMsg, commandCancelledMsg:
			// handled below
		default:
			model, cmd := m.currentCommandModel.Update(msg)
			m.currentCommandModel = model.(CommandModel)
			return m, cmd
		}
	}

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
		case modeEditStep:
			return m.handleEditStepInput(msg)
		case modeCommandSelection:
			return m.handleCommandSelectionInput(msg)
		}

	case commandConfiguredMsg:
		// Handle command configuration completion
		if m.currentCommandModel != nil {
			config := msg.Config
			// Update or create the step with the new configuration
			m.updateStepWithConfig(config)
			m.currentCommandModel = nil
			m.mode = modeSteps
		}
		return m, nil

	case commandCancelledMsg:
		// Handle command configuration cancellation
		m.currentCommandModel = nil
		m.mode = modeSteps
		return m, nil
	}

	return m, cmd
}

func (m *WorkflowEditorModel) handleNameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "enter":
		m.workflow.Name = m.nameInput.Value()
		m.mode = modeDescription
		return m, m.setFocus(1) // Focus desc input (index 1)
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
		return m, m.setFocus(3) // No input focused when in steps mode (index 3)
	case "esc":
		return m, func() tea.Msg { return cancelEditMsg{} }
	case "ctrl+s":
		m.workflow.Description = m.descInput.Value()
		return m, m.saveWorkflow()
	case "shift+tab":
		m.workflow.Description = m.descInput.Value()
		m.mode = modeName
		return m, m.setFocus(0) // Focus name input (index 0)
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
			return m, m.setFocus(1) // Focus desc input (index 1)
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
		// Edit step using command selection
		if len(m.workflow.Steps) > 0 && m.stepCursor < len(m.workflow.Steps) {
			m.mode = modeCommandSelection
			m.stepTypeCursor = 0
			// Find current step type in available types
			currentStep := m.workflow.Steps[m.stepCursor]
			for i, stepType := range m.availableTypes {
				if stepType == currentStep.Type {
					m.stepTypeCursor = i
					break
				}
			}
		}
		return m, nil
	case "tab":
		m.mode = modeName
		return m, m.setFocus(0)
	case "shift+tab":
		m.mode = modeDescription
		return m, m.setFocus(1)
	case "esc":
		m.mode = modeDescription
		return m, m.setFocus(1)
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
		// Create command model for selected step type
		if m.stepTypeCursor < len(m.availableTypes) {
			stepType := m.availableTypes[m.stepTypeCursor]
			commandModel := m.commandRegistry.CreateCommandModel(stepType, ModeConfigure)
			if commandModel != nil {
				m.currentCommandModel = commandModel
				// Set up for new step creation
				m.stepCursor = len(m.workflow.Steps) // Will be set when config is saved
				return m, commandModel.Init()
			}
		}
		return m, nil
	case "esc":
		m.mode = modeSteps
		return m, nil
	}

	return m, nil
}

func (m *WorkflowEditorModel) handleEditStepInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.workflow.Steps) == 0 || m.stepCursor >= len(m.workflow.Steps) {
		return m, nil
	}

	step := &m.workflow.Steps[m.stepCursor]

	switch msg.String() {
	case "up", "k":
		if m.editingParamKey == "" && !m.editingParamValue && m.paramCursor > 0 {
			m.paramCursor--
		}
		return m, nil
	case "down", "j":
		if m.editingParamKey == "" && !m.editingParamValue && m.paramCursor < len(m.paramKeys)-1 {
			m.paramCursor++
		}
		return m, nil
	case "enter":
		if m.editingParamKey == "new" {
			newKey := m.paramInput.Value()
			if newKey != "" {
				step.Parameters[newKey] = ""
				m.updateParamKeys()
				for i, key := range m.paramKeys {
					if key == newKey {
						m.paramCursor = i
						break
					}
				}
			}
			m.paramInput.SetValue("")
			m.paramInput.Placeholder = "Enter value"
			m.editingParamKey = ""
			return m, m.setFocus(3)
		}

		if m.editingParamValue {
			if m.editingParamKey != "" {
				step.Parameters[m.editingParamKey] = m.paramInput.Value()
			}
			m.editingParamKey = ""
			m.editingParamValue = false
			m.previousParamValue = ""
			m.paramInput.SetValue("")
			return m, m.setFocus(3)
		}

		if m.paramCursor < len(m.paramKeys) {
			key := m.paramKeys[m.paramCursor]
			m.editingParamKey = key
			m.editingParamValue = true
			m.previousParamValue = step.Parameters[key]
			m.paramInput.Placeholder = "Enter value"
			m.paramInput.SetValue(step.Parameters[key])
			return m, m.setFocus(2)
		}
		return m, nil
	case "d":
		if m.editingParamKey == "" && !m.editingParamValue && m.paramCursor < len(m.paramKeys) {
			key := m.paramKeys[m.paramCursor]
			delete(step.Parameters, key)
			m.updateParamKeys()
			if m.paramCursor >= len(m.paramKeys) && m.paramCursor > 0 {
				m.paramCursor--
			}
		}
		return m, nil
	case "a":
		if m.editingParamKey != "" || m.editingParamValue {
			break
		}
		m.paramInput.SetValue("")
		m.paramInput.Placeholder = "Enter new parameter name"
		m.editingParamKey = "new"
		return m, m.setFocus(2)
	case "tab":
		if m.editingParamKey == "" && !m.editingParamValue && m.paramCursor < len(m.paramKeys)-1 {
			m.paramCursor++
		}
		return m, nil
	case "shift+tab":
		if m.editingParamKey == "" && !m.editingParamValue && m.paramCursor > 0 {
			m.paramCursor--
		}
		return m, nil
	case "esc":
		if m.editingParamKey != "" || m.editingParamValue {
			m.editingParamKey = ""
			m.editingParamValue = false
			m.previousParamValue = ""
			m.paramInput.SetValue("")
			return m, m.setFocus(3)
		}
		m.mode = modeSteps
		return m, m.setFocus(3)
	case "ctrl+s":
		return m, m.saveWorkflow()
	}

	if m.editingParamKey != "" || m.editingParamValue {
		var cmd tea.Cmd
		m.paramInput, cmd = m.paramInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *WorkflowEditorModel) updateParamKeys() {
	if len(m.workflow.Steps) == 0 || m.stepCursor >= len(m.workflow.Steps) {
		m.paramKeys = []string{}
		return
	}

	step := m.workflow.Steps[m.stepCursor]
	m.paramKeys = make([]string, 0, len(step.Parameters))
	for key := range step.Parameters {
		m.paramKeys = append(m.paramKeys, key)
	}
	sort.Strings(m.paramKeys)
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
		lines = append(lines, m.styles.Help.Render("enter:configure | esc:cancel"))
	} else if m.mode == modeCommandSelection {
		// Show command type selector for editing
		lines = append(lines, m.styles.Info.Render("  Select command type:"))
		for i, st := range m.availableTypes {
			cursor := "  "
			if i == m.stepTypeCursor {
				cursor = m.styles.Key.Render("▸ ")
			}
			lines = append(lines, fmt.Sprintf("%s%s", cursor, st.String()))
		}
		lines = append(lines, "")
		lines = append(lines, m.styles.Help.Render("enter:configure | esc:cancel"))
	} else if m.mode == modeEditStep {
		// Show step parameter editor
		if len(m.workflow.Steps) > 0 && m.stepCursor < len(m.workflow.Steps) {
			step := m.workflow.Steps[m.stepCursor]
			lines = append(lines, m.styles.Info.Render(fmt.Sprintf("  Editing parameters for: %s", step.Type.String())))
			lines = append(lines, "")

			if len(m.paramKeys) == 0 {
				lines = append(lines, m.styles.Dimmed.Render("  (No parameters. Press 'a' to add)"))
			} else {
				for i, key := range m.paramKeys {
					cursor := "  "
					if i == m.paramCursor {
						cursor = m.styles.Key.Render("▸ ")
					}
					value := step.Parameters[key]
					line := fmt.Sprintf("%s%s: %s", cursor, key, value)
					if i == m.paramCursor {
						lines = append(lines, m.styles.Warning.Render(line))
					} else {
						lines = append(lines, line)
					}
				}
			}
			lines = append(lines, "")
			lines = append(lines, m.paramInput.View())
			lines = append(lines, "")
		}
		lines = append(lines, m.styles.Help.Render("↑/↓:nav | enter:save | a:add | d:del | esc:back"))
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

	// If there's an active command model, display it instead of the normal UI
	if m.currentCommandModel != nil {
		return m.currentCommandModel.View()
	}

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// handleCommandSelectionInput handles input for command selection mode
func (m *WorkflowEditorModel) handleCommandSelectionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		// Create command model for selected step type
		if m.stepTypeCursor < len(m.availableTypes) {
			stepType := m.availableTypes[m.stepTypeCursor]
			commandModel := m.commandRegistry.CreateCommandModel(stepType, ModeConfigure)
			if commandModel != nil {
				if m.stepCursor < len(m.workflow.Steps) {
					commandModel.SetParameters(m.workflow.Steps[m.stepCursor].Parameters)
				}
				m.currentCommandModel = commandModel
				return m, commandModel.Init()
			}
		}
		return m, nil
	case "esc":
		m.mode = modeSteps
		return m, nil
	}
	return m, nil
}

// updateStepWithConfig updates the current step with new configuration
func (m *WorkflowEditorModel) updateStepWithConfig(config models.CommandConfig) {
	if len(m.workflow.Steps) == 0 || m.stepCursor >= len(m.workflow.Steps) {
		// Create new step
		newStep := models.WorkflowStep{
			Type:        config.StepType,
			Parameters:  config.Parameters,
			Description: fmt.Sprintf("%s step", config.StepType.String()),
		}
		m.workflow.Steps = append(m.workflow.Steps, newStep)
		m.stepCursor = len(m.workflow.Steps) - 1
	} else {
		// Update existing step
		step := &m.workflow.Steps[m.stepCursor]
		step.Type = config.StepType
		step.Parameters = config.Parameters
		step.Description = fmt.Sprintf("%s step", config.StepType.String())
	}
}
