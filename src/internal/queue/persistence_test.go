package queue

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/tunez/tunez/internal/provider"
)

func TestPersistenceSaveLoad(t *testing.T) {
	// Create temp db
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "queue.db")

	store, err := NewPersistenceStore(dbPath)
	if err != nil {
		t.Fatalf("NewPersistenceStore: %v", err)
	}
	defer store.Close()

	// Create a queue with some tracks
	q := New()
	tracks := []provider.Track{
		{ID: "t1", Title: "Track 1", ArtistName: "Artist 1"},
		{ID: "t2", Title: "Track 2", ArtistName: "Artist 2"},
		{ID: "t3", Title: "Track 3", ArtistName: "Artist 3"},
	}
	q.Add(tracks...)
	_ = q.SetCurrent(1)
	q.CycleRepeat() // RepeatAll

	// Save
	ctx := context.Background()
	if err := store.Save(ctx, q, "filesystem", "home-files"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Load
	result, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(result.Tracks) != 3 {
		t.Errorf("expected 3 tracks, got %d", len(result.Tracks))
	}
	if result.CurrentIndex != 1 {
		t.Errorf("expected current index 1, got %d", result.CurrentIndex)
	}
	if result.Repeat != RepeatAll {
		t.Errorf("expected RepeatAll, got %d", result.Repeat)
	}
	if result.ProfileID != "home-files" {
		t.Errorf("expected profile 'home-files', got %q", result.ProfileID)
	}

	// Verify track data
	if result.Tracks[0].ID != "t1" || result.Tracks[0].Title != "Track 1" {
		t.Errorf("track 0 mismatch: %+v", result.Tracks[0])
	}
	if result.Tracks[1].ID != "t2" || result.Tracks[1].Title != "Track 2" {
		t.Errorf("track 1 mismatch: %+v", result.Tracks[1])
	}
}

func TestPersistenceEmptyQueue(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "queue.db")

	store, err := NewPersistenceStore(dbPath)
	if err != nil {
		t.Fatalf("NewPersistenceStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Load from empty db should work
	result, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(result.Tracks) != 0 {
		t.Errorf("expected 0 tracks, got %d", len(result.Tracks))
	}
	if result.CurrentIndex != -1 {
		t.Errorf("expected current index -1, got %d", result.CurrentIndex)
	}
}

func TestPersistenceClear(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "queue.db")

	store, err := NewPersistenceStore(dbPath)
	if err != nil {
		t.Fatalf("NewPersistenceStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Add and save some tracks
	q := New()
	q.Add(provider.Track{ID: "t1", Title: "Track 1"})
	if err := store.Save(ctx, q, "filesystem", "test"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Clear
	if err := store.Clear(ctx); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	// Load should return empty
	result, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(result.Tracks) != 0 {
		t.Errorf("expected 0 tracks after clear, got %d", len(result.Tracks))
	}
}

func TestPersistenceInvalidIndex(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "queue.db")

	store, err := NewPersistenceStore(dbPath)
	if err != nil {
		t.Fatalf("NewPersistenceStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create queue with 2 tracks, current at 5 (invalid)
	q := New()
	q.Add(provider.Track{ID: "t1"}, provider.Track{ID: "t2"})
	q.current = 5 // Force invalid index

	if err := store.Save(ctx, q, "test", "test"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	result, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Should clamp to valid range
	if result.CurrentIndex != 1 {
		t.Errorf("expected clamped index 1, got %d", result.CurrentIndex)
	}
}

func TestPersistenceShuffleState(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "queue.db")

	store, err := NewPersistenceStore(dbPath)
	if err != nil {
		t.Fatalf("NewPersistenceStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	q := New()
	q.Add(provider.Track{ID: "t1"}, provider.Track{ID: "t2"}, provider.Track{ID: "t3"})
	q.ToggleShuffle() // Enable shuffle

	if err := store.Save(ctx, q, "test", "test"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	result, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !result.Shuffled {
		t.Error("expected shuffled=true")
	}
}

func TestPersistenceDefaultPath(t *testing.T) {
	// Skip if running in CI without home dir
	if os.Getenv("HOME") == "" && os.Getenv("USERPROFILE") == "" {
		t.Skip("no home directory")
	}

	path, err := defaultQueueDBPath()
	if err != nil {
		t.Fatalf("defaultQueueDBPath: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
	if filepath.Base(path) != "queue.db" {
		t.Errorf("expected queue.db, got %s", filepath.Base(path))
	}
}
