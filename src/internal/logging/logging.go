package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Setup creates a slog.Logger that writes to a rotating log file in the user
// state directory. The caller is responsible for closing the file.
func Setup() (*slog.Logger, *os.File, error) {
	stateDir, err := StateDir()
	if err != nil {
		return nil, nil, fmt.Errorf("state dir: %w", err)
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create state dir: %w", err)
	}
	path := filepath.Join(stateDir, fmt.Sprintf("tunez-%s.log", time.Now().Format("20060102")))
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}
	handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(handler), f, nil
}

// StateDir returns the path to the tunez state directory (~/.config/tunez/state)
func StateDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tunez", "state"), nil
}
