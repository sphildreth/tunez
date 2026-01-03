package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("synthwave", Synthwave)
}

// Synthwave is an 80s retro synthwave/outrun theme with neon colors.
func Synthwave(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	hotPink := lipgloss.Color("#FF1493")
	electricBlue := lipgloss.Color("#00D4FF")
	purple := lipgloss.Color("#9D00FF")
	yellow := lipgloss.Color("#FFE900")
	dimPurple := lipgloss.Color("#4A0080")

	return Theme{
		Name:      "synthwave",
		Accent:    lipgloss.NewStyle().Foreground(hotPink).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dimPurple),
		Text:      lipgloss.NewStyle().Foreground(electricBlue),
		Title:     lipgloss.NewStyle().Foreground(hotPink).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(electricBlue).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(yellow).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(purple),
		Highlight: lipgloss.NewStyle().Foreground(yellow).Bold(true),
	}
}
