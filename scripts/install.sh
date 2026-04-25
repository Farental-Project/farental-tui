#!/bin/sh
# Usage: curl -fsSL https://example.com/install.sh | sh
set -eu

APP_NAME="myapp"
REPO="youruser/yourrepo"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# --- detect OS ---
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
linux) OS="linux" ;;
darwin) OS="darwin" ;;
msys* | mingw* | cygwin*) OS="windows" ;;
*)
  echo "Unsupported OS: $OS" >&2
  exit 1
  ;;
esac

# --- detect arch ---
ARCH="$(uname -m)"
case "$ARCH" in
x86_64 | amd64) ARCH="amd64" ;;
arm64 | aarch64) ARCH="arm64" ;;
*)
  echo "Unsupported arch: $ARCH" >&2
  exit 1
  ;;
esac

# --- resolve latest version (or pin one) ---
VERSION="${VERSION:-$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
  grep '"tag_name":' | head -n1 | cut -d'"' -f4)}"

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version" >&2
  exit 1
fi

# --- download ---
ASSET="${APP_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading ${URL}..."
curl -fsSL "$URL" -o "$TMP/$ASSET"

# Optional: verify checksum
# curl -fsSL "${URL}.sha256" -o "$TMP/$ASSET.sha256"
# (cd "$TMP" && sha256sum -c "$ASSET.sha256")

# --- install ---
tar -xzf "$TMP/$ASSET" -C "$TMP"
mkdir -p "$INSTALL_DIR"
mv "$TMP/$APP_NAME" "$INSTALL_DIR/$APP_NAME"
chmod +x "$INSTALL_DIR/$APP_NAME"

echo "Installed $APP_NAME $VERSION to $INSTALL_DIR/$APP_NAME"

# --- PATH hint ---
case ":$PATH:" in
*":$INSTALL_DIR:"*) ;;
*) echo "Note: add $INSTALL_DIR to your PATH" ;;
esac
