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
theme = "rainbow"          # rainbow (default) | mono (v1+) | green (v1+) | ...

[player]
mpv_path = "mpv"
ipc = "auto"              # auto | unix | pipe
initial_volume = 70
cache_secs = 30            # mpv cache-secs
network_timeout_ms = 8000

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
"app.quit" = ["ctrl+c"]
"app.help" = ["?"]
"nav.search" = ["/"]
"player.play_pause" = ["space"]
"player.next" = ["n"]
"player.prev" = ["p"]
"player.seek_forward_small" = ["l"]
"player.seek_back_small" = ["h"]
"player.volume_up" = ["+"]
"player.volume_down" = ["-"]
```

## Validation rules (MVP)
- `active_profile` must exist and be enabled
- mpv must be discoverable (PATH or `mpv_path`)
- Filesystem roots must exist
- Melodee base_url must be valid

## UI theme (requirement)
- The default theme is intended to be very colorful with rainbow-like ANSI effects.
- Additional themes will be implemented later (v1+), including monochromatic and “green terminal” styles.
