# Coding Agent Prompt — Implement Tunez Phase 1 (Go + Bubble Tea + mpv)

**Date:** 2026-01-03

You are a coding agent working in an existing repository named **tunez**.

Your goal is to implement **Phase 1 / MVP** of Tunez using the docs in `docs/` as the source of truth (treat them as requirements). Do not preserve or reference any prior/legacy implementations; implement as if starting fresh.

---

## 0) Non-negotiables

- Language: **Go**
- UI: **Bubble Tea** + Bubbles + Lip Gloss
- Playback: **mpv** controlled via **JSON IPC**
- **No blocking** provider I/O or mpv I/O on the Bubble Tea update loop.
- Every I/O action must be cancelable via `context.Context` when possible.
- Code must build and run on **Linux/macOS/Windows** (where mpv exists).
- Default UX theme: **very colorful with rainbow-like ANSI effects** (additional themes later).

---

## 1) Read and follow these docs (in-repo)

Treat these as the authoritative spec:
- `docs/PRD.md`
- `docs/TECH_DESIGN.md`
- `docs/TUI_UX.md`
- `docs/PROVIDERS.md`
- `docs/PROVIDER_FILESYSTEM.md`
- `docs/PROVIDER_MELODEE_API.md`
- `docs/CONFIG.md`
- `docs/TEST_STRATEGY.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/WBS_GUIDE.md`

If you see ambiguity, choose the simplest solution that meets MVP acceptance criteria in `docs/PRD.md` and document any assumptions in `docs/DECISIONS.md`.

---

## 2) Deliverables for Phase 1 (MVP)

Implement Phase 1 end-to-end. The recommended order is the milestones in `docs/IMPLEMENTATION_PLAN.md`:

- Milestone 0 — Project skeleton
- Milestone 1 — Player (mpv IPC)
- Milestone 2 — Core queue + Now Playing
- Milestone 3 — Provider contract + contract tests
- Milestone 4 — Filesystem Provider (MVP)
- Milestone 5 — Melodee Provider (MVP)
- Milestone 6 — TUI polish

### Must ship for Phase 1 (MVP)
1. **Project skeleton** (folders, go module, logging, config)
2. **mpv player service** with IPC (spawn, commands, event loop)
3. **Provider contract** matching `docs/PROVIDERS.md` + contract test harness
4. **Filesystem Provider (MVP)**: scan/index and support browse/search/play per `docs/PROVIDER_FILESYSTEM.md`
5. **Melodee Provider (MVP)**: auth, browse, search, playlists, and streaming per `docs/PROVIDER_MELODEE_API.md`
6. **TUI screens** per `docs/TUI_UX.md` (at minimum: Splash/Loading, Now Playing, Library, Search, Queue, Help overlay, Error toast/modal, Config summary)
7. **Theme**: implement the default `ui.theme = "rainbow"` as the colorful default (other themes can be placeholders for v1+)
8. A runnable CLI: `tunez` that launches the TUI

---

## 3) Definition of Done for Phase 1 (MVP)

Phase 1 is successful when:

- `go test ./...` passes
- `go vet ./...` passes
- `tunez` runs and shows a TUI with:
  - Library browse views (Artists/Albums/Tracks) with paging/infinite scroll
  - Search (`/`) with results across tracks/albums/artists (and playlists when supported)
  - Queue view with add/remove/clear/move
  - ability to select a track and press **Enter** to play
  - playback controls work: play/pause (space), seek (h/l), volume (+/-), next/prev (n/p)
  - top bar shows provider/profile and theme name; bottom Now Playing bar updates (track + time)
- UI remains responsive while loading files (scan occurs async with spinner/status)
- Errors are shown in a status line or modal (no panics)

---

## 4) Suggested repository layout (create if missing)

**All** source code goes into the `src/` folder.

Use this structure (adjust only if the repo already has conventions):

```
src/
  cmd/tunez/main.go
  internal/app/              # bubble tea root model + navigation
  internal/config/           # config load/validate
  internal/player/           # mpv IPC controller
  internal/provider/         # provider contract types + errors + contract tests
  internal/providers/filesystem/
  internal/providers/melodee/
  internal/ui/               # reusable UI components + theme
  internal/logging/          # slog/zerolog setup
test/fixtures/               # optional small fixture library
docs/                        # already provided
```

---

## 5) Implementation steps (do in order)

### Step A — Go module + dependencies
- Initialize Go module (if not present)
- Add deps (suggested; pick equivalents if better):
  - `github.com/charmbracelet/bubbletea`
  - `github.com/charmbracelet/bubbles`
  - `github.com/charmbracelet/lipgloss`
  - TOML parser: `github.com/pelletier/go-toml/v2` or `github.com/BurntSushi/toml`
  - Logging: Go `log/slog` (preferred) or `zerolog`

### Step B — Config loader
- Implement config reading from OS-appropriate config path (and `--config` override)
- Implement validation per `docs/CONFIG.md`:
  - active_profile exists and enabled
  - filesystem roots exist
  - mpv is discoverable (`mpv_path`)
- Implement UI defaults:
  - `ui.theme` defaults to `"rainbow"` when unset
  - `ui.no_emoji` supported
- Provide `examples/config.example.toml` and a `tunez config init` command (optional but nice)

### Step C — Provider contract (minimal)
- Create `internal/provider` package matching `docs/PROVIDERS.md`
- Implement normalized errors as sentinel errors + helpers (e.g., `IsUnauthorized(err)`)

### Step D — Filesystem provider (MVP)
- Implement scan/index with tag-first metadata and folder/filename fallback per `docs/PROVIDER_FILESYSTEM.md`.
- Provide browse and search across Artists/Albums/Tracks with paging.
- Implement `GetStream(trackId)` returning a valid `file://` URL to an absolute path.

### Step E — mpv player service (core of PR)
Implement a player controller that:
- Spawns mpv with IPC:
  - Linux/macOS: unix socket path
  - Windows: named pipe path
- Sends JSON commands:
  - loadfile URL replace
  - set pause yes/no
  - seek +/- seconds
  - set volume
- Subscribes to playback state:
  - Observe `time-pos`, `duration`, `pause`, `volume`, and `end-file`
- Exposes a Go API used by UI:
  - `Play(url, headers)` (headers can be ignored for filesystem)
  - `TogglePause()`, `Seek(delta)`, `SetVolume()`, `Next()/Prev()` (queue driven)
  - Event channel emitting player state updates
- Provide tests using a **fake IPC server** as described in `docs/TEST_STRATEGY.md`

### Step F — Queue + Bubble Tea UI
- Implement root model with:
  - Library screen (list)
  - Search screen
  - Queue screen (list)
  - Config summary screen (read-only)
  - Help overlay (MVP: static defaults acceptable if config-driven help isn’t wired yet)
  - Bottom player bar
- Make all I/O return as `tea.Msg` (no blocking):
  - filesystem scan runs in goroutine; results delivered as message
  - mpv events delivered as messages (wrap from channel)
- Implement keybindings per `docs/TUI_UX.md` (at least the playback keys and navigation keys)

### Step G — Melodee provider (Phase 1 required)
- Implement endpoints listed in `docs/PROVIDER_MELODEE_API.md`:
  - authenticate + refresh
  - browse artists/albums/tracks with paging
  - search
  - playlists list + playlist tracks
  - GetStream returns URL (+ headers if required)
- Ensure provider errors map to normalized errors from `docs/PROVIDERS.md`.
- Ensure secrets are not logged.

### Step H — “Doctor” check (nice-to-have)
- Implement `tunez doctor` command:
  - verify mpv executable
  - verify filesystem roots exist
  - print actionable output

---

## 6) Testing requirements

- Unit tests:
  - queue operations
  - mpv IPC encoding/decoding
  - fake IPC server tests for player
- Provider tests:
  - filesystem provider returns streams that are valid file URLs
- Keep tests fast; integration tests can be behind a build tag.

---

## 7) Commit discipline

- Make small, logical commits.
- Keep PR green throughout.
- Avoid huge refactors or “framework building” beyond what MVP needs.

---

## 8) Notes / constraints

- Implement Melodee after the player + provider contract are stable; keep secrets redacted and never log tokens.
- Keep UI snappy: never walk the filesystem on the UI thread.
- Favor clarity and correctness over cleverness.

---

## 9) Final checklist before marking done

- [ ] `tunez` launches TUI
- [ ] Filesystem scan/index supports browse + search
- [ ] Melodee provider supports browse + search + playlists
- [ ] Enter plays selected track via mpv (file:// and https://)
- [ ] Playback controls work
- [ ] Search (`/`) works across providers
- [ ] Help overlay works (`?`)
- [ ] UI doesn’t freeze during scan
- [ ] Logs are created and secrets are redacted
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
