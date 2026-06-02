package workflow

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

// Engine executes workflows
type Engine struct {
	gitSvc git.GitService
	mode   models.ExecutionMode
}

// NewEngine creates a new workflow engine
func NewEngine(gitSvc git.GitService, mode models.ExecutionMode) *Engine {
	return &Engine{
		gitSvc: gitSvc,
		mode:   mode,
	}
}

// Execute runs a workflow and returns an execution record
func (e *Engine) Execute(workflow models.Workflow) (*models.ExecutionRecord, error) {
	record := &models.ExecutionRecord{
		ID:         uuid.New().String(),
		WorkflowID: workflow.ID,
		StartedAt:  time.Now(),
		Status:     models.ExecutionRunning,
		Steps:      make([]models.StepExecution, 0, len(workflow.Steps)),
		Output:     []string{},
	}

	for i, step := range workflow.Steps {
		record.Output = append(record.Output, fmt.Sprintf("Executing step %d: %s", i+1, step.Type.String()))

		stepExec := models.StepExecution{
			Step:      step,
			Status:    models.ExecutionRunning,
			StartedAt: time.Now(),
		}

		err := e.ExecuteStep(step)

		now := time.Now()
		stepExec.CompletedAt = &now

		if err != nil {
			stepExec.Status = models.ExecutionFailed
			stepExec.Error = err.Error()
			record.Status = models.ExecutionFailed
			record.Output = append(record.Output, fmt.Sprintf("Step %d failed: %v", i+1, err))

			// Handle error based on mode
			switch e.mode {
			case models.StopOnError:
				now := time.Now()
				record.CompletedAt = &now
				return record, err
			case models.ContinueOnError:
				// Continue to next step
				record.Output = append(record.Output, "Continuing despite error...")
			case models.AskUser:
				// For now, stop (in TUI, this would show a prompt)
				now := time.Now()
				record.CompletedAt = &now
				return record, err
			}
		} else {
			stepExec.Status = models.ExecutionCompleted
			record.Output = append(record.Output, fmt.Sprintf("Step %d completed successfully", i+1))
		}

		record.Steps = append(record.Steps, stepExec)
	}

	now := time.Now()
	record.CompletedAt = &now
	record.Status = models.ExecutionCompleted
	record.Output = append(record.Output, "Workflow completed successfully!")

	return record, nil
}

// ExecuteStep executes a single workflow step
func (e *Engine) ExecuteStep(step models.WorkflowStep) error {
	switch step.Type {
	case models.StepStatus:
		_, err := e.gitSvc.Status()
		return err

	case models.StepCommit:
		message := step.Parameters["message"]
		if message == "" {
			message = "Commit from workflow"
		}

		// Auto-add if specified
		if step.Parameters["autoAdd"] == "true" {
			if err := e.gitSvc.AddAll(); err != nil {
				return err
			}
		}

		return e.gitSvc.Commit(message)

	case models.StepPush:
		remote := step.Parameters["remote"]
		if remote == "" {
			remote = "origin"
		}

		branch := step.Parameters["branch"]
		if branch == "" {
			var err error
			branch, err = e.gitSvc.CurrentBranch()
			if err != nil {
				return err
			}
		}

		return e.gitSvc.PushOptions(remote, branch, step.Parameters["force"] == "true")

	case models.StepPull:
		return e.gitSvc.PullOptions(
			step.Parameters["remote"],
			step.Parameters["rebase"] == "true",
			step.Parameters["branch"],
		)

	case models.StepCheckout:
		branch := step.Parameters["branch"]
		if branch == "" {
			return errors.New("checkout step requires a branch parameter")
		}
		return e.gitSvc.Checkout(branch)

	case models.StepMerge:
		target := step.Parameters["target"]
		if target == "" {
			target = step.Parameters["targetBranch"]
		}
		if target == "" {
			return errors.New("merge step requires a target parameter")
		}
		return e.gitSvc.Merge(target)

	case models.StepRebase:
		target := step.Parameters["target"]
		if target == "" {
			target = step.Parameters["targetBranch"]
		}
		if target == "" {
			return errors.New("rebase step requires a target parameter")
		}
		return e.gitSvc.Rebase(target)

	default:
		return fmt.Errorf("unknown step type: %v", step.Type)
	}
}

// ValidateWorkflow checks if a workflow is valid
func (e *Engine) ValidateWorkflow(workflow models.Workflow) []string {
	var errors []string

	if workflow.Name == "" {
		errors = append(errors, "Workflow must have a name")
	}

	if len(workflow.Steps) == 0 {
		errors = append(errors, "Workflow must have at least one step")
	}

	// Check for common anti-patterns
	for i := 0; i < len(workflow.Steps)-1; i++ {
		current := workflow.Steps[i]
		next := workflow.Steps[i+1]

		// Push before commit
		if current.Type == models.StepPush && next.Type == models.StepCommit {
			errors = append(errors, fmt.Sprintf("Step %d: Push before commit may fail if there are no changes", i+1))
		}

		// Multiple consecutive pushes
		if current.Type == models.StepPush && next.Type == models.StepPush {
			errors = append(errors, fmt.Sprintf("Step %d: Consecutive push operations are redundant", i+1))
		}
	}

	return errors
}
