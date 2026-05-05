# Focus Command Design

## Overview

Add CLI commands for TickTick Focus (Pomodoro/正计时) feature, following the existing layered architecture (domain → client → app → cli → output).

## API Surface (from developer.dida365.com)

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/open/v1/focus/{focusId}` | Get focus by ID |
| GET | `/open/v1/focus?startDate={ISO}&endDate={ISO}` | List focus sessions by time range |
| POST | `/open/v1/focus` | Start a new focus session |
| POST | `/open/v1/focus/{focusId}` | Stop an active focus session |

### Focus Data Model

```
id          string   Focus ID
mode        int      1=正计时, 2=番茄钟
status      int      0=进行中, 1=已完成
projectId   string   Associated project ID
taskId      string   Associated task ID
title       string   Focus session title
content     string   Description
startDate   string   ISO 8601 datetime
timezone    string   Time zone
abandonReason string Reason for abandonment
tags        []string Tags
creators    []string Creator IDs
sortOrder   int      Sort order
```

### Query Parameters (List)

- `startDate` — required, ISO 8601 datetime (e.g. `2021-09-01T00:00:00+08:00`)
- `endDate` — required, ISO 8601 datetime

## Proposed Commands (Scheme A — confirmed)

### `tick focus ls`

List focus sessions in a time range. Defaults to last 7 days.

```
tick focus ls [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--project NAME] [--json]
```

Flags:
- `--from` — start date (default: 7 days ago)
- `--to` — end date (default: today)
- `--project` — filter by project name/ID
- `--json` — JSON output
- `--output` — `table` or `json`

Output (table): `ID | TITLE | PROJECT | START | END | MODE | STATUS`

### `tick focus get <focus-id>`

Get a single focus session by ID.

```
tick focus get <focus-id> [--json]
```

### `tick focus start <title>`

Start a new focus session.

```
tick focus start <title> [--content TEXT] [--project NAME] [--mode 1|2] [--task ID] [--start "YYYY-MM-DD hh:mm"] [--json]
```

Flags:
- `--content` — session description
- `--project` — associate with project (falls back to `task.default_project` config)
- `--mode` — `1` for 正计时 (default), `2` for 番茄钟
- `--task` — associate with a specific task ID
- `--start` — override start time (default: now)

If `--project` is omitted, falls back to interactive project selection (same pattern as `task add`).

### `tick focus stop <focus-id>`

Stop an active focus session.

```
tick focus stop <focus-id>
```

Prints: `Stopped`

## Architecture

### New Files

| Layer | File | Responsibility |
|-------|------|----------------|
| Domain | `internal/domain/focus.go` | `Focus` struct, `FocusMode` type/constants, `FocusStatus` type/constants |
| Client | `internal/ticktick/focus.go` | DTOs (`focusDTO`, `focusListResponse`), `GetFocus`, `ListFocus`, `StartFocus`, `StopFocus` |
| App | `internal/app/focus.go` | `FocusAPI` interface, `FocusApp` struct, `ListFocusInput`, `StartFocusInput`, `Get`, `List`, `Start`, `Stop` |
| CLI | `internal/cli/focus.go` | `NewFocusCommand`, `FocusResolver`, `ls`/`get`/`start`/`stop` subcommands |
| Output | `internal/output/table.go` | `PrintFocusTable` function |
| Tests | `internal/cli/focus_test.go` | Stub implementations and command tests |
| Tests | `internal/app/focus_test.go` | App layer tests |

### Modified Files

| File | Change |
|------|--------|
| `internal/cli/root.go` | Add `FocusResolver` to `RootOptions`, wire `NewFocusCommand` |
| `cmd/tick/main.go` | Add `FocusResolver` closure in `buildRuntime` |

### Design Decisions

1. **Time range defaults**: `ls` defaults to last 7 days. This matches typical usage patterns where users want to review recent focus history.
2. **Project resolution**: Uses the same `ResolveProject` helper as tasks, with fallback to config and interactive selection.
3. **Mode defaults to 1 (正计时)**: More commonly used than Pomodoro for ad-hoc sessions.
4. **No `update` or `delete` commands**: The API does not expose these operations. Only start/stop are supported.
5. **Date parsing**: Uses `domain.ParseUserTime` (same as task commands) for `--start` flag.
6. **Table columns**: `ID`, `TITLE`, `PROJECT`, `START`, `END`, `MODE`, `STATUS` — omitting `content`, `tags`, `creators` to keep the table readable. Full details available via `get` or `--json`.

### Error Handling

- `ReferenceError` for ambiguous/missing project references (same as task commands).
- `RemoteError` for API failures (handled by client layer).
- Invalid `--mode` values validated at CLI layer before app call.
