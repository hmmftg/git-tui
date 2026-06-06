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
	ctx    *tui.AppContext
	router *Router
}

// NewApp creates a new App.
func NewApp(ctx *tui.AppContext, initialScreen tui.Screen) *App {
	return &App{
		ctx:    ctx,
		router: NewRouter(initialScreen),
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
		if s == "q" {
			return a, tea.Quit
		}
		if s == "home" {
			a.router.Navigate(screens.NewHomeScreen(a.ctx))
			return a, nil
		}

	case tui.OperationRequestMsg:
		a.router.Navigate(screens.NewOperationScreen(a.ctx, msg.Title, msg.Run))
		return a, nil

	case tui.HomeRequestMsg:
		a.router.Navigate(screens.NewHomeScreen(a.ctx))
		return a, nil
	}

	// Delegate to current screen
	screen, cmd := a.router.Current().Update(msg)
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
	branch := "(unknown)"
	if status, err := a.ctx.GitService.Status(); err == nil {
		branch = status.Branch
	}
	header := fmt.Sprintf(" GitFlow TUI | 🔀 %s ", branch)
	style := a.ctx.Styles.Header
	if a.ctx.Width > 0 {
		style = style.Width(a.ctx.Width)
	}
	return style.Render(header)
}

func (a *App) renderFooter() string {
	help := "q: quit | home: go home | ?: help"
	style := a.ctx.Styles.Footer
	if a.ctx.Width > 0 {
		style = style.Width(a.ctx.Width)
	}
	return style.Render(help)
}
