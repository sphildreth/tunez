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
	"github.com/tunez/tunez/internal/ui"
)

var version = "0.1.0"

func main() {
	cfgPath := flag.String("config", "", "config file path")
	doctor := flag.Bool("doctor", false, "run diagnostics")
	showVersion := flag.Bool("version", false, "print version")
	searchArtist := flag.String("artist", "", "search for artist")
	searchAlbum := flag.String("album", "", "search for album")
	autoPlay := flag.Bool("play", false, "auto-play first search result")
	randomPlay := flag.Bool("random", false, "play random songs")
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

	// NO_COLOR env var support per accessibility spec
	noColor := os.Getenv("NO_COLOR") != "" || cfg.UI.NoEmoji
	theme := ui.Rainbow(noColor)

	// Build startup options from CLI flags
	startupOpts := app.StartupOptions{
		SearchArtist: *searchArtist,
		SearchAlbum:  *searchAlbum,
		AutoPlay:     *autoPlay,
		RandomPlay:   *randomPlay,
	}

	model := app.New(cfg, prov, func(p config.Profile) (provider.Provider, error) {
		return buildProvider(p)
	}, ctrl, profile.Settings, theme, startupOpts)
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
	fmt.Println("Config file OK")
	mpvPath, err := exec.LookPath(cfg.Player.MPVPath)
	if err != nil {
		fmt.Printf("mpv path (%s): %v\n", cfg.Player.MPVPath, err)
	} else {
		fmt.Printf("mpv path (%s): OK (resolved: %s)\n", cfg.Player.MPVPath, mpvPath)
	}
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		fmt.Printf("ffprobe: NOT FOUND (duration detection disabled)\n")
	} else {
		fmt.Printf("ffprobe: OK (resolved: %s)\n", ffprobePath)
	}
	profile, _ := cfg.ProfileByID(cfg.ActiveProfile)
	prov, err := buildProvider(profile)
	if err != nil {
		fmt.Println("provider:", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := prov.Initialize(ctx, profile.Settings); err != nil {
		fmt.Println("provider init:", err)
		return
	}
	ok, details := prov.Health(ctx)
	fmt.Printf("provider health: %v (%s)\n", ok, details)
	logger.Info("doctor complete")
}
