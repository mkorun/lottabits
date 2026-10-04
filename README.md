# LottaBits

> **LottaBits – physical randomness without dice.**
>
> **Seed. Password. Passphrase.**
>
> **One bag. Numbered chips. No software-generated randomness.**

**Status: work in progress. Nothing here is usable yet. Do not use it for real secrets.**

LottaBits is not a new cryptographic primitive. It explores a simpler human interface to physical randomness:
numbered chips drawn from one bag with replacement create BIP39 seeds, passwords and passphrases.
The chips create secrets. The computer only calculates.

The name refers to *drawing lots*: every draw from the bag yields a few bits of physical randomness.

## Development

Toolchain pinned in `mise.toml` (Go, golangci-lint, gitleaks). `mise install`, then `mise run check` runs every gate.
Working rules: `AGENTS.md`. Decisions: `docs/adr/`.

## License

Code: MIT (`LICENSE`). Documentation and printables: CC BY 4.0 (`LICENSE-DOCS`).
Third-party word lists keep their own licenses; their provenance is documented next to them.
