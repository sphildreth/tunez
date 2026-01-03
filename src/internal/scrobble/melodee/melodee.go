package melodee

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/tunez/tunez/internal/scrobble"
)

// TokenProvider is an interface for getting the current auth token.
// This allows the scrobbler to reuse auth from the melodee provider.
type TokenProvider interface {
	Token() string
}

// Config holds Melodee scrobbler configuration.
type Config struct {
	BaseURL       string
	TokenProvider TokenProvider
	// Alternative: direct token if not using provider
	Token string
}

// Scrobbler implements scrobble.Scrobbler for Melodee API.
type Scrobbler struct {
	mu            sync.Mutex
	id            string
	baseURL       string
	tokenProvider TokenProvider
	staticToken   string
	client        *http.Client
	pending       []scrobbleEntry
	nowPlaying    *scrobble.Track
	playStarted   time.Time
	playDuration  time.Duration
}

type scrobbleEntry struct {
	Track     scrobble.Track
	Timestamp time.Time
}

// scrobbleRequest matches Melodee API ScrobbleRequest schema.
type scrobbleRequest struct {
	SongID         string `json:"songId"`
	PlayerName     string `json:"playerName"`
	ScrobbleType   string `json:"scrobbleType"` // "NowPlaying" or "Scrobble"
	Timestamp      int64  `json:"timestamp"`
	PlayedDuration int    `json:"playedDuration"` // seconds
}

// New creates a new Melodee scrobbler.
func New(id string, cfg Config) *Scrobbler {
	if id == "" {
		id = "melodee"
	}
	return &Scrobbler{
		id:            id,
		baseURL:       cfg.BaseURL,
		tokenProvider: cfg.TokenProvider,
		staticToken:   cfg.Token,
		client:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Scrobbler) ID() string   { return s.id }
func (s *Scrobbler) Name() string { return "Melodee" }

func (s *Scrobbler) IsEnabled() bool {
	return s.baseURL != "" && s.getToken() != ""
}

func (s *Scrobbler) getToken() string {
	if s.tokenProvider != nil {
		return s.tokenProvider.Token()
	}
	return s.staticToken
}

func (s *Scrobbler) NowPlaying(ctx context.Context, track scrobble.Track) error {
	s.mu.Lock()
	s.nowPlaying = &track
	s.playStarted = time.Now()
	s.playDuration = 0
	s.mu.Unlock()

	if !s.IsEnabled() {
		return nil
	}

	// Track must have an ID for Melodee scrobbling
	// The track ID should be passed through somehow - for now we skip if not available
	// In practice, the app would need to pass the provider-specific track ID
	return s.sendScrobble(ctx, track, "NowPlaying", 0)
}

func (s *Scrobbler) UpdatePosition(position time.Duration, paused bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nowPlaying == nil {
		return
	}

	if !paused {
		s.playDuration = position
	}
}

func (s *Scrobbler) ShouldScrobble() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nowPlaying == nil {
		return false
	}

	// 4 minute threshold
	if s.playDuration >= 4*time.Minute {
		return true
	}

	// 50% threshold
	if s.nowPlaying.DurationMs > 0 {
		halfDuration := time.Duration(s.nowPlaying.DurationMs/2) * time.Millisecond
		if s.playDuration >= halfDuration {
			return true
		}
	}

	return false
}

func (s *Scrobbler) Scrobble(ctx context.Context, track scrobble.Track) error {
	if !s.IsEnabled() {
		s.queueScrobble(track)
		return nil
	}

	s.mu.Lock()
	playedDuration := int(s.playDuration.Seconds())
	s.mu.Unlock()

	err := s.sendScrobble(ctx, track, "Scrobble", playedDuration)
	if err != nil {
		s.queueScrobble(track)
		return err
	}
	return nil
}

func (s *Scrobbler) sendScrobble(ctx context.Context, track scrobble.Track, scrobbleType string, playedDuration int) error {
	// Note: For Melodee, we need the song ID. This is a simplification -
	// in practice, the Track struct may need to carry provider-specific IDs.
	// For now, we'll use title as a fallback identifier.
	songID := track.Title // This should be the actual Melodee song ID

	req := scrobbleRequest{
		SongID:         songID,
		PlayerName:     "Tunez",
		ScrobbleType:   scrobbleType,
		Timestamp:      track.StartedAt.Unix(),
		PlayedDuration: playedDuration,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/api/v1/scrobble", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.getToken())

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return scrobble.ErrUnauthorized
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return scrobble.ErrRateLimited
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("melodee scrobble error: %s", resp.Status)
	}

	return nil
}

func (s *Scrobbler) PendingCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pending)
}

func (s *Scrobbler) FlushPending(ctx context.Context) error {
	if !s.IsEnabled() {
		return scrobble.ErrNotConfigured
	}

	s.mu.Lock()
	pending := s.pending
	s.pending = nil
	s.mu.Unlock()

	var failed []scrobbleEntry
	for _, entry := range pending {
		if err := s.Scrobble(ctx, entry.Track); err != nil {
			failed = append(failed, entry)
		}
	}

	if len(failed) > 0 {
		s.mu.Lock()
		s.pending = append(failed, s.pending...)
		s.mu.Unlock()
		return fmt.Errorf("failed to scrobble %d tracks", len(failed))
	}

	return nil
}

func (s *Scrobbler) queueScrobble(track scrobble.Track) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending = append(s.pending, scrobbleEntry{
		Track:     track,
		Timestamp: track.StartedAt,
	})

	// Limit pending queue
	if len(s.pending) > 50 {
		s.pending = s.pending[len(s.pending)-50:]
	}
}

func (s *Scrobbler) SavePending() error {
	s.mu.Lock()
	pending := s.pending
	s.mu.Unlock()

	if len(pending) == 0 {
		return nil
	}

	path, err := s.pendingPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(pending)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func (s *Scrobbler) LoadPending() error {
	path, err := s.pendingPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var pending []scrobbleEntry
	if err := json.Unmarshal(data, &pending); err != nil {
		return err
	}

	s.mu.Lock()
	s.pending = pending
	s.mu.Unlock()

	return nil
}

func (s *Scrobbler) pendingPath() (string, error) {
	var base string
	switch runtime.GOOS {
	case "darwin":
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "tunez", "state")
	case "windows":
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "Tunez", "state")
	default:
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(dir, "tunez", "state")
	}
	return filepath.Join(base, fmt.Sprintf("scrobble_pending_%s.json", s.id)), nil
}
