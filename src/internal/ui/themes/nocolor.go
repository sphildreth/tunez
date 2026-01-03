package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("nocolor", NoColor)
}

// NoColor is a high-contrast theme for NO_COLOR environments.
// Uses only bold, underline, and reverse instead of colors.
func NoColor(_ bool) Theme {
	reset := lipgloss.NewStyle()
	return Theme{
		Name:      "nocolor",
		Accent:    reset.Bold(true),
		Dim:       reset,
		Text:      reset,
		Title:     reset.Bold(true),
		Error:     reset.Bold(true),
		Success:   reset.Bold(true),
		Warning:   reset.Bold(true),
		Border:    reset,
		Highlight: reset.Reverse(true),
	}
}
