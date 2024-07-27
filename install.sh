#!/bin/bash

set -e 

ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

if [ "$ARCH" = "arm64" ] && [ "$OS" = "darwin" ]; then
    BINARY_NAME="cipr-darwin-arm64"
else
    echo "Unsupported architecture: $ARCH on $OS"
    exit 1
fi

RELEASE_URL=$(curl -s https://api.github.com/repos/kaumnen/cipr/releases/latest \
    | grep "browser_download_url.*$BINARY_NAME" \
    | cut -d : -f 2,3 \
    | tr -d \" \
    | xargs)

if [ -z "$RELEASE_URL" ]; then
    echo "Failed to fetch the latest release URL. Please check your internet connection and try again."
    exit 1
fi

echo "Downloading the latest release from: $RELEASE_URL"

curl -L -o cipr "$RELEASE_URL"

INSTALL_DIR="$HOME/.cipr/bin"
mkdir -p "$INSTALL_DIR"
chmod +x cipr
mv cipr "$INSTALL_DIR"

echo "Installation complete. The 'cipr' command has been installed to $INSTALL_DIR"
echo "Please ensure $INSTALL_DIR is in your PATH."
echo "You may need to add the following line to your shell configuration file (e.g., .bashrc, .zshrc):"
echo "export PATH=\"\$HOME/.cipr/bin:\$PATH\""
echo "After adding the line, restart your terminal or run 'source ~/.bashrc' (or the appropriate config file) to update your PATH."