package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("blue", Blue)
}

// Blue is a monochrome blue theme.
func Blue(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	bright := lipgloss.Color("#4488FF")
	medium := lipgloss.Color("#0066CC")
	dark := lipgloss.Color("#003388")
	dim := lipgloss.Color("#002255")

	return Theme{
		Name:      "blue",
		Accent:    lipgloss.NewStyle().Foreground(bright).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dim),
		Text:      lipgloss.NewStyle().Foreground(medium),
		Title:     lipgloss.NewStyle().Foreground(bright).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(bright).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA44")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(dark),
		Highlight: lipgloss.NewStyle().Foreground(bright).Bold(true).Underline(true),
	}
}
