#!/usr/bin/env bash

set -e

SOURCE_DIR=".git-samples"
HOOKS_DIR=".git/hooks"

echo "Installing Git hooks from $SOURCE_DIR to $HOOKS_DIR..."

# Ensure hooks directory exists
mkdir -p "$HOOKS_DIR"

# Loop over all files in .git-samples
for file in "$SOURCE_DIR"/*; do
  if [ -f "$file" ]; then
    echo "Processing $(basename "$file")..."

    # Convert line endings to LF
    dos2unix "$file" >/dev/null 2>&1 || true

    # Copy file to .git/hooks
    cp "$file" "$HOOKS_DIR/$(basename "$file")"

    # Make the hook executable
    chmod +x "$HOOKS_DIR/$(basename "$file")"
  fi
done

echo "Git hooks installed successfully!"