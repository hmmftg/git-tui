package components

import (
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"gitflow-tui/internal/tui/styles"
)

// AsyncExecutor provides standardized async operation execution.
type AsyncExecutor struct {
	spinner  spinner.Model
	running  bool
	done     bool
	err      error
	output   []string
	maxLines int
	command  string // The command being executed
}

// NewAsyncExecutor creates a new async executor.
func NewAsyncExecutor(s styles.Styles) *AsyncExecutor {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(s.Success.GetForeground())
	return &AsyncExecutor{
		spinner:  sp,
		output:   []string{},
		maxLines: 20,
	}
}

// Start begins the async operation.
func (a *AsyncExecutor) Start() tea.Cmd {
	a.running = true
	a.done = false
	a.err = nil
	a.output = []string{}
	return a.spinner.Tick
}

// SetCommand sets the command being executed.
func (a *AsyncExecutor) SetCommand(command string) {
	a.command = command
}

// Stop marks the operation as successfully completed.
func (a *AsyncExecutor) Stop() {
	a.running = false
	a.done = true
	a.err = nil
}

// Fail marks the operation as failed.
func (a *AsyncExecutor) Fail(err error) {
	a.running = false
	a.done = true
	a.err = err
	a.AddOutput(fmt.Sprintf("✘ Error: %v", err))
}

// IsRunning returns true if the operation is running.
func (a *AsyncExecutor) IsRunning() bool { return a.running }

// IsDone returns true if the operation has completed.
func (a *AsyncExecutor) IsDone() bool { return a.done }

// HasError returns true if the operation failed.
func (a *AsyncExecutor) HasError() bool { return a.err != nil }

// GetError returns the error if one occurred.
func (a *AsyncExecutor) GetError() error { return a.err }

// AddOutput adds a line of output.
func (a *AsyncExecutor) AddOutput(line string) {
	a.output = append(a.output, line)
	if len(a.output) > a.maxLines {
		a.output = a.output[len(a.output)-a.maxLines:]
	}
}

// GetOutput returns all output lines.
func (a *AsyncExecutor) GetOutput() []string { return a.output }

// Update handles spinner ticks.
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

// View renders the executor state.
func (a *AsyncExecutor) View(s styles.Styles, title string) string {
	var lines []string
	if a.running {
		statusLine := fmt.Sprintf("%s %s", a.spinner.View(), title)
		lines = append(lines, statusLine)
		if a.command != "" {
			lines = append(lines, s.Dimmed.Render(fmt.Sprintf("  Executing: %s", a.command)))
		}
	} else if a.done {
		if a.err != nil {
			lines = append(lines, s.Error.Render(fmt.Sprintf("✘ %s failed", title)))
		} else {
			lines = append(lines, s.Success.Render(fmt.Sprintf("✔ %s completed", title)))
		}
		if a.command != "" {
			lines = append(lines, s.Dimmed.Render(fmt.Sprintf("  Command: %s", a.command)))
		}
	} else {
		// Initial state before running
		lines = append(lines, fmt.Sprintf("Preparing %s...", title))
		if a.command != "" {
			lines = append(lines, s.Dimmed.Render(fmt.Sprintf("  Command: %s", a.command)))
		}
	}
	if len(a.output) > 0 {
		lines = append(lines, "")
		lines = append(lines, s.Info.Render("Output:"))
		for _, line := range a.output {
			lines = append(lines, "  "+line)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
