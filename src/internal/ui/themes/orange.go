package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("orange", Orange)
}

// Orange is a warm orange/amber theme.
func Orange(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	bright := lipgloss.Color("#FF9933")
	medium := lipgloss.Color("#CC6600")
	dark := lipgloss.Color("#884400")
	dim := lipgloss.Color("#552200")

	return Theme{
		Name:      "orange",
		Accent:    lipgloss.NewStyle().Foreground(bright).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dim),
		Text:      lipgloss.NewStyle().Foreground(medium),
		Title:     lipgloss.NewStyle().Foreground(bright).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#88FF88")).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(dark),
		Highlight: lipgloss.NewStyle().Foreground(bright).Bold(true).Underline(true),
	}
}
