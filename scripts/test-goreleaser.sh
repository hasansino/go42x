#!/bin/bash

# Test GoReleaser configuration locally
# This script validates the configuration and performs a snapshot build

set -e

echo "=== GoReleaser Configuration Test ==="
echo

# Check if goreleaser is installed
if ! command -v goreleaser &> /dev/null; then
    echo "GoReleaser is not installed. Installing..."
    go install github.com/goreleaser/goreleaser/v2@latest
fi

# Validate configuration
echo "1. Validating .goreleaser.yaml..."
if goreleaser check; then
    echo "✓ Configuration is valid"
else
    echo "✗ Configuration has errors"
    exit 1
fi

echo
echo "2. Running snapshot build (no upload)..."
echo "This will build all artifacts locally without creating a release"
echo

# Create a snapshot build
if goreleaser release --snapshot --clean --skip=publish,sign; then
    echo
    echo "✓ Snapshot build successful!"
    echo
    echo "Built artifacts:"
    ls -la dist/ 2>/dev/null | grep -E '\.(tar\.gz|deb|rpm|apk)$' || echo "Check dist/ directory for artifacts"
else
    echo "✗ Snapshot build failed"
    exit 1
fi

echo
echo "=== Test Complete ==="
echo "Configuration is valid and builds successfully!"
echo
echo "To create a real release:"
echo "1. Commit and push all changes"
echo "2. Create and push a semver tag: git tag v1.0.0 && git push origin v1.0.0"
echo "3. The GitHub Action will automatically build and release"