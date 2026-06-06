package forms

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
)

// CommitForm handles commit configuration and execution.
type CommitForm struct {
	ctx      *tui.AppContext
	form     *huh.Form
	subject  string
	body     string
	autoAdd  bool
	stepMode bool
	returnTo tui.Screen
}

// NewCommitForm creates a new commit form.
func NewCommitForm(ctx *tui.AppContext, stepMode bool, returnTo ...tui.Screen) *CommitForm {
	f := &CommitForm{ctx: ctx, stepMode: stepMode}
	if len(returnTo) > 0 {
		f.returnTo = returnTo[0]
	}
	f.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Commit Subject").
				Placeholder("Enter commit subject...").
				Value(&f.subject),
			huh.NewText().
				Title("Commit Body").
				Placeholder("Optional commit body...").
				Lines(5).
				Value(&f.body),
			huh.NewConfirm().
				Title("Auto-add all changes?").
				Value(&f.autoAdd),
		),
	)
	return f
}

// Init initializes the form.
func (f *CommitForm) Init() tea.Cmd {
	return f.form.Init()
}

// Update handles messages.
func (f *CommitForm) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	m, cmd := f.form.Update(msg)
	f.form = m.(*huh.Form)

	switch f.form.State {
	case huh.StateCompleted:
		message := strings.TrimSpace(f.subject)
		body := strings.TrimSpace(f.body)
		if body != "" && message != "" {
			message = message + "\n\n" + body
		} else if body != "" {
			message = body
		}
		if f.stepMode && f.returnTo != nil {
			return f.returnTo, func() tea.Msg {
				return tui.StepCreatedMsg{Step: f.BuildStep()}
			}
		}
		return f, func() tea.Msg {
			return tui.OperationRequestMsg{
				Title: "Commit",
				Command: func() string {
					cmd := "git commit -m \"" + message + "\""
					if f.autoAdd {
						cmd = "git add . && " + cmd
					}
					return cmd
				}(),
				Run: func() error {
					if f.autoAdd {
						if err := f.ctx.GitService.AddAll(); err != nil {
							return err
						}
					}
					return f.ctx.GitService.Commit(message)
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
func (f *CommitForm) View() tea.View {
	return tea.NewView(f.form.View())
}

// BuildStep builds a workflow step from the form values.
func (f *CommitForm) BuildStep() models.WorkflowStep {
	message := strings.TrimSpace(f.subject)
	body := strings.TrimSpace(f.body)
	if body != "" && message != "" {
		message = message + "\n\n" + body
	} else if body != "" {
		message = body
	}
	params := map[string]string{"message": message}
	if f.autoAdd {
		params["autoAdd"] = "true"
	}
	return models.WorkflowStep{
		Type:        models.StepCommit,
		Parameters:  params,
		Description: "Commit: " + f.subject,
	}
}
