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

// Layout styles
var (
	borderColor    = lipgloss.Color("#7C7CFF")
	accentColor    = lipgloss.Color("#FF6FF7")
	dimColor       = lipgloss.Color("#6C6F93")
	titleColor     = lipgloss.Color("#8EEBFF")
	highlightColor = lipgloss.Color("#FFA7C4")

	topBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(borderColor).
			Padding(0, 1)

	navStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(borderColor).
			Padding(0, 1)

	mainPaneStyle = lipgloss.NewStyle().
			Padding(0, 1)

	playerBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(borderColor).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	navItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	navSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(lipgloss.Color("#3C3C5C")).
				Foreground(titleColor).
				Bold(true)
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
	noEmoji         bool
	healthOK        bool
	healthDetails   string
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

func New(cfg *config.Config, prov provider.Provider, factory ProviderFactory, player *player.Controller, settings any, theme ui.Theme) Model {
	return Model{
		cfg:             cfg,
		provider:        prov,
		factory:         factory,
		player:          player,
		queue:           queue.New(),
		theme:           theme,
		screen:          screenLoading,
		status:          "Loading…",
		profileSettings: settings,
		noEmoji:         cfg.UI.NoEmoji,
		volume:          float64(cfg.Player.InitialVolume),
		healthOK:        true,
		healthDetails:   "OK",
	}
}

type initMsg struct {
	err error
}

type healthMsg struct {
	ok      bool
	details string
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initProviderCmd(), m.watchPlayerCmd(), m.healthCheckCmd())
}

func (m Model) healthCheckCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ok, details := m.provider.Health(ctx)
		return healthMsg{ok: ok, details: details}
	})
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

// matchKey returns true if the key matches the binding.
// Handles both single keys and aliases (e.g., "space" matches " ").
func matchKey(key, binding string) bool {
	if key == binding {
		return true
	}
	// Handle special cases
	if binding == "space" && key == " " {
		return true
	}
	return false
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

type seekMsg struct {
	err error
}

func (m Model) seekCmd(delta float64) tea.Cmd {
	return func() tea.Msg {
		if err := m.player.Seek(delta); err != nil {
			return seekMsg{err: err}
		}
		return seekMsg{}
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
	case healthMsg:
		m.healthOK = msg.ok
		m.healthDetails = msg.details
		return m, m.healthCheckCmd() // Schedule next check
	case seekMsg:
		if msg.err != nil {
			return m.setError(msg.err)
		}
		return m, nil
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
		m.healthOK = true
		m.healthDetails = "OK"
		return m, tea.Batch(m.initProviderCmd(), m.watchPlayerCmd(), m.healthCheckCmd())
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
		key := msg.String()

		// Handle configurable keybindings first (player controls)
		if matchKey(key, m.cfg.Keybindings.Quit) {
			return m, tea.Quit
		}
		if matchKey(key, m.cfg.Keybindings.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if matchKey(key, m.cfg.Keybindings.Mute) {
			m.muted = !m.muted
			return m, func() tea.Msg {
				if err := m.player.SetMute(m.muted); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		}
		if matchKey(key, m.cfg.Keybindings.Shuffle) {
			m.queue.ToggleShuffle()
			return m, nil
		}
		if matchKey(key, m.cfg.Keybindings.Repeat) {
			m.queue.CycleRepeat()
			return m, nil
		}
		if matchKey(key, m.cfg.Keybindings.PlayPause) {
			m.paused = !m.paused
			return m, func() tea.Msg {
				if err := m.player.TogglePause(m.paused); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		}
		if matchKey(key, m.cfg.Keybindings.NextTrack) {
			if t, err := m.queue.Next(); err == nil {
				return m, m.playTrackCmd(t)
			}
			return m, nil
		}
		if matchKey(key, m.cfg.Keybindings.PrevTrack) {
			if t, err := m.queue.Prev(); err == nil {
				return m, m.playTrackCmd(t)
			}
			return m, nil
		}
		if matchKey(key, m.cfg.Keybindings.VolumeDown) {
			m.volume -= float64(m.cfg.Player.VolumeStep)
			if m.volume < 0 {
				m.volume = 0
			}
			return m, func() tea.Msg {
				if err := m.player.SetVolume(m.volume); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		}
		if matchKey(key, m.cfg.Keybindings.VolumeUp) || key == "=" {
			m.volume += float64(m.cfg.Player.VolumeStep)
			if m.volume > 100 {
				m.volume = 100
			}
			return m, func() tea.Msg {
				if err := m.player.SetVolume(m.volume); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		}
		if matchKey(key, m.cfg.Keybindings.Search) {
			m.screen = screenSearch
			m.searchQ = ""
			m.searchResults = provider.SearchResults{}
			m.status = "Enter search query"
			return m, nil
		}

		// Non-configurable keys use switch
		switch key {
		case "H":
			return m, m.seekCmd(float64(-m.cfg.Player.SeekLarge))
		case "L":
			return m, m.seekCmd(float64(m.cfg.Player.SeekLarge))
		case "a":
			if t, ok := m.selectedTrack(); ok {
				return m, m.addTrackCmd(t)
			}
		case "A":
			if t, ok := m.selectedTrack(); ok {
				return m, m.addNextTrackCmd(t)
			}
		case "P":
			if t, ok := m.selectedTrack(); ok {
				return m, m.addNextTrackCmd(t)
			}
		case "tab":
			m.screen = (m.screen + 1) % 8
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
			m.selection = 0
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
			return m, m.seekCmd(float64(-m.cfg.Player.SeekSmall))
		case "l", "right":
			if m.screen == screenLibrary {
				return m.handleEnter()
			}
			return m, m.seekCmd(float64(m.cfg.Player.SeekSmall))
		case "f":
			if m.screen == screenSearch {
				m.searchFilter = (m.searchFilter + 1) % 3
				m.selection = 0
				return m, nil
			}
		case "enter":
			return m.handleEnter()
		case "x":
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
		case "d":
			if m.screen == screenQueue {
				if m.selection < m.queue.Len()-1 {
					_ = m.queue.Move(m.selection, m.selection+1)
					m.selection++
				}
				return m, nil
			}
		case "u":
			if m.screen == screenQueue {
				if m.selection > 0 {
					_ = m.queue.Move(m.selection, m.selection-1)
					m.selection--
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
		case "C":
			if m.screen == screenQueue {
				m.queue.Clear()
				m.selection = 0
				return m, nil
			}
		default:
			if m.screen == screenSearch && len(key) == 1 && msg.Runes != nil {
				m.searchQ += key
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
		if m.queue.Len() > 0 {
			if err := m.queue.SetCurrent(m.selection); err == nil {
				if t, err := m.queue.Current(); err == nil {
					return m, m.playTrackCmd(t)
				}
			}
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
		return m.renderHelpOverlay()
	}

	// Calculate dimensions
	width := m.width
	if width < 80 {
		width = 80
	}
	height := m.height
	if height < 24 {
		height = 24
	}

	// Top bar
	topBar := m.renderTopBar(width)

	// Left navigation (fixed width)
	navWidth := 16
	nav := m.renderNavigation(navWidth, height-6) // Account for top/bottom bars

	// Main content
	mainWidth := width - navWidth - 4 // Account for borders/padding
	var mainContent string
	switch m.screen {
	case screenLoading:
		mainContent = m.renderLoading(mainWidth)
	case screenNowPlaying:
		mainContent = m.renderNowPlaying()
	case screenLibrary:
		mainContent = m.renderLibrary()
	case screenSearch:
		mainContent = m.renderSearch()
	case screenQueue:
		mainContent = m.renderQueue()
	case screenPlaylists:
		mainContent = m.renderPlaylists()
	case screenLyrics:
		mainContent = m.renderLyrics()
	case screenConfig:
		mainContent = m.renderConfig()
	}

	// Apply main pane styling
	mainPane := mainPaneStyle.Width(mainWidth).Render(mainContent)

	// Combine nav and main horizontally
	middle := lipgloss.JoinHorizontal(lipgloss.Top, nav, mainPane)

	// Bottom player bar
	playerBar := m.renderPlayerBar()

	// Status line (if error)
	statusLine := ""
	if m.errorMsg != "" {
		statusLine = m.theme.Error.Render(" ⚠ " + m.errorMsg)
	}

	// Combine all vertically
	if statusLine != "" {
		return lipgloss.JoinVertical(lipgloss.Left, topBar, middle, statusLine, playerBar)
	}
	return lipgloss.JoinVertical(lipgloss.Left, topBar, middle, playerBar)
}

func (m Model) renderTopBar(width int) string {
	// Provider info
	profile, _ := m.cfg.ProfileByID(m.cfg.ActiveProfile)
	providerInfo := fmt.Sprintf("Provider: %s (%s)", profile.Provider, profile.Name)

	// Health status - use actual health check result
	var health string
	if m.healthOK {
		if m.noEmoji {
			health = m.theme.Success.Render("[OK]")
		} else {
			health = m.theme.Success.Render("● OK")
		}
	} else {
		if m.noEmoji {
			health = m.theme.Error.Render("[ERR]")
		} else {
			health = m.theme.Error.Render("● " + m.healthDetails)
		}
	}

	// Queue count
	queueInfo := fmt.Sprintf("Queue: %d", m.queue.Len())

	// Help hint
	helpHint := m.theme.Dim.Render("[? Help]")

	// Build top bar
	left := m.theme.Title.Render("♪ Tunez") + "  " + m.theme.Dim.Render(providerInfo)
	right := health + "  " + queueInfo + "  " + helpHint

	// Calculate spacing
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	spaces := width - leftLen - rightLen - 4
	if spaces < 1 {
		spaces = 1
	}

	bar := left + strings.Repeat(" ", spaces) + right
	return topBarStyle.Width(width).Render(bar)
}

func (m Model) renderNavigation(width, height int) string {
	items := []struct {
		screen screen
		label  string
		icon   string
	}{
		{screenNowPlaying, "Now Playing", "♪"},
		{screenSearch, "Search", "⌕"},
		{screenLibrary, "Library", "≡"},
		{screenQueue, "Queue", "☰"},
	}

	// Add capability-gated items
	caps := m.provider.Capabilities()
	if caps[provider.CapPlaylists] {
		items = append(items, struct {
			screen screen
			label  string
			icon   string
		}{screenPlaylists, "Playlists", "♫"})
	}
	if caps[provider.CapLyrics] {
		items = append(items, struct {
			screen screen
			label  string
			icon   string
		}{screenLyrics, "Lyrics", "¶"})
	}
	items = append(items, struct {
		screen screen
		label  string
		icon   string
	}{screenConfig, "Config", "⚙"})

	var lines []string
	for _, item := range items {
		icon := item.icon
		if m.noEmoji {
			icon = ">"
		}
		label := fmt.Sprintf("%s %s", icon, item.label)

		if item.screen == m.screen {
			lines = append(lines, navSelectedStyle.Render(label))
		} else {
			lines = append(lines, navItemStyle.Render(label))
		}
	}

	content := strings.Join(lines, "\n")
	return navStyle.Width(width).Height(height).Render(content)
}

func (m Model) renderLoading(width int) string {
	var b strings.Builder

	// ASCII art title
	title := `
   ▄▄▄█████▓ █    ██  ███▄    █ ▓█████ ▒███████▒
   ▓  ██▒ ▓▒ ██  ▓██▒ ██ ▀█   █ ▓█   ▀ ▒ ▒ ▒ ▄▀░
   ▒ ▓██░ ▒░▓██  ▒██░▓██  ▀█ ██▒▒███   ░ ▒ ▄▀▒░ 
   ░ ▓██▓ ░ ▓▓█  ░██░▓██▒  ▐▌██▒▒▓█  ▄   ▄▀▒   ░
     ▒██▒ ░ ▒▒█████▓ ▒██░   ▓██░░▒████▒▒███████▒
     ▒ ░░   ░▒▓▒ ▒ ▒ ░ ▒░   ▒ ▒ ░░ ▒░ ░░▒▒ ▓░▒░▒
`
	b.WriteString(m.theme.Accent.Render(title))
	b.WriteString("\n\n")

	// Status lines
	steps := []struct {
		label  string
		status string
	}{
		{"Loading config", "OK"},
		{"Starting mpv", "OK"},
		{"Initializing provider", m.status},
	}

	for _, step := range steps {
		icon := "✓"
		style := m.theme.Success
		if step.status != "OK" {
			icon = "○"
			style = m.theme.Dim
		}
		if m.noEmoji {
			if step.status == "OK" {
				icon = "[OK]"
			} else {
				icon = "[..]"
			}
		}
		line := fmt.Sprintf("  %s  %-30s %s", style.Render(icon), step.label, style.Render(step.status))
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("  Tip: Press ? at any time for help"))

	return b.String()
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

	if m.nowPlaying.Title == "" {
		// Nothing playing state
		b.WriteString(boxStyle.Render(
			lipgloss.JoinVertical(lipgloss.Center,
				"",
				m.theme.Dim.Render("♪ Nothing playing"),
				"",
				m.theme.Dim.Render("Select a track from Library or Search"),
				"",
			),
		))
	} else {
		// Track info box
		trackInfo := lipgloss.JoinVertical(lipgloss.Left,
			m.theme.Dim.Render("Track: ")+m.theme.Accent.Render(m.nowPlaying.Title),
			m.theme.Dim.Render("Artist: ")+m.theme.Text.Render(m.nowPlaying.ArtistName),
			m.theme.Dim.Render("Album: ")+m.theme.Text.Render(m.nowPlaying.AlbumTitle),
		)
		if m.nowPlaying.Codec != "" {
			trackInfo = lipgloss.JoinVertical(lipgloss.Left,
				trackInfo,
				m.theme.Dim.Render(fmt.Sprintf("Codec: %s  |  Bitrate: %dkbps", m.nowPlaying.Codec, m.nowPlaying.BitrateKbps)),
			)
		}
		b.WriteString(boxStyle.Render(trackInfo))
		b.WriteString("\n\n")

		// Visual progress bar
		barWidth := 50
		pct := 0.0
		if m.duration > 0 {
			pct = m.timePos / m.duration
		}
		filled := int(float64(barWidth) * pct)
		empty := barWidth - filled
		if filled < 0 {
			filled = 0
		}
		if empty < 0 {
			empty = 0
		}

		progressBar := m.theme.Highlight.Render(strings.Repeat("▓", filled)) +
			m.theme.Dim.Render(strings.Repeat("░", empty))

		tPos := fmt.Sprintf("%d:%02d", int(m.timePos)/60, int(m.timePos)%60)
		dur := fmt.Sprintf("%d:%02d", int(m.duration)/60, int(m.duration)%60)
		timeStr := fmt.Sprintf("%s / %s", tPos, dur)

		b.WriteString("  " + progressBar + "  " + m.theme.Dim.Render(timeStr) + "\n\n")

		// Visualizer placeholder
		b.WriteString(m.theme.Dim.Render("  Visualizer: "))
		vizBars := "║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║"
		b.WriteString(m.theme.Accent.Render(vizBars) + "\n\n")
	}

	// Up Next section
	b.WriteString(m.theme.Title.Render("Up Next") + "\n")
	upNextCount := 0
	items := m.queue.Items()
	currentIdx := m.queue.CurrentIndex()
	for i := currentIdx + 1; i < len(items) && upNextCount < 5; i++ {
		t := items[i]
		prefix := fmt.Sprintf("  %d) ", upNextCount+1)
		b.WriteString(m.theme.Dim.Render(prefix) + m.theme.Text.Render(fmt.Sprintf("%s — %s", t.ArtistName, t.Title)) + "\n")
		upNextCount++
	}
	if upNextCount == 0 {
		b.WriteString(m.theme.Dim.Render("  (End of queue)") + "\n")
	}

	// Action hints
	b.WriteString("\n" + m.theme.Dim.Render("[Space]Play/Pause [n/p]Next/Prev [h/l]Seek [s]Shuffle [r]Repeat"))

	return b.String()
}

func (m Model) renderLibrary() string {
	var b strings.Builder

	// Determine current view and show header with pagination
	var viewTitle string
	var itemCount, totalCount int

	if len(m.tracks) > 0 {
		viewTitle = "Tracks"
		itemCount = len(m.tracks)
		totalCount = itemCount // Will need provider total when available
	} else if len(m.albums) > 0 {
		viewTitle = "Albums"
		itemCount = len(m.albums)
		totalCount = itemCount
	} else {
		viewTitle = "Artists"
		itemCount = len(m.artists)
		totalCount = itemCount
	}

	// Header with view mode and pagination
	header := fmt.Sprintf("%s", viewTitle)
	if totalCount > 0 {
		header += fmt.Sprintf("  %d/%d", m.selection+1, totalCount)
	}
	b.WriteString(m.theme.Title.Render(header) + "\n")

	// List box content
	var listContent strings.Builder

	if len(m.tracks) > 0 {
		for i, t := range m.tracks {
			prefix := "   "
			style := m.theme.Text
			if i == m.selection {
				prefix = " ▶ "
				style = selectedStyle
			}
			dur := fmt.Sprintf("%d:%02d", t.DurationMs/60000, (t.DurationMs/1000)%60)
			line := fmt.Sprintf("%s%02d  %s — %s  %s", prefix, i+1, t.ArtistName, t.Title, m.theme.Dim.Render(dur))
			listContent.WriteString(style.Render(line) + "\n")
		}
	} else if len(m.albums) > 0 {
		for i, a := range m.albums {
			prefix := " ▢ "
			style := m.theme.Text
			if i == m.selection {
				prefix = " ▣ "
				style = selectedStyle
			}
			line := fmt.Sprintf("%s%s — %s (%d)", prefix, a.Title, a.ArtistName, a.Year)
			listContent.WriteString(style.Render(line) + "\n")
		}
	} else {
		for i, a := range m.artists {
			prefix := " ▢ "
			style := m.theme.Text
			if i == m.selection {
				prefix = " ▣ "
				style = selectedStyle
			}
			albumText := "albums"
			if a.AlbumCount == 1 {
				albumText = "album"
			}
			line := fmt.Sprintf("%s%s  (%d %s)", prefix, a.Name, a.AlbumCount, albumText)
			listContent.WriteString(style.Render(line) + "\n")
		}
	}

	b.WriteString(boxStyle.Render(listContent.String()))
	b.WriteString("\n")

	// Details panel for selected item
	if len(m.albums) > 0 && m.selection < len(m.albums) {
		a := m.albums[m.selection]
		details := fmt.Sprintf("%s (%d)\n%s\nTracks: %d", a.Title, a.Year, a.ArtistName, a.TrackCount)
		b.WriteString("\n" + m.theme.Accent.Render("Details") + "\n")
		b.WriteString(boxStyle.Render(details) + "\n")
	} else if len(m.artists) > 0 && m.selection < len(m.artists) {
		a := m.artists[m.selection]
		details := fmt.Sprintf("%s\nAlbums: %d", a.Name, a.AlbumCount)
		b.WriteString("\n" + m.theme.Accent.Render("Details") + "\n")
		b.WriteString(boxStyle.Render(details) + "\n")
	}

	// Action hints
	b.WriteString("\n" + m.theme.Dim.Render("[Enter]Open/Play  [a]Add to Queue  [A]Play Next  [Backspace]Back"))

	return b.String()
}

func (m Model) renderSearch() string {
	var b strings.Builder

	// Header with query
	header := fmt.Sprintf("Search: %s", m.searchQ)
	if m.searchQ == "" {
		header = "Search: (press / to search)"
	}
	b.WriteString(m.theme.Title.Render(header) + "\n\n")

	// Filters
	filters := []string{"Tracks", "Albums", "Artists"}
	var filterLine strings.Builder
	filterLine.WriteString("Filter: ")
	for i, f := range filters {
		if searchFilter(i) == m.searchFilter {
			filterLine.WriteString(m.theme.Accent.Render("[" + f + "]"))
		} else {
			filterLine.WriteString(m.theme.Dim.Render(" " + f + " "))
		}
		filterLine.WriteString(" ")
	}
	b.WriteString(filterLine.String() + "\n\n")

	// Results with pagination info
	var itemCount int
	switch m.searchFilter {
	case filterTracks:
		itemCount = len(m.searchResults.Tracks.Items)
	case filterAlbums:
		itemCount = len(m.searchResults.Albums.Items)
	case filterArtists:
		itemCount = len(m.searchResults.Artists.Items)
	}

	resultsHeader := fmt.Sprintf("Results (%s)", m.searchFilter)
	if itemCount > 0 {
		resultsHeader += fmt.Sprintf("  %d/%d", m.selection+1, itemCount)
	}
	b.WriteString(m.theme.Accent.Render(resultsHeader) + "\n")

	// Results list in a box
	var listContent strings.Builder

	if m.searchQ == "" || itemCount == 0 {
		if m.searchQ == "" {
			listContent.WriteString(m.theme.Dim.Render("  Enter a search query to find music"))
		} else {
			listContent.WriteString(m.theme.Dim.Render("  No results found"))
		}
	} else {
		switch m.searchFilter {
		case filterTracks:
			for i, t := range m.searchResults.Tracks.Items {
				prefix := "   "
				style := m.theme.Text
				if i == m.selection {
					prefix = " ▶ "
					style = selectedStyle
				}
				dur := fmt.Sprintf("%d:%02d", t.DurationMs/60000, (t.DurationMs/1000)%60)
				line := fmt.Sprintf("%s%02d  %s — %s  %s", prefix, i+1, t.ArtistName, t.Title, m.theme.Dim.Render(dur))
				listContent.WriteString(style.Render(line) + "\n")
			}
		case filterAlbums:
			for i, a := range m.searchResults.Albums.Items {
				prefix := " ▢ "
				style := m.theme.Text
				if i == m.selection {
					prefix = " ▣ "
					style = selectedStyle
				}
				line := fmt.Sprintf("%s%s — %s (%d)", prefix, a.Title, a.ArtistName, a.Year)
				listContent.WriteString(style.Render(line) + "\n")
			}
		case filterArtists:
			for i, a := range m.searchResults.Artists.Items {
				prefix := " ▢ "
				style := m.theme.Text
				if i == m.selection {
					prefix = " ▣ "
					style = selectedStyle
				}
				line := fmt.Sprintf("%s%s", prefix, a.Name)
				listContent.WriteString(style.Render(line) + "\n")
			}
		}
	}

	b.WriteString(boxStyle.Render(listContent.String()))
	b.WriteString("\n\n")

	// Action hints
	b.WriteString(m.theme.Dim.Render("[/]Search  [f]Cycle Filter  [Enter]Play  [a]Add to Queue  [A]Play Next"))

	return b.String()
}

func (m Model) renderQueue() string {
	var b strings.Builder
	items := m.queue.Items()
	currentIdx := m.queue.CurrentIndex()

	// Header with queue stats
	header := fmt.Sprintf("Queue  Items: %d", len(items))

	// Mode indicators
	modeStr := "Normal"
	if m.queue.IsShuffled() {
		modeStr = "Shuffled"
	}
	shuffleStr := "Off"
	if m.queue.IsShuffled() {
		shuffleStr = "On"
	}
	repeatStr := "Off"
	switch m.queue.RepeatMode() {
	case queue.RepeatAll:
		repeatStr = "All"
	case queue.RepeatOne:
		repeatStr = "One"
	}

	header += fmt.Sprintf("   Mode: %s   Shuffle: %s   Repeat: %s", modeStr, shuffleStr, repeatStr)
	b.WriteString(m.theme.Title.Render(header) + "\n\n")

	// Queue list in a box
	var listContent strings.Builder

	if len(items) == 0 {
		listContent.WriteString(m.theme.Dim.Render("  Queue is empty. Add tracks from Library or Search."))
	} else {
		for i, t := range items {
			// Determine prefix based on selection and playing state
			prefix := "    "
			style := m.theme.Text
			isPlaying := i == currentIdx
			isSelected := i == m.selection

			if isPlaying && isSelected {
				prefix = "▶▣ "
				style = selectedStyle
			} else if isPlaying {
				prefix = "▶  "
				style = m.theme.Accent
			} else if isSelected {
				prefix = " ▣ "
				style = selectedStyle
			}

			dur := fmt.Sprintf("%d:%02d", t.DurationMs/60000, (t.DurationMs/1000)%60)
			line := fmt.Sprintf("%s%02d  %s — %s  %s", prefix, i+1, t.ArtistName, t.Title, m.theme.Dim.Render(dur))
			listContent.WriteString(style.Render(line) + "\n")
		}
	}

	b.WriteString(boxStyle.Render(listContent.String()))
	b.WriteString("\n\n")

	// Action hints
	b.WriteString(m.theme.Dim.Render("[Enter]Play  [x]Remove  [C]Clear  [u/d]Move Up/Down  [P]Play Next"))

	return b.String()
}

func (m Model) renderPlaylists() string {
	var b strings.Builder

	// Header with pagination
	header := "Playlists"
	if len(m.playlists) > 0 {
		header += fmt.Sprintf("  %d/%d", m.selection+1, len(m.playlists))
	}
	b.WriteString(m.theme.Title.Render(header) + "\n\n")

	// Playlists list in a box
	var listContent strings.Builder

	if len(m.playlists) == 0 {
		listContent.WriteString(m.theme.Dim.Render("  No playlists available"))
	} else {
		for i, p := range m.playlists {
			prefix := " ▢ "
			style := m.theme.Text
			if i == m.selection {
				prefix = " ▣ "
				style = selectedStyle
			}
			trackText := "tracks"
			if p.TrackCount == 1 {
				trackText = "track"
			}
			line := fmt.Sprintf("%s%s  (%d %s)", prefix, p.Name, p.TrackCount, trackText)
			listContent.WriteString(style.Render(line) + "\n")
		}
	}

	b.WriteString(boxStyle.Render(listContent.String()))
	b.WriteString("\n")

	// Selected playlist details
	if len(m.playlists) > 0 && m.selection < len(m.playlists) {
		p := m.playlists[m.selection]
		b.WriteString("\n" + m.theme.Accent.Render("Playlist Details") + "\n")

		details := fmt.Sprintf("%s\nTracks: %d", p.Name, p.TrackCount)
		b.WriteString(boxStyle.Render(details) + "\n")
	}
	b.WriteString("\n")

	// Action hints
	b.WriteString(m.theme.Dim.Render("[Enter]Open  [A]Add All to Queue  [p]Play Playlist"))

	return b.String()
}

func (m Model) renderLyrics() string {
	var b strings.Builder

	// Header with track info
	trackInfo := "(no track)"
	if m.nowPlaying.Title != "" {
		trackInfo = fmt.Sprintf("%s — %s", m.nowPlaying.ArtistName, m.nowPlaying.Title)
	}
	b.WriteString(m.theme.Title.Render("Lyrics") + "  " + m.theme.Dim.Render(trackInfo) + "\n\n")

	// Lyrics content in a box
	var lyricsContent string

	// Check if provider supports lyrics
	caps := m.provider.Capabilities()
	if !caps[provider.CapLyrics] {
		lyricsContent = m.theme.Dim.Render("  Lyrics not supported by this provider")
	} else if m.nowPlaying.Title == "" {
		lyricsContent = m.theme.Dim.Render("  No track playing")
	} else {
		// TODO: Fetch and display actual lyrics from provider
		lyricsContent = m.theme.Dim.Render("  No lyrics available for this track")
	}

	b.WriteString(boxStyle.Render(lyricsContent))
	b.WriteString("\n\n")

	// Action hints
	b.WriteString(m.theme.Dim.Render("[j/k]Scroll  [g/G]Top/Bottom"))

	return b.String()
}

func (m Model) renderConfig() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render("Config") + "\n\n")

	// Config sections list
	sections := []struct {
		name string
		desc string
	}{
		{"Providers & Profiles", "Manage provider connections and profiles"},
		{"Theme & ANSI", "Appearance settings"},
		{"Keybindings", "View keybindings (view only)"},
		{"Cache / Offline", "Cache settings and status"},
		{"Logging & Diagnostics", "Log settings and debug info"},
	}

	b.WriteString(m.theme.Accent.Render("Sections") + "\n")
	var sectionsContent strings.Builder
	for i, s := range sections {
		prefix := " ▢ "
		style := m.theme.Text
		if i == m.selection {
			prefix = " ▣ "
			style = selectedStyle
		}
		sectionsContent.WriteString(style.Render(prefix+s.name) + "\n")
	}
	b.WriteString(boxStyle.Render(sectionsContent.String()))
	b.WriteString("\n\n")

	// Details panel based on selection
	b.WriteString(m.theme.Accent.Render("Details") + "\n")
	var detailsContent strings.Builder

	switch m.selection {
	case 0: // Providers & Profiles
		activeProfile := m.cfg.ActiveProfile
		for _, p := range m.cfg.Profiles {
			if p.ID == activeProfile {
				detailsContent.WriteString(fmt.Sprintf("Active Provider: %s\n", p.Provider))
				detailsContent.WriteString(fmt.Sprintf("Profile: %s\n", p.Name))
			}
		}
		detailsContent.WriteString(fmt.Sprintf("Total Profiles: %d\n", len(m.cfg.Profiles)))

		// Provider capabilities
		caps := m.provider.Capabilities()
		capList := []string{}
		if caps[provider.CapPlaylists] {
			capList = append(capList, "Playlists")
		}
		if caps[provider.CapLyrics] {
			capList = append(capList, "Lyrics")
		}
		if caps[provider.CapArtwork] {
			capList = append(capList, "Artwork")
		}
		if len(capList) > 0 {
			detailsContent.WriteString(fmt.Sprintf("Capabilities: %s", strings.Join(capList, ", ")))
		}

	case 1: // Theme & ANSI
		detailsContent.WriteString(fmt.Sprintf("Theme: %s\n", m.cfg.UI.Theme))
		noEmoji := "No"
		if m.cfg.UI.NoEmoji {
			noEmoji = "Yes"
		}
		detailsContent.WriteString(fmt.Sprintf("No Emoji: %s\n", noEmoji))
		detailsContent.WriteString(fmt.Sprintf("Page Size: %d", m.cfg.UI.PageSize))

	case 2: // Keybindings
		detailsContent.WriteString("Navigation: j/k, Tab/Shift+Tab\n")
		detailsContent.WriteString("Player: Space, n/p, h/l, +/-\n")
		detailsContent.WriteString("Queue: x, C, u/d, P\n")
		detailsContent.WriteString("Help: ?")

	case 3: // Cache / Offline
		detailsContent.WriteString("Cache: Not configured\n")
		detailsContent.WriteString("(MVP: view only)")

	case 4: // Logging & Diagnostics
		detailsContent.WriteString(fmt.Sprintf("MPV Path: %s\n", m.cfg.Player.MPVPath))
		detailsContent.WriteString(fmt.Sprintf("Seek Small: %ds\n", m.cfg.Player.SeekSmall))
		detailsContent.WriteString(fmt.Sprintf("Seek Large: %ds\n", m.cfg.Player.SeekLarge))
		detailsContent.WriteString(fmt.Sprintf("Volume Step: %d%%", m.cfg.Player.VolumeStep))
	}

	b.WriteString(boxStyle.Render(detailsContent.String()))
	b.WriteString("\n\n")

	// Footer hint
	b.WriteString(m.theme.Dim.Render("Config file: ~/.config/tunez/config.toml"))
	b.WriteString("\n")
	b.WriteString(m.theme.Dim.Render("[Enter]Open Section  [Esc]Back"))

	return b.String()
}

func (m Model) renderHelpOverlay() string {
	lines := []string{
		m.theme.Accent.Render("Global"),
		"  tab/shift+tab : Switch screens",
		"  ?             : Toggle help",
		"  ctrl+c        : Quit",
		"",
		m.theme.Accent.Render("Player"),
		"  space         : Play/Pause",
		"  n / p         : Next / Previous track",
		"  h / l         : Seek -5s / +5s",
		"  H / L         : Seek -30s / +30s",
		"  - / +         : Volume Down / Up",
		"  m             : Mute",
		"  s             : Toggle Shuffle",
		"  r             : Cycle Repeat (off/all/one)",
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
		"  x             : Remove item",
		"  u / d         : Move item up / down",
		"  C             : Clear queue",
		"  P             : Play next (add after current)",
		"",
		m.theme.Accent.Render("Library"),
		"  a             : Add to queue",
		"  A             : Add to queue (play next)",
		"",
		m.theme.Dim.Render("Press ? or Esc to close"),
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.Title.Render("  ═══ Help / Keybindings ═══  "),
		"",
		strings.Join(lines, "\n"),
	)

	// Put in a styled box and center it
	helpBox := boxStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpBox)
}

func (m Model) renderPlayerBar() string {
	// Play state icon
	state := "⏵"
	if m.paused {
		state = "⏸"
	}
	if m.noEmoji {
		if m.paused {
			state = "||"
		} else {
			state = ">"
		}
	}

	// Track info
	name := "(not playing)"
	if m.nowPlaying.Title != "" {
		name = fmt.Sprintf("%s — %s", m.nowPlaying.ArtistName, m.nowPlaying.Title)
	}

	// Time and visual progress bar
	var timeAndProgress string
	if m.duration > 0 {
		tPos := fmt.Sprintf("%d:%02d", int(m.timePos)/60, int(m.timePos)%60)
		dur := fmt.Sprintf("%d:%02d", int(m.duration)/60, int(m.duration)%60)

		// Visual progress bar
		barWidth := 20
		pct := m.timePos / m.duration
		if pct > 1 {
			pct = 1
		}
		filled := int(float64(barWidth) * pct)
		empty := barWidth - filled
		bar := strings.Repeat("▓", filled) + strings.Repeat("░", empty)

		timeAndProgress = fmt.Sprintf("[%s/%s] %s", tPos, dur, bar)
	}

	// Volume
	volStr := fmt.Sprintf("Vol: %.0f%%", m.volume)
	if m.muted {
		volStr = "Muted"
	}

	// Shuffle/Repeat indicators
	shuffle := ""
	if m.queue.IsShuffled() {
		if m.noEmoji {
			shuffle = " [Shuf]"
		} else {
			shuffle = " 🔀"
		}
	}
	repeat := ""
	switch m.queue.RepeatMode() {
	case queue.RepeatAll:
		if m.noEmoji {
			repeat = " [Rep:All]"
		} else {
			repeat = " 🔁"
		}
	case queue.RepeatOne:
		if m.noEmoji {
			repeat = " [Rep:One]"
		} else {
			repeat = " 🔂"
		}
	}

	// First line: track info
	line1 := fmt.Sprintf("%s  %s  %s  %s%s%s", state, name, timeAndProgress, volStr, shuffle, repeat)
	// Second line: action hints
	line2 := m.theme.Dim.Render("[Space]Play/Pause [n/p]Next/Prev [h/l]Seek [+/-]Vol [s]Shuffle [r]Repeat [?]Help")

	return line1 + "\n" + line2
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
	case screenPlaylists:
		return "Playlists"
	case screenLyrics:
		return "Lyrics"
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
	case screenPlaylists:
		return len(m.playlists)
	case screenConfig:
		return 5 // Number of config sections
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
