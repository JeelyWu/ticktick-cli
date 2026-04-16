# tick

`tick` is a Go-based CLI for TickTick's official Open API.

## Prerequisites

1. Create a TickTick developer application at `https://developer.ticktick.com/manage`
2. Copy the app's `client_id`, `client_secret`, and redirect URL

## Build

```bash
make build
```

## First-time auth

```bash
export TICK_CLIENT_SECRET=YOUR_CLIENT_SECRET
tick auth login \
  --client-id YOUR_CLIENT_ID \
  --redirect-url http://localhost:14573/callback
```

## Common commands

```bash
tick auth status
tick project ls
tick task ls --project Work
tick task ls --today
tick task ls --project Work --overdue
tick today
tick inbox
tick task add "Write spec" --project Work --due 2026-04-10
tick quick add "Write spec #Work !5 ^2026-04-10"
tick config set task.default_project Work
```

## Release artifacts

```bash
make release
```
