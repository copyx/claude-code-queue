# Installation Process Improvements

**Date:** 2026-02-08
**Status:** Approved

## Problem Statement

The current installation process has several issues:

1. **Unnecessary Claude Code dependency**: Installation is implemented as a Claude Code skill, but regular users should be able to install without Claude Code
2. **Validation false positives**: Post-installation verification runs `ccq` which triggers interactive tmux prefix setup, causing the script to misinterpret this as an error
3. **Dependency check timing**: tmux dependency is checked after installation instead of before, leading to poor UX

## Solution Overview

Separate installation into a standalone shell script that can be used independently, while keeping the Claude Code skill as a simple wrapper for convenience.

## Design

### File Structure

```
/
├── install.sh              # New: standalone installation script
├── README.md              # Updated: installation guide
└── plugins/ccq/skills/
    └── install-cli/       # Modified: wrapper for install.sh
        └── SKILL.md
```

### Installation Flow

#### For Regular Users (without Claude Code)

```bash
curl -fsSL https://raw.githubusercontent.com/copyx/claude-code-queue/main/install.sh | bash
```

Script execution order:
1. Check tmux existence → exit immediately if missing with installation instructions
2. Detect OS/architecture (darwin/linux, amd64/arm64)
3. Download binary from GitHub Releases
4. Place at `~/.local/bin/ccq` with execute permissions
5. Verify installation with `ccq --version`
6. Display success message and PATH instructions

#### For Claude Code Users

- Run `/install-cli` skill
- Skill internally executes `curl | bash` with the same install.sh
- Same installation process

### install.sh Implementation

#### 1. tmux Dependency Check (First Priority)

```bash
if ! command -v tmux &> /dev/null; then
    echo "Error: tmux is required but not installed."
    echo ""
    echo "Install tmux first:"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "  brew install tmux"
    elif command -v apt &> /dev/null; then
        echo "  sudo apt install tmux"
    elif command -v yum &> /dev/null; then
        echo "  sudo yum install tmux"
    else
        echo "  (see https://github.com/tmux/tmux/wiki)"
    fi
    exit 1
fi
```

**Rationale**: Strict validation prevents installation when tmux is missing, ensuring the tool will work after installation.

#### 2. OS/Architecture Detection

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [[ "$OS" != "darwin" && "$OS" != "linux" ]]; then
    echo "Unsupported OS: $OS"
    exit 1
fi
```

#### 3. Binary Download and Installation

```bash
BINARY_URL="https://github.com/copyx/claude-code-queue/releases/latest/download/ccq-${OS}-${ARCH}"
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="$INSTALL_DIR/ccq"

mkdir -p "$INSTALL_DIR"
curl -fsSL "$BINARY_URL" -o "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"
```

#### 4. Installation Verification

```bash
if ! "$INSTALL_PATH" --version &> /dev/null; then
    echo "Error: Installation failed - binary cannot execute"
    exit 1
fi

echo "✓ ccq installed successfully to $INSTALL_PATH"
echo ""
echo "Make sure ~/.local/bin is in your PATH"
```

**Note**: Requires `--version` flag implementation in ccq CLI.

### ccq CLI Modifications

Add `--version` flag to enable non-interactive validation:

```go
const Version = "1.0.0" // or injected at build time

func main() {
    if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
        fmt.Printf("ccq version %s\n", Version)
        os.Exit(0)
    }

    // existing logic...
}
```

### Claude Code Skill Update

**`plugins/ccq/skills/install-cli/SKILL.md`:**

```markdown
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
   - Install to ~/.local/bin/ccq
   - Verify installation with --version

Use the Bash tool to execute:
```bash
curl -fsSL https://raw.githubusercontent.com/copyx/claude-code-queue/main/install.sh | bash
```
```

### README.md Update

Add Installation section:

```markdown
## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/copyx/claude-code-queue/main/install.sh | bash
```

**Prerequisites:** tmux must be installed first.

**Claude Code users:** Run `/install-cli` skill instead.

### Manual Installation

1. Download binary for your platform from [Releases](https://github.com/copyx/claude-code-queue/releases/latest)
2. Move to `~/.local/bin/ccq`
3. Make executable: `chmod +x ~/.local/bin/ccq`
```

## Implementation Checklist

1. **Create install.sh** (new file)
   - tmux check logic
   - OS/architecture detection
   - Binary download
   - Installation and verification

2. **Add --version flag to ccq CLI**
   - Modify main.go
   - Define version constant or inject at build time

3. **Update skill**
   - Simplify `plugins/ccq/skills/install-cli/SKILL.md`

4. **Update README**
   - Add Installation section

5. **Testing**
   - Test install.sh on macOS/Linux
   - Verify error handling when tmux is missing
   - Confirm --version flag works
   - Test /install-cli from Claude Code

## Benefits

- **User-friendly**: Regular users can install without Claude Code
- **Reliable validation**: Non-interactive --version check prevents false errors
- **Better UX**: Dependency check happens first, failing fast with clear instructions
- **Maintainable**: Single source of truth (install.sh) for installation logic
