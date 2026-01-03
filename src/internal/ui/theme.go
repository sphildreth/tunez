package ui

import "github.com/tunez/tunez/internal/ui/themes"

// Theme is an alias for themes.Theme for backwards compatibility.
type Theme = themes.Theme

// GetTheme returns a theme by name. Returns Rainbow if name not found.
func GetTheme(name string, noColor bool) Theme {
	return themes.Get(name, noColor)
}

// ValidTheme returns true if the theme name is valid.
func ValidTheme(name string) bool {
	return themes.Valid(name)
}

// ThemeNames returns the list of available theme names.
func ThemeNames() []string {
	return themes.Names()
}

// Re-export theme constructors for direct access
var (
	Rainbow       = themes.Rainbow
	Monochrome    = themes.Monochrome
	GreenTerminal = themes.GreenTerminal
	NoColor       = themes.NoColor
)
