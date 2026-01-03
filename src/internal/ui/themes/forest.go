package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("forest", Forest)
}

// Forest is an earthy green and brown nature-inspired theme.
func Forest(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	leafGreen := lipgloss.Color("#228B22")
	moss := lipgloss.Color("#8FBC8F")
	bark := lipgloss.Color("#8B4513")
	_ = lipgloss.Color("#006400") // darkGreen
	dimBrown := lipgloss.Color("#3D2314")
	gold := lipgloss.Color("#DAA520")
	cream := lipgloss.Color("#FFFDD0")

	return Theme{
		Name:      "forest",
		Accent:    lipgloss.NewStyle().Foreground(leafGreen).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dimBrown),
		Text:      lipgloss.NewStyle().Foreground(moss),
		Title:     lipgloss.NewStyle().Foreground(leafGreen).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#CD5C5C")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(moss).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(gold).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(bark),
		Highlight: lipgloss.NewStyle().Foreground(cream).Bold(true),
	}
}
