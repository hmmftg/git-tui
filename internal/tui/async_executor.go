package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AsyncExecutor provides standardized async operation execution with spinner and output tracking.
// Use this for operations like push, pull, and other long-running git commands.
type AsyncExecutor struct {
	spinner  spinner.Model
	running  bool
	done     bool
	err      error
	output   []string
	maxLines int
}

// NewAsyncExecutor creates a new async executor with the given styles
func NewAsyncExecutor(styles Styles) *AsyncExecutor {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Success.GetForeground())

	return &AsyncExecutor{
		spinner:  s,
		output:   []string{},
		maxLines: 20,
	}
}

// Start begins the async operation and returns the initial command (spinner tick)
func (a *AsyncExecutor) Start() tea.Cmd {
	a.running = true
	a.done = false
	a.err = nil
	a.output = []string{}
	return a.spinner.Tick
}

// Stop marks the operation as successfully completed
func (a *AsyncExecutor) Stop(successMsg string) {
	a.running = false
	a.done = true
	a.err = nil
	a.AddOutput(successMsg)
}

// StopWithError marks the operation as failed with an error
func (a *AsyncExecutor) StopWithError(err error) {
	a.running = false
	a.done = true
	a.err = err
	a.AddOutput(fmt.Sprintf("✘ Error: %v", err))
}

// IsRunning returns true if the operation is currently running
func (a *AsyncExecutor) IsRunning() bool {
	return a.running
}

// IsDone returns true if the operation has completed (success or error)
func (a *AsyncExecutor) IsDone() bool {
	return a.done
}

// HasError returns true if the operation failed
func (a *AsyncExecutor) HasError() bool {
	return a.err != nil
}

// GetError returns the error if one occurred
func (a *AsyncExecutor) GetError() error {
	return a.err
}

// AddOutput adds a line of output, maintaining maxLines limit
func (a *AsyncExecutor) AddOutput(line string) {
	a.output = append(a.output, line)
	if len(a.output) > a.maxLines {
		a.output = a.output[len(a.output)-a.maxLines:]
	}
}

// GetOutput returns all output lines
func (a *AsyncExecutor) GetOutput() []string {
	return a.output
}

// Reset clears the executor state for reuse
func (a *AsyncExecutor) Reset() {
	a.running = false
	a.done = false
	a.err = nil
	a.output = []string{}
}

// Update handles spinner ticks and other messages
func (a *AsyncExecutor) Update(msg tea.Msg) tea.Cmd {
	if !a.running {
		return nil
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return cmd
	}

	return nil
}

// ViewStatusLine renders the current status line (spinner + status text)
func (a *AsyncExecutor) ViewStatusLine(statusText string) string {
	if a.running {
		return fmt.Sprintf("%s %s", a.spinner.View(), statusText)
	}
	if a.done {
		return statusText
	}
	return ""
}

// ViewOutput renders the output section with styling
func (a *AsyncExecutor) ViewOutput(styles Styles, title string) string {
	if len(a.output) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, styles.Info.Render(title))
	for _, line := range a.output {
		lines = append(lines, "  "+line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// View renders the complete async executor view (status + output)
func (a *AsyncExecutor) View(styles Styles, runningText, doneText string) string {
	var lines []string

	// Status line
	if a.running {
		lines = append(lines, a.ViewStatusLine(runningText))
	} else if a.done {
		if a.err != nil {
			lines = append(lines, styles.Error.Render(doneText))
		} else {
			lines = append(lines, styles.Success.Render(doneText))
		}
	}

	lines = append(lines, "")

	// Output
	if output := a.ViewOutput(styles, "Output:"); output != "" {
		lines = append(lines, output)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// GetHelpText returns appropriate help text based on state
func (a *AsyncExecutor) GetHelpText(styles Styles, defaultHelp string) string {
	if a.done || a.err != nil {
		return "Press esc to go back"
	}
	return defaultHelp
}
