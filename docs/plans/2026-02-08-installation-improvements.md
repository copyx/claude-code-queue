# Installation Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Separate installation into a standalone shell script with proper tmux validation and add --version flag to ccq CLI.

**Architecture:** Create install.sh script as single source of truth for installation. Add --version flag to ccq CLI for non-interactive validation. Update Claude Code skill to be a simple wrapper. Update README with installation instructions.

**Tech Stack:** Bash (install.sh), Go (CLI modifications), Markdown (documentation)

---

## Task 1: Add --version Flag to ccq CLI

**Files:**
- Modify: `main.go:10-55`

**Step 1: Add version constant and flag handling**

Add version constant at the top of main.go and handle --version flag before other argument parsing:

```go
package main

import (
	"fmt"
	"os"

	"github.com/jingikim/ccq/internal/cmd"
)

const Version = "1.0.0"

func printHelp() {
	fmt.Print(`ccq - Claude Code Queue Manager

FIFO queue-based auto-switcher for multiple Claude Code sessions via tmux.

Usage:
  ccq             Start ccq or add a new session
  ccq -h, --help  Show this help
  ccq --version   Show version

Keybindings (inside ccq session):
  prefix + a      Toggle auto/manual switching
  prefix + n/p    Next/previous window
  prefix + w      Window list
  prefix + d      Detach from session
`)
}

func main() {
	var err error

	if len(os.Args) < 2 {
		err = cmd.Root()
	} else {
		switch os.Args[1] {
		case "-h", "--help", "help":
			printHelp()
		case "--version", "-v":
			fmt.Printf("ccq version %s\n", Version)
			return
		case "_hook":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: ccq _hook <idle|busy|remove>")
				os.Exit(1)
			}
			err = cmd.Hook(os.Args[2])
		case "_toggle":
			err = cmd.Toggle()
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			fmt.Fprintln(os.Stderr, "Run 'ccq -h' for usage.")
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Build and test --version flag**

Run: `make build && ./ccq --version`
Expected: Output "ccq version 1.0.0"

Run: `./ccq -v`
Expected: Output "ccq version 1.0.0"

**Step 3: Verify exit code is 0**

Run: `./ccq --version; echo $?`
Expected: Output shows version then "0"

**Step 4: Commit**

```bash
git add main.go
git commit -m "feat: add --version flag for non-interactive validation"
```

---

## Task 2: Create install.sh Script

**Files:**
- Create: `install.sh`

**Step 1: Write install.sh with tmux check first**

Create complete installation script with all sections:

```bash
#!/usr/bin/env bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check tmux first (hard requirement)
if ! command -v tmux &> /dev/null; then
    echo -e "${RED}Error: tmux is required but not installed.${NC}"
    echo ""
    echo "Install tmux first:"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "  brew install tmux"
    elif command -v apt &> /dev/null; then
        echo "  sudo apt install tmux"
    elif command -v yum &> /dev/null; then
        echo "  sudo yum install tmux"
    else
        echo "  See https://github.com/tmux/tmux/wiki"
    fi
    exit 1
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

if [[ "$OS" != "darwin" && "$OS" != "linux" ]]; then
    echo -e "${RED}Error: Unsupported OS: $OS${NC}"
    exit 1
fi

# Download and install
BINARY_URL="https://github.com/copyx/claude-code-queue/releases/latest/download/ccq-${OS}-${ARCH}"
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="$INSTALL_DIR/ccq"

echo "Downloading ccq for ${OS}-${ARCH}..."
mkdir -p "$INSTALL_DIR"

if ! curl -fsSL "$BINARY_URL" -o "$INSTALL_PATH"; then
    echo -e "${RED}Error: Failed to download binary from $BINARY_URL${NC}"
    exit 1
fi

chmod +x "$INSTALL_PATH"

# Verify installation
if ! "$INSTALL_PATH" --version &> /dev/null; then
    echo -e "${RED}Error: Installation failed - binary cannot execute${NC}"
    rm -f "$INSTALL_PATH"
    exit 1
fi

VERSION=$("$INSTALL_PATH" --version)
echo -e "${GREEN}✓ $VERSION installed successfully to $INSTALL_PATH${NC}"
echo ""

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo -e "${YELLOW}Warning: ~/.local/bin is not in your PATH${NC}"
    echo ""
    echo "Add this to your shell profile (~/.bashrc or ~/.zshrc):"
    echo '  export PATH="$HOME/.local/bin:$PATH"'
    echo ""
    echo "Then reload your shell or run:"
    echo "  source ~/.bashrc  # or ~/.zshrc"
else
    echo "You can now run: ccq"
fi
```

**Step 2: Make install.sh executable**

Run: `chmod +x install.sh`

**Step 3: Test with tmux installed (manual verification)**

This step requires manual testing - the script should complete successfully and show green checkmark.

Run: `./install.sh`
Expected: Downloads binary, shows version, success message

**Step 4: Commit**

```bash
git add install.sh
git commit -m "feat: add standalone installation script with tmux validation"
```

---

## Task 3: Update Claude Code Skill

**Files:**
- Modify: `plugins/ccq/skills/install-cli/SKILL.md:1-26`

**Step 1: Simplify SKILL.md to wrapper**

Replace entire content with simpler wrapper approach:

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

If you need to test with the local version:

```bash
bash install.sh
```
```

**Step 2: Verify skill file is valid**

Run: `head -n 5 plugins/ccq/skills/install-cli/SKILL.md`
Expected: Shows YAML frontmatter with name and description

**Step 3: Commit**

```bash
git add plugins/ccq/skills/install-cli/SKILL.md
git commit -m "refactor: simplify install-cli skill to wrapper for install.sh"
```

---

## Task 4: Update README Installation Section

**Files:**
- Modify: `README.md:13-42`

**Step 1: Replace installation section**

Replace the current Installation section (lines 13-42) with:

```markdown
## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/copyx/claude-code-queue/main/install.sh | bash
```

**Prerequisites:** tmux must be installed first. See [tmux installation guide](https://github.com/tmux/tmux/wiki/Installing).

**Claude Code users:** Install the plugin first, then run `/install-cli` skill.

### Claude Code Plugin

Add the marketplace and install the plugin:

```bash
/plugin marketplace add copyx/claude-code-queue
/plugin install ccq@claude-code-queue
```

Then install the CLI binary:

```bash
/install-cli
```

### Manual Installation

1. Download the binary for your platform from [Releases](https://github.com/copyx/claude-code-queue/releases/latest):
   - macOS (Intel): `ccq-darwin-amd64`
   - macOS (Apple Silicon): `ccq-darwin-arm64`
   - Linux (x64): `ccq-linux-amd64`
   - Linux (ARM64): `ccq-linux-arm64`
2. Move to `~/.local/bin/ccq`
3. Make executable: `chmod +x ~/.local/bin/ccq`
4. Ensure `~/.local/bin` is in your PATH

### Build from Source

```bash
git clone https://github.com/copyx/claude-code-queue.git
cd claude-code-queue
make build
make install   # copies the binary to ~/.local/bin/ccq
```

Make sure `~/.local/bin` is on your `PATH`.
```

**Step 2: Verify markdown formatting**

Run: `head -n 60 README.md | tail -n 40`
Expected: Shows well-formatted installation section

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update installation section with install.sh and improved structure"
```

---

## Task 5: Test Installation Script (Manual)

**Files:**
- None (manual testing only)

**Step 1: Test successful installation**

This requires manual execution. Document the test:

Run: `./install.sh`
Expected:
- Checks tmux (should pass)
- Downloads binary
- Installs to ~/.local/bin/ccq
- Runs --version check
- Shows green success message

**Step 2: Test tmux missing scenario**

This requires temporarily renaming/removing tmux from PATH for testing:

Run: `PATH=/usr/bin:/bin ./install.sh`
Expected:
- Red error message about tmux
- Suggests installation command
- Exits with code 1
- Does NOT install binary

**Step 3: Test --version verification**

Run: `~/.local/bin/ccq --version`
Expected: `ccq version 1.0.0`

**Step 4: Document test results**

Create test results summary in commit message for final commit.

---

## Task 6: Final Integration Test

**Files:**
- None (integration testing)

**Step 1: Clean test - remove existing binary**

Run: `rm -f ~/.local/bin/ccq`

**Step 2: Run installation script**

Run: `./install.sh`
Expected: Complete installation with success message

**Step 3: Verify ccq works**

Run: `ccq --version`
Expected: `ccq version 1.0.0`

Run: `ccq --help`
Expected: Help text with --version listed

**Step 4: Final commit**

```bash
git add -A
git commit -m "test: verify installation improvements integration

Tested:
- install.sh completes successfully
- tmux dependency check blocks installation when missing
- --version flag works for non-interactive validation
- Claude Code skill wrapper approach
- README installation instructions

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Notes for Implementation

### Testing Considerations

- **tmux dependency testing** requires temporarily making tmux unavailable (adjust PATH)
- **Binary download testing** requires working internet and valid GitHub release
- **PATH checking** logic assumes standard shell profile locations

### Error Handling

- Script uses `set -e` to exit on any command failure
- Explicit error messages with colors for visibility
- Cleanup on verification failure (removes broken binary)

### Future Improvements (Out of Scope)

- Version injection at build time via ldflags
- Checksum validation for downloaded binary
- Support for custom installation directory
- Uninstall script
