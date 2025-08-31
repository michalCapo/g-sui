#!/usr/bin/env bash
set -euo pipefail

# Go to repo root (requires git). Fallback to script dir if not a git repo.
if command -v git >/dev/null 2>&1 && git rev-parse --show-toplevel >/dev/null 2>&1; then
    cd "$(git rev-parse --show-toplevel)"
else
    cd "$(dirname "$0")"
fi

if ! command -v reflex >/dev/null 2>&1; then
    echo "Error: 'reflex' not found in PATH." >&2
    echo "Install with: go install github.com/cespare/reflex@latest" >&2
    exit 1
fi

# -s = sequential (kill previous before starting new)
# -r '\.go$' = watch all .go files
exec reflex -s -r '\.go$' -- sh -c 'go run ./examples'
