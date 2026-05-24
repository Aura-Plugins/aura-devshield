#!/usr/bin/env bash
# Aura DevShield installer
# Usage: curl -sSfL https://raw.githubusercontent.com/Aura-Plugins/aura-devshield/main/install.sh | bash
# Override install directory: INSTALL_DIR=/usr/bin bash install.sh

set -euo pipefail

REPO="Aura-Plugins/aura-devshield"
APP="aura-devshield"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ── Detect OS ──────────────────────────────────────────────────────────────────
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# ── Detect architecture ────────────────────────────────────────────────────────
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# arm64 Linux is not yet distributed — fall back gracefully
if [ "$OS" = "linux" ] && [ "$ARCH" = "arm64" ]; then
  echo "Linux arm64 binaries are not yet available. Build from source:" >&2
  echo "  git clone https://github.com/$REPO && cd aura-devshield && make install" >&2
  exit 1
fi

# ── Get latest release tag ─────────────────────────────────────────────────────
if command -v curl &>/dev/null; then
  FETCH="curl -sSfL"
elif command -v wget &>/dev/null; then
  FETCH="wget -qO-"
else
  echo "curl or wget is required" >&2; exit 1
fi

LATEST=$($FETCH "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": "\(.*\)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Could not determine latest release. Check https://github.com/${REPO}/releases" >&2
  exit 1
fi

# ── Download ───────────────────────────────────────────────────────────────────
BINARY="${APP}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${LATEST}/checksums.txt"
TMP=$(mktemp)

echo "Installing aura-devshield ${LATEST} (${OS}/${ARCH})..."
$FETCH "$URL" -o "$TMP"
chmod +x "$TMP"

# ── Verify checksum (if sha256sum is available) ────────────────────────────────
if command -v sha256sum &>/dev/null; then
  EXPECTED=$($FETCH "$CHECKSUM_URL" | grep "$BINARY" | awk '{print $1}')
  ACTUAL=$(sha256sum "$TMP" | awk '{print $1}')
  if [ "$EXPECTED" != "$ACTUAL" ]; then
    echo "Checksum mismatch — aborting" >&2
    rm -f "$TMP"
    exit 1
  fi
  echo "Checksum verified."
elif command -v shasum &>/dev/null; then
  EXPECTED=$($FETCH "$CHECKSUM_URL" | grep "$BINARY" | awk '{print $1}')
  ACTUAL=$(shasum -a 256 "$TMP" | awk '{print $1}')
  if [ "$EXPECTED" != "$ACTUAL" ]; then
    echo "Checksum mismatch — aborting" >&2
    rm -f "$TMP"
    exit 1
  fi
  echo "Checksum verified."
fi

# ── Install ────────────────────────────────────────────────────────────────────
if [ ! -d "$INSTALL_DIR" ]; then
  echo "Install directory does not exist: $INSTALL_DIR" >&2
  exit 1
fi

if [ ! -w "$INSTALL_DIR" ]; then
  echo "No write permission to $INSTALL_DIR — trying with sudo..."
  sudo mv "$TMP" "${INSTALL_DIR}/${APP}"
else
  mv "$TMP" "${INSTALL_DIR}/${APP}"
fi

echo ""
echo "Installed: ${INSTALL_DIR}/${APP}"
echo "Version:   ${LATEST}"
echo ""
echo "Run: aura-devshield scan"
