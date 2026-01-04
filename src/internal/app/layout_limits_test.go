package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/tunez/tunez/internal/config"
	"github.com/tunez/tunez/internal/provider"
	"github.com/tunez/tunez/internal/queue"
	"github.com/tunez/tunez/internal/ui"
)

// TestViewLayoutLimits ensures that the View() method never returns a string
// with more lines than the Model's height, preventing scroll/ghosting.
func TestViewLayoutLimits(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"Standard 80x24", 80, 24},
		{"Narrow 50x24", 50, 24},
		{"Wide 120x40", 120, 40},
		{"Short 80x15", 80, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup minimalist model per test to avoid state leakage
			cfg := &config.Config{
				UI:      config.UIConfig{Theme: "rainbow"},
				Artwork: config.ArtworkConfig{Enabled: false},
			}
			theme := ui.GetTheme("rainbow", true)

			q := queue.New()
			// Add enough tracks to ensure we have content to scroll
			for i := 0; i < 50; i++ {
				q.Add(provider.Track{
					ID:         "t1",
					Title:      "A Very Long Title That Might Wrap If The Term Is Narrow Enough to Cause Issues " + strings.Repeat("x", i),
					ArtistName: "Some Artist",
					DurationMs: 180000,
				})
			}

			m := New(cfg, &mockProvider{}, nil, nil, nil, theme, StartupOptions{}, nil, nil, nil, nil)
			m.queue = q
			m.screen = screenQueue
			m.startupDone = true

			// Set dimensions
			m.width = tt.width
			m.height = tt.height
			m.focusedPane = paneContent

			output := m.View()
			gotHeight := lipgloss.Height(output)

			t.Logf("Dimensions: %dx%d, Output Height: %d", tt.width, tt.height, gotHeight)

			if gotHeight > tt.height {
				t.Errorf("View output height %d exceeds terminal height %d", gotHeight, tt.height)
			}

			// Strictly check width of EVERY line to ensure no wrapping
			lines := strings.Split(output, "\n")
			for i, line := range lines {
				// lipgloss.Width handles ansi codes correctly
				visualWidth := lipgloss.Width(line)

				if visualWidth > tt.width {
					t.Errorf("[%s] Line %d width %d exceeds terminal width %d\nContent: %q",
						tt.name, i+1, visualWidth, tt.width, line)
				}
			}
			// Log output if failed (for debugging)
			if t.Failed() {
				t.Logf("Check Output lines:")
				for i, l := range lines {
					t.Logf("%02d: %s", i+1, l)
				}
			}
		})
	}
}
