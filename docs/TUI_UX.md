# Tunez — TUI/UX Specification (Bubble Tea)

**Last updated:** 2026-01-03

This document defines the full screen set and interactions for Tunez. It is derived from the original Tunez TUI mockups, updated to the **Go + Bubble Tea + mpv** stack and the provider contract in `docs/PROVIDERS.md`.

---

## Visual Legend (ASCII)

These conventions appear in the reference layouts below:

- `█` / `▓` / `░` are intensity/progress bars.
- Bracketed labels like `[F1]` indicate keybinding hints.
- Icons (like `⏵`) are optional; if `ui.no_emoji = true` or the font renders poorly, replace with ASCII (`>` / `||` etc.).
- Color is implied via style tokens (Accent/Dim/Warn) and must not be the only carrier of meaning.

## Theme (Requirement)

- Default theme: very colorful with rainbow-like ANSI effects.
- Additional themes will be added later (v1+), including monochromatic and “green terminal” styles.
- Regardless of theme, the UI must not rely on color alone to convey meaning.

---

## Global Layout Regions

Tunez uses a stable layout with four regions:

1. **Top Bar**
  - App title
  - Active provider + profile
  - Provider health status (OK / Degraded / Offline)
  - Network status (OK / Degraded / Offline) when applicable
  - Scrobble status (ON / OFF) when applicable
  - Theme name (for support/debugging)
  - Clock (optional)
  - Quick hint for help

2. **Left Navigation**
   - Library
   - Search
   - Queue
   - Playlists (capability-gated)
   - Lyrics (capability-gated)
   - Configuration
   - Help

3. **Main Pane**
   - The active screen content (lists, details, editors, etc.)
   - Supports paging / infinite scroll for large libraries

4. **Bottom Player Bar**
   - Play state icon (playing/paused)
   - Track/artist/album
   - Progress bar + elapsed/remaining
   - Volume, shuffle, repeat

Example layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez  ▸ Provider: filesystem (music)  Net: OK  Scrobble: OFF  Theme: Default │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │                                                              │
│ Search        │                         MAIN PANE                             │
│ Library       │                                                              │
│ Playlists     │                                                              │
│ Queue         │                                                              │
│ Lyrics        │                                                              │
│ Config        │                                                              │
│ Help          │                                                              │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏵  Artist — Track Title  [01:23/04:56]  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 70%   [? Help] │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Responsiveness rule:** No provider I/O or mpv IPC runs in the Bubble Tea update loop. All I/O returns results via `tea.Msg`.

**Event model (recommended):**
- UI ticks: a periodic `UiTickMsg` (e.g., 16–100ms adaptive) for progress bars and visualizer.
- Player events: `PlayerProgressMsg`, `TrackChangedMsg`, `PlayerErrorMsg`.
- Provider events: `SearchResultsMsg`, `BrowsePageMsg`, `ProviderErrorMsg`.

**Overlays:** Help and modals are overlays rendered above the active screen. Prefer a simple overlay stack to avoid screen-specific modal logic.

---

## Screen 0 — Splash / Loading

**Purpose**
- Show startup progress: config load, profile init, filesystem scan, remote auth, cache warmup.

**Elements**
- Spinner + status lines:
  - “Loading config…”
  - “Starting mpv…”
  - “Initializing Provider: …”
  - “Scanning library…” (filesystem)
  - “Authenticating…” (remote)

**Transitions**
- On success → Screen 1 (Main / Now Playing)
- On fatal error → Screen 11 (Error Modal) with “Exit / Open config path / Retry”

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez — Terminal music player in full ANSI color                              │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│                              ░▒▓█ T U N E Z █▓▒░                             │
│                                                                              │
│                      Loading config…                [ OK ]                   │
│                      Discovering providers…         [ OK ]                   │
│                      Restoring session…             [ .. ]                   │
│                                                                              │
│                      Tip: Press ? at any time for keys                       │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ Status: Starting…   Log: ~/.local/state/tunez/tunez.log                       │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 1 — Main / Now Playing

**Purpose**
- Default landing screen.
- Shows current track details, progress, and playback controls.
- Offers “Up Next” preview from queue.

**Main Pane**
- Track title, artist, album
- Optional: codec/bitrate (if known)
- Large progress bar
- Up Next list (next 3–10 items)

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez  ▸ Provider: melodee (home)  Net: OK  Scrobble: ON  Queue: 12  [? Help] │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Track: Men At Work — Down Under                              │
│ Search        │  Album: Business as Usual (1981)                               │
│ Library       │  Artist: Men At Work                                          │
│ Playlists     │  Codec: FLAC  |  Rate: 44.1kHz  |  Stream: 320kbps            │
│ Queue         │                                                               │
│ Lyrics        │  Visualizer: Spectrum (adaptive FPS)                          │
│ Config        │                                                               │
│ Help          │  ║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║  ║▁▃▅▇█▇▅▃▁║  ║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║            │
│               │                                                               │
│               │  Up Next:                                                     │
│               │   1) Be Good Johnny                                           │
│               │   2) Touching the Untouchables                                │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏸  Men At Work — Down Under   03:14/03:42  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%  Rep:Off │
│ [Space]Play/Pause [h/l]Seek [n/p]Next/Prev [s]Shuffle [r]Repeat [q]Queue      │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Actions**
- `space`: play/pause
- `n/p`: next/prev
- `h/l`: seek -5/+5 seconds
- `H/L`: seek -30/+30 seconds
- `-/+`: volume down/up
- `m`: mute
- `s`: shuffle toggle
- `r`: repeat cycle (off → all → one)

---

## Screen 2 — Search

**Purpose**
- Global search across Tracks/Albums/Artists/Playlists (capability-gated).

**Interaction**
- `/` opens inline search input (or a modal input).
- Results grouped sections:
  - Tracks
  - Albums
  - Artists
  - Playlists (if supported)

**Result-type switching**
- `tab` cycles result type (Tracks/Albums/Artists/Playlists)
- Optional accelerators: `t` (tracks), `a` (albums), `r` (artists), `p` (playlists)

**Selection behavior**
- Track: `enter` plays now (replace queue) or enqueue+play based on config
- Album/Artist/Playlist: `enter` jumps to Library/Playlists scoped view

**Large-library requirement**
- Search supports paging and incremental loading (“Loading more…” row).

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Search  Provider: melodee (home)         / Query: "men at work cargo" │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Filters: [Artist] Men At Work   [Album] Cargo   [Year] any   │
│ Search        │           [Type] Tracks (t)  Albums (a)  Playlists (p)        │
│ Library       │                                                               │
│ Playlists     │  Results (Tracks)                                         1/12│
│ Queue         │  ┌──────────────────────────────────────────────────────────┐│
│ Lyrics        │  │  ▶  01  Dr. Heckyll & Mr. Jive     3:39  Cargo (1983)     ││
│ Config        │  │     02  Overkill                   3:45  Cargo (1983)     ││
│ Help          │  │     03  It's a Mistake             4:33  Cargo (1983)     ││
│               │  │     04  High Wire                  3:06  Cargo (1983)     ││
│               │  └──────────────────────────────────────────────────────────┘│
│               │                                                               │
│               │  Actions: [Enter]Play  [A]Add to Queue  [P]Play Next  [I]Info │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏵  (not playing)  Tip: Press TAB to cycle result type (Tracks/Albums/…)       │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 3 — Library (Browse)

**Purpose**
- Browse Artists, Albums, Tracks.

**Main Pane Modes**
- Artists list
- Albums list (optionally filtered by selected artist)
- Tracks list (optionally filtered by album/artist)

**Expected controls**
- `tab` cycles library sub-modes (Artists/Albums/Tracks)
- `enter`:
  - on Artist → filter Albums/Tracks
  - on Album → show Tracks
  - on Track → play/enqueue

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Library  Provider: melodee (home)   View: Albums  Sort: Recently Added│
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Albums                                                   1/40│
│ Search        │  ┌──────────────────────────────────────────────────────────┐│
│ Library       │  │  ▣ Cargo — Men At Work (1983)                            ││
│ Playlists     │  │  ▢ Business as Usual — Men At Work (1981)                ││
│ Queue         │  │  ▢ The Visitors — ABBA (1981)                            ││
│ Lyrics        │  │  ▢ Purple Rain — Prince (1984)                           ││
│ Config        │  └──────────────────────────────────────────────────────────┘│
│ Help          │                                                               │
│               │  Details                                                     │
│               │  ┌──────────────────────────────────────────────────────────┐│
│               │  │ Cargo (1983)                                              ││
│               │  │ Men At Work                                               ││
│               │  │ Tracks: 10  Duration: 38:12                               ││
│               │  │ [Enter]Open  [p]Play  [A]Add Album  [S]Shuffle Album      ││
│               │  └──────────────────────────────────────────────────────────┘│
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏵  Men At Work — Down Under   03:14/03:42  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%          │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Performance**
- Uses paging/infinite scroll.
- Shows a spinner or “Loading…” row during fetches.

---

## Screen 4 — Queue

**Purpose**
- View and manage the play queue.

**Actions**
- `enter`: jump+play selected queue item
- `x`: remove selected item
- `C`: clear queue
- `u/d`: move item up/down
- `n/p`: next/prev still operate globally

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Queue  Items: 12   Mode: Normal   Shuffle: Off   Repeat: Off          │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  ┌──────────────────────────────────────────────────────────┐│
│ Search        │  │  ▶  01  Down Under                 3:42                   ││
│ Library       │  │     02  Be Good Johnny             3:33                   ││
│ Playlists     │  │     03  Touching the Untouchables  3:39                   ││
│ Queue         │  │     04  Catch a Star               3:28                   ││
│ Lyrics        │  │     05  Overkill                   3:45                   ││
│ Config        │  └──────────────────────────────────────────────────────────┘│
│ Help          │                                                               │
│               │  Actions: [x]Remove  [C]Clear  [u/d]Move Up/Down              │
│               │           [Enter]Play Selected  [P]Play Next                  │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏸  Men At Work — Down Under   03:14/03:42  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%          │
└──────────────────────────────────────────────────────────────────────────────┘
```

**UX**
- Current playing item visually marked.
- Supports large queues with paging if needed.

---

## Screen 5 — Playlists (Capability-gated)

**Purpose**
- Browse playlists, open one, enqueue/play.

**Provider requirements**
- Only visible if `CapPlaylists` is true for the active provider.

**Views**
- Playlists list
- Playlist detail (tracks)

**Actions**
- `enter` on playlist → open playlist tracks
- `a` (optional) add playlist to queue
- `enter` on track → play/enqueue

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Playlists  Provider: melodee (home)     / Search: "workout"           │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Playlists                                               1/18│
│ Search        │  ┌──────────────────────────────────────────────────────────┐│
│ Library       │  │  ▣ Night Drive (42 tracks)                               ││
│ Playlists     │  │  ▢ Workout Mix (85 tracks)                               ││
│ Queue         │  │  ▢ 80s Classics (120 tracks)                             ││
│ Lyrics        │  └──────────────────────────────────────────────────────────┘│
│ Config        │                                                               │
│ Help          │  Tracks (selected playlist)                                   │
│               │  ┌──────────────────────────────────────────────────────────┐│
│               │  │  01  Down Under — Men At Work                             ││
│               │  │  02  Africa — Toto                                        ││
│               │  │  03  Take On Me — a-ha                                    ││
│               │  └──────────────────────────────────────────────────────────┘│
│               │  Actions: [Enter]Open  [A]Add All  [p]Play Playlist  [I]Info  │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏵  (not playing)                                                             │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 6 — Lyrics (Capability-gated)

**Purpose**
- Display lyrics for currently playing track or selected track.

**Provider requirements**
- Only visible if provider supports lyrics.

**Behavior**
- When track changes, lyrics panel attempts to load lyrics asynchronously.
- States:
  - Loading…
  - No lyrics available
  - Lyrics text (scrollable)

**Controls**
- `j/k` scroll
- `g/G` top/bottom
- `q/esc` back

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Lyrics  Provider: melodee (home)   Track: Down Under                  │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  ┌──────────────────────────────────────────────────────────┐│
│ Search        │  │ Traveling in a fried-out combie                           ││
│ Library       │  │ On a hippie trail, head full of zombie                    ││
│ Playlists     │  │ I met a strange lady, she made me nervous                ││
│ Queue         │  │ She took me in and gave me breakfast                      ││
│ Lyrics        │  │ …                                                        ││
│ Config        │  │ …                                                        ││
│ Help          │  └──────────────────────────────────────────────────────────┘│
│               │  [j/k]Scroll  [g/G]Top/Bottom                                 │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏸  Men At Work — Down Under   03:14/03:42  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%          │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 7 — Configuration (Main)

**Purpose**
- Read-only summary of current configuration and quick navigation to sub-screens.

**Content**
- Active profile
- Provider summary (with secrets redacted)
- mpv path + IPC mode
- Keybinding summary
- Cache status

**Actions**
- `enter` selects a config section:
  - Providers & Profiles
  - Theme & ANSI
  - Cache / Offline (if relevant)
  - Keybindings (view)
  - Scrobbling
  - Logging & Diagnostics

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Config                                                               │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Sections                                                     │
│ Search        │  ┌──────────────────────────────────────────────────────────┐│
│ Library       │  │  ▣ Providers & Profiles                                   ││
│ Playlists     │  │  ▢ Theme & ANSI                                           ││
│ Queue         │  │  ▢ Keybindings                                            ││
│ Lyrics        │  │  ▢ Cache / Offline                                        ││
│ Config        │  │  ▢ Scrobbling                                             ││
│ Help          │  │  ▢ Logging & Diagnostics                                  ││
│               │  └──────────────────────────────────────────────────────────┘│
│               │                                                               │
│               │  Details (selected section)                                   │
│               │  ┌──────────────────────────────────────────────────────────┐│
│               │  │ Default Provider:  melodee                                ││
│               │  │ Profile:           home                                   ││
│               │  │ Theme:             Default                                ││
│               │  │ Visualizer:        Spectrum (bars)                        ││
│               │  │ Scrobbling:        Enabled (melodee)                      ││
│               │  │ Cache:             Off (provider unsupported)             ││
│               │  └──────────────────────────────────────────────────────────┘│
│               │  [Enter]Open  [Esc]Back                                       │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ Tip: Config file: ~/.config/tunez/config.toml   Secrets: OS Keyring           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 8 — Config: Providers & Profiles

**Purpose**
- Manage active profile selection at runtime (MVP: select; edit via file).

**Views**
- Profiles list
- Profile detail summary (redacted)

**Actions**
- `enter`: set active profile (re-initialize provider with spinner)
- `o` (optional): open config file path (print path + instructions)
- `r` (optional): retry provider initialization

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Config ▸ Providers & Profiles                                         │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Config        │  Providers                                                     │
│               │  ┌──────────────────────────────────────────────────────────┐│
│               │  │  ▣ melodee     (remote)  profiles: home, lab              ││
│               │  │  ▢ filesystem  (local)   profiles: music, downloads       ││
│               │  └──────────────────────────────────────────────────────────┘│
│               │                                                               │
│               │  Provider Details                                              │
│               │  ┌──────────────────────────────────────────────────────────┐│
│               │  │ Provider: melodee                                         ││
│               │  │ Base URL:  https://music.example.com                      ││
│               │  │ User:      steven@example.com                             ││
│               │  │ Auth:      Logged in (token in keyring)                   ││
│               │  │ Capabilities: playlists, lyrics, scrobble                 ││
│               │  └──────────────────────────────────────────────────────────┘│
│               │  Actions: [Enter]Select  [r]Retry  [o]Config Path             │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ [Tab]Switch lists  [Esc]Back                                                  │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 9 — Config: Cache / Offline (Provider-gated)

**Purpose**
- Show cache status and offline-mode controls.

**MVP**
- View-only is acceptable:
  - cache DB path
  - cache size estimate
  - last refresh time (if known)

**v1+**
- Cache clear/rebuild
- Offline mode toggles (provider-gated)

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Config ▸ Cache / Offline                                              │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Config        │  Provider: filesystem (music)                                  │
│               │                                                               │
│               │  Offline Download: ENABLED (supported by provider)            │
│               │                                                               │
│               │  Download Location:  /mnt/music/.tunez-cache                  │
│               │  Max Cache Size:      20 GB                                   │
│               │  Eviction Policy:     LRU                                     │
│               │  TTL:                14 days                                  │
│               │                                                               │
│               │  [Enter]Edit  [S]Save  [C]Clear Cache                         │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ Note: Rights/DRM concerns are between user and provider, not Tunez.           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 10 — Help / Keybindings Overlay

**Purpose**
- Show keybindings for global + current screen actions.

**Requirement**
- Help MUST reflect the current keybinding map from config (not hard-coded) when possible.
- MVP: may ship with defaults hard-coded *only if* config keybind parsing is not yet implemented; document this in `docs/DECISIONS.md`.

**Controls**
- `?` toggles overlay
- `esc/q` closes

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Help — Keys (press ? to close)                                                │
├──────────────────────────────────────────────────────────────────────────────┤
│ Navigation      j/k: up/down   h/l: left/right   Enter: select/open           │
│ Playback        Space: play/pause   n/p: next/prev   ←/→: seek                │
│ Queue           A: add to queue   P: play next   x: remove   C: clear         │
│ Search          /: focus search   Tab: change result type                     │
│ Views           Tab/Shift+Tab: cycle tabs   Esc: back/close overlay           │
│ Misc            Ctrl+C: quit                                                  │
│                                                                              │
│ Tips                                                                     [OK] │
│ - Use `tunez play --artist ... --album ... -p` to jump straight into playback │
│ - Press `I` on items for details                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 11 — Error Modal / Toast

Tunez uses two mechanisms:

### Toast
- Non-blocking message in status line area
- Auto-dismiss after N seconds
- Used for transient issues (retrying provider call, mpv reconnect, etc.)

### Modal
- Blocking overlay requiring dismissal
- Used for fatal or user-action-required errors:
  - mpv not found
  - provider unauthorized
  - config invalid

Modal actions:
- **Retry**
- **Open config path** (prints path + instructions)
- **Exit**

Reference layouts (ASCII):

Toast:

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ [WARN] Stream failed (timeout). Retrying… (2/5)                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

Modal:

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Error                                                                        │
├──────────────────────────────────────────────────────────────────────────────┤
│ Could not decode track: unsupported codec or corrupted stream.               │
│ Action: skipped track and moved to next in queue.                            │
│                                                                              │
│ [View Logs]   [OK]                                                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 12 — CLI “Play then Launch TUI” Flow (visual)

Tunez supports a CLI mode that can start playback and then drop into the TUI:

Examples:
- `tunez play --track <id>`
- `tunez play --search "name"`

Flow:
1. Resolve track(s) using active profile/provider
2. Start mpv playback
3. Launch TUI directly into Now Playing with queue initialized

MVP: optional. If not implemented in the first phase, keep a placeholder command that prints “Not implemented yet” and returns non-zero.

Reference layout (ASCII):

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Resolving request…  Provider: melodee (home)                          │
├──────────────────────────────────────────────────────────────────────────────┤
│ Searching: artist="Men At Work" album="Cargo"                                 │
│ Best match: Cargo (1983)                                                      │
│ Loading tracks…  [#####-----] 6/10                                            │
│ Starting playback…                                                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Default Keybindings (Reference)

Global:
- `q` : back/close (or quit at root)
- `ctrl+c` : quit
- `?` : help
- `/` : search
- `tab` / `shift+tab` : next/prev left-nav section

Navigation:
- `j/k` or `down/up` : selection
- `g/G` : top/bottom
- `enter` : open/play
- `esc` : back/close

Playback:
- `space` : play/pause
- `n` : next
- `p` : previous
- `h/l` : seek -5s / +5s
- `H/L` : seek -30s / +30s
- `- / +` : volume down/up
- `m` : mute
- `s` : shuffle toggle
- `r` : repeat cycle

Library/Search common actions:
- `A` : add selection to queue (track/album/playlist)
- `P` : play next (enqueue as next)
- `I` : info/details

Queue:
- `x` : remove
- `C` : clear
- `u/d` : move up/down

---

## Terminal Compatibility & Accessibility

- Must degrade gracefully at 80×24.
- Avoid color-only meaning; use symbols + text labels.
- Support `ui.no_emoji = true` to avoid emoji icons if fonts render poorly.
