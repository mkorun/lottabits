#!/bin/sh
# The single gate: the same command runs locally (`mise run check`), as a pre-push hook and in CI.
set -eu
cd "$(dirname "$0")/.."
step() { printf '\n== %s\n' "$1"; }

step "formatting (gofmt)"
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  printf 'not gofmt-formatted:\n%s\n' "$unformatted" >&2
  exit 1
fi

step "standard library only (no module requirements)"
modules=$(go list -m all | sed 1d)
if [ -n "$modules" ]; then
  printf 'external modules need an ADR and are not allowed yet:\n%s\n' "$modules" >&2
  exit 1
fi

step "vet"
go vet ./...

step "lint (golangci-lint, including the import bans)"
golangci-lint run ./...

step "tests (race detector, no cache)"
go test -race -count=1 ./...

printf '\ncheck: all gates passed\n'
