package app

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunez/tunez/internal/config"
	"github.com/tunez/tunez/internal/player"
	"github.com/tunez/tunez/internal/provider"
	"github.com/tunez/tunez/internal/ui"
)

type mockProvider struct {
	artists []provider.Artist
	albums  []provider.Album
	tracks  []provider.Track
}

func (m *mockProvider) ID() string                          { return "mock" }
func (m *mockProvider) Name() string                        { return "Mock" }
func (m *mockProvider) Capabilities() provider.Capabilities { return nil }
func (m *mockProvider) Initialize(ctx context.Context, cfg any) error {
	return nil
}
func (m *mockProvider) Health(ctx context.Context) (bool, string) { return true, "ok" }
func (m *mockProvider) ListArtists(ctx context.Context, req provider.ListReq) (provider.Page[provider.Artist], error) {
	return provider.Page[provider.Artist]{Items: m.artists}, nil
}
func (m *mockProvider) GetArtist(ctx context.Context, id string) (provider.Artist, error) {
	return provider.Artist{}, nil
}
func (m *mockProvider) ListAlbums(ctx context.Context, artistId string, req provider.ListReq) (provider.Page[provider.Album], error) {
	return provider.Page[provider.Album]{Items: m.albums}, nil
}
func (m *mockProvider) GetAlbum(ctx context.Context, id string) (provider.Album, error) {
	return provider.Album{}, nil
}
func (m *mockProvider) ListTracks(ctx context.Context, albumId string, artistId string, playlistId string, req provider.ListReq) (provider.Page[provider.Track], error) {
	return provider.Page[provider.Track]{Items: m.tracks}, nil
}
func (m *mockProvider) GetTrack(ctx context.Context, id string) (provider.Track, error) {
	return provider.Track{}, nil
}
func (m *mockProvider) Search(ctx context.Context, q string, req provider.ListReq) (provider.SearchResults, error) {
	return provider.SearchResults{}, nil
}
func (m *mockProvider) ListPlaylists(ctx context.Context, req provider.ListReq) (provider.Page[provider.Playlist], error) {
	return provider.Page[provider.Playlist]{}, nil
}
func (m *mockProvider) GetPlaylist(ctx context.Context, id string) (provider.Playlist, error) {
	return provider.Playlist{}, nil
}
func (m *mockProvider) GetStream(ctx context.Context, trackId string) (provider.StreamInfo, error) {
	return provider.StreamInfo{URL: "mock://stream"}, nil
}
func (m *mockProvider) GetLyrics(ctx context.Context, trackId string) (provider.Lyrics, error) {
	return provider.Lyrics{}, nil
}
func (m *mockProvider) GetArtwork(ctx context.Context, ref string, sizePx int) (provider.Artwork, error) {
	return provider.Artwork{}, nil
}

func TestNavigation(t *testing.T) {
	cfg := &config.Config{
		UI: config.UIConfig{Theme: "rainbow"},
		Player: config.PlayerConfig{
			SeekSmall:  5,
			VolumeStep: 5,
		},
		Queue: config.QueueConfig{Persist: false},
	}
	prov := &mockProvider{
		artists: []provider.Artist{{ID: "1", Name: "Artist 1"}},
		albums:  []provider.Album{{ID: "10", Title: "Album 1", ArtistID: "1", Year: 2000}},
		tracks:  []provider.Track{{ID: "100", Title: "Track 1", AlbumID: "10", ArtistID: "1"}},
	}
	// Mock player that doesn't start process
	pl := player.New(player.Options{DisableProcess: true})
	theme := ui.Rainbow(false)

	m := New(cfg, prov, func(p config.Profile) (provider.Provider, error) {
		return prov, nil
	}, pl, nil, theme, StartupOptions{}, nil, nil)

	// 1. Initial State
	if m.screen != screenLoading {
		t.Errorf("expected loading screen, got %d", m.screen)
	}

	// 2. Simulate Init success
	// We can't easily run the Cmd produced by Init() because it calls m.provider which is safe,
	// but we want to simulate the RESULT of that cmd.
	// The Init() returns a batch with initProviderCmd.
	// Let's manually trigger the initMsg.
	m, _ = updateModel(m, initMsg{err: nil})
	if m.status != "Ready" {
		t.Errorf("expected status Ready, got %s", m.status)
	}

	// 3. Simulate Artist Load (which is part of Init chain usually)
	m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})
	if m.screen != screenNowPlaying {
		t.Errorf("expected NowPlaying screen, got %d", m.screen)
	}

	// Switch to Library screen for testing navigation
	m.screen = screenLibrary
	if len(m.artists) != 1 {
		t.Errorf("expected 1 artist, got %d", len(m.artists))
	}

	// 4. Select Artist -> Enter -> Albums
	// Selection is 0 by default.
	m, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	_ = cmd // This would be the loadAlbumsCmd

	// Simulate Albums loaded
	m, _ = updateModel(m, albumsMsg{page: provider.Page[provider.Album]{Items: prov.albums}})
	if len(m.albums) != 1 {
		t.Errorf("expected 1 album, got %d", len(m.albums))
	}
	if m.status != "Albums loaded (1)" {
		t.Errorf("expected Albums status, got %s", m.status)
	}

	// 5. Select Album -> Enter -> Tracks
	m, cmd = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	_ = cmd
	m, _ = updateModel(m, tracksMsg{page: provider.Page[provider.Track]{Items: prov.tracks}})
	if len(m.tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(m.tracks))
	}

	// 6. Navigation Back -> Albums
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyBackspace})
	if len(m.tracks) != 0 {
		t.Error("tracks should be cleared")
	}
	if len(m.albums) == 0 {
		t.Error("albums should be present")
	}

	// 7. Navigation Back -> Artists
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyLeft})
	if len(m.albums) != 0 {
		t.Error("albums should be cleared")
	}
	if len(m.artists) == 0 {
		t.Error("artists should be present")
	}
}

func updateModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	nm, cmd := m.Update(msg)
	return nm.(Model), cmd
}

func TestNextPrevDoesNotAddToQueue(t *testing.T) {
	cfg := &config.Config{
		UI: config.UIConfig{Theme: "rainbow"},
		Player: config.PlayerConfig{
			SeekSmall:  5,
			VolumeStep: 5,
		},
		Queue: config.QueueConfig{Persist: false},
		Keybindings: config.KeybindConfig{
			NextTrack: "n",
			PrevTrack: "N",
		},
	}
	tracks := []provider.Track{
		{ID: "1", Title: "Track 1", ArtistID: "a1", AlbumID: "al1"},
		{ID: "2", Title: "Track 2", ArtistID: "a1", AlbumID: "al1"},
		{ID: "3", Title: "Track 3", ArtistID: "a1", AlbumID: "al1"},
	}
	prov := &mockProvider{
		artists: []provider.Artist{{ID: "a1", Name: "Artist 1"}},
		albums:  []provider.Album{{ID: "al1", Title: "Album 1", ArtistID: "a1"}},
		tracks:  tracks,
	}
	pl := player.New(player.Options{DisableProcess: true})
	theme := ui.Rainbow(false)

	m := New(cfg, prov, func(p config.Profile) (provider.Provider, error) {
		return prov, nil
	}, pl, nil, theme, StartupOptions{}, nil, nil)

	// Initialize model
	m, _ = updateModel(m, initMsg{err: nil})
	m, _ = updateModel(m, artistsMsg{page: provider.Page[provider.Artist]{Items: prov.artists}})

	// Add tracks to queue
	m.queue.Add(tracks...)
	initialQueueLen := m.queue.Len()

	if initialQueueLen != 3 {
		t.Fatalf("expected initial queue length 3, got %d", initialQueueLen)
	}

	// Current should be at index 0
	if m.queue.CurrentIndex() != 0 {
		t.Errorf("expected current index 0, got %d", m.queue.CurrentIndex())
	}

	// Press 'n' (next track) multiple times
	for i := 0; i < 5; i++ {
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	}

	// Queue length should remain the same
	if m.queue.Len() != initialQueueLen {
		t.Errorf("queue length changed after next: expected %d, got %d", initialQueueLen, m.queue.Len())
	}

	// Press 'N' (prev track) multiple times
	for i := 0; i < 5; i++ {
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	}

	// Queue length should still be the same
	if m.queue.Len() != initialQueueLen {
		t.Errorf("queue length changed after prev: expected %d, got %d", initialQueueLen, m.queue.Len())
	}

	// Verify queue items are unchanged
	items := m.queue.Items()
	for i, track := range tracks {
		if items[i].ID != track.ID {
			t.Errorf("queue item %d changed: expected %s, got %s", i, track.ID, items[i].ID)
		}
	}
}
