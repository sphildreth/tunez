package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("ocean", Ocean)
}

// Ocean is a deep blue ocean-inspired theme.
func Ocean(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	seafoam := lipgloss.Color("#20B2AA")
	aqua := lipgloss.Color("#00CED1")
	teal := lipgloss.Color("#008B8B")
	navy := lipgloss.Color("#000080")
	_ = lipgloss.Color("#00008B") // deepBlue
	coral := lipgloss.Color("#FF7F50")
	sand := lipgloss.Color("#F4A460")

	return Theme{
		Name:      "ocean",
		Accent:    lipgloss.NewStyle().Foreground(aqua).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(navy),
		Text:      lipgloss.NewStyle().Foreground(seafoam),
		Title:     lipgloss.NewStyle().Foreground(aqua).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(coral).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(seafoam).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(sand).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(teal),
		Highlight: lipgloss.NewStyle().Foreground(aqua).Bold(true).Underline(true),
	}
}
