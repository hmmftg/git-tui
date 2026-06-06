package workflow

import (
	"time"

	"github.com/google/uuid"

	"gitflow-tui/internal/models"
)

// DefaultTemplates returns built-in workflow templates.
func DefaultTemplates() []models.Workflow {
	now := time.Now()
	return []models.Workflow{
		{
			ID:          uuid.New().String(),
			Name:        "Quick Commit",
			Description: "Stage, commit and push all changes",
			CreatedAt:   now,
			Steps: []models.WorkflowStep{
				{Type: models.StepStatus, Description: "Check repository status"},
				{Type: models.StepCommit, Description: "Commit all changes", Parameters: map[string]string{"autoAdd": "true", "message": "WIP: auto commit"}},
				{Type: models.StepPush, Description: "Push to remote"},
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Release",
			Description: "Prepare a release: commit, push, merge to main",
			CreatedAt:   now,
			Steps: []models.WorkflowStep{
				{Type: models.StepStatus, Description: "Check repository status"},
				{Type: models.StepCommit, Description: "Commit changes", Parameters: map[string]string{"message": "chore: prepare release"}},
				{Type: models.StepPush, Description: "Push to remote"},
				{Type: models.StepMerge, Description: "Merge to main", Parameters: map[string]string{"target": "main"}},
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Hotfix",
			Description: "Quick fix: commit, push, merge to main and develop",
			CreatedAt:   now,
			Steps: []models.WorkflowStep{
				{Type: models.StepStatus, Description: "Check repository status"},
				{Type: models.StepCommit, Description: "Commit hotfix", Parameters: map[string]string{"message": "fix: hotfix"}},
				{Type: models.StepPush, Description: "Push to remote"},
				{Type: models.StepMerge, Description: "Merge to main", Parameters: map[string]string{"target": "main"}},
				{Type: models.StepCheckout, Description: "Switch to develop", Parameters: map[string]string{"branch": "develop"}},
				{Type: models.StepMerge, Description: "Merge hotfix to develop", Parameters: map[string]string{"target": "develop"}},
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Sync",
			Description: "Pull latest changes from remote",
			CreatedAt:   now,
			Steps: []models.WorkflowStep{
				{Type: models.StepPull, Description: "Pull from remote"},
				{Type: models.StepStatus, Description: "Check current status"},
			},
		},
	}
}
