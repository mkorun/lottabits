# AGENTS.md

Rules for every agent (and human) working in this repository. `CLAUDE.md` imports this file.

LottaBits is a small, auditable human interface to physical randomness: numbered tokens in one bag create
BIP39 seeds, passwords and passphrases; software only calculates deterministically where mathematics requires it.
Paper first: the printables are a primary interface, not decorative assets.

## Sources of truth
- `SPEC.md` is normative. Nothing is implemented that is not specified there; behaviour changes start with the specification.
  CLI, printables and tests implement the same rules and never diverge silently.
- Architecture decisions are one-page ADRs in `docs/adr/`.

## Non-negotiables
- No software-generated randomness: production code never imports `crypto/rand`, `math/rand` or `math/rand/v2` (checked by the gate).
- No network code, telemetry, update checks, clipboard access, persistence or logging of draws, words, seeds, passwords or passphrases.
  Secret input is read from stdin or an interactive prompt, never from command-line arguments.
- UI language and word-list language are separate concepts: `--lang` never changes a BIP39 or passphrase word list.
  Translations change text only, never indexes, mappings, word lists or test vectors.
- Word lists are pinned data with documented provenance (upstream URL, immutable reference, license, SHA-256, derivation rule).
- Printables are generated from the same canonical data and templates for every language, A4, black-and-white friendly,
  with no external fonts, scripts, CDNs or network resources, and carry the release version.
- Claims: never "more random than dice", "unhackable", "audited" (unless an audit happened), "new cryptography" or "the first".

## Go
- Toolchain pinned in `mise.toml` (`mise install`; then `mise exec -- go ...` or put `~/.local/share/mise/shims` on `PATH`).
- Standard library only. Any dependency, including `golang.org/x`, needs an ADR first; the default answer is no.
- Layout: `internal/{bip39,password,passphrase}` hold pure functions (no I/O, no globals with state);
  `cmd/lottabits` is a thin shell for input, output and messages. Functional core, imperative shell.
- Errors are values; a message says what happened, why and what to do. Exit codes: 0 success, 1 failure, 2 usage error.
  stdout carries the result, stderr diagnostics.
- Keep it small: functions short, low nesting, no boolean-flag parameters, no premature abstraction. Delete before adding.

## Tests and gate
- One gate, `mise run check` (`scripts/check.sh`), runs the same locally, in the pre-push hook and in CI: `gofmt`,
  standard library only, `go vet`, `golangci-lint` (including the `depguard` import bans in `.golangci.yml`), `go test -race`,
  `tools/crosscheck.py verify` (Python standard library only), and on the compiled binary `lottabits selftest` plus
  `tools/crosscheck.py cli` (end-to-end comparison of the CLI output with the independent implementation).
  CI adds `govulncheck`; `scripts/pre-push.sh` and CI add `gitleaks`. The gate must pass before every commit.
  A rule is either checked automatically or dropped.
- Mappings are tested exhaustively (all 4096 seed draw pairs, all 88 password draws, all 7744 passphrase coordinates).
  Test vectors live as data in `vectors/` (embedded for `lottabits selftest`); BIP39 results are cross-checked against
  official vectors and the independent implementation `tools/crosscheck.py`.
- Tests are hermetic: no network, no clock, no randomness.

## Conventions
- Chat with the human is German. All repository content is English, except the German locale data, the German word list
  and the generated German printables.
- Small coherent commits with English messages.
- Report what could not be verified (for example hardware-wallet behaviour or a physical test print) instead of assuming it.
