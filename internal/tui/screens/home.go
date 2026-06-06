package screens

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"gitflow-tui/internal/tui"
	"gitflow-tui/internal/tui/forms"
)

// HomeAction represents a menu action.
type HomeAction string

const (
	ActionStatus    HomeAction = "Status"
	ActionCommit    HomeAction = "Commit"
	ActionPush      HomeAction = "Push"
	ActionPull      HomeAction = "Pull"
	ActionCheckout  HomeAction = "Checkout"
	ActionMerge     HomeAction = "Merge"
	ActionRebase    HomeAction = "Rebase"
	ActionNewBranch HomeAction = "New Branch"
	ActionWorkflows HomeAction = "Workflows"
	ActionExit      HomeAction = "Exit"
)

// HomeScreen is the main menu screen using Huh.
type HomeScreen struct {
	ctx    *tui.AppContext
	form   *huh.Form
	action *HomeAction
}

// NewHomeScreen creates a new home screen.
func NewHomeScreen(ctx *tui.AppContext) *HomeScreen {
	var action HomeAction
	h := &HomeScreen{ctx: ctx, action: &action}
	h.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[HomeAction]().
				Title("GitFlow TUI").
				Description("Select an action").
				Options(
					huh.NewOption("📊 Status", ActionStatus),
					huh.NewOption("📝 Commit", ActionCommit),
					huh.NewOption("📤 Push", ActionPush),
					huh.NewOption("📥 Pull", ActionPull),
					huh.NewOption("🔀 Checkout", ActionCheckout),
					huh.NewOption("⛙ Merge", ActionMerge),
					huh.NewOption("🔄 Rebase", ActionRebase),
					huh.NewOption("🌿 New Branch", ActionNewBranch),
					huh.NewOption("⚡ Workflows", ActionWorkflows),
					huh.NewOption("❌ Exit", ActionExit),
				).
				Value(&action),
		),
	)
	return h
}

// Init initializes the screen.
func (h *HomeScreen) Init() tea.Cmd {
	return h.form.Init()
}

// Update handles messages.
func (h *HomeScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	m, cmd := h.form.Update(msg)
	h.form = m.(*huh.Form)

	if h.form.State == huh.StateCompleted {
		switch *h.action {
		case ActionStatus:
			return NewStatusScreen(h.ctx), nil
		case ActionCommit:
			return forms.NewCommitForm(h.ctx, false), nil
		case ActionPush:
			return forms.NewPushForm(h.ctx, false), nil
		case ActionPull:
			return forms.NewPullForm(h.ctx, false), nil
		case ActionCheckout:
			return forms.NewCheckoutForm(h.ctx, false), nil
		case ActionMerge:
			return forms.NewMergeForm(h.ctx, false), nil
		case ActionRebase:
			return forms.NewRebaseForm(h.ctx, false), nil
		case ActionNewBranch:
			return forms.NewNewBranchForm(h.ctx, false), nil
		case ActionWorkflows:
			return NewWorkflowListScreen(h.ctx), nil
		case ActionExit:
			return h, tea.Quit
		}
	}

	return h, cmd
}

// View renders the screen.
func (h *HomeScreen) View() tea.View {
	return tea.NewView(h.form.View())
}
