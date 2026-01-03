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
	screenLibrary
	screenSearch
	screenQueue
	screenConfig
)

type Model struct {
	cfg      *config.Config
	provider provider.Provider
	player   *player.Controller
	queue    *queue.Queue
	theme    ui.Theme

	screen     screen
	status     string
	err        error
	artists    []provider.Artist
	tracks     []provider.Track
	searchQ    string
	searchRes  []provider.Track
	selection  int
	width      int
	height     int
	showHelp   bool
	nowPlaying provider.Track
	paused     bool
	timePos    float64
	duration   float64
	volume     float64
}

func New(cfg *config.Config, prov provider.Provider, player *player.Controller) Model {
	return Model{
		cfg:      cfg,
		provider: prov,
		player:   player,
		queue:    queue.New(),
		theme:    ui.Rainbow(cfg.UI.NoEmoji),
		screen:   screenLoading,
		status:   "Loading…",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadArtistsCmd(), m.watchPlayerCmd())
}

func (m Model) loadArtistsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		page, err := m.provider.ListArtists(ctx, provider.ListReq{PageSize: m.cfg.UI.PageSize})
		return artistsMsg{page: page, err: err}
	}
}

func (m Model) loadTracksCmd(artistID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		page, err := m.provider.ListTracks(ctx, "", artistID, "", provider.ListReq{PageSize: m.cfg.UI.PageSize})
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

type tracksMsg struct {
	page provider.Page[provider.Track]
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "tab":
			m.screen = (m.screen + 1) % 5
			return m, nil
		case "shift+tab":
			if m.screen == 0 {
				m.screen = 4
			} else {
				m.screen--
			}
			return m, nil
		case "j", "down":
			if m.selection < m.currentListLen()-1 {
				m.selection++
			}
			return m, nil
		case "k", "up":
			if m.selection > 0 {
				m.selection--
			}
			return m, nil
		case "/":
			m.screen = screenSearch
			m.searchQ = ""
			m.status = "Enter search query"
			return m, nil
		case "enter":
			return m.handleEnter()
		case " ":
			m.paused = !m.paused
			return m, tea.Batch(func() tea.Msg {
				_ = m.player.TogglePause(m.paused)
				return nil
			})
		case "n":
			if t, err := m.queue.Next(); err == nil {
				return m, m.playTrackCmd(t)
			}
		case "p":
			if t, err := m.queue.Prev(); err == nil {
				return m, m.playTrackCmd(t)
			}
		case "h":
			_ = m.player.Seek(float64(-m.cfg.Player.SeekSmall))
		case "l":
			_ = m.player.Seek(float64(m.cfg.Player.SeekSmall))
		case "-":
			m.volume -= float64(m.cfg.Player.VolumeStep)
			return m, func() tea.Msg {
				_ = m.player.SetVolume(m.volume)
				return nil
			}
		case "+":
			m.volume += float64(m.cfg.Player.VolumeStep)
			return m, func() tea.Msg {
				_ = m.player.SetVolume(m.volume)
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
			m.err = msg.err
			m.status = msg.err.Error()
		} else {
			m.artists = msg.page.Items
			m.status = "Artists loaded"
			m.screen = screenLibrary
		}
	case tracksMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = msg.err.Error()
		} else {
			m.tracks = msg.page.Items
			m.status = "Tracks loaded"
		}
	case searchMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = msg.err.Error()
		} else {
			m.searchRes = msg.res.Tracks.Items
			m.status = fmt.Sprintf("Found %d tracks", len(m.searchRes))
		}
	case playTrackMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = msg.err.Error()
		} else {
			m.nowPlaying = msg.track
			m.paused = false
			m.queue.Add(msg.track)
			m.status = "Playing " + msg.track.Title
		}
	case playerMsg:
		m.timePos = msg.TimePos
		if msg.Duration > 0 {
			m.duration = msg.Duration
		}
		if msg.Volume > 0 {
			m.volume = msg.Volume
		}
		if msg.Err != nil {
			m.err = msg.Err
			m.status = msg.Err.Error()
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

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenLibrary:
		if len(m.tracks) > 0 {
			idx := clamp(m.selection, 0, len(m.tracks)-1)
			track := m.tracks[idx]
			return m, m.playTrackCmd(track)
		}
		if len(m.artists) > 0 {
			idx := clamp(m.selection, 0, len(m.artists)-1)
			artist := m.artists[idx]
			return m, m.loadTracksCmd(artist.ID)
		}
	case screenSearch:
		if len(m.searchRes) > 0 {
			idx := clamp(m.selection, 0, len(m.searchRes)-1)
			track := m.searchRes[idx]
			return m, m.playTrackCmd(track)
		}
	case screenQueue:
		if t, err := m.queue.Current(); err == nil {
			return m, m.playTrackCmd(t)
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
	if m.showHelp {
		return m.renderHelp()
	}
	var main string
	switch m.screen {
	case screenLoading:
		main = m.theme.Title.Render("Loading… " + m.status)
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
	bottom := m.renderPlayerBar()
	return lipgloss.JoinVertical(lipgloss.Left, top, main, status, bottom)
}

func (m Model) renderLibrary() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render("Artists\n"))
	for i, a := range m.artists {
		prefix := "  "
		if i == m.selection {
			prefix = "⏵ "
		}
		b.WriteString(prefix + m.theme.Text.Render(a.Name) + "\n")
	}
	if len(m.tracks) > 0 {
		b.WriteString(m.theme.Title.Render("\nTracks\n"))
		for i, t := range m.tracks {
			prefix := "  "
			if i == m.selection {
				prefix = "⏵ "
			}
			line := fmt.Sprintf("%s%s — %s", prefix, t.ArtistName, t.Title)
			b.WriteString(m.theme.Text.Render(line) + "\n")
		}
	}
	return b.String()
}

func (m Model) renderSearch() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render("Search: " + m.searchQ + "\n"))
	for i, t := range m.searchRes {
		prefix := "  "
		if i == m.selection {
			prefix = "⏵ "
		}
		b.WriteString(prefix + fmt.Sprintf("%s — %s\n", t.ArtistName, t.Title))
	}
	return b.String()
}

func (m Model) renderQueue() string {
	var b strings.Builder
	items := m.queue.Items()
	b.WriteString(m.theme.Title.Render("Queue\n"))
	for i, t := range items {
		prefix := "  "
		if i == m.selection {
			prefix = "⏵ "
		}
		b.WriteString(prefix + fmt.Sprintf("%d. %s — %s\n", i+1, t.ArtistName, t.Title))
	}
	return b.String()
}

func (m Model) renderConfig() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render("Config\n"))
	b.WriteString(fmt.Sprintf("Active profile: %s\n", m.cfg.ActiveProfile))
	b.WriteString(fmt.Sprintf("Theme: %s\n", m.cfg.UI.Theme))
	b.WriteString(fmt.Sprintf("MPV path: %s\n", m.cfg.Player.MPVPath))
	return b.String()
}

func (m Model) renderHelp() string {
	lines := []string{
		"Controls:",
		" j/k: navigate   enter: select/play",
		" space: play/pause   n/p: next/prev   h/l: seek",
		" /: search   tab/shift+tab: switch views",
		" ?: toggle help   ctrl+c: quit",
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
	return fmt.Sprintf("%s %s%s  Vol: %.0f%%", state, name, progress, m.volume)
}

func (m Model) screenTitle() string {
	switch m.screen {
	case screenLoading:
		return "Loading"
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
		return len(m.artists)
	case screenSearch:
		return len(m.searchRes)
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
