#!/usr/bin/env sh
set -e

if ! command -v air >/dev/null 2>&1; then
  echo "[dev] 'air' is not installed. Installing to GOPATH/bin..."
  # Requires Go toolchain installed locally
  go install github.com/air-verse/air@latest
  export PATH="$(go env GOPATH)/bin:$PATH"
fi

echo "[dev] Starting Air with .air.toml"
exec air -c .air.toml
