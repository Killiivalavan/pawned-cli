#!/bin/bash
# Installer script for pawned

set -e

REPO="Killiivalavan/pawned-cli"
BIN_NAME="pawned"
BIN_DIR="$HOME/.local/bin"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_TARGET="linux";;
    Darwin*)    OS_TARGET="darwin";;
    *)          echo "Unsupported OS: ${OS}"; exit 1;;
esac

# Detect Arch
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64*)    ARCH_TARGET="amd64";;
    arm64*|aarch64*) ARCH_TARGET="arm64";;
    *)          echo "Unsupported Architecture: ${ARCH}"; exit 1;;
esac

echo "Detected OS: ${OS_TARGET}"
echo "Detected Architecture: ${ARCH_TARGET}"

# Construct the asset name and URL
ASSET_NAME="${BIN_NAME}-${OS_TARGET}-${ARCH_TARGET}"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

echo "Fetching latest release information..."
LATEST_RELEASE_JSON=$(curl -fsSL "${API_URL}")

if [ -z "$LATEST_RELEASE_JSON" ]; then
    echo "Error: Failed to fetch release information from GitHub API."
    exit 1
fi

# Extract the download URL for the specific asset using cross-platform grep (works on both Linux and macOS)
DOWNLOAD_URL=$(echo "$LATEST_RELEASE_JSON" | grep "browser_download_url" | grep "${ASSET_NAME}" | sed -n 's/.*"\(https:\/\/[^"]*\)".*/\1/p')

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find release asset for ${ASSET_NAME}"
    exit 1
fi

echo "Downloading ${ASSET_NAME}..."
mkdir -p "$BIN_DIR"
curl -fsSL "$DOWNLOAD_URL" -o "${BIN_DIR}/${BIN_NAME}"

# Make executable
chmod +x "${BIN_DIR}/${BIN_NAME}"

echo ""
echo "========================================"
echo "${BIN_NAME} installed successfully to ${BIN_DIR}/${BIN_NAME}"

# Check if BIN_DIR is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo ""
    echo "WARNING: ${BIN_DIR} is not in your PATH."
    echo "Please add the following line to your shell profile (e.g., ~/.bashrc or ~/.zshrc):"
    echo ""
    echo "export PATH=\"\$PATH:${BIN_DIR}\""
    echo ""
fi

echo "Run: ${BIN_NAME} help"
echo "========================================"
