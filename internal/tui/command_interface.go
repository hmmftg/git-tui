package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"gitflow-tui/internal/models"
)

// CommandMode represents the operating mode of a command model
type CommandMode string

const (
	ModeExecute   CommandMode = "execute"   // Execute the command and show output
	ModeConfigure CommandMode = "configure" // Collect parameters without executing
)

// CommandConfig represents the configuration data returned by a command model
type CommandConfig struct {
	StepType   models.StepType
	Parameters map[string]string
}

// CommandModel defines the interface for all command models that can be used
// in both execution and configuration modes
type CommandModel interface {
	tea.Model

	// SetMode sets the operating mode of the command model
	SetMode(mode CommandMode)

	// GetMode returns the current operating mode
	GetMode() CommandMode

	// GetParameters returns the collected parameters (only valid in configure mode)
	GetParameters() map[string]string

	// Execute returns a command to execute the operation (only valid in execute mode)
	Execute() tea.Cmd

	// GetStepType returns the step type this command model represents
	GetStepType() models.StepType
}

// Message types for command configuration
type commandConfiguredMsg struct {
	Config CommandConfig
}

type commandCancelledMsg struct{}
