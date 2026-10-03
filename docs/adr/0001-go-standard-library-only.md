# ADR 0001: Go reference implementation, standard library only

Status: accepted, 2026-10-03.

## Context
The prototype computed the last BIP39 word with Bash scripts that depended on `bc`, `xxd`, `sha256sum` and Bash substring
syntax, behaved differently on GNU, BSD and BusyBox, and drew three entropy bits from `/dev/urandom`.
Tokenware needs one small, auditable implementation that runs offline on Windows, Linux, macOS and ARM boards.

## Decision
- The reference implementation is a Go command-line program, built as one static binary per platform.
- Only the Go standard library is used. Any external module, including `golang.org/x`, needs its own ADR first;
  the gate rejects any module requirement in `go.mod`.
- Imports that contradict the project's principles are banned by the linter for production code and tests alike:
  `crypto/rand`, `math/rand` (all versions), `net` (all subpackages), `os/exec`, `log` (including `log/slog`), `unsafe`, `plugin`.
- Toolchain versions are pinned in `mise.toml`; `go.mod` declares the oldest supported language version and CI tests it
  together with the pinned one.

## Consequences
- SHA-256, encoding and templating come from the standard library; the audit surface is our own code plus Go itself.
- Hidden terminal input (echo off) has no portable standard-library API; how secret input is read is decided in the specification.
- A future second implementation for cross-verification (for example minimal JavaScript) is a separate decision.
