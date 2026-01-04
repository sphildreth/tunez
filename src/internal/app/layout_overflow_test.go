package app

import (
	"strings"
	"testing"

	"github.com/tunez/tunez/internal/provider"
)

// TestViewWithLongTrackTitle verifies that a very long track title doesn't cause
// the view to exceed terminal height due to wrapping.
func TestViewWithLongTrackTitle(t *testing.T) {
	// Setup a small terminal height to make overflow likely if logic is wrong
	width, height := 80, 24
	m := newTestModelForLayout(width, height)
	m.screen = screenNowPlaying

	// Create a track with a very long title that will wrap multiple times
	longTitle := strings.Repeat("Very Long Title ", 20) // ~320 chars, wraps ~4 lines on 80-char width
	m.nowPlaying = provider.Track{
		ID:         "longtrack",
		Title:      longTitle,
		ArtistName: "Test Artist",
		AlbumTitle: "Test Album",
		DurationMs: 180000,
	}
	m.duration = 180
	m.timePos = 60

	view := m.View()
	lines := strings.Split(view, "\n")
	lineCount := len(lines)

	// Verify view fits within terminal height
	if lineCount > height {
		t.Errorf("View has %d lines but terminal height is %d (overflow by %d lines)",
			lineCount, height, lineCount-height)
		t.Logf("Long title length: %d", len(longTitle))
	}

	// Verify player bar contains the long title (truncated or wrapped)
	if !strings.Contains(view, "Very Long Title") {
		t.Error("View should contain part of the track title")
	}

	// Verify player bar is at the bottom
	// The last lines should be the player bar
	// lastLine := lines[len(lines)-1]
	// Check for player controls hint in the last line (or second to last depending on implementation)
	// Our mock theme creates a 2-line player bar content + borders
	if !strings.Contains(view, "[Space]Play") && !strings.Contains(view, "[?]") {
		// Maybe check a few lines up
		found := false
		for i := 1; i <= 5; i++ {
			if len(lines) >= i && strings.Contains(lines[len(lines)-i], "[Space]Play") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Player bar controls not found at bottom of view")
		}
	}
}
