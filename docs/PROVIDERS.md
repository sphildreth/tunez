# Tunez — Provider Interface Contract

**Last updated:** 2026-01-02  
**Applies to:** Tunez (Go + Bubble Tea + mpv)

## 1. Concepts

- **Provider**: a compiled-in Go package that supplies library data and streaming info.
- **Profile**: a runtime configuration instance for a Provider (e.g., a specific server URL + credentials).
- **Capability**: optional features a Provider may support (playlists, lyrics, artwork, etc.).
- **Cursor-based paging**: list operations return a `nextCursor` token; callers request more by passing it back.

## 2. Normalized domain models

Tunez core operates on Provider-neutral models:

- `Artist`: id, name, sortName?, albumCount?, songCount?
- `Album`: id, title, artistName, year?, trackCount?, artworkRef?
- `Track`: id, title, artistName, albumTitle, durationMs?, trackNo?, discNo?, codec?, bitrateKbps?, artworkRef?
- `Playlist`: id, name, trackCount?

**IDs**
- Provider IDs are opaque strings.
- Tunez MUST namespace IDs internally as: `{providerId}:{entityId}` to avoid collisions.

## 3. Provider interface (Go)

The interface below is *normative*. Providers may implement a subset; unsupported calls must return `ErrNotSupported`.

```go
// Package provider defines the contract between Tunez core and Providers.
package provider

import "context"

type Capability string

const (
    CapPlaylists Capability = "playlists"
    CapLyrics    Capability = "lyrics"
    CapArtwork   Capability = "artwork"
)

type Capabilities map[Capability]bool

type ListReq struct {
    Cursor   string // empty for first page
    PageSize int    // core provides a default; Provider may clamp
    Sort     string // optional (e.g., "name", "recent")
}

type Page[T any] struct {
    Items      []T
    NextCursor string
    TotalHint  int // optional; -1 if unknown
}

type StreamInfo struct {
    URL     string            // file:// or https:// etc.
    Headers map[string]string // optional (e.g., Authorization)
}

type Provider interface {
    ID() string
    Name() string
    Capabilities() Capabilities

    // Initialize prepares internal caches/clients for this profile.
    Initialize(ctx context.Context, profileCfg any) error

    // Health reports current availability. Used for status UI.
    Health(ctx context.Context) (ok bool, details string)

    // Browse
    ListArtists(ctx context.Context, req ListReq) (Page[Artist], error)
    GetArtist(ctx context.Context, id string) (Artist, error)

    ListAlbums(ctx context.Context, artistId string, req ListReq) (Page[Album], error) // artistId optional
    GetAlbum(ctx context.Context, id string) (Album, error)

    ListTracks(ctx context.Context, albumId string, artistId string, playlistId string, req ListReq) (Page[Track], error)
    GetTrack(ctx context.Context, id string) (Track, error)

    // Search
    Search(ctx context.Context, q string, req ListReq) (SearchResults, error)

    // Playlists (CapPlaylists)
    ListPlaylists(ctx context.Context, req ListReq) (Page[Playlist], error)
    GetPlaylist(ctx context.Context, id string) (Playlist, error)

    // Playback
    GetStream(ctx context.Context, trackId string) (StreamInfo, error)

    // Optional capabilities
    GetLyrics(ctx context.Context, trackId string) (Lyrics, error)
    GetArtwork(ctx context.Context, ref string, sizePx int) (Artwork, error)
}

type SearchResults struct {
    Tracks    Page[Track]
    Albums    Page[Album]
    Artists   Page[Artist]
    Playlists Page[Playlist]
}
```

## 4. Error contract

Providers MUST map errors to a small normalized set so UI/logic is consistent:

- `ErrNotSupported`
- `ErrNotFound`
- `ErrUnauthorized`
- `ErrOffline` (no network or server unreachable)
- `ErrRateLimited` (include retry-after if known)
- `ErrTemporary` (timeouts, 5xx)
- `ErrInvalidConfig`

Tunez core will:
- display user-friendly messages
- retry certain temporary failures (with backoff)
- disable capability UI affordances when unsupported

## 5. Paging contract

**MUST**
- For large datasets, Providers MUST support paging using `Cursor` + `NextCursor`.
- Providers MUST return stable ordering for a given sort mode.

**SHOULD**
- For remote sources, cursor tokens SHOULD be opaque and reflect backend paging/offsets.
- For local sources, cursor may be an offset in sorted order (stable).

## 6. Caching expectations

- Providers MAY maintain their own internal caches.
- Tunez core MAY also cache pages for fast back/forward navigation.
- Providers MUST not cache secrets in logs or expose them in error strings.
