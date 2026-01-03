package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("gruvbox", Gruvbox)
}

// Gruvbox is based on the Gruvbox dark color scheme - retro groove.
func Gruvbox(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	bg := lipgloss.Color("#282828")
	fg := lipgloss.Color("#EBDBB2")
	gray := lipgloss.Color("#928374")
	red := lipgloss.Color("#FB4934")
	green := lipgloss.Color("#B8BB26")
	yellow := lipgloss.Color("#FABD2F")
	_ = lipgloss.Color("#83A598") // blue
	_ = lipgloss.Color("#D3869B") // purple
	aqua := lipgloss.Color("#8EC07C")
	orange := lipgloss.Color("#FE8019")

	_ = bg

	return Theme{
		Name:      "gruvbox",
		Accent:    lipgloss.NewStyle().Foreground(orange).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(gray),
		Text:      lipgloss.NewStyle().Foreground(fg),
		Title:     lipgloss.NewStyle().Foreground(yellow).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(red).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(green).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(orange).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(gray),
		Highlight: lipgloss.NewStyle().Foreground(aqua).Bold(true),
	}
}
