# Verification

Do not trust LottaBits; check it. This page describes how to verify a download, a result and a printout. Commands assume a
release tag `vX.Y.Z` and the binary name of your platform.

## 1. The download

Every release lists `SHA256SUMS`, an SPDX SBOM (`lottabits-vX.Y.Z.spdx.json`) and carries build attestations created by
GitHub Actions.

```sh
sha256sum --check --ignore-missing SHA256SUMS          # macOS: shasum -a 256 --check --ignore-missing SHA256SUMS
gh attestation verify lottabits-vX.Y.Z-linux-amd64 --repo mkorun/lottabits
```

The attestation shows that the file was built by this repository's release workflow from the tagged commit.

## 2. Rebuild it yourself

Releases are reproducible: the same source and the Go version pinned in `mise.toml` give byte-identical binaries.

```sh
git clone https://github.com/mkorun/lottabits && cd lottabits && git checkout vX.Y.Z
mise install                                           # or install the Go version from mise.toml
mise exec -- sh scripts/build-release.sh vX.Y.Z        # writes dist/release/ and dist/release/SHA256SUMS
cd dist/release && sha256sum --check --ignore-missing /path/to/downloaded/SHA256SUMS
```

All binaries, the printables archive (HTML) and the SBOM must match. Only the PDF archive is not reproducible (PDFs carry
creation dates); its checksum is listed for download integrity only.

## 3. The program on your device

```sh
lottabits version      # release, source revision, Go version, identifiers and data hashes
lottabits selftest     # recomputes the official BIP39 vectors and all LottaBits vectors, checks every data hash
```

The hashes printed by `version` must match `SPEC.md` sections 5.1 and 7.2.

## 4. A result

Every result is a function of the draws alone. Recompute it with the independent Python implementation (standard library
only, readable in a few minutes) on the same offline device or with test draws:

```sh
python3 tools/crosscheck.py seed < draws.txt
python3 tools/crosscheck.py passphrase --wordlist de < draws.txt
```

For a seed you can also check on paper: words 1 to 23 must match your booklet lookups, word 24 must lie in the block of your
extra bits, and a BIP39 wallet must accept the phrase. Any BIP39 tool used for checking must run offline as well.

## 5. A printout

- Every data page shows a data hash (`SPEC.md` section 9.1). Compare it with the same page of the PDF from a release you
  verified, or with `manifest.json` of the printables you generated.
- `python3 tools/check_printables.py DIR` checks generated printables completely: every entry, every hash, both languages,
  the booklet imposition.
- Spot checks without a computer, for printouts of unknown origin: pick a few entries and compare them with an independent
  copy of the word list. For the seed booklet, the word with number `N` is line `N` of the official BIP39 English list;
  on page `p` the entry for second draw `b` has number `(p − 1) × 64 + b` for `p` from 1 to 32. For a passphrase
  booklet, the word on page `a` for second draw `b` is line `(a − 1) × 88 + b` of the word list file.
- On chip sheets the calibration line must measure 50 mm.

## 6. The repository

Every change runs the same gate locally and in CI: formatting, vet, lint with banned imports, tests with the race detector,
the independent Python implementation, the compiled binary's self-test and an end-to-end comparison, and the printables
check. See `scripts/check.sh` and `.github/workflows/`.
