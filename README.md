<p align="center">
  <img src="assets/logo.png" alt="Aether logo" width="180">
</p>

# Aether

[Русская версия](README.ru.md)

A minimal terminal radio player built with Go, Bubble Tea, Lip Gloss, and mpv.

![Aether TUI screenshot](assets/screenshot.png)

## Features

- Compact TUI for internet radio playback.
- mpv-based audio playback controlled through IPC.
- Radio Browser station search.
- Stream state indicator: `ON AIR`, `TUNING`, `BUFFERING`, `NO AUDIO`, `PAUSED`.
- Listening history with a configurable retention limit.
- Station favorites and separate favorite tracks.
- Desktop notifications via `notify-send`.
- Environment diagnostics via `aether doctor`.

## Requirements

Required:

- Go
- mpv

On Arch/CachyOS:

```bash
sudo pacman -S --needed go mpv
```

Optional:

- `mpv-mpris` for desktop/media-key integration.
- `notify-send` for desktop notifications.
- `wl-copy`, `xclip`, or `xsel` for clipboard support.

## Run

```bash
go run ./cmd/aether
```

Install into your Go bin directory:

```bash
make install
aether
```

Make sure your Go bin directory is in `PATH`, usually:

```bash
export PATH="$PATH:$HOME/go/bin"
```

## Makefile

```bash
make run      # Start TUI via go run
make doctor   # Run diagnostics
make test     # Run tests
make build    # Build ./bin/aether
make install  # Install aether into Go bin dir
make clean    # Remove ./bin
```

## CLI

```bash
aether                    # Start TUI
aether doctor             # Run diagnostics
aether config path        # Print stations config path
aether config app-path    # Print app config path
aether history path       # Print history JSONL path
aether history list 20    # Print recent history
aether favorites path     # Print favorite tracks JSONL path
aether favorites list     # Print favorite tracks
aether search phonk       # Search stations via Radio Browser
aether version            # Print version
```

## Station search

Press `/` in the TUI to open station search.

Search controls:

- type text — enter query;
- `Enter` — search / preview selected result;
- `↑/↓` — move selection;
- `←/→` — pages;
- `Tab` — add selected station;
- `Ctrl+C` — close search.

## Configuration

Aether follows the XDG Base Directory conventions.

Created on first run:

```text
~/.config/aether/stations.toml
~/.config/aether/config.toml
```

`stations.toml` stores radio stations. The initial station set includes BadRadio, Lofi Radio, and Radio Paradise Main Mix FLAC.

Example station:

```toml
[[station]]
name = "BadRadio"
url = "https://s2.radio.co/s2b2b68744/listen"
provider = "generic"
```

`provider` is optional. The default `generic` provider reads metadata from mpv / ICY stream metadata.

`config.toml` stores app settings:

```toml
[player]
volume = 70
max_volume = 200
volume_step = 5

[search]
page_size = 8

[history]
max_entries = 200

[notifications]
enabled = true
```

UI tabs and hotkeys are also configurable through `ui.*` and `keys.*` sections. Example:

```toml
[keys.global]
pause = [" "]
```

## User data

History and favorite tracks are stored under the XDG data directory:

```text
~/.local/share/aether/history.jsonl
~/.local/share/aether/favorite_tracks.jsonl
```

If `XDG_CONFIG_HOME` or `XDG_DATA_HOME` are set, Aether uses those paths instead of the default `~/.config` and `~/.local/share` locations.

## History

If a stream provides metadata, tracks are automatically written to `history.jsonl`.

Consecutive duplicates are skipped.

By default, Aether keeps the latest `200` history entries:

```toml
[history]
max_entries = 200
```

Old entries beyond this limit are trimmed on startup.

## Favorites

Aether has two separate favorite concepts:

- Station favorites:
  - toggled with `F`;
  - stored in `stations.toml` as `favorite = true`;
  - displayed with `★` and sorted above regular stations.
- Favorite tracks:
  - added with `B`;
  - stored in `favorite_tracks.jsonl`.

Favorite tracks screen opens with `T`:

- `↑/↓` — navigation;
- `Enter` — copy `Artist — Title` to clipboard;
- `Delete` — delete selected track;
- `Ctrl+D` — clear all tracks, confirm with `Enter`;
- `Ctrl+C` — back / cancel.

## Hotkeys

Default global hotkeys:

- `↑/↓` — select station;
- `Tab` — open / close channels dropdown;
- `Enter` — play selected station;
- `Space` — pause / resume;
- `/` — search stations;
- `+` / `-` — volume;
- `B` — add current track to favorite tracks;
- `F` — toggle station favorite;
- `T` — open favorite tracks;
- `R` — rename selected station when channels dropdown is open;
- `D` — delete selected station with confirmation when channels dropdown is open;
- `4` — Help tab;
- `Ctrl+C` — back / quit.

## Backups

Before rewriting `stations.toml`, Aether creates backups in:

```text
~/.config/aether/backups/
```

Only the latest `3` station config backups are kept.

## Release package

Create a Linux amd64 binary tarball:

```bash
make release
```

The archive is written to:

```text
dist/aether_1.0.0_linux_amd64.tar.gz
```

## Troubleshooting

Run:

```bash
aether doctor
```

It checks:

- `mpv` availability;
- optional notification and clipboard tools;
- config paths;
- stations loading;
- history and favorite tracks paths.

If playback does not start, first check that `mpv` is installed and that the station URL is reachable.

## License

MIT — see [LICENSE](LICENSE).
