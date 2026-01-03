# Phase 1 Quick Review Checklist

**Copy this for a focused code review of Phase 1 implementation.**

---

## Quick Start
```bash
cd /home/steven/source/tunez/src
go build ./... && go test ./... && go fmt ./... && go vet ./...
```

---

## Critical Checks (Must Pass)

### Architecture
- [ ] **No blocking in Update loop** - Search for `http.`, `os.Open`, `db.Query` in `app.go` Update()
- [ ] **All I/O returns tea.Msg** - Check commands use goroutines
- [ ] **Context used everywhere** - All long operations have context
- [ ] **Errors wrapped** - `fmt.Errorf("...: %w", err)` pattern

### TUI Screens (All 12)
- [ ] **Screen 1**: Now Playing shows track, progress, Up Next
- [ ] **Screen 2**: Search works with tab cycling, paging
- [ ] **Screen 3**: Library browse (Artists→Albums→Tracks)
- [ ] **Screen 4**: Queue full management (add/remove/move/clear)
- [ ] **Screen 5**: Playlists (capability-gated)
- [ ] **Screen 6**: Lyrics (placeholder/disabled)
- [ ] **Screens 7-9**: Config screens (read-only + profile select)
- [ ] **Screen 10**: Help with actual keybindings
- [ ] **Screen 11**: Error toasts + modals
- [ ] **Screen 12**: CLI flow (optional)

### Playback & Queue
- [ ] All controls work: play/pause, next/prev, seek, volume, shuffle, repeat
- [ ] Queue: add, remove, clear, move, jump
- [ ] Progress bar displays correctly
- [ ] Volume bounds 0-100%

### Providers
- [ ] **Filesystem**: SQLite indexing, search, paging
- [ ] **Melodee**: Auth, token refresh, search, paging, headers
- [ ] Capability gating works (playlists/lyrics/artwork)
- [ ] Normalized errors (NotFound, Unauthorized, etc.)

### Configuration
- [ ] TOML validation catches all errors
- [ ] Profile switching stops playback & clears queue
- [ ] mpv discovery works
- [ ] No secrets in logs

### Theme & Accessibility
- [ ] Rainbow theme works
- [ ] NO_COLOR env var respected
- [ ] `ui.no_emoji` support
- [ ] Works at 80×24

---

## Major Checks (Should Pass)

### Performance
- [ ] First page loads < 1s
- [ ] Paging/infinite scroll implemented
- [ ] Cancelable requests when navigating away
- [ ] No busy-waiting loops

### Error Handling
- [ ] Structured logging to file
- [ ] Status line for transient errors
- [ ] Modals for fatal errors
- [ ] Retry with backoff

### Testing
- [ ] Unit tests for queue, config, IPC
- [ ] Provider contract tests
- [ ] Integration tests (build-tagged)
- [ ] Fake mpv server

### Security
- [ ] No secrets in logs
- [ ] HTTPS by default
- [ ] Input validation
- [ ] No command injection

---

## Code Quality Checks

### Files to Review
- `src/internal/app/app.go` - TUI model, Update loop
- `src/internal/player/player.go` - mpv IPC
- `src/internal/providers/filesystem/provider.go` - Filesystem
- `src/internal/providers/melodee/provider.go` - Melodee
- `src/internal/config/config.go` - Validation
- `src/internal/queue/queue.go` - Queue ops
- `src/internal/ui/theme.go` - Theme system

### Common Issues to Find
- [ ] Blocking I/O in Update()
- [ ] Goroutine leaks
- [ ] Race conditions
- [ ] Ignored errors
- [ ] Unclosed resources
- [ ] Missing context cancellation
- [ ] Inefficient queries (N+1)
- [ ] Platform-specific code (Unix vs Windows)

---

## Test Verification

Run these and verify:
```bash
go build ./...                    # ✓ Compiles
go test ./...                     # ✓ All pass
go fmt ./...                      # ✓ Clean
go vet ./...                      # ✓ No warnings
```

---

## Pass/Fail Criteria

**PASS if:**
- All critical checks pass
- All 12 screens implemented
- All playback controls work
- Both providers end-to-end
- Config complete
- All tests pass
- No critical issues

**FAIL if:**
- Update loop has blocking I/O
- Any screen missing
- Playback controls broken
- Provider not working
- Config validation missing
- Tests failing
- Critical security issues

---

## Report Format

For each issue found:
1. **File:Line** - Where the issue is
2. **Issue** - What's wrong
3. **Severity** - Critical/Major/Minor
4. **Fix** - How to resolve

---

**Start with: `go build ./... && go test ./...` then check critical items above.**