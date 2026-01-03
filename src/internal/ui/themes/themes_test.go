package themes

import (
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		noColor  bool
		wantName string
	}{
		{"rainbow", false, "rainbow"},
		{"mono", false, "mono"},
		{"green", false, "green"},
		{"nocolor", false, "nocolor"},
		{"invalid", false, "rainbow"}, // falls back to rainbow
		{"rainbow", true, "nocolor"},  // noColor overrides
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := Get(tt.name, tt.noColor)
			if theme.Name != tt.wantName {
				t.Errorf("Get(%q, %v) = %q, want %q", tt.name, tt.noColor, theme.Name, tt.wantName)
			}
		})
	}
}

func TestValid(t *testing.T) {
	validNames := []string{"rainbow", "mono", "green", "nocolor"}
	for _, name := range validNames {
		if !Valid(name) {
			t.Errorf("Valid(%q) = false, want true", name)
		}
	}

	if Valid("invalid") {
		t.Error("Valid(\"invalid\") = true, want false")
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) < 4 {
		t.Errorf("Names() returned %d themes, want at least 4", len(names))
	}

	// Check all expected themes are present
	expected := map[string]bool{"rainbow": false, "mono": false, "green": false, "nocolor": false}
	for _, name := range names {
		expected[name] = true
	}
	for name, found := range expected {
		if !found {
			t.Errorf("Names() missing %q", name)
		}
	}
}

func TestThemesNotPanic(t *testing.T) {
	// Ensure all themes can be constructed without panic
	for _, name := range Names() {
		t.Run(name, func(t *testing.T) {
			theme := Get(name, false)
			if theme.Name == "" {
				t.Errorf("theme %q has empty name", name)
			}
		})
	}
}
