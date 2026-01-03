package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("purple", Purple)
}

// Purple is a monochrome purple/violet theme.
func Purple(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	bright := lipgloss.Color("#BB66FF")
	medium := lipgloss.Color("#9933CC")
	dark := lipgloss.Color("#662299")
	dim := lipgloss.Color("#441166")

	return Theme{
		Name:      "purple",
		Accent:    lipgloss.NewStyle().Foreground(bright).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dim),
		Text:      lipgloss.NewStyle().Foreground(medium),
		Title:     lipgloss.NewStyle().Foreground(bright).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#55FF55")).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(dark),
		Highlight: lipgloss.NewStyle().Foreground(bright).Bold(true).Underline(true),
	}
}
