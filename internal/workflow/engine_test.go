package workflow

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
)

func TestNewEngine(t *testing.T) {
	mock := git.NewMockGitService()
	engine := NewEngine(mock, models.StopOnError)

	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if engine.mode != models.StopOnError {
		t.Error("Expected StopOnError mode")
	}
}

func TestEngine_Execute_StatusStep(t *testing.T) {
	mock := git.NewMockGitService()
	engine := NewEngine(mock, models.StopOnError)

	workflow := models.Workflow{
		ID:   uuid.New().String(),
		Name: "Test Workflow",
		Steps: []models.WorkflowStep{
			{
				Type:        models.StepStatus,
				Description: "Check status",
			},
		},
		CreatedAt: time.Now(),
	}

	record, err := engine.Execute(workflow)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if record.Status != models.ExecutionCompleted {
		t.Error("Expected workflow to complete")
	}

	if len(record.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(record.Steps))
	}
}

func TestEngine_Execute_StopOnError(t *testing.T) {
	mock := git.NewMockGitService()
	mock.ShouldFail["commit"] = true

	engine := NewEngine(mock, models.StopOnError)

	workflow := models.Workflow{
		ID:   uuid.New().String(),
		Name: "Test Workflow",
		Steps: []models.WorkflowStep{
			{
				Type:        models.StepStatus,
				Description: "Check status",
			},
			{
				Type:        models.StepCommit,
				Description: "Commit changes",
				Parameters: map[string]string{
					"message": "test commit",
				},
			},
		},
		CreatedAt: time.Now(),
	}

	record, err := engine.Execute(workflow)
	if err == nil {
		t.Error("Expected error from failed commit step")
	}

	if record.Status != models.ExecutionFailed {
		t.Error("Expected workflow to fail")
	}
}

func TestEngine_Execute_ContinueOnError(t *testing.T) {
	mock := git.NewMockGitService()
	mock.ShouldFail["commit"] = true

	engine := NewEngine(mock, models.ContinueOnError)

	workflow := models.Workflow{
		ID:   uuid.New().String(),
		Name: "Test Workflow",
		Steps: []models.WorkflowStep{
			{
				Type:        models.StepCommit,
				Description: "Commit changes (will fail)",
				Parameters: map[string]string{
					"message": "test commit",
				},
			},
			{
				Type:        models.StepStatus,
				Description: "Check status (should still run)",
			},
		},
		CreatedAt: time.Now(),
	}

	record, err := engine.Execute(workflow)
	// ContinueOnError means we don't return error
	if err != nil {
		t.Errorf("Expected no error with ContinueOnError mode, got %v", err)
	}

	if len(record.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(record.Steps))
	}
}

func TestEngine_ValidateWorkflow(t *testing.T) {
	mock := git.NewMockGitService()
	engine := NewEngine(mock, models.StopOnError)

	tests := []struct {
		name     string
		workflow models.Workflow
		wantErr  bool
		errCount int
	}{
		{
			name: "valid workflow",
			workflow: models.Workflow{
				Name:  "Valid",
				Steps: []models.WorkflowStep{{Type: models.StepStatus}},
			},
			wantErr:  false,
			errCount: 0,
		},
		{
			name: "empty name",
			workflow: models.Workflow{
				Name:  "",
				Steps: []models.WorkflowStep{{Type: models.StepStatus}},
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "no steps",
			workflow: models.Workflow{
				Name:  "No Steps",
				Steps: []models.WorkflowStep{},
			},
			wantErr:  true,
			errCount: 1,
		},
		{
			name: "push before commit",
			workflow: models.Workflow{
				Name: "Bad Order",
				Steps: []models.WorkflowStep{
					{Type: models.StepPush},
					{Type: models.StepCommit},
				},
			},
			wantErr:  true,
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := engine.ValidateWorkflow(tt.workflow)
			if tt.wantErr && len(errors) == 0 {
				t.Error("Expected validation errors but got none")
			}
			if !tt.wantErr && len(errors) > 0 {
				t.Errorf("Expected no errors but got: %v", errors)
			}
			if tt.errCount > 0 && len(errors) != tt.errCount {
				t.Errorf("Expected %d errors, got %d: %v", tt.errCount, len(errors), errors)
			}
		})
	}
}

func TestEngine_executeStep(t *testing.T) {
	mock := git.NewMockGitService()
	engine := NewEngine(mock, models.StopOnError)

	tests := []struct {
		name    string
		step    models.WorkflowStep
		wantErr bool
	}{
		{
			name:    "status step",
			step:    models.WorkflowStep{Type: models.StepStatus},
			wantErr: false,
		},
		{
			name: "commit step without message",
			step: models.WorkflowStep{
				Type:        models.StepCommit,
				Description: "Commit",
			},
			wantErr: false, // Will use default message
		},
		{
			name: "commit step with message",
			step: models.WorkflowStep{
				Type:        models.StepCommit,
				Description: "Commit",
				Parameters:  map[string]string{"message": "custom message"},
			},
			wantErr: false,
		},
		{
			name:    "pull step",
			step:    models.WorkflowStep{Type: models.StepPull},
			wantErr: false,
		},
		{
			name: "checkout without branch",
			step: models.WorkflowStep{
				Type:        models.StepCheckout,
				Description: "Checkout",
			},
			wantErr: true,
		},
		{
			name: "merge without target",
			step: models.WorkflowStep{
				Type:        models.StepMerge,
				Description: "Merge",
			},
			wantErr: true,
		},
		{
			name: "rebase without target",
			step: models.WorkflowStep{
				Type:        models.StepRebase,
				Description: "Rebase",
			},
			wantErr: true,
		},
		{
			name:    "unknown step type",
			step:    models.WorkflowStep{Type: models.StepType(999)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ExecuteStep(tt.step)
			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestEngine_ExecuteStep_TUIProducedMergeRebaseParameters(t *testing.T) {
	mock := git.NewMockGitService()
	engine := NewEngine(mock, models.StopOnError)

	if err := engine.ExecuteStep(models.WorkflowStep{Type: models.StepMerge, Parameters: map[string]string{"target": "develop"}}); err != nil {
		t.Fatalf("expected TUI-produced merge config to execute, got %v", err)
	}
	if err := engine.ExecuteStep(models.WorkflowStep{Type: models.StepRebase, Parameters: map[string]string{"target": "main"}}); err != nil {
		t.Fatalf("expected TUI-produced rebase config to execute, got %v", err)
	}
}

func TestEngine_ExecuteStep_PushPullOptions(t *testing.T) {
	mock := git.NewMockGitService()
	engine := NewEngine(mock, models.StopOnError)

	if err := engine.ExecuteStep(models.WorkflowStep{Type: models.StepPush, Parameters: map[string]string{
		"remote": "upstream",
		"branch": "feature/test",
		"force":  "true",
	}}); err != nil {
		t.Fatalf("expected push options to execute, got %v", err)
	}

	if err := engine.ExecuteStep(models.WorkflowStep{Type: models.StepPull, Parameters: map[string]string{
		"remote": "origin",
		"branch": "develop",
		"rebase": "true",
	}}); err != nil {
		t.Fatalf("expected pull options to execute, got %v", err)
	}

	commands := ""
	for _, command := range mock.Commands {
		commands += command + "\n"
	}
	if !strings.Contains(commands, "push --force upstream feature/test") {
		t.Fatalf("expected force push command, got:\n%s", commands)
	}
	if !strings.Contains(commands, "pull --rebase origin develop") {
		t.Fatalf("expected rebase pull command, got:\n%s", commands)
	}
}
