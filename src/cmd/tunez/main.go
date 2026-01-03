package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunez/tunez/internal/app"
	"github.com/tunez/tunez/internal/config"
	"github.com/tunez/tunez/internal/logging"
	"github.com/tunez/tunez/internal/player"
	"github.com/tunez/tunez/internal/provider"
	"github.com/tunez/tunez/internal/providers/filesystem"
	"github.com/tunez/tunez/internal/providers/melodee"
	"github.com/tunez/tunez/internal/queue"
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
	searchArtist := flag.String("artist", "", "")
	searchAlbum := flag.String("album", "", "")
	autoPlay := flag.Bool("play", false, "")
	randomPlay := flag.Bool("random", false, "")
	flag.Parse()

	if *showVersion {
		fmt.Println("tunez", version)
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

	// NO_COLOR env var support per accessibility spec
	noColor := os.Getenv("NO_COLOR") != "" || cfg.UI.NoEmoji
	theme := ui.GetTheme(cfg.UI.Theme, noColor)

	// Build startup options from CLI flags
	startupOpts := app.StartupOptions{
		SearchArtist: *searchArtist,
		SearchAlbum:  *searchAlbum,
		AutoPlay:     *autoPlay,
		RandomPlay:   *randomPlay,
	}

	model := app.New(cfg, prov, func(p config.Profile) (provider.Provider, error) {
		return buildProvider(p)
	}, ctrl, profile.Settings, theme, startupOpts, queueStore, logger)
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

func runDoctor(cfg *config.Config, logger *slog.Logger) {
	fmt.Println("Tunez doctor")
	fmt.Println("Config file: OK")

	// Check mpv
	mpvPath, err := exec.LookPath(cfg.Player.MPVPath)
	if err != nil {
		fmt.Printf("mpv (%s): NOT FOUND\n", cfg.Player.MPVPath)
	} else {
		fmt.Printf("mpv: OK (%s)\n", mpvPath)
	}

	// Check ffprobe
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		fmt.Println("ffprobe: NOT FOUND (optional, for duration detection)")
	} else {
		fmt.Printf("ffprobe: OK (%s)\n", ffprobePath)
	}

	// Check profile
	profile, ok := cfg.ProfileByID(cfg.ActiveProfile)
	if !ok {
		fmt.Printf("Active profile (%s): NOT FOUND\n", cfg.ActiveProfile)
		return
	}
	fmt.Printf("Active profile: %s (%s provider)\n", profile.Name, profile.Provider)

	// Check provider can be built (but don't initialize/scan)
	_, err = buildProvider(profile)
	if err != nil {
		fmt.Printf("Provider: ERROR - %v\n", err)
		return
	}
	fmt.Println("Provider: OK")

	logger.Info("doctor complete")
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
