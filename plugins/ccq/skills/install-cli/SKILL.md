---
name: install-cli
description: Install the ccq CLI binary using the official installation script
---

# Install ccq CLI

Installs the ccq binary by running the official installation script.

## What it does

1. Downloads and executes the installation script from GitHub
2. The script will:
   - Check for tmux (required dependency)
   - Detect OS and architecture
   - Download the appropriate binary
   - Verify checksum for security
   - Install to ~/.local/bin/ccq
   - Verify installation with --version

Use the Bash tool to execute:

```bash
curl -fsSL https://raw.githubusercontent.com/copyx/claude-code-queue/main/install.sh | bash
```

If you need to test with the local version:

```bash
bash install.sh
```
