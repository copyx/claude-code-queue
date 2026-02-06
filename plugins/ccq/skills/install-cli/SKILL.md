---
name: install-cli
description: Install the ccq CLI binary and verify tmux dependency
---

# Install ccq CLI

Install the ccq binary for managing Claude Code session queues.

## Steps

1. Detect the current OS and architecture
2. Download the appropriate binary from GitHub Releases:
   - `https://github.com/copyx/claude-code-queue/releases/latest/download/ccq-{os}-{arch}`
   - OS: `darwin` or `linux`
   - Arch: `amd64` or `arm64`
3. Place the binary at `~/.local/bin/ccq`
4. Make it executable: `chmod +x ~/.local/bin/ccq`
5. Verify `~/.local/bin` is in PATH. If not, suggest adding it.
6. Run `ccq` to verify installation.
7. Check if `tmux` is installed. If not:
   - macOS: suggest `brew install tmux`
   - Linux: suggest `sudo apt install tmux` or `sudo yum install tmux`

Use Bash commands to accomplish each step. Report success or failure clearly.
