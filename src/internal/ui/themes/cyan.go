package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("cyan", Cyan)
}

// Cyan is a cool cyan/teal theme.
func Cyan(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	bright := lipgloss.Color("#00FFFF")
	medium := lipgloss.Color("#00CCCC")
	dark := lipgloss.Color("#008888")
	dim := lipgloss.Color("#005555")

	return Theme{
		Name:      "cyan",
		Accent:    lipgloss.NewStyle().Foreground(bright).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dim),
		Text:      lipgloss.NewStyle().Foreground(medium),
		Title:     lipgloss.NewStyle().Foreground(bright).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(bright).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(dark),
		Highlight: lipgloss.NewStyle().Foreground(bright).Bold(true).Underline(true),
	}
}
