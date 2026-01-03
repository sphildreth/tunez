# Tunez — Melodee API Provider (Built-in)

**Last updated:** 2026-01-02  
**Provider ID:** `melodee`

## Overview
Connects to a Melodee server over HTTPS for browsing/searching and streaming tracks. The full API specification is defined in [melodee-api-v1.json](melodee-api-v1.json).

## Capabilities
- **Playlists**: Supported via `/api/v1/user/playlists` and `/api/v1/playlists/{id}/songs`.
- **Lyrics**: Supported via `Song.lyrics` field in song details.
- **Artwork**: Supported via `Album.thumbnailUrl` and `Album.imageUrl`.
- **Scrobbling**: Supported via scrobble endpoints (v1+).

## Authentication
- **Initial Auth**: `POST /api/v1/auth/authenticate` with username/password.
  - Request: `{"username": "user", "password": "pass"}`
  - Response: `{"accessToken": "...", "refreshToken": "...", "expiresIn": 3600}`
- **Token Refresh**: `POST /api/v1/auth/refresh-token` with refresh token before expiry.
- **Error Mapping**: 401/403 → `ErrUnauthorized`, 404 → `ErrNotFound`, 429 → `ErrRateLimited`.

## Endpoints (Phase 1)
- `POST /api/v1/auth/authenticate` — Authenticate user.
- `POST /api/v1/auth/refresh-token` — Refresh access token.
- `GET /api/v1/artists?page=&pageSize=` — List artists with paging.
- `GET /api/v1/artists/{id}/albums` — Get albums for artist.
- `GET /api/v1/albums?page=&pageSize=` — List albums with paging.
- `GET /api/v1/albums/{id}/songs` — Get tracks for album.
- `GET /api/v1/user/playlists?page=&limit=` — List user playlists.
- `GET /api/v1/playlists/{id}/songs?page=&pageSize=` — Get tracks in playlist.
- `GET /api/v1/search/songs?q=&page=&pageSize=` — Search tracks.
- `GET /api/v1/songs/{id}` — Get song details (for streaming URL).

## Data Mapping
- **Artist**: Maps `Artist` schema to provider `Artist` (id, name, albumCount, songCount).
- **Album**: Maps `Album` schema to provider `Album` (id, title, artistName, year, trackCount, artworkRef).
- **Track**: Maps `Song` schema to provider `Track` (id, title, artistName, albumTitle, durationMs, trackNo, codec, bitrateKbps, artworkRef, streamUrl).
- **Playlist**: Maps `Playlist` schema to provider `Playlist` (id, name, trackCount).

## Streaming
- Use `Song.streamUrl` from `GET /api/v1/songs/{id}`.
- If additional headers are needed (e.g., auth), include in `StreamInfo.Headers`.
- Supports gapless playback if mpv handles it.

## Error Handling
- Network timeouts: Map to `ErrTemporary`.
- Server errors (5xx): Map to `ErrTemporary`.
- Client errors: As above.
- All errors logged without secrets.

## MVP Acceptance
- Browse/search/play streamed tracks without blocking UI.
- Handle auth refresh transparently.
- Support paging for large libraries.
