package forms

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
)

// NewBranchForm handles new branch creation.
type NewBranchForm struct {
	ctx      *tui.AppContext
	form     *huh.Form
	name     string
	base     string
	stepMode bool
	returnTo tui.Screen
}

// NewNewBranchForm creates a new branch form.
func NewNewBranchForm(ctx *tui.AppContext, stepMode bool, returnTo ...tui.Screen) *NewBranchForm {
	f := &NewBranchForm{ctx: ctx, stepMode: stepMode}
	if len(returnTo) > 0 {
		f.returnTo = returnTo[0]
	}

	branches, _ := ctx.GitService.GetBranches()
	branchOptions := []huh.Option[string]{huh.NewOption("(current branch)", "")}
	for _, b := range branches {
		branchOptions = append(branchOptions, huh.NewOption(b, b))
	}

	f.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Branch name").
				Placeholder("feature/name").
				Value(&f.name),
			huh.NewSelect[string]().
				Title("Base branch (optional)").
				Options(branchOptions...).
				Value(&f.base),
		),
	)
	return f
}

// Init initializes the form.
func (f *NewBranchForm) Init() tea.Cmd {
	return f.form.Init()
}

// Update handles messages.
func (f *NewBranchForm) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
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
			cmd := "git checkout -b " + f.name
			if f.base != "" {
				cmd += " " + f.base
			}
			return tui.OperationRequestMsg{
				Title:   "New Branch",
				Command: cmd,
				Run: func() error {
					return f.ctx.GitService.CreateBranch(f.name, f.base)
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
func (f *NewBranchForm) View() tea.View {
	return tea.NewView(f.form.View())
}

// BuildStep builds a workflow step.
func (f *NewBranchForm) BuildStep() models.WorkflowStep {
	params := map[string]string{"branch": f.name}
	if f.base != "" {
		params["base"] = f.base
	}
	return models.WorkflowStep{
		Type:        models.StepNewBranch,
		Parameters:  params,
		Description: "Create branch " + f.name,
	}
}
