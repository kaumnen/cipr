#!/bin/bash

set -euo pipefail

REPO="kaumnen/cipr"
INSTALL_DIR="$HOME/.cipr/bin"

OS_RAW=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH_RAW=$(uname -m)

case "$OS_RAW" in
    darwin) OS="darwin" ;;
    linux)  OS="linux" ;;
    *)
        echo "Unsupported OS: $OS_RAW"
        exit 1
        ;;
esac

case "$ARCH_RAW" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo "Unsupported architecture: $ARCH_RAW"
        exit 1
        ;;
esac

ARCHIVE_NAME="cipr_${OS}_${ARCH}.tar.gz"

echo "Detecting latest release for $OS/$ARCH..."

API_RESPONSE=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest")

ARCHIVE_URL=$(echo "$API_RESPONSE" \
    | grep "browser_download_url.*${ARCHIVE_NAME}" \
    | head -n 1 \
    | cut -d : -f 2,3 \
    | tr -d \" \
    | xargs)

CHECKSUMS_URL=$(echo "$API_RESPONSE" \
    | grep "browser_download_url.*checksums.txt" \
    | head -n 1 \
    | cut -d : -f 2,3 \
    | tr -d \" \
    | xargs)

if [ -z "$ARCHIVE_URL" ]; then
    echo "Could not find a release asset matching ${ARCHIVE_NAME}."
    echo "Check https://github.com/${REPO}/releases for available builds."
    exit 1
fi

if [ -z "$CHECKSUMS_URL" ]; then
    echo "Could not find checksums.txt in the latest release."
    exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading $ARCHIVE_URL"
curl -fsSL -o "$TMPDIR/$ARCHIVE_NAME" "$ARCHIVE_URL"

echo "Downloading checksums"
curl -fsSL -o "$TMPDIR/checksums.txt" "$CHECKSUMS_URL"

echo "Verifying checksum"
(
    cd "$TMPDIR"
    if command -v sha256sum >/dev/null 2>&1; then
        grep " $ARCHIVE_NAME\$" checksums.txt | sha256sum -c -
    elif command -v shasum >/dev/null 2>&1; then
        grep " $ARCHIVE_NAME\$" checksums.txt | shasum -a 256 -c -
    else
        echo "Neither sha256sum nor shasum is available; cannot verify checksum."
        exit 1
    fi
)

echo "Extracting"
tar -xzf "$TMPDIR/$ARCHIVE_NAME" -C "$TMPDIR"

mkdir -p "$INSTALL_DIR"
mv "$TMPDIR/cipr" "$INSTALL_DIR/cipr"
chmod +x "$INSTALL_DIR/cipr"

echo "Installation complete. The 'cipr' command has been installed to $INSTALL_DIR"
echo "Please ensure $INSTALL_DIR is in your PATH."
echo "You may need to add the following line to your shell configuration file (e.g., .bashrc, .zshrc):"
echo "export PATH=\"\$HOME/.cipr/bin:\$PATH\""
echo "After adding the line, restart your terminal or run 'source ~/.bashrc' (or the appropriate config file) to update your PATH."
