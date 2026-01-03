# Phase 1 Implementation Prompt for Coding Agent

**Copy and paste this entire prompt into a coding agent to implement Phase 1 (MVP) of Tunez.**

---

## Mission

Implement **Phase 1 (MVP)** of Tunez - a fast, responsive terminal music player in Go with Bubble Tea TUI and mpv playback.

**Goal:** When complete, Tunez will be a fully functional terminal music player with all essential features from the PRD.

**Reference:** All requirements are mapped in `/home/steven/source/tunez/docs/PHASE_PLAN.md` Phase 1 section.

---

## Current State (Phase 0 Complete)

**Already implemented:**
- ✅ Basic app scaffold (main.go, app.go, config, logging)
- ✅ Player controller with mpv IPC
- ✅ Queue implementation
- ✅ Provider interface contract
- ✅ Filesystem provider (basic browse/search/play)
- ✅ Melodee provider (basic auth/browse/search/play)
- ✅ Theme system (rainbow + NO_COLOR)
- ✅ Basic TUI screens (Loading, Library, Search, Queue, Config)
- ✅ Basic keybindings

**What needs to be built:** Everything in Phase 1 checklist below.

---

## Phase 1 Implementation Checklist

### 1. Enhanced TUI Screens (All 12 Screens)

#### Screen 1 - Main / Now Playing
- [ ] Display track details: title, artist, album
- [ ] Show codec/bitrate when known
- [ ] Large progress bar with elapsed/remaining time
- [ ] Up Next preview (next 3-10 items from queue)
- [ ] Visualizer placeholder (adaptive FPS)
- [ ] Proper layout with left nav and bottom bar

#### Screen 2 - Search (Enhanced)
- [ ] Global search input (`/` key)
- [ ] Results grouped by type: Tracks, Albums, Artists
- [ ] Tab cycling between result types
- [ ] Selection behavior:
  - Track: play now or enqueue (based on config)
  - Album/Artist: jump to library view
- [ ] Paging with "Loading more..." indicator

#### Screen 3 - Library (Enhanced)
- [ ] Artists list with paging
- [ ] Albums list (filtered by selected artist)
- [ ] Tracks list (filtered by album/artist)
- [ ] Tab cycling between Artists/Albums/Tracks
- [ ] Details panel for selected item
- [ ] Actions: [Enter]Open, [p]Play, [A]Add Album, [S]Shuffle Album
- [ ] Infinite scroll/paging

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

#### Screen 5 - Playlists (Capability-gated)
- [ ] Only visible if provider supports CapPlaylists
- [ ] Playlists list with track counts
- [ ] Playlist detail view (tracks)
- [ ] Actions: `enter` open, `A` add all, `p` play playlist
- [ ] Placeholder when unsupported

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
- [ ] `enter` to open section

#### Screen 8 - Config: Providers & Profiles
- [ ] Providers list with capabilities
- [ ] Profiles list
- [ ] Profile detail summary (secrets redacted)
- [ ] Actions:
  - `enter`: set active profile (re-init with spinner)
  - `o`: open config file path
  - `r`: retry provider initialization
- [ ] Show provider health status

#### Screen 9 - Config: Cache/Offline (View-only)
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

#### Screen 12 - CLI "Play then Launch TUI" (Optional)
- [ ] `tunez play --track <id>`
- [ ] `tunez play --search "name"`
- [ ] Resolve track(s), start playback, launch TUI
- [ ] Placeholder if not implemented

### 2. Playback Control & Queue Management

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

### 3. Configuration & Profiles

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

### 4. Theme & Accessibility

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

### 5. Provider Architecture & Capabilities

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

### 6. Performance & Responsiveness

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

### 7. Error Handling & Observability

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

### 8. Security & Privacy

- [ ] Secrets never written to logs
- [ ] Config tokens supported
- [ ] HTTPS by default for remote providers
- [ ] OS keychain support (v1+)
- [ ] No telemetry by default

### 9. Testing

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

---

## Implementation Guidelines

### Architecture Rules
1. **No blocking in Update loop** - All I/O in goroutines, return tea.Msg
2. **Context everywhere** - Use context.Context for cancellation
3. **Error wrapping** - Use `fmt.Errorf("...: %w", err)`
4. **Stable sorts** - Deterministic ordering
5. **Explicit timeouts** - Don't hang indefinitely

### Code Quality
- [ ] Run `go test ./...` after each feature
- [ ] Run `go fmt ./...` before committing
- [ ] No unused exports or dead code
- [ ] Complete, compiling Go code (no pseudo-code)

### Provider Implementation
- [ ] Implement full interface from `docs/PROVIDERS.md`
- [ ] Return normalized errors
- [ ] Support cursor-based paging
- [ ] Clear capability detection

### TUI Implementation
- [ ] Follow layouts from `docs/TUI_UX.md`
- [ ] Use Lip Gloss for styling
- [ ] Support `NO_COLOR` and `ui.no_emoji`
- [ ] Graceful degradation at 80×24

### Player Implementation
- [ ] Use mpv JSON IPC
- [ ] Unix sockets (Linux/macOS) or named pipes (Windows)
- [ ] Subscribe to property changes
- [ ] Handle headers for remote streams

---

## File Structure

```
/home/steven/source/tunez/
├── docs/
│   ├── PHASE_PLAN.md          ← Reference for all requirements
│   ├── PRD.md                 ← Requirements
│   ├── TUI_UX.md              ← Screen specs
│   ├── PROVIDERS.md           ← Interface contract
│   └── CONFIG.md              ← Config format
└── src/
    ├── cmd/tunez/main.go      ← Entry point
    ├── internal/
    │   ├── app/               ← Bubble Tea model (enhance existing)
    │   ├── config/            ← Config (enhance existing)
    │   ├── logging/           ← Logging (enhance existing)
    │   ├── player/            ← mpv controller (enhance existing)
    │   ├── provider/          ← Interface (enhance existing)
    │   ├── providers/
    │   │   ├── filesystem/    ← Filesystem provider (enhance existing)
    │   │   └── melodee/       ← Melodee provider (enhance existing)
    │   ├── queue/             ← Queue (enhance existing)
    │   └── ui/                ← Theme (enhance existing)
    └── test/
        └── fixtures/          ← Test data
```

---

## Success Criteria

When Phase 1 is complete:
- [ ] All 12 TUI screens are implemented
- [ ] All playback controls work
- [ ] Queue management is fully functional
- [ ] Both providers work end-to-end
- [ ] Config system is complete with validation
- [ ] Theme system works with NO_COLOR
- [ ] Error handling covers all cases
- [ ] All tests pass
- [ ] Documentation is updated

---

## Commands to Run

```bash
# From /home/steven/source/tunez/src
go test ./...                    # Run all tests
go fmt ./...                     # Format code
go build -o ./bin/tunez ./cmd/tunez  # Build
./bin/tunez -doctor              # Validate setup
./bin/tunez                      # Run
```

---

## Questions to Ask

If you encounter:
1. **Ambiguous requirements** → Check `docs/PHASE_PLAN.md` Phase 1 section
2. **Provider interface questions** → Check `docs/PROVIDERS.md`
3. **TUI layout questions** → Check `docs/TUI_UX.md`
4. **Config format questions** → Check `docs/CONFIG.md`
5. **Architecture questions** → Check `docs/TECH_DESIGN.md`

---

**Ready to implement? Start with the first unchecked item in the checklist above and work through systematically.**