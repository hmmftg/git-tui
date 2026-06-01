package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme colors
var (
	PrimaryColor    = lipgloss.Color("#7B68EE") // MediumSlateBlue
	SecondaryColor  = lipgloss.Color("#00CED1") // DarkTurquoise
	SuccessColor    = lipgloss.Color("#32CD32") // LimeGreen
	ErrorColor      = lipgloss.Color("#FF6347") // Tomato
	WarningColor    = lipgloss.Color("#FFD700") // Gold
	InfoColor       = lipgloss.Color("#87CEEB") // SkyBlue
	TextColor       = lipgloss.Color("#FFFFFF") // White
	DimmedColor     = lipgloss.Color("#808080") // Gray
	BackgroundColor = lipgloss.Color("#1a1a2e") // Dark blue
)

// Styles provides centralized styling
type Styles struct {
	Title      lipgloss.Style
	Header     lipgloss.Style
	Footer     lipgloss.Style
	Menu       lipgloss.Style
	MenuItem   lipgloss.Style
	MenuActive lipgloss.Style
	Status     lipgloss.Style
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	Info       lipgloss.Style
	Dimmed     lipgloss.Style
	Box        lipgloss.Style
	Input      lipgloss.Style
	List       lipgloss.Style
	ListItem   lipgloss.Style
	Help       lipgloss.Style
	Key        lipgloss.Style
	Value      lipgloss.Style
}

// DefaultStyles returns the default styles
func DefaultStyles() Styles {
	s := Styles{}

	s.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Background(BackgroundColor).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor)

	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(TextColor).
		Background(PrimaryColor).
		Padding(0, 2).
		Width(100)

	s.Footer = lipgloss.NewStyle().
		Foreground(DimmedColor).
		Background(BackgroundColor).
		Padding(0, 1).
		Width(100)

	s.Menu = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Width(30)

	s.MenuItem = lipgloss.NewStyle().
		Foreground(TextColor).
		Padding(0, 1)

	s.MenuActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(BackgroundColor).
		Background(SecondaryColor).
		Padding(0, 1).
		Margin(0, 0, 0, -1)

	s.Status = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(InfoColor).
		Padding(1, 2).
		Width(70)

	s.Success = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	s.Error = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	s.Warning = lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)

	s.Info = lipgloss.NewStyle().
		Foreground(InfoColor)

	s.Dimmed = lipgloss.NewStyle().
		Foreground(DimmedColor)

	s.Box = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Margin(1, 0)

	s.Input = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(SecondaryColor).
		Padding(0, 1).
		Width(50)

	s.List = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(0, 1).
		Height(15).
		Width(50)

	s.ListItem = lipgloss.NewStyle().
		Padding(0, 1)

	s.Help = lipgloss.NewStyle().
		Foreground(DimmedColor).
		Padding(0, 1)

	s.Key = lipgloss.NewStyle().
		Bold(true).
		Foreground(SecondaryColor)

	s.Value = lipgloss.NewStyle().
		Foreground(TextColor)

	return s
}

// StatusIndicator returns an indicator for a step state
func StatusIndicator(state string) string {
	switch state {
	case "success":
		return lipgloss.NewStyle().Foreground(SuccessColor).Render("✔")
	case "error":
		return lipgloss.NewStyle().Foreground(ErrorColor).Render("✘")
	case "running":
		return lipgloss.NewStyle().Foreground(WarningColor).Render("⟳")
	case "pending":
		return lipgloss.NewStyle().Foreground(DimmedColor).Render("○")
	case "skipped":
		return lipgloss.NewStyle().Foreground(DimmedColor).Render("⊘")
	default:
		return "○"
	}
}
