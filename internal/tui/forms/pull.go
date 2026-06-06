package forms

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
)

// PullForm handles pull configuration.
type PullForm struct {
	ctx      *tui.AppContext
	form     *huh.Form
	remote   string
	branch   string
	rebase   bool
	stepMode bool
	returnTo tui.Screen
}

// NewPullForm creates a new pull form.
func NewPullForm(ctx *tui.AppContext, stepMode bool, returnTo ...tui.Screen) *PullForm {
	f := &PullForm{ctx: ctx, stepMode: stepMode}
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
				Title("Rebase instead of merge?").
				Value(&f.rebase),
		),
	)
	return f
}

// Init initializes the form.
func (f *PullForm) Init() tea.Cmd {
	return f.form.Init()
}

// Update handles messages.
func (f *PullForm) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
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
				Title: "Pull",
				Run: func() error {
					return f.ctx.GitService.PullOptions(f.remote, f.rebase, f.branch)
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
func (f *PullForm) View() tea.View {
	return tea.NewView(f.form.View())
}

// BuildStep builds a workflow step.
func (f *PullForm) BuildStep() models.WorkflowStep {
	params := map[string]string{
		"remote": f.remote,
		"branch": f.branch,
	}
	if f.rebase {
		params["rebase"] = "true"
	}
	return models.WorkflowStep{
		Type:        models.StepPull,
		Parameters:  params,
		Description: "Pull from " + f.remote,
	}
}
