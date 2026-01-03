package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("sunset", Sunset)
}

// Sunset is a warm orange-to-purple gradient inspired theme.
func Sunset(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	gold := lipgloss.Color("#FFD700")
	orange := lipgloss.Color("#FF8C00")
	coral := lipgloss.Color("#FF6347")
	magenta := lipgloss.Color("#FF00FF")
	purple := lipgloss.Color("#9400D3")
	dimPurple := lipgloss.Color("#4B0082")

	return Theme{
		Name:      "sunset",
		Accent:    lipgloss.NewStyle().Foreground(gold).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dimPurple),
		Text:      lipgloss.NewStyle().Foreground(orange),
		Title:     lipgloss.NewStyle().Foreground(gold).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(coral).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(gold).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(orange).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(purple),
		Highlight: lipgloss.NewStyle().Foreground(magenta).Bold(true),
	}
}
