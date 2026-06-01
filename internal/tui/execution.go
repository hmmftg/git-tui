package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/models"
	"gitflow-tui/internal/workflow"
)

// LogEntry represents a single log line with timestamp
type LogEntry struct {
	Timestamp time.Time
	Message   string
	Level     string // "info", "success", "error"
}

// ExecutionModel handles the workflow execution dashboard
type ExecutionModel struct {
	workflow models.Workflow
	engine   *workflow.Engine
	gitSvc   git.GitService
	styles   Styles

	currentStep int
	stepStates  []models.StepState
	logs        []LogEntry

	spinner         spinner.Model
	running         bool
	done            bool
	failed          bool
	continueOnError bool

	scrollPos  int
	logsHeight int
}

// NewExecutionModel creates a new execution model
func NewExecutionModel(gitSvc git.GitService, styles Styles, wf models.Workflow, mode models.ExecutionMode) *ExecutionModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Warning.GetForeground())

	// Initialize step states
	stepStates := make([]models.StepState, len(wf.Steps))
	for i := range stepStates {
		stepStates[i] = models.StepPending
	}

	// Default to StopOnError
	continueOnError := mode == models.ContinueOnError

	return &ExecutionModel{
		workflow:        wf,
		engine:          workflow.NewEngine(gitSvc, mode),
		gitSvc:          gitSvc,
		styles:          styles,
		currentStep:     0,
		stepStates:      stepStates,
		logs:            []LogEntry{},
		spinner:         s,
		running:         false,
		done:            false,
		failed:          false,
		continueOnError: continueOnError,
		logsHeight:      10,
	}
}

// Init initializes the model and starts execution
func (m *ExecutionModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startStep(0),
	)
}

// Update handles messages
func (m *ExecutionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scrollPos > 0 {
				m.scrollPos--
			}
		case "down", "j":
			maxScroll := len(m.logs) - m.logsHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollPos < maxScroll {
				m.scrollPos++
			}
		case "pgup":
			m.scrollPos -= m.logsHeight
			if m.scrollPos < 0 {
				m.scrollPos = 0
			}
		case "pgdown", " ":
			maxScroll := len(m.logs) - m.logsHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.scrollPos += m.logsHeight
			if m.scrollPos > maxScroll {
				m.scrollPos = maxScroll
			}
		case "r":
			if m.done || m.failed {
				return m.retryExecution()
			}
		case "esc":
			if m.done || m.failed {
				return m, nil // Allow parent to handle navigation back
			}
		}

	case spinner.TickMsg:
		if m.running {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case stepStartMsg:
		idx := int(msg)
		if idx < len(m.stepStates) {
			m.stepStates[idx] = models.StepRunning
			m.running = true
			m.addLog("info", fmt.Sprintf("Starting step %d: %s", idx+1, m.workflow.Steps[idx].Type.String()))
		}
		return m, m.spinner.Tick

	case stepCompleteMsg:
		idx := int(msg)
		if idx < len(m.stepStates) {
			m.stepStates[idx] = models.StepSuccess
			m.addLog("success", fmt.Sprintf("Step %d completed successfully", idx+1))

			// Move to next step
			if idx+1 < len(m.workflow.Steps) {
				return m, m.startStep(idx + 1)
			} else {
				// All steps complete
				m.running = false
				m.done = true
				m.addLog("success", "Workflow completed successfully!")
				return m, nil
			}
		}

	case stepErrorMsg:
		idx := msg.index
		if idx < len(m.stepStates) {
			m.stepStates[idx] = models.StepFailed
			m.addLog("error", fmt.Sprintf("Step %d failed: %v", idx+1, msg.err))

			if m.continueOnError && idx+1 < len(m.workflow.Steps) {
				// Continue to next step
				m.addLog("info", "Continuing despite error...")
				return m, m.startStep(idx + 1)
			} else {
				// Stop execution
				m.running = false
				m.failed = true
				m.addLog("error", "Workflow execution stopped due to error")
				return m, nil
			}
		}
	}

	return m, cmd
}

// View renders the execution dashboard
func (m *ExecutionModel) View() string {
	var lines []string

	// Title
	lines = append(lines, m.styles.Title.Render(fmt.Sprintf(" Executing: %s ", m.workflow.Name)))
	lines = append(lines, "")

	// Step progress
	if len(m.workflow.Steps) > 0 {
		lines = append(lines, m.styles.Info.Render("Progress:"))
		stepLine := ""
		for i, state := range m.stepStates {
			if i > 0 {
				stepLine += " → "
			}

			stepName := m.workflow.Steps[i].Type.String()
			var indicator string

			switch state {
			case models.StepPending:
				indicator = m.styles.Dimmed.Render("○ " + stepName)
			case models.StepRunning:
				indicator = fmt.Sprintf("%s %s", m.spinner.View(), m.styles.Warning.Render(stepName))
			case models.StepSuccess:
				indicator = m.styles.Success.Render("✔ " + stepName)
			case models.StepFailed:
				indicator = m.styles.Error.Render("✘ " + stepName)
			case models.StepSkipped:
				indicator = m.styles.Dimmed.Render("⊘ " + stepName)
			}

			stepLine += indicator
		}
		lines = append(lines, stepLine)
		lines = append(lines, "")
	}

	// Status summary
	if m.done {
		lines = append(lines, m.styles.Success.Render("✔ All steps completed successfully"))
	} else if m.failed {
		lines = append(lines, m.styles.Error.Render("✘ Workflow failed"))
	}

	if m.done || m.failed {
		lines = append(lines, "")
	}

	// Log panel
	lines = append(lines, m.styles.Info.Render("Execution Log:"))

	// Calculate visible log lines
	start := m.scrollPos
	end := start + m.logsHeight
	if end > len(m.logs) {
		end = len(m.logs)
	}
	if start < 0 {
		start = 0
	}

	if len(m.logs) == 0 {
		lines = append(lines, m.styles.Dimmed.Render("  (No logs yet)"))
	} else {
		for i := start; i < end; i++ {
			entry := m.logs[i]
			timeStr := entry.Timestamp.Format("15:04:05")
			logLine := fmt.Sprintf("  [%s] %s", timeStr, entry.Message)

			switch entry.Level {
			case "success":
				lines = append(lines, m.styles.Success.Render(logLine))
			case "error":
				lines = append(lines, m.styles.Error.Render(logLine))
			default:
				lines = append(lines, logLine)
			}
		}

		// Show scroll indicator
		if len(m.logs) > m.logsHeight {
			lines = append(lines, "")
			lines = append(lines, m.styles.Help.Render(fmt.Sprintf("  Showing %d-%d of %d lines", start+1, end, len(m.logs))))
		}
	}

	// Help footer
	lines = append(lines, "")
	if m.done || m.failed {
		help := "esc: back | r: retry | ↑/↓/pgup/pgdn: scroll logs"
		lines = append(lines, m.styles.Help.Render(help))
	} else {
		help := "esc: cancel | ↑/↓/pgup/pgdn: scroll logs"
		lines = append(lines, m.styles.Help.Render(help))
	}

	return m.styles.Box.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// startStep begins execution of a specific step
func (m *ExecutionModel) startStep(index int) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			// Send the start message immediately
			return stepStartMsg(index)
		},
		func() tea.Msg {
			// Execute the step
			if index >= 0 && index < len(m.workflow.Steps) {
				err := m.engine.ExecuteStep(m.workflow.Steps[index])
				if err != nil {
					return stepErrorMsg{index: index, err: err}
				}
				return stepCompleteMsg(index)
			}
			return stepErrorMsg{index: index, err: fmt.Errorf("invalid step index: %d", index)}
		},
	)
}

// retryExecution resets and restarts the workflow
func (m *ExecutionModel) retryExecution() (tea.Model, tea.Cmd) {
	// Reset state
	m.currentStep = 0
	m.done = false
	m.failed = false
	m.running = false
	m.scrollPos = 0

	// Reset step states
	for i := range m.stepStates {
		m.stepStates[i] = models.StepPending
	}

	// Clear logs and add initial entry
	m.logs = []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Retrying workflow: %s", m.workflow.Name),
			Level:     "info",
		},
	}

	// Start execution
	return m, tea.Batch(
		m.spinner.Tick,
		m.startStep(0),
	)
}

// addLog adds a log entry
func (m *ExecutionModel) addLog(level, message string) {
	m.logs = append(m.logs, LogEntry{
		Timestamp: time.Now(),
		Message:   message,
		Level:     level,
	})

	// Auto-scroll to bottom if near the end
	maxScroll := len(m.logs) - m.logsHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollPos >= maxScroll-2 {
		m.scrollPos = maxScroll
	}
}

// Message types
type stepStartMsg int
type stepCompleteMsg int

type stepErrorMsg struct {
	index int
	err   error
}
