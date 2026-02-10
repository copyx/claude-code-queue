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
