package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name      string
	Accent    lipgloss.Style
	Dim       lipgloss.Style
	Text      lipgloss.Style
	Title     lipgloss.Style
	Error     lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Border    lipgloss.Style
	Highlight lipgloss.Style
}

// themeRegistry maps theme names to constructors.
var themeRegistry = map[string]func(bool) Theme{
	"rainbow": Rainbow,
	"mono":    Monochrome,
	"green":   GreenTerminal,
	"nocolor": NoColor,
}

// ThemeNames returns the list of available theme names.
func ThemeNames() []string {
	return []string{"rainbow", "mono", "green", "nocolor"}
}

// GetTheme returns a theme by name. Returns Rainbow if name not found.
func GetTheme(name string, noColor bool) Theme {
	// NO_COLOR environment variable overrides theme selection
	if noColor {
		return NoColor(noColor)
	}
	if fn, ok := themeRegistry[name]; ok {
		return fn(noColor)
	}
	return Rainbow(noColor)
}

// ValidTheme returns true if the theme name is valid.
func ValidTheme(name string) bool {
	_, ok := themeRegistry[name]
	return ok
}

// Rainbow is the default colorful theme.
func Rainbow(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	return Theme{
		Name:      "rainbow",
		Accent:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6FF7")),
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6F93")),
		Text:      lipgloss.NewStyle().Foreground(lipgloss.Color("#E6E6FA")),
		Title:     lipgloss.NewStyle().Foreground(lipgloss.Color("#8EEBFF")).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F56")).Bold(true),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#5CFF5C")).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD166")).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(lipgloss.Color("#7C7CFF")),
		Highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA7C4")).Bold(true),
	}
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

// GreenTerminal is a classic green-on-black terminal theme.
func GreenTerminal(noColor bool) Theme {
	if noColor {
		return NoColor(noColor)
	}
	brightGreen := lipgloss.Color("#00FF00")
	mediumGreen := lipgloss.Color("#00CC00")
	darkGreen := lipgloss.Color("#008800")
	dimGreen := lipgloss.Color("#005500")

	return Theme{
		Name:      "green",
		Accent:    lipgloss.NewStyle().Foreground(brightGreen).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(dimGreen),
		Text:      lipgloss.NewStyle().Foreground(mediumGreen),
		Title:     lipgloss.NewStyle().Foreground(brightGreen).Bold(true),
		Error:     lipgloss.NewStyle().Foreground(brightGreen).Bold(true).Reverse(true),
		Success:   lipgloss.NewStyle().Foreground(brightGreen).Bold(true),
		Warning:   lipgloss.NewStyle().Foreground(mediumGreen).Bold(true),
		Border:    lipgloss.NewStyle().Foreground(darkGreen),
		Highlight: lipgloss.NewStyle().Foreground(brightGreen).Bold(true).Underline(true),
	}
}

// NoColor is a high-contrast theme for NO_COLOR environments.
// Uses only bold, underline, and reverse instead of colors.
func NoColor(_ bool) Theme {
	reset := lipgloss.NewStyle()
	return Theme{
		Name:      "nocolor",
		Accent:    reset.Bold(true),
		Dim:       reset,
		Text:      reset,
		Title:     reset.Bold(true),
		Error:     reset.Bold(true),
		Success:   reset.Bold(true),
		Warning:   reset.Bold(true),
		Border:    reset,
		Highlight: reset.Reverse(true),
	}
}
