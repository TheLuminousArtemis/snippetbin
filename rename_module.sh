#!/bin/bash

# Set your current and new module paths
export CUR="github.com/theluminousartemis/letsgo_snippetbox"
export NEW="github.com/theluminousartemis/snippetbin"

# Check if module paths are set
if [ -z "$CUR" ] || [ -z "$NEW" ]; then
    echo "ERROR: Both CUR and NEW environment variables must be set"
    exit 1
fi

# Create a backup of go.mod
cp go.mod go.mod.bak

# Update the module path in go.mod
echo "Updating module path in go.mod..."
go mod edit -module ${NEW}

# Update all import paths in .go files
echo "Updating imports in .go files..."
find . -type f -name '*.go' -exec perl -pi -e 's/$ENV{CUR}/$ENV{NEW}/g' {} \;

# Clean up dependencies
echo "Cleaning up dependencies..."
go mod tidy

echo "Module rename complete!"