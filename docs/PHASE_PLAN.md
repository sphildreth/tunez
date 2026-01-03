# Tunez — Phase Plan

**Last updated:** 2026-01-03  
**Status:** Phase 1 Complete, Phase 2 Ready to Start

This document tracks phase completion status and provides a quick reference for implementation priorities.

---

## Phase Summary

| Phase | Status | Focus |
|-------|--------|-------|
| **Phase 1 (MVP)** | ✅ Complete | Core playback, browsing, TUI |
| **Phase 2 (v1)** | 🔲 Ready | Lyrics, artwork, caching, themes |
| **Phase 3 (v2)** | 🔲 Pending | Command palette, CLI, polish |

---

## Phase 1: MVP — COMPLETE ✅

All acceptance criteria met:

- [x] mpv playback via IPC (Linux/macOS)
- [x] Filesystem provider (scan/index/browse/play)
- [x] Melodee provider (auth/browse/search/stream)
- [x] TUI screens (Loading, Now Playing, Library, Search, Queue, Playlists, Lyrics placeholder, Config, Help, Error handling)
- [x] Config system (load/validate/profiles/keybindings)
- [x] Non-blocking UI (all I/O async)
- [x] NO_COLOR accessibility support
- [x] Provider health status in top bar
- [x] Configurable keybindings (player controls)

---

## Phase 2: v1 Features — READY TO START

**Priority Order:**

### High Priority
1. **Queue Persistence** (2.1)
   - Persist queue to SQLite
   - Restore on startup
   - Files: `queue/persistence.go`, `app.go`, `config.go`

2. **Lyrics Display** (2.2)
   - Fetch from provider (CapLyrics)
   - Scrollable display
   - Files: `providers/*/provider.go`, `app.go`

### Medium Priority
3. **Additional Themes** (2.4)
   - Monochrome and green themes
   - Theme switching
   - Files: `ui/theme.go`, `config.go`

4. **Artwork Display** (2.3)
   - Image-to-ANSI conversion
   - Cache locally
   - Files: `artwork/` (new package), `app.go`

### Lower Priority
5. **Offline Cache** (2.5)
   - Track caching for offline play
   - Files: `cache/` (new package)

6. **Scrobbling** (2.6)
   - Last.fm integration
   - Files: `scrobble/` (new package)

**Full details:** See [PRD.md](PRD.md) Phase 2 section.

---

## Phase 3: v2 Features — PENDING

**Priority Order:**

1. **Command Palette** (3.1)
   - Fuzzy search commands
   - Files: `app/commands.go`, `app/palette.go`

2. **CLI Play Flow** (3.2)
   - `tunez play --search "query"`
   - Files: `cmd/tunez/play.go`

3. **Help Shows Config Keybindings** (3.4)
   - Dynamic help content
   - Files: `app.go`

4. **CLI Utilities** (3.5)
   - `tunez version`, `config init`, `doctor`
   - Files: `cmd/tunez/*.go`

5. **Diagnostics Overlay** (3.3)
   - Debug metrics display
   - Files: `app/diagnostics.go`

**Full details:** See [PRD.md](PRD.md) Phase 3 section.

---

## Implementation Guidelines

### For Coding Agents

1. **Read PRD.md first** - Contains detailed task breakdowns with file lists
2. **Follow existing patterns** - Check existing code for style/structure
3. **Test after each feature** - Run `go test ./...` before moving on
4. **Update docs** - Mark tasks complete, document decisions
5. **Small commits** - One feature per logical change

### File Organization
```
src/
├── cmd/tunez/          # CLI entry point
├── internal/
│   ├── app/            # TUI application (Bubble Tea)
│   ├── artwork/        # Phase 2: Image conversion (new)
│   ├── cache/          # Phase 2: Track caching (new)
│   ├── config/         # Configuration loading
│   ├── logging/        # Structured logging
│   ├── player/         # mpv IPC controller
│   ├── provider/       # Provider interface
│   ├── providers/      # Provider implementations
│   │   ├── filesystem/
│   │   └── melodee/
│   ├── queue/          # Queue management
│   ├── scrobble/       # Phase 2: Last.fm (new)
│   └── ui/             # Theme definitions
```

### Key Constraints
- **No blocking in Update()** - All I/O via `tea.Cmd`
- **Use context.Context** - For cancellation
- **Wrap errors** - With `fmt.Errorf("...: %w", err)`
- **Test coverage** - For non-trivial logic

---

## Completion Criteria

### Phase 2 Complete When:
- [ ] Queue persists across restarts
- [ ] Lyrics display functional
- [ ] 2+ additional themes available
- [ ] Artwork displays (optional)
- [ ] Cache system works (optional)
- [ ] Scrobbling works (optional)
- [ ] All tests pass

### Phase 3 Complete When:
- [ ] Command palette works
- [ ] CLI play flow works
- [ ] Help shows config keybindings
- [ ] CLI utilities work
- [ ] Diagnostics overlay works (optional)
- [ ] All tests pass

---

## References

- [PRD.md](PRD.md) — Detailed requirements and task lists
- [TECH_DESIGN.md](TECH_DESIGN.md) — Architecture overview
- [TUI_UX.md](TUI_UX.md) — Screen specifications
