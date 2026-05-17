#!/usr/bin/env bash
# install.sh - build the commit-composer binary into ./bin so the
# Claude plugin launcher can find it.
#
# Use this for local development. Production deployments should publish the
# binary alongside the plugin (under .claude-plugin/bin/) so users don't need
# a Go toolchain.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

if ! command -v go >/dev/null 2>&1; then
  printf 'install.sh: go toolchain not on PATH\n' >&2
  exit 1
fi

OUT_DIR="${OUT_DIR:-$REPO_ROOT/.claude-plugin/bin}"
mkdir -p "$OUT_DIR"

printf 'building commit-composer -> %s\n' "$OUT_DIR/commit-composer"
go build -trimpath -o "$OUT_DIR/commit-composer" ./cmd/commit-composer

chmod +x "$OUT_DIR/commit-composer"
printf 'done. Binary: %s\n' "$OUT_DIR/commit-composer"
