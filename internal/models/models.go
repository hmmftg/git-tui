package models

import (
	"time"
)

// ViewType represents the current application view
type ViewType int

const (
	ViewHome ViewType = iota
	ViewStatus
	ViewCommit
	ViewWorkflowBuilder
	ViewWorkflowEditor
	ViewExecution
	ViewMerge
	ViewRebase
	ViewSettings
	ViewCheckout
	ViewPull
	ViewPush
	ViewConflictResolver
)

// Status represents the parsed git status
type Status struct {
	Branch             string
	Ahead              int
	Behind             int
	Modified           []string
	Added              []string
	Deleted            []string
	Untracked          []string
	Renamed            []string
	Conflicted         []string
	IsClean            bool
	IsMergeInProgress  bool
	IsRebaseInProgress bool
}

// RepositoryInfo holds information about the current repository
type RepositoryInfo struct {
	Path       string
	IsValid    bool
	Status     Status
	LastUpdate time.Time
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Type        StepType
	Parameters  map[string]string
	Description string
}

// StepType represents the type of workflow step
type StepType int

const (
	StepStatus StepType = iota
	StepCommit
	StepPush
	StepPull
	StepCheckout
	StepMerge
	StepRebase
)

func (s StepType) String() string {
	switch s {
	case StepStatus:
		return "Status"
	case StepCommit:
		return "Commit"
	case StepPush:
		return "Push"
	case StepPull:
		return "Pull"
	case StepCheckout:
		return "Checkout"
	case StepMerge:
		return "Merge"
	case StepRebase:
		return "Rebase"
	default:
		return "Unknown"
	}
}

// Workflow represents a sequence of git operations
type Workflow struct {
	ID          string
	Name        string
	Description string
	Steps       []WorkflowStep
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ExecutionMode defines how to handle errors during workflow execution
type ExecutionMode int

const (
	StopOnError ExecutionMode = iota
	ContinueOnError
	AskUser
)

// ExecutionRecord tracks a workflow execution
type ExecutionRecord struct {
	ID          string
	WorkflowID  string
	StartedAt   time.Time
	CompletedAt *time.Time
	Status      ExecutionStatus
	Steps       []StepExecution
	Output      []string
}

// ExecutionStatus represents the state of a workflow execution
type ExecutionStatus int

const (
	ExecutionPending ExecutionStatus = iota
	ExecutionRunning
	ExecutionCompleted
	ExecutionFailed
	ExecutionCancelled
)

// StepExecution tracks the execution of a single step
type StepExecution struct {
	Step        WorkflowStep
	Status      ExecutionStatus
	StartedAt   time.Time
	CompletedAt *time.Time
	Output      []string
	Error       string
}

// StepState represents a step's state in the execution dashboard
type StepState int

const (
	StepPending StepState = iota
	StepRunning
	StepSuccess
	StepFailed
	StepSkipped
)

// CommandConfig represents the configuration data for a workflow step command
type CommandConfig struct {
	StepType   StepType
	Parameters map[string]string
}
