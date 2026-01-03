package app

import (
	"bytes"
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/tunez/tunez/internal/config"
	"github.com/tunez/tunez/internal/player"
	"github.com/tunez/tunez/internal/provider"
	"github.com/tunez/tunez/internal/ui"
)

// testProvider extends mockProvider with richer test data
type testProvider struct {
	mockProvider
}

func newTestProvider() *testProvider {
	return &testProvider{
		mockProvider: mockProvider{
			artists: []provider.Artist{
				{ID: "1", Name: "The Beatles", AlbumCount: 12},
				{ID: "2", Name: "Pink Floyd", AlbumCount: 15},
				{ID: "3", Name: "Led Zeppelin", AlbumCount: 9},
				{ID: "4", Name: "Queen", AlbumCount: 15},
				{ID: "5", Name: "David Bowie", AlbumCount: 26},
			},
			albums: []provider.Album{
				{ID: "10", Title: "Abbey Road", ArtistID: "1", ArtistName: "The Beatles", Year: 1969, TrackCount: 17},
				{ID: "11", Title: "Let It Be", ArtistID: "1", ArtistName: "The Beatles", Year: 1970, TrackCount: 12},
			},
			tracks: []provider.Track{
				{ID: "100", Title: "Come Together", AlbumID: "10", ArtistID: "1", ArtistName: "The Beatles", DurationMs: 259000},
				{ID: "101", Title: "Something", AlbumID: "10", ArtistID: "1", ArtistName: "The Beatles", DurationMs: 183000},
				{ID: "102", Title: "Here Comes the Sun", AlbumID: "10", ArtistID: "1", ArtistName: "The Beatles", DurationMs: 185000},
			},
		},
	}
}

func (p *testProvider) Search(ctx context.Context, q string, req provider.ListReq) (provider.SearchResults, error) {
	// Return mock search results based on query
	if q == "" {
		return provider.SearchResults{}, nil
	}
	return provider.SearchResults{
		Tracks: provider.Page[provider.Track]{Items: p.tracks},
		Albums: provider.Page[provider.Album]{Items: p.albums},
		Artists: provider.Page[provider.Artist]{Items: p.artists},
	}, nil
}

// createTestModel creates a Model suitable for teatest
func createTestModel(t *testing.T) Model {
	t.Helper()
	
	cfg := &config.Config{
		UI: config.UIConfig{Theme: "rainbow"},
		Player: config.PlayerConfig{
			SeekSmall:  5,
			VolumeStep: 5,
		},
		Keybindings: config.KeybindConfig{
			PlayPause:    "space",
			NextTrack:    "n",
			PrevTrack:    "N",
			SeekForward:  "l",
			SeekBackward: "h",
			VolumeUp:     "+",
			VolumeDown:   "-",
			Mute:         "m",
			Shuffle:      "S",
			Repeat:       "r",
			Help:         "?",
			Quit:         "q",
		},
	}
	prov := newTestProvider()
	pl := player.New(player.Options{DisableProcess: true})
	theme := ui.Rainbow(false)

	m := New(cfg, prov, func(p config.Profile) (provider.Provider, error) {
		return prov, nil
	}, pl, nil, theme, StartupOptions{}, nil)

	return m
}

// initializeModel simulates the app initialization sequence
func initializeModel(m Model, prov *testProvider) Model {
	// Simulate successful init
	m, _ = updateModel(m, initMsg{err: nil})
	// Simulate artists loaded
	m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
	return m
}

// TestScreensGolden tests each screen renders correctly using golden files
func TestScreensGolden(t *testing.T) {
	prov := newTestProvider()
	
	tests := []struct {
		name   string
		setup  func(m Model) Model
	}{
		{
			name: "now_playing_empty",
			setup: func(m Model) Model {
				m, _ = updateModel(m, initMsg{err: nil})
				m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
				m.screen = screenNowPlaying
				return m
			},
		},
		{
			name: "library_artists",
			setup: func(m Model) Model {
				m, _ = updateModel(m, initMsg{err: nil})
				m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
				m.screen = screenLibrary
				return m
			},
		},
		{
			name: "library_albums",
			setup: func(m Model) Model {
				m, _ = updateModel(m, initMsg{err: nil})
				m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
				m.screen = screenLibrary
				// Select first artist and load albums
				m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
				m, _ = updateModel(m, albumsMsg{page: provider.Page[provider.Album]{Items: prov.albums}})
				return m
			},
		},
		{
			name: "library_tracks",
			setup: func(m Model) Model {
				m, _ = updateModel(m, initMsg{err: nil})
				m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
				m.screen = screenLibrary
				m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
				m, _ = updateModel(m, albumsMsg{page: provider.Page[provider.Album]{Items: prov.albums}})
				m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
				m, _ = updateModel(m, tracksMsg{page: provider.Page[provider.Track]{Items: prov.tracks}})
				return m
			},
		},
		{
			name: "queue_empty",
			setup: func(m Model) Model {
				m, _ = updateModel(m, initMsg{err: nil})
				m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
				m.screen = screenQueue
				return m
			},
		},
		{
			name: "search_empty",
			setup: func(m Model) Model {
				m, _ = updateModel(m, initMsg{err: nil})
				m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
				m.screen = screenSearch
				return m
			},
		},
		{
			name: "config_screen",
			setup: func(m Model) Model {
				m, _ = updateModel(m, initMsg{err: nil})
				m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
				m.screen = screenConfig
				return m
			},
		},
		{
			name: "help_overlay",
			setup: func(m Model) Model {
				m, _ = updateModel(m, initMsg{err: nil})
				m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
				m.screen = screenNowPlaying
				m.showHelp = true
				return m
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel(t)
			m = tt.setup(m)
			
			// Render the view
			output := m.View()
			
			// Compare against golden file
			teatest.RequireEqualOutput(t, []byte(output))
		})
	}
}

// TestInteractiveNavigation tests full interactive session
func TestInteractiveNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping interactive test in short mode")
	}

	m := createTestModel(t)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 40))

	// Sequence of actions to test
	actions := []struct {
		name string
		key  tea.KeyMsg
		wait time.Duration
	}{
		{"down_to_library", tea.KeyMsg{Type: tea.KeyDown}, 100 * time.Millisecond},
		{"select_down", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 50 * time.Millisecond},
		{"select_down2", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 50 * time.Millisecond},
		{"select_up", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, 50 * time.Millisecond},
		{"open_help", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, 100 * time.Millisecond},
		{"close_help", tea.KeyMsg{Type: tea.KeyEscape}, 50 * time.Millisecond},
		{"next_screen", tea.KeyMsg{Type: tea.KeyDown}, 50 * time.Millisecond},
		{"quit", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, 50 * time.Millisecond},
	}

	for _, action := range actions {
		t.Logf("Action: %s", action.name)
		tm.Send(action.key)
		time.Sleep(action.wait)
	}

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

// TestViewOutput captures and validates screen output
func TestViewOutput(t *testing.T) {
	m := createTestModel(t)
	prov := newTestProvider()
	m = initializeModel(m, prov)
	
	screens := []struct {
		name     string
		screen   screen
		contains []string
	}{
		{
			name:     "NowPlaying",
			screen:   screenNowPlaying,
			contains: []string{"Now Playing", "Nothing playing"},
		},
		{
			name:     "Library",
			screen:   screenLibrary,
			contains: []string{"Artists", "The Beatles", "Pink Floyd"},
		},
		{
			name:     "Queue",
			screen:   screenQueue,
			contains: []string{"Queue", "empty"},
		},
		{
			name:     "Search",
			screen:   screenSearch,
			contains: []string{"Search"},
		},
		{
			name:     "Config",
			screen:   screenConfig,
			contains: []string{"Config", "Providers"},
		},
	}

	for _, sc := range screens {
		t.Run(sc.name, func(t *testing.T) {
			m.screen = sc.screen
			output := m.View()
			
			for _, expected := range sc.contains {
				if !bytes.Contains([]byte(output), []byte(expected)) {
					t.Errorf("screen %s: expected to contain %q\nGot:\n%s", sc.name, expected, output)
				}
			}
		})
	}
}

// TestKeyboardShortcuts validates keybinding behavior
func TestKeyboardShortcuts(t *testing.T) {
	m := createTestModel(t)
	prov := newTestProvider()
	m = initializeModel(m, prov)
	m.screen = screenLibrary
	m.focusedPane = paneContent // j/k work in content pane

	tests := []struct {
		name     string
		key      tea.KeyMsg
		validate func(t *testing.T, m Model)
	}{
		{
			name: "j_moves_selection_down",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			validate: func(t *testing.T, m Model) {
				if m.selection != 1 {
					t.Errorf("expected selection 1, got %d", m.selection)
				}
			},
		},
		{
			name: "k_moves_selection_up",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			validate: func(t *testing.T, m Model) {
				if m.selection != 0 {
					t.Errorf("expected selection 0, got %d", m.selection)
				}
			},
		},
		{
			name: "question_opens_help",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			validate: func(t *testing.T, m Model) {
				if !m.showHelp {
					t.Error("expected help to be shown")
				}
			},
		},
		{
			name: "tab_switches_pane",
			key:  tea.KeyMsg{Type: tea.KeyTab},
			validate: func(t *testing.T, m Model) {
				if m.focusedPane != paneNav {
					t.Errorf("expected paneNav, got %d", m.focusedPane)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testModel := m // copy for isolation
			testModel, _ = updateModel(testModel, tt.key)
			tt.validate(t, testModel)
		})
	}
}
