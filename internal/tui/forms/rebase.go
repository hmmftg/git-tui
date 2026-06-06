package forms

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
)

// RebaseForm handles rebase configuration.
type RebaseForm struct {
	ctx         *tui.AppContext
	form        *huh.Form
	branch      string
	interactive bool
	stepMode    bool
	returnTo    tui.Screen
}

// NewRebaseForm creates a new rebase form.
func NewRebaseForm(ctx *tui.AppContext, stepMode bool, returnTo ...tui.Screen) *RebaseForm {
	f := &RebaseForm{ctx: ctx, stepMode: stepMode}
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
				Title("Target branch").
				Options(branchOptions...).
				Value(&f.branch),
			huh.NewConfirm().
				Title("Interactive?").
				Value(&f.interactive),
		),
	)
	return f
}

// Init initializes the form.
func (f *RebaseForm) Init() tea.Cmd {
	return f.form.Init()
}

// Update handles messages.
func (f *RebaseForm) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
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
			cmd := "git rebase " + f.branch
			if f.interactive {
				cmd = "git rebase -i " + f.branch
			}
			return tui.OperationRequestMsg{
				Title:   "Rebase",
				Command: cmd,
				Run: func() error {
					return f.ctx.GitService.Rebase(f.branch)
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
func (f *RebaseForm) View() tea.View {
	return tea.NewView(f.form.View())
}

// BuildStep builds a workflow step.
func (f *RebaseForm) BuildStep() models.WorkflowStep {
	params := map[string]string{"target": f.branch}
	if f.interactive {
		params["interactive"] = "true"
	}
	return models.WorkflowStep{
		Type:        models.StepRebase,
		Parameters:  params,
		Description: "Rebase onto " + f.branch,
	}
}
