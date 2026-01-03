package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("neon", Neon)
}

// Neon is an electric, high-contrast neon signs theme.
func Neon(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	electricPink := lipgloss.Color("#FF10F0")
	electricBlue := lipgloss.Color("#00F5FF")
	electricGreen := lipgloss.Color("#39FF14")
	electricYellow := lipgloss.Color("#FFFF00")
	electricOrange := lipgloss.Color("#FF5F00")
	darkPurple := lipgloss.Color("#1A0033")

	return Theme{
		Name:      "neon",
		Accent:    lipgloss.NewStyle().Foreground(electricPink).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(darkPurple),
		Text:      lipgloss.NewStyle().Foreground(electricBlue),
		Title:     lipgloss.NewStyle().Foreground(electricPink).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(electricOrange).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(electricGreen).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(electricYellow).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(electricBlue),
		Highlight: lipgloss.NewStyle().Foreground(electricGreen).Bold(true),
	}
}
