package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("red", Red)
}

// Red is a monochrome red theme.
func Red(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	bright := lipgloss.Color("#FF4444")
	medium := lipgloss.Color("#CC0000")
	dark := lipgloss.Color("#880000")
	dim := lipgloss.Color("#550000")

	return Theme{
		Name:      "red",
		Accent:    lipgloss.NewStyle().Foreground(bright).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dim),
		Text:      lipgloss.NewStyle().Foreground(medium),
		Title:     lipgloss.NewStyle().Foreground(bright).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(bright).Bold(true).Reverse(true),
		Success:   lipgloss.NewStyle().Foreground(bright).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(medium).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(dark),
		Highlight: lipgloss.NewStyle().Foreground(bright).Bold(true).Underline(true),
	}
}
