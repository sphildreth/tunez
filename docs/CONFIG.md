# Tunez — Configuration

**Last updated:** 2026-01-03

Tunez uses a TOML config file: `config.toml`.

## Example `config.toml`

```toml
config_version = 1
active_profile = "home-files"

[ui]
page_size = 100
no_emoji = false
theme = "rainbow"          # rainbow (default) | mono | green | nocolor

[player]
mpv_path = "mpv"
ipc = "auto"              # auto | unix | pipe
initial_volume = 70
cache_secs = 30            # mpv cache-secs
network_timeout_ms = 8000

[queue]
persist = true             # Persist queue across restarts

[artwork]
enabled = true             # Show album artwork in Now Playing
width = 20                 # Artwork width in characters
cache_days = 30            # Days to cache converted artwork

[scrobble]
enabled = false            # Master switch for all scrobblers

[[scrobblers]]
id = "lastfm"
type = "lastfm"
enabled = true
[scrobblers.settings]
api_key = "your_lastfm_api_key"
api_secret = "your_lastfm_api_secret"
session_key = "your_session_key"

[[scrobblers]]
id = "melodee"
type = "melodee"
enabled = true
[scrobblers.settings]
provider = "melodee"       # Reuse auth from this provider

[[profiles]]
id = "home-files"
name = "Home Files"
provider = "filesystem"
enabled = true

[profiles.settings]
roots = ["/music"]
index_db = "filesystem.sqlite"
scan_on_start = true

[[profiles]]
id = "melodee-home"
name = "Melodee (Home)"
provider = "melodee"
enabled = true

[profiles.settings]
base_url = "https://music.example.com"
username = "steven"
password_env = "TUNEZ_MELODEE_PASSWORD"
page_size = 200
cache_db = "melodee_cache.sqlite"

[keybindings]
play_pause = "space"
next_track = "n"
prev_track = "p"
seek_forward = "l"
seek_backward = "h"
volume_up = "+"
volume_down = "-"
mute = "m"
shuffle = "s"
repeat = "r"
search = "/"
help = "?"
quit = "ctrl+c"
```

## Configuration Sections

### `[ui]`
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `page_size` | int | 100 | Items per page in lists |
| `no_emoji` | bool | false | Disable emoji in UI |
| `theme` | string | "rainbow" | Color theme: rainbow, mono, green, nocolor |

### `[player]`
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `mpv_path` | string | "mpv" | Path to mpv binary |
| `ipc` | string | "auto" | IPC method: auto, unix, pipe |
| `initial_volume` | int | 70 | Starting volume (0-100) |
| `cache_secs` | int | 30 | mpv cache seconds |
| `network_timeout_ms` | int | 8000 | Network timeout in milliseconds |
| `seek_small_seconds` | int | 5 | Small seek step |
| `seek_large_seconds` | int | 30 | Large seek step |
| `volume_step` | int | 5 | Volume adjustment step |

### `[queue]`
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `persist` | bool | true | Save queue across restarts |

### `[artwork]`
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | true | Show artwork in Now Playing |
| `width` | int | 20 | Artwork width in characters (auto-adjusted if too large for terminal) |
| `height` | int | 10 | Artwork height in characters |
| `quality` | string | "medium" | Image quality: low, medium, or high |
| `scale_mode` | string | "fit" | Scaling: fit, fill, or stretch |
| `cache_days` | int | 30 | Days to cache converted artwork |

**Note:** Artwork width is automatically adjusted if it exceeds your terminal width to prevent scrolling. For best results, use values that fit your terminal (e.g., 15-25 width for standard 80-column terminals).

### `[scrobble]`
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | false | Master switch for scrobbling |

### `[[scrobblers]]`
Array of scrobbler configurations.

| Key | Type | Description |
|-----|------|-------------|
| `id` | string | Unique identifier |
| `type` | string | Scrobbler type: lastfm, melodee |
| `enabled` | bool | Enable this scrobbler |
| `settings` | table | Type-specific settings |

**Last.fm settings:**
- `api_key` - Last.fm API key
- `api_secret` - Last.fm API secret
- `session_key` - Authenticated session key

**Melodee settings:**
- `provider` - Provider ID to reuse auth from
- `base_url` - API base URL (if not using provider)
- `token` - Static auth token (if not using provider)

## Themes

| Theme | Description |
|-------|-------------|
| `rainbow` | Colorful default theme with accent colors |
| `mono` | Grayscale theme using white/gray tones |
| `green` | Classic green-on-black terminal aesthetic |
| `nocolor` | Plain text, no ANSI colors (accessibility) |

The `nocolor` theme is automatically selected when the `NO_COLOR` environment variable is set.

## Validation Rules
- `active_profile` must exist and be enabled
- mpv must be discoverable (PATH or `mpv_path`)
- Filesystem roots must exist
- Melodee base_url must be valid URL
- Theme must be one of: rainbow, mono, green, nocolor
