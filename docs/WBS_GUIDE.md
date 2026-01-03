# Tunez — Work Breakdown Structure (WBS Guide)

**Last updated:** 2026-01-02

## Slice-based plan

### Slice 1 — App scaffold
- Go module + folder structure
- Config loader + validation
- Logging + log file location
- Root Bubble Tea model + nav

### Slice 2 — mpv Player service
- Start/stop mpv with IPC
- JSON IPC client
- Playback events → core state

### Slice 3 — Queue + playback UX
- Queue add/remove/reorder
- Next/prev
- Now Playing UI + bottom bar
- Keybindings dispatcher

### Slice 4 — Provider contract
- Provider interface
- Paging conventions
- Contract test suite

### Slice 5 — Filesystem Provider
- Scanner + tag reader
- SQLite index
- Browse + paging
- Stream URL mapping

### Slice 6 — Melodee Provider
- HTTP client + auth
- Browse + paging
- Search
- StreamInfo mapping

### Slice 7 — Large list UX
- Paged list component (infinite scroll)
- Cancelable loads
- Loading/error rows
