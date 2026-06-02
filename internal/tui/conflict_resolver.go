package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitflow-tui/internal/git"
)

// resolutionState tracks the state of a file resolution
type resolutionState struct {
	choice git.ResolutionChoice
	done   bool
}

// ConflictResolverModel handles interactive conflict resolution
type ConflictResolverModel struct {
	gitSvc      git.GitService
	styles      Styles
	files       []string
	resolutions map[string]resolutionState
	cursor      int
	content     string
	loading     bool
	err         error
	message     string
}

// NewConflictResolver creates a new conflict resolver model
func NewConflictResolver(gitSvc git.GitService, styles Styles) *ConflictResolverModel {
	return &ConflictResolverModel{
		gitSvc:      gitSvc,
		styles:      styles,
		resolutions: make(map[string]resolutionState),
	}
}

// Init initializes the model by loading conflicted files
func (m *ConflictResolverModel) Init() tea.Cmd {
	return m.loadConflictedFiles()
}

// Update handles messages
func (m *ConflictResolverModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't handle keys while loading
		if m.loading {
			return m, nil
		}

		if len(m.files) == 0 {
			// No conflicts, only exit available
			switch msg.String() {
			case "q", "esc":
				return m, func() tea.Msg {
					return conflictsResolvedMsg{}
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				return m, m.loadFileContent()
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
				return m, m.loadFileContent()
			}
		case "o":
			// Choose ours (current)
			return m, m.resolveCurrent(git.Ours)
		case "t":
			// Choose theirs (incoming)
			return m, m.resolveCurrent(git.Theirs)
		case "b":
			// Choose both (keep markers)
			return m, m.resolveCurrent(git.Both)
		case "enter":
			// Apply resolution and move to next
			if m.cursor < len(m.files)-1 {
				m.cursor++
				return m, m.loadFileContent()
			}
		case "c":
			// Complete/continue - check if all resolved
			if m.allResolved() {
				return m, func() tea.Msg {
					return conflictsResolvedMsg{}
				}
			}
			m.message = "⚠ Not all conflicts resolved!"
		case "a":
			// Abort
			return m, func() tea.Msg {
				return conflictsAbortedMsg{}
			}
		case "esc":
			return m, func() tea.Msg {
				return conflictsCancelledMsg{}
			}
		}

	case conflictedFilesLoadedMsg:
		m.files = msg.files
		m.loading = false
		if len(m.files) > 0 {
			return m, m.loadFileContent()
		}
		return m, nil

	case fileContentLoadedMsg:
		m.content = string(msg)
		m.loading = false
		return m, nil

	case fileResolvedMsg:
		m.resolutions[msg.file] = resolutionState{
			choice: msg.choice,
			done:   true,
		}
		m.message = fmt.Sprintf("✔ %s: %s", msg.file, choiceLabel(msg.choice))
		// Move to next unresolved file
		for i := m.cursor + 1; i < len(m.files); i++ {
			if !m.resolutions[m.files[i]].done {
				m.cursor = i
				return m, m.loadFileContent()
			}
		}
		// All files after current are resolved, check before
		for i := 0; i < m.cursor; i++ {
			if !m.resolutions[m.files[i]].done {
				m.cursor = i
				return m, m.loadFileContent()
			}
		}
		return m, nil

	case error:
		m.err = msg
		m.loading = false
		m.message = fmt.Sprintf("✘ Error: %v", msg)
		return m, nil
	}

	return m, nil
}

// View renders the conflict resolver screen
func (m *ConflictResolverModel) View() string {
	if m.loading {
		return m.styles.Box.Render(m.styles.Info.Render("Loading conflicted files..."))
	}

	if len(m.files) == 0 {
		return m.styles.Box.Render(
			m.styles.Success.Render("✔ No conflicts to resolve!") + "\n\n" +
				m.styles.Help.Render("Press esc or q to continue"),
		)
	}

	var leftPanel []string
	var rightPanel []string

	// Left panel: file list
	leftPanel = append(leftPanel, m.styles.Title.Render(" Conflicted Files "))
	leftPanel = append(leftPanel, "")

	for i, file := range m.files {
		state := m.resolutions[file]
		cursor := "  "
		if m.cursor == i {
			cursor = m.styles.Key.Render("▸ ")
		}

		icon := "○" // Pending
		if state.done {
			switch state.choice {
			case git.Ours:
				icon = m.styles.Success.Render("C") // Current
			case git.Theirs:
				icon = m.styles.Success.Render("I") // Incoming
			case git.Both:
				icon = m.styles.Warning.Render("B") // Both
			}
		}

		if m.cursor == i {
			leftPanel = append(leftPanel, fmt.Sprintf("%s%s %s", cursor, icon, m.styles.Key.Render(file)))
		} else {
			leftPanel = append(leftPanel, fmt.Sprintf("%s%s %s", cursor, icon, file))
		}
	}

	leftPanel = append(leftPanel, "")
	leftPanel = append(leftPanel, m.styles.Help.Render("o: current | t: incoming | b: both"))
	leftPanel = append(leftPanel, m.styles.Help.Render("enter: next | c: continue | a: abort"))

	// Right panel: content preview
	rightPanel = append(rightPanel, m.styles.Title.Render(" Preview "))
	rightPanel = append(rightPanel, "")

	if m.content == "" {
		rightPanel = append(rightPanel, m.styles.Info.Render("Loading..."))
	} else {
		// Show content with syntax highlighting for conflict markers
		lines := strings.Split(m.content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "<<<<<<<") {
				rightPanel = append(rightPanel, m.styles.Warning.Render(line))
			} else if strings.HasPrefix(line, "=======") {
				rightPanel = append(rightPanel, m.styles.Warning.Render(line))
			} else if strings.HasPrefix(line, ">>>>>>>") {
				rightPanel = append(rightPanel, m.styles.Warning.Render(line))
			} else {
				rightPanel = append(rightPanel, line)
			}
		}
	}

	// Status message
	if m.message != "" {
		rightPanel = append(rightPanel, "")
		if m.err != nil {
			rightPanel = append(rightPanel, m.styles.Error.Render(m.message))
		} else {
			rightPanel = append(rightPanel, m.styles.Success.Render(m.message))
		}
	}

	// Render panels side by side
	left := lipgloss.JoinVertical(lipgloss.Left, leftPanel...)
	right := lipgloss.JoinVertical(lipgloss.Left, rightPanel...)

	// Use a simple split layout
	leftStyle := lipgloss.NewStyle().Width(40)
	rightStyle := lipgloss.NewStyle().Width(60)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftStyle.Render(left), rightStyle.Render(right))

	return m.styles.Box.Render(content)
}

// loadConflictedFiles loads the list of conflicted files
func (m *ConflictResolverModel) loadConflictedFiles() tea.Cmd {
	return func() tea.Msg {
		files, err := m.gitSvc.GetConflictedFiles()
		if err != nil {
			return err
		}
		return conflictedFilesLoadedMsg{files: files}
	}
}

// loadFileContent loads the content of the current file
func (m *ConflictResolverModel) loadFileContent() tea.Cmd {
	if m.cursor >= len(m.files) {
		return nil
	}
	file := m.files[m.cursor]
	return func() tea.Msg {
		content, err := m.gitSvc.ReadFile(file)
		if err != nil {
			return err
		}
		return fileContentLoadedMsg(content)
	}
}

// resolveCurrent applies resolution to the current file
func (m *ConflictResolverModel) resolveCurrent(choice git.ResolutionChoice) tea.Cmd {
	if m.cursor >= len(m.files) {
		return nil
	}
	file := m.files[m.cursor]
	return func() tea.Msg {
		err := m.gitSvc.ResolveConflict(file, choice)
		if err != nil {
			return err
		}
		return fileResolvedMsg{file: file, choice: choice}
	}
}

// allResolved checks if all conflicts are resolved
func (m *ConflictResolverModel) allResolved() bool {
	if len(m.files) == 0 {
		return true
	}
	for _, file := range m.files {
		if !m.resolutions[file].done {
			return false
		}
	}
	return true
}

// choiceLabel returns a label for the resolution choice
func choiceLabel(choice git.ResolutionChoice) string {
	switch choice {
	case git.Ours:
		return "current (ours)"
	case git.Theirs:
		return "incoming (theirs)"
	case git.Both:
		return "both (keep markers)"
	default:
		return "unknown"
	}
}

// Message types
type conflictedFilesLoadedMsg struct {
	files []string
}
type fileContentLoadedMsg string
type fileResolvedMsg struct {
	file   string
	choice git.ResolutionChoice
}
type conflictsResolvedMsg struct{}
type conflictsAbortedMsg struct{}
type conflictsCancelledMsg struct{}

// Use git.Ours, git.Theirs, git.Both for resolution choices
