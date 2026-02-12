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
│  ccq _hook idle/prompt ←── Claude Code Hooks      │
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
        │       │ ────► UserPromptSubmit ► │        │
        └───────┘ ────► PreToolUse ──────► └────────┘
```

### Hook Events

| Hook | Command | Trigger |
|---|---|---|
| `Notification` (idle_prompt, permission_prompt, elicitation_dialog) | `ccq _hook idle` | Claude Code is waiting for user input |
| `PreToolUse` | `ccq _hook busy` | Tool is about to execute (catches permission/elicitation answers) |
| `UserPromptSubmit` | `ccq _hook prompt` | User submitted a prompt |
| `SessionEnd` | `ccq _hook remove` | Claude Code session ended |

### Hook Handlers

| Command | Action |
|---|---|
| `ccq _hook idle` | Set `@ccq_state=idle` and `@ccq_idle_since=<timestamp>` on the window, queuing it for the next auto-switch. If `@ccq_return_to` is set (initial setup), return to previous window/detach instead. No auto-switch — idle windows wait in the queue. |
| `ccq _hook busy` | If the window is idle (user just answered a permission/elicitation), mark busy and auto-switch. If already busy, no-op (avoids redundant writes during normal tool execution). |
| `ccq _hook prompt` | Set `@ccq_state=busy` on the window (override idle). Attempt auto-switch to the oldest idle window. |
| `ccq _hook remove` | Unset `@ccq_state` and `@ccq_idle_since` from the window. |

## Tmux Variables

| Variable | Scope | Values | Purpose |
|---|---|---|---|
| `@ccq_state` | window | `idle`, `busy` | Current window state |
| `@ccq_idle_since` | window | Unix timestamp | When the window became idle (FIFO ordering) |
| `@ccq_return_to` | window | window ID or `__detach__[:<tty>]` | Return target after initial setup |
| `@ccq_auto_switch` | session | `on`, `off` | Auto-switch toggle |

## Auto-Switch Rules

Auto-switch is triggered by user actions only: `UserPromptSubmit` (submitting a prompt) and `PreToolUse` on an idle window (answering a permission/elicitation). Idle windows queue up via `Notification` hooks and wait their turn.

1. If `@ccq_auto_switch` is `off`, only mark state — do not switch.
2. If the current (active) window is idle, never switch (user may be typing).
3. Switch only when the current window is busy — select the oldest idle window.
4. When toggled ON, immediately check the queue and switch if conditions are met.

## CLI Commands

| Command | Action |
|---|---|
| `ccq` | Add new Claude window + conditional attach (see below) |
| `ccq attach` | Attach to existing session (no new window) |
| `ccq status` | Show detailed session status in terminal |

## Smart Re-attach (`ccq` default behavior)

When `ccq` adds a new window, it briefly shows it for initial Claude Code setup (trust prompt, etc.). What happens after the first `idle_prompt` hook fires depends on context:

| Condition | After init |
|---|---|
| Inside ccq tmux | `@ccq_return_to` = previous window ID → `select-window` back |
| Outside tmux, no other clients | `@ccq_return_to` not set → stay attached (normal auto-switch) |
| Outside tmux, other clients attached | `@ccq_return_to` = `__detach__:<tty>` → `detach-client` to return to original terminal |

## Edge Cases

### Why PreToolUse Instead of PostToolUse

`PreToolUse` is used (sync) instead of `PostToolUse` (async) to detect when the user answers a permission or elicitation prompt. `PreToolUse` fires right when the tool starts — immediately after the user's action. `PostToolUse` was removed because its async execution raced with `Notification` hooks, corrupting `@ccq_state`.

`HandleBusy` guards against redundant work: if the window is already busy (normal tool execution), it's a no-op. Only when the window transitions from idle → busy (user answered a question) does it mark state and trigger auto-switch.

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
