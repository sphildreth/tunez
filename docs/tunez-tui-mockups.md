# Tunez TUI — Initial Screen Mockups (v0)

These are **ASCII layout mockups** to guide UI structure, navigation, and information density.
They’re designed to be implementable in **ratatui** with a consistent frame layout.

Legend:
- `█` / `▓` / `░` represent intensity bars (visualizer/progress).
- `⟲` / `⟳` etc. are optional; replace with ASCII if desired.
- Bracketed labels like `[F1]` indicate keybinding hints.
- Color is implied via **style tokens** (e.g., *Accent*, *Dim*, *Warn*), not shown literally.

---

## Global Layout Regions

Most screens share:
- **Top Status Bar**: app + provider/profile + network + scrobble + clock
- **Left Nav**: tabs + provider selector
- **Main Pane**: context-specific (lists, details, lyrics)
- **Bottom Player Bar**: now playing + progress + controls + volume

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez  ▸ Provider: melodee (home)  Net: OK  Scrobble: ON  Theme: AfterDark   │
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
│ ⏵  Men At Work — Down Under  [03:14/03:42]  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%  ♫    │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 0 — Splash / Loading

Purpose: instant feedback while Tunez loads config, providers, and restores state.

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
│                      Tip: Press ? at any time for keys                        │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ Status: Starting…   Log: ~/.local/state/tunez/tunez.log                       │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 1 — Main / Now Playing

Purpose: primary “dashboard” with visualizer and quick actions.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez  ▸ Provider: melodee (home)  Net: OK  Scrobble: ON  Queue: 12  [F1 Help]│
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Track: Men At Work — Down Under                               │
│ Search        │  Album: Business as Usual (1981)                                │
│ Library       │  Artist: Men At Work                                           │
│ Playlists     │  Codec: FLAC  |  Rate: 44.1kHz  |  Stream: 320kbps             │
│ Queue         │                                                               │
│ Lyrics        │  Visualizer: Spectrum (60 FPS adaptive)                        │
│ Config        │                                                               │
│ Help          │  ║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║  ║▁▃▅▇█▇▅▃▁║  ║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║            │
│               │  ║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║  ║▁▃▅▇█▇▅▃▁║  ║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║            │
│               │  ║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║  ║▁▃▅▇█▇▅▃▁║  ║▁▂▃▄▅▆▇█▇▆▅▄▃▂▁║            │
│               │                                                               │
│               │  Up Next:                                                     │
│               │   1) Be Good Johnny                                           │
│               │   2) Touching the Untouchables                                │
│               │                                                               │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏸  Men At Work — Down Under   03:14/03:42  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%  Rep:Off│
│ [Space]Play/Pause [←→]Seek [n/p]Next/Prev [s]Shuffle [r]Repeat [q]Queue       │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 2 — Search

Purpose: fast keyboard-driven search with filters and immediate play/queue.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Search  Provider: melodee (home)         / Query: "men at work cargo" │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Filters: [Artist] Men At Work   [Album] Cargo   [Year] any    │
│ Search        │           [Type] Tracks (t)  Albums (a)  Playlists (p)         │
│ Library       │                                                               │
│ Playlists     │  Results (Tracks)                                         1/12 │
│ Queue         │  ┌──────────────────────────────────────────────────────────┐ │
│ Lyrics        │  │  ▶  01  Dr. Heckyll & Mr. Jive     3:39  Cargo (1983)    │ │
│ Config        │  │     02  Overkill                   3:45  Cargo (1983)    │ │
│ Help          │  │     03  It's a Mistake             4:33  Cargo (1983)    │ │
│               │  │     04  High Wire                  3:06  Cargo (1983)    │ │
│               │  └──────────────────────────────────────────────────────────┘ │
│               │                                                               │
│               │  Actions: [Enter]Play  [A]Add to Queue  [P]Play Next  [I]Info  │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏵  (not playing)  Tip: Press TAB to cycle result type (Tracks/Albums/Playlists)│
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 3 — Library (Browse)

Purpose: provider-driven browse (Artists/Albums/Genres/Recently etc.)

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Library  Provider: melodee (home)   View: Albums  Sort: Recently Added │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Albums                                                   1/40 │
│ Search        │  ┌──────────────────────────────────────────────────────────┐ │
│ Library       │  │  ▣ Cargo — Men At Work (1983)                            │ │
│ Playlists     │  │  ▢ Business as Usual — Men At Work (1981)                │ │
│ Queue         │  │  ▢ The Visitors — ABBA (1981)                            │ │
│ Lyrics        │  │  ▢ Purple Rain — Prince (1984)                           │ │
│ Config        │  └──────────────────────────────────────────────────────────┘ │
│ Help          │                                                               │
│               │  Details                                                     │
│               │  ┌──────────────────────────────────────────────────────────┐ │
│               │  │ Cargo (1983)                                              │ │
│               │  │ Men At Work                                               │ │
│               │  │ Tracks: 10  Duration: 38:12                               │ │
│               │  │ [Enter]Open  [p]Play  [A]Add Album  [S]Shuffle Album      │ │
│               │  └──────────────────────────────────────────────────────────┘ │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏵  Men At Work — Down Under   03:14/03:42  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%          │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 4 — Queue

Purpose: manage playback queue (reorder, remove, play next).

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Queue  Items: 12   Mode: Normal   Shuffle: Off   Repeat: Off          │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  ┌──────────────────────────────────────────────────────────┐ │
│ Search        │  │  ▶  01  Down Under                 3:42                   │ │
│ Library       │  │     02  Be Good Johnny             3:33                   │ │
│ Playlists     │  │     03  Touching the Untouchables  3:39                   │ │
│ Queue         │  │     04  Catch a Star               3:28                   │ │
│ Lyrics        │  │     05  Overkill                   3:45                   │ │
│ Config        │  └──────────────────────────────────────────────────────────┘ │
│ Help          │                                                               │
│               │  Actions: [d]Remove  [D]Clear  [u/j]Move Up/Down  [S]Shuffle  │
│               │           [p]Play Selected  [P]Play Next  [s]Toggle Shuffle   │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏸  Men At Work — Down Under   03:14/03:42  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%          │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 5 — Playlists

Purpose: list/search playlists (provider-driven), open, queue.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Playlists  Provider: melodee (home)     / Search: "workout"           │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Playlists                                               1/18 │
│ Search        │  ┌──────────────────────────────────────────────────────────┐ │
│ Library       │  │  ▣ Night Drive (42 tracks)                               │ │
│ Playlists     │  │  ▢ Workout Mix (85 tracks)                               │ │
│ Queue         │  │  ▢ 80s Classics (120 tracks)                             │ │
│ Lyrics        │  └──────────────────────────────────────────────────────────┘ │
│ Config        │                                                               │
│ Help          │  Tracks (selected playlist)                                   │
│               │  ┌──────────────────────────────────────────────────────────┐ │
│               │  │  01  Down Under — Men At Work                             │ │
│               │  │  02  Africa — Toto                                       │ │
│               │  │  03  Take On Me — a-ha                                   │ │
│               │  └──────────────────────────────────────────────────────────┘ │
│               │  Actions: [Enter]Open  [A]Add All  [p]Play Playlist  [I]Info  │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏵  (not playing)                                                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 6 — Lyrics

Purpose: read lyrics with scroll, optional “follow along” later.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Lyrics  Provider: melodee (home)   Track: Down Under                 │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  ┌──────────────────────────────────────────────────────────┐ │
│ Search        │  │ Traveling in a fried-out combie                           │ │
│ Library       │  │ On a hippie trail, head full of zombie                    │ │
│ Playlists     │  │ I met a strange lady, she made me nervous                │ │
│ Queue         │  │ She took me in and gave me breakfast                     │ │
│ Lyrics        │  │ …                                                       │ │
│ Config        │  │ …                                                       │ │
│ Help          │  └──────────────────────────────────────────────────────────┘ │
│               │  [↑↓]Scroll  [g/G]Top/Bottom  [f]Follow (v2)                  │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ ⏸  Men At Work — Down Under   03:14/03:42  ▓▓▓▓▓▓▓▓▓░░░░░░  Vol: 72%          │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 7 — Configuration (Main)

Purpose: configure provider, profile, theme, keybindings, cache, scrobbling.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Config                                                              │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Now Playing   │  Sections                                                    │
│ Search        │  ┌──────────────────────────────────────────────────────────┐ │
│ Library       │  │  ▣ Providers & Profiles                                   │ │
│ Playlists     │  │  ▢ Theme & ANSI                                            │ │
│ Queue         │  │  ▢ Keybindings                                             │ │
│ Lyrics        │  │  ▢ Cache / Offline                                          │ │
│ Config        │  │  ▢ Scrobbling                                               │ │
│ Help          │  │  ▢ Logging & Diagnostics                                   │ │
│               │  └──────────────────────────────────────────────────────────┘ │
│               │                                                               │
│               │  Details (selected section)                                   │
│               │  ┌──────────────────────────────────────────────────────────┐ │
│               │  │ Default Provider:  melodee                                │ │
│               │  │ Profile:           home                                   │ │
│               │  │ Theme:             AfterDark                              │ │
│               │  │ Visualizer:         Spectrum (bars)                       │ │
│               │  │ Scrobbling:         Enabled (melodee)                     │ │
│               │  │ Cache:             Off (provider unsupported)             │ │
│               │  └──────────────────────────────────────────────────────────┘ │
│               │  [Enter]Edit  [S]Save  [Esc]Back                              │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ Tip: Config file: ~/.config/tunez/config.toml   Secrets: OS Keyring          │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 8 — Config: Providers & Profiles

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Config ▸ Providers & Profiles                                        │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Config        │  Providers                                                    │
│               │  ┌──────────────────────────────────────────────────────────┐ │
│               │  │  ▣ melodee     (remote)  profiles: home, lab              │ │
│               │  │  ▢ filesystem  (local)   profiles: music, downloads       │ │
│               │  └──────────────────────────────────────────────────────────┘ │
│               │                                                               │
│               │  Provider Details                                             │
│               │  ┌──────────────────────────────────────────────────────────┐ │
│               │  │ Provider: melodee                                         │ │
│               │  │ Base URL:  https://music.example.com                      │ │
│               │  │ User:      steven@example.com                             │ │
│               │  │ Auth:      Logged in (token in keyring)                   │ │
│               │  │ Capabilities: playlists, lyrics, scrobble                 │ │
│               │  │ Offline Download: NO                                      │ │
│               │  └──────────────────────────────────────────────────────────┘ │
│               │  Actions: [L]Login  [O]Logout  [E]Edit Profile               │
├───────────────┴──────────────────────────────────────────────────────────────┤
│ [Tab]Switch lists  [Esc]Back                                                 │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 9 — Config: Cache / Offline (Provider-gated)

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Tunez ▸ Config ▸ Cache / Offline                                             │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Config        │  Provider: filesystem (music)                                 │
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
│ Note: Rights/DRM concerns are between user and provider, not Tunez.          │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 10 — Help / Keybindings Overlay

Purpose: quick reference; appears as overlay on any screen.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Help — Keys (press ? to close)                                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ Navigation      j/k: up/down   h/l: left/right   Enter: select/open           │
│ Playback        Space: play/pause   n/p: next/prev   ←/→: seek                │
│ Queue           A: add to queue   P: play next   d: remove   D: clear         │
│ Search          /: focus search   Tab: change search type                     │
│ Views           1-9: jump tabs   Esc: back/close modal                         │
│ Misc            : (colon) command palette (v2)   Ctrl+C: quit                 │
│                                                                              │
│ Tips                                                                     [OK] │
│ - Use `tunez play --artist ... --album ... -p` to jump straight into playback │
│ - Press `I` on items for details                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 11 — Error Modal / Toast

Purpose: non-blocking error display; skip invalid tracks as required.

### Toast
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ [WARN] Stream failed (timeout). Retrying… (2/5)                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Modal
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Error                                                                         │
├──────────────────────────────────────────────────────────────────────────────┤
│ Could not decode track: unsupported codec or corrupted stream.                │
│ Action: skipped track and moved to next in queue.                             │
│                                                                              │
│ [View Logs]   [OK]                                                            │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 12 — CLI “Play then Launch TUI” Flow (visual)

This is the experience for:
`tunez play --provider melodee --artist "Men At Work" --album "Cargo" -p`

```
1) TUI launches instantly → shows “Searching…” state
2) Queue is filled as results arrive
3) Playback begins as soon as first track stream URL is resolved
4) Screen transitions to Now Playing with spectrum active
```

Mock:
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

## Notes for Implementation in ratatui
- Use a central `AppState` and an event loop with:
  - `UiTick` (e.g., 16–50ms adaptive)
  - `PlayerEvent` (progress, track changed, error)
  - `ProviderEvent` (search results, browse pages)
- Keep modals/overlays as a stack of views (`Vec<Modal>`).
- Make visualizer resolution adaptive:
  - fewer bins and lower FPS when terminal is small or CPU is constrained.

