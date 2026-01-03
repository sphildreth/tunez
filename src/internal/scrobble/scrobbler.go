package scrobble

import (
	"context"
	"sync"
	"time"
)

// Scrobbler is the interface implemented by all scrobbling backends.
type Scrobbler interface {
	// ID returns a unique identifier for this scrobbler instance.
	ID() string
	// Name returns a human-readable name for the scrobbler.
	Name() string
	// IsEnabled returns true if this scrobbler is configured and ready.
	IsEnabled() bool

	// NowPlaying reports the currently playing track.
	NowPlaying(ctx context.Context, track Track) error
	// Scrobble reports a completed track play.
	Scrobble(ctx context.Context, track Track) error

	// UpdatePosition updates the current playback position for timing calculations.
	UpdatePosition(position time.Duration, paused bool)
	// ShouldScrobble returns true if enough of the track has been played to scrobble.
	ShouldScrobble() bool

	// SavePending persists any pending offline scrobbles to disk.
	SavePending() error
	// LoadPending loads pending offline scrobbles from disk.
	LoadPending() error
	// PendingCount returns the number of pending offline scrobbles.
	PendingCount() int
	// FlushPending attempts to submit all pending offline scrobbles.
	FlushPending(ctx context.Context) error
}

// Manager coordinates multiple scrobblers, fanning out events to all enabled backends.
type Manager struct {
	mu         sync.RWMutex
	scrobblers []Scrobbler
	wg         sync.WaitGroup
}

// NewManager creates a new scrobbler manager.
func NewManager() *Manager {
	return &Manager{}
}

// Register adds a scrobbler to the manager.
func (m *Manager) Register(s Scrobbler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scrobblers = append(m.scrobblers, s)
}

// Scrobblers returns all registered scrobblers.
func (m *Manager) Scrobblers() []Scrobbler {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Scrobbler, len(m.scrobblers))
	copy(result, m.scrobblers)
	return result
}

// EnabledCount returns the number of enabled scrobblers.
func (m *Manager) EnabledCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, s := range m.scrobblers {
		if s.IsEnabled() {
			count++
		}
	}
	return count
}

// NowPlaying reports the currently playing track to all enabled scrobblers.
// Errors are logged but not returned to avoid blocking playback.
func (m *Manager) NowPlaying(ctx context.Context, track Track) {
	m.mu.RLock()
	scrobblers := m.scrobblers
	m.mu.RUnlock()

	for _, s := range scrobblers {
		if s.IsEnabled() {
			// Fire and forget - don't block on network
			m.wg.Add(1)
			go func(scrobbler Scrobbler) {
				defer m.wg.Done()
				_ = scrobbler.NowPlaying(ctx, track)
			}(s)
		}
	}
}

// Scrobble reports a completed track to all enabled scrobblers.
func (m *Manager) Scrobble(ctx context.Context, track Track) {
	m.mu.RLock()
	scrobblers := m.scrobblers
	m.mu.RUnlock()

	for _, s := range scrobblers {
		// Call on all scrobblers - they handle queueing if not enabled
		m.wg.Add(1)
		go func(scrobbler Scrobbler) {
			defer m.wg.Done()
			_ = scrobbler.Scrobble(ctx, track)
		}(s)
	}
}

// Wait blocks until all in-flight scrobble operations complete or the context is canceled.
func (m *Manager) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// UpdatePosition updates playback position on all scrobblers.
func (m *Manager) UpdatePosition(position time.Duration, paused bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.scrobblers {
		s.UpdatePosition(position, paused)
	}
}

// ShouldScrobble returns true if any scrobbler thinks the track should be scrobbled.
func (m *Manager) ShouldScrobble() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.scrobblers {
		if s.ShouldScrobble() {
			return true
		}
	}
	return false
}

// SavePending saves pending scrobbles for all scrobblers.
func (m *Manager) SavePending() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.scrobblers {
		if err := s.SavePending(); err != nil {
			return err
		}
	}
	return nil
}

// LoadPending loads pending scrobbles for all scrobblers.
func (m *Manager) LoadPending() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.scrobblers {
		if err := s.LoadPending(); err != nil {
			return err
		}
	}
	return nil
}

// FlushPending flushes pending scrobbles for all enabled scrobblers.
func (m *Manager) FlushPending(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.scrobblers {
		if s.IsEnabled() {
			if err := s.FlushPending(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// TotalPendingCount returns total pending scrobbles across all scrobblers.
func (m *Manager) TotalPendingCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0
	for _, s := range m.scrobblers {
		total += s.PendingCount()
	}
	return total
}
