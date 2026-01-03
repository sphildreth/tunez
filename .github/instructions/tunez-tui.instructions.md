---
description: "Bubble Tea + Lip Gloss TUI patterns for Tunez"
applyTo: "internal/tui/**/*.go"
---

## TUI design principles
- Keep the UI responsive: no disk/network I/O in `Update`.
- Model state must be serializable-ish (avoid embedding live connections directly).
- Use message types (`tea.Msg`) to represent events:
  - user input (keys)
  - provider results (library pages, search results)
  - playback events (track started/paused/position)

## Commands and background work
- Wrap background work as `tea.Cmd` (start a goroutine, return a msg).
- Prefer one-way messages; avoid shared mutable state across goroutines.
- Use contexts for cancellation when switching views or providers.

## View & styling
- Use Lip Gloss styles centrally (theme tokens).
- Ensure the app works without fancy terminal features (no hard dependency on sixel).
- Avoid over-styling; prioritize readability and consistent layout.

## Navigation
- Keymaps should be discoverable (help footer / key legend).
- Provide sane defaults:
  - arrows/jk to move
  - enter to select
  - space to play/pause
  - / to search
  - q/esc to back/quit depending on screen
