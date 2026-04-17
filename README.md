# tick

[中文](README.zh-CN.md) | **English**

[![Unit Tests](https://github.com/JeelyWu/ticktick-cli/actions/workflows/unit-tests.yml/badge.svg?branch=master)](https://github.com/JeelyWu/ticktick-cli/actions/workflows/unit-tests.yml)

`tick` is a Go CLI for TickTick international and Dida365 via the official Open API.

It covers a practical daily loop:

- authenticate once
- inspect projects and tasks
- add, update, move, and complete tasks
- use quick-add syntax for fast capture
- store a few local defaults so common commands stay short

## Features

- Supports both TickTick international and Dida365 mainland China
- OAuth login flow with automatic localhost callback capture
- Manual callback fallback for SSH sessions and remote machines
- Project commands: `add`, `get`, `ls`, `rm`, `update`
- Task commands: `add`, `get`, `ls`, `update`, `move`, `done`, `rm`
- Convenience commands: `today`, `inbox`, `quick add`
- Configurable local defaults for output format, default project, inbox project, and service region

## Install

### GitHub Releases

Every tagged release publishes platform-specific archives on GitHub Releases:

- macOS: `darwin/arm64`, `darwin/amd64`
- Linux: `linux/arm64`, `linux/amd64`
- Windows: `windows/amd64`

Release page:

```text
https://github.com/JeelyWu/ticktick-cli/releases
```

Each archive currently contains:

- the `tick` executable, or `tick.exe` on Windows
- `README.md`

Install manually by downloading the archive for your platform, extracting it, and moving the executable into a directory on your `PATH`.

Examples:

```bash
tar -xzf tick_0.0.1_linux_amd64.tar.gz
install -m 0755 tick /usr/local/bin/tick
```

```bash
tar -xzf tick_0.0.1_darwin_arm64.tar.gz
install -m 0755 tick "$HOME/.local/bin/tick"
```

On Windows, extract `tick_0.0.1_windows_amd64.zip` and place `tick.exe` somewhere on your `PATH`.

### macOS and Linux install script

The repository includes a release installer for Unix-like systems:

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | bash
```

Install a specific release:

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | \
  VERSION=v0.0.1 bash
```

Install into a custom directory:

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | \
  VERSION=v0.0.1 INSTALL_DIR="$HOME/.local/bin" bash
```

If you prefer not to pipe to `bash`, download [scripts/install.sh](scripts/install.sh) first and run it locally.

### Build from source

Requirements:

- Go, version from [go.mod](go.mod)

Build the local binary:

```bash
make build
```

The binary will be written to `bin/tick`.

## Prerequisites

Before logging in, create a developer application for the service you want to use:

- TickTick international: `https://developer.ticktick.com/manage`
- Dida365 mainland China: `https://developer.dida365.com/manage`

Collect these values from your app:

- `client_id`
- `client_secret`
- redirect URL

For local use, the recommended redirect URL is:

```text
http://localhost:14573/callback
```

`tick` expects the client secret in the environment:

```bash
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
```

## First-Time Setup

### 1. Choose the service region

`tick` defaults to `ticktick`. Switch to `dida365` if you use the mainland China service.

```bash
tick config set service.region ticktick
```

```bash
tick config set service.region dida365
```

Check the current value:

```bash
tick config get service.region
```

If you change regions after you already logged in, clear the stored token and authenticate again:

```bash
tick auth logout
tick auth login --client-id YOUR_CLIENT_ID --redirect-url http://localhost:14573/callback
```

### 2. Run login

TickTick international:

```bash
tick config set service.region ticktick
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
tick auth login \
  --client-id YOUR_CLIENT_ID \
  --redirect-url http://localhost:14573/callback
```

Dida365:

```bash
tick config set service.region dida365
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
tick auth login \
  --client-id YOUR_CLIENT_ID \
  --redirect-url http://localhost:14573/callback
```

### 3. Verify auth state

```bash
tick auth status
```

You can also print version and configured region together:

```bash
tick version --verbose
```

## Login Behavior On Local And Remote Machines

For local development, `tick auth login` first tries to capture the localhost callback automatically.

If the browser cannot reach the callback listener, for example on a remote machine or SSH session, `tick` falls back to manual callback paste and prints:

```text
Paste the full callback URL:
```

In that mode, copy the entire browser address after authorization, for example:

```text
http://localhost:14573/callback?code=abc123&state=xyz456
```

Paste that full URL back into the terminal.

## Common Commands

Top-level commands:

```bash
tick auth --help
tick project --help
tick task --help
tick quick --help
tick config --help
tick today --help
tick inbox --help
tick version --help
```

### Project commands

List projects:

```bash
tick project ls
```

Show one project by exact name or ID:

```bash
tick project get Work
```

Create a project:

```bash
tick project add Work
```

Create a note-type project with a color:

```bash
tick project add Notes --kind NOTE --color '#F18181'
```

### Task listing

List open tasks:

```bash
tick task ls
```

List tasks in one project:

```bash
tick task ls --project Work
```

List overdue tasks:

```bash
tick task ls --project Work --overdue
```

List tasks due today or overdue:

```bash
tick task ls --today
```

Filter by status:

```bash
tick task ls --status completed
```

Filter by priority:

```bash
tick task ls --priority 5
```

Filter by date range:

```bash
tick task ls --from 2026-04-01 --to 2026-04-30
```

Print JSON:

```bash
tick task ls --json
```

Equivalent explicit output flag:

```bash
tick task ls --output json
```

### Task creation and updates

Create a task in a project:

```bash
tick task add "Write spec" --project Work --due 2026-04-20
```

Create an all-day task:

```bash
tick task add "Review roadmap" --project Work --due 2026-04-20 --all-day
```

Create a high-priority task with description and content:

```bash
tick task add "Ship v0.0.2" \
  --project Work \
  --priority 5 \
  --desc "Publish binaries and verify release assets" \
  --content "Double-check the GitHub Release page"
```

Show one task by exact title or ID:

```bash
tick task get "Write spec"
```

Update title and due date:

```bash
tick task update "Write spec" --title "Write detailed spec" --due 2026-04-21
```

Move a task to another project:

```bash
tick task move "Write spec" --to Personal
```

Move a task when the same title exists in more than one project:

```bash
tick task move "Write spec" --project Work --to Personal
```

Mark a task as done:

```bash
tick task done "Write spec"
```

Delete a task:

```bash
tick task rm "Write spec"
```

### Convenience commands

Show tasks due today or overdue:

```bash
tick today
```

Show the configured inbox project:

```bash
tick inbox
```

Use JSON with convenience commands:

```bash
tick today --json
tick inbox --json
```

### Quick add

`tick quick add` parses a compact task-entry format:

- plain text becomes the task title
- `#ProjectName` sets the project
- `!1`, `!3`, `!5` set priority
- `^YYYY-MM-DD` sets the due date

Examples:

```bash
tick quick add "Write spec #Work !5 ^2026-04-10"
```

```bash
tick quick add "Buy milk #Personal ^2026-04-18"
```

If `task.default_project` is configured, quick add can omit `#ProjectName`:

```bash
tick config set task.default_project Work
tick quick add "Prepare launch notes !3 ^2026-04-22"
```

### Configuration

Show the full local config:

```bash
tick config list
```

Read one value:

```bash
tick config get service.region
```

Set the default task output format:

```bash
tick config set output.default json
```

Set a default project for `quick add`:

```bash
tick config set task.default_project Work
```

Set the inbox project ID used by `tick inbox`:

```bash
tick config set task.inbox_project_id YOUR_INBOX_PROJECT_ID
```

## Output And Priorities

Supported output formats:

- `table`
- `json`

You can set the default with:

```bash
tick config set output.default table
```

Priority values:

- `0` for none
- `1` for low
- `3` for medium
- `5` for high

## Release Process

Validate the GoReleaser config locally:

```bash
make release-check
```

Build snapshot release archives locally:

```bash
make release
```

Publish a real release:

```bash
git tag v0.0.2
git push origin v0.0.2
```

Pushing a `v*` tag triggers [release.yml](.github/workflows/release.yml), which runs tests, builds archives with GoReleaser, generates checksums, and uploads everything to GitHub Releases.
