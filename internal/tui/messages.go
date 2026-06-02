package tui

import (
	"fmt"

	"gitflow-tui/internal/models"
)

// SuccessMsg is a generic success message with typed payload
type SuccessMsg[T any] struct {
	Data    T
	Message string
}

// ErrorMsg is a generic error message with context
type ErrorMsg struct {
	Err     error
	Context string
}

func (e ErrorMsg) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s: %v", e.Context, e.Err)
	}
	return e.Err.Error()
}

// Operation lifecycle messages
type OperationStartedMsg struct {
	Operation string
}

type OperationCompletedMsg struct {
	Operation string
	Success   bool
	Message   string
}

// Branch-related messages
type BranchesLoadedMsg []string

// Status-related messages
type StatusLoadedMsg struct {
	Branch     string
	IsClean    bool
	Modified   []string
	Added      []string
	Deleted    []string
	Untracked  []string
	Conflicted []string
	Ahead      int
	Behind     int
}

// Command configuration messages
type CommandConfiguredMsg struct {
	Config models.CommandConfig
}

type CommandCancelledMsg struct{}

// Commit messages
type CommitSuccessMsg struct{}
type CommitErrorMsg error
type AddSuccessMsg struct{}

// Checkout messages
type CheckoutSuccessMsg string

// Push messages
type PushSuccessMsg struct{}
type PushErrorMsg error

// Pull messages
type PullSuccessMsg struct{}
type PullErrorMsg error

// Merge messages
type MergeSuccessMsg string
type MergeContinueMsg struct{}
type MergeAbortedMsg struct{}

// Rebase messages
type RebaseSuccessMsg string
type RebaseContinueMsg struct{}
type RebaseAbortedMsg struct{}

// Workflow messages
type SaveWorkflowMsg models.Workflow
type CancelEditMsg struct{}

// Conflict messages
type OpenConflictResolverMsg struct{}
type ConflictsResolvedMsg struct{}
type ConflictsAbortedMsg struct{}
type ConflictsCancelledMsg struct{}

// View navigation
type ViewChangeMsg models.ViewType
type RepoInfoMsg models.RepositoryInfo

// Simple string messages
type ErrorMsgString string
type SuccessMsgString string
