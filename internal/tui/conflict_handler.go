package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConflictHandler manages conflict state and UI for merge/rebase operations.
// Embed this struct in models that support conflict resolution.
type ConflictHandler struct {
	conflict     bool
	targetBranch string
	operation    string // "merge" or "rebase"
}

// NewConflictHandler creates a new conflict handler for the specified operation
func NewConflictHandler(operation string) *ConflictHandler {
	return &ConflictHandler{
		operation: operation,
	}
}

// SetConflict marks that a conflict has occurred with the target branch
func (c *ConflictHandler) SetConflict(targetBranch string) {
	c.conflict = true
	c.targetBranch = targetBranch
}

// ClearConflict marks conflicts as resolved
func (c *ConflictHandler) ClearConflict() {
	c.conflict = false
}

// HasConflict returns true if there's an active conflict
func (c *ConflictHandler) HasConflict() bool {
	return c.conflict
}

// GetTargetBranch returns the branch that caused the conflict
func (c *ConflictHandler) GetTargetBranch() string {
	return c.targetBranch
}

// GetOperation returns the operation type ("merge" or "rebase")
func (c *ConflictHandler) GetOperation() string {
	return c.operation
}

// ViewBanner renders the conflict warning banner
func (c *ConflictHandler) ViewBanner(styles Styles) string {
	if !c.conflict {
		return ""
	}

	upperOp := string([]byte{c.operation[0] - 32}) + c.operation[1:]
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Warning.Render(fmt.Sprintf(" ⚠ %s CONFLICTS DETECTED ", upperOp)),
		styles.Warning.Render("Please resolve conflicts in your editor, then:"),
	)
}

// ViewMessage renders the conflict message with appropriate styling
func (c *ConflictHandler) ViewMessage(styles Styles, baseMsg string) string {
	if !c.conflict {
		return ""
	}
	return styles.Warning.Render(baseMsg)
}

// GetHelpText returns the appropriate help text based on conflict state
func (c *ConflictHandler) GetHelpText(normalHelp string) string {
	if c.conflict {
		return "r: resolve | c: continue | a: abort | esc: back"
	}
	return normalHelp
}

// GetConflictMessage returns a formatted conflict message for the target branch
func (c *ConflictHandler) GetConflictMessage(targetBranch string) string {
	return fmt.Sprintf("⚠ %s of %s has conflicts! Resolve them and press 'c' to continue.", c.operation, targetBranch)
}

// GetResolvedMessage returns a success message for resolved conflicts
func (c *ConflictHandler) GetResolvedMessage() string {
	upperOp := string([]byte{c.operation[0] - 32}) + c.operation[1:]
	return fmt.Sprintf("✔ %s completed successfully!", upperOp)
}

// GetConflictErrorMessage returns an error message when conflicts are detected
func (c *ConflictHandler) GetConflictErrorMessage(err error) string {
	return fmt.Sprintf("⚠ %s conflict! Resolve files and press 'c' to continue.\nError: %v", c.operation, err)
}

// GetStillConflictMessage returns a message when conflicts still exist
func (c *ConflictHandler) GetStillConflictMessage() string {
	return "⚠ Still have conflicts! Resolve them before continuing."
}

// HandleConflictKeys processes conflict-related key inputs.
// Returns (handled, cmd) where handled is true if the key was processed.
func (c *ConflictHandler) HandleConflictKeys(key string, hasConflicts bool) (bool, tea.Cmd) {
	if !hasConflicts {
		return false, nil
	}

	switch key {
	case "r":
		// Open conflict resolver
		return true, func() tea.Msg { return OpenConflictResolverMsg{} }
	case "c":
		// Continue operation - subclasses implement this
		return true, nil
	case "a":
		// Abort operation - subclasses implement this
		return true, nil
	}

	return false, nil
}

// Reset clears the conflict state (call when entering configure mode)
func (c *ConflictHandler) Reset() {
	c.conflict = false
	c.targetBranch = ""
}
