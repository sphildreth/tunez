package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("mono", Monochrome)
}

// Monochrome is a grayscale theme using white, gray, and dark gray.
func Monochrome(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	return Theme{
		Name:      "mono",
		Accent:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		Text:      lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")),
		Title:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Underline(true),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Underline(true),
	}
}
