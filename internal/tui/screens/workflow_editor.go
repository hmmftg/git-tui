package screens

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
	"gitflow-tui/internal/tui/forms"
)

// WorkflowEditorScreen edits a workflow.
type WorkflowEditorScreen struct {
	ctx       *tui.AppContext
	workflow  *models.Workflow
	form      *huh.Form
	name      string
	desc      string
	stepIndex int
	mode      editorMode
}

type editorMode int

const (
	modeMain editorMode = iota
	modeAddStep
	modeEditStep
)

// NewWorkflowEditorScreen creates a new workflow editor.
func NewWorkflowEditorScreen(ctx *tui.AppContext, wf *models.Workflow) *WorkflowEditorScreen {
	if wf == nil {
		wf = &models.Workflow{ID: uuid.New().String()}
	}
	e := &WorkflowEditorScreen{
		ctx:      ctx,
		workflow: wf,
		name:     wf.Name,
		desc:     wf.Description,
	}
	e.buildForm()
	return e
}

func (e *WorkflowEditorScreen) buildForm() {
	e.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Workflow Name").
				Value(&e.name),
			huh.NewInput().
				Title("Description").
				Value(&e.desc),
		),
	)
}

// Init initializes the screen.
func (e *WorkflowEditorScreen) Init() tea.Cmd {
	return e.form.Init()
}

// Update handles messages.
func (e *WorkflowEditorScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return NewWorkflowListScreen(e.ctx), nil
		case "s":
			e.save()
			return e, func() tea.Msg {
				return tui.WorkflowSavedMsg{Workflow: *e.workflow}
			}
		case "a":
			return e.openStepTypeForm()
		case "d":
			if e.stepIndex >= 0 && e.stepIndex < len(e.workflow.Steps) {
				e.workflow.Steps = append(e.workflow.Steps[:e.stepIndex], e.workflow.Steps[e.stepIndex+1:]...)
				if e.stepIndex >= len(e.workflow.Steps) && e.stepIndex > 0 {
					e.stepIndex--
				}
			}
		case "up", "k":
			if e.stepIndex > 0 {
				e.stepIndex--
			}
		case "down", "j":
			if e.stepIndex < len(e.workflow.Steps)-1 {
				e.stepIndex++
			}
		}
	case tui.StepCreatedMsg:
		e.workflow.Steps = append(e.workflow.Steps, msg.Step)
		return e, nil
	}

	m, cmd := e.form.Update(msg)
	e.form = m.(*huh.Form)
	return e, cmd
}

func (e *WorkflowEditorScreen) openStepTypeForm() (tui.Screen, tea.Cmd) {
	// For simplicity, open commit form directly. Could be extended with a step type selector.
	form := forms.NewCommitForm(e.ctx, true, e)
	return form, form.Init()
}

func (e *WorkflowEditorScreen) save() {
	e.workflow.Name = e.name
	e.workflow.Description = e.desc
	_ = e.ctx.WorkflowStore.Save(*e.workflow)
}

// View renders the editor.
func (e *WorkflowEditorScreen) View() tea.View {
	var lines []string
	lines = append(lines, e.ctx.Styles.Title.Render(" Workflow Editor "))
	lines = append(lines, "")

	lines = append(lines, e.form.View())
	lines = append(lines, "")

	if len(e.workflow.Steps) > 0 {
		lines = append(lines, e.ctx.Styles.Info.Render("Steps:"))
		for i, step := range e.workflow.Steps {
			prefix := "  "
			if i == e.stepIndex {
				prefix = e.ctx.Styles.Key.Render("> ")
			}
			lines = append(lines, fmt.Sprintf("%s%s - %s", prefix, step.Type.String(), step.Description))
		}
	} else {
		lines = append(lines, e.ctx.Styles.Help.Render("No steps yet"))
	}

	lines = append(lines, "")
	lines = append(lines, e.ctx.Styles.Help.Render("a: add step | d: delete step | ↑/↓: select | s: save | esc: back"))

	return tea.NewView(e.ctx.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
}
