package app

import (
	"strings"
	"testing"

	"github.com/tunez/tunez/internal/config"
	"github.com/tunez/tunez/internal/player"
	"github.com/tunez/tunez/internal/provider"
	"github.com/tunez/tunez/internal/ui"
)

func newTestModelForLayout(width, height int) Model {
	cfg := &config.Config{
		ActiveProfile: "test",
		Profiles: []config.Profile{
			{ID: "test", Name: "Test", Provider: "mock", Enabled: true},
		},
		UI: config.UIConfig{
			PageSize: 50,
			Theme:    "rainbow",
		},
		Player: config.PlayerConfig{
			SeekSmall:  5,
			SeekLarge:  30,
			VolumeStep: 5,
		},
		Queue: config.QueueConfig{Persist: false},
		Keybindings: config.KeybindConfig{
			Quit:         "q",
			Help:         "?",
			PlayPause:    "space",
			NextTrack:    "n",
			PrevTrack:    "p",
			VolumeUp:     "+",
			VolumeDown:   "-",
			Mute:         "m",
			Shuffle:      "s",
			Repeat:       "r",
			Search:       "/",
			SeekForward:  "l",
			SeekBackward: "h",
		},
	}

	theme := ui.Rainbow(false)
	prov := &mockProvider{}
	pl := player.New(player.Options{DisableProcess: true})

	m := New(cfg, prov, nil, pl, nil, theme, StartupOptions{}, nil, nil, nil, nil)
	m.width = width
	m.height = height
	m.screen = screenNowPlaying
	m.healthOK = true
	m.healthDetails = "OK"

	return m
}

// TestViewLineCount verifies the View output fits within terminal bounds
func TestViewLineCount(t *testing.T) {
	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"small terminal", 80, 24},
		{"medium terminal", 120, 40},
		{"large terminal", 200, 60},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModelForLayout(tc.width, tc.height)
			view := m.View()

			lines := strings.Split(view, "\n")
			lineCount := len(lines)

			// View should not exceed terminal height
			if lineCount > tc.height {
				t.Errorf("View has %d lines but terminal height is %d (overflow by %d lines)",
					lineCount, tc.height, lineCount-tc.height)
				// Debug: print last few lines on failure
				for i := max(0, lineCount-5); i < lineCount; i++ {
					t.Logf("  Line %d: %q", i, lines[i])
				}
			}
		})
	}
}

// TestPlayerBarAppearsOnce verifies the player bar content appears exactly once
func TestPlayerBarAppearsOnce(t *testing.T) {
	m := newTestModelForLayout(120, 40)
	view := m.View()

	// Count occurrences of player bar hint text (should appear exactly once)
	hintText := "[Space]Play/Pause"
	count := strings.Count(view, hintText)

	if count != 1 {
		t.Errorf("Player bar hint '%s' appears %d times, expected 1", hintText, count)
	}

	// Count occurrences of "(not playing)" - should appear exactly once when nothing is playing
	notPlayingCount := strings.Count(view, "(not playing)")
	if notPlayingCount != 1 {
		t.Errorf("'(not playing)' appears %d times, expected 1", notPlayingCount)
	}
}

// TestViewStructure verifies the basic structure of the View output
func TestViewStructure(t *testing.T) {
	m := newTestModelForLayout(120, 40)
	view := m.View()

	// Should contain the app title
	if !strings.Contains(view, "Tunez") {
		t.Error("View should contain app title 'Tunez'")
	}

	// Should contain navigation items
	navItems := []string{"Now Playing", "Search", "Library", "Queue"}
	for _, item := range navItems {
		if !strings.Contains(view, item) {
			t.Errorf("View should contain nav item '%s'", item)
		}
	}

	// Should contain player bar elements
	if !strings.Contains(view, "Vol:") {
		t.Error("View should contain volume indicator")
	}
}

// TestViewHeightConsistencyAcrossScreens verifies content height is consistent across screens
func TestViewHeightConsistencyAcrossScreens(t *testing.T) {
	screens := []struct {
		name   string
		screen screen
	}{
		{"Now Playing", screenNowPlaying},
		{"Library", screenLibrary},
		{"Search", screenSearch},
		{"Queue", screenQueue},
		{"Config", screenConfig},
	}

	width, height := 120, 40

	for _, s := range screens {
		t.Run(s.name, func(t *testing.T) {
			m := newTestModelForLayout(width, height)
			m.screen = s.screen
			// Add some test data for screens that need it
			if s.screen == screenLibrary {
				m.artists = []provider.Artist{{ID: "1", Name: "Test Artist"}}
			}
			view := m.View()
			lines := strings.Split(view, "\n")

			// All screens should not overflow terminal height
			if len(lines) > height {
				t.Errorf("Screen %s has %d lines, exceeds terminal height %d",
					s.name, len(lines), height)
			}

			// Player bar should appear exactly once
			hintCount := strings.Count(view, "[Space]Play/Pause")
			if hintCount != 1 {
				t.Errorf("Screen %s: Player bar appears %d times, expected 1", s.name, hintCount)
			}

			t.Logf("Screen %s: %d lines", s.name, len(lines))
		})
	}
}

// TestNowPlayingWithTrack verifies Now Playing screen with an active track
func TestNowPlayingWithTrack(t *testing.T) {
	m := newTestModelForLayout(120, 40)
	m.screen = screenNowPlaying
	m.nowPlaying = provider.Track{
		ID:         "track1",
		Title:      "Test Song",
		ArtistName: "Test Artist",
		AlbumTitle: "Test Album",
		DurationMs: 180000, // 3 minutes
	}
	m.duration = 180
	m.timePos = 60

	view := m.View()

	// Should contain track info
	if !strings.Contains(view, "Test Song") {
		t.Error("View should contain track title")
	}
	if !strings.Contains(view, "Test Artist") {
		t.Error("View should contain artist name")
	}
	if !strings.Contains(view, "Test Album") {
		t.Error("View should contain album title")
	}

	// Should NOT contain "(not playing)" when a track is playing
	notPlayingCount := strings.Count(view, "(not playing)")
	if notPlayingCount != 0 {
		t.Errorf("View contains '(not playing)' %d times when track is active, expected 0", notPlayingCount)
	}

	// Line count should not exceed terminal height
	lines := strings.Split(view, "\n")
	if len(lines) > 40 {
		t.Errorf("View has %d lines, exceeds terminal height 40", len(lines))
	}

	// Player bar should show track info (not "not playing")
	// and appear exactly once
	hintCount := strings.Count(view, "[Space]Play/Pause")
	if hintCount != 1 {
		t.Errorf("Player bar appears %d times, expected 1", hintCount)
	}
}

// TestQueueScreenWithManyItems verifies Queue screen doesn't overflow with many items
func TestQueueScreenWithManyItems(t *testing.T) {
	m := newTestModelForLayout(120, 40)
	m.screen = screenQueue

	// Add many tracks to queue
	for i := 0; i < 50; i++ {
		m.queue.Add(provider.Track{
			ID:         string(rune('a' + i)),
			Title:      "Track " + string(rune('A'+i%26)),
			ArtistName: "Artist",
			DurationMs: 180000,
		})
	}

	view := m.View()
	lines := strings.Split(view, "\n")

	if len(lines) > 40 {
		t.Errorf("Queue view has %d lines with 50 items, exceeds terminal height 40", len(lines))
	}

	// Player bar should still appear exactly once
	hintCount := strings.Count(view, "[Space]Play/Pause")
	if hintCount != 1 {
		t.Errorf("Player bar appears %d times in queue view, expected 1", hintCount)
	}
}
