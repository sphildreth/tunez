# Tunez — Melodee.API Provider — PRD

**Status:** Draft  
**Last updated:** 2026-01-01  
**Applies to:** Tunez Phase 1 (built-in Providers)

## 1. Overview

This document specifies a built-in Tunez Provider that integrates with **Melodee.API (v1)** so Tunez can browse/search a Melodee library and play tracks via streaming URLs.

**Melodee API spec source:** [docs/melodee-api-v1.json](docs/melodee-api-v1.json)

### 1.1 Goals
- Allow Tunez to authenticate to a Melodee server.
- Allow Tunez to search and browse songs, artists, albums, and playlists.
- Allow Tunez to play songs by resolving a stream URL from Melodee.

### 1.2 Non-goals
- Implementing Melodee server features unrelated to playback (requests, shares, analytics, charts, equalizer presets, etc.).
- Using Melodee’s server-side scrobble endpoints; Tunez uses its own Scrobbler abstraction.
- Offline download support.
- Administrative / editor workflows.

### 1.3 Assumptions / Open Questions
Some details are not explicit in the provided OpenAPI document and must be verified during implementation:
- **How API auth is applied to requests** (the spec defines `401` responses and returns a `token`, but does not define a `securitySchemes` section or required auth header/cookie).
- **Whether `Song.streamUrl` is absolute or relative**.
- **Whether the stream URL always works as-is**, or if Tunez must construct `/song/stream/{apiKey}/{userApiKey}/{authToken}`.

This PRD defines the expected Tunez behavior and the best-effort fallback rules when these values are ambiguous.

---

## 2. Provider Identity and Configuration

### 2.1 Provider ID
- Provider ID: `melodee`
- Display name: `Melodee`

### 2.2 Required Configuration
Tunez MUST support configuring a Melodee connection profile with:
- `base_url`: Melodee server base URL (example from spec: `http://localhost:5157/`).

### 2.3 Credentials
Tunez MUST support authenticating with one of:
- Username + password (`/api/v1/auth/authenticate`)
- Refresh token (`/api/v1/auth/refresh-token`) when present

**Security requirements**
- Tunez MUST NOT store plaintext passwords or tokens in a config file.
- Tunez MUST store access token + refresh token (if provided) in the OS keyring.
- Tunez MUST NOT log tokens, refresh tokens, or Authorization headers.

---

## 3. Capabilities

### 3.1 Capability Flags
The Melodee Provider SHOULD advertise the following capabilities:
- `supports_search_tracks`: yes
- `supports_browse_artists`: yes
- `supports_browse_albums`: yes
- `supports_playlists`: yes
- `supports_lyrics`: yes (via `/api/v1/songs/{id}/lyrics`)
- `supports_artwork`: yes (via `thumbnailUrl` / `imageUrl` fields)
- `supports_offline_download`: no

If any capability cannot be supported due to missing server permissions or unavailable endpoints, the Provider MUST return `NotSupported` for those operations.

---

## 4. Functional Requirements (EARS)

### 4.1 Authentication
- WHEN the user selects the `melodee` Provider and no valid access token is available, THE SYSTEM SHALL prompt for credentials (per Tunez core UX) and authenticate via `POST /api/v1/auth/authenticate`.
- WHEN the Melodee server returns an `AuthenticationResponse`, THE SYSTEM SHALL store the returned `token` and `expiresAt` securely.
- WHEN the server returns `refreshToken` and `refreshTokenExpiresAt`, THE SYSTEM SHALL store the refresh token securely.
- WHEN an authenticated request receives `401 Unauthorized`, THE SYSTEM SHALL attempt a token refresh (if a refresh token exists) and then retry the request once.
- IF refresh is unavailable or fails, THEN THE SYSTEM SHALL surface an authentication error to the user and stop retrying.

### 4.2 Search Tracks
- WHEN the user performs a track search, THE SYSTEM SHALL query Melodee using `GET /api/v1/search/songs`.
- WHEN `page`/`pageSize` are used, THE SYSTEM SHALL map Tunez paging to Melodee paging and respect `PaginationMetadata` (`totalCount`, `pageSize`, `currentPage`, `totalPages`).

### 4.3 Track Metadata
- WHEN Tunez needs full track metadata (e.g., for Now Playing), THE SYSTEM SHALL fetch it via `GET /api/v1/songs/{id}`.

### 4.4 Stream URL Resolution (Provider Stream Contract)
Tunez Phase 1 provider contract is “**provider returns a stream URL only**”.

- WHEN Tunez calls `get_stream_url(track_id)`, THE SYSTEM SHALL produce a playable URL without embedding secrets into logs.
- THE SYSTEM SHALL prefer `Song.streamUrl` from Melodee as the stream URL.
- IF `Song.streamUrl` is a relative URL, THEN THE SYSTEM SHALL join it with `base_url`.
- IF `Song.streamUrl` is missing/empty, THEN THE SYSTEM MAY fall back to constructing `/song/stream/{apiKey}/{userApiKey}/{authToken}` only if all required values are available.

### 4.5 Playlists
- WHEN the user lists playlists, THE SYSTEM SHALL call `GET /api/v1/user/playlists`.
- WHEN the user opens a playlist, THE SYSTEM SHALL call `GET /api/v1/playlists/{apiKey}/songs` to obtain tracks.

### 4.6 Albums and Artists (Browsing)
- WHEN the user browses artists, THE SYSTEM SHALL call `GET /api/v1/artists` and support paging.
- WHEN the user browses albums, THE SYSTEM SHALL call `GET /api/v1/albums` and support paging.
- WHEN the user views an album’s tracks, THE SYSTEM SHALL call `GET /api/v1/albums/{id}/songs`.
- WHEN the user views an artist’s tracks, THE SYSTEM SHALL call `GET /api/v1/artists/{id}/songs`.

### 4.7 Lyrics
- WHEN the user requests lyrics for a track and lyrics are available, THE SYSTEM SHALL call `GET /api/v1/songs/{id}/lyrics`.
- IF lyrics are unavailable, THEN THE SYSTEM SHALL return `NotFound` or `NotSupported` based on server response.

---

## 5. API Mapping (Tunez ↔ Melodee)

### 5.1 Base URL
- Base URL is user-configured.
- Only `http` and `https` schemes SHOULD be allowed.

### 5.2 Endpoints Used

**Authentication**
- `POST /api/v1/auth/authenticate` → returns `AuthenticationResponse { token, expiresAt, refreshToken?, refreshTokenExpiresAt?, user }`
- `POST /api/v1/auth/refresh-token` (with `RefreshTokenRequest { refreshToken }`) → returns `AuthenticationResponse`

**Search**
- `GET /api/v1/search/songs?q=&page=&pageSize=&filterByArtistApiKey=` → returns `SongPagedResponse { meta, data: Song[] }`
- (Optional, not required for MVP) `POST /api/v1/search` → returns `SearchResultResponse { meta, data: SearchResult }`

**Songs**
- `GET /api/v1/songs/{id}` → returns `Song`
- `GET /api/v1/songs/{id}/lyrics` → returns `Lyrics`

**Streaming**
- Primary: `Song.streamUrl`
- Fallback (if needed): `GET /song/stream/{apiKey}/{userApiKey}/{authToken}` (+ optional header `X-Api-Version`)

**Playlists**
- `GET /api/v1/user/playlists?page=&limit=` → returns `PlaylistPagedResponse`
- `GET /api/v1/playlists/{apiKey}/songs?page=&pageSize=` → returns `SongPagedResponse`

**Albums / Artists**
- `GET /api/v1/albums?page=&pageSize=` → returns `AlbumPagedResponse`
- `GET /api/v1/albums/{id}` → returns `Album`
- `GET /api/v1/albums/{id}/songs` → returns `SongPagedResponse`
- `GET /api/v1/artists?page=&pageSize=` → returns `ArtistPagedResponse`
- `GET /api/v1/artists/{id}` → returns `Artist`
- `GET /api/v1/artists/{id}/albums` → returns `AlbumPagedResponse`
- `GET /api/v1/artists/{id}/songs` → returns `SongPagedResponse`

---

## 6. Data Model Mapping

### 6.1 Track Identity
- Tunez track id MUST be treated as an opaque provider-scoped string.
- For Melodee, Tunez SHOULD use `Song.id` (UUID string) as the provider track id.

### 6.2 Track Metadata (minimum)
Tunez track model SHOULD map:
- Title: `Song.title`
- Artist name: `Song.artist.name`
- Album title: `Song.album.name`
- Duration seconds: `floor(Song.durationMs / 1000)`
- Artwork: `Song.thumbnailUrl` (small), `Song.imageUrl` (large)
- Stream URL: `Song.streamUrl`

### 6.3 Paging
Melodee uses `PaginationMetadata`:
- `totalCount`, `pageSize`, `currentPage`, `totalPages`, `hasPrevious`, `hasNext`

Tunez SHOULD preserve deterministic ordering within a query while paging.

---

## 7. Error Handling Requirements

- The Provider MUST map HTTP `401` to `AuthenticationError`.
- The Provider MUST map HTTP `404` to `NotFound`.
- The Provider MUST map HTTP `400` to a user-visible `Other` (or `InvalidRequest`) error.
- Network failures MUST map to `NetworkError`.
- The Provider MUST NOT panic on API errors.

---

## 8. Performance and UX Requirements

- The Provider MUST be non-blocking to the TUI (network calls run off the UI thread / task).
- The Provider SHOULD apply basic caching for entity lookups (e.g., `get_track` after search), bounded by size/time.
- The Provider SHOULD clamp requested page sizes to a safe maximum if Tunez requests an unreasonably large page.

---

## 9. Validation (Acceptance Criteria)

### 9.1 MVP Acceptance Criteria
- A user can configure `base_url`, authenticate, search songs, select a song, and begin playback using the returned stream URL.
- A user can list playlists and enqueue tracks from a playlist.
- Errors are surfaced as user-friendly messages and do not crash Tunez.

### 9.2 Test Expectations (Provider-level)
- Unit tests for:
  - URL joining behavior for `Song.streamUrl`
  - Paging translation and metadata handling
  - Error mapping (`401`, `404`, network)
- Contract tests using mocked HTTP responses (no real network required).
