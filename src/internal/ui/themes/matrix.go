package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("matrix", Matrix)
}

// Matrix is a green-on-black "Matrix" style hacker theme.
func Matrix(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	// Various shades of the iconic Matrix green
	bright := lipgloss.Color("#00FF41")
	medium := lipgloss.Color("#00D135")
	dark := lipgloss.Color("#009929")
	dim := lipgloss.Color("#004D14")
	white := lipgloss.Color("#FFFFFF")

	return Theme{
		Name:      "matrix",
		Accent:    lipgloss.NewStyle().Foreground(bright).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dim),
		Text:      lipgloss.NewStyle().Foreground(medium),
		Title:     lipgloss.NewStyle().Foreground(bright).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(white).Background(dark).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(bright).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(bright).Blink(true),
		Border:    lipgloss.NewStyle().Foreground(dark),
		Highlight: lipgloss.NewStyle().Foreground(bright).Bold(true).Underline(true),
	}
}
