# Phase 1 Code Review Prompt for Coding Agent

**Copy and paste this entire prompt into a coding agent to review Phase 1 implementation of Tunez.**

---

## Mission

Perform a comprehensive code review of Phase 1 (MVP) implementation for Tunez - a terminal music player in Go with Bubble Tea TUI and mpv playback.

**Reference Documents:**
- `/home/steven/source/tunez/docs/PHASE_PLAN.md` Phase 1 section (requirements)
- `/home/steven/source/tunez/docs/PRD.md` (product requirements)
- `/home/steven/source/tunez/docs/TUI_UX.md` (screen specifications)
- `/home/steven/source/tunez/docs/PROVIDERS.md` (provider contract)
- `/home/steven/source/tunez/docs/CONFIG.md` (config format)
- `/home/steven/source/tunez/PHASE1_PROMPT.md` (implementation guidelines)

---

## Review Scope

Review ALL code in `/home/steven/source/tunez/src/` for Phase 1 completeness, correctness, and quality.

---

## Review Checklist

### 1. Architecture & Code Quality

#### General
- [ ] **No blocking in Bubble Tea Update loop** - All I/O must be async
- [ ] **Context usage** - All long-running operations use `context.Context`
- [ ] **Error handling** - All errors wrapped with `fmt.Errorf("...: %w", err)`
- [ ] **Error normalization** - Provider errors use normalized error types
- [ ] **No dead code** - All exports are used
- [ ] **Code formatting** - Run `go fmt ./...` - should be clean
- [ ] **Build succeeds** - `go build ./...` without errors
- [ ] **Tests pass** - `go test ./...` all pass

#### Concurrency
- [ ] **Goroutine leaks** - All goroutines have proper cleanup/cancellation
- [ ] **Channels** - Proper closing, no blocking sends
- [ ] **Tea.Msg pattern** - All async work returns via tea.Msg

### 2. TUI Screens (All 12 Screens)

#### Screen 0 - Splash/Loading
- [ ] Shows startup progress
- [ ] Transitions to Screen 1 on success
- [ ] Shows error modal on fatal error

#### Screen 1 - Main / Now Playing
- [ ] Displays track: title, artist, album
- [ ] Shows codec/bitrate when known
- [ ] Large progress bar with elapsed/remaining
- [ ] Up Next preview (next 3-10 from queue)
- [ ] Visualizer placeholder
- [ ] Proper layout with left nav and bottom bar

#### Screen 2 - Search
- [ ] Global search with `/` key
- [ ] Results grouped by type (Tracks/Albums/Artists)
- [ ] Tab cycling between types
- [ ] Correct selection behavior:
  - Track: play now or enqueue
  - Album/Artist: jump to library
- [ ] Paging with "Loading more..." indicator

#### Screen 3 - Library
- [ ] Artists list with paging
- [ ] Albums list (filtered by artist)
- [ ] Tracks list (filtered by album/artist)
- [ ] Tab cycling between modes
- [ ] Details panel for selected item
- [ ] Actions: [Enter]Open, [p]Play, [A]Add Album, [S]Shuffle Album
- [ ] Infinite scroll/paging

#### Screen 4 - Queue
- [ ] Full queue display with numbering
- [ ] Current playing item visually marked
- [ ] Actions work:
  - `enter`: jump+play
  - `x`: remove
  - `C`: clear
  - `u/d`: move
  - `P`: play next
- [ ] Queue mode display (Normal/Shuffle/Repeat)

#### Screen 5 - Playlists (Capability-gated)
- [ ] Only visible if `CapPlaylists` true
- [ ] Playlists list with track counts
- [ ] Playlist detail view
- [ ] Actions: `enter`, `A`, `p`, `I`
- [ ] Placeholder when unsupported

#### Screen 6 - Lyrics (Capability-gated)
- [ ] Only visible if `CapLyrics` true
- [ ] Loading state
- [ ] "No lyrics available" state
- [ ] Lyrics text (scrollable)
- [ ] Controls: `j/k`, `g/G`
- [ ] Placeholder in MVP

#### Screen 7 - Configuration (Main)
- [ ] Read-only summary
- [ ] Sections navigation
- [ ] Details panel
- [ ] `enter` to open sections

#### Screen 8 - Config: Providers & Profiles
- [ ] Providers list with capabilities
- [ ] Profiles list
- [ ] Profile detail (secrets redacted)
- [ ] Actions: `enter`, `o`, `r`
- [ ] Health status display

#### Screen 9 - Config: Cache/Offline
- [ ] View-only in MVP
- [ ] Cache DB path
- [ ] Cache size estimate
- [ ] Last refresh time
- [ ] "Unsupported" messaging

#### Screen 10 - Help Overlay
- [ ] Toggle with `?`
- [ ] Global keybindings
- [ ] Screen-specific keybindings
- [ ] Reflects actual config bindings
- [ ] Close with `esc` or `q`

#### Screen 11 - Error Handling
- [ ] **Toast** for transient errors (auto-dismiss)
- [ ] **Modal** for fatal errors (blocking)
- [ ] Actions: Retry, Open config, Exit

#### Screen 12 - CLI Flow (Optional)
- [ ] `tunez play --track <id>`
- [ ] `tunez play --search "name"`
- [ ] Resolves, plays, launches TUI

### 3. Playback Controls

- [ ] Play/pause (`space`)
- [ ] Next/prev (`n`/`p`)
- [ ] Seek small (`h`/`l`, configurable)
- [ ] Seek large (`H`/`L`, configurable)
- [ ] Volume up/down (`-`/`+`)
- [ ] Mute (`m`)
- [ ] Shuffle toggle (`s`)
- [ ] Repeat cycle (`r`): off → all → one
- [ ] Volume bounds: 0-100%
- [ ] Progress display with bar

### 4. Queue Management

- [ ] Add track(s) to queue
- [ ] Play next (`P`)
- [ ] Add to queue (`A`)
- [ ] Remove from queue (`x`)
- [ ] Clear queue (`C`)
- [ ] Jump to queue item (`enter`)
- [ ] Move items up/down (`u`/`d`)
- [ ] Queue persistence (optional MVP)

### 5. Configuration System

#### Config File
- [ ] TOML format
- [ ] Multiple provider profiles
- [ ] Keybindings customization
- [ ] Playback settings (mpv path, cache, volume, seek)
- [ ] UI settings (theme, page size, no_emoji)

#### Validation
- [ ] Active profile exists and enabled
- [ ] mpv discoverable (PATH or config)
- [ ] Filesystem roots exist
- [ ] Melodee base_url valid
- [ ] No secrets in error messages

#### Profile Switching
- [ ] Runtime selection menu
- [ ] Stops playback on switch
- [ ] Clears queue on switch
- [ ] Re-initializes provider with spinner

### 6. Theme & Accessibility

- [ ] Rainbow default theme
- [ ] NO_COLOR env var support
- [ ] High-contrast fallback
- [ ] `ui.no_emoji` support
- [ ] Color not only carrier of meaning
- [ ] Graceful at 80×24

### 7. Provider Implementation

#### Filesystem Provider
- [ ] Tag-based SQLite indexing
- [ ] Folder fallback for untagged
- [ ] Fast startup with persisted index
- [ ] Browse artists/albums/tracks
- [ ] Search all fields
- [ ] `file://` stream URLs
- [ ] Paging support

#### Melodee Provider
- [ ] HTTPS authentication
- [ ] Token refresh
- [ ] Browse artists/albums/playlists
- [ ] Search across library
- [ ] Stream URLs with headers
- [ ] Paging support
- [ ] Capability detection

#### Provider Interface
- [ ] Capability detection works
- [ ] Incremental paging
- [ ] Normalized errors
- [ ] Clear capability exposure

### 8. Performance & Responsiveness

- [ ] No blocking in Update loop
- [ ] All I/O async
- [ ] Loading states for async ops
- [ ] Cancelable requests
- [ ] Adaptive UI tick rate
- [ ] First page < 1s on healthy LAN
- [ ] Paging/infinite scroll
- [ ] "Loading more..." indicators

### 9. Error Handling & Observability

- [ ] Structured logging to file
- [ ] Log file: `~/.config/tunez/state/tunez-YYYYMMDD.log`
- [ ] Status line for transient errors
- [ ] Modal for fatal errors
- [ ] Retry with backoff
- [ ] mpv crash handling

### 10. Security & Privacy

- [ ] Secrets never in logs
- [ ] Config tokens supported
- [ ] HTTPS by default
- [ ] No telemetry by default

### 11. Testing

#### Unit Tests
- [ ] Queue operations
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

#### Integration Tests
- [ ] Filesystem provider with fixtures
- [ ] Melodee provider with mocked HTTP

#### Fake mpv Server
- [ ] Accept commands
- [ ] Emit property-change events
- [ ] Emit end-file events
- [ ] Deterministic tests

### 12. Documentation

- [ ] Code comments for complex logic
- [ ] Updated docs if needed
- [ ] Clear error messages
- [ ] User-facing documentation

---

## Specific Code Review Questions

### For `src/internal/app/app.go`
- [ ] Does `Update()` contain any blocking I/O?
- [ ] Are all commands returning `tea.Msg`?
- [ ] Is screen state managed correctly?
- [ ] Are keybindings handled properly?
- [ ] Is selection logic correct?

### For `src/internal/player/player.go`
- [ ] Does mpv IPC handle all platforms (Unix/Windows)?
- [ ] Are events properly subscribed and handled?
- [ ] Is process lifecycle managed (start/stop/reconnect)?
- [ ] Are headers applied for remote streams?

### For `src/internal/providers/filesystem/provider.go`
- [ ] Is SQLite schema correct?
- [ ] Does scanning handle all file types?
- [ ] Are tags read correctly?
- [ ] Is fallback to folder structure working?
- [ ] Are queries efficient?

### For `src/internal/providers/melodee/provider.go`
- [ ] Is auth flow correct?
- [ ] Does token refresh work?
- [ ] Are HTTP errors mapped correctly?
- [ ] Are headers passed to stream?
- [ ] Is paging implemented correctly?

### For `src/internal/config/config.go`
- [ ] Does validation catch all edge cases?
- [ ] Are defaults applied correctly?
- [ ] Is profile switching safe?
- [ ] Are secrets handled properly?

### For `src/internal/queue/queue.go`
- [ ] Are all operations O(1) where possible?
- [ ] Does current index stay valid after removes/moves?
- [ ] Are edge cases handled (empty queue, boundary)?

### For `src/internal/ui/theme.go`
- [ ] Does NO_COLOR work everywhere?
- [ ] Are all styles defined?
- [ ] Is color not the only carrier of meaning?

---

## Common Pitfalls to Check

- [ ] **Blocking in Update**: Any `os.Open`, `http.Get`, `db.Query` in Update?
- [ ] **Context leaks**: Are contexts cancelled?
- [ ] **Race conditions**: Are shared resources protected?
- [ ] **Memory leaks**: Are channels/goroutines cleaned up?
- [ ] **Error handling**: Are all errors handled, not ignored?
- [ ] **Resource cleanup**: Are files/connections closed?
- [ ] **Platform issues**: Unix sockets vs named pipes?
- [ ] **Config validation**: Are all edge cases covered?
- [ ] **Provider errors**: Are they normalized and user-friendly?

---

## Test Coverage Requirements

For each major component, verify tests exist for:
- [ ] Happy path
- [ ] Error cases
- [ ] Edge cases (empty, boundary, invalid input)
- [ ] Concurrency (where applicable)

---

## Performance Checks

- [ ] **No N+1 queries** in provider implementations
- [ ] **Pagination** implemented for large lists
- [ ] **Caching** where appropriate
- [ ] **No busy-waiting** or polling loops
- [ ] **Efficient data structures** (maps vs slices)

---

## Security Checks

- [ ] **No secrets in logs** - verify all logging
- [ ] **Input validation** - all user/provider input validated
- [ ] **Path traversal** - filesystem provider safe
- [ ] **HTTP security** - HTTPS enforced, no insecure defaults
- [ ] **Command injection** - no shell string concatenation

---

## User Experience Checks

- [ ] **Responsive UI** - no lag, no blocking
- [ ] **Clear error messages** - actionable, not cryptic
- [ ] **Loading states** - all async ops show progress
- [ ] **Keyboard navigation** - all functions accessible
- [ ] **Help accuracy** - reflects actual keybindings
- [ ] **Accessibility** - works with NO_COLOR, 80×24

---

## Documentation Review

- [ ] **Code comments** - complex logic explained
- [ ] **Error messages** - user-friendly
- [ ] **Config examples** - accurate and complete
- [ ] **Provider docs** - match implementation

---

## Final Verification

Run these commands and verify success:
```bash
cd /home/steven/source/tunez/src
go build ./...                    # Should succeed
go test ./...                     # Should all pass
go fmt ./...                      # Should be clean
go vet ./...                      # Should have no warnings
```

---

## Report Format

Provide a review report with:

### Summary
- Overall assessment (Pass/Needs Work)
- Critical issues found
- Major issues found
- Minor issues found

### Detailed Findings
For each issue found:
1. **File and line number**
2. **Issue description**
3. **Severity** (Critical/Major/Minor)
4. **Recommendation** (how to fix)
5. **Code snippet** (if applicable)

### Pass/Fail by Component
- TUI Screens: Pass/Fail
- Playback Controls: Pass/Fail
- Queue Management: Pass/Fail
- Configuration: Pass/Fail
- Theme/Accessibility: Pass/Fail
- Providers: Pass/Fail
- Performance: Pass/Fail
- Error Handling: Pass/Fail
- Security: Pass/Fail
- Testing: Pass/Fail

### Recommendations
- Immediate fixes needed
- Improvements for later
- Best practices to adopt

---

## Approval Criteria

**Phase 1 is approved when:**
- [ ] All Phase 1 requirements from PHASE_PLAN.md are met
- [ ] All 12 TUI screens are implemented correctly
- [ ] All playback controls work
- [ ] Queue management is complete
- [ ] Both providers work end-to-end
- [ ] Config system is complete and validated
- [ ] Theme works with NO_COLOR
- [ ] All tests pass
- [ ] No critical or major issues found
- [ ] Code follows architecture rules
- [ ] Performance is acceptable

---

**Begin review by running tests and building the project, then systematically check each item in this checklist.**