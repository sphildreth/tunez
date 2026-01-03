package ui

import "testing"

func TestRainbow(t *testing.T) {
	theme := Rainbow(false)
	if theme.Name != "rainbow" {
		t.Errorf("expected name 'rainbow', got %q", theme.Name)
	}
	if theme.Accent.GetForeground() == nil {
		t.Error("Rainbow(false) should have colors")
	}
}

func TestMonochrome(t *testing.T) {
	theme := Monochrome(false)
	if theme.Name != "mono" {
		t.Errorf("expected name 'mono', got %q", theme.Name)
	}
	if theme.Accent.GetForeground() == nil {
		t.Error("Monochrome(false) should have colors")
	}
}

func TestGreenTerminal(t *testing.T) {
	theme := GreenTerminal(false)
	if theme.Name != "green" {
		t.Errorf("expected name 'green', got %q", theme.Name)
	}
	if theme.Accent.GetForeground() == nil {
		t.Error("GreenTerminal(false) should have colors")
	}
}

func TestNoColor(t *testing.T) {
	theme := NoColor(true)
	if theme.Name != "nocolor" {
		t.Errorf("expected name 'nocolor', got %q", theme.Name)
	}
	// NoColor should use bold for title
	if !theme.Title.GetBold() {
		t.Error("NoColor should use bold for title")
	}
}

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name     string
		noColor  bool
		expected string
	}{
		{"rainbow", false, "rainbow"},
		{"mono", false, "mono"},
		{"green", false, "green"},
		{"nocolor", false, "nocolor"},
		{"invalid", false, "rainbow"}, // defaults to rainbow
		{"rainbow", true, "nocolor"},  // noColor overrides
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetTheme(tt.name, tt.noColor)
			if theme.Name != tt.expected {
				t.Errorf("GetTheme(%q, %v) = %q, want %q", tt.name, tt.noColor, theme.Name, tt.expected)
			}
		})
	}
}

func TestValidTheme(t *testing.T) {
	validThemes := []string{"rainbow", "mono", "green", "nocolor"}
	for _, name := range validThemes {
		if !ValidTheme(name) {
			t.Errorf("ValidTheme(%q) should be true", name)
		}
	}

	if ValidTheme("invalid") {
		t.Error("ValidTheme('invalid') should be false")
	}
}

func TestThemeNames(t *testing.T) {
	names := ThemeNames()
	if len(names) != 4 {
		t.Errorf("expected 4 themes, got %d", len(names))
	}
}
