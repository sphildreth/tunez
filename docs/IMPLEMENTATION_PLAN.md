# Tunez — Implementation Plan (MVP → v1)

**Last updated:** 2026-01-02

## Milestone 0 — Project skeleton
- Go module layout
- Bubble Tea root model scaffold
- Config loading + logging
- `tunez doctor` (mpv presence)

## Milestone 1 — Player (mpv IPC)
- Spawn mpv with IPC
- Commands: load, pause, seek, volume
- Events: time-pos/duration/pause/end-file
- Tests with fake IPC server

## Milestone 2 — Core queue + Now Playing
- Queue structure
- Next/prev semantics
- Now Playing screen + bottom bar
- Keybindings wired to actions

## Milestone 3 — Provider contract + contract tests
- Provider interface + normalized errors
- Contract test harness

## Milestone 4 — Filesystem Provider (MVP)
- Scan/index to SQLite
- Browse + paging
- `file://` stream URLs

## Milestone 5 — Melodee Provider (MVP)
- Auth + token refresh
- Browse + paging
- Search
- Stream URL (+ headers if needed)

## Milestone 6 — TUI polish
- Infinite scroll paging component
- Help overlay generated from bindings
- Config screen (read-only)
- Error/status UX
