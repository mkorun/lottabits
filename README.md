# LottaBits

> **LottaBits – physical randomness without dice.**
>
> **Seed. Password. Passphrase.**
>
> **One bag. Numbered chips. No software-generated randomness.**
>
> **The chips create secrets. The computer only calculates.**
>
> **Paper first. Software only where mathematics requires it.**

LottaBits is not a new cryptographic primitive. It explores a simpler human interface to physical randomness: numbered
chips, drawn from one opaque bag with replacement, create a BIP39 seed, a password or a passphrase. Everything except the
checksum of a seed's last word can be done on paper; the small command-line tool only calculates.

The name refers to *drawing lots*: every draw from the bag yields a few bits of physical randomness. Chips in, bits out
– they even rhyme a little.

It started in 2024 with a bag of counting chips from a school-supply shop, number stickers and a printed booklet, as a way
to make a Bitcoin seed without rolling dice a hundred times. LottaBits is the cleaned-up, tested and documented version
of that experiment.

> **Status:** release candidate, not audited. Before you use it for anything of value, read [SECURITY.md](SECURITY.md)
> and try the seed workflow with a test wallet first.

## Three modes

| Mode | Chips | Draws | Result | Strength | Computer needed? |
|---|---|---|---|---|---|
| Seed | 01–64 | 46 | 24 BIP39 words | 256 bits | only for the checksum of word 24, unless your hardware wallet offers it |
| Password | 01–88 | 1 per character | characters from a fixed 88-character alphabet | 6.46 bits per draw | no |
| Passphrase | 01–88 | 2 per word | words from a fixed 7744-word list (English or German) | 12.92 bits per word | no |

Twelve draws give about 77.5 bits in either password or passphrase mode: twelve characters or six words.

## How the seed works

Two draws `a`, `b` from chips 01–64 give 4096 equally likely pairs. LottaBits folds the first draw onto 32 groups, so
every BIP39 word has exactly two pairs and stays equally likely:

```text
word number = ((a − 1) mod 32) × 64 + b          (1 … 2048)
extra bit   = 0 if a ≤ 32, 1 if a ≥ 33
```

23 pairs (46 draws) give the first 23 words, 253 bits. The extra bits of the first three pairs are the three missing
entropy bits, so all 256 bits come from the chips. Word 24 is three entropy bits plus the 8-bit SHA-256 checksum; the
three bits fix one of eight blocks of 256 words. Some hardware wallets list the valid final words or accept the final
bits; with those you can finish the seed on the device itself ([which ones](docs/hardware-wallets.md); a few need their
companion app to start a recovery). The normative definition, test vectors included, is in [SPEC.md](SPEC.md).

## What you need

- an opaque bag and 88 chips that differ only in their printed number (01–64 and 65–88 may differ in colour so that both
  sets can share the bag),
- the printables of your language (English or German): booklets, record sheets, character map, quick references,
- for a seed without a suitable hardware wallet: the `lottabits` command on an offline device ([offline guide](docs/offline-use.md)).

## The command

Needed only for a seed whose word 24 is not chosen on a hardware wallet. For passwords and passphrases it is an optional
convenience: it saves the lookups, prints the character classes and can double-check a result made on paper.

```sh
lottabits selftest                      # run on the device you will use: checks vectors and data hashes
lottabits seed                          # asks for 23 pairs of draws, prints 24 words
lottabits password                      # optional: one draw per character, empty line to finish
lottabits passphrase --wordlist de      # two draws per word; the word list is always chosen explicitly
lottabits seed --lang de                # German messages; never changes a word list
```

Draws are read from standard input, never from command-line arguments. The program has no network code, writes no files,
keeps no logs and never generates random numbers; the build rejects any import that could
([ADR 0001](docs/adr/0001-go-standard-library-only.md)).

## Printables

Generated as self-contained HTML (fonts embedded, no external resources) in English and German from the same data the
command uses: `mise run printables`, PDFs with `mise run pdf`. Booklets are A5, printed double-sided, folded and stapled.
Every data page carries a short hash of its data ([SPEC.md](SPEC.md) section 9).

## Verification

Every result can be recomputed independently: `tools/crosscheck.py` is a second implementation in plain Python, and the
official BIP39 test vectors are part of the self-test. Release downloads carry checksums, build attestations and can be
rebuilt byte for byte. See [docs/verification.md](docs/verification.md).

## Documentation

- [SPEC.md](SPEC.md) – the normative specification
- [SECURITY.md](SECURITY.md) – safe use and reporting vulnerabilities
- [docs/threat-model.md](docs/threat-model.md) – what can go wrong and what LottaBits does about it
- [docs/offline-use.md](docs/offline-use.md) – where to run the command for a real seed
- [docs/hardware-wallets.md](docs/hardware-wallets.md) – which devices support the paper workflow for word 24
- [docs/verification.md](docs/verification.md) – checking downloads, results and printouts
- [docs/prior-art.md](docs/prior-art.md) – earlier physical methods this project builds on
- [docs/adr/](docs/adr/) – design decisions

## What LottaBits does not claim

It does not claim that chips are more random than fair dice, that the method is new or the first of its kind, that the
software is audited, or that anything is unhackable. Physical randomness depends on identical chips and thorough mixing.

## Development

Toolchain pinned in `mise.toml` (Go, golangci-lint, gitleaks). `mise install`, then `mise run check` runs every gate.
Working rules: [AGENTS.md](AGENTS.md).

## License

Code: MIT ([LICENSE](LICENSE)). Documentation and printables: CC BY 4.0 ([LICENSE-DOCS](LICENSE-DOCS)). Word lists and
fonts keep their own licenses; their provenance is documented next to them.
