package themes

import "github.com/charmbracelet/lipgloss"

func init() {
	Register("solarized", Solarized)
}

// Solarized is based on the Solarized dark color scheme.
func Solarized(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	base03 := lipgloss.Color("#002B36")
	base01 := lipgloss.Color("#586E75")
	base0 := lipgloss.Color("#839496")
	_ = lipgloss.Color("#93A1A1") // base1
	yellow := lipgloss.Color("#B58900")
	orange := lipgloss.Color("#CB4B16")
	red := lipgloss.Color("#DC322F")
	magenta := lipgloss.Color("#D33682")
	blue := lipgloss.Color("#268BD2")
	cyan := lipgloss.Color("#2AA198")
	green := lipgloss.Color("#859900")

	_ = base03
	_ = magenta

	return Theme{
		Name:      "solarized",
		Accent:    lipgloss.NewStyle().Foreground(blue).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(base01),
		Text:      lipgloss.NewStyle().Foreground(base0),
		Title:     lipgloss.NewStyle().Foreground(cyan).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(red).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(green).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(orange).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(base01),
		Highlight: lipgloss.NewStyle().Foreground(yellow).Bold(true),
	}
}
