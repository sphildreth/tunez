# Tunez — TUI/UX Specification (Bubble Tea)

**Last updated:** 2026-01-02

## 1. Global layout

```
┌──────────────────────────────────────────────────────────────────────┐
│ Tunez  [Profile: Home Files]  [Status: OK]            /search…        │
├───────────────┬──────────────────────────────────────────────────────┤
│ Library       │ Main Pane                                              │
│ Search        │ (paged lists, details, lyrics)                          │
│ Queue         │                                                        │
│ Playlists     │                                                        │
│ Now Playing   │                                                        │
│ Config        │                                                        │
│ Help (?)      │                                                        │
├───────────────┴──────────────────────────────────────────────────────┤
│ ⏵  Artist — Album — Track Title                     01:23 / 04:56  65% │
│ [━━━━━━━╺━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━]  Vol 70  Shuffle Off  Repeat Off│
└──────────────────────────────────────────────────────────────────────┘
```

## 2. Screens
- Splash / Loading
- Library (Artists / Albums / Tracks)
- Search
- Queue
- Now Playing
- Help overlay
- Config (read-only in MVP)

## 3. Keybindings (defaults)

Global:
- `q` : back/close (or quit at root)
- `ctrl+c` : quit
- `?` : help overlay
- `/` : search
- `tab` : next nav section
- `shift+tab` : previous nav section

Navigation:
- `j/k` or `down/up` : move selection
- `g/G` : top/bottom
- `enter` : open/play
- `esc` : back/close modal

Playback:
- `space` : play/pause
- `n` : next
- `p` : previous
- `h/l` : seek -5s / +5s
- `H/L` : seek -30s / +30s
- `- / +` : volume down/up
- `m` : mute
- `s` : shuffle toggle
- `r` : cycle repeat (off → all → one)

Queue:
- `x` : remove selected item
- `C` : clear queue
- `u/d` : move item up/down

## 4. Accessibility and terminal compatibility
- Works in 80x24 (degrades gracefully)
- Avoid color-only cues
- Optional “no emoji” mode
