package tui

import (
	tea "charm.land/bubbletea/v2"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/tui/styles"
	"gitflow-tui/internal/workflow"
)

// Screen is the interface for all navigable screens.
type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() tea.View
}

// AppContext holds shared application state and services.
type AppContext struct {
	GitService    git.GitService
	WorkflowStore workflow.Store
	Styles        styles.Styles
	Width         int
	Height        int
}

// NewAppContext creates a new app context.
func NewAppContext(gitSvc git.GitService, store workflow.Store) *AppContext {
	return &AppContext{
		GitService:    gitSvc,
		WorkflowStore: store,
		Styles:        styles.DefaultStyles(),
	}
}
