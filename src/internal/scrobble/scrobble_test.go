package scrobble_test

import (
	"testing"
	"time"

	"github.com/tunez/tunez/internal/scrobble"
	"github.com/tunez/tunez/internal/scrobble/lastfm"
)

func TestManager(t *testing.T) {
	mgr := scrobble.NewManager()

	s1 := lastfm.New("lastfm1", lastfm.Config{APIKey: "key", APISecret: "secret", SessionKey: "session"})
	s2 := lastfm.New("lastfm2", lastfm.Config{}) // Disabled

	mgr.Register(s1)
	mgr.Register(s2)

	if len(mgr.Scrobblers()) != 2 {
		t.Errorf("expected 2 scrobblers, got %d", len(mgr.Scrobblers()))
	}

	if mgr.EnabledCount() != 1 {
		t.Errorf("expected 1 enabled, got %d", mgr.EnabledCount())
	}
}

func TestManagerShouldScrobble(t *testing.T) {
	mgr := scrobble.NewManager()

	s1 := lastfm.New("lastfm1", lastfm.Config{APIKey: "key", APISecret: "secret", SessionKey: "session"})
	mgr.Register(s1)

	track := scrobble.Track{
		Title:      "Test",
		DurationMs: 60000,
		StartedAt:  time.Now(),
	}

	// Fan out NowPlaying (async, won't fail without network)
	mgr.NowPlaying(nil, track)

	// Wait briefly for async call
	time.Sleep(10 * time.Millisecond)

	// Manager delegates to registered scrobblers
	mgr.UpdatePosition(35*time.Second, false) // 58%

	if !mgr.ShouldScrobble() {
		t.Error("expected manager ShouldScrobble true at 58%")
	}
}

func TestManagerTotalPending(t *testing.T) {
	mgr := scrobble.NewManager()

	s1 := lastfm.New("lastfm1", lastfm.Config{}) // Disabled - will queue
	s2 := lastfm.New("lastfm2", lastfm.Config{}) // Disabled - will queue

	mgr.Register(s1)
	mgr.Register(s2)

	track := scrobble.Track{
		Title:     "Test",
		StartedAt: time.Now(),
	}

	// Scrobble fans out to both (they'll queue since disabled)
	mgr.Scrobble(nil, track)

	// Wait for async
	time.Sleep(20 * time.Millisecond)

	if mgr.TotalPendingCount() != 2 {
		t.Errorf("expected 2 total pending, got %d", mgr.TotalPendingCount())
	}
}
