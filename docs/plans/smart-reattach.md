# Smart Re-attach: Client-aware Session Handling

## Problem

When a user detaches from a ccq session and later runs `ccq` again, the current behavior **always creates a new Claude window and auto-detaches** regardless of context. This is unintuitive — a detached user most likely wants to re-attach to their existing session, not just spawn a background window.

## Design

### Command Interface

```
ccq                  # Add new Claude window (+ conditional attach)
ccq attach           # Attach to existing session (no new window)
ccq status           # Show detailed session status
```

### `ccq` Default Behavior (revised)

`ccq` always adds a new Claude window. The only variable is what happens after initialization.

| Condition | Detection | Action | After init |
|-----------|-----------|--------|------------|
| No session (or 0 windows) | `HasSession()` + `ListWindows()` | Create session + window, attach | Stay attached |
| Inside ccq tmux | `$TMUX` points to ccq session | Add window, switch to it | `@ccq_return_to` switches back to previous window |
| Outside, no clients | `ListClients()` empty | Add window, attach for init | Stay attached (`@ccq_return_to` unset) |
| Outside, clients exist | `ListClients()` non-empty | Add window, attach for init | Auto-detach (`@ccq_return_to = __detach__:<tty>`) |

Key change: "outside, no clients" no longer sets `@ccq_return_to = __detach__`, so the user stays attached after initialization completes.

### `ccq attach`

Plain attach to existing session. No new window created.

- Uses `attachOrSwitch()` — works from outside tmux (attach) and inside other tmux sessions (switch-client).
- No session exists → prints error and exits.
- Inside ccq tmux → tmux's own nested session error handles this naturally.

### `ccq status`

Detailed per-window status for the terminal:

```
ccq: 3 windows, 1 client attached, auto-switch on

  #1  my-project      busy           ~/dev/my-project
  #2  api-service     idle   5m32s   ~/dev/api-service
  #3  frontend        idle   1m15s   ~/dev/frontend
```

Edge cases:
- No session → `ccq: no active session`
- Session, 0 windows → `ccq: session active, no windows`

## Implementation

### 1. `internal/tmux/tmux.go` — Add `ListClients()`

```go
func (t *Tmux) ListClients() []string {
    out, err := t.Output("list-clients", "-t", t.session, "-F", "#{client_tty}")
    if err != nil {
        return nil
    }
    // parse and return non-empty lines
}
```

### 2. `internal/cmd/root.go` — Modify `addWindow()` return-to logic

Current logic (outside tmux):
```
always set @ccq_return_to = __detach__:<tty>
```

New logic (outside tmux):
```
if len(ListClients()) == 0 {
    // Don't set @ccq_return_to → stay attached after init
} else {
    @ccq_return_to = __detach__:<tty>
}
```

The window creation, claude launch, and attach logic remain unchanged.

### 3. `main.go` — Add subcommands

```go
case "attach": err = cmd.Attach()
case "status": err = cmd.SessionStatus()
```

### 4. `internal/cmd/attach.go` — New file

```go
func Attach() error {
    tm := tmux.New(sessionName)
    if !tm.HasSession() {
        fmt.Fprintln(os.Stderr, "ccq: no active session")
        os.Exit(1)
    }
    return attachOrSwitch(tm)
}
```

### 5. `internal/cmd/session_status.go` — New file

Reads the same tmux variables as the dashboard (`status.go`) but formats for terminal output:
- Window list: `ListWindows()`
- State + idle duration: `GetWindowOption(@ccq_state)` + `@ccq_idle_since`
- Working directory: `GetWindowPanePath(windowID)`
- Client count: `ListClients()`
- Auto-switch: `GetSessionOption(@ccq_auto_switch)`

## Testing

- Unit test: `ListClients()` parsing
- Integration scenarios:
  - `ccq` with no session → creates session + attaches
  - `ccq` inside ccq session → adds window, switches, returns to previous
  - `ccq` with detached session (no clients) → adds window, stays attached
  - `ccq` with attached session (clients exist) → adds window, auto-detaches
  - `ccq attach` with session → attaches
  - `ccq attach` without session → error
  - `ccq status` with session → shows window details
  - `ccq status` without session → shows "no active session"
