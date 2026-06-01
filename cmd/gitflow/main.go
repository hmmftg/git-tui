package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"gitflow-tui/internal/git"
	"gitflow-tui/internal/tui"
)

func main() {
	// Create git service
	gitSvc := git.NewGitService(".")

	// Check if we're in a git repository
	if !gitSvc.IsGitRepo() {
		fmt.Fprintf(os.Stderr, "Error: Not a git repository\n")
		os.Exit(1)
	}

	// Create and run the TUI application
	app := tui.NewApp(gitSvc)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
