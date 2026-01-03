# Tunez (Restart) — Go + Bubble Tea + mpv

**Last updated:** 2026-01-02

Tunez is a cross-platform terminal music player built with:
- **Go** (core app + providers)
- **Bubble Tea** (TUI) + Bubbles + Lip Gloss
- **mpv** (playback) via **JSON IPC** (a small `mpv` sidecar process controlled by Tunez)

This zip contains a fresh set of product + technical docs that define Tunez from scratch.

## What’s inside
- `docs/PRD.md` — product requirements and acceptance criteria
- `docs/TECH_DESIGN.md` — architecture + key technical decisions
- `docs/TUI_UX.md` — screens, interactions, keybindings, accessibility
- `docs/PROVIDERS.md` — provider interface contract
- `docs/PROVIDER_FILESYSTEM.md` — built-in local filesystem provider spec
- `docs/PROVIDER_MELODEE_API.md` — built-in Melodee API provider spec
- `docs/CONFIG.md` — config schema + profiles + keybindings
- `docs/IMPLEMENTATION_PLAN.md` — milestone plan (MVP → v1)
- `docs/WBS_GUIDE.md` — work breakdown structure (slices)
- `docs/TEST_STRATEGY.md` — unit/contract/integration testing approach
- `docs/SECURITY_PRIVACY.md` — secrets, auth, privacy expectations

## Quick MVP definition
MVP is complete when:
1. User can browse/search library from **Filesystem** and **Melodee API** providers.
2. User can enqueue and play tracks with **instant** UI response (no blocking while loading).
3. Playback works reliably through **mpv IPC** (play/pause/seek/next/prev/volume).
4. Config + keybindings work, and the app has a built-in help overlay.
