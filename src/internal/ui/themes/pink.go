package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("pink", Pink)
}

// Pink is a soft pink/magenta theme.
func Pink(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	bright := lipgloss.Color("#FF66B2")
	medium := lipgloss.Color("#CC3399")
	dark := lipgloss.Color("#991166")
	dim := lipgloss.Color("#660044")

	return Theme{
		Name:      "pink",
		Accent:    lipgloss.NewStyle().Foreground(bright).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dim),
		Text:      lipgloss.NewStyle().Foreground(medium),
		Title:     lipgloss.NewStyle().Foreground(bright).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#66FF66")).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(dark),
		Highlight: lipgloss.NewStyle().Foreground(bright).Bold(true).Underline(true),
	}
}
