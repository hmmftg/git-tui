package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"gitflow-tui/internal/tui"
	"gitflow-tui/internal/tui/screens"
)

// App is the global Bubble Tea application model.
type App struct {
	ctx          *tui.AppContext
	router       *Router
	cachedBranch string
}

// NewApp creates a new App.
func NewApp(ctx *tui.AppContext, initialScreen tui.Screen) *App {
	branch := "(unknown)"
	if status, err := ctx.GitService.Status(); err == nil {
		branch = status.Branch
	}
	return &App{
		ctx:          ctx,
		router:       NewRouter(initialScreen),
		cachedBranch: branch,
	}
}

// Init initializes the application.
func (a *App) Init() tea.Cmd {
	return a.router.Current().Init()
}

// Update handles messages and delegates to the current screen.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.ctx.Width = msg.Width
		a.ctx.Height = msg.Height
		return a, nil

	case tea.KeyMsg:
		s := msg.String()
		if s == "ctrl+c" {
			return a, tea.Quit
		}
		// Only allow 'q' to quit from specific screens (not forms)
		// Let screens handle their own quit logic
		if s == "home" {
			newScreen := screens.NewHomeScreen(a.ctx)
			a.router.Navigate(newScreen)
			if status, err := a.ctx.GitService.Status(); err == nil {
				a.cachedBranch = status.Branch
			}
			return a, newScreen.Init()
		}

	case tui.OperationRequestMsg:
		newScreen := screens.NewOperationScreen(a.ctx, msg.Title, msg.Command, msg.Run)
		a.router.Navigate(newScreen)
		return a, newScreen.Init()

	case tui.HomeRequestMsg:
		newScreen := screens.NewHomeScreen(a.ctx)
		a.router.Navigate(newScreen)
		// Update branch cache when going home
		if status, err := a.ctx.GitService.Status(); err == nil {
			a.cachedBranch = status.Branch
		}
		return a, newScreen.Init()
	}

	// Delegate to current screen
	screen, cmd := a.router.Current().Update(msg)
	if screen != a.router.Current() {
		// Screen changed, initialize the new screen
		a.router.Navigate(screen)
		initCmd := screen.Init()
		return a, tea.Batch(cmd, initCmd)
	}
	a.router.Navigate(screen)
	return a, cmd
}

// View renders the full UI.
func (a *App) View() tea.View {
	if a.ctx.Width == 0 || a.ctx.Height == 0 {
		return tea.NewView("Loading...")
	}

	content := a.router.Current().View()

	v := tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		a.renderHeader(),
		content.Content,
		a.renderFooter(),
	))

	v.AltScreen = true

	return v
}

func (a *App) renderHeader() string {
	header := fmt.Sprintf(" GitFlow TUI | 🔀 %s ", a.cachedBranch)
	style := a.ctx.Styles.Header
	if a.ctx.Width > 0 {
		style = style.Width(a.ctx.Width)
	}
	return style.Render(header)
}

func (a *App) renderFooter() string {
	var help string
	// Check if we're on the home screen
	if _, isHome := a.router.Current().(*screens.HomeScreen); isHome {
		help = "q: quit | ?: help"
	} else {
		help = "ctrl+c: quit | home: go home | ?: help"
	}
	style := a.ctx.Styles.Footer
	if a.ctx.Width > 0 {
		style = style.Width(a.ctx.Width)
	}
	return style.Render(help)
}
