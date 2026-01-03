# Tunez Documentation

This directory contains all design and planning documents for the Tunez terminal music player.

## Quick Navigation

### 🎯 **Start Here**
- **[PHASE_PLAN.md](PHASE_PLAN.md)** - **Comprehensive phase breakdown** mapping ALL requirements to deliverable phases. This is your primary reference for implementation.

### 📋 **Requirements & Specifications**
- **[PRD.md](PRD.md)** - Product Requirements Document (high-level requirements)
- **[TUI_UX.md](TUI_UX.md)** - Complete TUI/UX specification with screen layouts and interactions
- **[CONFIG.md](CONFIG.md)** - Configuration file format and settings
- **[PROVIDERS.md](PROVIDERS.md)** - Provider interface contract

### 🏗️ **Architecture & Design**
- **[TECH_DESIGN.md](TECH_DESIGN.md)** - Technical architecture, process model, and Bubble Tea strategy
- **[SECURITY_PRIVACY.md](SECURITY_PRIVACY.md)** - Security and privacy considerations
- **[DECISIONS.md](DECISIONS.md)** - Architectural trade-offs and clarifications

### 🔧 **Provider-Specific Documentation**
- **[PROVIDER_FILESYSTEM.md](PROVIDER_FILESYSTEM.md)** - Filesystem provider details
- **[PROVIDER_MELODEE_API.md](PROVIDER_MELODEE_API.md)** - Melodee API provider details
- **[melodee-api-v1.json](melodee-api-v1.json)** - Melodee API schema

### 🧪 **Testing**
- **[TEST_STRATEGY.md](TEST_STRATEGY.md)** - Unit, integration, and provider contract testing approaches

## Documentation Hierarchy

```
PHASE_PLAN.md (Master implementation plan)
    ↓
PRD.md (Requirements)
    ↓
TUI_UX.md, TECH_DESIGN.md, PROVIDERS.md, CONFIG.md (Detailed specs)
    ↓
Provider-specific docs, TEST_STRATEGY.md (Implementation details)
```

## Key Documents Summary

### PHASE_PLAN.md
The comprehensive phase plan that maps ALL requirements from all documents into three deliverable phases:
- **Phase 1 (MVP)**: Core terminal music player
- **Phase 2 (v1)**: Enhanced features (cache, lyrics, artwork, scrobbling)
- **Phase 3 (v2)**: Advanced UX (command palette, diagnostics, CLI flow)

When all phases are complete, the application will be code-complete.

### TUI_UX.md
Defines 12 screens with ASCII layouts, interactions, and keybindings:
- Screen 0: Splash/Loading
- Screen 1: Main/Now Playing
- Screen 2: Search
- Screen 3: Library
- Screen 4: Queue
- Screen 5: Playlists (capability-gated)
- Screen 6: Lyrics (capability-gated)
- Screens 7-9: Configuration
- Screen 10: Help
- Screen 11: Error handling
- Screen 12: CLI flow

### PROVIDERS.md
Defines the Provider interface contract:
- Capability system (playlists, lyrics, artwork)
- Paging conventions
- Error normalization
- Stream info structure

## Implementation Status

**Current State (Phase 0 - Foundation Complete):**
- ✅ Basic app scaffold
- ✅ Config, logging, provider interface
- ✅ Player controller with mpv IPC
- ✅ Queue implementation
- ✅ Theme system
- ✅ Filesystem & Melodee providers (basic)
- ✅ Basic TUI screens

**Next Steps:**
See PHASE_PLAN.md Phase 1 for complete checklist of MVP features to implement.

## Related Resources

- **Source code**: `/home/steven/source/tunez/src/`
- **Example config**: `/home/steven/source/tunez/examples/config.example.toml`
- **Test fixtures**: `/home/steven/source/tunez/src/test/fixtures/`