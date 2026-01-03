package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("coffee", Coffee)
}

// Coffee is a warm brown coffee/mocha inspired theme.
func Coffee(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	espresso := lipgloss.Color("#3C2415")
	latte := lipgloss.Color("#C4A484")
	_ = lipgloss.Color("#967259") // mocha
	cream := lipgloss.Color("#FFFDD0")
	caramel := lipgloss.Color("#FFD59A")
	chocolate := lipgloss.Color("#7B3F00")

	return Theme{
		Name:      "coffee",
		Accent:    lipgloss.NewStyle().Foreground(cream).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(espresso),
		Text:      lipgloss.NewStyle().Foreground(latte),
		Title:     lipgloss.NewStyle().Foreground(cream).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#CD5C5C")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(caramel).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(caramel).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(chocolate),
		Highlight: lipgloss.NewStyle().Foreground(cream).Bold(true).Underline(true),
	}
}
