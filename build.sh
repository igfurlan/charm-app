#!/usr/bin/env bash
# Cross-compile charm-budget for every major OS/architecture.
# Go produces a single self-contained native binary per target — no runtime
# or dependencies needed on the machine that runs it.
#
# Usage: ./build.sh
# Output: ./dist/<os>-<arch>/charm-budget[.exe]
set -euo pipefail

APP="charm-budget"
OUT="dist"
rm -rf "$OUT"

# target = "GOOS/GOARCH"
targets=(
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"   # -> charm-budget.exe (double-click on Windows)
  "windows/arm64"
  "darwin/amd64"    # macOS Intel
  "darwin/arm64"    # macOS Apple Silicon (M1/M2/M3...)
)

for t in "${targets[@]}"; do
  os="${t%/*}"
  arch="${t#*/}"
  ext=""
  [ "$os" = "windows" ] && ext=".exe"

  dir="$OUT/$os-$arch"
  mkdir -p "$dir"

  echo "building $os/$arch..."
  # CGO disabled -> fully static binary, nothing to install on the target.
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags="-s -w" -o "$dir/$APP$ext" .
done

echo
echo "Done. Binaries are in ./$OUT/"
