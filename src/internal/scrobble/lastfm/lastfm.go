package lastfm

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tunez/tunez/internal/scrobble"
)

const apiURL = "https://ws.audioscrobbler.com/2.0/"

// Config holds Last.fm scrobbler configuration.
type Config struct {
	APIKey     string
	APISecret  string
	SessionKey string
}

// Scrobbler implements scrobble.Scrobbler for Last.fm.
type Scrobbler struct {
	mu           sync.Mutex
	id           string
	apiKey       string
	apiSecret    string
	sessionKey   string
	enabled      bool
	client       *http.Client
	pending      []scrobbleEntry
	nowPlaying   *scrobble.Track
	playStarted  time.Time
	playDuration time.Duration
}

type scrobbleEntry struct {
	Track     scrobble.Track
	Timestamp time.Time
}

// New creates a new Last.fm scrobbler.
func New(id string, cfg Config) *Scrobbler {
	if id == "" {
		id = "lastfm"
	}
	return &Scrobbler{
		id:         id,
		apiKey:     cfg.APIKey,
		apiSecret:  cfg.APISecret,
		sessionKey: cfg.SessionKey,
		enabled:    cfg.APIKey != "" && cfg.APISecret != "",
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Scrobbler) ID() string   { return s.id }
func (s *Scrobbler) Name() string { return "Last.fm" }

func (s *Scrobbler) IsEnabled() bool {
	return s.enabled && s.sessionKey != ""
}

// SetSessionKey sets the session key for authenticated requests.
func (s *Scrobbler) SetSessionKey(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionKey = key
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

	params := map[string]string{
		"method":  "track.updateNowPlaying",
		"track":   track.Title,
		"artist":  track.Artist,
		"album":   track.Album,
		"api_key": s.apiKey,
		"sk":      s.sessionKey,
	}
	if track.DurationMs > 0 {
		params["duration"] = fmt.Sprintf("%d", track.DurationMs/1000)
	}

	return s.signedPost(ctx, params)
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

	params := map[string]string{
		"method":    "track.scrobble",
		"track":     track.Title,
		"artist":    track.Artist,
		"album":     track.Album,
		"timestamp": fmt.Sprintf("%d", track.StartedAt.Unix()),
		"api_key":   s.apiKey,
		"sk":        s.sessionKey,
	}
	if track.DurationMs > 0 {
		params["duration"] = fmt.Sprintf("%d", track.DurationMs/1000)
	}

	err := s.signedPost(ctx, params)
	if err != nil {
		s.queueScrobble(track)
		return err
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

func (s *Scrobbler) signedPost(ctx context.Context, params map[string]string) error {
	params["api_sig"] = s.sign(params)
	params["format"] = "json"

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
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
		return fmt.Errorf("lastfm error: %s", resp.Status)
	}

	var result struct {
		Error   int    `json:"error"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		if result.Error != 0 {
			return fmt.Errorf("lastfm error %d: %s", result.Error, result.Message)
		}
	}

	return nil
}

func (s *Scrobbler) sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "format" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var sig strings.Builder
	for _, k := range keys {
		sig.WriteString(k)
		sig.WriteString(params[k])
	}
	sig.WriteString(s.apiSecret)

	hash := md5.Sum([]byte(sig.String()))
	return hex.EncodeToString(hash[:])
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
	// Use scrobbler ID in filename to support multiple instances
	return filepath.Join(base, fmt.Sprintf("scrobble_pending_%s.json", s.id)), nil
}
