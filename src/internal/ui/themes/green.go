package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("green", GreenTerminal)
}

// GreenTerminal is a classic green-on-black terminal theme.
func GreenTerminal(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	brightGreen := lipgloss.Color("#00FF00")
	mediumGreen := lipgloss.Color("#00CC00")
	darkGreen := lipgloss.Color("#008800")
	dimGreen := lipgloss.Color("#005500")

	return Theme{
		Name:      "green",
		Accent:    lipgloss.NewStyle().Foreground(brightGreen).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dimGreen),
		Text:      lipgloss.NewStyle().Foreground(mediumGreen),
		Title:     lipgloss.NewStyle().Foreground(brightGreen).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(brightGreen).Bold(true).Reverse(true),
		Success:   lipgloss.NewStyle().Foreground(brightGreen).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(mediumGreen).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(darkGreen),
		Highlight: lipgloss.NewStyle().Foreground(brightGreen).Bold(true).Underline(true),
	}
}
