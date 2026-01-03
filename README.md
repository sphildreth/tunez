<p align="center">
  <img src="graphics/tunez-logo.png" alt="Tunez Logo" width="120" />
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey?style=flat" alt="Platform" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat" alt="License" />
  <img src="https://img.shields.io/badge/Status-Active%20Development-blue?style=flat" alt="Status" />
</p>

<h1 align="center">Tunez</h1>

<p align="center">
  <strong>A fast, beautiful terminal music player</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#quickstart">Quickstart</a> •
  <a href="#keybindings">Keybindings</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#contributing">Contributing</a>
</p>
---

Tunez is a **keyboard-driven terminal music player** written in Go. It features a responsive [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI, [mpv](https://mpv.io/) for high-quality audio playback, and support for multiple music sources through a flexible provider system.

## Features

- 🎵 **Beautiful TUI** — Rainbow-colored interface with smooth navigation
- ⚡ **Responsive** — Non-blocking UI, all I/O happens in the background
- 🎧 **High-quality playback** — Powered by mpv with gapless playback support
- 🔀 **Queue management** — Add, remove, reorder, shuffle, and repeat
- 🔍 **Fast search** — Search across tracks, albums, and artists
- 📚 **Multiple providers** — Local filesystem or Melodee API server
- ⚙️ **Configurable** — Custom keybindings, themes, and profiles
- ♿ **Accessible** — NO_COLOR support, works at 80×24

## Installation

### Prerequisites

- **Go 1.22+**
- **mpv** media player

#### Install mpv

```bash
# Debian/Ubuntu
sudo apt-get install -y mpv

# macOS (Homebrew)
brew install mpv

# Arch Linux
sudo pacman -S mpv

# Fedora
sudo dnf install mpv
```

### Build from source

```bash
git clone https://github.com/yourusername/tunez.git
cd tunez/src
go build -o tunez ./cmd/tunez
./tunez --version
```

## Quickstart

### 1. Create a config file

```bash
mkdir -p ~/.config/tunez
cp examples/config.example.toml ~/.config/tunez/config.toml
```

### 2. Edit the config

Point Tunez at your music library:

```toml
active_profile = "local"

[[profiles]]
id = "local"
name = "My Music"
provider = "filesystem"
enabled = true

[profiles.settings]
roots = ["/home/you/Music"]
```

### 3. Run Tunez

```bash
./tunez
```

Or run the doctor to verify your setup:

```bash
./tunez -doctor
```

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select / Play |
| `Tab` | Next screen |
| `Shift+Tab` | Previous screen |
| `/` | Search |
| `?` | Help |
| `Ctrl+C` | Quit |

### Playback

| Key | Action |
|-----|--------|
| `Space` | Play / Pause |
| `n` | Next track |
| `p` | Previous track |
| `h` / `l` | Seek -5s / +5s |
| `H` / `L` | Seek -30s / +30s |
| `-` / `+` | Volume down / up |
| `m` | Mute |
| `s` | Toggle shuffle |
| `r` | Cycle repeat (off → all → one) |

### Queue

| Key | Action |
|-----|--------|
| `a` | Add to queue |
| `A` / `P` | Play next |
| `x` | Remove from queue |
| `u` / `d` | Move up / down |
| `C` | Clear queue |

## Configuration

Tunez uses a TOML configuration file located at:
- **Linux/macOS**: `~/.config/tunez/config.toml`
- **Windows**: `%APPDATA%\Tunez\config.toml`

### Example config

```toml
config_version = 1
active_profile = "local"

[ui]
theme = "rainbow"
page_size = 100
no_emoji = false

[player]
mpv_path = "mpv"
initial_volume = 70
seek_small_seconds = 5
seek_large_seconds = 30
volume_step = 5

[[profiles]]
id = "local"
name = "Local Music"
provider = "filesystem"
enabled = true

[profiles.settings]
roots = ["/home/you/Music", "/mnt/nas/music"]
scan_on_start = false

[[profiles]]
id = "melodee"
name = "Melodee Server"
provider = "melodee"
enabled = true

[profiles.settings]
base_url = "https://music.example.com"
username = "user"
password_env = "TUNEZ_MELODEE_PASSWORD"
```

See [docs/CONFIG.md](docs/CONFIG.md) for the full configuration reference.

## Themes & Accessibility

Tunez respects the `NO_COLOR` environment variable for accessibility:

```bash
NO_COLOR=1 ./tunez
```

You can also disable emoji in the config:

```toml
[ui]
no_emoji = true
```

## Documentation

| Document | Description |
|----------|-------------|
| [PHASE_PLAN.md](docs/PHASE_PLAN.md) | Development roadmap and phase breakdown |
| [PRD.md](docs/PRD.md) | Product requirements |
| [TUI_UX.md](docs/TUI_UX.md) | Screen specifications and interactions |
| [CONFIG.md](docs/CONFIG.md) | Configuration reference |
| [PROVIDERS.md](docs/PROVIDERS.md) | Provider interface contract |
| [TECH_DESIGN.md](docs/TECH_DESIGN.md) | Architecture decisions |

## Troubleshooting

### Logs

Tunez writes logs to `~/.config/tunez/state/tunez-YYYYMMDD.log`

### Common issues

**"mpv not found"**
```bash
# Verify mpv is installed
mpv --version

# Or set an explicit path in config
[player]
mpv_path = "/usr/local/bin/mpv"
```

**Filesystem provider fails**
- Ensure `roots` paths exist and are readable
- Check log file for detailed errors

**Melodee authentication fails**
- Set the password environment variable:
  ```bash
  export TUNEZ_MELODEE_PASSWORD="your-password"
  ./tunez
  ```

## Contributing

Contributions are welcome! Please read the following before submitting:

1. Keep the Bubble Tea `Update` loop free of blocking I/O
2. Add tests for non-trivial logic
3. Run `go test ./...` before submitting
4. Follow existing code style (`go fmt`)

```bash
cd src
go test ./...      # Run tests
go fmt ./...       # Format code
go vet ./...       # Check for issues
```

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<p align="center">
  Made with 🎵 and Go
</p>
