# Smart Re-attach: Client-aware Session Handling

## Problem

When a user detaches from a ccq session and later runs `ccq` again, the current behavior **always creates a new Claude window** regardless of intent. This is unintuitive — a detached user most likely wants to re-attach to their existing session, not spawn a new Claude instance.

## Current Behavior

```
ccq (session exists) → always addWindow() → attach
```

No distinction between "I want to come back" and "I want another Claude."

## Design

### Decision Matrix

| Condition | Detected Via | Action | UX Message |
|-----------|-------------|--------|------------|
| No session | `HasSession()` = false | Create new session + attach | `ccq: creating new session...` |
| Session + inside ccq tmux | `$TMUX` set, same session | Add new window + switch | `ccq: adding new Claude window to current session (N windows)...` |
| Session + no client attached | `list-clients -t ccq` empty | Re-attach only (no new window) | `ccq: session found (N windows, no client attached). Re-attaching...` |
| Session + client attached + external | `list-clients` non-empty, external terminal | Prompt user | See prompt below |

### Prompt (client attached + external)

```
ccq: session found (N windows, M client(s) attached)
  1) Attach to existing session (default)
  2) Add new Claude window and attach
  Select [1]:
```

### Key Principles

1. **State visibility**: Always tell the user what state was detected and what action is taken
2. **Least surprise**: Default to the most common intent for each scenario
3. **Explicit over implicit**: When intent is ambiguous (client exists + external), ask

## Implementation

### Changes to `root.go`

1. Add `listClients()` helper using `tmux list-clients -t ccq`
2. Add `windowCount()` helper using `tmux list-windows -t ccq`
3. Modify `Root()` to branch on client presence:

```go
func Root() error {
    // ... tmux check ...
    tm := tmux.New(sessionName)

    if !tm.HasSession() {
        // Create new session (unchanged)
    }

    // Session exists — determine action
    clients := tm.ListClients()
    windowCount := len(tm.ListWindows())
    inCcqSession := isInCcqSession()  // $TMUX points to ccq session

    if inCcqSession {
        fmt.Printf("ccq: adding new Claude window to current session (%d windows)...\n", windowCount+1)
        return addWindow(tm)
    }

    if len(clients) == 0 {
        fmt.Printf("ccq: session found (%d windows, no client attached). Re-attaching...\n", windowCount)
        return attachOrSwitch(tm)
    }

    // Client attached + external: prompt
    fmt.Printf("ccq: session found (%d windows, %d client(s) attached)\n", windowCount, len(clients))
    choice := promptReattachOrAdd()
    if choice == "add" {
        return addWindow(tm)
    }
    return attachOrSwitch(tm)
}
```

### Changes to `internal/tmux/tmux.go`

Add `ListClients() []string` method:
```go
func (t *Tmux) ListClients() []string {
    out, err := t.Output("list-clients", "-t", t.session, "-F", "#{client_tty}")
    if err != nil {
        return nil
    }
    // parse and return non-empty lines
}
```

### Edge Cases

- **Session exists, 0 windows** (all Claude instances exited): `windowCount == 0` → treat as re-attach (tmux may auto-destroy, or user sees empty session and can `ccq` again to add)
- **Multiple clients**: Show count in message, still prompt for action
- **Nested tmux detection**: Use `$TMUX` env var to determine if inside ccq session specifically (parse session name from `$TMUX` socket path + `display-message -p "#{session_name}"`)

## Testing

- Unit test: `ListClients()` parsing
- Integration scenarios:
  - `ccq` with no session → creates session
  - `ccq` inside ccq session → adds window (no prompt)
  - `ccq` with detached session → re-attaches (no new window)
  - `ccq` with attached session from external → prompts user
