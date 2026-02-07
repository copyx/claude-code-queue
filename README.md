# ccq — Claude Code Queue Manager

FIFO queue-based auto-switcher for multiple Claude Code sessions via tmux.

Run several Claude Code sessions in a single tmux session. When you submit a prompt in the current window, ccq automatically switches you to the session that has been waiting for input the longest. No more missed idle prompts.

## Prerequisites

- **tmux** 3.0+
- **Go** 1.22+ (only needed to build from source)
- **Claude Code** CLI

## Installation

### Claude Code Plugin (recommended)

Add the marketplace and install the plugin:

```
/plugin marketplace add copyx/claude-code-queue
/plugin install ccq@claude-code-queue
```

Then install the CLI binary from inside Claude Code:

```
/ccq:install-cli
```

This downloads the `ccq` binary and checks that tmux is available.

### Build from source

```bash
git clone https://github.com/copyx/claude-code-queue.git
cd claude-code-queue
make build
make install   # copies the binary to ~/.local/bin/ccq
```

Make sure `~/.local/bin` is on your `PATH`.

## Usage

### Start a session

```bash
ccq
```

Creates a `ccq` tmux session, launches Claude Code inside it, and attaches your terminal. On first run you will be prompted to choose a tmux prefix key (`Ctrl+Space` is recommended since Claude Code uses most `Ctrl` combinations).

### Add more sessions

From a different terminal (or a different project directory), just run `ccq` again:

```bash
ccq
```

If a session already exists, ccq adds a new window and starts Claude Code in it. You'll see the new window briefly for initial setup (trust prompt, etc.), then ccq automatically returns you to your previous view.

### Auto-switching

While you work, ccq tracks every window's state through Claude Code hooks:

1. You type a prompt in window 1 and press Enter.
2. Window 1 becomes **busy**.
3. ccq switches you to the window that has been **idle** the longest.
4. When that window also becomes busy, you move to the next idle one, and so on.

If all windows are busy, ccq stays on the current window until one becomes idle.

### Keybindings

All keybindings use the tmux prefix you chose during setup.

| Key | Action |
|---|---|
| `prefix + a` | Toggle auto/manual mode |
| `prefix + n` | Next window (tmux built-in) |
| `prefix + p` | Previous window (tmux built-in) |

In **manual** mode, state tracking still happens but ccq will not switch windows for you. Press `prefix + a` again to re-enable auto-switching (ccq immediately checks the queue and switches if needed).

## How it works

ccq is a hook-driven state machine with no long-running daemon.

1. The Claude Code plugin registers hooks for key events (`Notification`, `UserPromptSubmit`, `PostToolUse`, `PostToolUseFailure`, `SessionEnd`).
2. Each hook invokes `ccq _hook idle`, `ccq _hook busy`, or `ccq _hook remove` as a short-lived process.
3. The hook handler reads and writes tmux window variables (`@ccq_state`, `@ccq_idle_since`) to track which windows are idle and when they became idle.
4. When the current window is busy and at least one other window is idle, `ccq` issues a `tmux select-window` to the oldest idle window.
5. No external database or lock file is needed. tmux itself serializes all commands, and any transient inconsistency self-corrects on the next hook invocation.

## Configuration

Configuration is stored at `~/.config/ccq/config` (JSON):

```json
{
  "prefix": "C-Space"
}
```

| Key | Description | Default |
|---|---|---|
| `prefix` | tmux prefix key | Set on first run |

## License

MIT
