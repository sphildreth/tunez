package artwork

import (
	"os"
	"testing"
)

func TestDetectProtocol(t *testing.T) {
	// Helper to reset state
	reset := func() {
		ResetProtocolDetection()
		os.Unsetenv("TERM")
		os.Unsetenv("KITTY_WINDOW_ID")
		os.Unsetenv("TERM_PROGRAM")
		os.Unsetenv("WT_SESSION")
		os.Unsetenv("LC_TERMINAL")
		os.Unsetenv("ITERM_SESSION_ID")
		os.Unsetenv("WEZTERM_EXECUTABLE")
		os.Unsetenv("XTERM_VERSION")
		os.Unsetenv("CONTOUR_SESSION_ID")
	}

	tests := []struct {
		name     string
		setup    func()
		expected Protocol
	}{
		{
			name: "Kitty via TERM",
			setup: func() {
				os.Setenv("TERM", "xterm-kitty")
			},
			expected: ProtocolKitty,
		},
		{
			name: "Kitty via Window ID",
			setup: func() {
				os.Setenv("TERM", "xterm")
				os.Setenv("KITTY_WINDOW_ID", "1234")
			},
			expected: ProtocolKitty,
		},
		{
			name: "Sixel via iTerm2 TERM_PROGRAM",
			setup: func() {
				os.Setenv("TERM_PROGRAM", "iTerm.app")
			},
			expected: ProtocolSixel,
		},
		{
			name: "Sixel via iTerm2 LC_TERMINAL",
			setup: func() {
				os.Setenv("LC_TERMINAL", "iTerm2")
			},
			expected: ProtocolSixel,
		},
		{
			name: "Sixel via WezTerm",
			setup: func() {
				os.Setenv("WEZTERM_EXECUTABLE", "/usr/bin/wezterm")
			},
			expected: ProtocolSixel,
		},
		{
			name: "Sixel via Foot",
			setup: func() {
				os.Setenv("TERM", "foot")
			},
			expected: ProtocolSixel,
		},
		{
			name: "Sixel via Windows Terminal",
			setup: func() {
				os.Setenv("WT_SESSION", "uuid")
			},
			expected: ProtocolSixel,
		},
		{
			name: "Fallback to ANSI",
			setup: func() {
				os.Setenv("TERM", "xterm-256color")
			},
			expected: ProtocolANSI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reset()
			defer reset()
			tt.setup()

			got := DetectProtocol()
			if got != tt.expected {
				t.Errorf("DetectProtocol() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestForceProtocol(t *testing.T) {
	ResetProtocolDetection()
	ForceProtocol(ProtocolKitty)
	if p := DetectProtocol(); p != ProtocolKitty {
		t.Errorf("expected Kitty after force, got %v", p)
	}

	// Should persist even if we try to detect again
	if p := DetectProtocol(); p != ProtocolKitty {
		t.Errorf("expected Kitty to persist, got %v", p)
	}
}
