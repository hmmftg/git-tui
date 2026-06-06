package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"gitflow-tui/internal/app"
	"gitflow-tui/internal/git"
	"gitflow-tui/internal/tui"
	"gitflow-tui/internal/tui/screens"
	"gitflow-tui/internal/workflow"
)

func main() {
	gitSvc := git.NewGitService(".")
	if !gitSvc.IsGitRepo() {
		fmt.Fprintf(os.Stderr, "Error: Not a git repository\n")
		os.Exit(1)
	}

	store := workflow.NewViperStore()
	ctx := tui.NewAppContext(gitSvc, store)
	initialScreen := screens.NewHomeScreen(ctx)
	application := app.NewApp(ctx, initialScreen)

	p := tea.NewProgram(application)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
