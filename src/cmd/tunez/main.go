package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunez/tunez/internal/app"
	"github.com/tunez/tunez/internal/artwork"
	"github.com/tunez/tunez/internal/config"
	"github.com/tunez/tunez/internal/logging"
	"github.com/tunez/tunez/internal/player"
	"github.com/tunez/tunez/internal/provider"
	"github.com/tunez/tunez/internal/providers/filesystem"
	"github.com/tunez/tunez/internal/providers/melodee"
	"github.com/tunez/tunez/internal/queue"
	"github.com/tunez/tunez/internal/scrobble"
	"github.com/tunez/tunez/internal/scrobble/lastfm"
	scrobblemelodee "github.com/tunez/tunez/internal/scrobble/melodee"
	"github.com/tunez/tunez/internal/ui"
)

var version = "0.1.0"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Tunez - A terminal music player

Usage: tunez [options]

Options:
  -config string
        Path to config file (default: ~/.config/tunez/config.toml)
  -version
        Print version and exit
  -config-init
        Create example config file

Diagnostics:
  -doctor
        Check configuration and dependencies (fast, no library scan)
  -scan
        Scan/rescan music library

Playback:
  -artist string
        Search for artist and add matching tracks to queue
  -album string
        Search for album and add matching tracks to queue
  -random
        Add random tracks to queue (uses ui.page_size from config)
  -play
        Auto-play first track in queue (use with -artist, -album, or -random)

Examples:
  tunez                                    # Start interactive TUI
  tunez --config-init                      # Create example config
  tunez --doctor                           # Check setup
  tunez --scan                             # Rescan music library
  tunez --random --play                    # Play random tracks
  tunez --artist "Pink Floyd" --play       # Play artist
  tunez --artist "Queen" --album "News"    # Queue matching album

`)
	}

	cfgPath := flag.String("config", "", "")
	doctor := flag.Bool("doctor", false, "")
	scan := flag.Bool("scan", false, "")
	showVersion := flag.Bool("version", false, "")
	configInit := flag.Bool("config-init", false, "")
	searchArtist := flag.String("artist", "", "")
	searchAlbum := flag.String("album", "", "")
	autoPlay := flag.Bool("play", false, "")
	randomPlay := flag.Bool("random", false, "")
	flag.Parse()

	if *showVersion {
		fmt.Println("tunez", version)
		return
	}

	if *configInit {
		runConfigInit()
		return
	}

	cfg, resolvedPath, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	logger, logFile, err := logging.Setup()
	if err != nil {
		log.Fatalf("setup logging: %v", err)
	}
	defer logFile.Close()
	logger.Info("starting tunez", slog.String("config", resolvedPath))

	if *doctor {
		runDoctor(cfg, logger)
		return
	}

	if *scan {
		runScan(cfg, logger)
		return
	}

	profile, _ := cfg.ProfileByID(cfg.ActiveProfile)
	prov, err := buildProvider(profile)
	if err != nil {
		logger.Error("provider init", slog.Any("err", err))
		log.Fatalf("init provider: %v", err)
	}

	ctrl := player.New(player.Options{
		MPVPath: cfg.Player.MPVPath,
		Logger:  logger,
	})
	if err := ctrl.Start(context.Background()); err != nil {
		logger.Error("start player", slog.Any("err", err))
		log.Fatalf("start player: %v", err)
	}
	defer ctrl.Stop()

	// Initialize queue persistence store if enabled
	var queueStore *queue.PersistenceStore
	if cfg.Queue.Persist {
		queueStore, err = queue.NewPersistenceStore("")
		if err != nil {
			logger.Warn("queue persistence unavailable", slog.Any("err", err))
		} else {
			defer queueStore.Close()
		}
	}

	// Initialize scrobble manager if enabled
	var scrobbleMgr *scrobble.Manager
	if cfg.Scrobble.Enabled {
		scrobbleMgr = buildScrobbleManager(cfg, prov, logger)
		if scrobbleMgr != nil {
			// Load pending scrobbles from disk
			if err := scrobbleMgr.LoadPending(); err != nil {
				logger.Warn("failed to load pending scrobbles", slog.Any("err", err))
			}
			// Save pending scrobbles on shutdown
			defer func() {
				waitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := scrobbleMgr.Wait(waitCtx); err != nil {
					logger.Warn("pending scrobbles not flushed", slog.Any("err", err))
				}
				if err := scrobbleMgr.SavePending(); err != nil {
					logger.Warn("failed to save pending scrobbles", slog.Any("err", err))
				}
			}()
		}
	}

	// NO_COLOR env var support per accessibility spec
	noColor := os.Getenv("NO_COLOR") != "" || cfg.UI.NoEmoji
	theme := ui.GetTheme(cfg.UI.Theme, noColor)

	// Initialize artwork cache if enabled
	var artCache *artwork.Cache
	if cfg.Artwork.Enabled {
		artCache, err = artwork.NewCache("", cfg.Artwork.CacheDays)
		if err != nil {
			logger.Warn("artwork cache unavailable", slog.Any("err", err))
		}
	}

	// Build startup options from CLI flags
	startupOpts := app.StartupOptions{
		SearchArtist: *searchArtist,
		SearchAlbum:  *searchAlbum,
		AutoPlay:     *autoPlay,
		RandomPlay:   *randomPlay,
	}

	model := app.New(cfg, prov, func(p config.Profile) (provider.Provider, error) {
		return buildProvider(p)
	}, ctrl, profile.Settings, theme, startupOpts, queueStore, scrobbleMgr, artCache, logger)
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		logger.Error("run tui", slog.Any("err", err))
		log.Fatalf("tui: %v", err)
	}
}

func buildProvider(p config.Profile) (provider.Provider, error) {
	switch p.Provider {
	case "filesystem":
		return filesystem.New(), nil
	case "melodee":
		return melodee.New(), nil
	default:
		return nil, fmt.Errorf("unknown provider %s", p.Provider)
	}
}

// buildScrobbleManager creates and configures the scrobble manager based on config.
func buildScrobbleManager(cfg *config.Config, prov provider.Provider, logger *slog.Logger) *scrobble.Manager {
	if len(cfg.Scrobblers) == 0 {
		return nil
	}

	mgr := scrobble.NewManager()

	for _, entry := range cfg.Scrobblers {
		if !entry.Enabled {
			continue
		}

		var s scrobble.Scrobbler
		switch entry.Type {
		case "lastfm":
			lfmCfg := lastfm.Config{}
			if v, ok := entry.Settings["api_key"].(string); ok {
				lfmCfg.APIKey = v
			}
			if v, ok := entry.Settings["api_secret"].(string); ok {
				lfmCfg.APISecret = v
			}
			if v, ok := entry.Settings["session_key"].(string); ok {
				lfmCfg.SessionKey = v
			}
			s = lastfm.New(entry.ID, lfmCfg)
			logger.Info("registered scrobbler", slog.String("id", entry.ID), slog.String("type", "lastfm"))

		case "melodee":
			melCfg := scrobblemelodee.Config{}
			// Check if we should reuse auth from a melodee provider
			if provID, ok := entry.Settings["provider"].(string); ok && provID != "" {
				// Try to get token from current provider if it's melodee
				if mp, ok := prov.(*melodee.Provider); ok && prov.ID() == provID {
					melCfg.TokenProvider = mp
					melCfg.BaseURL = mp.BaseURL()
				}
			}
			// Fallback to explicit settings
			if melCfg.BaseURL == "" {
				if v, ok := entry.Settings["base_url"].(string); ok {
					melCfg.BaseURL = v
				}
			}
			if melCfg.TokenProvider == nil {
				if v, ok := entry.Settings["token"].(string); ok {
					melCfg.Token = v
				}
			}
			s = scrobblemelodee.New(entry.ID, melCfg)
			logger.Info("registered scrobbler", slog.String("id", entry.ID), slog.String("type", "melodee"))

		default:
			logger.Warn("unknown scrobbler type", slog.String("id", entry.ID), slog.String("type", entry.Type))
			continue
		}

		if s != nil {
			mgr.Register(s)
		}
	}

	if mgr.EnabledCount() == 0 && len(mgr.Scrobblers()) == 0 {
		return nil
	}

	return mgr
}

func runDoctor(cfg *config.Config, logger *slog.Logger) {
	fmt.Println("┌─────────────────────────────────────────┐")
	fmt.Println("│           Tunez Doctor Report           │")
	fmt.Println("└─────────────────────────────────────────┘")
	fmt.Println()

	allOK := true
	warnings := 0

	// Config file
	printCheck("Config file", "OK", true, "")

	// Check mpv (required)
	mpvPath, err := exec.LookPath(cfg.Player.MPVPath)
	if err != nil {
		printCheck("mpv", "NOT FOUND", false, cfg.Player.MPVPath)
		allOK = false
	} else {
		// Get mpv version
		out, _ := exec.Command(mpvPath, "--version").Output()
		version := ""
		if len(out) > 0 {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				version = strings.TrimSpace(lines[0])
			}
		}
		printCheck("mpv", "OK", true, version)
	}

	// Check ffprobe (optional)
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		printCheck("ffprobe", "NOT FOUND", false, "optional - for duration/codec detection")
		warnings++
	} else {
		out, _ := exec.Command(ffprobePath, "-version").Output()
		version := ""
		if len(out) > 0 {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				parts := strings.Fields(lines[0])
				if len(parts) >= 3 {
					version = parts[2]
				}
			}
		}
		printCheck("ffprobe", "OK", true, version)
	}

	// Check cava (optional - for visualizer)
	cavaPath, err := exec.LookPath("cava")
	if err != nil {
		printCheck("cava", "NOT FOUND", false, "optional - for audio visualizer")
		warnings++
	} else {
		out, _ := exec.Command(cavaPath, "-v").CombinedOutput()
		version := strings.TrimSpace(string(out))
		printCheck("cava", "OK", true, version)
	}

	fmt.Println()

	// Check profile
	profile, ok := cfg.ProfileByID(cfg.ActiveProfile)
	if !ok {
		printCheck("Active profile", "NOT FOUND", false, cfg.ActiveProfile)
		allOK = false
	} else {
		printCheck("Active profile", profile.Name, true, profile.Provider+" provider")

		// Check provider can be built
		_, err = buildProvider(profile)
		if err != nil {
			printCheck("Provider", "ERROR", false, err.Error())
			allOK = false
		} else {
			printCheck("Provider", "OK", true, "")
		}
	}

	// Check directories
	fmt.Println()
	stateDir, _ := os.UserConfigDir()
	cacheDir, _ := os.UserCacheDir()
	printCheck("Config dir", "OK", true, filepath.Join(stateDir, "tunez"))
	printCheck("Cache dir", "OK", true, filepath.Join(cacheDir, "tunez"))

	// Summary
	fmt.Println()
	fmt.Println("─────────────────────────────────────────")
	if allOK && warnings == 0 {
		fmt.Println("✓ All checks passed!")
	} else if allOK {
		fmt.Printf("✓ All required checks passed (%d optional warnings)\n", warnings)
	} else {
		fmt.Println("✗ Some checks failed. Please resolve the issues above.")
	}

	logger.Info("doctor complete")
}

func printCheck(name, status string, ok bool, detail string) {
	icon := "✓"
	if !ok {
		icon = "✗"
	}
	if detail != "" {
		fmt.Printf("  %s %-15s %s (%s)\n", icon, name+":", status, detail)
	} else {
		fmt.Printf("  %s %-15s %s\n", icon, name+":", status)
	}
}

func runConfigInit() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Error: cannot determine config directory: %v\n", err)
		os.Exit(1)
	}

	tunezDir := filepath.Join(configDir, "tunez")
	configPath := filepath.Join(tunezDir, "config.toml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file already exists: %s\n", configPath)
		fmt.Println("Remove the existing file first if you want to regenerate it.")
		os.Exit(1)
	}

	// Create directory if needed
	if err := os.MkdirAll(tunezDir, 0755); err != nil {
		fmt.Printf("Error: cannot create config directory: %v\n", err)
		os.Exit(1)
	}

	// Example config content
	exampleConfig := `# Tunez Configuration
# See docs/CONFIG.md for full reference

config_version = 1
active_profile = "local"

[ui]
theme = "rainbow"     # rainbow, mono, green, dracula, nord, synthwave, etc.
page_size = 100
no_emoji = false

[player]
mpv_path = "mpv"
initial_volume = 70
seek_small_seconds = 5
seek_large_seconds = 30
volume_step = 5

[queue]
persist = true        # Remember queue across restarts

[artwork]
enabled = true
width = 40
cache_days = 30

[scrobble]
enabled = false       # Set to true and configure scrobblers below

# Uncomment to enable Last.fm scrobbling:
# [[scrobblers]]
# id = "lastfm"
# type = "lastfm"
# enabled = true
# [scrobblers.settings]
# api_key = "YOUR_API_KEY"
# api_secret = "YOUR_API_SECRET"
# session_key = "YOUR_SESSION_KEY"

[keybindings]
play_pause = "space"
next_track = "n"
prev_track = "p"
seek_forward = "l"
seek_backward = "h"
volume_up = "+"
volume_down = "-"
mute = "m"
shuffle = "s"
repeat = "r"
search = "/"
help = "?"
quit = "q,ctrl+c"

# Local filesystem profile
[[profiles]]
id = "local"
name = "My Music"
provider = "filesystem"
enabled = true

[profiles.settings]
roots = ["/home/` + os.Getenv("USER") + `/Music"]
scan_on_start = false

# Melodee API profile (uncomment to enable)
# [[profiles]]
# id = "melodee"
# name = "Melodee Server"
# provider = "melodee"
# enabled = true
#
# [profiles.settings]
# base_url = "https://music.example.com"
# username = "your-username"
# password_env = "TUNEZ_MELODEE_PASSWORD"
`

	if err := os.WriteFile(configPath, []byte(exampleConfig), 0644); err != nil {
		fmt.Printf("Error: cannot write config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created config file: %s\n", configPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit the config file to set your music library path")
	fmt.Println("  2. Run 'tunez --scan' to index your library")
	fmt.Println("  3. Run 'tunez' to start playing!")
}

func runScan(cfg *config.Config, logger *slog.Logger) {
	profile, ok := cfg.ProfileByID(cfg.ActiveProfile)
	if !ok {
		fmt.Printf("Profile '%s' not found\n", cfg.ActiveProfile)
		return
	}

	prov, err := buildProvider(profile)
	if err != nil {
		fmt.Printf("Provider error: %v\n", err)
		return
	}

	fmt.Printf("Scanning library for profile '%s' (%s)...\n", profile.Name, profile.Provider)

	// Force scan by setting scan_on_init in settings with progress callback
	var settings any = profile.Settings
	if settings == nil {
		settings = map[string]any{}
	}
	if m, ok := settings.(map[string]any); ok {
		m["scan_on_init"] = true
		// Add progress callback for CLI feedback
		m["scan_progress"] = func(count int, path string) {
			// Truncate path for display
			displayPath := path
			if len(displayPath) > 60 {
				displayPath = "..." + displayPath[len(displayPath)-57:]
			}
			fmt.Printf("\r\033[K  Scanned %d tracks: %s", count, displayPath)
		}
		settings = m
	}

	ctx := context.Background() // No timeout for scan
	start := time.Now()
	if err := prov.Initialize(ctx, settings); err != nil {
		fmt.Printf("\nScan error: %v\n", err)
		return
	}

	// Clear progress line and show completion
	fmt.Printf("\r\033[K")

	// Get counts
	healthy, details := prov.Health(ctx)
	if !healthy {
		fmt.Printf("Health check failed: %s\n", details)
		return
	}

	fmt.Printf("Scan complete in %s\n", time.Since(start).Round(time.Millisecond))
	fmt.Printf("  %s\n", details)
	logger.Info("scan complete", slog.Duration("duration", time.Since(start)))
}
