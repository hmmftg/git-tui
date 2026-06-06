package forms

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
)

// CheckoutForm handles checkout/branch creation.
type CheckoutForm struct {
	ctx       *tui.AppContext
	form      *huh.Form
	branch    string
	createNew bool
	newName   string
	stepMode  bool
	returnTo  tui.Screen
}

// NewCheckoutForm creates a new checkout form.
func NewCheckoutForm(ctx *tui.AppContext, stepMode bool, returnTo ...tui.Screen) *CheckoutForm {
	f := &CheckoutForm{ctx: ctx, stepMode: stepMode}
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
				Title("Select branch").
				Options(branchOptions...).
				Value(&f.branch),
			huh.NewConfirm().
				Title("Create new branch?").
				Value(&f.createNew),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("New branch name").
				Placeholder("feature/name").
				Value(&f.newName),
		).WithHideFunc(func() bool { return !f.createNew }),
	)
	return f
}

// Init initializes the form.
func (f *CheckoutForm) Init() tea.Cmd {
	return f.form.Init()
}

// Update handles messages.
func (f *CheckoutForm) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
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
				Title: "Checkout",
				Run: func() error {
					if f.createNew {
						return f.ctx.GitService.CreateBranch(f.newName, f.branch)
					}
					return f.ctx.GitService.Checkout(f.branch)
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
func (f *CheckoutForm) View() tea.View {
	return tea.NewView(f.form.View())
}

// BuildStep builds a workflow step.
func (f *CheckoutForm) BuildStep() models.WorkflowStep {
	params := map[string]string{"branch": f.branch}
	if f.createNew {
		params["create"] = "true"
		params["branch"] = f.newName
		if f.branch != "" {
			params["base"] = f.branch
		}
	}
	return models.WorkflowStep{
		Type:        models.StepCheckout,
		Parameters:  params,
		Description: "Checkout " + f.branch,
	}
}
