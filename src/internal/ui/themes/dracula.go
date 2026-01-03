package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("dracula", Dracula)
}

// Dracula is based on the popular Dracula color scheme.
func Dracula(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	purple := lipgloss.Color("#BD93F9")
	pink := lipgloss.Color("#FF79C6")
	green := lipgloss.Color("#50FA7B")
	cyan := lipgloss.Color("#8BE9FD")
	orange := lipgloss.Color("#FFB86C")
	red := lipgloss.Color("#FF5555")
	_ = lipgloss.Color("#F1FA8C") // yellow
	comment := lipgloss.Color("#6272A4")
	foreground := lipgloss.Color("#F8F8F2")

	return Theme{
		Name:      "dracula",
		Accent:    lipgloss.NewStyle().Foreground(pink).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(comment),
		Text:      lipgloss.NewStyle().Foreground(foreground),
		Title:     lipgloss.NewStyle().Foreground(purple).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(red).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(green).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(orange).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(comment),
		Highlight: lipgloss.NewStyle().Foreground(cyan).Bold(true),
	}
}
