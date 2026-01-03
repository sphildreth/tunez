package app

import (
	"testing"

	"github.com/tunez/tunez/internal/config"
	"github.com/tunez/tunez/internal/queue"
)

func TestCommandRegistry(t *testing.T) {
	// Create a minimal model for testing
	m := &Model{}
	m.cfg = &config.Config{}
	m.queue = queue.New()

	registry := NewCommandRegistry(m)

	t.Run("has commands", func(t *testing.T) {
		cmds := registry.Commands()
		if len(cmds) == 0 {
			t.Error("expected commands to be registered")
		}
	})

	t.Run("has navigation commands", func(t *testing.T) {
		cmds := registry.Commands()
		hasNav := false
		for _, cmd := range cmds {
			if cmd.Category == "Navigation" {
				hasNav = true
				break
			}
		}
		if !hasNav {
			t.Error("expected Navigation category")
		}
	})

	t.Run("has playback commands", func(t *testing.T) {
		cmds := registry.Commands()
		hasPlayback := false
		for _, cmd := range cmds {
			if cmd.Category == "Playback" {
				hasPlayback = true
				break
			}
		}
		if !hasPlayback {
			t.Error("expected Playback category")
		}
	})

	t.Run("searchable names match commands", func(t *testing.T) {
		names := registry.SearchableNames()
		cmds := registry.Commands()
		if len(names) != len(cmds) {
			t.Errorf("expected %d names, got %d", len(cmds), len(names))
		}
	})
}

func TestPaletteState(t *testing.T) {
	// Create a minimal model for testing
	m := &Model{}
	m.cfg = &config.Config{}
	m.queue = queue.New()

	registry := NewCommandRegistry(m)
	palette := NewPaletteState(registry)

	t.Run("initial state", func(t *testing.T) {
		if palette.Input() != "" {
			t.Error("expected empty input")
		}
		if palette.selected != 0 {
			t.Error("expected selection at 0")
		}
	})

	t.Run("insert char", func(t *testing.T) {
		palette.Reset()
		palette.InsertChar('g')
		palette.InsertChar('o')
		if palette.Input() != "go" {
			t.Errorf("expected 'go', got '%s'", palette.Input())
		}
	})

	t.Run("backspace", func(t *testing.T) {
		palette.Reset()
		palette.SetInput("test")
		palette.Backspace()
		if palette.Input() != "tes" {
			t.Errorf("expected 'tes', got '%s'", palette.Input())
		}
	})

	t.Run("cursor navigation", func(t *testing.T) {
		palette.Reset()
		palette.SetInput("test")
		if palette.cursor != 4 {
			t.Errorf("expected cursor at 4, got %d", palette.cursor)
		}
		palette.CursorLeft()
		if palette.cursor != 3 {
			t.Errorf("expected cursor at 3, got %d", palette.cursor)
		}
		palette.CursorRight()
		if palette.cursor != 4 {
			t.Errorf("expected cursor at 4, got %d", palette.cursor)
		}
	})

	t.Run("selection navigation", func(t *testing.T) {
		palette.Reset()
		palette.SelectDown()
		if palette.selected != 1 {
			t.Errorf("expected selection at 1, got %d", palette.selected)
		}
		palette.SelectUp()
		if palette.selected != 0 {
			t.Errorf("expected selection at 0, got %d", palette.selected)
		}
	})

	t.Run("fuzzy search filters commands", func(t *testing.T) {
		palette.Reset()
		palette.SetInput("play")
		if len(palette.matches) == 0 {
			t.Error("expected fuzzy matches for 'play'")
		}
	})

	t.Run("selected command returns valid command", func(t *testing.T) {
		palette.Reset()
		cmd := palette.SelectedCommand()
		if cmd == nil {
			t.Error("expected a command to be selected")
		}
	})
}
