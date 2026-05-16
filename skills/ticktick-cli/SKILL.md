---
name: ticktick-cli
description: Use when the user wants to manage TickTick or Dida365 tasks, projects, habits, focus sessions, or local tick CLI configuration through the `tick` command-line tool.
---

# TickTick CLI

Use the local `tick` CLI to manage TickTick international or Dida365 mainland China data. This skill is for operating the CLI on behalf of the user, not for changing the CLI source code.

## Preflight

1. Prefer `tick` from `PATH`.
2. If `tick` is unavailable and the current directory is the source repository, use `bin/tick` if it exists or run `make build`.
3. Check authentication before user-data operations:

```bash
tick auth status
```

If authentication is missing or expired, ask the user to run:

```bash
tick auth login
```

Do not ask the user to paste OAuth client secrets, tokens, or callback URLs into the conversation unless they explicitly choose that workflow.

## Calling Rules

- Prefer JSON for read commands: add `--json` when the command supports it.
- Quote task titles, project names, habit names, and other user-provided strings.
- Convert relative dates such as "today", "tomorrow", or "next Friday" into explicit `YYYY-MM-DD` dates before calling the CLI.
- Resolve ambiguity before writes by listing or getting likely matches first.
- Preserve the user's configured region unless they explicitly ask to change it.
- Do not delete, archive, logout, switch regions, or overwrite config unless the user clearly requested that action.
- Use `--yes` for destructive commands only when the user already explicitly approved the action.
- If a write command affects a named task or project, prefer confirming the target with a read command when multiple matches are plausible.

## Common Commands

### Auth

```bash
tick auth status
tick auth login
tick auth logout
```

### Projects

```bash
tick project ls --json
tick project get "Work" --json
tick project add "Work"
tick project add "Notes" --kind NOTE
tick project update "Work" --name "New Name"
tick project rm "Work" --yes
```

### Tasks

```bash
tick task ls --json
tick task ls --project "Work" --json
tick task ls --today --json
tick task ls --overdue --json
tick task ls --status completed --json
tick task ls --priority 5 --json
tick task ls --from 2026-04-01 --to 2026-04-30 --json
tick task ls --tag urgent --json
tick task get "Write spec" --json

tick task add "Write spec" --project "Work" --due 2026-04-20
tick task add "Review" --project "Work" --due 2026-04-20 --all-day
tick task update "Write spec" --title "Write detailed spec" --due 2026-04-21
tick task done "Write spec"
tick task reopen "Write spec"
tick task move "Write spec" --to "Personal"
tick task move "Write spec" --project "Work" --to "Personal"
tick task rm "Write spec" --yes
```

### Today

```bash
tick today --json
```

### Quick Add

Use quick add for compact capture when the user's request maps naturally to one task:

```bash
tick quick add "Write spec #Work !5 ^2026-04-10"
```

Syntax:

- `#ProjectName` selects the project.
- `!1`, `!3`, and `!5` set low, medium, and high priority.
- `^YYYY-MM-DD` sets the due date.

If `default_project` is configured, `#ProjectName` may be omitted.

### Habits

```bash
tick habit ls --json
tick habit get "Read 30 min" --json
tick habit add "Read 30 min" --goal 30 --unit "min"
tick habit update "Read 30 min" --goal 60
tick habit checkin "Read 30 min" --value 30
tick habit log "Read 30 min" --json
tick habit archive "Read 30 min"
```

### Focus

```bash
tick focus ls --json
tick focus ls --type 0 --json
tick focus get "<focus-id>" --json
tick focus get "<focus-id>" --type 0 --json
```

`--type 0` means pomodoro. The default focus type is timer.

### Config

```bash
tick config list
tick config get region
tick config get output
tick config get default_project
tick config set output json
tick config set default_project "Work"
tick config set region ticktick
tick config set region dida365
```

Config commands print plain text rather than JSON.

Changing regions after login requires re-authentication. Ask for explicit confirmation before changing `region` or running `auth logout`.

## Error Handling

- If authentication fails, ask the user to run `tick auth login`.
- If a task, project, or habit name is ambiguous, list candidates and ask the user to choose.
- If JSON output is unavailable, parse table output conservatively and prefer a follow-up `get` before writing.
- If the binary is missing in the source repo, run `make build` and retry with `bin/tick`.
