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
| `Stop` | `ccq _hook idle` | Claude Code finished its response |
| `Notification` (idle_prompt, permission_prompt, elicitation_dialog) | `ccq _hook idle` | Claude Code is waiting for user input |
| `PreToolUse` | `ccq _hook busy` | Tool is about to execute (catches permission/elicitation answers) |
| `PreCompact` | `ccq _hook busy` | Claude Code starts compacting context |
| `UserPromptSubmit` | `ccq _hook prompt` | User submitted a prompt |
| `SessionEnd` | `ccq _hook remove` | Claude Code session ended |

### Hook Handlers

| Command | Action |
|---|---|
| `ccq _hook idle` | Set `@ccq_state=idle` and `@ccq_idle_since=<timestamp>` on the window, then attempt auto-switch. If the active window is busy, switch to the oldest idle window immediately; if the active window is also idle, the newly idle window waits in the queue. If `@ccq_return_to` is set (initial setup), return to previous window/detach instead. |
| `ccq _hook busy` | If the window is idle (user just answered a permission/elicitation), mark busy and auto-switch. If already busy, no-op (avoids redundant writes during normal tool execution). |
| `ccq _hook prompt` | Set `@ccq_state=busy` on the window (override idle). Attempt auto-switch to the oldest idle window. |
| `ccq _hook remove` | Unset `@ccq_state` and `@ccq_idle_since` from the window. |

## Tmux Variables

| Variable | Scope | Values | Purpose |
|---|---|---|---|
| `@ccq_state` | window | `idle`, `busy` | Current window state |
| `@ccq_idle_since` | window | Unix timestamp | When the window became idle (FIFO ordering) |
| `@ccq_return_to` | window | window ID or `__switch__:<session>` | Return target after initial setup |
| `@ccq_auto_switch` | session | `on`, `off` | Auto-switch toggle |

## Auto-Switch Rules

Auto-switch is triggered by three events: `Stop`/`Notification` (a window becomes idle), `UserPromptSubmit` (submitting a prompt), and `PreToolUse` on an idle window (answering a permission/elicitation). When a window becomes idle, it switches immediately only if the active window is busy; otherwise it queues up.

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

When `ccq` adds a new window, what happens depends on context:

| Condition | After init |
|---|---|
| Inside ccq tmux | `@ccq_return_to` = previous window ID → `HandleIdle` switches back |
| Inside different tmux session | `@ccq_return_to` = `__switch__:<session>` → `HandleIdle` switches back to original session |
| Outside tmux, no other clients | Attach to session (normal auto-switch) |
| Outside tmux, other clients attached | Don't attach — existing clients already see the new window via `select-window` |

## Edge Cases

### Why PreToolUse Instead of PostToolUse

`PreToolUse` is used (sync) instead of `PostToolUse` (async) to detect when the user answers a permission or elicitation prompt. `PreToolUse` fires right when the tool starts — immediately after the user's action. `PostToolUse` was removed because its async execution raced with `Notification` hooks, corrupting `@ccq_state`.

`HandleBusy` guards against redundant work: if the window is already busy (normal tool execution), it's a no-op. Only when the window transitions from idle → busy (user answered a question) does it mark state and trigger auto-switch.

### Abnormal Termination

`remain-on-exit off` ensures windows are automatically destroyed when their process exits. Even if `SessionEnd` hook doesn't fire, the window disappearing removes it from the queue naturally.

### Manual Window Close

If the active window is closed, tmux's default behavior (move to next window) takes over. The next `_hook idle` invocation self-corrects the state.

### Concurrent Hook Execution (TOCTOU Race)

When two windows become idle simultaneously (both fire `Stop` hook), their hook processes run concurrently. The background window's `HandleIdle` may call `TrySwitch` before the active window's hook has updated `@ccq_state` to idle — reading stale "busy" state and switching incorrectly.

**Mitigation**: `HandleIdle` sleeps 500ms for background windows before calling `TrySwitch`, giving concurrent hooks time to settle. The active window skips `TrySwitch` entirely (no delay needed — user is already there).

### Nested tmux

`$TMUX` environment variable determines behavior:
- Outside tmux: `tmux attach-session`
- Inside tmux: `tmux switch-client` (no nesting)

## Queue Implementation

The "FIFO queue" is not a traditional in-memory data structure. There is no slice, linked list, or channel holding queued items. Instead, each tmux window stores its own state as window-level variables (`@ccq_state`, `@ccq_idle_since`), and FIFO ordering is computed on demand.

```
Window @1: @ccq_state=idle  @ccq_idle_since=1709000100   ← oldest idle
Window @2: @ccq_state=busy  @ccq_idle_since=0
Window @3: @ccq_state=idle  @ccq_idle_since=1709000200
Window @4: @ccq_state=idle  @ccq_idle_since=1709000150

OldestIdle() → scans all windows → returns @1 (smallest timestamp)
```

- **Enqueue**: `MarkIdle` sets `@ccq_state=idle` and records the current Unix timestamp in `@ccq_idle_since`. If already idle, the timestamp is preserved (idempotent — prevents FIFO reordering on duplicate hook events).
- **Dequeue**: `OldestIdle` iterates all windows, filters by `@ccq_state=idle`, and returns the one with the smallest `@ccq_idle_since` timestamp.
- **Remove**: `MarkBusy` sets `@ccq_state=busy` and clears the timestamp.

This design has no persistent process or shared memory — each `ccq _hook` invocation is a short-lived process that reads tmux variables, computes the next action, and exits. tmux itself acts as the durable state store.

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
