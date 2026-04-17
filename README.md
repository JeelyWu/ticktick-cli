# tick

[![Unit Tests](https://github.com/JeelyWu/ticktick-cli/actions/workflows/unit-tests.yml/badge.svg?branch=master)](https://github.com/JeelyWu/ticktick-cli/actions/workflows/unit-tests.yml)

`tick` is a Go-based CLI for TickTick and Dida365 via the official Open API.

## Install

### GitHub Releases

Every tagged release publishes platform-specific archives to GitHub Releases:

- macOS: `darwin/arm64`, `darwin/amd64`
- Linux: `linux/arm64`, `linux/amd64`
- Windows: `windows/amd64`

Download the archive that matches your platform from the releases page and unpack `tick` into a directory on your `PATH`.

### macOS and Linux install script

The repository includes a release installer for Unix-like systems:

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | bash
```

Useful overrides:

```bash
curl -fsSL https://raw.githubusercontent.com/JeelyWu/ticktick-cli/master/scripts/install.sh | \
  VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" bash
```

If you prefer not to pipe to `bash`, download [scripts/install.sh](scripts/install.sh) first and run it locally.

## Prerequisites

1. Create a developer application for your region
   - TickTick international: `https://developer.ticktick.com/manage`
   - Dida365 mainland China: `https://developer.dida365.com/manage`
2. Copy the app's `client_id`, `client_secret`, and redirect URL

## Build

```bash
make build
```

To build release archives locally, install GoReleaser and run:

```bash
make release-check
make release
```

## First-time auth

Choose your service region before login. The default is `ticktick`.

If you switch regions after logging in, run `tick auth logout` and authenticate again so the stored token matches the selected service.

For local development, `tick auth login` will first try to capture the localhost callback automatically. If that fails, it falls back to manual callback URL paste.

TickTick international on your local machine:

```bash
tick config set service.region ticktick
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
tick auth login \
  --client-id YOUR_CLIENT_ID \
  --redirect-url http://localhost:14573/callback
```

Dida365 on your local machine:

```bash
tick config set service.region dida365
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
tick auth login \
  --client-id YOUR_CLIENT_ID \
  --redirect-url http://localhost:14573/callback
```

For a remote machine or SSH session, keep the same command and paste the full callback URL back into the terminal if the browser cannot reach the remote host:

```bash
tick config set service.region dida365
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
tick auth login \
  --client-id YOUR_CLIENT_ID \
  --redirect-url http://localhost:14573/callback
```

The terminal will either complete automatically after the browser callback, or print:

```text
Paste the full callback URL:
```

In manual fallback mode, copy the entire browser address bar after authorization, for example:

```text
http://localhost:14573/callback?code=abc123&state=xyz456
```

## Common commands

```bash
tick auth status
tick version --verbose
tick project ls
tick task ls --project Work
tick task ls --today
tick task ls --project Work --overdue
tick today
tick inbox
tick task add "Write spec" --project Work --due 2026-04-10
tick quick add "Write spec #Work !5 ^2026-04-10"
tick config list
tick config set task.default_project Work
tick config get service.region
```

## Release artifacts

```bash
git tag v0.1.0
git push origin v0.1.0
```

Pushing a `v*` tag triggers [release.yml](.github/workflows/release.yml), which runs tests, builds archives with GoReleaser, generates checksums, and uploads everything to GitHub Releases.

For local dry runs, install GoReleaser from https://goreleaser.com/install/ and use:

```bash
make release-check
make release
```
