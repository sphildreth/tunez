package app

import tea "github.com/charmbracelet/bubbletea"

// Command represents an action that can be invoked via the command palette.
type Command struct {
	ID          string
	Name        string
	Description string
	Category    string
	Keybinding  string
	Handler     func(m *Model) (Model, tea.Cmd)
}

// CommandRegistry holds all available commands.
type CommandRegistry struct {
	commands []Command
}

// NewCommandRegistry creates a registry with all available commands.
func NewCommandRegistry(m *Model) *CommandRegistry {
	r := &CommandRegistry{}

	// Navigation commands
	r.register(Command{
		ID:          "nav.now_playing",
		Name:        "Go to Now Playing",
		Description: "Show the now playing screen",
		Category:    "Navigation",
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.screen = screenNowPlaying
			return *m, nil
		},
	})
	r.register(Command{
		ID:          "nav.library",
		Name:        "Go to Library",
		Description: "Browse your music library",
		Category:    "Navigation",
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.screen = screenLibrary
			return *m, nil
		},
	})
	r.register(Command{
		ID:          "nav.queue",
		Name:        "Go to Queue",
		Description: "View and manage the play queue",
		Category:    "Navigation",
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.screen = screenQueue
			return *m, nil
		},
	})
	r.register(Command{
		ID:          "nav.search",
		Name:        "Go to Search",
		Description: "Search for tracks, albums, or artists",
		Category:    "Navigation",
		Keybinding:  m.cfg.Keybindings.Search,
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.screen = screenSearch
			return *m, nil
		},
	})
	r.register(Command{
		ID:          "nav.lyrics",
		Name:        "Go to Lyrics",
		Description: "View lyrics for the current track",
		Category:    "Navigation",
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.screen = screenLyrics
			return *m, nil
		},
	})
	r.register(Command{
		ID:          "nav.config",
		Name:        "Go to Config",
		Description: "View and edit settings",
		Category:    "Navigation",
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.screen = screenConfig
			return *m, nil
		},
	})

	// Playback commands
	r.register(Command{
		ID:          "playback.play_pause",
		Name:        "Play/Pause",
		Description: "Toggle playback",
		Category:    "Playback",
		Keybinding:  m.cfg.Keybindings.PlayPause,
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.paused = !m.paused
			return *m, func() tea.Msg {
				if err := m.player.TogglePause(m.paused); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		},
	})
	r.register(Command{
		ID:          "playback.next",
		Name:        "Next Track",
		Description: "Skip to the next track",
		Category:    "Playback",
		Keybinding:  m.cfg.Keybindings.NextTrack,
		Handler: func(m *Model) (Model, tea.Cmd) {
			next, err := m.queue.Next()
			if err != nil {
				return *m, nil
			}
			return *m, m.playTrackCmd(next)
		},
	})
	r.register(Command{
		ID:          "playback.prev",
		Name:        "Previous Track",
		Description: "Go to the previous track",
		Category:    "Playback",
		Keybinding:  m.cfg.Keybindings.PrevTrack,
		Handler: func(m *Model) (Model, tea.Cmd) {
			prev, err := m.queue.Prev()
			if err != nil {
				return *m, nil
			}
			return *m, m.playTrackCmd(prev)
		},
	})
	r.register(Command{
		ID:          "playback.shuffle",
		Name:        "Toggle Shuffle",
		Description: "Turn shuffle on or off",
		Category:    "Playback",
		Keybinding:  m.cfg.Keybindings.Shuffle,
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.queue.ToggleShuffle()
			return *m, nil
		},
	})
	r.register(Command{
		ID:          "playback.repeat",
		Name:        "Cycle Repeat",
		Description: "Cycle through repeat modes (off/all/one)",
		Category:    "Playback",
		Keybinding:  m.cfg.Keybindings.Repeat,
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.queue.CycleRepeat()
			return *m, nil
		},
	})
	r.register(Command{
		ID:          "playback.mute",
		Name:        "Toggle Mute",
		Description: "Mute or unmute audio",
		Category:    "Playback",
		Keybinding:  m.cfg.Keybindings.Mute,
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.muted = !m.muted
			return *m, func() tea.Msg {
				if err := m.player.SetMute(m.muted); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		},
	})
	r.register(Command{
		ID:          "playback.volume_up",
		Name:        "Volume Up",
		Description: "Increase volume",
		Category:    "Playback",
		Keybinding:  m.cfg.Keybindings.VolumeUp,
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.volume += float64(m.cfg.Player.VolumeStep)
			if m.volume > 100 {
				m.volume = 100
			}
			return *m, func() tea.Msg {
				if err := m.player.SetVolume(m.volume); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		},
	})
	r.register(Command{
		ID:          "playback.volume_down",
		Name:        "Volume Down",
		Description: "Decrease volume",
		Category:    "Playback",
		Keybinding:  m.cfg.Keybindings.VolumeDown,
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.volume -= float64(m.cfg.Player.VolumeStep)
			if m.volume < 0 {
				m.volume = 0
			}
			return *m, func() tea.Msg {
				if err := m.player.SetVolume(m.volume); err != nil {
					return playerMsg{Err: err}
				}
				return nil
			}
		},
	})

	// Queue commands
	r.register(Command{
		ID:          "queue.clear",
		Name:        "Clear Queue",
		Description: "Remove all tracks from the queue",
		Category:    "Queue",
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.queue.Clear()
			return *m, nil
		},
	})

	// UI commands
	r.register(Command{
		ID:          "ui.help",
		Name:        "Show Help",
		Description: "Display keybindings help",
		Category:    "UI",
		Keybinding:  m.cfg.Keybindings.Help,
		Handler: func(m *Model) (Model, tea.Cmd) {
			m.showHelp = !m.showHelp
			return *m, nil
		},
	})
	r.register(Command{
		ID:          "ui.quit",
		Name:        "Quit",
		Description: "Exit Tunez",
		Category:    "UI",
		Keybinding:  m.cfg.Keybindings.Quit,
		Handler: func(m *Model) (Model, tea.Cmd) {
			return *m, tea.Quit
		},
	})

	return r
}

func (r *CommandRegistry) register(cmd Command) {
	r.commands = append(r.commands, cmd)
}

// Commands returns all registered commands.
func (r *CommandRegistry) Commands() []Command {
	return r.commands
}

// SearchableNames returns command names for fuzzy matching.
func (r *CommandRegistry) SearchableNames() []string {
	names := make([]string, len(r.commands))
	for i, cmd := range r.commands {
		names[i] = cmd.Name
	}
	return names
}
