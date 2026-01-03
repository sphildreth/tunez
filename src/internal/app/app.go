package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tunez/tunez/internal/config"
	"github.com/tunez/tunez/internal/player"
	"github.com/tunez/tunez/internal/provider"
	"github.com/tunez/tunez/internal/queue"
	"github.com/tunez/tunez/internal/ui"
)

type screen int

const (
	screenLoading screen = iota
	screenNowPlaying
	screenLibrary
	screenSearch
	screenQueue
	screenPlaylists
	screenLyrics
	screenConfig
)

type ProviderFactory func(config.Profile) (provider.Provider, error)

type Model struct {
	cfg      *config.Config
	provider provider.Provider
	factory  ProviderFactory
	player   *player.Controller
	queue    *queue.Queue
	theme    ui.Theme

	screen          screen
	status          string
	errorMsg        string
	fatalErr        error
	artists         []provider.Artist
	artistsCursor   string
	albums          []provider.Album
	albumsCursor    string
	tracks          []provider.Track
	tracksCursor    string
	playlists       []provider.Playlist
	playlistsCursor string
	currentArtistID string
	currentAlbumID  string
	searchQ         string
	searchResults   provider.SearchResults
	searchFilter    searchFilter
	selection       int
	width           int
	height          int
	showHelp        bool
	nowPlaying      provider.Track
	paused          bool
	timePos         float64
	duration        float64
	volume          float64
	muted           bool
	profileSettings any
}

type searchFilter int

const (
	filterTracks searchFilter = iota
	filterAlbums
	filterArtists
)

func (f searchFilter) String() string {
	switch f {
	case filterTracks:
		return "Tracks"
	case filterAlbums:
		return "Albums"
	case filterArtists:
		return "Artists"
	default:
		return "Unknown"
	}
}

func New(cfg *config.Config, prov provider.Provider, factory ProviderFactory, player *player.Controller, settings any) Model {
	return Model{
		cfg:             cfg,
		provider:        prov,
		factory:         factory,
		player:          player,
		queue:           queue.New(),
		theme:           ui.Rainbow(cfg.UI.NoEmoji),
		screen:          screenLoading,
		status:          "Loading…",
		profileSettings: settings,
	}
}

type initMsg struct {
	err error
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initProviderCmd(), m.watchPlayerCmd())
}

func (m Model) initProviderCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := m.provider.Initialize(ctx, m.profileSettings); err != nil {
			return initMsg{err: err}
		}
		// Load initial data
		page, err := m.provider.ListArtists(ctx, provider.ListReq{PageSize: m.cfg.UI.PageSize})
		return artistsMsg{page: page, err: err}
	}
}

func (m Model) loadArtistsCmd(cursor string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		page, err := m.provider.ListArtists(ctx, provider.ListReq{PageSize: m.cfg.UI.PageSize, Cursor: cursor})
		return artistsMsg{page: page, err: err}
	}
}

func (m Model) loadAlbumsCmd(artistID, cursor string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		page, err := m.provider.ListAlbums(ctx, artistID, provider.ListReq{PageSize: m.cfg.UI.PageSize, Cursor: cursor})
		return albumsMsg{page: page, err: err}
	}
}

func (m Model) loadTracksCmd(artistID, albumID, cursor string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		page, err := m.provider.ListTracks(ctx, albumID, artistID, "", provider.ListReq{PageSize: m.cfg.UI.PageSize, Cursor: cursor})
		return tracksMsg{page: page, err: err}
	}
}

func (m Model) loadPlaylistsCmd(cursor string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		page, err := m.provider.ListPlaylists(ctx, provider.ListReq{PageSize: m.cfg.UI.PageSize, Cursor: cursor})
		return playlistsMsg{page: page, err: err}
	}
}

func (m Model) loadPlaylistTracksCmd(playlistID, cursor string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		page, err := m.provider.ListTracks(ctx, "", "", playlistID, provider.ListReq{PageSize: m.cfg.UI.PageSize, Cursor: cursor})
		return tracksMsg{page: page, err: err}
	}
}

func (m Model) searchCmd(q string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		res, err := m.provider.Search(ctx, q, provider.ListReq{PageSize: m.cfg.UI.PageSize})
		return searchMsg{res: res, err: err}
	}
}

func (m Model) watchPlayerCmd() tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-m.player.Events()
		if !ok {
			return nil
		}
		return playerMsg(evt)
	}
}

type artistsMsg struct {
	page provider.Page[provider.Artist]
	err  error
}

type albumsMsg struct {
	page provider.Page[provider.Album]
	err  error
}

type tracksMsg struct {
	page provider.Page[provider.Track]
	err  error
}

type playlistsMsg struct {
	page provider.Page[provider.Playlist]
	err  error
}

type searchMsg struct {
	res provider.SearchResults
	err error
}

type playerMsg player.Event

type playTrackMsg struct {
	track provider.Track
	err   error
}

type searchMoreMsg struct {
	res provider.SearchResults
	err error
}

func (m Model) searchMoreCmd(q, cursor string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		res, err := m.provider.Search(ctx, q, provider.ListReq{Cursor: cursor, PageSize: m.cfg.UI.PageSize})
		return searchMoreMsg{res: res, err: err}
	}
}

type profileSwitchedMsg struct {
	provider provider.Provider
	profile  config.Profile
}

func (m Model) switchProfileCmd(profile config.Profile) tea.Cmd {
	return func() tea.Msg {
		newProv, err := m.factory(profile)
		if err != nil {
			return initMsg{err: err}
		}
		_ = m.player.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := m.player.Start(ctx); err != nil {
			return initMsg{err: err}
		}
		return profileSwitchedMsg{provider: newProv, profile: profile}
	}
}

type clearErrorMsg struct{}

func (m Model) clearErrorCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m Model) setError(err error) (Model, tea.Cmd) {
	m.errorMsg = err.Error()
	return m, m.clearErrorCmd()
}

type addTrackMsg struct {
	track provider.Track
}

type addNextTrackMsg struct {
	track provider.Track
}

func (m Model) addTrackCmd(track provider.Track) tea.Cmd {
	return func() tea.Msg {
		return addTrackMsg{track: track}
	}
}

func (m Model) addNextTrackCmd(track provider.Track) tea.Cmd {
	return func() tea.Msg {
		return addNextTrackMsg{track: track}
	}
}

func (m Model) selectedTrack() (provider.Track, bool) {
	if m.screen == screenLibrary && len(m.tracks) > 0 {
		idx := clamp(m.selection, 0, len(m.tracks)-1)
		return m.tracks[idx], true
	}
	if m.screen == screenSearch && m.searchFilter == filterTracks && len(m.searchResults.Tracks.Items) > 0 {
		idx := clamp(m.selection, 0, len(m.searchResults.Tracks.Items)-1)
		return m.searchResults.Tracks.Items[idx], true
	}
	return provider.Track{}, false
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case addTrackMsg:
		m.queue.Add(msg.track)
		m.status = "Added to queue: " + msg.track.Title
		return m, nil
	case addNextTrackMsg:
		m.queue.AddNext(msg.track)
		m.status = "Playing next: " + msg.track.Title
		return m, nil
	case profileSwitchedMsg:
		m.provider = msg.provider
		m.cfg.ActiveProfile = msg.profile.ID
		m.profileSettings = msg.profile.Settings
		m.queue.Clear()
		m.tracks = nil
		m.albums = nil
		m.artists = nil
		m.playlists = nil
		m.searchResults = provider.SearchResults{}
		m.status = "Profile switched"
		return m, tea.Batch(m.initProviderCmd(), m.watchPlayerCmd())
	case clearErrorMsg:
		m.errorMsg = ""
		return m, nil
	case initMsg:
		if msg.err != nil {
			m.fatalErr = msg.err
			m.status = "Init failed"
		} else {
			m.status = "Ready"
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "m":
			m.muted = !m.muted
			return m, func() tea.Msg {
				if err := m.player.SetMute(m.muted); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		case "s":
			m.queue.ToggleShuffle()
			return m, nil
		case "r":
			m.queue.CycleRepeat()
			return m, nil
		case "H":
			if err := m.player.Seek(float64(-m.cfg.Player.SeekLarge)); err != nil {
				return m.setError(err)
			}
			return m, nil
		case "L":
			if err := m.player.Seek(float64(m.cfg.Player.SeekLarge)); err != nil {
				return m.setError(err)
			}
			return m, nil
		case "a":
			if t, ok := m.selectedTrack(); ok {
				return m, m.addTrackCmd(t)
			}
		case "A":
			if t, ok := m.selectedTrack(); ok {
				return m, m.addNextTrackCmd(t)
			}
		case "tab":
			m.screen = (m.screen + 1) % 7
			if m.screen == screenLoading {
				m.screen++
			}
			// Skip screens if capability missing
			if m.screen == screenPlaylists && !m.provider.Capabilities()[provider.CapPlaylists] {
				m.screen++
			}
			if m.screen == screenLyrics && !m.provider.Capabilities()[provider.CapLyrics] {
				m.screen++
			}
			if m.screen == screenPlaylists && len(m.playlists) == 0 {
				return m, m.loadPlaylistsCmd("")
			}
			return m, nil
		case "shift+tab":
			m.screen--
			if m.screen == screenLyrics && !m.provider.Capabilities()[provider.CapLyrics] {
				m.screen--
			}
			if m.screen == screenPlaylists && !m.provider.Capabilities()[provider.CapPlaylists] {
				m.screen--
			}
			if m.screen <= screenLoading {
				m.screen = screenConfig
			}
			return m, nil
		case "j", "down":
			if m.selection < m.currentListLen()-1 {
				m.selection++
			} else if m.screen == screenSearch {
				var nextCursor string
				switch m.searchFilter {
				case filterTracks:
					nextCursor = m.searchResults.Tracks.NextCursor
				case filterAlbums:
					nextCursor = m.searchResults.Albums.NextCursor
				case filterArtists:
					nextCursor = m.searchResults.Artists.NextCursor
				}
				if nextCursor != "" {
					return m, m.searchMoreCmd(m.searchQ, nextCursor)
				}
			} else if m.screen == screenLibrary {
				if len(m.tracks) > 0 && m.tracksCursor != "" {
					return m, m.loadTracksCmd(m.currentArtistID, m.currentAlbumID, m.tracksCursor)
				}
				if len(m.albums) > 0 && m.albumsCursor != "" {
					return m, m.loadAlbumsCmd(m.currentArtistID, m.albumsCursor)
				}
				if len(m.artists) > 0 && m.artistsCursor != "" {
					return m, m.loadArtistsCmd(m.artistsCursor)
				}
			}
			return m, nil
		case "k", "up":
			if m.selection > 0 {
				m.selection--
			}
			return m, nil
		case "h", "left", "backspace":
			if m.screen == screenLibrary {
				if len(m.tracks) > 0 {
					m.tracks = nil
					m.tracksCursor = ""
					m.currentAlbumID = ""
					m.selection = 0
					m.status = "Albums"
					return m, nil
				}
				if len(m.albums) > 0 {
					m.albums = nil
					m.albumsCursor = ""
					m.currentArtistID = ""
					m.selection = 0
					m.status = "Artists"
					return m, nil
				}
			}
			// Seeking for other screens
			if err := m.player.Seek(float64(-m.cfg.Player.SeekSmall)); err != nil {
				return m.setError(err)
			}
		case "l", "right":
			if m.screen == screenLibrary {
				return m.handleEnter()
			}
			if err := m.player.Seek(float64(m.cfg.Player.SeekSmall)); err != nil {
				return m.setError(err)
			}
		case "/":
			m.screen = screenSearch
			m.searchQ = ""
			m.searchResults = provider.SearchResults{}
			m.status = "Enter search query"
			return m, nil
		case "f":
			if m.screen == screenSearch {
				m.searchFilter = (m.searchFilter + 1) % 3
				m.selection = 0
				return m, nil
			}
		case "enter":
			return m.handleEnter()
		case " ":
			m.paused = !m.paused
			return m, tea.Batch(func() tea.Msg {
				if err := m.player.TogglePause(m.paused); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			})
		case "d":
			if m.screen == screenQueue {
				if err := m.queue.Remove(m.selection); err == nil {
					if m.selection >= m.queue.Len() {
						m.selection = m.queue.Len() - 1
					}
					if m.selection < 0 {
						m.selection = 0
					}
				}
				return m, nil
			}
		case "J":
			if m.screen == screenQueue {
				if m.selection < m.queue.Len()-1 {
					_ = m.queue.Move(m.selection, m.selection+1)
					m.selection++
				}
				return m, nil
			}
		case "K":
			if m.screen == screenQueue {
				if m.selection > 0 {
					_ = m.queue.Move(m.selection, m.selection-1)
					m.selection--
				}
				return m, nil
			}
		case "c":
			if m.screen == screenQueue {
				m.queue.Clear()
				m.selection = 0
				return m, nil
			}
		case "n":
			if t, err := m.queue.Next(); err == nil {
				return m, m.playTrackCmd(t)
			}
		case "p":
			if t, err := m.queue.Prev(); err == nil {
				return m, m.playTrackCmd(t)
			}

		case "-":
			m.volume -= float64(m.cfg.Player.VolumeStep)
			return m, func() tea.Msg {
				if err := m.player.SetVolume(m.volume); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		case "+":
			m.volume += float64(m.cfg.Player.VolumeStep)
			return m, func() tea.Msg {
				if err := m.player.SetVolume(m.volume); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		default:
			if m.screen == screenSearch && len(msg.String()) == 1 && msg.Runes != nil {
				m.searchQ += msg.String()
				return m, m.searchCmd(m.searchQ)
			}
		}
	case artistsMsg:
		if msg.err != nil {
			return m.setError(msg.err)
		} else {
			if m.artistsCursor == "" {
				m.artists = msg.page.Items
			} else {
				m.artists = append(m.artists, msg.page.Items...)
			}
			m.artistsCursor = msg.page.NextCursor
			m.status = fmt.Sprintf("Artists loaded (%d)", len(m.artists))
			if m.screen == screenLoading {
				m.screen = screenNowPlaying
			}
		}
	case albumsMsg:
		if msg.err != nil {
			return m.setError(msg.err)
		} else {
			if m.albumsCursor == "" {
				m.albums = msg.page.Items
			} else {
				m.albums = append(m.albums, msg.page.Items...)
			}
			m.albumsCursor = msg.page.NextCursor
			m.tracks = nil
			m.selection = 0
			m.status = fmt.Sprintf("Albums loaded (%d)", len(m.albums))
		}
	case tracksMsg:
		if msg.err != nil {
			return m.setError(msg.err)
		} else {
			if m.tracksCursor == "" {
				m.tracks = msg.page.Items
			} else {
				m.tracks = append(m.tracks, msg.page.Items...)
			}
			m.tracksCursor = msg.page.NextCursor
			m.status = fmt.Sprintf("Tracks loaded (%d)", len(m.tracks))
		}
	case playlistsMsg:
		if msg.err != nil {
			return m.setError(msg.err)
		} else {
			if m.playlistsCursor == "" {
				m.playlists = msg.page.Items
			} else {
				m.playlists = append(m.playlists, msg.page.Items...)
			}
			m.playlistsCursor = msg.page.NextCursor
			m.status = fmt.Sprintf("Playlists loaded (%d)", len(m.playlists))
		}
	case searchMsg:
		if msg.err != nil {
			return m.setError(msg.err)
		} else {
			m.searchResults = msg.res
			count := len(msg.res.Tracks.Items) + len(msg.res.Albums.Items) + len(msg.res.Artists.Items)
			m.status = fmt.Sprintf("Found %d results", count)
		}
	case searchMoreMsg:
		if msg.err != nil {
			return m.setError(msg.err)
		} else {
			if len(msg.res.Tracks.Items) > 0 {
				m.searchResults.Tracks.Items = append(m.searchResults.Tracks.Items, msg.res.Tracks.Items...)
				m.searchResults.Tracks.NextCursor = msg.res.Tracks.NextCursor
			}
			if len(msg.res.Albums.Items) > 0 {
				m.searchResults.Albums.Items = append(m.searchResults.Albums.Items, msg.res.Albums.Items...)
				m.searchResults.Albums.NextCursor = msg.res.Albums.NextCursor
			}
			if len(msg.res.Artists.Items) > 0 {
				m.searchResults.Artists.Items = append(m.searchResults.Artists.Items, msg.res.Artists.Items...)
				m.searchResults.Artists.NextCursor = msg.res.Artists.NextCursor
			}
			m.status = "Loaded more results"
		}
	case playTrackMsg:
		if msg.err != nil {
			return m.setError(msg.err)
		} else {
			m.nowPlaying = msg.track
			m.paused = false
			m.queue.Add(msg.track)
			m.status = "Playing " + msg.track.Title
		}
	case playerMsg:
		if msg.TimePos != nil {
			m.timePos = *msg.TimePos
		}
		if msg.Duration != nil {
			m.duration = *msg.Duration
		}
		if msg.Volume != nil {
			m.volume = *msg.Volume
		}
		if msg.Paused != nil {
			m.paused = *msg.Paused
		}
		if msg.Muted != nil {
			m.muted = *msg.Muted
		}
		if msg.Err != nil {
			return m.setError(msg.Err)
		}
		if msg.Ended {
			if t, err := m.queue.Next(); err == nil {
				return m, m.playTrackCmd(t)
			}
		}
		return m, m.watchPlayerCmd()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenLibrary:
		if len(m.tracks) > 0 {
			idx := clamp(m.selection, 0, len(m.tracks)-1)
			track := m.tracks[idx]
			return m, m.playTrackCmd(track)
		}
		if len(m.albums) > 0 {
			idx := clamp(m.selection, 0, len(m.albums)-1)
			album := m.albums[idx]
			m.currentAlbumID = album.ID
			m.currentArtistID = album.ArtistID
			return m, m.loadTracksCmd(album.ArtistID, album.ID, "")
		}
		if len(m.artists) > 0 {
			idx := clamp(m.selection, 0, len(m.artists)-1)
			artist := m.artists[idx]
			m.currentArtistID = artist.ID
			return m, m.loadAlbumsCmd(artist.ID, "")
		}
	case screenSearch:
		switch m.searchFilter {
		case filterTracks:
			if len(m.searchResults.Tracks.Items) > 0 {
				idx := clamp(m.selection, 0, len(m.searchResults.Tracks.Items)-1)
				track := m.searchResults.Tracks.Items[idx]
				return m, m.playTrackCmd(track)
			}
		case filterAlbums:
			if len(m.searchResults.Albums.Items) > 0 {
				idx := clamp(m.selection, 0, len(m.searchResults.Albums.Items)-1)
				album := m.searchResults.Albums.Items[idx]
				m.screen = screenLibrary
				m.currentAlbumID = album.ID
				m.currentArtistID = album.ArtistID
				return m, m.loadTracksCmd(album.ArtistID, album.ID, "")
			}
		case filterArtists:
			if len(m.searchResults.Artists.Items) > 0 {
				idx := clamp(m.selection, 0, len(m.searchResults.Artists.Items)-1)
				artist := m.searchResults.Artists.Items[idx]
				m.screen = screenLibrary
				m.currentArtistID = artist.ID
				return m, m.loadAlbumsCmd(artist.ID, "")
			}
		}
	case screenQueue:
		if t, err := m.queue.Current(); err == nil {
			return m, m.playTrackCmd(t)
		}
	case screenConfig:
		if len(m.cfg.Profiles) > 0 {
			idx := clamp(m.selection, 0, len(m.cfg.Profiles)-1)
			profile := m.cfg.Profiles[idx]
			if profile.ID != m.cfg.ActiveProfile && profile.Enabled {
				m.status = "Switching profile..."
				return m, m.switchProfileCmd(profile)
			}
		}
	}
	return m, nil
}

func (m Model) playTrackCmd(track provider.Track) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		stream, err := m.provider.GetStream(ctx, track.ID)
		if err != nil {
			return playTrackMsg{err: err}
		}
		if err := m.player.Play(stream.URL, stream.Headers); err != nil {
			return playTrackMsg{err: err}
		}
		return playTrackMsg{track: track}
	}
}

func (m Model) View() string {
	if m.fatalErr != nil {
		return m.renderFatalError()
	}
	if m.showHelp {
		return m.renderHelp()
	}
	var main string
	switch m.screen {
	case screenLoading:
		main = m.theme.Title.Render("Loading… " + m.status)
	case screenNowPlaying:
		main = m.renderNowPlaying()
	case screenLibrary:
		main = m.renderLibrary()
	case screenSearch:
		main = m.renderSearch()
	case screenQueue:
		main = m.renderQueue()
	case screenConfig:
		main = m.renderConfig()
	}
	top := lipgloss.NewStyle().Bold(true).Render("Tunez ▸ " + m.screenTitle())
	status := m.theme.Dim.Render(m.status)
	if m.errorMsg != "" {
		status = m.theme.Error.Render(m.errorMsg)
	}
	bottom := m.renderPlayerBar()
	return lipgloss.JoinVertical(lipgloss.Left, top, main, status, bottom)
}

func (m Model) renderFatalError() string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.theme.Border.Render(
			lipgloss.JoinVertical(lipgloss.Center,
				m.theme.Error.Render("Fatal Error"),
				"",
				m.theme.Text.Render(m.fatalErr.Error()),
				"",
				m.theme.Dim.Render("Press Ctrl+C to quit"),
			),
		),
	)
}

func (m Model) renderNowPlaying() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render("Now Playing") + "\n\n")

	if m.nowPlaying.Title == "" {
		b.WriteString(m.theme.Dim.Render("Nothing playing") + "\n")
	} else {
		b.WriteString(m.theme.Accent.Render(m.nowPlaying.Title) + "\n")
		b.WriteString(m.theme.Text.Render(m.nowPlaying.ArtistName) + "\n")
		if m.nowPlaying.AlbumTitle != "" {
			b.WriteString(m.theme.Dim.Render(m.nowPlaying.AlbumTitle) + "\n")
		}
		b.WriteString("\n")

		// Progress bar
		width := m.width - 4
		if width < 10 {
			width = 10
		}
		pct := 0.0
		if m.duration > 0 {
			pct = m.timePos / m.duration
		}
		filled := int(float64(width) * pct)
		empty := width - filled
		if filled < 0 {
			filled = 0
		}
		if empty < 0 {
			empty = 0
		}

		bar := strings.Repeat("━", filled) + strings.Repeat("─", empty)
		b.WriteString(m.theme.Highlight.Render(bar) + "\n")

		// Time
		tPos := fmt.Sprintf("%d:%02d", int(m.timePos)/60, int(m.timePos)%60)
		dur := fmt.Sprintf("%d:%02d", int(m.duration)/60, int(m.duration)%60)
		b.WriteString(m.theme.Dim.Render(fmt.Sprintf("%s / %s", tPos, dur)) + "\n")
	}

	b.WriteString("\n" + m.theme.Title.Render("Up Next") + "\n")
	if next, err := m.queue.PeekNext(); err == nil {
		b.WriteString(m.theme.Text.Render(fmt.Sprintf("%s — %s", next.ArtistName, next.Title)) + "\n")
	} else {
		b.WriteString(m.theme.Dim.Render("(End of queue)") + "\n")
	}

	return b.String()
}

func (m Model) renderLibrary() string {
	var b strings.Builder
	if len(m.tracks) > 0 {
		b.WriteString(m.theme.Title.Render("Tracks\n"))
		for i, t := range m.tracks {
			prefix := "  "
			if i == m.selection {
				prefix = "⏵ "
			}
			dur := fmt.Sprintf("%d:%02d", t.DurationMs/60000, (t.DurationMs/1000)%60)
			line := fmt.Sprintf("%s%s — %s (%s)", prefix, t.ArtistName, t.Title, dur)
			b.WriteString(m.theme.Text.Render(line) + "\n")
		}
		return b.String()
	}
	if len(m.albums) > 0 {
		b.WriteString(m.theme.Title.Render("Albums\n"))
		for i, a := range m.albums {
			prefix := "  "
			if i == m.selection {
				prefix = "⏵ "
			}
			line := fmt.Sprintf("%s%s (%d)", prefix, a.Title, a.Year)
			b.WriteString(m.theme.Text.Render(line) + "\n")
		}
		return b.String()
	}
	b.WriteString(m.theme.Title.Render("Artists\n"))
	for i, a := range m.artists {
		prefix := "  "
		if i == m.selection {
			prefix = "⏵ "
		}
		b.WriteString(prefix + m.theme.Text.Render(fmt.Sprintf("%s (%d albums)", a.Name, a.AlbumCount)) + "\n")
	}
	return b.String()
}

func (m Model) renderSearch() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render(fmt.Sprintf("Search (%s): %s\n", m.searchFilter, m.searchQ)))

	switch m.searchFilter {
	case filterTracks:
		for i, t := range m.searchResults.Tracks.Items {
			prefix := "  "
			if i == m.selection {
				prefix = "⏵ "
			}
			b.WriteString(prefix + fmt.Sprintf("%s — %s\n", t.ArtistName, t.Title))
		}
	case filterAlbums:
		for i, a := range m.searchResults.Albums.Items {
			prefix := "  "
			if i == m.selection {
				prefix = "⏵ "
			}
			b.WriteString(prefix + fmt.Sprintf("%s — %s (%d)\n", a.ArtistName, a.Title, a.Year))
		}
	case filterArtists:
		for i, a := range m.searchResults.Artists.Items {
			prefix := "  "
			if i == m.selection {
				prefix = "⏵ "
			}
			b.WriteString(prefix + fmt.Sprintf("%s\n", a.Name))
		}
	}
	return b.String()
}

func (m Model) renderQueue() string {
	var b strings.Builder
	items := m.queue.Items()
	currentIdx := m.queue.CurrentIndex()
	b.WriteString(m.theme.Title.Render("Queue\n"))
	for i, t := range items {
		prefix := "   "
		if i == currentIdx {
			prefix = "🔊 "
		}
		if i == m.selection {
			if i == currentIdx {
				prefix = "⏵🔊"
			} else {
				prefix = "⏵  "
			}
		}
		b.WriteString(prefix + fmt.Sprintf("%d. %s — %s\n", i+1, t.ArtistName, t.Title))
	}
	return b.String()
}

func (m Model) renderPlaylists() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render("Playlists\n"))
	for i, p := range m.playlists {
		prefix := "  "
		if i == m.selection {
			prefix = "⏵ "
		}
		b.WriteString(prefix + fmt.Sprintf("%s (%d tracks)\n", p.Name, p.TrackCount))
	}
	return b.String()
}

func (m Model) renderLyrics() string {
	return m.theme.Title.Render("Lyrics") + "\n\n" + m.theme.Dim.Render("No lyrics available.")
}

func (m Model) renderConfig() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render("Config") + "\n\n")

	b.WriteString(m.theme.Title.Render("Profiles") + "\n")
	for i, p := range m.cfg.Profiles {
		prefix := "  "
		if i == m.selection {
			prefix = "⏵ "
		}
		active := ""
		if p.ID == m.cfg.ActiveProfile {
			active = " (active)"
		}
		line := fmt.Sprintf("%s%s%s", prefix, p.Name, active)
		if !p.Enabled {
			line += " [disabled]"
		}
		b.WriteString(m.theme.Text.Render(line) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Theme: %s\n", m.cfg.UI.Theme))
	b.WriteString(fmt.Sprintf("MPV path: %s\n", m.cfg.Player.MPVPath))
	return b.String()
}

func (m Model) renderHelp() string {
	lines := []string{
		m.theme.Title.Render("Help"),
		"",
		m.theme.Accent.Render("Global"),
		"  tab/shift+tab : Switch screens",
		"  ?             : Toggle help",
		"  ctrl+c        : Quit",
		"",
		m.theme.Accent.Render("Player"),
		"  space         : Play/Pause",
		"  n / p         : Next / Previous track",
		"  h / l         : Seek -5s / +5s",
		"  - / +         : Volume Down / Up",
		"",
		m.theme.Accent.Render("Navigation"),
		"  j / k         : Move selection down / up",
		"  enter         : Select / Play / Drill down",
		"  backspace     : Go back (Library)",
		"",
		m.theme.Accent.Render("Search"),
		"  /             : Enter search mode",
		"  f             : Cycle filter (Tracks/Albums/Artists)",
		"",
		m.theme.Accent.Render("Queue"),
		"  d             : Remove item",
		"  J / K         : Move item down / up",
		"  c             : Clear queue",
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderPlayerBar() string {
	name := "(stopped)"
	if m.nowPlaying.Title != "" {
		name = fmt.Sprintf("%s — %s", m.nowPlaying.ArtistName, m.nowPlaying.Title)
	}
	state := "⏵"
	if m.paused {
		state = "⏸"
	}
	progress := ""
	if m.duration > 0 {
		progress = fmt.Sprintf(" %.0f/%.0fs", m.timePos, m.duration)
	}

	shuffle := ""
	if m.queue.IsShuffled() {
		shuffle = " 🔀"
	}
	repeat := ""
	switch m.queue.RepeatMode() {
	case queue.RepeatAll:
		repeat = " 🔁"
	case queue.RepeatOne:
		repeat = " 🔂"
	}

	volStr := fmt.Sprintf("Vol: %.0f%%", m.volume)
	if m.muted {
		volStr = "Muted"
	}

	return fmt.Sprintf("%s %s%s  %s%s%s", state, name, progress, volStr, shuffle, repeat)
}

func (m Model) screenTitle() string {
	switch m.screen {
	case screenLoading:
		return "Loading"
	case screenNowPlaying:
		return "Now Playing"
	case screenLibrary:
		return "Library"
	case screenSearch:
		return "Search"
	case screenQueue:
		return "Queue"
	case screenConfig:
		return "Config"
	default:
		return ""
	}
}

func (m Model) currentListLen() int {
	switch m.screen {
	case screenLibrary:
		if len(m.tracks) > 0 {
			return len(m.tracks)
		}
		if len(m.albums) > 0 {
			return len(m.albums)
		}
		return len(m.artists)
	case screenSearch:
		switch m.searchFilter {
		case filterTracks:
			return len(m.searchResults.Tracks.Items)
		case filterAlbums:
			return len(m.searchResults.Albums.Items)
		case filterArtists:
			return len(m.searchResults.Artists.Items)
		}
		return 0
	case screenQueue:
		return m.queue.Len()
	default:
		return 0
	}
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
