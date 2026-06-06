package tui

import "gitflow-tui/internal/models"

// OperationRequestMsg requests the app to navigate to an operation screen.
type OperationRequestMsg struct {
	Title string
	Run   func() error
}

// HomeRequestMsg requests the app to navigate home.
type HomeRequestMsg struct{}

// OperationFinishedMsg signals an async operation completed.
type OperationFinishedMsg struct {
	Err error
}

// WorkflowSavedMsg signals a workflow was saved.
type WorkflowSavedMsg struct {
	Workflow models.Workflow
}

// StepCreatedMsg signals a workflow step was created or edited.
type StepCreatedMsg struct {
	Step models.WorkflowStep
}

// RepoInfoMsg carries refreshed repository information.
type RepoInfoMsg models.RepositoryInfo

// ErrorMsgString is a simple string error message.
type ErrorMsgString string

// SuccessMsgString is a simple string success message.
type SuccessMsgString string
