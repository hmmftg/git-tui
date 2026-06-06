package forms

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
)

// MergeForm handles merge configuration.
type MergeForm struct {
	ctx      *tui.AppContext
	form     *huh.Form
	branch   string
	noFF     bool
	squash   bool
	stepMode bool
	returnTo tui.Screen
}

// NewMergeForm creates a new merge form.
func NewMergeForm(ctx *tui.AppContext, stepMode bool, returnTo ...tui.Screen) *MergeForm {
	f := &MergeForm{ctx: ctx, stepMode: stepMode}
	if len(returnTo) > 0 {
		f.returnTo = returnTo[0]
	}

	branches, _ := ctx.GitService.GetBranches()
	branchOptions := make([]huh.Option[string], 0, len(branches))
	for _, b := range branches {
		branchOptions = append(branchOptions, huh.NewOption(b, b))
	}

	f.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Source branch").
				Options(branchOptions...).
				Value(&f.branch),
			huh.NewConfirm().
				Title("No fast-forward?").
				Value(&f.noFF),
			huh.NewConfirm().
				Title("Squash?").
				Value(&f.squash),
		),
	)
	return f
}

// Init initializes the form.
func (f *MergeForm) Init() tea.Cmd {
	return f.form.Init()
}

// Update handles messages.
func (f *MergeForm) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
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
				Title: "Merge",
				Run: func() error {
					return f.ctx.GitService.Merge(f.branch)
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
func (f *MergeForm) View() tea.View {
	return tea.NewView(f.form.View())
}

// BuildStep builds a workflow step.
func (f *MergeForm) BuildStep() models.WorkflowStep {
	params := map[string]string{"target": f.branch}
	if f.noFF {
		params["noFF"] = "true"
	}
	if f.squash {
		params["squash"] = "true"
	}
	return models.WorkflowStep{
		Type:        models.StepMerge,
		Parameters:  params,
		Description: "Merge " + f.branch,
	}
}
