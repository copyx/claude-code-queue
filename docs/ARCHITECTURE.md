# Architecture

## Overview

ccq is a hook-driven state machine with no long-running daemon. The Claude Code plugin registers hooks that invoke `ccq _hook` as short-lived processes. Each invocation reads and writes tmux window variables to track state, then optionally issues a `select-window` to switch the user's view.

```
┌──────────────────────────────────────────────────┐
│                User Terminal                      │
│                                                   │
│  ┌─────────── ccq tmux session ───────────────┐  │
│  │                                             │  │
│  │  [window 1: claude]  ← current view         │  │
│  │  [window 2: claude]  (background)           │  │
│  │  [window 3: claude]  (background)           │  │
│  │                                             │  │
│  │  status bar: [AUTO] 3 windows               │  │
│  └─────────────────────────────────────────────┘  │
│                                                   │
│  ccq _hook idle/busy  ←── Claude Code Hooks       │
│  (tmux variable read/write + select-window)       │
└──────────────────────────────────────────────────┘
```

## Components

| Component | Role |
|---|---|
| `ccq` CLI (Go binary) | tmux session creation/management, queue logic, hook command processing |
| Claude Code Plugin | Hook definitions (`hooks/hooks.json`), `/ccq:install-cli` skill |
| tmux | Session/window/PTY management, view switching, status bar |

## State Machine

```
         Notification                    UserPromptSubmit
      (idle_prompt |                            │
  permission_prompt |                           │
  elicitation_dialog)                           │
            │                                   │
            ▼                                   ▼
        ┌───────┐                          ┌────────┐
        │ idle  │ ◄──── Notification ───── │  busy  │
        │       │ ────► PostToolUse ─────► │        │
        └───────┘      PostToolUseFailure  └────────┘
                       UserPromptSubmit
```

### Hook Events

| Hook | Command | Trigger |
|---|---|---|
| `Notification` (idle_prompt, permission_prompt, elicitation_dialog) | `ccq _hook idle` | Claude Code is waiting for user input |
| `UserPromptSubmit` | `ccq _hook busy` | User submitted a prompt |
| `PostToolUse` / `PostToolUseFailure` | `ccq _hook busy` (async) | Tool execution completed |
| `SessionEnd` | `ccq _hook remove` | Claude Code session ended |

### Hook Handlers

| Command | Action |
|---|---|
| `ccq _hook idle` | Set `@ccq_state=idle` and `@ccq_idle_since=<timestamp>` on the window. If `@ccq_return_to` is set (initial setup), return to previous window/detach instead of auto-switching. Otherwise attempt auto-switch. |
| `ccq _hook busy` | Set `@ccq_state=busy` on the window. Attempt auto-switch. |
| `ccq _hook remove` | Unset `@ccq_state` and `@ccq_idle_since` from the window. |

## Tmux Variables

| Variable | Scope | Values | Purpose |
|---|---|---|---|
| `@ccq_state` | window | `idle`, `busy` | Current window state |
| `@ccq_idle_since` | window | Unix timestamp | When the window became idle (FIFO ordering) |
| `@ccq_return_to` | window | window ID or `__detach__[:<tty>]` | Return target after initial setup |
| `@ccq_auto_switch` | session | `on`, `off` | Auto-switch toggle |

## Auto-Switch Rules

1. If `@ccq_auto_switch` is `off`, only mark state — do not switch.
2. If the current (active) window is idle, never switch (user may be typing).
3. Switch only when the current window is busy — select the oldest idle window.
4. When toggled ON, immediately check the queue and switch if conditions are met.

## Initial Setup Flow (`@ccq_return_to`)

When `ccq` adds a new window, it briefly shows the new window so the user can handle the initial Claude Code setup (trust prompt, etc.). Once the first `idle_prompt` hook fires:

- **Inside tmux**: `@ccq_return_to` contains the previous window ID → `select-window` back.
- **Outside tmux**: `@ccq_return_to` contains `__detach__:<tty>` → `detach-client` to return the user to their original terminal.

## Edge Cases

### Hook Detection Gap

| Scenario | Busy detection | Gap |
|---|---|---|
| Text prompt submit | `UserPromptSubmit` (immediate) | None |
| Permission approval | `PostToolUse` (after tool completes) | Tool execution time |
| Elicitation response | `PostToolUse` (next tool use) | Claude processing time |

During the gap, a window may be incorrectly marked as idle, causing an unnecessary switch. This self-corrects on the next hook invocation.

**Race Condition Prevention:** `PostToolUse` hooks fire asynchronously and may execute after a `Notification` hook has already marked the window as idle. To prevent incorrectly switching away from an active window where the user is typing, `HandleBusy` skips marking a window as busy if it's already idle (the idle state is more recent and accurate).

### Race Conditions

Two hooks firing simultaneously could cause non-atomic read-judge-write on tmux variables. Since the tmux server serializes all commands, the actual probability is low. Any inconsistency self-corrects on the next hook invocation. No external locks are used.

### Abnormal Termination

`remain-on-exit off` ensures windows are automatically destroyed when their process exits. Even if `SessionEnd` hook doesn't fire, the window disappearing removes it from the queue naturally.

### Manual Window Close

If the active window is closed, tmux's default behavior (move to next window) takes over. The next `_hook idle` invocation self-corrects the state.

### Nested tmux

`$TMUX` environment variable determines behavior:
- Outside tmux: `tmux attach-session`
- Inside tmux: `tmux switch-client` (no nesting)

## Project Layout

```
├── main.go                          # CLI entry point
├── internal/
│   ├── cmd/                         # Command handlers (root, hook, toggle)
│   ├── tmux/                        # tmux CLI wrapper
│   ├── queue/                       # FIFO queue logic (mark idle/busy, find oldest)
│   ├── switcher/                    # Auto-switch decision logic
│   ├── hook/                        # Hook event handlers
│   └── config/                      # User config (~/.config/ccq/config)
├── plugins/ccq/                     # Claude Code plugin
│   ├── .claude-plugin/plugin.json
│   ├── hooks/hooks.json
│   └── skills/install-cli/SKILL.md
├── .claude-plugin/marketplace.json  # Marketplace catalog
├── .github/workflows/release.yml    # CI: build + GitHub Release on tag push
└── test/                            # Integration tests
```
