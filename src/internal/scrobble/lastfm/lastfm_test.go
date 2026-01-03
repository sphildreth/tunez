package lastfm

import (
	"testing"
	"time"

	"github.com/tunez/tunez/internal/scrobble"
)

func TestNew(t *testing.T) {
	s := New("test", Config{})
	if s.IsEnabled() {
		t.Error("expected disabled without api key")
	}

	s = New("test", Config{APIKey: "key", APISecret: "secret"})
	if s.IsEnabled() {
		t.Error("expected disabled without session key")
	}

	s = New("test", Config{APIKey: "key", APISecret: "secret", SessionKey: "session"})
	if !s.IsEnabled() {
		t.Error("expected enabled with all keys")
	}
}

func TestShouldScrobble(t *testing.T) {
	s := New("test", Config{APIKey: "key", APISecret: "secret", SessionKey: "session"})

	// No track playing
	if s.ShouldScrobble() {
		t.Error("expected false with no track")
	}

	// Start playing a short track
	track := scrobble.Track{
		Title:      "Short Song",
		DurationMs: 60000, // 1 minute
		StartedAt:  time.Now(),
	}
	_ = s.NowPlaying(nil, track)

	// Played less than 50%
	s.UpdatePosition(20*time.Second, false) // 33%
	if s.ShouldScrobble() {
		t.Error("expected false at 33%")
	}

	// Played more than 50%
	s.UpdatePosition(35*time.Second, false) // 58%
	if !s.ShouldScrobble() {
		t.Error("expected true at 58%")
	}

	// Long track, 4+ minutes played
	longTrack := scrobble.Track{
		Title:      "Long Song",
		DurationMs: 600000, // 10 minutes
		StartedAt:  time.Now(),
	}
	_ = s.NowPlaying(nil, longTrack)
	s.UpdatePosition(4*time.Minute, false)
	if !s.ShouldScrobble() {
		t.Error("expected true at 4 minutes")
	}

	// Long track, less than 4 minutes and less than 50%
	_ = s.NowPlaying(nil, longTrack) // Reset
	s.UpdatePosition(3*time.Minute, false) // 30%
	if s.ShouldScrobble() {
		t.Error("expected false at 3 minutes / 30%")
	}
}

func TestPendingQueue(t *testing.T) {
	s := New("test", Config{}) // Disabled scrobbler

	track := scrobble.Track{
		Title:     "Test Track",
		Artist:    "Test Artist",
		Album:     "Test Album",
		StartedAt: time.Now(),
	}

	// Scrobble when disabled queues the track
	_ = s.Scrobble(nil, track)
	if s.PendingCount() != 1 {
		t.Errorf("expected 1 pending, got %d", s.PendingCount())
	}

	// Queue more
	for i := 0; i < 60; i++ {
		_ = s.Scrobble(nil, track)
	}

	// Should be capped at 50
	if s.PendingCount() != 50 {
		t.Errorf("expected 50 pending (capped), got %d", s.PendingCount())
	}
}

func TestUpdatePosition(t *testing.T) {
	s := New("test", Config{APIKey: "key", APISecret: "secret", SessionKey: "session"})

	// No track - should not panic
	s.UpdatePosition(30*time.Second, false)

	// With track
	track := scrobble.Track{
		Title:      "Test",
		DurationMs: 180000,
		StartedAt:  time.Now(),
	}
	_ = s.NowPlaying(nil, track)

	s.UpdatePosition(60*time.Second, false)

	// Check via ShouldScrobble behavior
	// At 60s of a 180s track = 33%, should not scrobble
	if s.ShouldScrobble() {
		t.Error("expected false at 33%")
	}

	s.UpdatePosition(100*time.Second, false) // 55%
	if !s.ShouldScrobble() {
		t.Error("expected true at 55%")
	}

	// Reset and test paused behavior
	_ = s.NowPlaying(nil, track)
	s.UpdatePosition(60*time.Second, false)
	s.UpdatePosition(90*time.Second, true) // Paused - should not update

	// Should still be at ~60s equivalent behavior
	if s.ShouldScrobble() {
		t.Error("expected false when paused didn't advance position")
	}
}

func TestSign(t *testing.T) {
	s := New("test", Config{APIKey: "key", APISecret: "secret", SessionKey: "session"})

	params := map[string]string{
		"method":  "track.scrobble",
		"track":   "Test",
		"artist":  "Artist",
		"api_key": "key",
	}

	sig := s.sign(params)
	if sig == "" {
		t.Error("expected non-empty signature")
	}
	if len(sig) != 32 {
		t.Errorf("expected 32 char MD5 hex, got %d", len(sig))
	}

	// Same params should produce same signature
	sig2 := s.sign(params)
	if sig != sig2 {
		t.Error("expected deterministic signature")
	}
}
