package tui

import (
	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// CommandRegistry manages the mapping between StepTypes and CommandModel creators
type CommandRegistry struct {
	gitSvc git.GitService
	styles Styles
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry(gitSvc git.GitService, styles Styles) *CommandRegistry {
	return &CommandRegistry{
		gitSvc: gitSvc,
		styles: styles,
	}
}

// CreateCommandModel creates a new CommandModel instance for the given StepType
func (r *CommandRegistry) CreateCommandModel(stepType models.StepType, mode CommandMode) CommandModel {
	switch stepType {
	case models.StepCommit:
		model := NewCommitModel(r.gitSvc, r.styles)
		model.SetMode(mode)
		return model
	case models.StepCheckout:
		model := NewCheckoutModel(r.gitSvc, r.styles)
		model.SetMode(mode)
		return model
	case models.StepStatus:
		model := NewStatusModel(r.gitSvc, r.styles)
		model.SetMode(mode)
		return model
	case models.StepPush:
		model := NewPushModel(r.gitSvc, r.styles)
		model.SetMode(mode)
		return model
	case models.StepPull:
		model := NewPullModel(r.gitSvc, r.styles)
		model.SetMode(mode)
		return model
	case models.StepMerge:
		model := NewMergeModel(r.gitSvc, r.styles)
		model.SetMode(mode)
		return model
	case models.StepRebase:
		model := NewRebaseModel(r.gitSvc, r.styles)
		model.SetMode(mode)
		return model
	default:
		return nil
	}
}

// GetAvailableStepTypes returns all available step types
func (r *CommandRegistry) GetAvailableStepTypes() []models.StepType {
	return []models.StepType{
		models.StepCommit,
		models.StepCheckout,
		models.StepStatus,
		models.StepPush,
		models.StepPull,
		models.StepMerge,
		models.StepRebase,
	}
}

// GetStepTypeName returns a human-readable name for a step type
func (r *CommandRegistry) GetStepTypeName(stepType models.StepType) string {
	return stepType.String()
}
