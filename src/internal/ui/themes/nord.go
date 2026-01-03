package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("nord", Nord)
}

// Nord is based on the popular Nord color palette - arctic, north-bluish colors.
func Nord(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	// Polar Night
	nord0 := lipgloss.Color("#2E3440")
	nord3 := lipgloss.Color("#4C566A")
	// Snow Storm
	nord4 := lipgloss.Color("#D8DEE9")
	nord6 := lipgloss.Color("#ECEFF4")
	// Frost
	_ = lipgloss.Color("#8FBCBB") // nord7
	nord8 := lipgloss.Color("#88C0D0")
	nord9 := lipgloss.Color("#81A1C1")
	// Aurora
	nord11 := lipgloss.Color("#BF616A") // red
	nord13 := lipgloss.Color("#EBCB8B") // yellow
	nord14 := lipgloss.Color("#A3BE8C") // green

	_ = nord0

	return Theme{
		Name:      "nord",
		Accent:    lipgloss.NewStyle().Foreground(nord8).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(nord3),
		Text:      lipgloss.NewStyle().Foreground(nord4),
		Title:     lipgloss.NewStyle().Foreground(nord9).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(nord11).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(nord14).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(nord13).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(nord3),
		Highlight: lipgloss.NewStyle().Foreground(nord6).Bold(true),
	}
}
