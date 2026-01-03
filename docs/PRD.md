# Tunez — Product Requirements Document (PRD)

**Status:** Active Development  
**Last updated:** 2026-01-03  
**Stack:** Go + Bubble Tea + Lip Gloss + mpv

---

## Overview

Tunez is a terminal-first music player with a rich, keyboard-driven TUI for browsing music libraries and controlling playback. It supports multiple compiled-in "Providers" (music sources) and uses **mpv** as the playback engine via JSON IPC.

**Built-in Providers:**
- **Filesystem Provider**: Local directories with tag-based indexing (SQLite)
- **Melodee API Provider**: Remote server via HTTPS

---

## Phase Status

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 (MVP) | ✅ Complete | Core playback, browsing, TUI screens |
| Phase 2 (v1) | 🟡 In Progress | Lyrics, artwork, caching, themes |
| Phase 3 (v2) | 🔲 Not Started | Command palette, CLI flow, polish |

---

## Phase 1: MVP (COMPLETE)

### Summary
Phase 1 delivers a fully functional terminal music player with:
- mpv playback control (play/pause/seek/next/prev/volume/shuffle/repeat)
- Two providers (Filesystem + Melodee) with browse/search/play
- Complete TUI with 11 screens per TUI_UX.md spec
- Configurable keybindings
- NO_COLOR accessibility support

### Acceptance Criteria (All Met)
1. ✅ mpv playback works via IPC on Linux/macOS
2. ✅ Filesystem provider: scan/index/browse/play local files
3. ✅ Melodee provider: authenticate/browse/search/stream
4. ✅ TUI complete: Loading, Now Playing, Library, Search, Queue, Playlists, Lyrics (placeholder), Config, Help overlay, Error handling
5. ✅ Config: load/validate/switch profiles, keybindings from config
6. ✅ Non-blocking UI: all I/O via tea.Cmd
7. ✅ NO_COLOR support with high-contrast fallback

---

## Phase 2: v1 Features

**Goal:** Enhanced features for daily use - lyrics, artwork, caching, persistence, themes.

### 2.1 Queue Persistence

**Priority:** HIGH  
**Complexity:** Medium  
**Status:** ✅ Complete

Persist the play queue across application restarts.

#### Requirements
- Store queue in SQLite database (`~/.config/tunez/state/queue.db`)
- Restore queue on startup (paused at position 0)
- Clear queue when switching profiles
- Handle missing files gracefully (remove from queue, show toast)

#### Implementation Tasks
```
[x] Create queue persistence schema in internal/queue/persistence.go
    - Table: queue_items (position, track_id, provider_id, track_json, added_at)
    - Table: queue_state (current_index, shuffle_enabled, repeat_mode)

[x] Add Save() method to PersistenceStore
    - Serialize current state to SQLite
    - Called after any queue modification

[x] Add Load() method to PersistenceStore
    - Read from SQLite on startup
    - Validate tracks still exist (provider.GetTrack)
    - Remove invalid entries with toast notification

[x] Update app.go Init() to restore queue
    - Load queue after provider initialization
    - Set status "Restored X tracks" or "Queue empty"

[x] Add config option: queue.persist (bool, default: true)

[x] Add tests for persistence
    - Save/load round-trip
    - Handle corrupted database
    - Handle missing tracks
```

#### Files Modified
- `internal/queue/persistence.go` - SQLite persistence store
- `internal/queue/persistence_test.go` - Persistence tests
- `internal/app/app.go` - Queue restoration on init, save on modifications
- `internal/config/config.go` - QueueConfig section
- `cmd/tunez/main.go` - Initialize persistence store

---

### 2.2 Lyrics Display (Functional)

**Priority:** HIGH  
**Complexity:** Medium  
**Status:** ✅ Complete

Display lyrics for the currently playing track.

#### Requirements
- Fetch lyrics from provider (capability-gated: CapLyrics)
- Display in Screen 6 with scrolling
- Handle loading/error/no-lyrics states
- Sync scroll position with playback (basic: paragraph-level)

#### Implementation Tasks
```
[x] Add GetLyrics to provider interface (already defined)
    - Filesystem: read from embedded tags or .lrc sidecar files
    - Melodee: GET /api/v1/songs/{id} returns lyrics field

[x] Implement GetLyrics for Filesystem provider
    - Check ID3v2 USLT frame for embedded lyrics
    - Check for {filename}.lrc sidecar file
    - Check for {filename}.txt sidecar file
    - Return ErrNotSupported if neither found

[x] Implement GetLyrics for Melodee provider
    - Parse lyrics from Song response
    - Handle plain text and timestamped formats

[x] Create lyrics state in app Model
    - lyrics string
    - lyricsLoading bool
    - lyricsError error
    - lyricsScrollOffset int
    - lyricsTrackID string

[x] Add lyricsCmd to fetch lyrics when track changes
    - Triggered by playTrackMsg
    - Cancel previous fetch on track change (via trackID check)

[x] Update renderLyrics() for functional display
    - Show loading spinner during fetch
    - Show "No lyrics available" when empty
    - Show lyrics text with scroll support
    - Handle j/k for scroll, g/G for top/bottom
    - Strip LRC timestamps for display

[x] Add lyrics loading indicator in Now Playing screen
    - Lyrics screen accessible via navigation
```

#### Files Modified
- `internal/providers/filesystem/provider.go` - GetLyrics with ID3 USLT and sidecar support
- `internal/providers/melodee/provider.go` - GetLyrics from API
- `internal/app/app.go` - Lyrics state, fetchLyricsCmd, renderLyrics, keybindings

---

### 2.3 Artwork Display

**Priority:** MEDIUM  
**Complexity:** High  
**Status:** ✅ Complete

Display album artwork in the TUI.

#### Requirements
- Fetch artwork from provider (capability-gated: CapArtwork)
- Convert to ANSI art for terminal display
- Show in Now Playing screen
- Cache artwork locally

#### Implementation Tasks
```
[x] Add GetArtwork to provider implementations
    - Filesystem: extract from audio files or folder.jpg
    - Melodee: use thumbnailUrl from Album response

[x] Create internal/artwork package
    - Image download with caching
    - Image-to-ANSI conversion using 256-color mode
    - Configurable size (default: 20x10 chars)

[x] Add artwork cache
    - Store converted ANSI art in ~/.cache/tunez/artwork/
    - Key by artwork reference hash + width
    - TTL-based expiration (default: 30 days)

[x] Add artwork state to app Model
    - artworkANSI string
    - artworkLoading bool
    - artworkTrackID string

[x] Update renderNowPlaying() to show artwork
    - Display to the left of track info
    - Show placeholder when loading/unavailable

[x] Add config options
    - artwork.enabled (bool, default: true)
    - artwork.width (int, default: 20)
    - artwork.cache_days (int, default: 30)

[x] Add tests
    - Image conversion
    - Cache hit/miss
    - Placeholder generation
```

#### Files Modified
- `internal/artwork/artwork.go` - ANSI conversion, cache, placeholder
- `internal/artwork/artwork_test.go` - Conversion and cache tests
- `internal/app/app.go` - Artwork state, fetchArtworkCmd, renderNowPlaying
- `internal/config/config.go` - ArtworkConfig section
- `cmd/tunez/main.go` - Initialize artwork cache

---

### 2.4 Additional Themes

**Priority:** MEDIUM  
**Complexity:** Low  
**Status:** ✅ Complete

Add alternative color themes beyond the default rainbow theme.

#### Requirements
- Monochrome theme (grayscale only)
- Green terminal theme (classic green-on-black)
- Theme selection via config and runtime
- Theme preview in config screen

#### Implementation Tasks
```
[x] Define theme interface in internal/ui/theme.go
    - Already exists: ui.Theme struct with lipgloss styles

[x] Add MonochromeTheme() function
    - All text in grayscale (white, gray, dark gray)
    - Borders in medium gray
    - Highlights via bold/underline instead of color

[x] Add GreenTerminalTheme() function
    - Primary: bright green (#00FF00)
    - Secondary: dark green (#008000)
    - Background: black
    - Classic terminal aesthetic

[x] Add NoColorTheme() function
    - Plain text only, no ANSI colors
    - Supports NO_COLOR environment variable

[x] Add theme registry
    - Map theme names to constructors
    - GetTheme(), ValidTheme(), ThemeNames() functions
    - "rainbow" (default), "mono", "green", "nocolor"

[x] Update config loading
    - Read ui.theme from config
    - Validate theme name exists
    - Fall back to rainbow if invalid

[x] Add tests
    - Theme loading
    - NO_COLOR override
    - All themes render without panic
```

#### Files Modified
- `internal/ui/theme.go` - Added Monochrome, GreenTerminal, NoColor themes + registry
- `internal/ui/theme_test.go` - Tests for all themes
- `cmd/tunez/main.go` - Use GetTheme() for theme selection

---

### 2.5 Offline Cache / Download

**Priority:** LOW (v1)  
**Complexity:** High

Cache streamed tracks for offline playback.

#### Requirements
- Download tracks to local cache (provider-gated)
- Configurable cache size and eviction policy
- Offline mode toggle
- Cache management UI in Config Screen 9

#### Implementation Tasks
```
[ ] Create internal/cache package
    - SQLite metadata: track_id, provider_id, file_path, size, last_accessed
    - File storage: ~/.config/tunez/cache/tracks/{hash}.audio

[ ] Implement download manager
    - Background download queue
    - Progress reporting
    - Resume interrupted downloads

[ ] Add cache eviction
    - LRU (least recently used)
    - Size-based limit (configurable, default: 10GB)
    - TTL expiration (configurable, default: 30 days)

[ ] Update GetStream to check cache first
    - If cached, return file:// URL
    - If not cached, return remote URL
    - Optionally trigger background cache

[ ] Implement Config Screen 9 (Cache/Offline)
    - Show cache size and item count
    - Show cache location
    - Actions: Clear cache, Set size limit
    - Offline mode toggle

[ ] Add config options
    - cache.enabled (bool, default: false)
    - cache.max_size_gb (int, default: 10)
    - cache.ttl_days (int, default: 30)
    - cache.location (string, default: auto)

[ ] Add tests
    - Cache hit/miss
    - Eviction logic
    - Offline mode fallback
```

#### Files to Modify
- `internal/cache/cache.go` - New package
- `internal/cache/download.go` - Download manager
- `internal/app/app.go` - Cache screen rendering
- `internal/config/config.go` - Cache config

---

### 2.6 Scrobbling

**Priority:** LOW (v1)  
**Complexity:** Medium  
**Status:** ✅ Complete

Report played tracks to Last.fm or similar services.

#### Requirements
- Scrobble to Last.fm API
- Scrobble to Melodee API (native scrobbling)
- Configurable enable/disable (master switch + per-scrobbler)
- Handle offline scrobble queue

#### Implementation Tasks
```
[x] Create internal/scrobble package
    - Scrobbler interface for multiple backends
    - Manager for fan-out to multiple scrobblers
    - Offline queue with persistence

[x] Implement Last.fm scrobbler
    - Last.fm API client with MD5 signing
    - OAuth authentication support
    - Scrobble and NowPlaying methods

[x] Implement Melodee scrobbler
    - POST /api/v1/scrobble endpoint
    - Reuses auth token from Melodee provider
    - ScrobbleType: NowPlaying, Scrobble

[x] Implement scrobble triggers
    - Scrobble after 50% of track played OR 4 minutes
    - "Now Playing" notification at track start

[x] Add scrobble state to app
    - scrobbled bool (per-track flag)
    - Scrobble manager in Model

[x] Add config options
    - scrobble.enabled (bool, default: false) - master switch
    - [[scrobblers]] array with id, type, enabled, settings

[x] Add offline queue persistence
    - Save pending scrobbles on shutdown
    - Load pending scrobbles on startup
    - Flush when connection restored

[x] Add tests
    - Scrobble timing logic (50%/4min)
    - Pending queue management
    - Manager fan-out
```

#### Files Modified
- `internal/scrobble/scrobble.go` - Track type, errors
- `internal/scrobble/scrobbler.go` - Scrobbler interface, Manager
- `internal/scrobble/lastfm/lastfm.go` - Last.fm implementation
- `internal/scrobble/lastfm/lastfm_test.go` - Last.fm tests
- `internal/scrobble/melodee/melodee.go` - Melodee implementation
- `internal/scrobble/scrobble_test.go` - Manager tests
- `internal/app/app.go` - Scrobble integration with playback events
- `internal/config/config.go` - ScrobbleConfig, ScrobblerEntry
- `cmd/tunez/main.go` - Scrobble manager initialization

---

### Phase 2 Acceptance Criteria

Phase 2 is complete when:
1. [x] Queue persists across restarts
2. [x] Lyrics display works for providers with CapLyrics
3. [x] Artwork displays in Now Playing (optional based on config)
4. [x] At least 2 additional themes available (mono, green, nocolor)
5. [ ] Cache system works for offline playback (deferred to v1.1)
6. [x] Scrobbling works with Last.fm and Melodee
7. [x] All Phase 2 tests pass

---

## Phase 3: v2 Features

**Goal:** Advanced UX, CLI workflows, and polish.

### 3.1 Command Palette

**Priority:** HIGH  
**Complexity:** Medium

Quick command access via fuzzy search.

#### Requirements
- Open with `:` or `Ctrl+P`
- Fuzzy search across all actions
- Show keybinding hints
- Execute selected command

#### Implementation Tasks
```
[ ] Create command registry
    - Action name, description, keybinding, handler
    - Categories: Navigation, Playback, Queue, Config

[ ] Implement fuzzy matcher
    - Use github.com/sahilm/fuzzy or similar
    - Score and rank results

[ ] Create palette overlay UI
    - Input field at top
    - Scrollable results list
    - Show keybinding for each result

[ ] Wire up command execution
    - Return tea.Cmd from selected action
    - Close palette on execution

[ ] Add config option
    - keybindings.command_palette (default: ":")

[ ] Add tests
    - Fuzzy matching
    - Command execution
```

#### Files to Modify
- `internal/app/commands.go` - New file for command registry
- `internal/app/palette.go` - New file for palette UI
- `internal/app/app.go` - Integrate palette

---

### 3.2 CLI Play Flow

**Priority:** MEDIUM  
**Complexity:** Medium

Start playback from command line, then launch TUI.

#### Requirements
- `tunez play --search "query"` - Search and play first result
- `tunez play --track ID` - Play specific track
- `tunez play --album ID` - Queue album and play
- Launch TUI after queueing with Now Playing screen active

#### Implementation Tasks
```
[ ] Add play subcommand to CLI
    - Parse flags: --search, --track, --album, --artist
    - Initialize provider
    - Execute search/lookup
    - Queue results

[ ] Implement search-and-play logic
    - Search with query
    - Take first track result (or prompt if multiple)
    - Add to queue and start playback

[ ] Launch TUI after queue populated
    - Start at Now Playing screen
    - Playback already started

[ ] Add --no-tui flag for headless playback
    - Play track(s) without launching TUI
    - Exit after queue exhausted

[ ] Add tests
    - Flag parsing
    - Search integration
    - Queue initialization
```

#### Files to Modify
- `cmd/tunez/main.go` - Add play subcommand
- `cmd/tunez/play.go` - New file for play command
- `internal/app/app.go` - Support pre-populated queue

---

### 3.3 Advanced Diagnostics

**Priority:** LOW  
**Complexity:** Low

Debug overlay for troubleshooting.

#### Requirements
- Toggle with `Ctrl+D`
- Show provider request latency
- Show cache hit rates
- Show mpv connection status
- Show memory usage

#### Implementation Tasks
```
[ ] Create diagnostics state
    - lastRequestLatency time.Duration
    - cacheHits, cacheMisses int
    - mpvConnected bool
    - memoryUsage uint64

[ ] Implement diagnostics overlay
    - Semi-transparent overlay
    - Key metrics in corner
    - Auto-refresh every second

[ ] Add request timing
    - Wrap provider calls with timing
    - Store last N request latencies

[ ] Add mpv health monitoring
    - Periodic ping to mpv
    - Track reconnection attempts

[ ] Add config option
    - debug.enabled (bool, default: false)
    - debug.overlay_key (string, default: "ctrl+d")

[ ] Add tests
    - Timing accuracy
    - Overlay rendering
```

#### Files to Modify
- `internal/app/diagnostics.go` - New file
- `internal/app/app.go` - Integrate diagnostics
- `internal/player/player.go` - Add health check

---

### 3.4 Help Reflects Config Keybindings

**Priority:** MEDIUM  
**Complexity:** Low

Show actual configured keybindings in help overlay.

#### Requirements
- Read keybindings from config
- Display configured values in help screen
- Highlight customized bindings

#### Implementation Tasks
```
[ ] Update renderHelpOverlay() to use config values
    - Replace hard-coded strings with cfg.Keybindings.*
    - Format: "action : key"

[ ] Add indicator for customized bindings
    - Show "(custom)" next to non-default bindings

[ ] Add "Reset to defaults" option
    - Show in config screen
    - Regenerate keybindings section

[ ] Add tests
    - Help displays correct bindings
    - Custom binding display
```

#### Files to Modify
- `internal/app/app.go` - Update renderHelpOverlay()

---

### 3.5 CLI Utilities

**Priority:** LOW  
**Complexity:** Low

Helpful CLI commands for setup and troubleshooting.

#### Requirements
- `tunez version` - Show version info
- `tunez config init` - Create example config
- `tunez doctor` - Check mpv and provider connectivity

#### Implementation Tasks
```
[ ] Add version command
    - Print version, build date, Go version
    - Use ldflags for build-time injection

[ ] Add config init command
    - Write example config.toml to default location
    - Don't overwrite existing config
    - Print path on success

[ ] Add doctor command
    - Check mpv in PATH
    - Check config file exists and is valid
    - Test provider connectivity (with timeout)
    - Print summary with pass/fail for each check

[ ] Add tests
    - Version output format
    - Config init creates valid config
    - Doctor reports correctly
```

#### Files to Modify
- `cmd/tunez/main.go` - Add subcommands
- `cmd/tunez/version.go` - Version command
- `cmd/tunez/config.go` - Config init command
- `cmd/tunez/doctor.go` - Doctor command

---

### Phase 3 Acceptance Criteria

Phase 3 is complete when:
1. [ ] Command palette works with fuzzy search
2. [ ] `tunez play` commands work end-to-end
3. [ ] Diagnostics overlay shows useful metrics
4. [ ] Help overlay shows actual configured keybindings
5. [ ] CLI utilities (version, config init, doctor) work
6. [ ] All Phase 3 tests pass

---

## Testing Requirements

### Unit Tests (All Phases)
- Queue operations (add, remove, move, shuffle, repeat, persistence)
- Config validation (profiles, paths, keybindings)
- Provider error mapping
- mpv IPC encode/decode
- Lyrics parsing
- Artwork conversion
- Scrobble timing

### Integration Tests (Build-tagged)
- Filesystem provider with fixture library
- Melodee provider with mocked HTTP
- Full TUI flow with fake mpv
- Cache system with temp directory

### Provider Contract Tests
- Paging behavior
- Browse flows (artists → albums → tracks)
- Search across entity types
- GetStream returns valid URL
- Error normalization

---

## Configuration Reference

```toml
config_version = 1
active_profile = "home-files"

[ui]
page_size = 100
no_emoji = false
theme = "rainbow"  # rainbow | mono | green

[player]
mpv_path = "mpv"
ipc = "auto"
initial_volume = 70
seek_small_seconds = 5
seek_large_seconds = 30
volume_step = 5

[queue]
persist = true  # Phase 2

[artwork]
enabled = true  # Phase 2
width = 20
cache_days = 30

[cache]
enabled = false  # Phase 2
max_size_gb = 10
ttl_days = 30

[scrobble]
enabled = false  # Phase 2
service = "lastfm"

[keybindings]
play_pause = "space"
next_track = "n"
prev_track = "p"
seek_forward = "l"
seek_backward = "h"
volume_up = "+"
volume_down = "-"
mute = "m"
shuffle = "s"
repeat = "r"
search = "/"
help = "?"
quit = "ctrl+c"
command_palette = ":"  # Phase 3

[[profiles]]
id = "home-files"
name = "Home Files"
provider = "filesystem"
enabled = true

[profiles.settings]
roots = ["/music"]
index_db = "filesystem.sqlite"
scan_on_start = true

[[profiles]]
id = "melodee-home"
name = "Melodee (Home)"
provider = "melodee"
enabled = true

[profiles.settings]
base_url = "https://music.example.com"
username = "user"
password_env = "TUNEZ_MELODEE_PASSWORD"
```

---

## Related Documents

| Document | Description |
|----------|-------------|
| [TECH_DESIGN.md](TECH_DESIGN.md) | Architecture, process model, Bubble Tea strategy |
| [PROVIDERS.md](PROVIDERS.md) | Provider interface contract |
| [TUI_UX.md](TUI_UX.md) | Screen specifications and interactions |
| [CONFIG.md](CONFIG.md) | Configuration format details |
| [SECURITY_PRIVACY.md](SECURITY_PRIVACY.md) | Security requirements |
| [TEST_STRATEGY.md](TEST_STRATEGY.md) | Testing approach |
| [PROVIDER_FILESYSTEM.md](PROVIDER_FILESYSTEM.md) | Filesystem provider details |
| [PROVIDER_MELODEE_API.md](PROVIDER_MELODEE_API.md) | Melodee provider details |
