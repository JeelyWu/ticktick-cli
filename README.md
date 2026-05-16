# tick - TickTick/Dida365 CLI

[中文](README.zh-CN.md) | **English**

[![Unit Tests](https://github.com/JeelyWu/ticktick-cli/actions/workflows/unit-tests.yml/badge.svg?branch=master)](https://github.com/JeelyWu/ticktick-cli/actions/workflows/unit-tests.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`tick` is an unofficial Go CLI for TickTick / Dida365, including the mainland China Dida365 service known as 滴答清单. It uses the official Open API to bring task capture, project management, habits, focus sessions, and scriptable output into your terminal.

## Why tick?

- Capture tasks quickly from the terminal with compact quick-add syntax.
- Manage TickTick/Dida365 projects, tasks, habits, and focus sessions from one binary.
- Use table output for humans and JSON output for shell scripts, launchers, and automation.
- Log in with an interactive OAuth flow that can capture localhost callbacks automatically.
- Switch between TickTick international and Dida365 mainland China regions with local defaults.

## Install

### GitHub Releases

Download a platform archive from [GitHub Releases](https://github.com/JeelyWu/ticktick-cli/releases) and extract the binary to your `PATH`.

```bash
# Linux / macOS example
tar -xzf tick_0.0.1_linux_amd64.tar.gz
install -m 0755 tick /usr/local/bin/tick
```

### Install script (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | bash
```

### Go install

```bash
go install github.com/jeelywu/ticktick-cli/cmd/tick@latest
```

### Build from source

```bash
make build
```

Binary is written to `bin/tick`.

## Quick Start

### 1. Create a developer app

- TickTick international: `https://developer.ticktick.com/manage`
- Dida365 mainland China: `https://developer.dida365.com/manage`

Save the **Client ID** and **Client Secret**.

### 2. Log in

```bash
tick auth login
```

This starts an interactive OAuth flow. It will prompt for:

- Region (`ticktick` or `dida365`)
- Client ID
- Client Secret

On local machines it automatically captures the localhost callback. On remote/SSH sessions it falls back to manual callback URL paste.

### 3. Verify

```bash
tick auth status
```

## Usage

Use `tick <command> --help` for details on any command.

### Projects

```bash
tick project ls
tick project get Work
tick project add Work
tick project add Notes --kind NOTE
tick project update Work --name "New Name"
tick project rm Work --yes
```

### Tasks

```bash
# List
tick task ls
tick task ls --project Work
tick task ls --today
tick task ls --overdue
tick task ls --status completed
tick task ls --priority 5
tick task ls --from 2026-04-01 --to 2026-04-30
tick task ls --tag urgent

# Get / add / update / delete
tick task get "Write spec"
tick task add "Write spec" --project Work --due 2026-04-20
tick task add "Review" --project Work --due 2026-04-20 --all-day
tick task update "Write spec" --title "Write detailed spec" --due 2026-04-21
tick task done "Write spec"
tick task reopen "Write spec"
tick task rm "Write spec"          # prompts for confirmation
tick task rm "Write spec" --yes    # skip confirmation

# Move
tick task move "Write spec" --to Personal
tick task move "Write spec" --project Work --to Personal
```

### Today

```bash
tick today
tick today --json
```

### Quick add

`tick quick add` parses compact task-entry syntax:

- plain text → title
- `#ProjectName` → project
- `!1`, `!3`, `!5` → priority
- `^YYYY-MM-DD` → due date

```bash
tick quick add "Write spec #Work !5 ^2026-04-10"
tick quick add "Buy milk #Personal ^2026-04-18"
```

If `default_project` is configured, you can omit `#ProjectName`:

```bash
tick config set default_project Work
tick quick add "Prepare launch notes !3 ^2026-04-22"
```

### Habits

```bash
tick habit ls
tick habit get "Read 30 min"
tick habit add "Read 30 min" --goal 30 --unit "min"
tick habit update "Read 30 min" --goal 60
tick habit archive "Read 30 min"
tick habit checkin "Read 30 min" --value 30
tick habit log "Read 30 min"
```

### Focus sessions

```bash
tick focus ls
tick focus ls --type 0                    # pomodoro (default: 1=timer)
tick focus get <focus-id>
tick focus get <focus-id> --type 0        # pomodoro (default: 1=timer)
```

### Configuration

```bash
tick config list
tick config get region
tick config set region ticktick      # or dida365
tick config set output json          # default output format
tick config set default_project Work # default for quick add / task add
```

Switching regions after login requires re-authentication:

```bash
tick auth logout
tick auth login
```

### Output formats and priorities

- Output: `table` (default) or `json`. Use `--json` or `tick config set output json`.
- Priority: `0` none, `1` low, `3` medium, `5` high.

## Disclaimer

`tick` is an independent, unofficial community project. It is not affiliated with, endorsed by, or sponsored by TickTick, Dida365, or Appest.

## License

[MIT](LICENSE)
