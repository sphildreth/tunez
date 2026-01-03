# Tunez — Technical Design

**Last updated:** 2026-01-02  
**Language:** Go  
**UI:** Bubble Tea (Charmbracelet)  
**Playback:** mpv (external process) via JSON IPC

## 1. Architecture overview

Tunez consists of four major subsystems:

1. **TUI (Bubble Tea)**  
   - Renders screens (Library, Search, Queue, Now Playing, Help)
   - Dispatches user intents as commands
   - Shows progress/spinners for async work

2. **Core state + navigation**  
   - Owns the current screen, selection, and view models
   - Maintains queue and playback state
   - Coordinates provider calls and caching

3. **Provider layer**  
   - Implements the `Provider` contract (`docs/PROVIDERS.md`)
   - Built-in Providers in Phase 1:
     - Filesystem
     - Melodee API

4. **Player (mpv controller)**  
   - Spawns mpv as a child process
   - Communicates over mpv JSON IPC
   - Emits playback events to core (time position, pause, end-of-file, errors)

## 2. Process model and concurrency

### 2.1 Golden rule
**No provider IO and no mpv IPC may run on the Bubble Tea update loop.**  
All IO runs in goroutines and returns results to the UI as `tea.Msg`.

### 2.2 Recommended internal channels
- `uiMsgCh` (provider results, errors)
- `playerEvtCh` (mpv events: property-change, end-file, etc.)
- `workQueue` (background jobs: scan, prefetch, cache refresh)

### 2.3 Cancellation
- Every provider request MUST have a `context.Context`.
- When user navigates away from a view, in-flight requests SHOULD be cancelled.
- When scrolling triggers paging, older paging requests may be superseded/cancelled.

## 3. Bubble Tea screen strategy

### 3.1 Root model
- Root `Model` holds:
  - active profile/provider
  - active screen enum
  - shared components (top bar, bottom player bar)
  - child screen model (Library/Search/Queue/etc.)
  - modal overlays (help, error popup)

### 3.2 Lists at scale (800k albums)
Bubble Tea’s standard `list` component is great for small/medium lists. For very large lists, Tunez SHOULD implement a **paged list adapter**:

- Keep only the currently visible window + a small buffer in memory.
- Fetch additional pages when selection approaches end.
- Show “Loading more…” row while fetching.
- Maintain a `nextCursor` per list.

## 4. Playback with mpv JSON IPC

### 4.1 IPC transport
- Linux/macOS: Unix socket (`--input-ipc-server=/tmp/tunez-mpv.sock`)
- Windows: named pipe (`\\.\pipe\tunez-mpv`)

Tunez spawns mpv with:
- `--idle=yes` (stay alive between tracks)
- `--input-ipc-server=...`
- `--no-terminal`
- `--force-window=no`

### 4.2 Commands
Tunez uses IPC `command` messages:
- `loadfile <url> replace`
- `set pause yes/no`
- `seek <seconds> relative|absolute`
- `set volume <0-100>`

### 4.3 Observing state
Tunez subscribes to property changes:
- `time-pos`, `duration`, `pause`, `volume`, `media-title`
- `end-file` event

### 4.4 Stream headers (remote providers)
Provider returns `StreamInfo { URL, Headers }`.
Player applies headers via mpv options/properties before calling `loadfile`.

## 5. Data storage and caching

### 5.1 Local database (recommended)
Use SQLite for:
- Filesystem index
- Remote metadata cache (artists/albums/tracks pages)
- Queue persistence (v1)

Cross-platform option: `modernc.org/sqlite` (pure Go).

### 5.2 Cache layers
- **In-memory LRU** for “recent pages”
- **SQLite** for durable index/cache
- **HTTP ETags** where supported

## 6. Configuration
- TOML config (see `docs/CONFIG.md`)
- Profiles + keybindings

## 7. Testing approach (high level)
See `docs/TEST_STRATEGY.md`:
- Provider contract tests
- Fake mpv IPC server
- Integration tests behind build tag `integration`
