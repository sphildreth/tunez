# Tunez — Filesystem Provider — PRD

**Status:** Draft  
**Last updated:** 2026-01-01  
**Applies to:** Tunez Phase 1 (built-in Providers)

## 1. Overview

This document specifies a built-in Tunez Provider that integrates with the **local filesystem** (e.g., `/home/<user>/Music`) so Tunez can scan a music library, browse/search, and play tracks from local files.

### 1.1 Goals
- Allow Tunez to load a local library rooted at one or more directories.
- Allow Tunez to search and browse tracks/artists/albums.
- Allow Tunez to play tracks by returning a local stream URL.

### 1.2 Non-goals
- Managing or editing audio metadata tags (write operations).
- Downloading / caching remote content.
- Online lyrics fetching (local embedded lyrics only, if present).
- DRM/protected formats.

### 1.3 Assumptions / Open Questions
- Which audio codecs are supported is primarily determined by Tunez’s audio pipeline (decoder support). The Provider will only expose playable files based on extension/metadata + what the decoder layer can handle.
- How Tunez wants to represent “browse” for filesystem sources (tag-based library vs. directory tree) should follow the canonical Tunez PRD. This Provider defaults to **tag-based** browsing (Artists/Albums) with directory browsing treated as optional.

---

## 2. Provider Identity and Configuration

### 2.1 Provider ID
- Provider ID: `filesystem`
- Display name: `Filesystem`

### 2.2 Required Configuration
Tunez MUST support configuring a Filesystem provider profile with:
- `roots`: one or more absolute directory paths to scan (examples: `/home/<user>/Music`, `/mnt/media/Music`).

### 2.3 Optional Configuration
- `follow_symlinks`: boolean (default: false)
- `excluded_paths`: list of absolute paths/prefixes to skip
- `extensions_allowlist`: list of file extensions (default: common audio extensions)
- `rescan_on_startup`: boolean (default: true)

**Security requirements**
- The Provider MUST treat configured paths as untrusted input for path traversal purposes.
- The Provider MUST normalize and canonicalize scan roots and ensure all discovered files are within configured roots.

---

## 3. Capabilities

### 3.1 Capability Flags
The Filesystem Provider SHOULD advertise the following capabilities:
- `supports_search_tracks`: yes
- `supports_browse_artists`: yes
- `supports_browse_albums`: yes
- `supports_playlists`: optional (see below)
- `supports_lyrics`: optional (embedded-only)
- `supports_artwork`: optional (embedded cover art or folder images)
- `supports_offline_download`: yes (content is already local)

### 3.2 Playlists Capability
Filesystem does not inherently provide playlists; however, Tunez MAY support playlists via:
- `.m3u` / `.m3u8` playlist files inside scan roots, or
- Tunez-managed playlists stored outside the Provider (core feature).

For Phase 1, the Filesystem Provider SHOULD implement:
- `list_playlists` / `get_playlist` / `list_playlist_tracks` only if `.m3u/.m3u8` is implemented.
- Otherwise it MUST return `ProviderError::NotSupported`.

---

## 4. Functional Requirements (EARS)

### 4.1 Library Scan and Index
- WHEN the user selects the `filesystem` Provider, THE SYSTEM SHALL load (or build) a local index of tracks found under the configured roots.
- WHEN `rescan_on_startup` is enabled, THE SYSTEM SHALL rescan roots on startup to reflect file changes.
- IF a file cannot be read or parsed, THEN THE SYSTEM SHALL skip it, log a non-fatal warning, and continue.

### 4.2 Track Identity
- WHEN the Provider returns tracks, THE SYSTEM SHALL use a stable provider-scoped track id.
- THE SYSTEM SHALL treat filesystem track ids as opaque strings.

Recommended stability rule (Phase 1):
- Track id SHOULD be derived from a canonicalized absolute path (and optionally file metadata such as mtime/size) in a way that is stable across runs.

### 4.3 Search Tracks
- WHEN the user performs a track search, THE SYSTEM SHALL search the local index by title/artist/album (case-insensitive) and return deterministic results.
- WHEN paging is requested, THE SYSTEM SHALL return stable ordering for a given query while paging.

### 4.4 Browse Artists and Albums
- WHEN the user browses artists, THE SYSTEM SHALL return distinct artist entities derived from tags.
- WHEN the user browses albums, THE SYSTEM SHALL return distinct album entities derived from tags.
- WHEN an album is selected, THE SYSTEM SHALL return album tracks ordered by track number when present, otherwise by title.

### 4.5 Track Metadata
- WHEN Tunez requests full track metadata, THE SYSTEM SHALL return cached metadata from the index.
- IF metadata is missing (e.g., no tags), THEN THE SYSTEM SHALL fall back to filename-based fields.

### 4.6 Stream URL Resolution (Provider Stream Contract)
Tunez Phase 1 provider contract is “**provider returns a stream URL only**”.

- WHEN Tunez calls `get_stream_url(track_id)`, THE SYSTEM SHALL return a playable URL for the local file.
- THE SYSTEM SHALL return a `file://` URL for the canonicalized absolute path.
- IF the file no longer exists or is inaccessible, THEN THE SYSTEM SHALL return `NotFound`.

### 4.7 Lyrics (Optional)
- WHEN lyrics are requested and embedded lyrics are available, THE SYSTEM SHALL return them.
- IF lyrics are not available, THEN THE SYSTEM SHALL return `NotFound` or `NotSupported`.

---

## 5. Provider ↔ Tunez Mapping

### 5.1 Tunez Provider Operations
The Filesystem Provider MUST implement, at minimum:
- `search_tracks(query, filters, paging)` → local index search
- `browse(kind: artists|albums|genres, paging)` → local index browse
- `get_track(track_id)` → local index lookup
- `get_stream_url(track_id)` → `file://` URL

The Provider SHOULD implement (if feasible in Phase 1):
- `get_album` / `list_album_tracks`
- `get_artist` / `list_artist_tracks` (or equivalent)

The Provider MAY implement:
- `list_playlists` / `get_playlist` / `list_playlist_tracks` based on `.m3u/.m3u8` discovery

---

## 6. Data Model Mapping

### 6.1 Track Metadata (minimum)
Tunez track model SHOULD include:
- Title
- Primary artist name
- Album title (when present)
- Duration seconds (when known)
- A stable provider-scoped track id

### 6.2 Suggested Tag Sources
- ID3 (MP3)
- Vorbis comments (FLAC/OGG)
- MP4 atoms (M4A)

When tags are missing:
- Title defaults to filename (without extension)
- Artist defaults to `Unknown Artist`
- Album defaults to `Unknown Album`

---

## 7. Error Handling Requirements

- The Provider MUST map permission/IO failures to `Other` (or a dedicated IO error category if Tunez core defines one).
- The Provider MUST NOT panic on unreadable files.
- The Provider MUST handle very large libraries without unbounded memory growth (bounded caches, streaming scan).

---

## 8. Performance and UX Requirements

- The Provider MUST avoid blocking the TUI during scans (scan/indexing runs in a background task).
- The Provider SHOULD persist an index to speed up startup (location defined by Tunez core config rules).
- The Provider SHOULD support incremental rescans (only changed files) when a persisted index exists.

---

## 9. Validation (Acceptance Criteria)

### 9.1 MVP Acceptance Criteria
- A user can configure `roots`, search for tracks, select a track, and begin playback.
- Browsing Artists → Albums → Tracks works for tagged libraries.
- Missing/unreadable tracks do not crash Tunez.

### 9.2 Test Expectations (Provider-level)
- Unit tests for:
  - Canonical path handling and root containment (prevents traversal)
  - Search ranking / deterministic ordering
  - Stream URL generation (`file://`)
- Contract tests using a temporary directory with fixture files (no real music required).
