package forms

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
)

// PushForm handles push configuration.
type PushForm struct {
	ctx      *tui.AppContext
	form     *huh.Form
	remote   string
	branch   string
	force    bool
	stepMode bool
	returnTo tui.Screen
}

// NewPushForm creates a new push form.
func NewPushForm(ctx *tui.AppContext, stepMode bool, returnTo ...tui.Screen) *PushForm {
	f := &PushForm{ctx: ctx, stepMode: stepMode}
	if len(returnTo) > 0 {
		f.returnTo = returnTo[0]
	}
	f.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Remote").
				Placeholder("origin").
				Value(&f.remote),
			huh.NewInput().
				Title("Branch").
				Placeholder("leave empty for current branch").
				Value(&f.branch),
			huh.NewConfirm().
				Title("Force push?").
				Value(&f.force),
		),
	)
	return f
}

// Init initializes the form.
func (f *PushForm) Init() tea.Cmd {
	return f.form.Init()
}

// Update handles messages.
func (f *PushForm) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	m, cmd := f.form.Update(msg)
	f.form = m.(*huh.Form)

	switch f.form.State {
	case huh.StateCompleted:
		if f.stepMode && f.returnTo != nil {
			return f.returnTo, func() tea.Msg {
				return tui.StepCreatedMsg{Step: f.BuildStep()}
			}
		}
		return f, func() tea.Msg {
			return tui.OperationRequestMsg{
				Title: "Push",
				Run: func() error {
					return f.ctx.GitService.PushOptions(f.remote, f.branch, f.force)
				},
			}
		}
	case huh.StateAborted:
		if f.returnTo != nil {
			return f.returnTo, nil
		}
		return f, func() tea.Msg { return tui.HomeRequestMsg{} }
	}

	return f, cmd
}

// View renders the form.
func (f *PushForm) View() tea.View {
	return tea.NewView(f.form.View())
}

// BuildStep builds a workflow step.
func (f *PushForm) BuildStep() models.WorkflowStep {
	params := map[string]string{
		"remote": f.remote,
		"branch": f.branch,
	}
	if f.force {
		params["force"] = "true"
	}
	return models.WorkflowStep{
		Type:        models.StepPush,
		Parameters:  params,
		Description: "Push to " + f.remote,
	}
}
