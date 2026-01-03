# Tunez — Product Requirements Document (PRD)

**Status:** Draft  
**Last updated:** 2026-01-02  
**Stack (normative):** Go + Bubble Tea + mpv

## 1. Overview

Tunez is a terminal-first music player. It provides a rich, keyboard-driven TUI for browsing and searching a music library and controlling playback. Tunez supports multiple “Providers” (music sources) compiled into the binary.

**Phase 1 includes two built-in Providers:**
- **Filesystem Provider**: local directories (tags-based library + folder fallback)
- **Melodee API Provider**: remote server providing library/search/streaming via HTTPS

Tunez uses **mpv** as the playback engine. Tunez controls mpv via **JSON IPC** so playback is robust, performant, and consistent across platforms.

## 2. Goals and Non-goals

### 2.1 Goals (MVP / Phase 1)
1. **Responsive playback control**: play/pause/seek/next/prev/volume with immediate UI feedback.
2. **Fast browsing & search**:
   - Local libraries: handle hundreds of albums smoothly.
   - Remote libraries: handle *hundreds of thousands* of albums/tracks using paging and incremental loading.
3. **Keyboard-first TUI**: browsing views, search, queue, now playing, help overlay.
4. **Cross-platform**: Linux/macOS/Windows support (terminal + mpv present).
5. **Modular provider architecture**: adding a new Provider is straightforward and does not require rewriting core UI/player logic.

### 2.2 Non-goals (MVP / Phase 1)
- Plugin marketplace or runtime plugin loading
- Multi-room audio/casting
- Full “Spotify client” feature set (social features, collaborative playlists)
- Metadata editing/writing tags (read-only library in Phase 1)
- DRM/protected content support

## 3. Users and Use Cases

### 3.1 Primary user persona
- Power user who lives in terminals (tmux/SSH)
- Prefers keyboard-only workflows
- Maintains local libraries and/or self-hosted music servers
- Comfortable editing config files

### 3.2 Core user stories
- Select a Provider profile and browse artists/albums/playlists.
- Search songs/albums/artists and start playback immediately.
- Queue songs and manage the play queue.
- See now playing status (track, album, artist, elapsed/remaining).
- Control playback with hotkeys.
- Rebind keys and see help reflect current bindings.

## 4. Functional Requirements

### 4.1 Provider architecture
**MUST**
- Tunez MUST define a stable internal Provider interface (see `docs/PROVIDERS.md`).
- Providers MUST support incremental paging for large datasets.
- Providers MUST clearly expose capability support (e.g., playlists, lyrics, artwork).
- Provider errors MUST be normalized (NotSupported, NotFound, Unauthorized, Offline, RateLimited, Temporary, etc.).

**MUST (Phase 1 built-ins)**
- Filesystem Provider (see `docs/PROVIDER_FILESYSTEM.md`)
- Melodee API Provider (see `docs/PROVIDER_MELODEE_API.md`)

### 4.2 Library browsing
Tunez MUST provide these browse views (capability-gated where appropriate):
- **Artists**
- **Albums** (all, and filtered by Artist)
- **Tracks** (filtered by Album/Artist/Playlist)
- **Playlists** (when Provider supports)

**UX requirement:** browsing MUST remain responsive while items are loading; long lists MUST paginate/infinite-scroll.

### 4.3 Search
**MUST**
- Global search (`/`) from any screen.
- Search returns best-effort results across: tracks, albums, artists (and playlists when supported).
- Selecting a result can:
  - Start playback immediately (tracks)
  - Open the relevant browse view (album/artist/playlist)

### 4.4 Playback
Tunez uses mpv for playback.

**MUST**
- Play/pause toggle
- Seek forward/back
- Next/previous track (in queue)
- Volume up/down + mute
- Shuffle + repeat modes (at least: off / all / one)
- Display elapsed and remaining time and a progress bar

**SHOULD**
- Gapless playback where mpv supports it
- Preloading next track (player-side) to minimize delay

### 4.5 Queue management
**MUST**
- Add track(s) to queue
- Play next / play later
- Remove from queue
- Clear queue
- Jump to a queue item
- Persist queue across restarts (optional for MVP; required for v1)

### 4.6 TUI screens
Tunez MUST provide:
- Splash/loading screen
- Now Playing
- Search results
- Library (artists/albums/tracks)
- Playlists (capability-gated)
- Queue
- Help / keybindings overlay
- Configuration (view-only in MVP; edit via file is acceptable)

### 4.7 Configuration & profiles
**MUST**
- Config file on disk (`config.toml`) with:
  - one or more Provider profiles
  - keybindings
  - playback settings (mpv path, cache settings, initial volume)
  - UI settings (theme, page sizes)
- Tunez MUST validate config on startup and show actionable errors without leaking secrets.
- Tunez MUST support selecting an active profile at runtime (menu).

**SHOULD**
- A `tunez config init` command that writes an example config.
- A `tunez doctor` command that checks mpv availability and provider connectivity.

### 4.8 Observability
**MUST**
- Structured logging to file with log level control.
- A visible “status line” for transient errors (e.g., provider timeout).

**SHOULD**
- A debug overlay showing last provider request latency and cache hits.

## 5. Non-functional Requirements

### 5.1 Performance targets
- UI frame loop MUST remain responsive under input (no “stuck” keypresses).
- Provider fetches MUST be asynchronous; UI MUST not block.
- For remote providers, list screens MUST load the first page within a reasonable time (target: < 1s on a healthy LAN; degrade gracefully).

### 5.2 Reliability
- Playback commands should not crash the app if mpv is missing; show actionable remediation.
- Network failures should be handled with retries/backoff and user-visible status.

### 5.3 Security & privacy
- Secrets (API tokens) MUST not be written to logs.
- Config may contain tokens; support OS keychain later (v1+).

## 6. MVP Acceptance Criteria (Definition of Done)
MVP is accepted when all items below are true:

1. **mpv playback control works** (play/pause/seek/next/prev/volume) via IPC on Linux/macOS/Windows.
2. **Filesystem Provider works**: scan/index local library; browse artists/albums/tracks; play a local file.
3. **Melodee API Provider works**: authenticate; browse artists/albums/playlists; search; play a streamed track.
4. **TUI is complete**: Now Playing, Library, Search, Queue, Help overlay.
5. **Config works**: load + validate config; switch profiles; keybindings apply.
6. **No UI blocking**: long provider requests show loading state and remain cancellable.

## 7. Roadmap (Post-MVP)
- Offline cache/download (provider-gated)
- Lyrics view (embedded + remote)
- Artwork (inline placeholders + optional image-to-ANSI)
- Scrobbling (e.g., Last.fm)
- Playlists create/edit (provider-gated)
- Advanced library sorting/filtering (genre/year/rating)
