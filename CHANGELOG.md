# Changelog

All notable changes are listed here. Versions follow [semantic versioning](https://semver.org/). Algorithm and data
identifiers (`lottabits-seed-v1` and so on) are versioned separately and never change their meaning.

## [1.0.0-rc.1]

First public release candidate.

### Modes
- `lottabits-seed-v1`: 46 draws of chips 01–64 give a 24-word BIP39 seed with 256 bits of physical entropy; the extra bits
  of the first three pairs choose the block of word 24.
- `lottabits-password-88-v1`: one draw of chips 01–88 per character from a frozen 88-character alphabet.
- `lottabits-passphrase-7744-v1`: two draws per word from `lottabits-en-7744-v1` (EFF long list) or `lottabits-de-7744-v1`
  (dys2p list), each the first 7744 entries of a pinned upstream file.

### Command line
- `lottabits seed | password | passphrase --wordlist en|de | selftest | version`, English and German messages
  (`--lang de`), draws only from standard input.

### Printables
- English and German: quick reference, chip inventory and cut-out sheets, seed booklet, record sheet and quick reference,
  password character map and record sheet, passphrase booklets and record sheet. Booklets are A5 for saddle stitching.
  Fonts are embedded; every data page shows a data hash.

### Verification
- Exhaustive mapping tests, official BIP39 vectors, an independent Python implementation compared with the compiled binary,
  independent checks of the generated printables, reproducible builds with SBOM and build attestations.

[1.0.0-rc.1]: https://github.com/mkorun/lottabits/releases/tag/v1.0.0-rc.1
