# Tunez

Tunez is a fast, responsive terminal music player written in Go with a Bubble Tea UI and `mpv` playback.

## Status
Tunez is in active development. Expect rough edges and breaking changes while MVP is being built.

## Key features (MVP)
- Browse/search music via built-in providers (Filesystem + Melodee API).
- Enqueue and play tracks with non-blocking UI.
- Playback through `mpv` (started and controlled by Tunez).
- Configurable profiles + keybindings.

## Repository layout
This repository has two top-level concerns:

- `docs/` — product + technical specs (requirements, UX, provider contract)
- `src/` — Go module (`go.mod`) and application source

The Tunez entrypoint is `src/cmd/tunez`.

## Prerequisites

### Required
- **Go 1.22+** (see `src/go.mod`)
- **mpv** installed and available on `PATH` (or configure `player.mpv_path`)

### Optional / provider-specific
- **Filesystem provider**: a directory of music files accessible on this machine
- **Melodee provider**: a reachable Melodee server + credentials (password is read from an env var)

## Quickstart (run locally)

### 1) Install mpv

On Linux (Debian/Ubuntu):

```bash
sudo apt-get update
sudo apt-get install -y mpv
```

On macOS (Homebrew):

```bash
brew install mpv
```

On Arch:

```bash
sudo pacman -S mpv
```

### 2) Create a config file

Tunez reads a TOML config file.

- Default location (Linux/macOS): `~/.config/tunez/config.toml`
- You can also pass `-config /path/to/config.toml`

Start from the example:

```bash
mkdir -p ~/.config/tunez
cp examples/config.example.toml ~/.config/tunez/config.toml
```

Then edit `~/.config/tunez/config.toml` and set at least:
- `active_profile`
- a valid provider profile
- `profiles.settings.roots` (for `filesystem`) pointing at a real folder

Config reference: `docs/CONFIG.md`.

### 3) Run Tunez

The Go module lives in `src/`:

```bash
cd src
go run ./cmd/tunez
```

Useful flags:

```bash
cd src
go run ./cmd/tunez -version
go run ./cmd/tunez -doctor
go run ./cmd/tunez -config ~/.config/tunez/config.toml
```

`-doctor` checks that the config parses, `mpv` is discoverable, and the active provider initializes.

## Testing

```bash
cd src
go test ./...
```

## Build

```bash
cd src
go build -o ./bin/tunez ./cmd/tunez
./bin/tunez -doctor
./bin/tunez
```

## Logs & troubleshooting

### Logs
Tunez writes logs to the user config dir under `tunez/state`.

On Linux this is typically:

- `~/.config/tunez/state/tunez-YYYYMMDD.log`

### Common issues

**"mpv not found"**
- Install `mpv` and ensure it’s on `PATH`, or set `player.mpv_path` to an absolute path.

**Filesystem provider fails validation**
- Ensure `profiles.settings.roots` exists and is readable.

**Melodee provider authentication**
- Set the password env var specified by `profiles.settings.password_env` (example uses `TUNEZ_MELODEE_PASSWORD`).

## Docs (source of truth)

**Start here:**
- **`docs/PHASE_PLAN.md`** — **Comprehensive phase breakdown mapping ALL requirements to deliverable phases**

**Full documentation:**
- `docs/README.md` — Documentation navigation and quick reference
- `docs/PRD.md` — Product requirements and acceptance criteria
- `docs/TUI_UX.md` — Screens, interactions, and keybindings
- `docs/TECH_DESIGN.md` — Architecture + key technical decisions
- `docs/PROVIDERS.md` — Provider interface contract
- `docs/CONFIG.md` — Config schema + profiles + keybindings
- `docs/TEST_STRATEGY.md` — Testing approach
- `docs/SECURITY_PRIVACY.md` — Handling secrets, auth, privacy expectations
- `docs/PROVIDER_FILESYSTEM.md` — Filesystem provider spec
- `docs/PROVIDER_MELODEE_API.md` — Melodee API provider spec
- `docs/DECISIONS.md` — Architectural trade-offs and clarifications

## Contributing (developer workflow)

- Keep the Bubble Tea `Update` loop free of blocking work; use background commands and messages.
- Prefer small changes with tests for non-trivial logic.
- Run `go test ./...` before opening a PR.
