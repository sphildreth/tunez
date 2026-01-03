package scrobble

import (
	"errors"
	"time"
)

var (
	ErrNotConfigured = errors.New("scrobbling not configured")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrRateLimited   = errors.New("rate limited")
)

// Track represents a track for scrobbling.
type Track struct {
	Title      string
	Artist     string
	Album      string
	DurationMs int
	StartedAt  time.Time
	// ProviderID is the provider-specific track ID (e.g., Melodee song ID).
	// Used by scrobblers that need provider-specific identifiers.
	ProviderID string
}
