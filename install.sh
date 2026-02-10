#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REPO="copyx/claude-code-queue"
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="$INSTALL_DIR/ccq"

# Cleanup on exit/interrupt
cleanup() {
    if [[ -f "$INSTALL_PATH.tmp" ]]; then
        rm -f "$INSTALL_PATH.tmp"
    fi
    if [[ -f "$INSTALL_DIR/checksums.txt" ]]; then
        rm -f "$INSTALL_DIR/checksums.txt"
    fi
}
trap cleanup EXIT INT TERM

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
DETECTED_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
DETECTED_ARCH=$(uname -m)

case $DETECTED_ARCH in
    x86_64) DETECTED_ARCH="amd64" ;;
    aarch64|arm64) DETECTED_ARCH="arm64" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $DETECTED_ARCH${NC}"
        exit 1
        ;;
esac

if [[ "$DETECTED_OS" != "darwin" && "$DETECTED_OS" != "linux" ]]; then
    echo -e "${RED}Error: Unsupported OS: $DETECTED_OS${NC}"
    exit 1
fi

# Backup existing installation
if [[ -f "$INSTALL_PATH" ]]; then
    BACKUP_PATH="$INSTALL_PATH.backup.$(date +%s)"
    echo "Backing up existing installation to $BACKUP_PATH"
    mv "$INSTALL_PATH" "$BACKUP_PATH"
fi

# Download binary and checksum
BINARY_URL="https://github.com/$GITHUB_REPO/releases/latest/download/ccq-${DETECTED_OS}-${DETECTED_ARCH}"
CHECKSUM_URL="https://github.com/$GITHUB_REPO/releases/latest/download/checksums.txt"

echo "Downloading ccq for ${DETECTED_OS}-${DETECTED_ARCH}..."
mkdir -p "$INSTALL_DIR"

if ! curl -fsSL "$BINARY_URL" -o "$INSTALL_PATH.tmp"; then
    echo -e "${RED}Error: Failed to download binary from $BINARY_URL${NC}"
    exit 1
fi

echo "Downloading checksums..."
if ! curl -fsSL "$CHECKSUM_URL" -o "$INSTALL_DIR/checksums.txt"; then
    echo -e "${RED}Error: Failed to download checksums${NC}"
    exit 1
fi

# Verify checksum
echo "Verifying checksum..."
EXPECTED_CHECKSUM=$(grep "ccq-${DETECTED_OS}-${DETECTED_ARCH}" "$INSTALL_DIR/checksums.txt" | awk '{print $1}')
if [[ -z "$EXPECTED_CHECKSUM" ]]; then
    echo -e "${RED}Error: Checksum not found for this platform${NC}"
    exit 1
fi

if command -v sha256sum &> /dev/null; then
    ACTUAL_CHECKSUM=$(sha256sum "$INSTALL_PATH.tmp" | awk '{print $1}')
elif command -v shasum &> /dev/null; then
    ACTUAL_CHECKSUM=$(shasum -a 256 "$INSTALL_PATH.tmp" | awk '{print $1}')
else
    echo -e "${YELLOW}Warning: No checksum tool found, skipping verification${NC}"
    ACTUAL_CHECKSUM="$EXPECTED_CHECKSUM"
fi

if [[ "$ACTUAL_CHECKSUM" != "$EXPECTED_CHECKSUM" ]]; then
    echo -e "${RED}Error: Checksum mismatch!${NC}"
    echo "Expected: $EXPECTED_CHECKSUM"
    echo "Got: $ACTUAL_CHECKSUM"
    exit 1
fi

# Move to final location
mv "$INSTALL_PATH.tmp" "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"

# Verify installation
if ! VERSION=$("$INSTALL_PATH" --version 2>&1); then
    echo -e "${RED}Error: Installation failed - binary cannot execute${NC}"
    rm -f "$INSTALL_PATH"
    exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^ccq\ version ]]; then
    echo -e "${RED}Error: Invalid binary - unexpected version format${NC}"
    rm -f "$INSTALL_PATH"
    exit 1
fi

echo -e "${GREEN}✓ $VERSION installed successfully to $INSTALL_PATH${NC}"
echo ""

# Check if ~/.local/bin is in PATH (handle all cases)
if [[ ":${PATH}:" != *":${HOME}/.local/bin:"* ]] && \
   [[ "${PATH}" != "${HOME}/.local/bin:"* ]] && \
   [[ "${PATH}" != *":${HOME}/.local/bin" ]] && \
   [[ "${PATH}" != "${HOME}/.local/bin" ]]; then
    echo -e "${YELLOW}Warning: ~/.local/bin is not in your PATH${NC}"
    echo ""
    echo "Add this to your shell profile:"
    echo '  export PATH="$HOME/.local/bin:$PATH"'
    echo ""
    echo "Common shell profiles:"
    echo "  - Bash: ~/.bashrc or ~/.bash_profile"
    echo "  - Zsh: ~/.zshrc"
    echo "  - Fish: ~/.config/fish/config.fish"
    echo ""
    echo "Then reload your shell or run:"
    echo "  source ~/.bashrc  # (or your shell's profile)"
else
    echo "You can now run: ccq"
fi
