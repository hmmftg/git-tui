package screens

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
	"gitflow-tui/internal/workflow"
)

// WorkflowListScreen displays and manages workflows.
type WorkflowListScreen struct {
	ctx       *tui.AppContext
	form      *huh.Form
	workflows []*models.Workflow
	selected  *models.Workflow
	message   string
}

// NewWorkflowListScreen creates a new workflow list screen.
func NewWorkflowListScreen(ctx *tui.AppContext) *WorkflowListScreen {
	w := &WorkflowListScreen{ctx: ctx}
	w.loadWorkflows()
	w.buildForm()
	return w
}

func (w *WorkflowListScreen) loadWorkflows() {
	workflows, _ := w.ctx.WorkflowStore.List()
	if len(workflows) == 0 {
		// Seed with templates
		templates := workflow.DefaultTemplates()
		for _, t := range templates {
			_ = w.ctx.WorkflowStore.Save(t)
		}
		workflows, _ = w.ctx.WorkflowStore.List()
	}
	w.workflows = make([]*models.Workflow, len(workflows))
	for i := range workflows {
		w.workflows[i] = &workflows[i]
	}
}

func (w *WorkflowListScreen) buildForm() {
	var selected *models.Workflow
	options := []huh.Option[*models.Workflow]{}
	for _, wf := range w.workflows {
		label := fmt.Sprintf("%s - %s", wf.Name, wf.Description)
		options = append(options, huh.NewOption(label, wf))
	}
	options = append(options, huh.NewOption[*models.Workflow]("+ Create New Workflow", nil))

	w.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*models.Workflow]().
				Title("Workflows").
				Description("Select a workflow to run, edit, or manage").
				Options(options...).
				Value(&selected),
		),
	)
}

// Init initializes the screen.
func (w *WorkflowListScreen) Init() tea.Cmd {
	return w.form.Init()
}

// Update handles messages.
func (w *WorkflowListScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return NewHomeScreen(w.ctx), nil
		}
	case tui.WorkflowSavedMsg:
		w.loadWorkflows()
		w.buildForm()
		w.message = fmt.Sprintf("Saved: %s", msg.Workflow.Name)
		return w, w.form.Init()
	}

	m, cmd := w.form.Update(msg)
	w.form = m.(*huh.Form)

	if w.form.State == huh.StateCompleted {
		// Huh select value binding happens after completion
		// We need to read the selected value from the form options
		// Since the value pointer was set in buildForm, we need a different approach
		// For now, use key-based actions after selection
		return w, nil
	}

	return w, cmd
}

// View renders the screen.
func (w *WorkflowListScreen) View() tea.View {
	var lines []string
	lines = append(lines, w.ctx.Styles.Title.Render(" Workflow Manager "))
	lines = append(lines, "")
	lines = append(lines, w.form.View())
	if w.message != "" {
		lines = append(lines, "")
		lines = append(lines, w.ctx.Styles.Success.Render(w.message))
	}
	lines = append(lines, "")
	lines = append(lines, w.ctx.Styles.Help.Render("esc: back | r: run selected | e: edit | d: delete | c: copy | n: new"))
	return tea.NewView(w.ctx.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
}

// RunWorkflow executes the selected workflow.
func (w *WorkflowListScreen) RunWorkflow() tui.Screen {
	if w.selected == nil {
		return w
	}
	return NewExecutionScreen(w.ctx, *w.selected)
}

// EditWorkflow opens the editor for the selected workflow.
func (w *WorkflowListScreen) EditWorkflow() tui.Screen {
	if w.selected == nil {
		return w
	}
	return NewWorkflowEditorScreen(w.ctx, w.selected)
}

// DeleteWorkflow removes the selected workflow.
func (w *WorkflowListScreen) DeleteWorkflow() tea.Cmd {
	if w.selected == nil || w.selected.ID == "" {
		return nil
	}
	_ = w.ctx.WorkflowStore.Delete(w.selected.ID)
	w.selected = nil
	w.loadWorkflows()
	w.buildForm()
	w.message = "Workflow deleted"
	return w.form.Init()
}

// DuplicateWorkflow creates a copy of the selected workflow.
func (w *WorkflowListScreen) DuplicateWorkflow() tea.Cmd {
	if w.selected == nil {
		return nil
	}
	copy := models.Workflow{
		ID:          uuid.New().String(),
		Name:        w.selected.Name + " Copy",
		Description: w.selected.Description,
		Steps:       make([]models.WorkflowStep, len(w.selected.Steps)),
		CreatedAt:   time.Now(),
	}
	for i, step := range w.selected.Steps {
		copy.Steps[i] = step
		if step.Parameters != nil {
			copy.Steps[i].Parameters = make(map[string]string)
			for k, v := range step.Parameters {
				copy.Steps[i].Parameters[k] = v
			}
		}
	}
	_ = w.ctx.WorkflowStore.Save(copy)
	w.loadWorkflows()
	w.buildForm()
	w.message = fmt.Sprintf("Duplicated: %s", copy.Name)
	return w.form.Init()
}
