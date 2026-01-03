package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config holds Tunez runtime configuration loaded from TOML.
type Config struct {
	ConfigVersion int              `toml:"config_version"`
	ActiveProfile string           `toml:"active_profile"`
	UI            UIConfig         `toml:"ui"`
	Player        PlayerConfig     `toml:"player"`
	Queue         QueueConfig      `toml:"queue"`
	Artwork       ArtworkConfig    `toml:"artwork"`
	Scrobble      ScrobbleConfig   `toml:"scrobble"`
	Keybindings   KeybindConfig    `toml:"keybindings"`
	Profiles      []Profile        `toml:"profiles"`
	Scrobblers    []ScrobblerEntry `toml:"scrobblers"`
}

// QueueConfig holds queue persistence settings.
type QueueConfig struct {
	Persist bool `toml:"persist"`
}

// ArtworkConfig holds artwork display settings.
type ArtworkConfig struct {
	Enabled   bool   `toml:"enabled"`
	Width     int    `toml:"width"`
	Height    int    `toml:"height"`
	Quality   string `toml:"quality"`    // low, medium, high
	ScaleMode string `toml:"scale_mode"` // fit, fill, stretch
	CacheDays int    `toml:"cache_days"`
}

// ScrobbleConfig holds global scrobbling settings.
type ScrobbleConfig struct {
	Enabled bool `toml:"enabled"` // Master switch for all scrobblers
}

// ScrobblerEntry defines a scrobbler configuration.
type ScrobblerEntry struct {
	ID       string         `toml:"id"`
	Type     string         `toml:"type"` // "lastfm", "melodee"
	Enabled  bool           `toml:"enabled"`
	Settings map[string]any `toml:"settings"`
}

type UIConfig struct {
	PageSize int    `toml:"page_size"`
	NoEmoji  bool   `toml:"no_emoji"`
	Theme    string `toml:"theme"`
}

type PlayerConfig struct {
	MPVPath         string `toml:"mpv_path"`
	IPC             string `toml:"ipc"`
	InitialVolume   int    `toml:"initial_volume"`
	CacheSeconds    int    `toml:"cache_secs"`
	NetworkTimeout  int    `toml:"network_timeout_ms"`
	SeekSmall       int    `toml:"seek_small_seconds"`
	SeekLarge       int    `toml:"seek_large_seconds"`
	VolumeStep      int    `toml:"volume_step"`
	EnableAutostart bool   `toml:"autostart"`
}

// KeybindConfig allows customizing keybindings.
type KeybindConfig struct {
	PlayPause    string `toml:"play_pause"`
	NextTrack    string `toml:"next_track"`
	PrevTrack    string `toml:"prev_track"`
	SeekForward  string `toml:"seek_forward"`
	SeekBackward string `toml:"seek_backward"`
	VolumeUp     string `toml:"volume_up"`
	VolumeDown   string `toml:"volume_down"`
	Mute         string `toml:"mute"`
	Shuffle      string `toml:"shuffle"`
	Repeat       string `toml:"repeat"`
	Search       string `toml:"search"`
	Help         string `toml:"help"`
	Quit         string `toml:"quit"`
}

type Profile struct {
	ID       string         `toml:"id"`
	Name     string         `toml:"name"`
	Provider string         `toml:"provider"`
	Enabled  bool           `toml:"enabled"`
	Settings map[string]any `toml:"settings"`
}

// Load reads configuration from disk. If path is empty, a default OS-specific
// location is used.
func Load(path string) (*Config, string, error) {
	cfgPath := path
	if cfgPath == "" {
		var err error
		cfgPath, err = defaultPath()
		if err != nil {
			return nil, "", fmt.Errorf("resolve config path: %w", err)
		}
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, cfgPath, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, cfgPath, fmt.Errorf("parse config: %w", err)
	}

	applyDefaults(&cfg)

	if err := Validate(cfg); err != nil {
		return nil, cfgPath, err
	}

	return &cfg, cfgPath, nil
}

func defaultPath() (string, error) {
	var base string
	switch runtime.GOOS {
	case "darwin":
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "tunez")
	case "windows":
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "Tunez")
	default:
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "tunez")
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(base, "config.toml"), nil
}

func applyDefaults(cfg *Config) {
	if cfg.UI.PageSize == 0 {
		cfg.UI.PageSize = 100
	}
	if cfg.UI.Theme == "" {
		cfg.UI.Theme = "rainbow"
	}
	if cfg.Player.MPVPath == "" {
		cfg.Player.MPVPath = "mpv"
	}
	if cfg.Player.InitialVolume == 0 {
		cfg.Player.InitialVolume = 70
	}
	if cfg.Player.SeekSmall == 0 {
		cfg.Player.SeekSmall = 5
	}
	if cfg.Player.SeekLarge == 0 {
		cfg.Player.SeekLarge = 30
	}
	if cfg.Player.VolumeStep == 0 {
		cfg.Player.VolumeStep = 5
	}
	if cfg.Player.NetworkTimeout == 0 {
		cfg.Player.NetworkTimeout = 8000
	}
	// Keybinding defaults
	if cfg.Keybindings.PlayPause == "" {
		cfg.Keybindings.PlayPause = "space"
	}
	if cfg.Keybindings.NextTrack == "" {
		cfg.Keybindings.NextTrack = "n"
	}
	if cfg.Keybindings.PrevTrack == "" {
		cfg.Keybindings.PrevTrack = "p"
	}
	if cfg.Keybindings.SeekForward == "" {
		cfg.Keybindings.SeekForward = "l"
	}
	if cfg.Keybindings.SeekBackward == "" {
		cfg.Keybindings.SeekBackward = "h"
	}
	if cfg.Keybindings.VolumeUp == "" {
		cfg.Keybindings.VolumeUp = "+"
	}
	if cfg.Keybindings.VolumeDown == "" {
		cfg.Keybindings.VolumeDown = "-"
	}
	if cfg.Keybindings.Mute == "" {
		cfg.Keybindings.Mute = "m"
	}
	if cfg.Keybindings.Shuffle == "" {
		cfg.Keybindings.Shuffle = "s"
	}
	if cfg.Keybindings.Repeat == "" {
		cfg.Keybindings.Repeat = "r"
	}
	if cfg.Keybindings.Search == "" {
		cfg.Keybindings.Search = "/"
	}
	if cfg.Keybindings.Help == "" {
		cfg.Keybindings.Help = "?"
	}
	if cfg.Keybindings.Quit == "" {
		cfg.Keybindings.Quit = "q,ctrl+c"
	}
	// Queue defaults - persist enabled by default
	if !cfg.Queue.Persist {
		// Default to true unless explicitly set to false in config
		// Note: TOML will parse missing as false, so we treat missing as "use default"
		cfg.Queue.Persist = true
	}
	// Artwork defaults - enabled by default
	if !cfg.Artwork.Enabled {
		cfg.Artwork.Enabled = true
	}
	if cfg.Artwork.Width == 0 {
		cfg.Artwork.Width = 20 // Reverted to smaller default
	}
	if cfg.Artwork.Height == 0 {
		cfg.Artwork.Height = 10 // Reverted to smaller default
	}
	if cfg.Artwork.Quality == "" {
		cfg.Artwork.Quality = "medium"
	}
	if cfg.Artwork.ScaleMode == "" {
		cfg.Artwork.ScaleMode = "fit"
	}
	if cfg.Artwork.CacheDays == 0 {
		cfg.Artwork.CacheDays = 30
	}
}

// Validate performs semantic validation of config according to docs/CONFIG.md.
func Validate(cfg Config) error {
	if cfg.ActiveProfile == "" {
		return errors.New("active_profile is required")
	}
	profile, ok := cfg.ProfileByID(cfg.ActiveProfile)
	if !ok {
		return fmt.Errorf("active_profile %q not found", cfg.ActiveProfile)
	}
	if !profile.Enabled {
		return fmt.Errorf("active_profile %q is disabled", cfg.ActiveProfile)
	}
	if cfg.Player.InitialVolume < 0 || cfg.Player.InitialVolume > 100 {
		return fmt.Errorf("player.initial_volume must be 0-100")
	}
	if _, err := os.Stat(cfg.Player.MPVPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if _, lookErr := execLookPath(cfg.Player.MPVPath); lookErr != nil {
				return fmt.Errorf("mpv not found (%s): %w", cfg.Player.MPVPath, lookErr)
			}
		}
	}

	switch profile.Provider {
	case "filesystem":
		if err := validateFilesystem(profile.Settings); err != nil {
			return err
		}
	case "melodee":
		if err := validateMelodee(profile.Settings); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown provider: %s", profile.Provider)
	}
	return nil
}

func validateFilesystem(settings map[string]any) error {
	roots, ok := settings["roots"].([]any)
	if !ok || len(roots) == 0 {
		return errors.New("filesystem.roots is required")
	}
	for _, r := range roots {
		s, _ := r.(string)
		if s == "" {
			return errors.New("filesystem.roots contains empty path")
		}
		if _, err := os.Stat(s); err != nil {
			return fmt.Errorf("filesystem root %s: %w", s, err)
		}
	}
	return nil
}

func validateMelodee(settings map[string]any) error {
	baseURL, _ := settings["base_url"].(string)
	if baseURL == "" {
		return errors.New("melodee.base_url is required")
	}
	return nil
}

// ProfileByID returns profile and true when found.
func (c Config) ProfileByID(id string) (Profile, bool) {
	for _, p := range c.Profiles {
		if p.ID == id {
			return p, true
		}
	}
	return Profile{}, false
}

// DeadlineContext returns a context with default timeout based on player network timeout.
func (c Config) DeadlineContext() (context.Context, context.CancelFunc) {
	d := time.Duration(c.Player.NetworkTimeout) * time.Millisecond
	if d == 0 {
		d = 8 * time.Second
	}
	return context.WithTimeout(context.Background(), d)
}

// execLookPath is a test seam.
var execLookPath = func(file string) (string, error) {
	return exec.LookPath(file)
}
