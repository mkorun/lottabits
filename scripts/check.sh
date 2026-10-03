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

step "independent implementation (tools/crosscheck.py verify)"
if ! crosscheck=$(python3 tools/crosscheck.py verify 2>&1); then
  printf '%s\n' "$crosscheck" >&2
  exit 1
fi
printf '%s\n' "$crosscheck" | tail -n 1

step "compiled binary: selftest and end-to-end comparison with tools/crosscheck.py"
build=$(mktemp -d "${TMPDIR:-/tmp}/lottabits-check.XXXXXX")
trap 'rm -rf -- "$build"' EXIT INT TERM
go build -trimpath -o "$build/lottabits" ./cmd/lottabits
if ! selftest=$("$build/lottabits" selftest 2>&1); then
  printf '%s\n' "$selftest" >&2
  exit 1
fi
printf '%s\n' "$selftest" | tail -n 1
if ! e2e=$(python3 tools/crosscheck.py cli "$build/lottabits" 2>&1); then
  printf '%s\n' "$e2e" >&2
  exit 1
fi
printf '%s\n' "$e2e" | tail -n 1

printf '\ncheck: all gates passed\n'
