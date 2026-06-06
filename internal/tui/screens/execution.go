package screens

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"gitflow-tui/internal/models"
	"gitflow-tui/internal/tui"
	"gitflow-tui/internal/tui/components"
	"gitflow-tui/internal/tui/styles"
	"gitflow-tui/internal/workflow"
)

// ExecutionScreen runs a workflow step by step.
type ExecutionScreen struct {
	ctx         *tui.AppContext
	workflow    models.Workflow
	engine      *workflow.Engine
	currentStep int
	stepStates  []models.StepState
	executor    *components.AsyncExecutor
	done        bool
	failed      bool
	cancelled   bool
}

// NewExecutionScreen creates a new execution screen.
func NewExecutionScreen(ctx *tui.AppContext, wf models.Workflow) *ExecutionScreen {
	stepStates := make([]models.StepState, len(wf.Steps))
	for i := range stepStates {
		stepStates[i] = models.StepPending
	}
	return &ExecutionScreen{
		ctx:         ctx,
		workflow:    wf,
		engine:      workflow.NewEngine(ctx.GitService, models.StopOnError),
		currentStep: 0,
		stepStates:  stepStates,
		executor:    components.NewAsyncExecutor(ctx.Styles),
	}
}

// Init starts execution.
func (s *ExecutionScreen) Init() tea.Cmd {
	return s.runCurrentStep()
}

// Update handles messages.
func (s *ExecutionScreen) Update(msg tea.Msg) (tui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			if s.done || s.failed || s.cancelled {
				return NewHomeScreen(s.ctx), nil
			}
			s.cancelled = true
			return s, nil
		}
		if msg.String() == "r" && s.failed {
			s.failed = false
			return s, s.runCurrentStep()
		}
	case tui.OperationFinishedMsg:
		if s.cancelled {
			return s, nil
		}
		if msg.Err != nil {
			s.stepStates[s.currentStep] = models.StepFailed
			s.executor.Fail(msg.Err)
			s.failed = true
		} else {
			s.stepStates[s.currentStep] = models.StepSuccess
			s.executor.Stop()
			s.currentStep++
			if s.currentStep >= len(s.workflow.Steps) {
				s.done = true
			} else {
				return s, s.runCurrentStep()
			}
		}
		return s, nil
	}

	cmd := s.executor.Update(msg)
	return s, cmd
}

func (s *ExecutionScreen) runCurrentStep() tea.Cmd {
	if s.currentStep >= len(s.workflow.Steps) {
		s.done = true
		return nil
	}
	s.stepStates[s.currentStep] = models.StepRunning
	step := s.workflow.Steps[s.currentStep]
	return tea.Batch(
		s.executor.Start(),
		func() tea.Msg {
			err := s.engine.ExecuteStep(step)
			return tui.OperationFinishedMsg{Err: err}
		},
	)
}

// View renders the execution dashboard.
func (s *ExecutionScreen) View() tea.View {
	var lines []string
	lines = append(lines, s.ctx.Styles.Title.Render(" "+s.workflow.Name+" "))
	lines = append(lines, "")

	for i, step := range s.workflow.Steps {
		state := s.stepStates[i]
		indicator := styles.StatusIndicator(stepStateString(state))
		name := step.Type.String()
		if step.Description != "" {
			name = step.Description
		}
		lines = append(lines, fmt.Sprintf("%s %s", indicator, name))
	}

	lines = append(lines, "")
	if s.currentStep < len(s.workflow.Steps) {
		step := s.workflow.Steps[s.currentStep]
		lines = append(lines, s.executor.View(s.ctx.Styles, step.Type.String()))
	}

	if s.done {
		lines = append(lines, "")
		lines = append(lines, s.ctx.Styles.Success.Render("✔ Workflow completed successfully"))
	} else if s.failed {
		lines = append(lines, "")
		lines = append(lines, s.ctx.Styles.Error.Render("✘ Step failed"))
		lines = append(lines, s.ctx.Styles.Help.Render("r: retry | esc: back"))
	} else if s.cancelled {
		lines = append(lines, "")
		lines = append(lines, s.ctx.Styles.Warning.Render("⚠ Cancelled"))
		lines = append(lines, s.ctx.Styles.Help.Render("esc: back"))
	}

	return tea.NewView(s.ctx.Styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
}

func stepStateString(state models.StepState) string {
	switch state {
	case models.StepPending:
		return "pending"
	case models.StepRunning:
		return "running"
	case models.StepSuccess:
		return "success"
	case models.StepFailed:
		return "error"
	case models.StepSkipped:
		return "skipped"
	default:
		return "pending"
	}
}
