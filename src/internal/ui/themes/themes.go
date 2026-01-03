// Package themes provides color themes for the Tunez TUI.
package themes

import "github.com/charmbracelet/lipgloss"

// Theme defines the color scheme for the TUI.
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

// ThemeFunc is a constructor function for a theme.
type ThemeFunc func(noColor bool) Theme

// registry maps theme names to constructors.
var registry = make(map[string]ThemeFunc)

// Register adds a theme to the registry.
func Register(name string, fn ThemeFunc) {
	registry[name] = fn
}

// Get returns a theme by name. Returns Rainbow if name not found.
func Get(name string, noColor bool) Theme {
	// NO_COLOR environment variable overrides theme selection
	if noColor {
		if fn, ok := registry["nocolor"]; ok {
			return fn(noColor)
		}
	}
	if fn, ok := registry[name]; ok {
		return fn(noColor)
	}
	// Default to rainbow
	if fn, ok := registry["rainbow"]; ok {
		return fn(noColor)
	}
	// Fallback if nothing registered
	return Theme{Name: "default"}
}

// Valid returns true if the theme name is valid.
func Valid(name string) bool {
	_, ok := registry[name]
	return ok
}

// Names returns the list of available theme names.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
