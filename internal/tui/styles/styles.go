package styles

import "charm.land/lipgloss/v2"

// Theme colors
var (
	PrimaryColor    = lipgloss.Color("#7B68EE")
	SecondaryColor  = lipgloss.Color("#00CED1")
	SuccessColor    = lipgloss.Color("#32CD32")
	ErrorColor      = lipgloss.Color("#FF6347")
	WarningColor    = lipgloss.Color("#FFD700")
	InfoColor       = lipgloss.Color("#87CEEB")
	TextColor       = lipgloss.Color("#FFFFFF")
	DimmedColor     = lipgloss.Color("#808080")
	BackgroundColor = lipgloss.Color("#1a1a2e")
)

// Styles provides centralized styling.
type Styles struct {
	Title   lipgloss.Style
	Header  lipgloss.Style
	Footer  lipgloss.Style
	Box     lipgloss.Style
	Info    lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Success lipgloss.Style
	Help    lipgloss.Style
	Key     lipgloss.Style
	Value   lipgloss.Style
	Dimmed  lipgloss.Style
}

// DefaultStyles returns the default styles.
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

	s.Box = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Margin(1, 0)

	s.Info = lipgloss.NewStyle().
		Foreground(InfoColor)

	s.Warning = lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)

	s.Error = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	s.Success = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	s.Help = lipgloss.NewStyle().
		Foreground(DimmedColor).
		Padding(0, 1)

	s.Key = lipgloss.NewStyle().
		Bold(true).
		Foreground(SecondaryColor)

	s.Value = lipgloss.NewStyle().
		Foreground(TextColor)

	s.Dimmed = lipgloss.NewStyle().
		Foreground(DimmedColor)

	return s
}

// StatusIndicator returns an indicator for a step state.
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
