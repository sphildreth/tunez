# Tunez Documentation

This directory contains design and planning documents for the Tunez terminal music player.

## Terminology

| Term | Definition |
|------|------------|
| **Track** | Any audio file or stream in the library. This is the canonical term used throughout Tunez. |
| **Song** | A specific type of track containing a musical composition with vocals. Not all tracks are songs (e.g., instrumentals, interludes, spoken word, sound effects). |
| **Album** | A collection of tracks released together. |
| **Artist** | The performer or creator associated with tracks/albums. |
| **Queue** | The ordered list of tracks pending playback. |
| **Now Playing** | The currently playing track. |

> **Convention:** Always use "track" in code, UI, logs, and documentation. Avoid "song" except when specifically referring to vocal music compositions.

## Document Overview

| Document | Purpose |
|----------|---------|
| **[PRD.md](PRD.md)** | **Primary reference** - Requirements, phase tasks, acceptance criteria |
| **[PHASE_PLAN.md](PHASE_PLAN.md)** | Quick reference for phase status and priorities |
| **[TUI_UX.md](TUI_UX.md)** | Screen layouts, interactions, keybindings |
| **[TECH_DESIGN.md](TECH_DESIGN.md)** | Architecture, process model, Bubble Tea patterns |
| **[PROVIDERS.md](PROVIDERS.md)** | Provider interface contract |
| **[CONFIG.md](CONFIG.md)** | Configuration file format |
| **[TEST_STRATEGY.md](TEST_STRATEGY.md)** | Testing approach |
| **[SECURITY_PRIVACY.md](SECURITY_PRIVACY.md)** | Security requirements |

## Provider Documentation

| Document | Purpose |
|----------|---------|
| **[PROVIDER_FILESYSTEM.md](PROVIDER_FILESYSTEM.md)** | Local filesystem provider |
| **[PROVIDER_MELODEE_API.md](PROVIDER_MELODEE_API.md)** | Melodee remote API provider |
| **[melodee-api-v1.json](melodee-api-v1.json)** | Melodee API schema |

## Implementation Status

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 (MVP) | ✅ Complete | Core playback, browsing, TUI |
| Phase 2 (v1) | 🔲 Ready | Lyrics, artwork, caching, themes |
| Phase 3 (v2) | 🔲 Pending | Command palette, CLI, polish |

## Getting Started

**For implementers:** Start with [PRD.md](PRD.md) - it contains detailed task breakdowns with file lists for each feature.

**For understanding the app:** Read [TUI_UX.md](TUI_UX.md) for screen layouts and [TECH_DESIGN.md](TECH_DESIGN.md) for architecture.

## Source Code

```
src/
├── cmd/tunez/           # CLI entry point
└── internal/
    ├── app/             # TUI application (Bubble Tea)
    ├── config/          # Configuration loading
    ├── logging/         # Structured logging
    ├── player/          # mpv IPC controller
    ├── provider/        # Provider interface
    ├── providers/       # Provider implementations
    │   ├── filesystem/
    │   └── melodee/
    ├── queue/           # Queue management
    └── ui/              # Theme definitions
```