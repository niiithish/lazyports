# lazyports

A terminal UI for viewing and managing processes bound to local dev ports — inspired by [lazydocker](https://github.com/jesseduffield/lazydocker).

```bash
lazyports
```

## Features

- **Dev ports only** by default — node, python, rails, docker, etc. No system noise
- Split-pane UI inspired by lazydocker
- Process name, PID, user, uptime, and command line
- Filter with `/`
- Kill with `d` (SIGTERM) or `K` (SIGKILL)
- Restart with `s`
- Auto-refresh every 2 seconds (configurable)

## Install

**One-liner** (builds and installs to `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/niiithish/lazyports/main/install.sh | bash
```

**With Go:**

```bash
go install github.com/niiithish/lazyports@latest
```

**From source:**

```bash
git clone https://github.com/niiithish/lazyports
cd lazyports
make install
```

Make sure `~/.local/bin` is on your `PATH`.

## Usage

```bash
lazyports
lazyports -refresh 5s
lazyports -refresh 0   # disable auto-refresh
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate ports |
| `a` | Toggle dev ports / all TCP listen |
| `/` | Filter |
| `d` / `x` | Kill process (SIGTERM) |
| `K` | Force kill (SIGKILL) |
| `s` | Restart process |
| `r` | Refresh |
| `?` | Help |
| `q` | Quit |

## Requirements

- Linux (uses `ss` and `/proc`)
- Go 1.23+ to build from source

## License

MIT
