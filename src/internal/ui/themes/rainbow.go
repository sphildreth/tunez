package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("rainbow", Rainbow)
}

// Rainbow is the default colorful theme.
func Rainbow(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	return Theme{
		Name:      "rainbow",
		Accent:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6FF7")),
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6F93")),
		Text:      lipgloss.NewStyle().Foreground(lipgloss.Color("#E6E6FA")),
		Title:     lipgloss.NewStyle().Foreground(lipgloss.Color("#8EEBFF")).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F56")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#5CFF5C")).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD166")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(lipgloss.Color("#7C7CFF")),
		Highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA7C4")).Bold(true),
	}
}
