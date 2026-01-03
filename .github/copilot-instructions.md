# Tunez – GitHub Copilot Repository Instructions

## Project goal
Build **Tunez**, a fast, reliable terminal music player written in **Go** using:
- **Bubble Tea** (TUI state machine + input handling)
- **Lip Gloss** (styling)
- **mpv** (playback) via IPC (Unix socket on Linux/macOS; named pipe fallback on Windows if supported later)

Tunez must feel responsive (no laggy UI), be correct over cleverness, and be easy to extend.

## How to work in this repo
1. **Read the docs** in `/doc/` before implementing a feature. Treat them as the source of truth.
2. Work in small, reviewable steps. Prefer a working vertical slice over scaffolding.
3. **No blocking work in Bubble Tea `Update`**:
   - I/O (filesystem scans, API calls), JSON parsing, and long computations must run in background commands/goroutines and return `tea.Msg`.
4. Always keep the code buildable:
   - `go test ./...`
   - `gofmt -w .` (or `go fmt ./...`)
   - Avoid unused exports and dead code.

## Architecture expectations (high-level)
- Keep a clean separation:
  - `internal/tui`: UI state, models, views, message types, keymaps.
  - `internal/player`: mpv integration, playback state, events.
  - `internal/providers`: library/metadata providers (filesystem, API).
  - `cmd/tunez`: wiring + config + dependency graph.
- Prefer small interfaces for seams (player, providers, cache, logger).
- Use `context.Context` for cancellation; avoid goroutine leaks.

## Quality bar
When generating or modifying code, you must:
- Include error handling with wrapped errors (`fmt.Errorf("...: %w", err)`).
- Add tests for non-trivial logic (parsing, filtering, sorting, provider behavior).
- Prefer deterministic code: stable sorts, explicit timeouts, bounded concurrency.
- Avoid “magic” or pseudo-code. Output complete, compiling Go code.

## mpv integration rules
- Do not shell out with string concatenation.
- Prefer `exec.CommandContext` with args, or mpv IPC JSON commands.
- Treat file paths and URLs as untrusted input (validate/sanitize).
- Ensure mpv process lifecycle is owned (start/stop/reconnect), and handle crashes.

## PR / agent workflow expectations
If acting as an agent:
- Create a plan (bullets), then implement step-by-step.
- After each step, run tests and fix issues.
- Keep diffs small; avoid unrelated refactors.
