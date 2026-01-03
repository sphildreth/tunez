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
| Phase 2 (v1) | 🔲 Not Started | Lyrics, artwork, caching, themes |
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

Persist the play queue across application restarts.

#### Requirements
- Store queue in SQLite database (`~/.config/tunez/state/queue.db`)
- Restore queue on startup (paused at position 0)
- Clear queue when switching profiles
- Handle missing files gracefully (remove from queue, show toast)

#### Implementation Tasks
```
[ ] Create queue persistence schema in internal/queue/persistence.go
    - Table: queue_items (position, track_id, provider_id, track_json, added_at)
    - Table: queue_state (current_index, shuffle_enabled, repeat_mode)

[ ] Add Save() method to Queue
    - Serialize current state to SQLite
    - Called after any queue modification

[ ] Add Load() method to Queue  
    - Read from SQLite on startup
    - Validate tracks still exist (provider.GetTrack)
    - Remove invalid entries with toast notification

[ ] Update app.go Init() to restore queue
    - Load queue after provider initialization
    - Set status "Restored X tracks" or "Queue empty"

[ ] Add config option: queue.persist (bool, default: true)

[ ] Add tests for persistence
    - Save/load round-trip
    - Handle corrupted database
    - Handle missing tracks
```

#### Files to Modify
- `internal/queue/queue.go` - Add persistence methods
- `internal/queue/persistence.go` - New file for SQLite operations
- `internal/app/app.go` - Load queue on init
- `internal/config/config.go` - Add queue config section

---

### 2.2 Lyrics Display (Functional)

**Priority:** HIGH  
**Complexity:** Medium

Display lyrics for the currently playing track.

#### Requirements
- Fetch lyrics from provider (capability-gated: CapLyrics)
- Display in Screen 6 with scrolling
- Handle loading/error/no-lyrics states
- Sync scroll position with playback (basic: paragraph-level)

#### Implementation Tasks
```
[ ] Add GetLyrics to provider interface (already defined)
    - Filesystem: read from embedded tags or .lrc sidecar files
    - Melodee: GET /api/v1/songs/{id} returns lyrics field

[ ] Implement GetLyrics for Filesystem provider
    - Check ID3v2 USLT frame for embedded lyrics
    - Check for {filename}.lrc sidecar file
    - Return ErrNotSupported if neither found

[ ] Implement GetLyrics for Melodee provider
    - Parse lyrics from Song response
    - Handle plain text and timestamped formats

[ ] Create lyrics state in app Model
    - lyrics string
    - lyricsLoading bool
    - lyricsError error
    - lyricsScrollOffset int

[ ] Add lyricsCmd to fetch lyrics when track changes
    - Triggered by playTrackMsg
    - Cancel previous fetch on track change

[ ] Update renderLyrics() for functional display
    - Show loading spinner during fetch
    - Show "No lyrics available" when empty
    - Show lyrics text with scroll support
    - Handle j/k for scroll, g/G for top/bottom

[ ] Add lyrics loading indicator in Now Playing screen
    - Show "Lyrics: Loading..." / "Available" / "None"

[ ] Add tests for lyrics parsing
    - Embedded lyrics extraction
    - LRC file parsing
    - Timestamped lyrics parsing
```

#### Files to Modify
- `internal/providers/filesystem/provider.go` - Implement GetLyrics
- `internal/providers/melodee/provider.go` - Implement GetLyrics
- `internal/app/app.go` - Add lyrics state and rendering
- `internal/provider/provider.go` - Lyrics type definition

---

### 2.3 Artwork Display

**Priority:** MEDIUM  
**Complexity:** High

Display album artwork in the TUI.

#### Requirements
- Fetch artwork from provider (capability-gated: CapArtwork)
- Convert to ANSI art for terminal display
- Show in Now Playing screen
- Cache artwork locally

#### Implementation Tasks
```
[ ] Add GetArtwork to provider implementations
    - Filesystem: extract from audio files or folder.jpg
    - Melodee: use thumbnailUrl from Album response

[ ] Create internal/artwork package
    - Image download with caching
    - Image-to-ANSI conversion (use github.com/qeesung/image2ascii or similar)
    - Configurable size (default: 20x10 chars)

[ ] Add artwork cache
    - Store converted ANSI art in ~/.config/tunez/cache/artwork/
    - Key by album ID hash
    - TTL-based expiration (default: 30 days)

[ ] Add artwork state to app Model
    - artworkANSI string
    - artworkLoading bool

[ ] Update renderNowPlaying() to show artwork
    - Display to the left of track info
    - Show placeholder when loading/unavailable

[ ] Add config options
    - artwork.enabled (bool, default: true)
    - artwork.width (int, default: 20)
    - artwork.cache_days (int, default: 30)

[ ] Add tests
    - Image conversion
    - Cache hit/miss
    - Missing artwork handling
```

#### Files to Modify
- `internal/artwork/artwork.go` - New package
- `internal/artwork/cache.go` - Artwork caching
- `internal/providers/*/provider.go` - Implement GetArtwork
- `internal/app/app.go` - Add artwork rendering
- `internal/config/config.go` - Add artwork config

---

### 2.4 Additional Themes

**Priority:** MEDIUM  
**Complexity:** Low

Add alternative color themes beyond the default rainbow theme.

#### Requirements
- Monochrome theme (grayscale only)
- Green terminal theme (classic green-on-black)
- Theme selection via config and runtime
- Theme preview in config screen

#### Implementation Tasks
```
[ ] Define theme interface in internal/ui/theme.go
    - Already exists: ui.Theme struct with lipgloss styles

[ ] Add MonochromeTheme() function
    - All text in grayscale (white, gray, dark gray)
    - Borders in medium gray
    - Highlights via bold/underline instead of color

[ ] Add GreenTerminalTheme() function
    - Primary: bright green (#00FF00)
    - Secondary: dark green (#008000)
    - Background: black
    - Classic terminal aesthetic

[ ] Add theme registry
    - Map theme names to constructors
    - "rainbow" (default), "mono", "green", "nocolor"

[ ] Update config loading
    - Read ui.theme from config
    - Validate theme name exists
    - Fall back to rainbow if invalid

[ ] Add theme preview in Config screen
    - Show sample text in each theme
    - Allow cycling through themes with 't' key

[ ] Add runtime theme switching
    - Store selected theme in state
    - Re-render on change

[ ] Add tests
    - Theme loading
    - NO_COLOR override
```

#### Files to Modify
- `internal/ui/theme.go` - Add new themes
- `internal/config/config.go` - Theme validation
- `internal/app/app.go` - Theme switching
- `cmd/tunez/main.go` - Theme initialization

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

Report played tracks to Last.fm or similar services.

#### Requirements
- Scrobble to Last.fm API
- Configurable enable/disable
- Show scrobble status in top bar
- Handle offline scrobble queue

#### Implementation Tasks
```
[ ] Create internal/scrobble package
    - Last.fm API client
    - OAuth authentication flow
    - Scrobble queue for offline

[ ] Implement scrobble triggers
    - Scrobble after 50% of track played OR 4 minutes
    - "Now Playing" notification at track start

[ ] Add scrobble state to app
    - scrobbleEnabled bool
    - lastScrobbleStatus string

[ ] Update top bar to show scrobble status
    - "Scrobble: ON" / "OFF" / "Pending (3)"

[ ] Add config options
    - scrobble.enabled (bool, default: false)
    - scrobble.service (string: "lastfm")
    - scrobble.api_key_env (string)
    - scrobble.session_key_env (string)

[ ] Add authentication flow
    - `tunez scrobble auth` command
    - Open browser for OAuth
    - Store session key securely

[ ] Add tests
    - Scrobble timing logic
    - Offline queue
    - API mocking
```

#### Files to Modify
- `internal/scrobble/scrobble.go` - New package
- `internal/scrobble/lastfm.go` - Last.fm client
- `internal/app/app.go` - Scrobble integration
- `cmd/tunez/main.go` - Scrobble command

---

### Phase 2 Acceptance Criteria

Phase 2 is complete when:
1. [ ] Queue persists across restarts
2. [ ] Lyrics display works for providers with CapLyrics
3. [ ] Artwork displays in Now Playing (optional based on terminal support)
4. [ ] At least 2 additional themes available (mono, green)
5. [ ] Cache system works for offline playback (provider-gated)
6. [ ] Scrobbling works with Last.fm
7. [ ] All Phase 2 tests pass

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
