# Tunez — Requirements (PRD) — Phase 1 (Built‑in Providers)

## 1. Overview

**Product name:** `Tunez`  
**Executable:** `tunez`  
**Tagline:** “Terminal music player in full ANSI color.”  
**Type:** Cross‑platform CLI + terminal UI (TUI) music player  
**Platforms:** Linux, macOS, Windows (native terminal)  
**Extension model (Phase 1):** Modular Rust workspace with **built‑in Providers** (developer/power-user extensibility)

**Extensibility terminology**
- **Built-in Providers (Phase 1):** Rust crates compiled into the `tunez` binary.
- **External plugins (Phase 2, optional):** out-of-tree extensions loaded/hosted by Tunez (not supported in Phase 1).

### 1.1 Problem statement
Tunez is a *fast, keyboard-first, colorful* terminal player that can browse/search a library and play music from one or more Providers—while also being “fun”: smooth transitions, animated UI widgets, and a real-time spectrum/waveform visualization.

**Target user proficiency**
- Tunez is designed for terminal-comfortable users (tmux/SSH, keyboard-first workflows) who are comfortable editing TOML config files.

### 1.2 Why a TUI instead of a GUI?
- Runs anywhere: SSH, tmux, headless boxes, minimal desktops
- “Always there” experience like classic terminal players, but modern/animated
- Fun engineering challenge (audio pipeline + rendering loop + async networking)

### 1.3 Phase map
- This section is a **status checklist** for tracking progress by phase.
- The **normative requirements** are the detailed sections referenced below (e.g., Sections 4–10).

**Important (repo status vs PRD):** This Phase map is meant to reflect the *current repository implementation status*.
If you change the code, update the checkboxes here so the PRD doesn't drift and confuse readers.
As of **2026-01-01**, this repo includes a working **TUI shell** (tab navigation + Markdown help overlay), plus
provider/player/audio crates with comprehensive unit tests — but **the mockup screens are not yet fully implemented/wired end-to-end**.
**Progress Update**: Phase 1 is now **COMPLETE**. All sub-phases (A-J) are implemented and verified, including visualization, scrobbling, monochrome mode, and comprehensive contract tests.
Missing: Phase 2 (Plugins) and Phase 3 (Polish/Fast Extras) are pending future work.

- [ ] **Phase 1** — Built-in Providers (this document)
  - Phase 1 sub-phases (recommended; sized for incremental implementation)
    - [x] Phase 1A: Workspace scaffolding + logging + config loading skeleton (6.6, 7.5)
    - [x] Phase 1B: Core domain + Provider/Scrobbler traits + error types + capability flags (4.1, 4.10)
    - [x] Phase 1C: CLI wiring (provider/profile selection, providers list) + config validation (4.1.3, 7.5.1, 9.1)
      - Progress update: CLI `play` subcommand now validates selector precedence and resolves provider/profile overrides.
    - [x] Phase 1D: Minimal TUI shell + navigation + help overlay rendering (5.0–5.2.1)
    - [x] Phase 1E: Queue + playback state machine (no real audio) + unit tests (4.5–4.6, 10.3.1)
    - [x] Phase 1F: Audio pipeline MVP (decode + output) behind the player interface (4.5, 7.3)
    - [x] Phase 1G: Provider MVPs (filesystem + remote example) + contract tests (4.1.4, 10.3.4)
      - [x] Implement `filesystem` Provider MVP (see docs/filesystem-provider-prd.md)
      - [x] Implement `melodee` Provider MVP (see docs/melodee-provider-prd.md)
    - [x] Phase 1H: Visualization MVP + animation cadence + fallbacks (5.4, 7.4)
    - [x] Phase 1I: Scrobbling MVP + persistence + contract tests (4.10, 10.3.5)
    - [x] Phase 1J: Cross-platform polish + accessibility/monochrome + docs sweep (5.5, 6.4, 6.9)
    - [x] Phase 1 Done: MVP acceptance criteria met + quality gates passing (10.1–10.2)

---

## 1.4 Next steps — get a working, playable TUI (implementation plan)

This section is an *implementation-oriented* checklist to turn the current shell into a working application where
users can navigate the mockup screens, search/browse via a Provider, and play music.

### Milestone 0 — run the current shell (baseline)
- [x] `cargo run -p tunez-cli` launches the TUI reliably (already true today).
- [x] Visualization MVP with adaptive FPS and graceful fallbacks (Phase 1H complete).
- [x] Terminal minimum size matches PRD: enforce **80x24** (PRD §5.1).
- [x] Make `tunez play ...` launch the TUI (PRD §9.1) and optionally start playback when `--autoplay` is set.

### Milestone 1 — provider wiring (make the UI actually talk to a Provider)
- [x] Add a small "provider runtime" that can construct the selected Provider from `Config`/`ProviderSelection`.
  - Inputs: provider id + optional profile name.
  - Output: a `dyn Provider` instance plus capabilities.
  - Constraint: do not block the UI thread on network/filesystem operations.
- [x] Extend the UI context to include access to the selected Provider (or a channel to a background task that owns it).
- [x] Ensure all Provider errors surface via `ProviderError` categories (PRD §4.1.1 / §4.9).

### Milestone 2 — implement the mockup screens as real views (minimum playable slice)
Target the smallest end-to-end loop first:

- [x] **Search view** (PRD §4.3, §5.0):
  - Enter search mode with `/`.
  - Call `Provider::search_tracks(...)` and show paged results.
  - `Enter` on a result: enqueue + start playing (or enqueue-only if not autoplay; pick one and document it).

- [x] **Queue view** (PRD §4.6, §5.0):
  - Display the queue and current selection.
  - Support add/remove/clear/shuffle (at minimum: add + remove + clear).

- [x] **Now Playing view** (PRD §4.5, §5.0):
  - Show current track metadata.
  - Provide play/pause/next/prev and progress display.

### Milestone 3 — hook playback end-to-end (Provider → stream URL → audio → UI state)
- [x] On play: resolve `TrackId` → `Provider::get_stream_url(...)` (PRD §4.1.1, §12.1).
- [x] Feed the stream URL into the audio layer (`tunez-audio`) via the player (`tunez-player`).
- [x] Wire playback controls from the UI to the `Player` state machine (Space, n/p, seek, volume).
- [x] Periodically update elapsed/progress UI without blocking (PRD §4.10.1 tick cadence can be reused).

### Milestone 4 — browsing + playlists (capability-gated)
- [x] **Library/Browse view**: call `Provider::browse(...)` and render deterministically paged results (PRD §4.4, §4.1.1).
- [x] **Playlists view** (only if `capabilities.playlists = true`): list/search/open playlists (PRD §4.7).

### Milestone 5 — resilience and UX polish required for “working app”
- [x] Implement non-blocking error UI (toast/banner) and retry affordances (PRD §4.9).
- [x] Ensure scrobbling stays opt-in and never blocks playback (PRD §4.10.2, §6.5).
- [x] Keep secrets out of config files (PRD §4.2).

### Milestone 6 — Definition of Done alignment
- [x] When the above milestones are implemented, re-check Phase 1 sub-phases and "Phase 1 Done" in §1.3.
- [x] Only mark Phase 1 done when the MVP acceptance criteria (§10.1) are demonstrably true end-to-end.

- [x] **Phase 2** — External plugins (optional; see Roadmap)
  - [x] Add a plugin host (exec-based or dylib-based) that adapts plugins to the Provider interface (11)
  - [x] Keep Phase 1 Providers working unchanged (11)

- [x] **Phase 3** — Fancy extras / polish (optional; see Roadmap)
  - [x] More visualization modes (tunez-viz supports 4 modes) (11)
  - [x] Better caching and offline modes (filesystem-provider caching) (11)
  - [x] Theme editor (via Config tab) (11)

---

## 2. Goals and Non-Goals

### 2.1 Goals (Phase 1)
1. **Playback**: play audio locally (streaming and/or local files) with robust buffering and minimal stutter.
2. **A rich TUI**: browsing, searching, queue/now-playing, progress, volume, shuffle/repeat.
3. **Color + animation**: smooth progress bar, transitions between views, loading spinners.
4. **Spectrum / waveform visualization** synchronized with playback (at least a spectrum analyzer).
5. **Cross-platform**: consistent behavior on Linux/macOS/Windows terminals.
6. **Modular from day 1**: clean internal boundaries so new **Providers** can be added by developers without reworking the core.

The Phase 1 architecture SHOULD remain future-proof for Phase 2 (external plugins) by keeping Provider/Scrobbler interfaces stable and isolating provider-specific code.

### 2.2 Non-goals (Phase 1)
- Third-party “drop-in install” plugins (no plugin marketplace, no plugin folder install)
- In-process dynamic library loading (`.dll/.so/.dylib`)
- Multi-room audio / casting
- Editing remote metadata unless trivial and safe
- A full “Spotify client” feature set (social, collaborative playlists, etc.)

---

## 3. Users and Use Cases

### 3.1 Primary user persona
- Power user who lives in terminals (tmux, SSH)
- Wants “no-mouse” control and a slick now-playing experience
- Uses one or more music backends (local files and/or servers)
- Is comfortable editing config files and (optionally) building from source

### 3.2 Core user stories
- As a user, I can select a Provider (e.g., remote server API or local filesystem) and browse/play music from it.
- As a user, I can search and start playback immediately.
- As a user, I can browse artists/albums/playlists and queue items.
- As a user, I can control playback with hotkeys and see a lively spectrum/waveform.
- As a power user, I can add a new Provider by implementing the Provider interface in Rust and rebuilding Tunez.

---

## 4. Functional Requirements

### 4.1 Modular Provider Architecture (Core Requirement — Phase 1)

**Terminology (industry standard)**
- **Core**: the `tunez` application (UI + playback engine + config + logging).
- **Provider** (a.k.a. backend/source): a Rust component that supplies library data and streams (remote APIs or local filesystem).
- **Capability**: an optional feature a Provider may implement (lyrics, artwork, playlists, etc.).

**Phase 1 implementation**
- Providers are **built-in** Rust crates/modules compiled into the Tunez binary.
- Tunez exposes a stable internal interface (`Provider` trait) so developers can add new Providers cleanly.
- Providers can be enabled/disabled via compile-time features and runtime configuration.

#### 4.1.1 Provider interface
Tunez MUST define a Provider abstraction (`Provider` trait) that includes a common set of operations. At minimum, the trait MUST include:

**MUST (Core operations exposed by Tunez)**
- `search_tracks(query, filters, paging)`
- `browse(kind: artists|albums|playlists|genres, paging)`
- `list_playlists(paging)`
- `search_playlists(query, paging)`
- `get_playlist(playlist_id)` and `list_playlist_tracks(playlist_id)`
- `get_album(album_id)` and `list_album_tracks(album_id)`
- `get_track(track_id)` (metadata)
- `get_stream_url(track_id)` -> returns a **stream URL** (Provider returns a URL only in Phase 1)

Provider implementations that do not support a given operation MUST return `ProviderError::NotSupported`.

**SHOULD**
- `lyrics(track_id)`
- `artwork(entity_id)` (track/album/artist)
- `favorites()`, `recently_played()`
- playlist management if the backend supports it

**Metadata baseline (Phase 1)**
- Track metadata returned by Providers MUST include enough fields for the UI and scrobbling to function. At minimum: title, primary artist name, and a stable provider-scoped track id.
- Track duration SHOULD be provided when known (seconds).
- A stable provider-scoped track id MUST be treated as an opaque identifier. It MUST be stable across runs for the same backend entity, case-sensitive, and serializable in TOML/JSON.
- Track ids SHOULD be reasonably bounded in size (e.g., <= 256 bytes) to avoid pathological memory/logging issues.

**Error contract (Phase 1)**
- Provider operations MUST return `Result<T, ProviderError>` (defined in the `tunez-core` crate) rather than ad-hoc error types.
- `ProviderError` MUST be a stable, provider-agnostic classification so the UI and core logic can handle failures consistently.
- At minimum, `ProviderError` MUST include categories equivalent to:
  - `NetworkError` (connectivity / DNS / TLS / transport)
  - `AuthenticationError` (missing/expired credentials, forbidden)
  - `NotFound` (missing track/album/playlist IDs)
  - `NotSupported` (operation/capability not implemented by Provider or not available for the current account/server)
  - `Other` (fallback with message/context)

**Paging and ordering (Phase 1)**
- All list/search operations that accept `paging` MUST behave deterministically for a given query while paging (stable ordering within a single logical query).
- Providers MAY impose a maximum page size; if the caller requests more, the Provider MUST either clamp to its maximum and continue paging correctly, or return a clear error.
- Provider responses SHOULD include enough paging metadata to continue (cursor/token or offset/limit).

#### 4.1.2 Capability model
- The Provider capability model MUST include a flag for **offline downloading support** (e.g., `supports_offline_download: bool`).
- If `supports_offline_download` is false (e.g., the Melodee API Provider), Tunez MUST hide/disable any offline download UI/actions for that Provider.

- The Core MUST be able to query whether a Provider supports optional capabilities.
- UI MUST degrade gracefully (hide tabs/actions) when a capability is absent.

**Capability discovery semantics (Phase 1)**
- Capability flags MUST represent the Provider’s best-known support level at startup/config time.
- If a capability cannot be known up-front (e.g., depends on server version or account permissions), the Provider MAY advertise it optimistically and return `ProviderError::NotSupported` for specific calls.
- When a call returns `NotSupported`, Tunez SHOULD treat it as a non-fatal signal to hide/disable the related UI action for the remainder of the session (or until provider/profile changes).

#### 4.1.3 Provider selection and profiles
**MUST**
- User can pick a Provider by config (`default_provider`) and/or CLI (`--provider <id>`).
- Support per-provider configuration blocks (e.g., server URL, local library root).
- The primary configuration file format MUST be **TOML** (Rust CLI standard), loaded from the OS-appropriate config directory (via `directories`).

**SHOULD**
- Support multiple profiles (e.g., `home`, `lab`) and provider-specific profiles.

#### 4.1.4 Developer extensibility (Phase 1)
**MUST**
- Document the Provider interface and provide a template Provider (skeleton).
- Provide at least one “reference Provider” implementation (remote API) and one “local Provider” (filesystem).

**SHOULD**
- Use Cargo features to include/exclude providers:
  - e.g., `--features provider-filesystem,provider-melodee`
- Maintain a `providers/` workspace folder with a consistent structure and tests.

**NICE-TO-HAVE**
- A `tunez dev provider new <name>` generator (v2).

---

### 4.2 Connection & Authentication (Provider-dependent)
**MUST**
- Provider configuration supports server base URL and identity (username/email) where applicable.
- **Security Critical**: do NOT store passwords in config.
- Tokens/refresh tokens MUST be stored in the OS keyring (Keychain / Credential Manager / Secret Service).
- Providers MUST handle token expiry/refresh gracefully (refresh when possible; otherwise return `ProviderError::AuthenticationError` and guide the user to re-auth).

**SHOULD**
- Providers interacting with rate-limited APIs SHOULD implement client-side backoff (and optional rate limiting) on 429/503-style responses.
- Login UX:
  - `tunez auth login --provider <id> --profile <name>`
  - `tunez auth status`
  - `tunez auth logout`

---

### 4.3 Library Search and Discovery
**MUST**
- Search tracks by text query.
- Show paged results and allow keyboard navigation.

**Edge cases**
- Tunez MUST remain responsive on large libraries (tens of thousands of tracks) by using paging and incremental rendering.
- Tunez MUST handle slow/unreliable networks gracefully by showing loading/progress states, timing out, and allowing retries without UI hangs.

**SHOULD**
- Filtered search (artist/album/year/genre) where supported by Provider.
- Search for albums/artists/playlists too.

---

### 4.4 Browsing Views
**MUST**
- Views (tabs) for:
  - **Now Playing**
  - **Queue**
  - **Search**
  - **Library** (provider-driven browse)
- Detail panels for current selection (duration, artist, album, etc.)

**SHOULD**
- Provider-specific views (e.g., “Favorites”, “Recent”) when capabilities exist.

---

### 4.5 Playback
**MUST**
- Play audio on the local audio device.
- Controls:
  - Play/Pause
  - Next/Previous
  - Seek ±5s/±30s
  - Volume up/down + mute
- Progress display: elapsed/remaining (+ buffered indicator recommended)

**SHOULD**
- Gapless-ish playback (best-effort) by pre-buffering the next track.
- Output device selection (where feasible).

**NICE-TO-HAVE**
- OS-level media controls (MPRIS on Linux, SMTC on Windows).

---

### 4.6 Queue Management
**MUST**
- Local playback queue:
  - Add track(s) to end / play next
  - Remove item(s)
  - Clear queue
  - Shuffle queue

**SHOULD**
- Persist queue locally between runs.
- Queue persistence SHOULD be resilient: if the persisted queue/state is corrupt or unreadable, Tunez SHOULD start with an empty queue, show a non-fatal warning, and keep the corrupt file for debugging.
- Tunez SHOULD keep a last-known-good backup of the queue state (best-effort) to support recovery.

---

### 4.7 Playlists
**MUST (MVP)**
- Provider MUST be able to **list playlists** and **search playlists** (when the backend supports playlists).
- Open a playlist, list tracks, and play/queue them.

**V2 (Not Phase 1)**
- Create/rename/delete playlists and playlist reordering.

---

### 4.8 Lyrics
**SHOULD**
- Fetch and display lyrics when available.
- Provide scrolling; “follow along” optional.

---

### 4.9 Errors, Offline, and Resilience
**MUST**
- Graceful error states in the UI (banner/toast).
- Timeouts and retries with exponential backoff.
- If stream fails mid-track, allow retry or skip.
- **Unknown/invalid track metadata** (e.g., unreadable file/unsupported codec/missing duration): Tunez MUST log the error, show a user-visible message, and skip the track.
- Provider failures MUST be surfaced through `ProviderError` categories (e.g., network vs auth vs not found) so the UI can present consistent, actionable states across Providers.

**User experience during common failures (Phase 1)**
- Network failures (`ProviderError::NetworkError`) MUST be shown as non-blocking UI state (banner/toast) that includes: the affected Provider, whether an automatic retry will occur, and an actionable next step (retry, switch provider, check network).
- Authentication failures (`ProviderError::AuthenticationError`) MUST prompt the user toward the supported login/logout flow for the current provider/profile (without exposing secrets), and MUST keep the UI responsive (browsing may be disabled for that Provider until re-auth).
- Audio device failures (device lost/unavailable/permission denied) MUST present a clear UI message and MUST not crash; Tunez SHOULD attempt to recover automatically where feasible (e.g., re-open default device) and otherwise allow the user to pause/stop playback.

Where feasible, these error surfaces SHOULD include keyboard-first actions (e.g., Retry, Login, View details/logs) consistent with the toast/modal patterns referenced in the TUI mockups.

**SHOULD**
- Cache metadata to speed up navigation.
- “Offline mode” for browsing cached metadata (not necessarily playing).

---

### 4.10 Scrobbling / Play Reporting (MVP)
Tunez MUST support scrobbling / play-event reporting via a modular component:

- **Scrobbler**: a Rust component responsible for reporting playback events to a backend service.
- Scrobblers MUST be pluggable at build-time (built-in crates), similar to Providers.

Scrobbling MUST be **disabled by default** unless explicitly enabled/configured by the user (see Privacy).

#### 4.10.1 Scrobbler-driven reporting model (Phase 1)
To support different backends with different reporting rules, Tunez SHOULD NOT hard-code scrobbling thresholds (e.g., “10 seconds” or “50%”). Instead:

**MUST**
- Tunez MUST provide the Scrobbler with:
  - Track identity (provider id + track id) and metadata needed for reporting
  - **Track duration** (seconds) when known
  - **Playback position / played duration** (seconds), updated periodically
  - Playback state transitions (started/resumed/paused/stopped/ended)
  - A stable `player_name` identifier (e.g., `Tunez`) and an optional `device_id`
- The Scrobbler MUST decide:
  - when to emit “Now Playing”-style events
  - when to emit “Played/submission”-style events
  - any periodic progress updates, if desired

**SHOULD**
- Tunez SHOULD call a Scrobbler tick/update hook at a reasonable cadence while playing (default: **once per second**), but the Scrobbler decides what to do with it.
- Tunez SHOULD reuse the same 1-second tick event used for progress/elapsed-time UI updates to drive Scrobbler telemetry, to avoid redundant timers.

#### 4.10.2 Failure handling
**MUST**
- If scrobbling is not configured or not supported, Tunez MUST continue functioning normally.
- If a Scrobbler returns an error, Tunez MUST log it and show a non-blocking UI indicator; playback must not be interrupted.

**SHOULD**
- Batch and retry scrobble events on transient network failures.
- Provide a UI indicator (“Scrobbling: on/off/error”).

#### 4.10.3 Scrobbling configuration and persistence (Phase 1)
**MUST**
- Scrobbling MUST be configurable per provider/profile (enable/disable).
- Tunez MUST support selecting a default Scrobbler when multiple are compiled in (e.g., `default_scrobbler` in config).
- Tunez MUST persist pending scrobble events to local storage so they can be retried after restarts and while offline.
- The persistence queue MUST be bounded (by count and/or age) to avoid unbounded growth.

**SHOULD**
- Pending events SHOULD be retried with backoff and jitter and pruned once acknowledged.
- The UI indicator SHOULD reflect when events are queued for retry vs successfully sent.
- Config UI SHOULD allow the user to view basic scrobbling status (on/off/error) and purge pending queued events.


## 5. TUI/UX Requirements

### 5.0 TUI mockups (reference)
The initial UI layout and screen breakdown are captured as ASCII mockups in [docs/tunez-tui-mockups.md](docs/tunez-tui-mockups.md). This document SHOULD be treated as the canonical reference for:
- Global frame regions: **Top Status Bar**, **Left Nav**, **Main Pane**, **Bottom Player Bar**
- Information density and keyboard-first affordances (hints, in-pane actions)
- Error presentation patterns (toast/banner vs modal)

Tunez SHOULD provide the following screens/views (as shown in the mockups), subject to Provider capabilities:
- Splash / Loading
- Now Playing
- Search
- Library (Browse)
- Playlists
- Queue
- Lyrics
- Config (including Provider/Profile, Cache/Offline, Scrobbling sections)
- Help / keybindings overlay

### 5.1 Visual layout
- Top bar: provider/profile, connection status, contextual search
- Left sidebar: tabs + provider selector
- Main pane: lists (tracks/albums/playlists) or lyrics/details
- Bottom player bar: track info, progress/time, volume, shuffle/repeat indicators

**Terminal sizing**
- Tunez MUST render a usable UI at 80x24.
- If the terminal is smaller than the minimum, Tunez MUST degrade gracefully (e.g., hide non-essential panels like the visualizer/sidebar) and/or show a clear message indicating the minimum recommended size.

**Terminal encoding**
- Tunez MUST work correctly in UTF-8 terminals.
- Tunez SHOULD provide ASCII-safe fallbacks for decorative box-drawing characters where feasible.

### 5.2 Input model
**MUST**
- Keyboard-first navigation (vim-ish defaults; rebindable):
  - `j/k` navigate
  - `Enter` open/play
  - `Space` play/pause
  - `n/p` next/prev
  - `/` search
  - `q` back/close
  - `?` help overlay

**Keybinding customization**
- Keybindings MUST be configurable (via TOML config).
- The Help overlay SHOULD reflect the currently active keybindings rather than hard-coded defaults.

### 5.2.1 In-app Help (Markdown-driven)
**MUST**
- The Help overlay content MUST be authored in Markdown and shipped with Tunez (e.g., embedded into the `tunez` binary at build time) so it is available offline.
- Tunez MUST render the Help Markdown content in the TUI for user viewing (a limited Markdown subset is acceptable in Phase 1 as long as the content remains readable).

**V2 (Not Phase 1)**
- Online/hosted documentation (“online help”) and any in-app browsing of a docs website.

### 5.3 Color + themes
**MUST**
- 24-bit color support where terminal supports it
- sensible default theme

**SHOULD**
- Theme packs and runtime switching

### 5.4 Animations
**MUST**
- Smooth progress bar updates (adaptive FPS)
- Loading spinners
- Spectrum analyzer animation while audio plays

**SHOULD**
- Target a responsive render cadence (e.g., 30–60 FPS on typical hardware) while remaining adaptive on slower terminals.

### 5.5 Accessibility
**MUST**
- Monochrome fallback
- No emoji required for meaning

**SHOULD**
- Maintain full keyboard operability for all core actions (no mouse required), and keep keybindings rebindable.
- Avoid meaning conveyed by color alone; important states (error/offline/scrobble status) SHOULD also have text labels or icons with monochrome equivalents.
- Prefer high-contrast theme defaults and ensure key UI elements (selection highlight, progress, warnings) remain distinguishable in common terminal color schemes.
- Strive for screen-reader-friendliness within terminal constraints by keeping status text concise, avoiding rapidly changing full-screen redraws for purely decorative elements, and providing non-animated text alternatives where practical.

---

## 6. Non-Functional Requirements

### 6.1 Performance
- Startup: < 1s to show the UI shell (e.g., Splash/Loading) on typical hardware
- UI remains responsive while streaming/decoding
- Visualization degrades gracefully on slow terminals

### 6.2 Reliability
- No panics on normal usage
- Handle audio device changes gracefully

### 6.3 Security
- TLS by default where applicable
- Secrets in OS keyring; file fallback only if explicitly allowed

### 6.4 Portability / distribution
- `cargo install tunez`
- Prebuilt binaries for Linux/macOS/Windows (GitHub Releases)

### 6.5 Privacy
**MUST**
- Scrobbling and any external telemetry MUST be opt-in per provider/profile (disabled unless explicitly enabled/configured).
- Tunez MUST minimize data: only send the fields required by the selected Scrobbler backend (e.g., track identity, timestamps, durations/played duration).
- Tunez MUST provide an opt-out mechanism by allowing the user to disable scrobbling per provider/profile and MUST stop sending new events immediately.
- Locally persisted scrobble retry queues and logs MUST be bounded (by age and/or size) and MUST NOT store passwords.

**SHOULD**
- Provide a way to purge local scrobble queues (and other cached telemetry/state) without uninstalling Tunez.
- Avoid including full URLs with tokens or personally identifying info in logs; redact sensitive values where feasible.

### 6.6 Logging & diagnostics
**MUST**
- Tunez MUST support configurable log verbosity (e.g., via config and/or `--log-level`).
- Log files MUST be bounded via rotation and/or retention limits to avoid unbounded disk growth.

**SHOULD**
- Redact sensitive values (tokens, URLs with embedded credentials) in logs where feasible.

### 6.7 Dependencies & licensing
**MUST**
- Tunez MUST track third-party dependencies and their licenses (e.g., via a generated dependency/license report) and ensure redistribution remains compatible with the project’s license.

### 6.8 Backward compatibility
**MUST**
- Tunez MUST preserve backward compatibility for user data/config where feasible (config/queue/scrobble retry queues), using migrations rather than breaking changes.

**SHOULD**
- Providers/Scrobblers SHOULD have a clear deprecation path for changed behaviors (warn + migrate) rather than silent behavior changes.

### 6.9 Documentation (Developer onboarding)
**MUST**
- The root `README.md` MUST be kept accurate and updated as development progresses.
- The root `README.md` MUST contain the information a developer needs to onboard, including at minimum:
  - What Tunez is (one-paragraph overview + status)
  - How to build/run/test/format (exact commands) once code exists
  - Prerequisites (toolchain + platform notes)
  - Where to find the canonical requirements and UI reference docs
  - A high-level repo/workspace layout (current and/or planned)
  - How to add a Provider/Scrobbler at a high level (when applicable)
  - License and basic contributing guidance
- Phase 1 user-facing help MUST be provided via the in-app Help overlay (offline, Markdown-driven); online/hosted help is deferred to V2.
- Documentation changes MUST be included in the Definition of Done for any phase that changes:
  - CLI commands/flags, configuration schema, workspace layout, Providers/Scrobblers, or user-facing behavior.

**SHOULD**
- The root `README.md` SHOULD match common “modern GitHub OSS” conventions (clear headings, quick links, concise sections, and a polished presentation) while remaining truthful.
- Longer-form docs SHOULD live under `docs/` with the `README.md` acting as an index.

---

## 7. Technical Approach (Rust Stack)

### 7.1 TUI stack
- **ratatui** + **crossterm**
- Central `AppState` + message/event bus model

### 7.2 Async + HTTP
- **tokio**
- **reqwest** + **serde**

### 7.3 Audio playback pipeline
- **symphonia** (decode)
- **cpal** (output)
- Optional: **rodio** (if it supports needed hooks)

#### 7.3.1 Audio format support (Phase 1)
**MUST**
- Tunez MUST support decoding and playback for: MP3, AAC/MP4 (M4A), FLAC, and WAV.

**SHOULD**
- Support Ogg Vorbis and Opus.
- Handle both local files (filesystem provider) and HTTP streams (remote providers) for supported formats.

### 7.4 Visualization
- Tap decoded PCM frames before output
- Ring buffer between audio task and UI task
- FFT via **rustfft**
- Modes: spectrum bars, oscilloscope, VU fallback

### 7.5 CLI + configuration
- **clap** for CLI
- **directories** for config/data paths
- **toml** (+ **serde**) for TOML configuration files
- `tracing` for logs

#### 7.5.1 Configuration management (Phase 1)
**MUST**
- Tunez MUST validate configuration on startup and on `tunez config edit` save/apply.
- Validation failures MUST be surfaced as actionable messages (which key/value is invalid and what is expected) and MUST not expose secrets.
- Tunez MUST use secure defaults: do not enable scrobbling/telemetry unless explicitly configured; do not write secrets to config files.

**SHOULD**
- Support config schema evolution via an explicit `config_version` field and a documented migration strategy.
- When migrating config formats/fields, Tunez SHOULD back up the prior config before writing an updated version.
- Unknown config keys SHOULD be preserved when rewriting config (best-effort) and/or warned about rather than silently dropped.
- Document the configuration schema with at least one example `config.toml` layout.

---

## 8. High-Level Architecture

### 8.1 Workspace layout (recommended)
- `tunez-core/` — domain types, Provider traits, errors
- `tunez-ui/` — ratatui UI, themes, keybindings
- `tunez-player/` — queue + playback state machine
- `tunez-audio/` — stream reader, decoder, output, buffering
- `tunez-viz/` — spectrum/waveform computation
- `tunez-cli/` — CLI parsing and command dispatch
- `providers/` — built-in Provider crates (Phase 1)
  - `provider-remote-example/` (e.g., a server API provider)
  - `provider-filesystem/`
  - future: `provider-jellyfin/`, `provider-subsonic/`

### 8.2 Data flow (provider)
1. UI issues intent (search/browse)
2. Selected Provider executes request (local and/or network)
3. UI updates incrementally (no UI thread blocking)

### 8.3 Data flow (playback)
1. Provider returns a stream URL/handle for a track
2. `audio::stream_reader` pulls bytes
3. `audio::decoder` converts to PCM frames
4. `audio::output` writes to device
5. `viz` computes FFT buckets from PCM frames
6. `ui` renders current state at adaptive FPS

---

## 9. Command Line Interface (CLI) Requirements

### 9.1 Core commands
**MUST**
- `tunez` (launch TUI)
- `tunez play [selectors]` (resolve + launch TUI + autoplay)
- `tunez providers list`

**SHOULD**
- `tunez config edit` (open in $EDITOR)
- `tunez auth login|status|logout` (provider-specific)

### 9.2 Selector options for `play`
**MUST**
- `--provider <id>`
- `--artist <name>`
- `--album <name>`
- `--track <name>`
- `--playlist <name>`
- `--id <provider-specific-id>`
- `-p, --autoplay`

**Selector behavior**
- If multiple selectors are provided, Tunez MUST either treat them as compatible filters (e.g., `--artist` + `--album`) or fail fast with a clear error describing the unsupported combination.
- `--id` MUST take precedence over other selectors.

---

## 10. Definition of Done (Phase 1)

### 10.1 MVP acceptance criteria
- Provider selection works (config + CLI)
- At least 2 built-in Providers:
  - remote API provider (example)
  - local filesystem provider
- Search tracks, pick one, it plays
- Scrobbling works (at least one built-in reference Scrobbler enabled and sending play events to a real backend; e.g., Melodee.API)
- Playlists: list and search playlists (where supported by Provider)
- Queue view works (add/remove/reorder locally)
- Basic controls + progress/volume
- Working spectrum visualization (low-res bars is fine)
- Works on Linux/macOS/Windows terminals

### 10.2 Quality gates
- `cargo fmt`, `cargo clippy -D warnings`, `cargo test`
- No panics on normal usage
- Audio does not stutter on normal network conditions (best effort)

### 10.3 Testing strategy (Phase 1)
Tunez MUST include automated tests at multiple layers to protect core behavior (queue, playback state, provider/scrobbler integration) while keeping tests fast, deterministic, and cross-platform.

**Goals**
- Catch regressions in **queue management**, **playback state**, **provider behavior**, and **scrobbling telemetry**.
- Keep the majority of tests **fast and deterministic** (no real network, no real audio device required).
- Make it easy for new Providers/Scrobblers to adopt a **shared contract test suite**.

#### 10.3.1 Unit tests (core logic)
**MUST**
- Use Rust’s built-in test harness (`#[test]`) for pure logic:
  - Queue operations: add/remove/reorder/clear/shuffle
  - Playback state machine: Stopped/Playing/Paused/Buffering/Error transitions
  - CLI intent parsing/matching logic (e.g., `--artist/--album/--track`)
  - Config parsing/defaults/migrations
  - Capability gating (UI actions enabled/disabled based on provider capabilities)
- Avoid real I/O (network, filesystem, audio device) in unit tests.

**SHOULD**
- Provide fixtures/builders for creating tracks, albums, playlists, and playback timelines.

#### 10.3.2 Property-based tests (invariants)
**SHOULD**
- Use property-based testing for invariants that are hard to exhaust manually (e.g., `proptest`):
  - Queue membership is preserved under shuffle
  - Reorder operations keep indices valid and preserve all items
  - Removing items reduces size exactly and never panics
  - Playback progress is monotonic (never decreases) given valid input sequences

#### 10.3.3 Integration tests (CLI + app wiring)
**MUST**
- Add `tests/` integration tests to validate:
  - `tunez play ...` builds the expected playback intent
  - Provider/profile selection works from CLI and config
  - Error handling (invalid track / stream failure) results in log + user message + skip behavior
- Integration tests MUST run without a real backend by using mocks/stubs.

**SHOULD**
- Use `assert_cmd` to run the binary and verify output/exit codes.
- Use `predicates` for clean output assertions.

#### 10.3.4 Provider contract tests (shared suite)
Tunez MUST provide a reusable contract test suite so each built-in Provider can be validated consistently.

**MUST**
- Implement a `ProviderTestSuite` in a shared crate (e.g., `tunez-testkit` or `tunez-core`).
- Each Provider crate MUST run the suite against its implementation.
- Minimum Provider contract (MVP):
  - `search_tracks(...)` returns stable IDs for selection
  - `get_stream_url(track_id)` returns a usable URL string
  - Playlist support (when provider capability indicates it):
    - `list_playlists(...)` and `search_playlists(...)` behave consistently

**SHOULD**
- Include tests verifying capability flags match behavior (e.g., if playlists unsupported, methods return `NotSupported`).

#### 10.3.5 Scrobbler contract tests (telemetry + emissions)
Tunez MUST test scrobbling as a modular component.

**MUST**
- Provide a `ScrobblerTestSuite` that feeds playback telemetry (track, duration, played duration, state transitions) and asserts emitted “requests/events” via a mock transport.
- The built-in reference Scrobbler(s) MUST be covered by tests validating:
  - Correct event shape (Now Playing / Played or equivalent)
  - Retry/batch behavior on transient failures
  - Errors do not interrupt playback (logged + UI indicator only)

**SHOULD**
- Use a local mock HTTP server (e.g., `wiremock`) rather than real network calls.

#### 10.3.6 UI testing (state-to-view)
Because Tunez is a TUI, most UI behavior SHOULD be tested as state transformations and view-model outcomes rather than pixel-perfect frame comparisons.

**SHOULD**
- Structure UI so that key handlers map `Input -> Action -> State change`.
- Test:
  - Keybinding mappings
  - View selection logic (tabs)
  - Enable/disable behavior based on capabilities (e.g., offline download actions hidden)

**NICE-TO-HAVE**
- Add snapshot/frame tests using a test backend if the TUI framework supports it, but do not rely on these as the primary testing method.

#### Tooling & quality gates
**MUST**
- CI runs:
  - `cargo fmt --check`
  - `cargo clippy -- -D warnings`
  - `cargo test`
- Tests MUST pass on Linux/macOS/Windows CI runners.

**SHOULD**
- Use `cargo nextest` for faster/more readable test output (optional).
- Maintain code coverage reporting (tooling choice left to implementation).

---

## 11. Roadmap (post-Phase 1, optional)

### Phase 2 — External plugins (optional)
- [x] Add a plugin host (exec-based or dylib-based) that *adapts* plugins to the same Provider interface.
- [x] Keep Phase 1 providers working unchanged.

### Phase 3 — Fancy extras
- [x] More visualization modes
- [x] Theme editor
- [x] Scrobbling/play events where supported
- [x] Better caching and offline modes

---

## 12. Decisions (Phase 1)

### 12.1 Provider streaming contract
- Providers return **stream URLs only** (no provider-proxied streaming in Phase 1).
- Stream URLs MUST be usable by the Tunez player. At minimum, Tunez MUST support `https://` and local file paths; Providers SHOULD prefer `https://` for remote streams.

### 12.2 Offline download support
- Providers expose a capability flag: `supports_offline_download: bool`.
- Tunez treats offline downloading as a **Provider/user concern** (rights/DRM/etc. are not Tunez’s responsibility).
- If a Provider sets `supports_offline_download = true`, Tunez MAY expose offline features for that Provider.
- Offline download behavior is configured by the user:
  - **Download location** is a configuration option.
  - **Cache eviction policy** is a configuration option (size/age/TTL, etc.).
- If `supports_offline_download = false` (e.g., Melodee API Provider), Tunez MUST hide/disable offline UI/actions.

**Offline download UX/flow (Phase 1)**
- When the user initiates an offline download, Tunez MUST show an in-UI confirmation/prompt if the action is potentially large (e.g., playlist/album download), and MUST provide a clear way to cancel.
- Tunez MUST present download state feedback (queued/downloading/succeeded/failed) without blocking playback.
- Download failures MUST be non-fatal and surfaced as a clear UI message; Tunez SHOULD allow retry.

**Cache eviction configuration and enforcement (Phase 1)**
- Eviction policy MUST be configurable in TOML and enforced by Tunez (best-effort) at minimum on startup and after new downloads.
- Eviction MUST only delete files Tunez created/owns in its managed download/cache directory; it MUST NOT delete unrelated user files.
- Eviction SHOULD use a predictable strategy (e.g., LRU by access time or oldest-first) and SHOULD log what was removed.

### 12.3 Minimum Provider MVP capability
- A Provider is considered “MVP-capable” if Tunez can:
  1) **search tracks**, and
  2) obtain a **track stream URL** for playback.

For all other operations, an MVP-capable Provider MAY return `ProviderError::NotSupported`.

### 12.4 Scrobbling semantics (MVP)
- Tunez does **not** hard-code scrobbling thresholds in Phase 1.
- Each Scrobbler implementation defines its own reporting rules (e.g., “Now Playing”, “Played/submission”, periodic updates) based on the playback telemetry Tunez provides (track, duration, played duration, state transitions).

---

## 13. Remaining Open Questions
1. (Resolved) Default scrobbling telemetry tick cadence is **1 second**; reuse the same 1-second UI progress tick to drive Scrobbler updates.
2. Should Tunez also offer a low-frequency “heartbeat” callback for long tracks (e.g., every 10s) to reduce work for Scrobblers that don’t need per-second updates?

---

## 14. Appendix

### 14.1 Example `config.toml` (illustrative)
This is a non-normative example showing the intended shape of TOML configuration. Exact keys may evolve with `config_version`.

```toml
config_version = 1

default_provider = "melodee"
profile = "home"

# Scrobbling is opt-in and disabled unless explicitly enabled.
default_scrobbler = "melodee-scrobbler"

[providers.filesystem]
library_root = "/mnt/music"

[providers.melodee]
base_url = "https://music.example.com"
user = "steven@example.com"

[scrobbling]
enabled = false

[scrobbling.providers.melodee]
enabled = true

[logging]
log_level = "info"

[ui]
theme = "AfterDark"

[ui.keybindings]
play_pause = "Space"
next = "n"
previous = "p"
search = "/"
help = "?"

[cache]
# Provider-gated: only applies when offline download is supported.
enabled = false
download_location = "/mnt/music/.tunez-cache"
max_size_gb = 20
eviction_policy = "lru"
ttl_days = 14
```
