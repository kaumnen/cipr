#!/bin/bash

set -e

INSTALL_DIR="$HOME/.cipr/bin"
CIPR_PATH="$INSTALL_DIR/cipr"
CONFIG_DIR="$HOME/.config/cipr"

if [ -f "$CIPR_PATH" ]; then
    rm "$CIPR_PATH"
    echo "Removed cipr binary: $CIPR_PATH"
else
    echo "cipr is not installed in $INSTALL_DIR."
fi

if [ -d "$INSTALL_DIR" ] && [ -z "$(ls -A "$INSTALL_DIR")" ]; then
    rmdir "$INSTALL_DIR"
    echo "Removed empty directory: $INSTALL_DIR"
fi

if [ -d "$CONFIG_DIR" ]; then
    rm -rf "$CONFIG_DIR"
    echo "Removed cipr configuration: $CONFIG_DIR"
else
    echo "No cipr configuration found in $CONFIG_DIR."
fi

echo "Please remember to remove the following line from your shell configuration file (e.g., .bashrc, .zshrc):"
echo "export PATH=\"\$HOME/.cipr/bin:\$PATH\""
echo "After removing the line, restart your terminal or run 'source ~/.bashrc' (or the appropriate config file) to update your PATH."

echo "Uninstallation complete."
