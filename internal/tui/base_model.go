package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// BaseModel provides common fields and helper methods for all command models.
// Embed this struct in your model to get standard functionality.
type BaseModel struct {
	GitSvc  git.GitService
	Styles  Styles
	Mode    CommandMode
	Message string
	Err     error
}

// NewBaseModel creates a new base model with default execute mode
func NewBaseModel(gitSvc git.GitService, styles Styles) BaseModel {
	return BaseModel{
		GitSvc: gitSvc,
		Styles: styles,
		Mode:   ModeExecute,
	}
}

// SetMode changes the operating mode and resets state for configure mode
func (b *BaseModel) SetMode(mode CommandMode) {
	b.Mode = mode
	if mode == ModeConfigure {
		b.ResetState()
	}
}

// GetMode returns the current operating mode
func (b *BaseModel) GetMode() CommandMode {
	return b.Mode
}

// IsConfigureMode returns true if the model is in configure mode
func (b *BaseModel) IsConfigureMode() bool {
	return b.Mode == ModeConfigure
}

// IsExecuteMode returns true if the model is in execute mode
func (b *BaseModel) IsExecuteMode() bool {
	return b.Mode == ModeExecute
}

// ResetState clears messages and errors (typically called when entering configure mode)
func (b *BaseModel) ResetState() {
	b.Message = ""
	b.Err = nil
}

// SetError sets an error message and returns the formatted error string
func (b *BaseModel) SetError(err error) string {
	b.Err = err
	b.Message = fmt.Sprintf("✘ Error: %v", err)
	return b.Message
}

// SetSuccess sets a success message and returns the formatted success string
func (b *BaseModel) SetSuccess(format string, args ...interface{}) string {
	b.Err = nil
	b.Message = fmt.Sprintf("✔ "+format, args...)
	return b.Message
}

// SetInfo sets an informational message
func (b *BaseModel) SetInfo(format string, args ...interface{}) string {
	b.Message = fmt.Sprintf(format, args...)
	return b.Message
}

// HasError returns true if there's an active error
func (b *BaseModel) HasError() bool {
	return b.Err != nil
}

// HasMessage returns true if there's a message to display
func (b *BaseModel) HasMessage() bool {
	return b.Message != ""
}

// GetMessageStyle returns the appropriate lipgloss style for the current message
func (b *BaseModel) GetMessageStyle() lipgloss.Style {
	if b.HasError() {
		return b.Styles.Error
	}
	if strings.Contains(b.Message, "✔") {
		return b.Styles.Success
	}
	return b.Styles.Info
}

// RenderMessage renders the current message with appropriate styling
func (b *BaseModel) RenderMessage() string {
	if !b.HasMessage() {
		return ""
	}
	return b.GetMessageStyle().Render(b.Message)
}

// QuitCmd returns a command to quit the application
func QuitCmd() tea.Cmd {
	return tea.Quit
}

// NavigationCmd creates a command that sends a view change message
func NavigationCmd(view ViewType) tea.Cmd {
	return func() tea.Msg {
		return ViewChangeMsg(view)
	}
}

// ViewType is an alias for models.ViewType to avoid import cycles
type ViewType int
