# Tunez — Comprehensive Phase Plan

**Last updated:** 2026-01-03  
**Status:** Draft

This document organizes ALL requirements from the PRD, TUI_UX, and related documents into clear, deliverable phases. Each phase builds upon the previous one, and when all phases are complete, the application will be code-complete.

## Phase 0: Foundation (COMPLETE)

**Status:** ✅ Complete - All core infrastructure is in place

### Completed Items:
- ✅ Go module structure
- ✅ Config loading and validation (TOML)
- ✅ Logging setup (structured, file-based)
- ✅ Provider interface contract
- ✅ Basic Bubble Tea app scaffold
- ✅ Player controller with mpv IPC
- ✅ Queue implementation
- ✅ Theme system (rainbow + NO_COLOR)
- ✅ Filesystem provider (basic browse/search/play)
- ✅ Melodee provider (basic auth/browse/search/play)
- ✅ Basic TUI screens (Loading, Library, Search, Queue, Config)
- ✅ Basic keybindings

### Files in place:
- `src/cmd/tunez/main.go`
- `src/internal/app/app.go` (basic implementation)
- `src/internal/config/config.go`
- `src/internal/logging/logging.go`
- `src/internal/player/player.go`
- `src/internal/provider/provider.go` + `errors.go`
- `src/internal/providers/filesystem/provider.go`
- `src/internal/providers/melodee/provider.go`
- `src/internal/queue/queue.go`
- `src/internal/ui/theme.go`

---

## Phase 1: MVP Core Features

**Goal:** Fully functional terminal music player with all essential features from PRD

### 1.1 Enhanced TUI Screens & Navigation

**Requirements from:** `docs/TUI_UX.md`

#### Screen 1 - Main / Now Playing (Enhanced)
- [ ] Track details (title, artist, album)
- [ ] Codec/bitrate display (when known)
- [ ] Large progress bar with elapsed/remaining
- [ ] Up Next preview (next 3-10 items from queue)
- [ ] Visualizer placeholder (adaptive FPS)
- [ ] Proper layout with left nav and bottom bar

#### Screen 2 - Search (Enhanced)
- [ ] Global search across tracks, albums, artists
- [ ] Inline search input or modal
- [ ] Results grouped by type (Tracks/Albums/Artists)
- [ ] Tab cycling between result types
- [ ] Selection behavior:
  - Track: play now or enqueue based on config
  - Album/Artist: jump to library view
- [ ] Paging/incremental loading with "Loading more..." row

#### Screen 3 - Library (Enhanced)
- [ ] Artists list with paging
- [ ] Albums list (filtered by artist when selected)
- [ ] Tracks list (filtered by album/artist)
- [ ] Tab cycling between Artists/Albums/Tracks
- [ ] Selection behavior:
  - Artist → show albums
  - Album → show tracks
  - Track → play/enqueue
- [ ] Details panel showing selected item info
- [ ] Actions: [Enter]Open, [p]Play, [A]Add Album, [S]Shuffle Album
- [ ] Infinite scroll/paging for large libraries

#### Screen 4 - Queue (Enhanced)
- [ ] Full queue display with numbering
- [ ] Current playing item visually marked
- [ ] Actions:
  - `enter`: jump+play selected
  - `x`: remove selected
  - `C`: clear queue
  - `u/d`: move up/down
  - `P`: play next
- [ ] Queue mode display (Normal/Shuffle/Repeat)
- [ ] Paging for large queues

#### Screen 5 - Playlists (Capability-gated)
- [ ] Only visible if provider supports CapPlaylists
- [ ] Playlists list with track counts
- [ ] Playlist detail view (tracks)
- [ ] Actions:
  - `enter`: open playlist
  - `A`: add all to queue
  - `p`: play playlist
  - `I`: info
- [ ] Placeholder UI when unsupported

#### Screen 6 - Lyrics (Capability-gated)
- [ ] Only visible if provider supports CapLyrics
- [ ] Loading state
- [ ] "No lyrics available" state
- [ ] Lyrics text display (scrollable)
- [ ] Controls: `j/k` scroll, `g/G` top/bottom
- [ ] Placeholder/disabled in MVP

#### Screen 7 - Configuration (Main)
- [ ] Read-only summary of current config
- [ ] Sections navigation:
  - Providers & Profiles
  - Theme & ANSI
  - Keybindings
  - Cache/Offline
  - Scrobbling
  - Logging & Diagnostics
- [ ] Details panel for selected section
- [ ] Actions: `enter` to open section

#### Screen 8 - Config: Providers & Profiles
- [ ] Providers list with capabilities
- [ ] Profiles list
- [ ] Profile detail summary (secrets redacted)
- [ ] Actions:
  - `enter`: set active profile (re-init with spinner)
  - `o`: open config file path
  - `r`: retry provider initialization
- [ ] Show provider health status

#### Screen 9 - Config: Cache/Offline (View-only MVP)
- [ ] Cache DB path
- [ ] Cache size estimate
- [ ] Last refresh time
- [ ] "Unsupported" messaging for provider-gated features
- [ ] Placeholder for v1 functional controls

#### Screen 10 - Help / Keybindings Overlay
- [ ] Toggle with `?`
- [ ] Global keybindings
- [ ] Screen-specific keybindings
- [ ] Reflect actual keybindings from config
- [ ] Close with `esc` or `q`

#### Screen 11 - Error Handling
- [ ] **Toast** (non-blocking, auto-dismiss):
  - Transient errors (retrying, reconnecting)
  - Status line area
- [ ] **Modal** (blocking, requires dismissal):
  - Fatal errors (mpv not found, unauthorized, invalid config)
  - Actions: Retry, Open config path, Exit
- [ ] Error messages with context

#### Screen 12 - CLI "Play then Launch TUI" (Optional MVP)
- [ ] `tunez play --track <id>`
- [ ] `tunez play --search "name"`
- [ ] Resolve track(s), start playback, launch TUI
- [ ] Placeholder if not implemented in MVP

### 1.2 Playback Control & Queue Management

**Requirements from:** `docs/PRD.md` section 4.4, 4.5

#### Playback Controls
- [ ] Play/pause toggle (`space`)
- [ ] Next/previous track (`n`/`p`)
- [ ] Seek forward/back:
  - Small: `h`/`l` (default 5s, configurable)
  - Large: `H`/`L` (default 30s, configurable)
- [ ] Volume up/down (`-`/`+`)
- [ ] Mute (`m`)
- [ ] Shuffle toggle (`s`)
- [ ] Repeat cycle (`r`): off → all → one
- [ ] Volume bounds: 0-100%, no boost beyond 100%
- [ ] Progress display: elapsed/remaining time + bar

#### Queue Management
- [ ] Add track(s) to queue
- [ ] Play next (`P`)
- [ ] Add to queue (`A`)
- [ ] Remove from queue (`x`)
- [ ] Clear queue (`C`)
- [ ] Jump to queue item (`enter`)
- [ ] Move items up/down (`u`/`d`)
- [ ] Queue persistence (optional MVP, required v1)

#### Playback State Display
- [ ] Now Playing in bottom bar
- [ ] Play state icon (⏵/⏸)
- [ ] Progress bar
- [ ] Volume display
- [ ] Shuffle/Repeat mode display

### 1.3 Configuration & Profiles

**Requirements from:** `docs/PRD.md` section 4.7, `docs/CONFIG.md`

#### Config File
- [ ] TOML format with validation
- [ ] Multiple provider profiles
- [ ] Keybindings customization
- [ ] Playback settings (mpv path, cache, volume, seek amounts)
- [ ] UI settings (theme, page size, no_emoji)
- [ ] Profile selection at runtime

#### Validation
- [ ] Active profile exists and enabled
- [ ] mpv discoverable (PATH or config path)
- [ ] Filesystem roots exist
- [ ] Melodee base_url valid
- [ ] Actionable error messages without leaking secrets

#### Profile Switching
- [ ] Runtime profile selection menu
- [ ] Stop playback on switch
- [ ] Clear queue on switch
- [ ] Re-initialize provider with spinner

### 1.4 Theme & Accessibility

**Requirements from:** `docs/PRD.md` section 4.7, `docs/TUI_UX.md`

#### Theme System
- [ ] Default: very colorful, rainbow-like ANSI effects
- [ ] NO_COLOR environment variable support
- [ ] High-contrast fallback when NO_COLOR=1
- [ ] Theme selection in config (`ui.theme`)
- [ ] Color not the only carrier of meaning

#### Accessibility
- [ ] Graceful degradation at 80×24
- [ ] Symbols + text labels (not color-only)
- [ ] `ui.no_emoji = true` support
- [ ] Clear error messages

### 1.5 Provider Architecture & Capabilities

**Requirements from:** `docs/PROVIDERS.md`, `docs/PRD.md` section 4.1

#### Provider Interface
- [ ] Stable interface with capability detection
- [ ] Incremental paging for large datasets
- [ ] Normalized errors (NotSupported, NotFound, Unauthorized, etc.)
- [ ] Clear capability exposure

#### Built-in Providers (Phase 1)
- [ ] **Filesystem Provider:**
  - [ ] Tag-based indexing (SQLite)
  - [ ] Folder fallback for untagged files
  - [ ] Fast startup with persisted index
  - [ ] Browse artists/albums/tracks
  - [ ] Search across all fields
  - [ ] `file://` stream URLs
  - [ ] Paging support

- [ ] **Melodee API Provider:**
  - [ ] HTTPS authentication
  - [ ] Token refresh
  - [ ] Browse artists/albums/playlists
  - [ ] Search across library
  - [ ] Stream URLs with headers
  - [ ] Paging support
  - [ ] Capability detection (playlists, lyrics, artwork)

#### Capability Gating
- [ ] Playlists screen only visible if CapPlaylists
- [ ] Lyrics screen only visible if CapLyrics
- [ ] Artwork placeholders when unsupported
- [ ] Cache/Offline controls show "unsupported" when provider doesn't support
- [ ] Clear messaging for disabled features

### 1.6 Performance & Responsiveness

**Requirements from:** `docs/PRD.md` section 5.1, `docs/TECH_DESIGN.md`

#### UI Responsiveness
- [ ] No blocking in Bubble Tea Update loop
- [ ] All I/O returns via tea.Msg
- [ ] Loading states for async operations
- [ ] Cancelable requests when navigating away
- [ ] Adaptive UI tick rate (16-100ms)

#### Provider Performance
- [ ] Asynchronous provider fetches
- [ ] First page load < 1s on healthy LAN
- [ ] Graceful degradation on slow networks
- [ ] Paging/infinite scroll for large lists
- [ ] "Loading more..." indicators

#### Player Performance
- [ ] Gapless playback where supported
- [ ] Preloading next track
- [ ] Immediate UI feedback for playback commands

### 1.7 Error Handling & Observability

**Requirements from:** `docs/PRD.md` sections 4.8, 5.2

#### Error Handling
- [ ] Structured logging to file
- [ ] Log level control
- [ ] Status line for transient errors
- [ ] Modal for fatal errors
- [ ] Retry with backoff for network failures
- [ ] User-visible status for provider issues
- [ ] mpv crash handling with remediation

#### Observability
- [ ] Structured logging (slog)
- [ ] Log file location: `~/.config/tunez/state/tunez-YYYYMMDD.log`
- [ ] Status line for transient errors
- [ ] Debug overlay (optional): last request latency, cache hits

### 1.8 Security & Privacy

**Requirements from:** `docs/SECURITY_PRIVACY.md`, `docs/PRD.md` section 5.3

- [ ] Secrets never written to logs
- [ ] Config tokens supported
- [ ] HTTPS by default for remote providers
- [ ] OS keychain support (v1+)
- [ ] No telemetry by default

### 1.9 Testing

**Requirements from:** `docs/TEST_STRATEGY.md`

#### Unit Tests
- [ ] Queue operations (existing + enhancements)
- [ ] Keybinding parsing/dispatch
- [ ] mpv IPC encode/decode
- [ ] Config validation
- [ ] Provider error mapping

#### Provider Contract Tests
- [ ] Paging behavior
- [ ] Browse flows
- [ ] Search sanity
- [ ] GetStream returns usable URL
- [ ] Error normalization

#### Integration Tests (build-tagged)
- [ ] Filesystem provider with fixtures
- [ ] Melodee provider with mocked HTTP

#### Fake mpv Server
- [ ] Accept commands
- [ ] Emit property-change events
- [ ] Emit end-file events
- [ ] Deterministic player tests

### 1.10 Documentation

- [ ] Update all docs to reflect implementation
- [ ] Document any MVP trade-offs in DECISIONS.md
- [ ] README with setup instructions
- [ ] Example config file

---

## Phase 2: v1 Features

**Goal:** Enhanced features for power users

### 2.1 Offline Cache & Download

**Requirements from:** `docs/PRD.md` section 7

- [ ] Functional cache controls in Config Screen 9
- [ ] Download tracks for offline use (provider-gated)
- [ ] Cache management (clear, rebuild)
- [ ] Offline mode toggles
- [ ] Eviction policies (LRU, TTL)
- [ ] Cache size limits

### 2.2 Lyrics (Functional)

**Requirements from:** `docs/PRD.md` section 7, `docs/TUI_UX.md` Screen 6

- [ ] Lyrics fetch from provider
- [ ] Real-time display for current track
- [ ] Follow-along with playback
- [ ] Scroll sync
- [ ] Loading states
- [ ] No lyrics available state

### 2.3 Artwork

**Requirements from:** `docs/PRD.md` section 7

- [ ] Artwork fetch from provider
- [ ] Image-to-ANSI conversion (optional)
- [ ] Inline placeholders
- [ ] Artwork in Now Playing screen
- [ ] Capability gating

### 2.4 Scrobbling

**Requirements from:** `docs/PRD.md` section 7

- [ ] Scrobble to Last.fm or similar
- [ ] Config toggle
- [ ] Status indicator in top bar
- [ ] Provider-gated

### 2.5 Queue Persistence

**Requirements from:** `docs/PRD.md` section 4.5

- [ ] Persist queue across restarts
- [ ] Store in SQLite
- [ ] Restore on startup
- [ ] Clear on profile switch

### 2.6 Additional Themes

**Requirements from:** `docs/PRD.md` section 7

- [ ] Monochrome theme
- [ ] Green terminal theme
- [ ] Theme selection UI
- [ ] Theme preview

### 2.7 Advanced Library Features

**Requirements from:** `docs/PRD.md` section 7

- [ ] Sorting by genre/year/rating
- [ ] Filtering options
- [ ] Advanced search operators
- [ ] Smart playlists

---

## Phase 3: v2 Features

**Goal:** Advanced UX and polish

### 3.1 Command Palette

**Requirements from:** `docs/PRD.md` section 7

- [ ] Global command search
- [ ] Quick actions
- [ ] Keyboard shortcuts
- [ ] Fuzzy matching

### 3.2 Advanced Diagnostics

**Requirements from:** `docs/PRD.md` section 7

- [ ] Debug overlay with metrics
- [ ] Request latency tracking
- [ ] Cache hit rates
- [ ] mpv health monitoring
- [ ] Provider health details

### 3.3 CLI "Play then Launch TUI"

**Requirements from:** `docs/TUI_UX.md` Screen 12

- [ ] Full implementation if not in MVP
- [ ] `tunez play --artist ... --album ...`
- [ ] `tunez play --search "..."`
- [ ] Queue initialization
- [ ] Direct to Now Playing

### 3.4 Advanced UX

- [ ] Lyrics follow-along with timing
- [ ] Visualizer enhancements
- [ ] Custom keybinding profiles
- [ ] Plugin architecture (if needed)

---

## Code Completion Criteria

The application will be considered **code-complete** when:

### Phase 1 Complete:
- [ ] All screens from TUI_UX.md are implemented
- [ ] All playback controls work
- [ ] Queue management is fully functional
- [ ] Both providers work end-to-end
- [ ] Config system is complete with validation
- [ ] Theme system works with NO_COLOR
- [ ] Error handling covers all cases
- [ ] All tests pass
- [ ] Documentation is updated

### Phase 2 Complete:
- [ ] Offline cache works
- [ ] Lyrics functional
- [ ] Artwork support
- [ ] Scrobbling works
- [ ] Queue persists
- [ ] Additional themes available

### Phase 3 Complete:
- [ ] Command palette implemented
- [ ] Advanced diagnostics available
- [ ] CLI flow complete
- [ ] All advanced UX features

---

## Implementation Order Recommendation

To maximize velocity and minimize risk:

1. **Phase 1.1** - Enhanced TUI Screens (build on existing scaffold)
2. **Phase 1.2** - Playback & Queue (enhance existing)
3. **Phase 1.3** - Config & Profiles (enhance existing)
4. **Phase 1.4** - Theme & Accessibility (enhance existing)
5. **Phase 1.5** - Provider Capabilities (enhance existing)
6. **Phase 1.6** - Performance (optimize existing)
7. **Phase 1.7** - Error Handling (enhance existing)
8. **Phase 1.8** - Security (validate existing)
9. **Phase 1.9** - Testing (comprehensive test suite)
10. **Phase 1.10** - Documentation

Then proceed to Phase 2 and 3 features.

---

## Notes

- **Capability Gating:** All features that depend on provider capabilities must gracefully handle unsupported providers with clear messaging.
- **MVP Trade-offs:** Any feature marked "optional" or "view-only" in MVP must be clearly documented in DECISIONS.md.
- **Performance:** All I/O must be async; no blocking in UI loop.
- **Testing:** Every new feature must include tests.
- **Documentation:** Update docs as implementation progresses.